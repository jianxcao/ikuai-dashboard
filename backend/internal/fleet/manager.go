package fleet

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"ikuai-dashboard/backend/internal/config"
	"ikuai-dashboard/backend/internal/ikuaiv4"
)

var (
	ErrUnconfigured   = errors.New("尚未配置爱快 V4 路由器节点")
	ErrRouterNotFound = errors.New("未找到指定的爱快 V4 路由器节点")
)

// ClientFactory allows tests to supply a custom HTTP transport when building V4 clients.
type ClientFactory func(router config.RouterConfig) *ikuaiv4.Client

// RouterHealthDTO summarizes live connectivity & telemetry for one V4 router in the fleet.
type RouterHealthDTO struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	URL           string   `json:"url"`
	Tags          []string `json:"tags"`
	Note          string   `json:"note"`
	Mock          bool     `json:"mock"`
	Source        string   `json:"source"`
	Active        bool     `json:"active"`
	Online        bool     `json:"online"`
	LatencyMs     int64    `json:"latency_ms"`
	Hostname      string   `json:"hostname"`
	Version       string   `json:"version"`
	Verstring     string   `json:"verstring"`
	Arch          string   `json:"arch"`
	Uptime        int64    `json:"uptime"`
	CPUPercent    float64  `json:"cpu_percent"`
	MemoryPercent float64  `json:"memory_percent"`
	UploadBps     int64    `json:"upload_bps"`
	DownloadBps   int64    `json:"download_bps"`
	OnlineUsers   int      `json:"online_users"`
	Connections   int      `json:"connections"`
	CheckedAt     int64    `json:"checked_at"`
	Error         string   `json:"error,omitempty"`
}

// RouterTestRequest is used by POST /api/v1/routers/test to validate connectivity before saving.
type RouterTestRequest struct {
	ID                 string `json:"id"`
	OriginalID         string `json:"original_id,omitempty"`
	Name               string `json:"name"`
	URL                string `json:"url"`
	Token              string `json:"token"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
	Mock               bool   `json:"mock"`
}

// Manager caches pure-V4 clients and handles multi-router switching & fleet probing.
type Manager struct {
	mu      sync.RWMutex
	clients map[string]*ikuaiv4.Client
	factory ClientFactory
}

var GlobalManager = NewManager()

// NewManager creates a fleet manager.
func NewManager() *Manager {
	return &Manager{
		clients: make(map[string]*ikuaiv4.Client),
		factory: func(r config.RouterConfig) *ikuaiv4.Client {
			return ikuaiv4.NewClient(r.URL, r.Token, r.InsecureSkipVerify)
		},
	}
}

// SetClientFactory sets a custom client builder (for unit testing).
func (m *Manager) SetClientFactory(factory ClientFactory) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.factory = factory
	m.clients = make(map[string]*ikuaiv4.Client)
}

// InvalidateCache clears cached clients after config updates.
func (m *Manager) InvalidateCache() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients = make(map[string]*ikuaiv4.Client)
}

// Status returns a concise summary for /health and config responses.
func (m *Manager) Status() map[string]any {
	cfg := config.Snapshot()
	if cfg == nil || len(cfg.Routers) == 0 {
		return map[string]any{
			"mode":         "unconfigured",
			"version":      "v4",
			"router_id":    "",
			"router_name":  "",
			"router_count": 0,
			"error":        ErrUnconfigured.Error(),
		}
	}
	active, ok := cfg.ActiveRouter()
	if !ok {
		return map[string]any{
			"mode":         "unconfigured",
			"version":      "v4",
			"router_id":    "",
			"router_name":  "",
			"router_count": len(cfg.Routers),
			"error":        ErrUnconfigured.Error(),
		}
	}
	mode := "real"
	if active.Mock {
		mode = "mock"
	}
	return map[string]any{
		"mode":         mode,
		"version":      "v4",
		"router_id":    active.ID,
		"router_name":  active.Name,
		"source":       active.Source,
		"router_count": len(cfg.Routers),
		"error":        "",
	}
}

// ResolveRouter resolves either the requested routerID or the active router, returning its V4 client (nil in mock mode).
func (m *Manager) ResolveRouter(routerID string) (config.RouterConfig, *ikuaiv4.Client, error) {
	return m.resolveRouterWith(routerID, func(cfg *config.AppConfig, key string) (config.RouterConfig, bool) {
		return cfg.FindRouter(key)
	})
}

// ResolveRouterByKey resolves either a router matching ID/Name or the active router when key is empty.
func (m *Manager) ResolveRouterByKey(key string) (config.RouterConfig, *ikuaiv4.Client, error) {
	return m.resolveRouterWith(key, func(cfg *config.AppConfig, k string) (config.RouterConfig, bool) {
		return cfg.FindRouterByIDOrName(k)
	})
}

func (m *Manager) resolveRouterWith(key string, finder func(*config.AppConfig, string) (config.RouterConfig, bool)) (config.RouterConfig, *ikuaiv4.Client, error) {
	cfg := config.Snapshot()
	if cfg == nil || len(cfg.Routers) == 0 {
		return config.RouterConfig{}, nil, ErrUnconfigured
	}

	trimmed := strings.TrimSpace(key)
	var (
		router config.RouterConfig
		ok     bool
	)
	if trimmed != "" {
		router, ok = finder(cfg, trimmed)
		if !ok {
			return config.RouterConfig{}, nil, fmt.Errorf("%w: %s", ErrRouterNotFound, trimmed)
		}
	} else {
		router, ok = cfg.ActiveRouter()
		if !ok {
			return config.RouterConfig{}, nil, ErrUnconfigured
		}
	}

	if router.Mock {
		return router, nil, nil
	}

	client := m.getOrCreateClient(router)
	return router, client, nil
}

func (m *Manager) getOrCreateClient(router config.RouterConfig) *ikuaiv4.Client {
	m.mu.RLock()
	existing, ok := m.clients[router.ID]
	factory := m.factory
	m.mu.RUnlock()

	if ok && existing != nil &&
		existing.BaseURL == ikuaiv4.NormalizeBaseURL(router.URL) &&
		existing.Token == strings.TrimSpace(router.Token) &&
		existing.InsecureSkipVerify == router.InsecureSkipVerify {
		return existing
	}

	client := factory(router)
	m.mu.Lock()
	m.clients[router.ID] = client
	m.mu.Unlock()
	return client
}

// ProbeAll probes all configured V4 routers concurrently with a bounded timeout.
func (m *Manager) ProbeAll(ctx context.Context) []RouterHealthDTO {
	cfg := config.Snapshot()
	if cfg == nil || len(cfg.Routers) == 0 {
		return []RouterHealthDTO{}
	}

	results := make([]RouterHealthDTO, len(cfg.Routers))
	var wg sync.WaitGroup

	for i, router := range cfg.Routers {
		wg.Add(1)
		go func(idx int, r config.RouterConfig) {
			defer wg.Done()
			results[idx] = m.probeOne(ctx, r, r.ID == cfg.ActiveRouterID)
		}(i, router)
	}
	wg.Wait()
	return results
}

// TestRouter validates connectivity to a candidate V4 router before saving.
func (m *Manager) TestRouter(ctx context.Context, req RouterTestRequest) (*RouterHealthDTO, error) {
	urlStr := ikuaiv4.NormalizeBaseURL(req.URL)
	if urlStr == "" {
		return nil, errors.New("路由器地址不能为空")
	}

	token := strings.TrimSpace(req.Token)
	if token == "" {
		if cfg := config.Snapshot(); cfg != nil {
			if id := strings.TrimSpace(req.ID); id != "" {
				if existing, ok := cfg.FindRouter(id); ok {
					token = existing.Token
				}
			}
			if token == "" {
				if orig := strings.TrimSpace(req.OriginalID); orig != "" {
					if existing, ok := cfg.FindRouter(orig); ok {
						token = existing.Token
					}
				}
			}
		}
	}

	if !req.Mock && token == "" {
		return nil, errors.New("真实连接测试需要填写 API Token")
	}

	candidate := config.RouterConfig{
		ID:                 strings.TrimSpace(req.ID),
		Name:               strings.TrimSpace(req.Name),
		URL:                urlStr,
		Token:              token,
		Mock:               req.Mock,
		InsecureSkipVerify: req.InsecureSkipVerify,
		Source:             "test",
	}
	if candidate.Name == "" {
		candidate.Name = candidate.URL
	}

	health := m.probeOne(ctx, candidate, false)
	return &health, nil
}

func (m *Manager) probeOne(ctx context.Context, r config.RouterConfig, active bool) RouterHealthDTO {
	tags := []string{}
	if len(r.Tags) > 0 {
		tags = append([]string(nil), r.Tags...)
	}
	source := r.Source
	if source == "" {
		if r.Mock {
			source = "mock"
		} else {
			source = "yaml"
		}
	}

	dto := RouterHealthDTO{
		ID:        r.ID,
		Name:      r.Name,
		URL:       r.URL,
		Tags:      tags,
		Note:      r.Note,
		Mock:      r.Mock,
		Source:    source,
		Active:    active,
		CheckedAt: time.Now().Unix(),
	}

	if r.Mock {
		return fillMockRouterHealth(dto)
	}

	probeCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	var client *ikuaiv4.Client
	if r.ID != "" && r.Source != "test" {
		client = m.getOrCreateClient(r)
	} else {
		m.mu.RLock()
		factory := m.factory
		m.mu.RUnlock()
		client = factory(r)
	}

	start := time.Now()
	raw, err := client.Get(probeCtx, "/monitoring/system", nil)
	dto.LatencyMs = time.Since(start).Milliseconds()
	if err != nil {
		dto.Online = false
		dto.Error = err.Error()
		return dto
	}

	dto.Online = true
	sys := ikuaiv4.ExtractObject(raw)
	dto.Hostname = ikuaiv4.Str(sys, "hostname")
	dto.Uptime = ikuaiv4.Int64(sys, "uptime")

	if ver, ok := sys["verinfo"].(map[string]any); ok {
		dto.Version = ikuaiv4.Str(ver, "version")
		dto.Verstring = ikuaiv4.Str(ver, "verstring")
		dto.Arch = ikuaiv4.Str(ver, "arch") + "/" + ikuaiv4.Str(ver, "sysbit")
		dto.Arch = strings.Trim(dto.Arch, "/")
	}
	if cpuArr, ok := sys["cpu"].([]any); ok && len(cpuArr) > 0 {
		dto.CPUPercent = ikuaiv4.ToFloat64(cpuArr[0])
	}
	if mem, ok := sys["memory"].(map[string]any); ok {
		dto.MemoryPercent = ikuaiv4.Float64(mem, "used")
	}
	if stream, ok := sys["stream"].(map[string]any); ok {
		dto.UploadBps = ikuaiv4.Int64(stream, "upload")
		dto.DownloadBps = ikuaiv4.Int64(stream, "download")
		dto.Connections = ikuaiv4.Int(stream, "connect_num")
	}
	if users, ok := sys["online_user"].(map[string]any); ok {
		dto.OnlineUsers = ikuaiv4.Int(users, "count")
	}
	return dto
}

func fillMockRouterHealth(dto RouterHealthDTO) RouterHealthDTO {
	dto.Online = true
	dto.LatencyMs = 4
	if strings.Contains(dto.ID, "lab") {
		dto.LatencyMs = 11
		dto.Hostname = "iKuai-Edge-Lab"
		dto.Version = "4.0.314"
		dto.Verstring = "4.0.314-beta arm64 Build20261013"
		dto.Arch = "arm64"
		dto.Uptime = 1284300
		dto.CPUPercent = 14.2
		dto.MemoryPercent = 47
		dto.UploadBps = 2_450_000
		dto.DownloadBps = 9_820_000
		dto.OnlineUsers = 16
		dto.Connections = 428
		return dto
	}
	dto.Hostname = "iKuai-HQ-Core"
	dto.Version = "4.0.314"
	dto.Verstring = "4.0.314-beta x64 Build20261013"
	dto.Arch = "x86/x64"
	dto.Uptime = 3158145
	dto.CPUPercent = 8.4
	dto.MemoryPercent = 39
	dto.UploadBps = 6_420_000
	dto.DownloadBps = 28_640_000
	dto.OnlineUsers = 42
	dto.Connections = 1246
	return dto
}

var _ = http.MethodGet

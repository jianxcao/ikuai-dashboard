package v4service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"ikuai-dashboard/backend/internal/config"
	"ikuai-dashboard/backend/internal/ikuaiv4"
)

// OverviewDTO aggregates real-time telemetry from V4 monitoring & physical NIC APIs.
type OverviewDTO struct {
	Router           RouterSummaryDTO      `json:"router"`
	Stream           StreamSummaryDTO      `json:"stream"`
	CPU              CPUSummaryDTO         `json:"cpu"`
	Memory           MemorySummaryDTO      `json:"memory"`
	Connections      ConnectionSummaryDTO  `json:"connections"`
	OnlineUsers      OnlineUserSummaryDTO  `json:"online_users"`
	WanChecks        []WanCheckDTO         `json:"wan_checks"`
	InterfaceTraffic []InterfaceTrafficDTO `json:"interface_traffic"`
	PhysicalPorts    []PhysicalPortDTO     `json:"physical_ports"`
	WanHistory       []WanHistoryPointDTO  `json:"wan_history"`
	CollectedAt      int64                 `json:"collected_at"`
}

type RouterSummaryDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Hostname   string `json:"hostname"`
	GWID       string `json:"gwid"`
	LanIP      string `json:"lan_ip"`
	LinkStatus int    `json:"link_status"`
	Uptime     int64  `json:"uptime"`
	Version    string `json:"version"`
	Verstring  string `json:"verstring"`
	Arch       string `json:"arch"`
	BuildDate  string `json:"build_date"`
	Mock       bool   `json:"mock"`
}

type StreamSummaryDTO struct {
	UploadBps   int64 `json:"upload_bps"`
	DownloadBps int64 `json:"download_bps"`
	TotalUp     int64 `json:"total_up"`
	TotalDown   int64 `json:"total_down"`
}

type CPUSummaryDTO struct {
	AvgPercent float64   `json:"avg_percent"`
	Cores      []float64 `json:"cores"`
	FreqMHz    []int64   `json:"freq_mhz"`
	Temps      []float64 `json:"temps"`
}

type MemorySummaryDTO struct {
	TotalKB     int64   `json:"total_kb"`
	AvailableKB int64   `json:"available_kb"`
	FreeKB      int64   `json:"free_kb"`
	CachedKB    int64   `json:"cached_kb"`
	BuffersKB   int64   `json:"buffers_kb"`
	UsedPercent float64 `json:"used_percent"`
}

type ConnectionSummaryDTO struct {
	Total int `json:"total"`
	TCP   int `json:"tcp"`
	UDP   int `json:"udp"`
	ICMP  int `json:"icmp"`
}

type OnlineUserSummaryDTO struct {
	Total    int `json:"total"`
	Wired    int `json:"wired"`
	Wireless int `json:"wireless"`
	WiFi2G   int `json:"wifi_2g"`
	WiFi5G   int `json:"wifi_5g"`
}

type WanCheckDTO struct {
	ID              int     `json:"id"`
	Interface       string  `json:"interface"`
	ParentInterface string  `json:"parent_interface"`
	IPAddr          string  `json:"ip_addr"`
	Gateway         string  `json:"gateway"`
	Internet        string  `json:"internet"`
	Result          string  `json:"result"`
	ErrMsg          string  `json:"errmsg"`
	AutoSwitch      string  `json:"auto_switch"`
	UpdateTime      string  `json:"updatetime"`
	Comment         string  `json:"comment"`
	AvgRTTMs        float64 `json:"avg_rtt_ms"`
	DropRate        float64 `json:"drop_rate"`
}

type InterfaceTrafficDTO struct {
	Interface   string `json:"interface"`
	IPAddr      string `json:"ip_addr"`
	Comment     string `json:"comment"`
	UploadBps   int64  `json:"upload_bps"`
	DownloadBps int64  `json:"download_bps"`
	TotalUp     int64  `json:"total_up"`
	TotalDown   int64  `json:"total_down"`
	Connections int    `json:"connections"`
	UpDropped   int64  `json:"updropped"`
	DownDropped int64  `json:"downdropped"`
	UpPacked    int64  `json:"uppacked"`
	DownPacked  int64  `json:"downpacked"`
}

type PhysicalPortDTO struct {
	Name      string `json:"name"`
	Interface string `json:"interface"`
	Link      int    `json:"link"`
	Speed     int    `json:"speed"`
	Duplex    int    `json:"duplex"`
	MAC       string `json:"mac"`
	Type      string `json:"type"`
	Driver    string `json:"driver"`
	Model     string `json:"model"`
	Lock      int    `json:"lock"`
}

type WanHistoryPointDTO struct {
	Timestamp       int64   `json:"timestamp"`
	Interface       string  `json:"interface"`
	AvgUploadKBps   float64 `json:"avg_upload_kbps"`
	AvgDownloadKBps float64 `json:"avg_download_kbps"`
	MaxUploadKBps   float64 `json:"max_upload_kbps"`
	MaxDownloadKBps float64 `json:"max_download_kbps"`
	AvgRTTMs        float64 `json:"avg_rtt_ms"`
	MaxRTTMs        float64 `json:"max_rtt_ms"`
	MinRTTMs        float64 `json:"min_rtt_ms"`
	DropRate        float64 `json:"drop_rate"`
}

type LoadHistoryDTO struct {
	Metric    string           `json:"metric"`
	TimeRange string           `json:"time_range"`
	Aggregate string           `json:"aggregate"`
	Series    []map[string]any `json:"series"`
	Raw       any              `json:"raw,omitempty"`
}

// Legacy compatibility structs for /api/v1/monitor/interface
type LegacySummaryInfo struct {
	UploadSpeed      int64 `json:"upload_speed"`
	DownloadSpeed    int64 `json:"download_speed"`
	TotalConnections int   `json:"total_connections"`
	OnlineUsers      int   `json:"online_users"`
}

type LegacyWanStatus struct {
	Name    string `json:"name"`
	IP      string `json:"ip"`
	Proto   string `json:"proto"`
	Status  string `json:"status"`
	Comment string `json:"comment"`
}

type LegacyTrafficDetail struct {
	Name          string `json:"name"`
	IP            string `json:"ip"`
	UploadSpeed   int64  `json:"upload_speed"`
	DownloadSpeed int64  `json:"download_speed"`
	TotalUp       int64  `json:"total_up"`
	TotalDown     int64  `json:"total_down"`
	Connections   int    `json:"connections"`
	Comment       string `json:"comment"`
}

type LegacyInterfaceDataResult struct {
	Summary        LegacySummaryInfo     `json:"summary"`
	WanStatus      []LegacyWanStatus     `json:"wan_status"`
	TrafficDetails []LegacyTrafficDetail `json:"traffic_details"`
}

// GetOverview collects /monitoring/system, /monitoring/interfaces-status, /monitoring/interfaces-traffic and /interfaces/physical.
func GetOverview(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client) (*OverviewDTO, error) {
	if router.Mock || client == nil {
		return mockOverview(router), nil
	}

	dto := &OverviewDTO{
		Router: RouterSummaryDTO{
			ID:   router.ID,
			Name: router.Name,
			Mock: false,
		},
		WanChecks:        []WanCheckDTO{},
		InterfaceTraffic: []InterfaceTrafficDTO{},
		PhysicalPorts:    []PhysicalPortDTO{},
		WanHistory:       []WanHistoryPointDTO{},
		CollectedAt:      time.Now().Unix(),
	}

	// Fetch core overview endpoints concurrently to minimize cross-WAN latency.
	var (
		sysRaw     json.RawMessage
		sysErr     error
		statusRaw  json.RawMessage
		statusErr  error
		trafficRaw json.RawMessage
		trafficErr error
		phyRaw     json.RawMessage
		phyErr     error
		wg         sync.WaitGroup
	)
	wg.Add(4)
	go func() {
		defer wg.Done()
		sysRaw, sysErr = client.Get(ctx, "/monitoring/system", nil)
	}()
	go func() {
		defer wg.Done()
		statusRaw, statusErr = client.Get(ctx, "/monitoring/interfaces-status", nil)
	}()
	go func() {
		defer wg.Done()
		trafficRaw, trafficErr = client.Get(ctx, "/monitoring/interfaces-traffic", nil)
	}()
	go func() {
		defer wg.Done()
		phyRaw, phyErr = client.Get(ctx, "/interfaces/physical", nil)
	}()
	wg.Wait()

	// 1. /monitoring/system (required baseline)
	if sysErr != nil {
		return nil, sysErr
	}
	sys := ikuaiv4.ExtractObject(sysRaw)
	dto.Router.Hostname = ikuaiv4.Str(sys, "hostname")
	dto.Router.GWID = ikuaiv4.Str(sys, "gwid")
	dto.Router.LanIP = ikuaiv4.Str(sys, "ip_addr")
	dto.Router.LinkStatus = ikuaiv4.Int(sys, "link_status")
	dto.Router.Uptime = ikuaiv4.Int64(sys, "uptime")

	if ver, ok := sys["verinfo"].(map[string]any); ok {
		dto.Router.Version = ikuaiv4.Str(ver, "version")
		dto.Router.Verstring = ikuaiv4.Str(ver, "verstring")
		arch := ikuaiv4.Str(ver, "arch")
		sysbit := ikuaiv4.Str(ver, "sysbit")
		if arch != "" && sysbit != "" {
			dto.Router.Arch = arch + "/" + sysbit
		} else {
			dto.Router.Arch = arch + sysbit
		}
		dto.Router.BuildDate = ikuaiv4.Str(ver, "build_date")
	}

	if cpuArr, ok := sys["cpu"].([]any); ok {
		for _, item := range cpuArr {
			dto.CPU.Cores = append(dto.CPU.Cores, ikuaiv4.ToFloat64(item))
		}
		if len(dto.CPU.Cores) > 0 {
			dto.CPU.AvgPercent = dto.CPU.Cores[0]
		}
	}
	if freqArr, ok := sys["freq"].([]any); ok {
		for _, item := range freqArr {
			if n, err := strconv.ParseInt(strings.TrimSpace(fmt.Sprint(item)), 10, 64); err == nil {
				dto.CPU.FreqMHz = append(dto.CPU.FreqMHz, n)
			}
		}
	}
	if tempArr, ok := sys["cputemp"].([]any); ok {
		for _, item := range tempArr {
			dto.CPU.Temps = append(dto.CPU.Temps, ikuaiv4.ToFloat64(item))
		}
	}

	if mem, ok := sys["memory"].(map[string]any); ok {
		dto.Memory = MemorySummaryDTO{
			TotalKB:     ikuaiv4.Int64(mem, "total"),
			AvailableKB: ikuaiv4.Int64(mem, "available"),
			FreeKB:      ikuaiv4.Int64(mem, "free"),
			CachedKB:    ikuaiv4.Int64(mem, "cached"),
			BuffersKB:   ikuaiv4.Int64(mem, "buffers"),
			UsedPercent: ikuaiv4.Float64(mem, "used"),
		}
	}

	if stream, ok := sys["stream"].(map[string]any); ok {
		dto.Stream = StreamSummaryDTO{
			UploadBps:   ikuaiv4.Int64(stream, "upload"),
			DownloadBps: ikuaiv4.Int64(stream, "download"),
			TotalUp:     ikuaiv4.Int64(stream, "total_up"),
			TotalDown:   ikuaiv4.Int64(stream, "total_down"),
		}
		dto.Connections = ConnectionSummaryDTO{
			Total: ikuaiv4.Int(stream, "connect_num"),
			TCP:   ikuaiv4.Int(stream, "tcp_connect_num"),
			UDP:   ikuaiv4.Int(stream, "udp_connect_num"),
			ICMP:  ikuaiv4.Int(stream, "icmp_connect_num"),
		}
	}

	if users, ok := sys["online_user"].(map[string]any); ok {
		dto.OnlineUsers = OnlineUserSummaryDTO{
			Total:    ikuaiv4.Int(users, "count"),
			Wired:    ikuaiv4.Int(users, "count_wired"),
			Wireless: ikuaiv4.Int(users, "count_wireless"),
			WiFi2G:   ikuaiv4.Int(users, "count_2g"),
			WiFi5G:   ikuaiv4.Int(users, "count_5g"),
		}
	}

	// 2. /monitoring/interfaces-status -> iface_check + iface_stream
	if statusErr == nil {
		if statusAny, decErr := ikuaiv4.DecodeAny(statusRaw); decErr == nil {
			if statusMap, ok := statusAny.(map[string]any); ok {
				if checks, ok := statusMap["iface_check"].([]any); ok {
					for _, item := range checks {
						if row, ok := item.(map[string]any); ok {
							dto.WanChecks = append(dto.WanChecks, WanCheckDTO{
								ID:              ikuaiv4.Int(row, "id"),
								Interface:       ikuaiv4.Str(row, "interface"),
								ParentInterface: ikuaiv4.Str(row, "parent_interface"),
								IPAddr:          ikuaiv4.Str(row, "ip_addr"),
								Gateway:         ikuaiv4.Str(row, "gateway"),
								Internet:        ikuaiv4.Str(row, "internet"),
								Result:          ikuaiv4.Str(row, "result"),
								ErrMsg:          ikuaiv4.Str(row, "errmsg"),
								AutoSwitch:      ikuaiv4.Str(row, "auto_switch"),
								UpdateTime:      ikuaiv4.Str(row, "updatetime"),
								Comment:         ikuaiv4.Str(row, "comment"),
							})
						}
					}
				}
				if streams, ok := statusMap["iface_stream"].([]any); ok {
					for _, item := range streams {
						if row, ok := item.(map[string]any); ok {
							dto.InterfaceTraffic = append(dto.InterfaceTraffic, InterfaceTrafficDTO{
								Interface:   ikuaiv4.Str(row, "interface"),
								IPAddr:      ikuaiv4.Str(row, "ip_addr"),
								Comment:     ikuaiv4.Str(row, "comment"),
								UploadBps:   ikuaiv4.Int64(row, "upload"),
								DownloadBps: ikuaiv4.Int64(row, "download"),
								TotalUp:     ikuaiv4.Int64(row, "total_up"),
								TotalDown:   ikuaiv4.Int64(row, "total_down"),
								Connections: ikuaiv4.Int(row, "connect_num"),
								UpDropped:   ikuaiv4.Int64(row, "updropped"),
								DownDropped: ikuaiv4.Int64(row, "downdropped"),
								UpPacked:    ikuaiv4.Int64(row, "uppacked"),
								DownPacked:  ikuaiv4.Int64(row, "downpacked"),
							})
						}
					}
				}
			}
		}
	}

	// 3. /monitoring/interfaces-traffic -> wans_stat_history
	if trafficErr == nil {
		rows, _ := ikuaiv4.ExtractCollection(trafficRaw)
		latestRTT := make(map[string]WanHistoryPointDTO)
		// Keep last 180 samples max so payload stays nimble
		start := 0
		if len(rows) > 180 {
			start = len(rows) - 180
		}
		for _, row := range rows[start:] {
			pt := WanHistoryPointDTO{
				Timestamp:       ikuaiv4.Int64(row, "timestamp"),
				Interface:       ikuaiv4.Str(row, "interface"),
				AvgUploadKBps:   ikuaiv4.Float64(row, "avg_upload"),
				AvgDownloadKBps: ikuaiv4.Float64(row, "avg_download"),
				MaxUploadKBps:   ikuaiv4.Float64(row, "max_upload"),
				MaxDownloadKBps: ikuaiv4.Float64(row, "max_download"),
				AvgRTTMs:        ikuaiv4.Float64(row, "avg_rtt"),
				MaxRTTMs:        ikuaiv4.Float64(row, "max_rtt"),
				MinRTTMs:        ikuaiv4.Float64(row, "min_rtt"),
				DropRate:        ikuaiv4.Float64(row, "drop_rate"),
			}
			dto.WanHistory = append(dto.WanHistory, pt)
			latestRTT[pt.Interface] = pt
		}
		for i := range dto.WanChecks {
			if pt, ok := latestRTT[dto.WanChecks[i].Interface]; ok {
				dto.WanChecks[i].AvgRTTMs = pt.AvgRTTMs
				dto.WanChecks[i].DropRate = pt.DropRate
			}
		}
	}

	// 4. /interfaces/physical -> ether_info map-of-maps
	if phyErr == nil {
		rows, _ := ikuaiv4.ExtractCollection(phyRaw)
		for _, row := range rows {
			dto.PhysicalPorts = append(dto.PhysicalPorts, PhysicalPortDTO{
				Name:      ikuaiv4.Str(row, "name"),
				Interface: ikuaiv4.Str(row, "interface"),
				Link:      ikuaiv4.Int(row, "link"),
				Speed:     ikuaiv4.Int(row, "speed"),
				Duplex:    ikuaiv4.Int(row, "duplex"),
				MAC:       ikuaiv4.Str(row, "mac"),
				Type:      ikuaiv4.Str(row, "type"),
				Driver:    ikuaiv4.Str(row, "driver"),
				Model:     ikuaiv4.Str(row, "model"),
				Lock:      ikuaiv4.Int(row, "lock"),
			})
		}
		sort.Slice(dto.PhysicalPorts, func(i, j int) bool {
			return dto.PhysicalPorts[i].Name < dto.PhysicalPorts[j].Name
		})
	}

	return dto, nil
}

// GetLoadHistory queries /monitoring/{cpu|memory|disk|cputemp|terminals|connections|network}.
func GetLoadHistory(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client, metric, timeRange, aggregate string) (*LoadHistoryDTO, error) {
	metric = strings.TrimSpace(metric)
	if metric == "" {
		metric = "cpu"
	}
	allowedMetrics := map[string]bool{
		"cpu": true, "memory": true, "disk": true, "cputemp": true,
		"terminals": true, "connections": true, "network": true,
	}
	if !allowedMetrics[metric] {
		return nil, fmt.Errorf("不支持的监控指标: %s", metric)
	}

	if timeRange == "" {
		timeRange = "hour"
	}
	var windowSeconds int64 = 3600
	switch timeRange {
	case "hour":
		windowSeconds = 3600
	case "day":
		windowSeconds = 86400
	case "week":
		windowSeconds = 604800
	case "month":
		windowSeconds = 2592000
	default:
		timeRange = "hour"
	}

	if aggregate != "max" {
		aggregate = "avg"
	}

	if router.Mock || client == nil {
		return mockLoadHistory(router, metric, timeRange, aggregate), nil
	}

	end := time.Now().Unix()
	start := end - windowSeconds
	params := map[string]string{
		"datetype":   timeRange,
		"start_time": strconv.FormatInt(start, 10),
		"end_time":   strconv.FormatInt(end, 10),
		"math":       aggregate,
	}
	raw, err := client.Get(ctx, "/monitoring/"+metric, params)
	if err != nil {
		return nil, err
	}

	rows, _ := ikuaiv4.ExtractCollection(raw)
	decoded, _ := ikuaiv4.DecodeAny(raw)
	return &LoadHistoryDTO{
		Metric:    metric,
		TimeRange: timeRange,
		Aggregate: aggregate,
		Series:    rows,
		Raw:       decoded,
	}, nil
}

// ToLegacyInterfaceData maps OverviewDTO to the legacy /api/v1/monitor/interface response shape.
func ToLegacyInterfaceData(ov *OverviewDTO) *LegacyInterfaceDataResult {
	res := &LegacyInterfaceDataResult{
		Summary: LegacySummaryInfo{
			UploadSpeed:      ov.Stream.UploadBps,
			DownloadSpeed:    ov.Stream.DownloadBps,
			TotalConnections: ov.Connections.Total,
			OnlineUsers:      ov.OnlineUsers.Total,
		},
		WanStatus:      make([]LegacyWanStatus, 0, len(ov.WanChecks)),
		TrafficDetails: make([]LegacyTrafficDetail, 0, len(ov.InterfaceTraffic)),
	}
	for _, w := range ov.WanChecks {
		res.WanStatus = append(res.WanStatus, LegacyWanStatus{
			Name:    w.Interface,
			IP:      w.IPAddr,
			Proto:   w.Internet,
			Status:  w.Result,
			Comment: w.Comment,
		})
	}
	for _, t := range ov.InterfaceTraffic {
		res.TrafficDetails = append(res.TrafficDetails, LegacyTrafficDetail{
			Name:          t.Interface,
			IP:            t.IPAddr,
			UploadSpeed:   t.UploadBps,
			DownloadSpeed: t.DownloadBps,
			TotalUp:       t.TotalUp,
			TotalDown:     t.TotalDown,
			Connections:   t.Connections,
			Comment:       t.Comment,
		})
	}
	return res
}

var _ = json.RawMessage{}

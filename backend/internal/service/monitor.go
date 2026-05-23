package service

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"ikuai-dashboard/backend/internal/config"
	"ikuai-dashboard/backend/internal/logger"

	ikuaiapi "github.com/zy84338719/ikuai-api"
	ikuaisvc "github.com/zy84338719/ikuai-api/service"
)

// ───────────── 对外暴露的 DTO 类型 ─────────────

// ClientDTO 按 MAC 聚合去重后（包含双栈 IP）的客户端记录
type ClientDTO struct {
	MAC           string   `json:"mac"`
	Hostname      string   `json:"hostname"`
	IPs           []string `json:"ips"`
	UploadSpeed   int64    `json:"upload_speed"`   // 字节/秒
	DownloadSpeed int64    `json:"download_speed"` // 字节/秒
	TotalUp       int64    `json:"total_up"`       // 字节
	TotalDown     int64    `json:"total_down"`     // 字节
	Connections   int      `json:"connections"`
	Comment       string   `json:"comment"`
	ClientType    string   `json:"client_type"`
	Uptime        string   `json:"uptime"`
}

// SummaryInfo 首页大卡片：总体流量速率与连接数
type SummaryInfo struct {
	UploadSpeed      int64 `json:"upload_speed"`      // 字节/秒
	DownloadSpeed    int64 `json:"download_speed"`    // 字节/秒
	TotalConnections int   `json:"total_connections"` // 总连接数
	OnlineUsers      int   `json:"online_users"`      // 在线终端数
}

// WanStatus WAN 口状态
type WanStatus struct {
	Name    string `json:"name"`
	IP      string `json:"ip"`
	Proto   string `json:"proto"`
	Status  string `json:"status"`
	Comment string `json:"comment"`
}

// TrafficDetail 每个物理接口的流量统计
type TrafficDetail struct {
	Name          string `json:"name"`
	IP            string `json:"ip"`
	UploadSpeed   int64  `json:"upload_speed"`
	DownloadSpeed int64  `json:"download_speed"`
	TotalUp       int64  `json:"total_up"`
	TotalDown     int64  `json:"total_down"`
	Connections   int    `json:"connections"`
	Comment       string `json:"comment"`
}

// InterfaceDataResult 首页合并后的完整响应
type InterfaceDataResult struct {
	Summary        SummaryInfo     `json:"summary"`
	WanStatus      []WanStatus     `json:"wan_status"`
	TrafficDetails []TrafficDetail `json:"traffic_details"`
}

// ───────────── 服务层 ─────────────

// MonitorService 监控服务层
type MonitorService struct {
	client          *ikuaiapi.Client
	v4Client        *ikuaiapi.V4Client
	api             ikuaisvc.APIClient
	router          config.RouterConfig
	mockMode        bool
	unconfigured    bool
	connectionError string
	mu              sync.Mutex
}

var GlobalMonitorService *MonitorService
var globalServiceMu sync.RWMutex
var ErrUnconfigured = errors.New("尚未配置爱快服务器")

// InitMonitorService 初始化监控服务，自动根据当前激活配置选择真实或模拟模式。
func InitMonitorService() {
	cfg := config.Snapshot()
	if cfg == nil {
		GlobalMonitorService = &MonitorService{unconfigured: true, connectionError: ErrUnconfigured.Error()}
		return
	}

	router, ok := cfg.ActiveRouter()
	if !ok {
		GlobalMonitorService = &MonitorService{unconfigured: true, connectionError: ErrUnconfigured.Error()}
		return
	}

	ReplaceMonitorService(NewMonitorService(router))
}

// NewMonitorService 基于单台爱快服务器配置创建服务实例。
func NewMonitorService(router config.RouterConfig) *MonitorService {
	svc := &MonitorService{
		router:   router,
		mockMode: router.Mock,
	}

	if router.Mock {
		logger.Log.Infof("✦ %s 使用高保真模拟模式", router.Name)
		return svc
	}

	version := router.Version
	if version == "" {
		version = config.RouterVersionV3
	}
	logger.Log.Infof("→ 正在连接爱快路由器 [%s/%s]: %s", router.Name, version, router.URL)

	if version == config.RouterVersionV4 {
		svc.v4Client = ikuaiapi.NewV4RESTClient(router.URL, router.Token)
		logger.Log.Infof("✓ 爱快路由器 [%s] 已启用 v4 Token 客户端", router.Name)
		return svc
	}

	client, err := ikuaiapi.NewV3ClientWithLogin(
		router.URL,
		router.Username,
		router.Password,
		ikuaiapi.WithInsecureSkipVerify(router.InsecureSkipVerify),
	)
	if err != nil {
		logger.Log.Warnf("⚠ 爱快连接失败，自动回退至模拟模式。原因: %v", err)
		svc.mockMode = true
		svc.connectionError = err.Error()
		return svc
	}

	svc.client = client
	svc.api = ikuaisvc.NewAPIClient(client)
	logger.Log.Infof("✓ 爱快路由器 [%s] 连接成功", router.Name)
	return svc
}

// ReplaceMonitorService 原子替换全局监控服务。
func ReplaceMonitorService(svc *MonitorService) {
	globalServiceMu.Lock()
	defer globalServiceMu.Unlock()
	GlobalMonitorService = svc
}

// CurrentMonitorService 返回当前监控服务。
func CurrentMonitorService() *MonitorService {
	globalServiceMu.RLock()
	defer globalServiceMu.RUnlock()
	return GlobalMonitorService
}

// ReloadMonitorService 根据当前全局配置重建监控服务。
func ReloadMonitorService() {
	cfg := config.Snapshot()
	if cfg == nil {
		ReplaceMonitorService(&MonitorService{unconfigured: true, connectionError: ErrUnconfigured.Error()})
		return
	}
	router, ok := cfg.ActiveRouter()
	if !ok {
		ReplaceMonitorService(&MonitorService{unconfigured: true, connectionError: ErrUnconfigured.Error()})
		return
	}
	ReplaceMonitorService(NewMonitorService(router))
}

// Status 返回当前服务状态。
func (s *MonitorService) Status() map[string]any {
	if s == nil {
		return map[string]any{"mode": "mock", "router_id": "", "router_name": "", "error": "服务未初始化"}
	}
	if s.unconfigured {
		return map[string]any{"mode": "unconfigured", "router_id": "", "router_name": "", "version": "", "error": s.connectionError}
	}
	mode := "real"
	if s.mockMode {
		mode = "mock"
	}
	return map[string]any{
		"mode":        mode,
		"router_id":   s.router.ID,
		"router_name": s.router.Name,
		"version":     s.router.Version,
		"error":       s.connectionError,
	}
}

// ───────────── 首页接口数据 ─────────────

// GetInterfaceData 获取首页看板所需的接口数据（摘要 + WAN 状态 + 流量详情）
func (s *MonitorService) GetInterfaceData(ctx context.Context) (*InterfaceDataResult, error) {
	if s == nil {
		return nil, errors.New("服务未初始化")
	}
	if s.unconfigured {
		return nil, ErrUnconfigured
	}
	if s.mockMode {
		return s.getMockInterfaceData(), nil
	}
	if s.v4Client != nil {
		return s.getV4InterfaceData(ctx)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var result InterfaceDataResult

	// 1. 首页摘要：上传/下载速率 + 总连接数（来自 homepage 接口）
	homepage, err := s.api.System().GetHomepage(ctx)
	if err != nil {
		logger.Log.Errorf("获取首页摘要数据失败: %v", err)
		// 降级处理：摘要失败不影响后面的接口数据
	} else {
		result.Summary = SummaryInfo{
			UploadSpeed:      int64(homepage.Stream.Upload),
			DownloadSpeed:    int64(homepage.Stream.Download),
			TotalConnections: homepage.Stream.ConnectNum,
			OnlineUsers:      homepage.OnlineUser.Count,
		}
	}

	// 2. 接口状态与流量：来自 monitor_iface 接口
	ifaces, err := s.api.Monitor().GetInterfaces(ctx)
	if err != nil {
		logger.Log.Errorf("获取接口流量数据失败: %v", err)
		return &result, nil
	}

	// WAN 接口状态
	for _, check := range ifaces.GetIFaceCheck() {
		result.WanStatus = append(result.WanStatus, WanStatus{
			Name:    check.Interface,
			IP:      check.IPAddr,
			Proto:   check.Internet,
			Status:  check.Result,
			Comment: check.Comment,
		})
	}

	// 各接口实时流量
	for _, stream := range ifaces.GetIFaceStream() {
		conns := 0
		if stream.ConnectNum != "" {
			if n, err := strconv.Atoi(stream.ConnectNum); err == nil {
				conns = n
			}
		}
		result.TrafficDetails = append(result.TrafficDetails, TrafficDetail{
			Name:          stream.Interface,
			IP:            stream.IPAddr,
			UploadSpeed:   int64(stream.Upload),
			DownloadSpeed: int64(stream.Download),
			TotalUp:       stream.TotalUp,
			TotalDown:     stream.TotalDown,
			Connections:   conns,
			Comment:       stream.Comment,
		})
	}

	return &result, nil
}

// ───────────── 局域网客户端列表 ─────────────

// GetLanClients 获取按 MAC 去重并双栈聚合后的局域网客户端列表
func (s *MonitorService) GetLanClients(ctx context.Context, filterComment string) ([]ClientDTO, error) {
	if s == nil {
		return nil, errors.New("服务未初始化")
	}
	if s.unconfigured {
		return nil, ErrUnconfigured
	}
	if s.mockMode {
		return s.getMockLanClients(filterComment), nil
	}
	if s.v4Client != nil {
		return s.getV4LanClients(ctx, filterComment)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 以 MAC（大写）为 key 的聚合 Map
	aggregatedMap := make(map[string]*ClientDTO)

	// 1. 获取 IPv4 客户端列表
	ipv4List, err := s.api.Monitor().GetLanIP(ctx)
	if err != nil {
		logger.Log.Errorf("获取 IPv4 客户端列表失败: %v", err)
		return nil, err
	}

	for _, item := range ipv4List {
		if item.Mac == "" {
			continue
		}
		macKey := strings.ToUpper(item.Mac)

		if entry, exists := aggregatedMap[macKey]; exists {
			// 追加 IP
			if !containsStr(entry.IPs, item.IPAddr) {
				entry.IPs = append(entry.IPs, item.IPAddr)
			}
			// 累加流量指标
			entry.UploadSpeed += int64(item.Upload)
			entry.DownloadSpeed += int64(item.Download)
			entry.TotalUp += item.TotalUp
			entry.TotalDown += item.TotalDown
			entry.Connections += item.ConnectNum
			if item.Comment != "" && !strings.Contains(entry.Comment, item.Comment) {
				if entry.Comment == "" {
					entry.Comment = item.Comment
				} else {
					entry.Comment += ", " + item.Comment
				}
			}
		} else {
			aggregatedMap[macKey] = &ClientDTO{
				MAC:           item.Mac,
				Hostname:      item.Hostname,
				IPs:           []string{item.IPAddr},
				UploadSpeed:   int64(item.Upload),
				DownloadSpeed: int64(item.Download),
				TotalUp:       item.TotalUp,
				TotalDown:     item.TotalDown,
				Connections:   item.ConnectNum,
				Comment:       item.Comment,
				ClientType:    item.ClientType,
				Uptime:        item.Uptime,
			}
		}
	}

	// 2. 获取 IPv6 客户端，合并同一 MAC 的 IPv6 地址
	ipv6List, err := s.api.Monitor().GetLanIPv6(ctx)
	if err != nil {
		logger.Log.Warnf("获取 IPv6 客户端列表失败（不影响 IPv4 结果）: %v", err)
	} else {
		for _, item := range ipv6List {
			if item.Mac == "" {
				continue
			}
			macKey := strings.ToUpper(item.Mac)
			if entry, exists := aggregatedMap[macKey]; exists {
				if !containsStr(entry.IPs, item.IPAddr) {
					entry.IPs = append(entry.IPs, item.IPAddr)
				}
			}
		}
	}

	// 3. 将 Map 转换为切片并做备注过滤
	result := make([]ClientDTO, 0, len(aggregatedMap))
	for _, dto := range aggregatedMap {
		// 备注模糊搜索（大小写无关）
		if filterComment != "" {
			haystack := strings.ToLower(dto.Comment + " " + dto.Hostname)
			if !strings.Contains(haystack, strings.ToLower(filterComment)) {
				continue
			}
		}
		result = append(result, *dto)
	}

	return result, nil
}

// ─────────── 高保真 Mock 数据源 ───────────

func (s *MonitorService) getMockInterfaceData() *InterfaceDataResult {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	upBase := int64(8 * 1024 * 1024) // ~8 MB/s 基础值
	downBase := int64(300 * 1024)    // ~300 KB/s 基础值
	upSpeed := upBase + rng.Int63n(2*1024*1024) - 1*1024*1024
	downSpeed := downBase + rng.Int63n(150*1024) - 75*1024
	if upSpeed < 512 {
		upSpeed = 512
	}
	if downSpeed < 512 {
		downSpeed = 512
	}

	return &InterfaceDataResult{
		Summary: SummaryInfo{
			UploadSpeed:      upSpeed,
			DownloadSpeed:    downSpeed,
			TotalConnections: 950 + rng.Intn(30) - 15,
			OnlineUsers:      42 + rng.Intn(10) - 5,
		},
		WanStatus: []WanStatus{
			{Name: "wan1", IP: "123.118.2.47", Proto: "PPPOE", Status: "success", Comment: "主力电信"},
			{Name: "wan2", IP: "123.118.3.80", Proto: "PPPOE", Status: "success", Comment: "备选联通"},
			{Name: "wan3", IP: "192.168.1.3", Proto: "DHCP", Status: "success", Comment: "旁路主路由"},
		},
		TrafficDetails: []TrafficDetail{
			{Name: "lan1", IP: "192.168.50.1", UploadSpeed: upSpeed * 9 / 10, DownloadSpeed: downSpeed * 9 / 10, TotalUp: 3375253815296 + rng.Int63n(1e6), TotalDown: 2253965529088 + rng.Int63n(2e6), Connections: 0, Comment: "内网局域网接口"},
			{Name: "wan1", IP: "123.118.2.47", UploadSpeed: upSpeed * 7 / 10, DownloadSpeed: downSpeed * 8 / 10, TotalUp: 1913149751296 + rng.Int63n(5e5), TotalDown: 1517329547264 + rng.Int63n(1e6), Connections: 497 + rng.Intn(10) - 5, Comment: "电信千兆宽带"},
			{Name: "wan2", IP: "123.118.3.80", UploadSpeed: upSpeed * 3 / 10, DownloadSpeed: downSpeed * 2 / 10, TotalUp: 1373149751296 + rng.Int63n(2e5), TotalDown: 647529547264 + rng.Int63n(4e5), Connections: 458 + rng.Intn(10) - 5, Comment: "联通备用宽带"},
			{Name: "wan3", IP: "192.168.1.3", UploadSpeed: 0, DownloadSpeed: 0, TotalUp: 5040000000, TotalDown: 7020000000, Connections: 0, Comment: "DHCP 备用链路"},
		},
	}
}

func (s *MonitorService) getMockLanClients(filterComment string) []ClientDTO {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	type mockDevice struct {
		mac        string
		hostname   string
		ips        []string
		comment    string
		clientType string
		uptime     string
		up         int64
		down       int64
		totUp      int64
		totDown    int64
		conns      int
	}

	devices := []mockDevice{
		{mac: "00:22:1F:97:F2:CD", hostname: "imac-pro-01", ips: []string{"10.10.10.104", "2408:8207:1920:6d52:014c:b5b4:c506:aafe"}, comment: "研发中心-iMac Pro", clientType: "PC", uptime: "8h 32m", up: 2887, down: 1925, totUp: 42144825344, totDown: 5282922496, conns: 18},
		{mac: "00:70:FA:3D:30:10", hostname: "CEO-MacBook", ips: []string{"10.10.10.103", "2408:8207:1920:89e2:5da4:a1da:cb1a:5876"}, comment: "李总的 MacBook Pro", clientType: "PC", uptime: "2h 15m", up: 39000, down: 39000, totUp: 33704825344, totDown: 4782922496, conns: 15},
		{mac: "00:9A:13:24:C8:A3", hostname: "meeting-tablet", ips: []string{"10.10.10.102", "2408:8207:1920:6d53:dc4f:d2cf:046d:5f13"}, comment: "公司智能会议平板", clientType: "Pad", uptime: "24h 0m", up: 0, down: 0, totUp: 216244825344, totDown: 71732954726, conns: 2},
		{mac: "00:5B:E5:28:A6:C0", hostname: "nas-local", ips: []string{"10.10.10.101", "2408:8207:1920:89e3:e1b9:0cfe:ffdb:f40d"}, comment: "极客专属-本地Nas", clientType: "NAS", uptime: "168h 0m", up: 20000, down: 20000, totUp: 211044825344, totDown: 32429224962, conns: 2},
		{mac: "DA:8A:9F:53:43:2D", hostname: "guest-ipad", ips: []string{"192.168.50.254"}, comment: "前台访客 iPad", clientType: "Pad", uptime: "0h 45m", up: 0, down: 0, totUp: 200000, totDown: 1500000, conns: 4},
		{mac: "EA:6C:25:F5:D7:43", hostname: "printer-finance", ips: []string{"192.168.50.223"}, comment: "财务部专用打印机", clientType: "OTHER", uptime: "72h 0m", up: 0, down: 0, totUp: 32023901, totDown: 644810232, conns: 3},
		{mac: "C6:B0:29:0D:27:36", hostname: "gate-machine", ips: []string{"192.168.50.222"}, comment: "智能前台闸机", clientType: "OTHER", uptime: "720h 0m", up: 0, down: 0, totUp: 252682910, totDown: 17321092, conns: 2},
		{mac: "8E:F6:C5:99:8A:50", hostname: "synology-nas", ips: []string{"192.168.50.100", "2408:8207:1920:89e1:0c21:4493:7ead:5c49"}, comment: "测试部-群晖NAS", clientType: "NAS", uptime: "360h 0m", up: 8765, down: 6882, totUp: 513192809122, totDown: 44921029100, conns: 76},
		{mac: "78:DF:72:9F:D1:1A", hostname: "iphone15-mkt", ips: []string{"192.168.50.68"}, comment: "市场部-iPhone 15 Pro", clientType: "Phone", uptime: "1h 20m", up: 0, down: 0, totUp: 6451290, totDown: 4012920, conns: 11},
		{mac: "BC:24:11:AE:93:20", hostname: "lenovo-hr", ips: []string{"192.168.50.63"}, comment: "HR-联想办公本", clientType: "PC", uptime: "6h 10m", up: 53248, down: 67584, totUp: 463581029, totDown: 277481092, conns: 18},
		{mac: "54:48:E6:95:66:9F", hostname: "sensor-temp", ips: []string{"192.168.50.61"}, comment: "智能温湿度传感器", clientType: "OTHER", uptime: "2160h 0m", up: 0, down: 0, totUp: 730000, totDown: 660000, conns: 1},
	}

	var dtos []ClientDTO
	for _, d := range devices {
		if filterComment != "" {
			hay := strings.ToLower(d.comment + " " + d.hostname)
			if !strings.Contains(hay, strings.ToLower(filterComment)) {
				continue
			}
		}

		dynamicUp := d.up
		dynamicDown := d.down
		if d.up > 0 {
			dynamicUp = d.up + rng.Int63n(500) - 250
			if dynamicUp < 50 {
				dynamicUp = 50
			}
		}
		if d.down > 0 {
			dynamicDown = d.down + rng.Int63n(500) - 250
			if dynamicDown < 50 {
				dynamicDown = 50
			}
		}

		dtos = append(dtos, ClientDTO{
			MAC:           d.mac,
			Hostname:      d.hostname,
			IPs:           d.ips,
			UploadSpeed:   dynamicUp,
			DownloadSpeed: dynamicDown,
			TotalUp:       d.totUp + rng.Int63n(5000),
			TotalDown:     d.totDown + rng.Int63n(5000),
			Connections:   d.conns + rng.Intn(4) - 2,
			Comment:       d.comment,
			ClientType:    d.clientType,
			Uptime:        d.uptime,
		})
	}

	return dtos
}

// ─────────── 工具函数 ───────────

func containsStr(arr []string, target string) bool {
	for _, s := range arr {
		if s == target {
			return true
		}
	}
	return false
}

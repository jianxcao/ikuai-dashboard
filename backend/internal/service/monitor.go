package service

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"ikuai4-backend/backend/internal/config"
	"ikuai4-backend/backend/internal/logger"

	ikuaiapi "github.com/zy84338719/ikuai-api"
)

// ClientRaw iKuai SDK 返回的原始局域网客户端信息
type ClientRaw struct {
	MAC           string `json:"mac"`
	IP            string `json:"ip"`
	UploadSpeed   int64  `json:"upload_speed"`  // 字节/秒
	DownloadSpeed int64  `json:"download_speed"` // 字节/秒
	TotalUp       int64  `json:"total_up"`       // 字节数
	TotalDown     int64  `json:"total_down"`     // 字节数
	Connections   int    `json:"connections"`
	Comment       string `json:"comment"`
}

// ClientDTO 聚合去重后，包含双栈 IP 且按 MAC 唯一的客户端数据
type ClientDTO struct {
	MAC           string   `json:"mac"`
	IPs           []string `json:"ips"`
	UploadSpeed   int64    `json:"upload_speed"`
	DownloadSpeed int64    `json:"download_speed"`
	TotalUp       int64    `json:"total_up"`
	TotalDown     int64    `json:"total_down"`
	Connections   int      `json:"connections"`
	Comment       string   `json:"comment"`
}

// SummaryInfo 首页大卡片数据
type SummaryInfo struct {
	UploadSpeed      int64 `json:"upload_speed"`      // 字节/秒
	DownloadSpeed     int64 `json:"download_speed"`     // 字节/秒
	TotalConnections int   `json:"total_connections"`
}

// WanStatus WAN 接口状态
type WanStatus struct {
	Name    string `json:"name"`
	IP      string `json:"ip"`
	Proto   string `json:"proto"`
	Status  string `json:"status"`
	Comment string `json:"comment"`
}

// TrafficDetail 接口流量统计详情
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

// InterfaceDataResult 首页合并后的最终返回数据
type InterfaceDataResult struct {
	Summary        SummaryInfo     `json:"summary"`
	WanStatus      []WanStatus     `json:"wan_status"`
	TrafficDetails []TrafficDetail `json:"traffic_details"`
}

// MonitorService 监控服务层
type MonitorService struct {
	ikuaiClient *ikuaiapi.Client
	mu          sync.Mutex
}

var GlobalMonitorService *MonitorService

// InitMonitorService 初始化监控服务
func InitMonitorService() {
	GlobalMonitorService = &MonitorService{}
	if !config.GlobalConfig.MockMode {
		logger.Log.Infof("正在尝试连接爱快路由器: %s", config.GlobalConfig.IKuaiURL)
		var client *ikuaiapi.Client
		var err error

		if config.GlobalConfig.APIToken != "" {
			// 如果提供了 Token 鉴权 (V4)
			client, err = ikuaiapi.NewClientWithToken(config.GlobalConfig.IKuaiURL, config.GlobalConfig.APIToken)
		} else {
			// 使用常规账号密码认证 (V3)
			client, err = ikuaiapi.NewClientWithLogin(config.GlobalConfig.IKuaiURL, config.GlobalConfig.Username, config.GlobalConfig.Password)
		}

		if err != nil {
			logger.Log.Warnf("爱快连接建立失败，服务自动回退至【高仿真模拟模式】。失败原因: %v", err)
			config.GlobalConfig.MockMode = true
		} else {
			GlobalMonitorService.ikuaiClient = client
			logger.Log.Info("爱快物理连接已成功建立！")
		}
	} else {
		logger.Log.Info("已开启高保真模拟模式（Mock Mode = true）")
	}
}

// GetInterfaceData 获取首页看板接口数据
func (s *MonitorService) GetInterfaceData(ctx context.Context) (*InterfaceDataResult, error) {
	if config.GlobalConfig.MockMode {
		return s.getMockInterfaceData(), nil
	}

	// 真实模式：并发从爱快抓取实时流量详情
	s.mu.Lock()
	defer s.mu.Unlock()

	var result InterfaceDataResult

	// 1. 获取 wan 口状态 (func_name: "wan", action: "show")
	var rawWan []map[string]interface{}
	err := s.ikuaiClient.Monitor().Show(ctx, "wan", nil, &rawWan)
	if err != nil {
		logger.Log.Errorf("获取 WAN 口状态失败: %v", err)
		return nil, err
	}

	for _, w := range rawWan {
		ip, _ := w["ip_addr"].(string)
		name, _ := w["interface"].(string)
		proto, _ := w["internet_proto"].(string)
		status, _ := w["status"].(string)
		comment, _ := w["comment"].(string)

		result.WanStatus = append(result.WanStatus, WanStatus{
			Name:    name,
			IP:      ip,
			Proto:   proto,
			Status:  status,
			Comment: comment,
		})
	}

	// 2. 获取实时流量统计 (func_name: "monitor_iface", action: "show")
	var rawIface []map[string]interface{}
	err = s.ikuaiClient.Monitor().Show(ctx, "monitor_iface", nil, &rawIface)
	if err != nil {
		logger.Log.Errorf("获取接口实时流量失败: %v", err)
		return nil, err
	}

	var totalUpSpeed int64
	var totalDownSpeed int64
	var totalConns int

	for _, f := range rawIface {
		name, _ := f["interface"].(string)
		ip, _ := f["ip_addr"].(string)
		upSpeed, _ := getInt64(f["upload_speed"])
		downSpeed, _ := getInt64(f["download_speed"])
		totalUp, _ := getInt64(f["total_up"])
		totalDown, _ := getInt64(f["total_down"])
		conns, _ := getInt(f["connections"])
		comment, _ := f["comment"].(string)

		result.TrafficDetails = append(result.TrafficDetails, TrafficDetail{
			Name:          name,
			IP:            ip,
			UploadSpeed:   upSpeed,
			DownloadSpeed: downSpeed,
			TotalUp:       totalUp,
			TotalDown:     totalDown,
			Connections:   conns,
			Comment:       comment,
		})

		// 汇总大卡片速率与连接数
		if strings.HasPrefix(strings.ToLower(name), "wan") {
			totalUpSpeed += upSpeed
			totalDownSpeed += downSpeed
			totalConns += conns
		}
	}

	result.Summary = SummaryInfo{
		UploadSpeed:      totalUpSpeed,
		DownloadSpeed:     totalDownSpeed,
		TotalConnections: totalConns,
	}

	return &result, nil
}

// GetLanClients 获取按 MAC 去重并智能聚合双栈 IP 后的局域网客户端列表
func (s *MonitorService) GetLanClients(ctx context.Context, filterComment string) ([]ClientDTO, error) {
	if config.GlobalConfig.MockMode {
		return s.getMockLanClients(filterComment), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var rawList []ClientRaw
	// 对应爱快系统底层：func_name = "monitor_lan", action = "show"
	err := s.ikuaiClient.Monitor().Show(ctx, "monitor_lan", nil, &rawList)
	if err != nil {
		logger.Log.Errorf("拉取爱快局域网客户端列表失败: %v", err)
		return nil, err
	}

	// 执行高保真去重与双栈合并算法
	aggregatedMap := make(map[string]*ClientDTO)
	for _, raw := range rawList {
		// 1. 备注模糊搜索过滤（大小写无关）
		if filterComment != "" && !strings.Contains(strings.ToLower(raw.Comment), strings.ToLower(filterComment)) {
			continue
		}

		macUpper := strings.ToUpper(raw.MAC)
		if entry, exists := aggregatedMap[macUpper]; exists {
			// 如果此 MAC 已存在，聚合 IP 记录且叠加流量指标
			if !contains(entry.IPs, raw.IP) {
				entry.IPs = append(entry.IPs, raw.IP)
			}
			entry.UploadSpeed += raw.UploadSpeed
			entry.DownloadSpeed += raw.DownloadSpeed
			entry.TotalUp += raw.TotalUp
			entry.TotalDown += raw.TotalDown
			entry.Connections += raw.Connections

			if raw.Comment != "" && !strings.Contains(entry.Comment, raw.Comment) {
				if entry.Comment == "" {
					entry.Comment = raw.Comment
				} else {
					entry.Comment = entry.Comment + ", " + raw.Comment
				}
			}
		} else {
			// 全新 MAC 纪录，初始化 DTO
			aggregatedMap[macUpper] = &ClientDTO{
				MAC:           raw.MAC,
				IPs:           []string{raw.IP},
				UploadSpeed:   raw.UploadSpeed,
				DownloadSpeed: raw.DownloadSpeed,
				TotalUp:       raw.TotalUp,
				TotalDown:     raw.TotalDown,
				Connections:   raw.Connections,
				Comment:       raw.Comment,
			}
		}
	}

	result := make([]ClientDTO, 0, len(aggregatedMap))
	for _, dto := range aggregatedMap {
		result = append(result, *dto)
	}

	return result, nil
}

// ==========================================
//           以下为高保真仿真数据源 (Mock Mode)
// ==========================================

func (s *MonitorService) getMockInterfaceData() *InterfaceDataResult {
	rand.Seed(time.Now().UnixNano())

	// 动态浮动大卡片上传/下载速率 (高逼真随机起伏)
	upBase := int64(8 * 1024 * 1024)   // 8MB 基础
	downBase := int64(300 * 1024)      // 300KB 基础
	upSpeed := upBase + rand.Int63n(800*1024) - 400*1024
	downSpeed := downBase + rand.Int63n(150*1024) - 75*1024
	if upSpeed < 0 {
		upSpeed = 1024
	}
	if downSpeed < 0 {
		downSpeed = 1024
	}

	return &InterfaceDataResult{
		Summary: SummaryInfo{
			UploadSpeed:      upSpeed,
			DownloadSpeed:     downSpeed,
			TotalConnections: 950 + rand.Intn(30) - 15,
		},
		WanStatus: []WanStatus{
			{Name: "wan1", IP: "123.118.2.47", Proto: "PPPOE", Status: "success", Comment: "主力电信"},
			{Name: "wan2", IP: "123.118.3.80", Proto: "PPPOE", Status: "success", Comment: "备选联通"},
			{Name: "wan3", IP: "192.168.1.3", Proto: "DHCP", Status: "success", Comment: "旁路主路由"},
		},
		TrafficDetails: []TrafficDetail{
			{Name: "lan1", IP: "192.168.50.1", UploadSpeed: upSpeed - 120000, DownloadSpeed: downSpeed - 50000, TotalUp: 3375253815296 + rand.Int63n(1000000), TotalDown: 2253965529088 + rand.Int63n(2000000), Connections: 0, Comment: "内网局域网接口"},
			{Name: "wan1", IP: "123.118.2.47", UploadSpeed: upSpeed * 7 / 10, DownloadSpeed: downSpeed * 8 / 10, TotalUp: 1913149751296 + rand.Int63n(500000), TotalDown: 1517329547264 + rand.Int63n(1000000), Connections: 497 + rand.Intn(10) - 5, Comment: "电信千兆宽带"},
			{Name: "wan2", IP: "123.118.3.80", UploadSpeed: upSpeed * 3 / 10, DownloadSpeed: downSpeed * 2 / 10, TotalUp: 1373149751296 + rand.Int63n(200000), TotalDown: 647529547264 + rand.Int63n(400000), Connections: 458 + rand.Intn(10) - 5, Comment: "联通备用宽带"},
			{Name: "wan3", IP: "192.168.1.3", UploadSpeed: 0, DownloadSpeed: 0, TotalUp: 5040000000 + rand.Int63n(10000), TotalDown: 7020000000 + rand.Int63n(20000), Connections: 0, Comment: "DHCP 备用链路"},
		},
	}
}

func (s *MonitorService) getMockLanClients(filterComment string) []ClientDTO {
	rand.Seed(time.Now().UnixNano())

	rawDevices := []struct {
		mac     string
		ips     []string
		comment string
		up      int64
		down    int64
		totUp   int64
		totDown int64
		conns   int
	}{
		{
			mac: "00:22:1F:97:F2:CD",
			ips: []string{"10.10.10.104", "2408:8207:1920:6d52:014c:b5b4:c506:aafe"},
			comment: "研发中心-iMac Pro",
			up: 2887, down: 1925, totUp: 42144825344, totDown: 5282922496, conns: 18,
		},
		{
			mac: "00:70:FA:3D:30:10",
			ips: []string{"10.10.10.103", "2408:8207:1920:89e2:5da4:a1da:cb1a:5876"},
			comment: "李总的 MacBook Pro",
			up: 39000, down: 39000, totUp: 33704825344, totDown: 4782922496, conns: 15,
		},
		{
			mac: "00:9A:13:24:C8:A3",
			ips: []string{"10.10.10.102", "2408:8207:1920:6d53:dc4f:d2cf:046d:5f13"},
			comment: "公司智能会议平板",
			up: 0, down: 0, totUp: 216244825344, totDown: 71732954726, conns: 2,
		},
		{
			mac: "00:5B:E5:28:A6:C0",
			ips: []string{"10.10.10.101", "2408:8207:1920:89e3:e1b9:0cfe:ffdb:f40d"},
			comment: "极客专属-本地Nas",
			up: 20000, down: 20000, totUp: 211044825344, totDown: 32429224962, conns: 2,
		},
		{
			mac: "DA:8A:9F:53:43:2D",
			ips: []string{"192.168.50.254"},
			comment: "前台访客 iPad",
			up: 0, down: 0, totUp: 200000, totDown: 1500000, conns: 4,
		},
		{
			mac: "EA:6C:25:F5:D7:43",
			ips: []string{"192.168.50.223"},
			comment: "财务部专用打印机",
			up: 0, down: 0, totUp: 32023901, totDown: 644810232, conns: 3,
		},
		{
			mac: "C6:B0:29:0D:27:36",
			ips: []string{"192.168.50.222"},
			comment: "智能前台闸机",
			up: 0, down: 0, totUp: 252682910, totDown: 17321092, conns: 2,
		},
		{
			mac: "8E:F6:C5:99:8A:50",
			ips: []string{"192.168.50.100", "2408:8207:1920:89e1:0c21:4493:7ead:5c49"},
			comment: "测试部-群晖NAS",
			up: 8765, down: 6882, totUp: 513192809122, totDown: 44921029100, conns: 76,
		},
		{
			mac: "78:DF:72:9F:D1:1A",
			ips: []string{"192.168.50.68"},
			comment: "市场部-iPhone 15 Pro",
			up: 0, down: 0, totUp: 6451290, totDown: 4012920, conns: 11,
		},
		{
			mac: "BC:24:11:AE:93:20",
			ips: []string{"192.168.50.63"},
			comment: "HR-联想办公本",
			up: 53248, down: 67584, totUp: 463581029, totDown: 277481092, conns: 18,
		},
		{
			mac: "54:48:E6:95:66:9F",
			ips: []string{"192.168.50.61"},
			comment: "智能温湿度传感器",
			up: 0, down: 0, totUp: 730000, totDown: 660000, conns: 1,
		},
	}

	var dtos []ClientDTO
	for _, d := range rawDevices {
		// 模糊搜索过滤
		if filterComment != "" && !strings.Contains(strings.ToLower(d.comment), strings.ToLower(filterComment)) {
			continue
		}

		// 动态产生少量即时速率波动
		var dynamicUp int64 = d.up
		var dynamicDown int64 = d.down
		if d.up > 0 {
			dynamicUp = d.up + rand.Int63n(500) - 250
			if dynamicUp < 0 {
				dynamicUp = 100
			}
		}
		if d.down > 0 {
			dynamicDown = d.down + rand.Int63n(500) - 250
			if dynamicDown < 0 {
				dynamicDown = 100
			}
		}

		dtos = append(dtos, ClientDTO{
			MAC:           d.mac,
			IPs:           d.ips,
			UploadSpeed:   dynamicUp,
			DownloadSpeed: dynamicDown,
			TotalUp:       d.totUp + rand.Int63n(5000),
			TotalDown:     d.totDown + rand.Int63n(5000),
			Connections:   d.conns + rand.Intn(4) - 2,
			Comment:       d.comment,
		})
	}

	return dtos
}

// 辅助转换方法
func getInt64(val interface{}) (int64, error) {
	if val == nil {
		return 0, nil
	}
	switch v := val.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case float64:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("invalid int64 type")
	}
}

func getInt(val interface{}) (int, error) {
	if val == nil {
		return 0, nil
	}
	switch v := val.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("invalid int type")
	}
}

func contains(arr []string, target string) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}
	return false
}

package v4service

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"ikuai-dashboard/backend/internal/config"
)

func isLabMock(router config.RouterConfig) bool {
	return strings.Contains(strings.ToLower(router.ID), "lab")
}

func mockOverview(router config.RouterConfig) *OverviewDTO {
	now := time.Now().Unix()
	rng := rand.New(rand.NewSource(now / 2))
	lab := isLabMock(router)

	upBase := int64(6_420_000)
	downBase := int64(28_640_000)
	hostname := "iKuai-HQ-Core"
	lanIP := "192.168.9.1"
	arch := "x86/x64"
	ver := "4.0.314-beta x64 Build202610131658"
	uptime := int64(3158145)
	conns := 1246
	users := 42
	cpuAvg := 8.4
	memPct := 39.0
	if lab {
		upBase = 2_450_000
		downBase = 9_820_000
		hostname = "iKuai-Edge-Lab"
		lanIP = "10.10.0.1"
		arch = "arm/arm64"
		ver = "4.0.314-beta arm64 Build202610131658"
		uptime = 1284300
		conns = 428
		users = 16
		cpuAvg = 14.2
		memPct = 47.0
	}

	upSpeed := upBase + rng.Int63n(900_000) - 450_000
	downSpeed := downBase + rng.Int63n(2_400_000) - 1_200_000

	wanChecks := []WanCheckDTO{
		{ID: 1, Interface: "wan1", ParentInterface: "wan1", IPAddr: "203.0.113.10", Gateway: "203.0.113.1", Internet: "静态IP", Result: "success", ErrMsg: "线路检测成功", AutoSwitch: "已启用", UpdateTime: "--", Comment: "电信千兆专线", AvgRTTMs: 4.4, DropRate: 0},
		{ID: 2, Interface: "wan2", ParentInterface: "wan2", IPAddr: "198.51.100.22", Gateway: "198.51.100.1", Internet: "DHCP", Result: "success", ErrMsg: "线路检测成功", AutoSwitch: "已启用", UpdateTime: fmt.Sprint(now - 120), Comment: "联通备用出口", AvgRTTMs: 7.2, DropRate: 0},
	}
	if lab {
		wanChecks = []WanCheckDTO{
			{ID: 1, Interface: "wan1", ParentInterface: "wan1", IPAddr: "198.51.100.88", Gateway: "198.51.100.1", Internet: "PPPoE", Result: "success", ErrMsg: "线路检测成功", AutoSwitch: "已启用", UpdateTime: fmt.Sprint(now - 60), Comment: "实验室拨号宽带", AvgRTTMs: 11.4, DropRate: 0.2},
		}
	}

	ifaceTraffic := []InterfaceTrafficDTO{
		{Interface: "lan1", IPAddr: lanIP, Comment: "内网桥接口", UploadBps: upSpeed * 9 / 10, DownloadBps: downSpeed * 9 / 10, TotalUp: 109318250800, TotalDown: 482612168600, Connections: 0, UpPacked: 156100, DownPacked: 295300},
		{Interface: "wan1", IPAddr: wanChecks[0].IPAddr, Comment: wanChecks[0].Comment, UploadBps: upSpeed * 7 / 10, DownloadBps: downSpeed * 75 / 100, TotalUp: 83088261020, TotalDown: 354135935800, Connections: conns * 7 / 10, UpPacked: 350640, DownPacked: 856368},
	}
	if !lab {
		ifaceTraffic = append(ifaceTraffic, InterfaceTrafficDTO{
			Interface: "wan2", IPAddr: wanChecks[1].IPAddr, Comment: wanChecks[1].Comment, UploadBps: upSpeed * 3 / 10, DownloadBps: downSpeed * 25 / 100, TotalUp: 26854541000, TotalDown: 128654033200, Connections: conns * 3 / 10, UpPacked: 175820, DownPacked: 359925,
		})
	}

	phyPorts := []PhysicalPortDTO{
		{Name: "eth0", Interface: "wan1", Link: 1, Speed: 2500, Duplex: 2, MAC: "bc:24:11:18:79:dc", Type: "TP", Driver: "igc", Model: "Intel I225-V 2.5GbE"},
		{Name: "eth1", Interface: "wan2", Link: 1, Speed: 1000, Duplex: 2, MAC: "bc:24:11:a5:78:0a", Type: "TP", Driver: "igc", Model: "Intel I225-V 2.5GbE"},
		{Name: "eth2", Interface: "lan1", Link: 1, Speed: 10000, Duplex: 2, MAC: "bc:24:11:a9:bc:78", Type: "SFP+", Driver: "ixgbe", Model: "Intel X520 10GbE SFP+"},
		{Name: "eth3", Interface: "Null", Link: 0, Speed: 0, Duplex: 0, MAC: "bc:24:11:a9:bc:79", Type: "TP", Driver: "igc", Model: "Intel I225-V 2.5GbE"},
	}

	history := make([]WanHistoryPointDTO, 0, 48)
	for i := 23; i >= 0; i-- {
		ts := now - int64(i*300)
		wave := math.Sin(float64(i)/3.0)*0.3 + 1.0
		history = append(history, WanHistoryPointDTO{
			Timestamp:       ts,
			Interface:       "all",
			AvgUploadKBps:   math.Round(float64(upBase)/1024*wave*100) / 100,
			AvgDownloadKBps: math.Round(float64(downBase)/1024*wave*100) / 100,
			MaxUploadKBps:   math.Round(float64(upBase)/1024*wave*1.35*100) / 100,
			MaxDownloadKBps: math.Round(float64(downBase)/1024*wave*1.35*100) / 100,
			AvgRTTMs:        math.Round((4.6+math.Cos(float64(i)/4.0)*0.8)*100) / 100,
			MaxRTTMs:        9.2,
			MinRTTMs:        3.8,
			DropRate:        0,
		})
	}

	return &OverviewDTO{
		Router: RouterSummaryDTO{
			ID:         router.ID,
			Name:       router.Name,
			Hostname:   hostname,
			GWID:       "c23aa104204a41439043f9c83b262c64",
			LanIP:      lanIP,
			LinkStatus: 1,
			Uptime:     uptime,
			Version:    "4.0.314",
			Verstring:  ver,
			Arch:       arch,
			BuildDate:  "202610131658",
			Mock:       true,
		},
		Stream: StreamSummaryDTO{
			UploadBps:   upSpeed,
			DownloadBps: downSpeed,
			TotalUp:     109854541000,
			TotalDown:   482612168600,
		},
		CPU: CPUSummaryDTO{
			AvgPercent: cpuAvg,
			Cores:      []float64{cpuAvg, cpuAvg * 0.72, cpuAvg * 1.15, cpuAvg * 0.64},
			FreqMHz:    []int64{3187, 3187, 3187, 3187},
			Temps:      []float64{46.5, 47.0},
		},
		Memory: MemorySummaryDTO{
			TotalKB:     4096000,
			AvailableKB: 2498560,
			FreeKB:      1820000,
			CachedKB:    524000,
			BuffersKB:   154560,
			UsedPercent: memPct,
		},
		Connections: ConnectionSummaryDTO{
			Total: conns,
			TCP:   conns * 68 / 100,
			UDP:   conns * 28 / 100,
			ICMP:  conns * 4 / 100,
		},
		OnlineUsers: OnlineUserSummaryDTO{
			Total:    users,
			Wired:    users * 45 / 100,
			Wireless: users * 55 / 100,
			WiFi2G:   users * 18 / 100,
			WiFi5G:   users * 37 / 100,
		},
		WanChecks:        wanChecks,
		InterfaceTraffic: ifaceTraffic,
		PhysicalPorts:    phyPorts,
		WanHistory:       history,
		CollectedAt:      now,
	}
}

func mockLoadHistory(router config.RouterConfig, metric, timeRange, aggregate string) *LoadHistoryDTO {
	now := time.Now().Unix()
	points := 30
	step := int64(120)
	if timeRange == "day" {
		step = 2880
	} else if timeRange == "week" {
		step = 20160
	}
	base := 18.0
	if isLabMock(router) {
		base = 24.0
	}
	switch metric {
	case "memory":
		base = 39.0
	case "disk":
		base = 28.0
	case "cputemp":
		base = 47.0
	case "connections":
		base = 1180.0
	case "terminals":
		base = 40.0
	case "network":
		base = 24000.0
	}

	series := make([]map[string]any, 0, points)
	for i := points - 1; i >= 0; i-- {
		ts := now - int64(i)*step
		val := math.Round((base+math.Sin(float64(i)/2.5)*base*0.18)*100) / 100
		series = append(series, map[string]any{
			"timestamp": ts,
			"value":     val,
			"upload":    math.Round(val * 0.28),
			"download":  math.Round(val),
		})
	}
	return &LoadHistoryDTO{
		Metric:    metric,
		TimeRange: timeRange,
		Aggregate: aggregate,
		Series:    series,
	}
}

func mockClientsPage(router config.RouterConfig, mode string, page, pageSize int, search string) *ClientPageDTO {
	lab := isLabMock(router)
	subnet := "192.168.9"
	if lab {
		subnet = "10.10.0"
	}

	all := []ClientRecordDTO{
		{ID: 1, MAC: "00:22:1F:97:F2:CD", Hostname: "workstation-01", IPAddr: subnet + ".101", IPs: []string{subnet + ".101"}, Interface: "lan1", UploadSpeed: 1_450_000, DownloadSpeed: 8_920_000, TotalUp: 42144825344, TotalDown: 152829224960, Connections: 184, Comment: "研发主力工作站", ClientType: "PC", Uptime: "8h 32m"},
		{ID: 2, MAC: "8E:F6:C5:99:8A:50", Hostname: "synology-nas", IPAddr: subnet + ".10", IPs: []string{subnet + ".10"}, Interface: "lan1", UploadSpeed: 4_120_000, DownloadSpeed: 12_400_000, TotalUp: 513192809122, TotalDown: 449210291000, Connections: 412, Comment: "企业级群晖 NAS", ClientType: "NAS", Uptime: "360h 15m"},
		{ID: 3, MAC: "00:70:FA:3D:30:10", Hostname: "macbook-pro", IPAddr: subnet + ".103", IPs: []string{subnet + ".103"}, Interface: "lan1", UploadSpeed: 286_000, DownloadSpeed: 3_420_000, TotalUp: 13704825344, TotalDown: 44782922496, Connections: 68, Comment: "MacBook Pro M4", ClientType: "PC", Uptime: "2h 15m", SSID: "iKuai_5G", Signal: "-52 dBm"},
		{ID: 4, MAC: "78:DF:72:9F:D1:1A", Hostname: "iphone-16-pro", IPAddr: subnet + ".118", IPs: []string{subnet + ".118"}, Interface: "lan1", UploadSpeed: 42_000, DownloadSpeed: 640_000, TotalUp: 845129000, TotalDown: 9401292000, Connections: 29, Comment: "移动端测试机", ClientType: "Phone", Uptime: "1h 20m", SSID: "iKuai_5G", Signal: "-58 dBm"},
		{ID: 5, MAC: "EA:6C:25:F5:D7:43", Hostname: "pve-cluster-01", IPAddr: subnet + ".20", IPs: []string{subnet + ".20"}, Interface: "lan1", UploadSpeed: 520_000, DownloadSpeed: 3_120_000, TotalUp: 93202390100, TotalDown: 164481023200, Connections: 145, Comment: "PVE 虚拟化宿主机", ClientType: "NAS", Uptime: "720h 0m"},
	}

	switch mode {
	case "offline":
		all = []ClientRecordDTO{
			{ID: 11, MAC: "DA:8A:9F:53:43:2D", Hostname: "guest-ipad", IPAddr: subnet + ".211", IPs: []string{subnet + ".211"}, Interface: "lan1", TotalUp: 24000000, TotalDown: 450000000, Uptime: "1h 12m", Downtime: "2h ago", ClientType: "Pad", Comment: "访客平板"},
			{ID: 12, MAC: "54:48:E6:95:66:9F", Hostname: "meeting-tv", IPAddr: subnet + ".225", IPs: []string{subnet + ".225"}, Interface: "lan1", TotalUp: 1200000, TotalDown: 860000000, Uptime: "4h 05m", Downtime: "18h ago", ClientType: "OTHER", Comment: "会议室投屏"},
		}
	case "ipv6-online":
		for i := range all {
			ip6 := fmt.Sprintf("2001:db8:%x::%d", i+1, 100+i)
			all[i].IPAddr = ip6
			all[i].IPs = []string{ip6}
		}
	case "ipv6-offline":
		all = []ClientRecordDTO{
			{ID: 21, MAC: "DA:8A:9F:53:43:2D", Hostname: "guest-ipad", IPAddr: "2001:db8:2::211", IPs: []string{"2001:db8:2::211"}, Interface: "lan1", TotalUp: 4200000, TotalDown: 112000000, Downtime: "2h ago", ClientType: "Pad"},
		}
	}

	filtered := make([]ClientRecordDTO, 0, len(all))
	for _, c := range all {
		if search != "" {
			hay := strings.ToLower(c.Hostname + " " + c.IPAddr + " " + c.MAC + " " + c.Comment)
			if !strings.Contains(hay, strings.ToLower(search)) {
				continue
			}
		}
		filtered = append(filtered, c)
	}

	return &ClientPageDTO{
		Mode:     mode,
		Page:     page,
		PageSize: pageSize,
		Total:    len(filtered),
		Items:    filtered,
	}
}

func mockClientDetail(ip, mac string) *ClientDetailDTO {
	now := time.Now().Unix()
	load := make([]map[string]any, 0, 12)
	for i := 11; i >= 0; i-- {
		load = append(load, map[string]any{
			"timestamp": now - int64(i*300),
			"upload":    250000 + i*12000,
			"download":  1800000 + (11-i)*140000,
			"conn_num":  42 + i*3,
		})
	}
	return &ClientDetailDTO{
		IP:          ip,
		MAC:         mac,
		TrafficLoad: load,
		AppProto: []map[string]any{
			{"appname": "HTTPS / TLS", "total_down": 1824000000, "total_up": 324000000},
			{"appname": "QUIC / HTTP3", "total_down": 945000000, "total_up": 128000000},
			{"appname": "SSH / Git", "total_down": 210000000, "total_up": 540000000},
		},
	}
}

func mockAppProtocolAnalytics() *AppProtocolAnalyticsDTO {
	return &AppProtocolAnalyticsDTO{
		Protocols: []map[string]any{
			{"proto_name": "HTTPS", "name": "HTTPS", "total_down": 248000000000, "total_up": 34000000000, "percent": 52.4},
			{"proto_name": "QUIC", "name": "QUIC", "total_down": 112000000000, "total_up": 14000000000, "percent": 23.6},
			{"proto_name": "SSH/SFTP", "name": "SSH/SFTP", "total_down": 48000000000, "total_up": 38000000000, "percent": 11.8},
			{"proto_name": "DNS", "name": "DNS", "total_down": 12000000000, "total_up": 9000000000, "percent": 4.2},
			{"proto_name": "其他协议", "name": "其他协议", "total_down": 28000000000, "total_up": 11000000000, "percent": 8.0},
		},
		AppSummary: []map[string]any{
			{"appid": 1001, "appname": "Cloudflare / CDN", "appname_level1": "网络服务", "total_down": 124500000000, "total_up": 12400000000, "connect_num": 312},
			{"appid": 1002, "appname": "GitHub / DevTools", "appname_level1": "研发工具", "total_down": 84200000000, "total_up": 28400000000, "connect_num": 195},
			{"appid": 1003, "appname": "哔哩哔哩视频", "appname_level1": "流媒体", "total_down": 68900000000, "total_up": 4100000000, "connect_num": 84},
			{"appid": 1004, "appname": "Docker Registry", "appname_level1": "容器镜像", "total_down": 51400000000, "total_up": 2200000000, "connect_num": 46},
			{"appid": 1005, "appname": "WireGuard VPN", "appname_level1": "加密隧道", "total_down": 32100000000, "total_up": 29800000000, "connect_num": 12},
		},
		AppLoad: []map[string]any{
			{"appid": 1001, "appname": "Cloudflare / CDN", "conn_cnt": 312, "upload": 842000, "download": 11240000, "total_down": 124500000000},
			{"appid": 1002, "appname": "GitHub / DevTools", "conn_cnt": 195, "upload": 2140000, "download": 6420000, "total_down": 84200000000},
			{"appid": 1003, "appname": "哔哩哔哩视频", "conn_cnt": 84, "upload": 180000, "download": 5420000, "total_down": 68900000000},
		},
	}
}

func mockNetworkMap(router config.RouterConfig) *NetworkMapData {
	ov := mockOverview(router)
	nodes := []MapNode{
		{ID: "router", Name: ov.Router.Hostname, Type: "router", IP: ov.Router.LanIP, Category: 0, Status: "online"},
	}
	links := []MapLink{}
	for _, w := range ov.WanChecks {
		id := "wan:" + w.Interface
		nodes = append(nodes, MapNode{ID: id, Name: fmt.Sprintf("%s (%s)", w.Interface, w.Comment), Type: "wan", IP: w.IPAddr, Category: 1, Status: w.Result})
		links = append(links, MapLink{Source: id, Target: "router"})
	}
	nodes = append(nodes, MapNode{ID: "lan:lan1", Name: "LAN1 · 核心内网", Type: "lan", IP: ov.Router.LanIP, Category: 2, Status: "up"})
	links = append(links, MapLink{Source: "router", Target: "lan:lan1"})

	clients := mockClientsPage(router, "online", 1, 50, "").Items
	for _, c := range clients {
		id := "dev:" + c.MAC
		nodes = append(nodes, MapNode{ID: id, Name: c.Comment, Type: "device", IP: c.IPAddr, Category: 3})
		links = append(links, MapLink{Source: "lan:lan1", Target: id})
	}
	return &NetworkMapData{Nodes: nodes, Links: links}
}

func mockSecurityHub(router config.RouterConfig) *SecurityHubData {
	clients := mockClientsPage(router, "online", 1, 50, "").Items
	var abnormal []LegacyClientDTO
	for _, c := range clients {
		if c.Connections > 150 {
			abnormal = append(abnormal, LegacyClientDTO{
				MAC: c.MAC, Hostname: c.Hostname, IPs: c.IPs, UploadSpeed: c.UploadSpeed, DownloadSpeed: c.DownloadSpeed,
				TotalUp: c.TotalUp, TotalDown: c.TotalDown, Connections: c.Connections, Comment: c.Comment, ClientType: c.ClientType, Uptime: c.Uptime,
			})
		}
	}
	return &SecurityHubData{
		HighRiskPorts: []PortMapping{
			{Name: "NAS 管理后台", ExtPort: "5001", IntIP: "192.168.9.10", IntPort: "5001", Protocol: "TCP", Enabled: true},
			{Name: "WireGuard 隧道", ExtPort: "51820", IntIP: "192.168.9.1", IntPort: "51820", Protocol: "UDP", Enabled: true},
		},
		AbnormalDevices: abnormal,
		ACLCount:        6,
		PeerconnCount:   2,
		URLBlackCount:   3,
	}
}

func mockMultiWan(router config.RouterConfig) *MultiWanData {
	ov := mockOverview(router)
	legacy := ToLegacyInterfaceData(ov)
	return &MultiWanData{
		WanStatus: legacy.WanStatus,
		Routes: []RouteRule{
			{Type: "Domain", Name: "cdn-direct", Interface: "wan1", Target: "example.com,cdn.example.net", Enabled: true},
			{Type: "FiveTuple", Name: "ssh-wan2", Interface: "wan2", Target: "tcp:22", Enabled: true},
			{Type: "LoadBalance", Name: "dual-wan-lb", Interface: "wan1,wan2", Target: "weight=2:1", Enabled: true},
		},
	}
}

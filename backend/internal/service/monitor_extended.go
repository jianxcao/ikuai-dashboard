package service

import (
	"context"

	"ikuai-dashboard/backend/internal/logger"
)

// DTOs
type NetworkMapData struct {
	Nodes []MapNode `json:"nodes"`
	Links []MapLink `json:"links"`
}
type MapNode struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"` // "router", "wan", "lan", "device"
	IP       string `json:"ip"`
	Category int    `json:"category"`
}
type MapLink struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type SecurityHubData struct {
	HighRiskPorts   []PortMapping `json:"high_risk_ports"`
	AbnormalDevices []ClientDTO   `json:"abnormal_devices"`
}
type PortMapping struct {
	Name     string `json:"name"`
	ExtPort  string `json:"ext_port"`
	IntIP    string `json:"int_ip"`
	IntPort  string `json:"int_port"`
	Protocol string `json:"protocol"`
}

type MultiWanData struct {
	WanStatus []WanStatus `json:"wan_status"`
	Routes    []RouteRule `json:"routes"`
}
type RouteRule struct {
	Type      string `json:"type"`
	Interface string `json:"interface"`
	Target    string `json:"target"`
	Enabled   bool   `json:"enabled"`
}

type MonitorInsightsData struct {
	Summary          SummaryInfo     `json:"summary"`
	TopClients       []ClientDTO     `json:"top_clients"`
	TopInterfaces    []TrafficDetail `json:"top_interfaces"`
	AbnormalClients  []ClientDTO     `json:"abnormal_clients"`
	HighRiskMappings []PortMapping   `json:"high_risk_mappings"`
}

func (s *MonitorService) GetNetworkMap(ctx context.Context) (*NetworkMapData, error) {
	if s == nil {
		return (&MonitorService{}).getMockNetworkMap(), nil
	}
	if s.unconfigured {
		return nil, ErrUnconfigured
	}
	if s.mockMode {
		return s.getMockNetworkMap(), nil
	}
	iface, err := s.GetInterfaceData(ctx)
	if err != nil {
		return nil, err
	}
	clients, err := s.GetLanClients(ctx, "")
	if err != nil {
		logger.Log.Warnf("构建网络拓扑时获取客户端失败: %v", err)
		clients = nil
	}
	return buildNetworkMapFromData(s.router.Name, iface, clients), nil
}

func (s *MonitorService) GetSecurityHub(ctx context.Context) (*SecurityHubData, error) {
	if s == nil {
		return (&MonitorService{}).getMockSecurityHub(), nil
	}
	if s.unconfigured {
		return nil, ErrUnconfigured
	}
	if s.mockMode {
		return s.getMockSecurityHub(), nil
	}
	def, _ := commonResourceCatalog().Lookup("dnat-rules")
	rows, err := s.getCommonResourceRows(ctx, def)
	if err != nil {
		logger.Log.Warnf("获取 DNAT 暴露面失败: %v", err)
	}
	clients, err := s.GetLanClients(ctx, "")
	if err != nil {
		logger.Log.Warnf("获取异常终端失败: %v", err)
	}
	return &SecurityHubData{
		HighRiskPorts:   portMappingsFromRows(rows),
		AbnormalDevices: abnormalClientsFromClients(clients),
	}, nil
}

func (s *MonitorService) GetMultiWan(ctx context.Context) (*MultiWanData, error) {
	if s == nil {
		return (&MonitorService{}).getMockMultiWan(), nil
	}
	if s.unconfigured {
		return nil, ErrUnconfigured
	}
	if s.mockMode {
		return s.getMockMultiWan(), nil
	}
	iface, err := s.GetInterfaceData(ctx)
	if err != nil {
		return nil, err
	}
	routes := make([]RouteRule, 0)
	for _, item := range []struct {
		resource string
		label    string
	}{
		{resource: "static-routes", label: "Static"},
		{resource: "policy-routes", label: "Policy"},
		{resource: "domain-rules", label: "Domain"},
	} {
		def, _ := commonResourceCatalog().Lookup(item.resource)
		rows, err := s.getCommonResourceRows(ctx, def)
		if err != nil {
			logger.Log.Warnf("获取 %s 失败: %v", item.resource, err)
			continue
		}
		routes = append(routes, routeRulesFromRows(item.label, rows)...)
	}
	return &MultiWanData{WanStatus: iface.WanStatus, Routes: routes}, nil
}

func (s *MonitorService) GetMonitorInsights(ctx context.Context) (*MonitorInsightsData, error) {
	if s == nil {
		return nil, nil
	}
	if s.unconfigured {
		return nil, ErrUnconfigured
	}
	if s.mockMode {
		iface := s.getMockInterfaceData()
		clients := s.getMockLanClients("")
		security := s.getMockSecurityHub()
		return &MonitorInsightsData{
			Summary:          iface.Summary,
			TopClients:       topClientsByTraffic(clients, 8),
			TopInterfaces:    topInterfacesByTraffic(iface.TrafficDetails, 8),
			AbnormalClients:  abnormalClientsFromClients(clients),
			HighRiskMappings: security.HighRiskPorts,
		}, nil
	}

	iface, err := s.GetInterfaceData(ctx)
	if err != nil {
		return nil, err
	}
	clients, err := s.GetLanClients(ctx, "")
	if err != nil {
		logger.Log.Warnf("获取监控分析终端失败: %v", err)
		clients = nil
	}
	security, err := s.GetSecurityHub(ctx)
	if err != nil {
		logger.Log.Warnf("获取监控分析安全摘要失败: %v", err)
		security = &SecurityHubData{}
	}
	return &MonitorInsightsData{
		Summary:          iface.Summary,
		TopClients:       topClientsByTraffic(clients, 8),
		TopInterfaces:    topInterfacesByTraffic(iface.TrafficDetails, 8),
		AbnormalClients:  abnormalClientsFromClients(clients),
		HighRiskMappings: security.HighRiskPorts,
	}, nil
}

// Mock generators
func (s *MonitorService) getMockNetworkMap() *NetworkMapData {
	data := &NetworkMapData{
		Nodes: []MapNode{
			{ID: "router", Name: "iKuai", Type: "router", IP: "192.168.50.1", Category: 0},
			{ID: "wan1", Name: "WAN1 (电信)", Type: "wan", IP: "123.118.2.47", Category: 1},
			{ID: "wan2", Name: "WAN2 (联通)", Type: "wan", IP: "123.118.3.80", Category: 1},
			{ID: "lan1", Name: "LAN", Type: "lan", IP: "192.168.50.x", Category: 2},
		},
		Links: []MapLink{
			{Source: "wan1", Target: "router"},
			{Source: "wan2", Target: "router"},
			{Source: "router", Target: "lan1"},
		},
	}

	clients := s.getMockLanClients("")
	for _, c := range clients {
		data.Nodes = append(data.Nodes, MapNode{
			ID: c.MAC, Name: c.Hostname, Type: "device", IP: c.IPs[0], Category: 3,
		})
		data.Links = append(data.Links, MapLink{Source: "lan1", Target: c.MAC})
	}
	return data
}

func (s *MonitorService) getMockSecurityHub() *SecurityHubData {
	return &SecurityHubData{
		HighRiskPorts: []PortMapping{
			{Name: "NAS SSH", ExtPort: "2222", IntIP: "192.168.50.100", IntPort: "22", Protocol: "TCP"},
			{Name: "Web", ExtPort: "80", IntIP: "192.168.50.222", IntPort: "80", Protocol: "TCP"},
		},
		AbnormalDevices: []ClientDTO{
			s.getMockLanClients("")[0],
		},
	}
}

func (s *MonitorService) getMockMultiWan() *MultiWanData {
	iface := s.getMockInterfaceData()
	return &MultiWanData{
		WanStatus: iface.WanStatus,
		Routes: []RouteRule{
			{Type: "Domain", Interface: "wan1", Target: "netflix.com", Enabled: true},
			{Type: "Domain", Interface: "wan2", Target: "bilibili.com", Enabled: true},
		},
	}
}

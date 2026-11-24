package v4service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"ikuai-dashboard/backend/internal/config"
	"ikuai-dashboard/backend/internal/ikuaiv4"
)

type AppProtocolAnalyticsDTO struct {
	Protocols  []map[string]any `json:"protocols"`
	AppSummary []map[string]any `json:"app_summary"`
	AppLoad    []map[string]any `json:"app_load"`
}

type MapNode struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	IP       string `json:"ip"`
	Category int    `json:"category"`
	Status   string `json:"status,omitempty"`
}

type MapLink struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type NetworkMapData struct {
	Nodes []MapNode `json:"nodes"`
	Links []MapLink `json:"links"`
}

type PortMapping struct {
	Name     string `json:"name"`
	ExtPort  string `json:"ext_port"`
	IntIP    string `json:"int_ip"`
	IntPort  string `json:"int_port"`
	Protocol string `json:"protocol"`
	Enabled  bool   `json:"enabled"`
}

type SecurityHubData struct {
	HighRiskPorts   []PortMapping     `json:"high_risk_ports"`
	AbnormalDevices []LegacyClientDTO `json:"abnormal_devices"`
	ACLCount        int               `json:"acl_count"`
	PeerconnCount   int               `json:"peerconn_count"`
	URLBlackCount   int               `json:"url_black_count"`
}

type RouteRule struct {
	Type      string `json:"type"`
	Interface string `json:"interface"`
	Target    string `json:"target"`
	Enabled   bool   `json:"enabled"`
	Name      string `json:"name,omitempty"`
}

type MultiWanData struct {
	WanStatus []LegacyWanStatus `json:"wan_status"`
	Routes    []RouteRule       `json:"routes"`
}

// GetAppProtocolAnalytics queries protocol distribution & top L7 app traffic.
func GetAppProtocolAnalytics(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client) (*AppProtocolAnalyticsDTO, error) {
	if router.Mock || client == nil {
		return mockAppProtocolAnalytics(), nil
	}

	dto := &AppProtocolAnalyticsDTO{
		Protocols:  []map[string]any{},
		AppSummary: []map[string]any{},
		AppLoad:    []map[string]any{},
	}

	var (
		protoRows []map[string]any
		sumRows   []map[string]any
		loadRows  []map[string]any
		wg        sync.WaitGroup
	)
	wg.Add(3)
	go func() {
		defer wg.Done()
		if raw, err := client.Get(ctx, "/monitoring/protocols", nil); err == nil {
			protoRows, _ = ikuaiv4.ExtractCollection(raw)
		}
	}()
	go func() {
		defer wg.Done()
		if raw, err := client.Get(ctx, "/monitoring/app-traffic-summary", map[string]string{"page": "1", "limit": "20"}); err == nil {
			sumRows, _ = ikuaiv4.ExtractCollection(raw)
		}
	}()
	go func() {
		defer wg.Done()
		if raw, err := client.Get(ctx, "/monitoring/app-protocols/load", map[string]string{
			"page":     "1",
			"limit":    "20",
			"order":    "desc",
			"order_by": "total_down",
		}); err == nil {
			loadRows, _ = ikuaiv4.ExtractCollection(raw)
		}
	}()
	wg.Wait()

	if protoRows != nil {
		dto.Protocols = protoRows
	}
	if sumRows != nil {
		dto.AppSummary = sumRows
	}
	if loadRows != nil {
		dto.AppLoad = loadRows
	}
	return dto, nil
}

// GetNetworkMap builds a real topology graph from WAN checks, LAN interfaces, VLANs, APs and online clients.
func GetNetworkMap(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client) (*NetworkMapData, error) {
	if router.Mock || client == nil {
		return mockNetworkMap(router), nil
	}

	ov, err := GetOverview(ctx, router, client)
	if err != nil {
		return nil, err
	}

	routerName := ov.Router.Hostname
	if routerName == "" {
		routerName = router.Name
	}

	data := &NetworkMapData{
		Nodes: []MapNode{
			{
				ID:       "router",
				Name:     routerName,
				Type:     "router",
				IP:       ov.Router.LanIP,
				Category: 0,
				Status:   "online",
			},
		},
		Links: []MapLink{},
	}

	for _, w := range ov.WanChecks {
		id := "wan:" + w.Interface
		label := w.Interface
		if w.Comment != "" {
			label = fmt.Sprintf("%s (%s)", w.Interface, w.Comment)
		}
		data.Nodes = append(data.Nodes, MapNode{
			ID:       id,
			Name:     label,
			Type:     "wan",
			IP:       w.IPAddr,
			Category: 1,
			Status:   w.Result,
		})
		data.Links = append(data.Links, MapLink{Source: id, Target: "router"})
	}

	lanID := "lan:lan1"
	hasLanNode := false
	for _, t := range ov.InterfaceTraffic {
		if strings.HasPrefix(strings.ToLower(t.Interface), "lan") {
			id := "lan:" + t.Interface
			if !hasLanNode {
				lanID = id
				hasLanNode = true
			}
			data.Nodes = append(data.Nodes, MapNode{
				ID:       id,
				Name:     strings.ToUpper(t.Interface),
				Type:     "lan",
				IP:       t.IPAddr,
				Category: 2,
				Status:   "up",
			})
			data.Links = append(data.Links, MapLink{Source: "router", Target: id})
		}
	}
	if !hasLanNode {
		data.Nodes = append(data.Nodes, MapNode{
			ID:       lanID,
			Name:     "LAN1",
			Type:     "lan",
			IP:       ov.Router.LanIP,
			Category: 2,
			Status:   "up",
		})
		data.Links = append(data.Links, MapLink{Source: "router", Target: lanID})
	}

	// Wireless APs if any (/network/ac/ap-config takes no paging args and uses tagname)
	if apRaw, err := client.Get(ctx, "/network/ac/ap-config", nil); err == nil {
		apRows, _ := ikuaiv4.ExtractCollection(apRaw)
		for _, ap := range apRows {
			apID := "ap:" + ikuaiv4.Str(ap, "mac", "id")
			apName := ikuaiv4.Str(ap, "tagname", "apname", "name", "mac")
			if apName == "" {
				continue
			}
			data.Nodes = append(data.Nodes, MapNode{
				ID:       apID,
				Name:     "AP · " + apName,
				Type:     "lan",
				IP:       ikuaiv4.Str(ap, "ip_addr", "ip"),
				Category: 2,
			})
			data.Links = append(data.Links, MapLink{Source: lanID, Target: apID})
		}
	}

	clients, err := GetMergedLanClients(ctx, router, client, "")
	if err == nil {
		limit := len(clients)
		if limit > 80 {
			limit = 80
		}
		for _, c := range clients[:limit] {
			nodeID := "dev:" + c.MAC
			if c.MAC == "" && len(c.IPs) > 0 {
				nodeID = "dev:" + c.IPs[0]
			}
			name := c.Comment
			if name == "" {
				name = c.Hostname
			}
			if name == "" && len(c.IPs) > 0 {
				name = c.IPs[0]
			}
			ip := ""
			if len(c.IPs) > 0 {
				ip = c.IPs[0]
			}
			data.Nodes = append(data.Nodes, MapNode{
				ID:       nodeID,
				Name:     name,
				Type:     "device",
				IP:       ip,
				Category: 3,
			})
			data.Links = append(data.Links, MapLink{Source: lanID, Target: nodeID})
		}
	}

	return data, nil
}

// GetSecurityHub aggregates exposed DNAT ports, ACL/Peerconn counts, and high-connection terminals.
func GetSecurityHub(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client) (*SecurityHubData, error) {
	if router.Mock || client == nil {
		return mockSecurityHub(router), nil
	}

	hub := &SecurityHubData{
		HighRiskPorts:   []PortMapping{},
		AbnormalDevices: []LegacyClientDTO{},
	}

	var (
		dnatRows    []map[string]any
		aclTotal    int
		peerTotal   int
		urlTotal    int
		clientsList []LegacyClientDTO
		wg          sync.WaitGroup
	)
	wg.Add(5)
	go func() {
		defer wg.Done()
		if raw, err := client.Get(ctx, "/network/dnat/rules", map[string]string{"page": "1", "limit": "100"}); err == nil {
			dnatRows, _ = ikuaiv4.ExtractCollection(raw)
		}
	}()
	go func() {
		defer wg.Done()
		if raw, err := client.Get(ctx, "/security/acl-rules", map[string]string{"page": "1", "limit": "1"}); err == nil {
			_, aclTotal = ikuaiv4.ExtractCollection(raw)
		}
	}()
	go func() {
		defer wg.Done()
		if raw, err := client.Get(ctx, "/security/peerconn/rules", map[string]string{"page": "1", "limit": "1"}); err == nil {
			_, peerTotal = ikuaiv4.ExtractCollection(raw)
		}
	}()
	go func() {
		defer wg.Done()
		if raw, err := client.Get(ctx, "/security/url-black/rules", map[string]string{"page": "1", "limit": "1"}); err == nil {
			_, urlTotal = ikuaiv4.ExtractCollection(raw)
		}
	}()
	go func() {
		defer wg.Done()
		if list, err := GetMergedLanClients(ctx, router, client, ""); err == nil {
			clientsList = list
		}
	}()
	wg.Wait()

	for _, r := range dnatRows {
		hub.HighRiskPorts = append(hub.HighRiskPorts, PortMapping{
			Name:     ikuaiv4.Str(r, "tagname", "comment", "name"),
			ExtPort:  ikuaiv4.Str(r, "wan_port"),
			IntIP:    ikuaiv4.Str(r, "lan_addr"),
			IntPort:  ikuaiv4.Str(r, "lan_port"),
			Protocol: strings.ToUpper(ikuaiv4.Str(r, "protocol")),
			Enabled:  ikuaiv4.Bool(r, "enabled"),
		})
	}
	hub.ACLCount = aclTotal
	hub.PeerconnCount = peerTotal
	hub.URLBlackCount = urlTotal
	for _, c := range clientsList {
		if c.Connections >= 300 || c.UploadSpeed >= 10*1024*1024 {
			hub.AbnormalDevices = append(hub.AbnormalDevices, c)
		}
	}

	return hub, nil
}

// GetMultiWan aggregates WAN status + static/domain/five-tuple/load-balance/updown routes.
func GetMultiWan(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client) (*MultiWanData, error) {
	if router.Mock || client == nil {
		return mockMultiWan(router), nil
	}

	ov, err := GetOverview(ctx, router, client)
	if err != nil {
		return nil, err
	}
	legacy := ToLegacyInterfaceData(ov)

	data := &MultiWanData{
		WanStatus: legacy.WanStatus,
		Routes:    []RouteRule{},
	}

	sources := []struct {
		label string
		path  string
	}{
		{"Static", "/routing/static-routes"},
		{"Domain", "/routing/domain-rules"},
		{"FiveTuple", "/routing/five-tuple-rules"},
		{"L7", "/routing/app-protocols"},
		{"LoadBalance", "/routing/load-balance-rules"},
		{"UpDown", "/routing/updown"},
	}

	rawResults := make([]json.RawMessage, len(sources))
	var wg sync.WaitGroup
	for i, src := range sources {
		wg.Add(1)
		go func(idx int, path string) {
			defer wg.Done()
			if raw, getErr := client.Get(ctx, path, map[string]string{"page": "1", "limit": "50"}); getErr == nil {
				rawResults[idx] = raw
			}
		}(i, src.path)
	}
	wg.Wait()

	for i, src := range sources {
		if len(rawResults[i]) == 0 {
			continue
		}
		rows, _ := ikuaiv4.ExtractCollection(rawResults[i])
		for _, r := range rows {
			target := ikuaiv4.Str(r, "domain", "dst_addr", "app_proto", "dst_port", "src_addr", "tagname")
			if target == "" {
				target = ikuaiv4.Str(r, "comment", "id")
			}
			iface := ikuaiv4.Str(r, "interface", "upiface", "gateway")
			data.Routes = append(data.Routes, RouteRule{
				Type:      src.label,
				Name:      ikuaiv4.Str(r, "tagname", "name"),
				Interface: iface,
				Target:    target,
				Enabled:   ikuaiv4.Bool(r, "enabled"),
			})
		}
	}
	return data, nil
}

package v4service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"ikuai-dashboard/backend/internal/config"
	"ikuai-dashboard/backend/internal/ikuaiv4"
)

// InterfaceOptionItem describes a selectable router interface option for forms.
type InterfaceOptionItem struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"` // "wan" | "lan" | "vlan" | "physical"
	IP      string `json:"ip,omitempty"`
	Proto   string `json:"proto,omitempty"`
	Comment string `json:"comment,omitempty"`
	Detail  string `json:"detail,omitempty"`
}

// InterfaceOptionsDTO groups router WAN, LAN/VLAN, and physical port options.
type InterfaceOptionsDTO struct {
	Wan      []InterfaceOptionItem `json:"wan"`
	Lan      []InterfaceOptionItem `json:"lan"`
	Physical []InterfaceOptionItem `json:"physical"`
}

const interfaceOptionsCacheTTL = 60 * time.Second

type interfaceOptionsCacheEntry struct {
	dto       *InterfaceOptionsDTO
	expiresAt time.Time
}

var (
	ifaceOptsCacheMu sync.Mutex
	ifaceOptsCache   = make(map[*ikuaiv4.Client]interfaceOptionsCacheEntry)
)

// GetInterfaceOptions returns selectable WAN, LAN/VLAN, and physical port names for the active router.
func GetInterfaceOptions(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client) (*InterfaceOptionsDTO, error) {
	if router.Mock || client == nil {
		return mockInterfaceOptions(router), nil
	}

	if cached := getCachedInterfaceOptions(client); cached != nil {
		return cached, nil
	}

	_, _, cachedPhy, hasPhyCache := getOverviewCachedRaw(client)

	var (
		wanCfgRaw  json.RawMessage
		wanCfgErr  error
		statusRaw  json.RawMessage
		statusErr  error
		wanVlanRaw json.RawMessage
		wanVlanErr error
		lanCfgRaw  json.RawMessage
		lanCfgErr  error
		vlanRaw    json.RawMessage
		vlanErr    error
		phyRaw     = cachedPhy
		phyErr     error
		wg         sync.WaitGroup
	)

	// Bound concurrent TLS/HTTP requests to embedded ARM routers so Nginx does not stall.
	sem := make(chan struct{}, 2)
	fetch := func(path string, params map[string]string, outRaw *json.RawMessage, outErr *error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			*outRaw, *outErr = client.Get(ctx, path, params)
		}()
	}

	fetch("/monitoring/interfaces-status", nil, &statusRaw, &statusErr)
	fetch("/interfaces/wan-config", nil, &wanCfgRaw, &wanCfgErr)
	fetch("/interfaces/lan-config", nil, &lanCfgRaw, &lanCfgErr)
	fetch("/interfaces/wan-vlan-config", nil, &wanVlanRaw, &wanVlanErr)
	fetch("/network/vlan", map[string]string{"page": "1", "limit": "100"}, &vlanRaw, &vlanErr)
	if !hasPhyCache {
		fetch("/interfaces/physical", nil, &phyRaw, &phyErr)
	}
	wg.Wait()

	if !hasPhyCache && phyErr == nil && len(phyRaw) > 0 {
		setOverviewCachedRaw(client, nil, false, phyRaw, true)
	}

	if wanCfgErr != nil && statusErr != nil && lanCfgErr != nil && phyErr != nil {
		return nil, statusErr
	}

	dto := &InterfaceOptionsDTO{
		Wan:      []InterfaceOptionItem{},
		Lan:      []InterfaceOptionItem{},
		Physical: []InterfaceOptionItem{},
	}

	wanIndex := make(map[string]int)
	upsertWan := func(item InterfaceOptionItem, overwriteLive bool) {
		item.Name = strings.TrimSpace(item.Name)
		if item.Name == "" || isNonWanInterface(item.Name) {
			return
		}
		key := strings.ToLower(item.Name)
		if idx, exists := wanIndex[key]; exists {
			existing := &dto.Wan[idx]
			if item.IP != "" && (overwriteLive || existing.IP == "") {
				existing.IP = item.IP
			}
			if item.Proto != "" && (overwriteLive || existing.Proto == "") {
				existing.Proto = item.Proto
			}
			if existing.Comment == "" && item.Comment != "" {
				existing.Comment = item.Comment
			}
			existing.Detail = formatOptionDetail(existing.IP, existing.Proto, existing.Comment)
			return
		}
		item.Kind = "wan"
		item.Detail = formatOptionDetail(item.IP, item.Proto, item.Comment)
		wanIndex[key] = len(dto.Wan)
		dto.Wan = append(dto.Wan, item)
	}

	if wanCfgErr == nil {
		rows, _ := ikuaiv4.ExtractCollection(wanCfgRaw)
		for _, row := range rows {
			name := ikuaiv4.Str(row, "tagname", "name", "interface")
			rawProto := ikuaiv4.Str(row, "internet")
			proto := mapWanInternetLabel(rawProto)
			var ipRaw string
			switch strings.TrimSpace(rawProto) {
			case "2":
				ipRaw = ikuaiv4.Str(row, "pppoe_ip_addr", "ip_addr", "ip_mask", "ip")
			case "1":
				ipRaw = ikuaiv4.Str(row, "dhcp_ip_addr", "ip_addr", "ip_mask", "ip")
			default:
				ipRaw = ikuaiv4.Str(row, "ip_addr", "ip_mask", "pppoe_ip_addr", "dhcp_ip_addr", "ip")
			}
			ip := cleanWanIP(ipRaw)
			comment := ikuaiv4.Str(row, "comment")
			upsertWan(InterfaceOptionItem{
				Name:    name,
				Kind:    "wan",
				IP:      ip,
				Proto:   proto,
				Comment: comment,
			}, false)
		}
	}

	if statusErr == nil {
		if statusAny, decErr := ikuaiv4.DecodeAny(statusRaw); decErr == nil {
			if statusMap, ok := statusAny.(map[string]any); ok {
				if checks, ok := statusMap["iface_check"].([]any); ok {
					for _, item := range checks {
						if row, ok := item.(map[string]any); ok {
							name := ikuaiv4.Str(row, "interface", "name", "tagname")
							ip := cleanWanIP(ikuaiv4.Str(row, "ip_addr", "ip", "ip_mask"))
							proto := mapWanInternetLabel(ikuaiv4.Str(row, "internet"))
							comment := ikuaiv4.Str(row, "comment")
							upsertWan(InterfaceOptionItem{
								Name:    name,
								Kind:    "wan",
								IP:      ip,
								Proto:   proto,
								Comment: comment,
							}, true)
						}
					}
				}
				if streams, ok := statusMap["iface_stream"].([]any); ok {
					for _, item := range streams {
						if row, ok := item.(map[string]any); ok {
							name := ikuaiv4.Str(row, "interface", "name", "tagname")
							if isNonWanInterface(name) {
								continue
							}
							ip := cleanWanIP(ikuaiv4.Str(row, "ip_addr", "ip", "ip_mask"))
							comment := ikuaiv4.Str(row, "comment")
							upsertWan(InterfaceOptionItem{
								Name:    name,
								Kind:    "wan",
								IP:      ip,
								Comment: comment,
							}, false)
						}
					}
				}
			}
		}
	}

	if wanVlanErr == nil {
		rows, _ := ikuaiv4.ExtractCollection(wanVlanRaw)
		for _, row := range rows {
			name := ikuaiv4.Str(row, "vlan_name", "tagname", "name")
			ip := cleanWanIP(ikuaiv4.Str(row, "ip_addr", "ip_mask", "ip"))
			comment := ikuaiv4.Str(row, "comment")
			upsertWan(InterfaceOptionItem{
				Name:    name,
				Kind:    "wan",
				IP:      ip,
				Proto:   "VLAN",
				Comment: comment,
			}, false)
		}
	}

	lanIndex := make(map[string]int)
	upsertLan := func(item InterfaceOptionItem) {
		item.Name = strings.TrimSpace(item.Name)
		if item.Name == "" {
			return
		}
		key := strings.ToLower(item.Name)
		if idx, exists := lanIndex[key]; exists {
			existing := &dto.Lan[idx]
			if existing.IP == "" && item.IP != "" {
				existing.IP = item.IP
			}
			if existing.Comment == "" && item.Comment != "" {
				existing.Comment = item.Comment
			}
			existing.Detail = formatOptionDetail(existing.IP, existing.Proto, existing.Comment)
			return
		}
		item.Detail = formatOptionDetail(item.IP, item.Proto, item.Comment)
		lanIndex[key] = len(dto.Lan)
		dto.Lan = append(dto.Lan, item)
	}

	if lanCfgErr == nil {
		rows, _ := ikuaiv4.ExtractCollection(lanCfgRaw)
		for _, row := range rows {
			name := ikuaiv4.Str(row, "tagname", "name", "interface")
			ip := strings.ReplaceAll(ikuaiv4.Str(row, "ip_mask", "ip_addr", "ip"), `\/`, "/")
			comment := ikuaiv4.Str(row, "comment")
			upsertLan(InterfaceOptionItem{
				Name:    name,
				Kind:    "lan",
				IP:      ip,
				Comment: comment,
			})
		}
	}

	if vlanErr == nil {
		rows, _ := ikuaiv4.ExtractCollection(vlanRaw)
		for _, row := range rows {
			name := ikuaiv4.Str(row, "vlan_name", "tagname", "name")
			ip := cleanWanIP(ikuaiv4.Str(row, "ip_addr", "ip_mask", "ip"))
			comment := ikuaiv4.Str(row, "comment")
			upsertLan(InterfaceOptionItem{
				Name:    name,
				Kind:    "vlan",
				IP:      ip,
				Proto:   "VLAN",
				Comment: comment,
			})
		}
	}

	if phyErr == nil {
		rows, _ := ikuaiv4.ExtractCollection(phyRaw)
		for _, row := range rows {
			name := ikuaiv4.Str(row, "name")
			if name == "" {
				continue
			}
			bound := ikuaiv4.Str(row, "interface")
			speed := ikuaiv4.Int(row, "speed")
			link := ikuaiv4.Int(row, "link")
			var parts []string
			if bound != "" && !strings.EqualFold(bound, "Null") {
				parts = append(parts, "已绑定 "+bound)
			} else {
				parts = append(parts, "未分配")
			}
			if speed > 0 {
				parts = append(parts, fmt.Sprintf("%dM", speed))
			} else if link == 1 {
				parts = append(parts, "已连接")
			} else {
				parts = append(parts, "未连接")
			}
			dto.Physical = append(dto.Physical, InterfaceOptionItem{
				Name:   name,
				Kind:   "physical",
				Detail: strings.Join(parts, " · "),
			})
		}
		sort.Slice(dto.Physical, func(i, j int) bool {
			return dto.Physical[i].Name < dto.Physical[j].Name
		})
	}

	setCachedInterfaceOptions(client, dto)
	return dto, nil
}

func mockInterfaceOptions(router config.RouterConfig) *InterfaceOptionsDTO {
	ov := mockOverview(router)
	dto := &InterfaceOptionsDTO{
		Wan:      make([]InterfaceOptionItem, 0, len(ov.WanChecks)),
		Lan:      []InterfaceOptionItem{},
		Physical: make([]InterfaceOptionItem, 0, len(ov.PhysicalPorts)),
	}
	for _, w := range ov.WanChecks {
		dto.Wan = append(dto.Wan, InterfaceOptionItem{
			Name:    w.Interface,
			Kind:    "wan",
			IP:      w.IPAddr,
			Proto:   w.Internet,
			Comment: w.Comment,
			Detail:  formatOptionDetail(w.IPAddr, w.Internet, w.Comment),
		})
	}
	for _, t := range ov.InterfaceTraffic {
		if strings.HasPrefix(strings.ToLower(t.Interface), "lan") {
			dto.Lan = append(dto.Lan, InterfaceOptionItem{
				Name:    t.Interface,
				Kind:    "lan",
				IP:      t.IPAddr,
				Comment: t.Comment,
				Detail:  formatOptionDetail(t.IPAddr, "", t.Comment),
			})
		}
	}
	if len(dto.Lan) == 0 {
		dto.Lan = append(dto.Lan, InterfaceOptionItem{
			Name:   "lan1",
			Kind:   "lan",
			IP:     ov.Router.LanIP,
			Detail: ov.Router.LanIP,
		})
	}
	for _, p := range ov.PhysicalPorts {
		detail := "未分配"
		if p.Interface != "" && !strings.EqualFold(p.Interface, "Null") {
			detail = "已绑定 " + p.Interface
		}
		if p.Speed > 0 {
			detail += fmt.Sprintf(" · %dM", p.Speed)
		} else if p.Link == 0 {
			detail += " · 未连接"
		}
		dto.Physical = append(dto.Physical, InterfaceOptionItem{
			Name:   p.Name,
			Kind:   "physical",
			Detail: detail,
		})
	}
	return dto
}

func mapWanInternetLabel(raw string) string {
	trimmed := strings.TrimSpace(raw)
	switch strings.ToLower(trimmed) {
	case "0":
		return "静态IP"
	case "1", "dhcp":
		return "DHCP"
	case "2", "pppoe":
		return "PPPoE"
	case "5", "wifi":
		return "WiFi"
	case "6", "lte":
		return "LTE"
	default:
		return trimmed
	}
}

func formatOptionDetail(ip, proto, comment string) string {
	var parts []string
	if ip != "" {
		parts = append(parts, ip)
	}
	if proto != "" {
		parts = append(parts, proto)
	}
	if comment != "" {
		parts = append(parts, comment)
	}
	return strings.Join(parts, " · ")
}

func getCachedInterfaceOptions(client *ikuaiv4.Client) *InterfaceOptionsDTO {
	ifaceOptsCacheMu.Lock()
	defer ifaceOptsCacheMu.Unlock()
	entry, ok := ifaceOptsCache[client]
	if !ok || entry.dto == nil || time.Now().After(entry.expiresAt) {
		return nil
	}
	return entry.dto
}

func setCachedInterfaceOptions(client *ikuaiv4.Client, dto *InterfaceOptionsDTO) {
	ifaceOptsCacheMu.Lock()
	defer ifaceOptsCacheMu.Unlock()
	if len(ifaceOptsCache) > 32 {
		ifaceOptsCache = make(map[*ikuaiv4.Client]interfaceOptionsCacheEntry)
	}
	ifaceOptsCache[client] = interfaceOptionsCacheEntry{
		dto:       dto,
		expiresAt: time.Now().Add(interfaceOptionsCacheTTL),
	}
}

package service

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	abnormalConnectionThreshold = 100
	abnormalUploadThreshold     = 10 * 1024 * 1024
)

func buildNetworkMapFromData(routerName string, iface *InterfaceDataResult, clients []ClientDTO) *NetworkMapData {
	if routerName == "" {
		routerName = "iKuai"
	}
	data := &NetworkMapData{
		Nodes: []MapNode{{ID: "router", Name: routerName, Type: "router", Category: 0}},
	}

	wanNames := map[string]struct{}{}
	for _, wan := range iface.WanStatus {
		name := strings.TrimSpace(wan.Name)
		if name == "" {
			continue
		}
		wanNames[strings.ToLower(name)] = struct{}{}
		nodeID := mapNodeID("wan", name)
		data.Nodes = append(data.Nodes, MapNode{
			ID:       nodeID,
			Name:     name,
			Type:     "wan",
			IP:       wan.IP,
			Category: 1,
		})
		data.Links = append(data.Links, MapLink{Source: nodeID, Target: "router"})
	}

	lanIDs := make([]string, 0)
	seenLAN := map[string]struct{}{}
	for _, detail := range iface.TrafficDetails {
		name := strings.TrimSpace(detail.Name)
		if name == "" {
			continue
		}
		lowerName := strings.ToLower(name)
		if _, isWAN := wanNames[lowerName]; isWAN || strings.HasPrefix(lowerName, "wan") {
			continue
		}
		if !strings.HasPrefix(lowerName, "lan") && !strings.HasPrefix(lowerName, "br") && !strings.HasPrefix(lowerName, "vlan") {
			continue
		}
		if _, exists := seenLAN[lowerName]; exists {
			continue
		}
		seenLAN[lowerName] = struct{}{}
		nodeID := mapNodeID("lan", name)
		lanIDs = append(lanIDs, nodeID)
		data.Nodes = append(data.Nodes, MapNode{
			ID:       nodeID,
			Name:     name,
			Type:     "lan",
			IP:       detail.IP,
			Category: 2,
		})
		data.Links = append(data.Links, MapLink{Source: "router", Target: nodeID})
	}

	if len(lanIDs) == 0 && len(clients) > 0 {
		lanIDs = append(lanIDs, "lan:default")
		data.Nodes = append(data.Nodes, MapNode{ID: "lan:default", Name: "LAN", Type: "lan", Category: 2})
		data.Links = append(data.Links, MapLink{Source: "router", Target: "lan:default"})
	}
	defaultLAN := ""
	if len(lanIDs) > 0 {
		defaultLAN = lanIDs[0]
	}

	seenDevices := map[string]struct{}{}
	for _, client := range clients {
		mac := strings.ToUpper(strings.TrimSpace(client.MAC))
		if mac == "" {
			continue
		}
		if _, exists := seenDevices[mac]; exists {
			continue
		}
		seenDevices[mac] = struct{}{}
		name := strings.TrimSpace(client.Hostname)
		if name == "" {
			name = strings.TrimSpace(client.Comment)
		}
		if name == "" {
			name = mac
		}
		ip := ""
		if len(client.IPs) > 0 {
			ip = client.IPs[0]
		}
		nodeID := mapNodeID("device", mac)
		data.Nodes = append(data.Nodes, MapNode{
			ID:       nodeID,
			Name:     name,
			Type:     "device",
			IP:       ip,
			Category: 3,
		})
		if defaultLAN != "" {
			data.Links = append(data.Links, MapLink{Source: defaultLAN, Target: nodeID})
		}
	}

	return data
}

func mapNodeID(prefix, value string) string {
	return prefix + ":" + strings.TrimSpace(value)
}

func portMappingsFromRows(rows []map[string]any) []PortMapping {
	ports := make([]PortMapping, 0, len(rows))
	for _, row := range rows {
		if !isEnabledValue(v4String(row, "enabled", "enable", "status")) {
			continue
		}
		extPort := v4String(row, "wan_port", "external_port", "ext_port", "wanport", "port")
		intPort := v4String(row, "lan_port", "internal_port", "int_port", "lanport")
		if !isHighRiskPort(extPort) && !isHighRiskPort(intPort) {
			continue
		}
		name := firstNonEmpty(
			v4String(row, "comment", "remark", "description", "desc"),
			v4String(row, "tagname", "name"),
			"DNAT "+v4String(row, "id", "rowid"),
		)
		ports = append(ports, PortMapping{
			Name:     strings.TrimSpace(name),
			ExtPort:  extPort,
			IntIP:    v4String(row, "lan_addr", "lan_ip", "internal_ip", "int_ip", "server_ip"),
			IntPort:  intPort,
			Protocol: strings.ToUpper(firstNonEmpty(v4String(row, "protocol", "proto"), "TCP")),
		})
	}
	return ports
}

func abnormalClientsFromClients(clients []ClientDTO) []ClientDTO {
	abnormal := make([]ClientDTO, 0)
	for _, client := range clients {
		if client.Connections >= abnormalConnectionThreshold || client.UploadSpeed >= abnormalUploadThreshold {
			abnormal = append(abnormal, client)
		}
	}
	sort.SliceStable(abnormal, func(i, j int) bool {
		if abnormal[i].Connections != abnormal[j].Connections {
			return abnormal[i].Connections > abnormal[j].Connections
		}
		return abnormal[i].UploadSpeed > abnormal[j].UploadSpeed
	})
	return abnormal
}

func routeRulesFromRows(ruleType string, rows []map[string]any) []RouteRule {
	rules := make([]RouteRule, 0, len(rows))
	for _, row := range rows {
		target := firstNonEmpty(
			v4String(row, "domain", "dst_addr", "destination", "target", "dst"),
			strings.TrimSpace(v4String(row, "src_addr", "source")+" → "+v4String(row, "dst_addr", "destination")),
			v4String(row, "comment", "remark", "description"),
		)
		target = strings.Trim(target, " →")
		if target == "" {
			target = "规则 " + v4String(row, "id", "rowid")
		}
		rules = append(rules, RouteRule{
			Type:      ruleType,
			Interface: firstNonEmpty(v4String(row, "interface", "iface", "wan", "out_interface"), "—"),
			Target:    target,
			Enabled:   isEnabledValue(v4String(row, "enabled", "enable", "status")),
		})
	}
	return rules
}

func topClientsByTraffic(clients []ClientDTO, limit int) []ClientDTO {
	ranked := append([]ClientDTO(nil), clients...)
	sort.SliceStable(ranked, func(i, j int) bool {
		left := ranked[i].UploadSpeed + ranked[i].DownloadSpeed
		right := ranked[j].UploadSpeed + ranked[j].DownloadSpeed
		if left != right {
			return left > right
		}
		return ranked[i].Connections > ranked[j].Connections
	})
	if limit > 0 && len(ranked) > limit {
		return ranked[:limit]
	}
	return ranked
}

func topInterfacesByTraffic(details []TrafficDetail, limit int) []TrafficDetail {
	ranked := append([]TrafficDetail(nil), details...)
	sort.SliceStable(ranked, func(i, j int) bool {
		left := ranked[i].UploadSpeed + ranked[i].DownloadSpeed
		right := ranked[j].UploadSpeed + ranked[j].DownloadSpeed
		if left != right {
			return left > right
		}
		return ranked[i].Connections > ranked[j].Connections
	})
	if limit > 0 && len(ranked) > limit {
		return ranked[:limit]
	}
	return ranked
}

func isEnabledValue(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return true
	}
	switch value {
	case "1", "yes", "true", "enable", "enabled", "success", "on":
		return true
	case "0", "no", "false", "disable", "disabled", "off":
		return false
	default:
		return true
	}
}

func isHighRiskPort(value string) bool {
	for _, token := range strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ';' || r == ' ' || r == '/'
	}) {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if strings.Contains(token, "-") {
			parts := strings.SplitN(token, "-", 2)
			start, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
			for port := range highRiskPorts {
				if port >= start && port <= end {
					return true
				}
			}
			continue
		}
		port, err := strconv.Atoi(token)
		if err != nil {
			continue
		}
		if _, ok := highRiskPorts[port]; ok {
			return true
		}
	}
	return false
}

var highRiskPorts = map[int]struct{}{
	21: {}, 22: {}, 23: {}, 25: {}, 53: {}, 80: {}, 110: {}, 143: {}, 443: {},
	445: {}, 993: {}, 995: {}, 1433: {}, 1521: {}, 2049: {}, 2375: {}, 2376: {},
	3306: {}, 3389: {}, 5432: {}, 5601: {}, 5900: {}, 6379: {}, 8080: {}, 8443: {},
	9200: {}, 9300: {}, 11211: {}, 27017: {},
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func rowsFromAny(value any) []map[string]any {
	return v4Rows(value)
}

func mapResult(message string, id int, raw any) map[string]any {
	result := map[string]any{"message": message}
	if id > 0 {
		result["id"] = id
	}
	if raw != nil {
		result["raw"] = raw
	}
	return result
}

func intFromString(value string) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("无效的资源 ID: %q", value)
	}
	return id, nil
}

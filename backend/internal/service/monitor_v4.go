package service

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"ikuai-dashboard/backend/internal/logger"
)

func (s *MonitorService) getV4InterfaceData(ctx context.Context) (*InterfaceDataResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := &InterfaceDataResult{}

	system, err := s.v4Get(ctx, "/monitoring/system")
	if err != nil {
		logger.Log.Errorf("获取 v4 首页摘要数据失败: %v", err)
	} else {
		result.Summary = SummaryInfo{
			UploadSpeed:      v4FindInt64(system, "upload_speed", "up_speed", "upload", "up", "tx_speed"),
			DownloadSpeed:    v4FindInt64(system, "download_speed", "down_speed", "download", "down", "rx_speed"),
			TotalConnections: int(v4FindInt64(system, "total_connections", "connect_num", "connections", "conn_count")),
			OnlineUsers:      int(v4FindInt64(system, "online_users", "online_user", "client_count", "clients")),
		}
	}

	statusPayload, err := s.v4Get(ctx, "/monitoring/interfaces-status")
	if err != nil {
		logger.Log.Warnf("获取 v4 接口状态失败: %v", err)
	} else {
		for _, row := range v4Rows(statusPayload) {
			result.WanStatus = append(result.WanStatus, WanStatus{
				Name:    v4String(row, "interface", "iface", "name", "if_name"),
				IP:      v4String(row, "ip", "ip_addr", "ipaddr", "address"),
				Proto:   v4String(row, "internet", "proto", "protocol", "mode"),
				Status:  v4String(row, "status", "result", "state", "link_state"),
				Comment: v4String(row, "comment", "remark", "description", "desc"),
			})
		}
	}

	trafficPayload, err := s.v4Get(ctx, "/monitoring/interfaces-traffic")
	if err != nil {
		logger.Log.Errorf("获取 v4 接口流量数据失败: %v", err)
		return result, nil
	}
	for _, row := range v4Rows(trafficPayload) {
		result.TrafficDetails = append(result.TrafficDetails, TrafficDetail{
			Name:          v4String(row, "interface", "iface", "name", "if_name"),
			IP:            v4String(row, "ip", "ip_addr", "ipaddr", "address"),
			UploadSpeed:   v4Int64(row, "upload_speed", "up_speed", "upload", "up", "tx_speed"),
			DownloadSpeed: v4Int64(row, "download_speed", "down_speed", "download", "down", "rx_speed"),
			TotalUp:       v4Int64(row, "total_up", "upload_total", "tx_bytes", "total_upload"),
			TotalDown:     v4Int64(row, "total_down", "download_total", "rx_bytes", "total_download"),
			Connections:   int(v4Int64(row, "connect_num", "connections", "conn_count")),
			Comment:       v4String(row, "comment", "remark", "description", "desc"),
		})
	}

	return result, nil
}

func (s *MonitorService) getV4LanClients(ctx context.Context, filterComment string) ([]ClientDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	aggregated := make(map[string]*ClientDTO)
	ipv4Payload, err := s.v4Get(ctx, "/monitoring/clients-online")
	if err != nil {
		logger.Log.Errorf("获取 v4 IPv4 客户端列表失败: %v", err)
		return nil, err
	}
	for _, row := range v4Rows(ipv4Payload) {
		mac := strings.ToUpper(v4String(row, "mac", "mac_addr", "macaddr"))
		if mac == "" {
			continue
		}
		aggregated[mac] = &ClientDTO{
			MAC:           mac,
			Hostname:      v4String(row, "hostname", "host_name", "name"),
			IPs:           []string{v4String(row, "ip", "ip_addr", "ipaddr", "address")},
			UploadSpeed:   v4Int64(row, "upload_speed", "up_speed", "upload", "up", "tx_speed"),
			DownloadSpeed: v4Int64(row, "download_speed", "down_speed", "download", "down", "rx_speed"),
			TotalUp:       v4Int64(row, "total_up", "upload_total", "tx_bytes", "total_upload"),
			TotalDown:     v4Int64(row, "total_down", "download_total", "rx_bytes", "total_download"),
			Connections:   int(v4Int64(row, "connect_num", "connections", "conn_count")),
			Comment:       v4String(row, "comment", "remark", "description", "desc"),
			ClientType:    v4String(row, "client_type", "type", "device_type"),
			Uptime:        v4String(row, "uptime", "online_time", "time"),
		}
	}

	ipv6Payload, err := s.v4Get(ctx, "/monitoring/clients-ip6-online")
	if err != nil {
		logger.Log.Warnf("获取 v4 IPv6 客户端列表失败（不影响 IPv4 结果）: %v", err)
	} else {
		for _, row := range v4Rows(ipv6Payload) {
			mac := strings.ToUpper(v4String(row, "mac", "mac_addr", "macaddr"))
			ip := v4String(row, "ip", "ip_addr", "ipaddr", "address")
			if mac == "" || ip == "" {
				continue
			}
			if entry, ok := aggregated[mac]; ok && !containsStr(entry.IPs, ip) {
				entry.IPs = append(entry.IPs, ip)
			}
		}
	}

	result := make([]ClientDTO, 0, len(aggregated))
	for _, dto := range aggregated {
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

func (s *MonitorService) v4Get(ctx context.Context, path string) (any, error) {
	var payload any
	if err := s.v4Client.Get(ctx, path, nil, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func v4Rows(value any) []map[string]any {
	switch typed := value.(type) {
	case []any:
		rows := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if row, ok := item.(map[string]any); ok {
				rows = append(rows, row)
			}
		}
		return rows
	case map[string]any:
		for _, key := range []string{"rows", "data", "list", "items", "records", "result", "results", "interfaces", "clients"} {
			if nested, ok := typed[key]; ok {
				if rows := v4Rows(nested); len(rows) > 0 {
					return rows
				}
			}
		}
		return []map[string]any{typed}
	default:
		return nil
	}
}

func v4String(row map[string]any, aliases ...string) string {
	value, ok := v4Lookup(row, aliases...)
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(toJSONString(typed)), `"`), `"`))
	}
}

func v4Int64(row map[string]any, aliases ...string) int64 {
	value, ok := v4Lookup(row, aliases...)
	if !ok {
		return 0
	}
	return v4Number(value)
}

func v4FindInt64(value any, aliases ...string) int64 {
	switch typed := value.(type) {
	case map[string]any:
		if got := v4Int64(typed, aliases...); got != 0 {
			return got
		}
		for _, nested := range typed {
			if got := v4FindInt64(nested, aliases...); got != 0 {
				return got
			}
		}
	case []any:
		for _, item := range typed {
			if got := v4FindInt64(item, aliases...); got != 0 {
				return got
			}
		}
	}
	return 0
}

func v4Lookup(row map[string]any, aliases ...string) (any, bool) {
	aliasSet := make(map[string]struct{}, len(aliases))
	for _, alias := range aliases {
		aliasSet[v4NormalizeKey(alias)] = struct{}{}
	}
	for key, value := range row {
		if _, ok := aliasSet[v4NormalizeKey(key)]; ok {
			return value, true
		}
	}
	return nil, false
}

func v4NormalizeKey(key string) string {
	key = strings.ToLower(key)
	key = strings.ReplaceAll(key, "_", "")
	key = strings.ReplaceAll(key, "-", "")
	return key
}

func v4Number(value any) int64 {
	switch typed := value.(type) {
	case int:
		return int64(typed)
	case int64:
		return typed
	case float64:
		return int64(typed)
	case json.Number:
		n, _ := typed.Int64()
		return n
	case string:
		n, _ := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		return int64(n)
	default:
		return 0
	}
}

func toJSONString(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}

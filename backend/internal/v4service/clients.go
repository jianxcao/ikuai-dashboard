package v4service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"ikuai-dashboard/backend/internal/config"
	"ikuai-dashboard/backend/internal/ikuaiv4"
)

// ClientRecordDTO represents a V4 LAN client, offline history item, or traffic summary row.
type ClientRecordDTO struct {
	ID            int      `json:"id"`
	MAC           string   `json:"mac"`
	Hostname      string   `json:"hostname"`
	IPAddr        string   `json:"ip_addr"`
	IPs           []string `json:"ips"`
	Interface     string   `json:"interface"`
	UploadSpeed   int64    `json:"upload_speed"`
	DownloadSpeed int64    `json:"download_speed"`
	TotalUp       int64    `json:"total_up"`
	TotalDown     int64    `json:"total_down"`
	Connections   int      `json:"connections"`
	Comment       string   `json:"comment"`
	ClientType    string   `json:"client_type"`
	ClientModel   string   `json:"client_model"`
	Uptime        string   `json:"uptime"`
	Downtime      string   `json:"downtime"`
	AuthType      string   `json:"auth_type"`
	Username      string   `json:"username"`
	SSID          string   `json:"ssid"`
	Signal        string   `json:"signal"`
}

// LegacyClientDTO preserves /api/v1/monitor/lan item contract.
type LegacyClientDTO struct {
	MAC           string   `json:"mac"`
	Hostname      string   `json:"hostname"`
	IPs           []string `json:"ips"`
	UploadSpeed   int64    `json:"upload_speed"`
	DownloadSpeed int64    `json:"download_speed"`
	TotalUp       int64    `json:"total_up"`
	TotalDown     int64    `json:"total_down"`
	Connections   int      `json:"connections"`
	Comment       string   `json:"comment"`
	ClientType    string   `json:"client_type"`
	Uptime        string   `json:"uptime"`
}

type ClientPageDTO struct {
	Mode     string            `json:"mode"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Total    int               `json:"total"`
	Items    []ClientRecordDTO `json:"items"`
}

type ClientDetailDTO struct {
	IP          string           `json:"ip"`
	MAC         string           `json:"mac"`
	TrafficLoad []map[string]any `json:"traffic_load"`
	AppProto    []map[string]any `json:"app_protocols"`
}

// GetClients fetches paged client collections by mode.
func GetClients(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client, mode string, page, pageSize int, search string) (*ClientPageDTO, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if mode == "" {
		mode = "online"
	}

	if router.Mock || client == nil {
		return mockClientsPage(router, mode, page, pageSize, search), nil
	}

	pathMap := map[string]string{
		"online":       "/monitoring/clients-online",
		"offline":      "/monitoring/clients-offline",
		"ipv6-online":  "/monitoring/clients-ip6-online",
		"ipv6-offline": "/monitoring/clients-ip6-offline",
		"summary":      "/monitoring/clients-traffic-summary",
	}
	apiPath, ok := pathMap[mode]
	if !ok {
		return nil, fmt.Errorf("不支持的客户端查询模式: %s", mode)
	}

	params := map[string]string{
		"page":  strconv.Itoa(page),
		"limit": strconv.Itoa(pageSize),
	}
	if strings.TrimSpace(search) != "" && mode != "summary" {
		params["key"] = "hostname,ip_addr,mac,comment"
		params["pattern"] = strings.TrimSpace(search)
	}

	raw, err := client.Get(ctx, apiPath, params)
	if err != nil {
		return nil, err
	}

	rows, total := ikuaiv4.ExtractCollection(raw)
	items := make([]ClientRecordDTO, 0, len(rows))
	for _, r := range rows {
		ipAddr := ikuaiv4.Str(r, "ip_addr", "ip", "ipaddr", "address")
		var ips []string
		if ipAddr != "" {
			ips = []string{ipAddr}
		} else {
			ips = []string{}
		}
		upSpeed := ikuaiv4.Int64(r, "upload", "upload_speed", "up_speed", "tx_speed")
		downSpeed := ikuaiv4.Int64(r, "download", "download_speed", "down_speed", "rx_speed")
		totUp := ikuaiv4.Int64(r, "total_up", "sum_total_up", "upload_total", "tx_bytes")
		totDown := ikuaiv4.Int64(r, "total_down", "sum_total_down", "download_total", "rx_bytes")

		rec := ClientRecordDTO{
			ID:            ikuaiv4.Int(r, "id"),
			MAC:           strings.ToUpper(ikuaiv4.Str(r, "mac", "mac_addr", "macaddr")),
			Hostname:      ikuaiv4.Str(r, "hostname", "host_name", "name"),
			IPAddr:        ipAddr,
			IPs:           ips,
			Interface:     ikuaiv4.Str(r, "interface", "iface"),
			UploadSpeed:   upSpeed,
			DownloadSpeed: downSpeed,
			TotalUp:       totUp,
			TotalDown:     totDown,
			Connections:   ikuaiv4.Int(r, "connect_num", "connections", "conn_count"),
			Comment:       ikuaiv4.Str(r, "comment", "remark"),
			ClientType:    ikuaiv4.Str(r, "client_type", "type"),
			ClientModel:   ikuaiv4.Str(r, "client_model", "model"),
			Uptime:        ikuaiv4.Str(r, "uptime", "online_time"),
			Downtime:      ikuaiv4.Str(r, "downtime"),
			AuthType:      ikuaiv4.Str(r, "auth_type"),
			Username:      ikuaiv4.Str(r, "username"),
			SSID:          ikuaiv4.Str(r, "ssid"),
			Signal:        ikuaiv4.Str(r, "signal"),
		}
		if search != "" && mode == "summary" {
			hay := strings.ToLower(rec.Hostname + " " + rec.IPAddr + " " + rec.MAC + " " + rec.Comment + " " + rec.Username)
			if !strings.Contains(hay, strings.ToLower(search)) {
				continue
			}
		}
		items = append(items, rec)
	}

	return &ClientPageDTO{
		Mode:     mode,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Items:    items,
	}, nil
}

// GetMergedLanClients aggregates online IPv4 + IPv6 clients by MAC for legacy compatibility.
func GetMergedLanClients(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client, search string) ([]LegacyClientDTO, error) {
	v4Page, err := GetClients(ctx, router, client, "online", 1, 500, search)
	if err != nil {
		return nil, err
	}

	byMAC := make(map[string]*LegacyClientDTO, len(v4Page.Items))
	order := make([]string, 0, len(v4Page.Items))

	for _, c := range v4Page.Items {
		mac := c.MAC
		if mac == "" {
			mac = c.IPAddr
		}
		if existing, ok := byMAC[mac]; ok {
			for _, ip := range c.IPs {
				if ip != "" && !containsString(existing.IPs, ip) {
					existing.IPs = append(existing.IPs, ip)
				}
			}
			existing.UploadSpeed += c.UploadSpeed
			existing.DownloadSpeed += c.DownloadSpeed
			existing.TotalUp += c.TotalUp
			existing.TotalDown += c.TotalDown
			existing.Connections += c.Connections
			continue
		}
		byMAC[mac] = &LegacyClientDTO{
			MAC:           c.MAC,
			Hostname:      c.Hostname,
			IPs:           append([]string(nil), c.IPs...),
			UploadSpeed:   c.UploadSpeed,
			DownloadSpeed: c.DownloadSpeed,
			TotalUp:       c.TotalUp,
			TotalDown:     c.TotalDown,
			Connections:   c.Connections,
			Comment:       c.Comment,
			ClientType:    c.ClientType,
			Uptime:        c.Uptime,
		}
		order = append(order, mac)
	}

	if v6Page, err := GetClients(ctx, router, client, "ipv6-online", 1, 500, ""); err == nil {
		for _, c6 := range v6Page.Items {
			if existing, ok := byMAC[c6.MAC]; ok {
				for _, ip := range c6.IPs {
					if ip != "" && !containsString(existing.IPs, ip) {
						existing.IPs = append(existing.IPs, ip)
					}
				}
			}
		}
	}

	out := make([]LegacyClientDTO, 0, len(order))
	for _, mac := range order {
		out = append(out, *byMAC[mac])
	}
	return out, nil
}

// GetClientDetail fetches per-client 5-minute load series and top L7 app protocols.
func GetClientDetail(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client, ip, mac string) (*ClientDetailDTO, error) {
	if strings.TrimSpace(ip) == "" || strings.TrimSpace(mac) == "" {
		return nil, fmt.Errorf("查询终端详情需要 ip 和 mac 参数")
	}
	if router.Mock || client == nil {
		return mockClientDetail(ip, mac), nil
	}

	detail := &ClientDetailDTO{
		IP:          ip,
		MAC:         mac,
		TrafficLoad: []map[string]any{},
		AppProto:    []map[string]any{},
	}

	if loadRaw, err := client.Get(ctx, "/monitoring/clients-traffic-load", map[string]string{"ip": ip, "mac": mac}); err == nil {
		detail.TrafficLoad, _ = ikuaiv4.ExtractCollection(loadRaw)
	}
	if appRaw, err := client.Get(ctx, "/monitoring/clients/app-protocols/load", map[string]string{
		"ip":    ip,
		"mac":   mac,
		"page":  "1",
		"limit": "15",
	}); err == nil {
		detail.AppProto, _ = ikuaiv4.ExtractCollection(appRaw)
	}
	return detail, nil
}

// KickOnlineSession kicks an authenticated user session offline.
func KickOnlineSession(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client, id string, dryRun bool) (map[string]any, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("需要有效的在线会话 ID")
	}
	path := "/auth/online-users/" + strings.TrimSpace(id)
	if dryRun {
		if client == nil {
			return map[string]any{"dry_run": true, "method": "DELETE", "url": "mock://api/v4.0" + path}, nil
		}
		return client.PreviewDryRun("DELETE", path, nil, nil), nil
	}
	if router.Mock || client == nil {
		return map[string]any{"message": "mock session kicked", "id": id}, nil
	}
	raw, err := client.Delete(ctx, path)
	if err != nil {
		return nil, err
	}
	return ikuaiv4.ExtractObject(raw), nil
}

func containsString(list []string, target string) bool {
	for _, s := range list {
		if s == target {
			return true
		}
	}
	return false
}

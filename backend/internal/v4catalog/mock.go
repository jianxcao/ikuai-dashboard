package v4catalog

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	mockMu          sync.Mutex
	mockCollections map[string][]map[string]any
	mockSingletons  map[string]map[string]any
	mockSeq         = 100
)

func ensureMockStore() {
	if mockCollections != nil {
		return
	}
	now := time.Now().Unix()
	mockCollections = map[string][]map[string]any{
		"wan-config": {
			{"id": 1, "tagname": "wan1", "internet": "静态IP", "ip_mask": "203.0.113.10/24", "gateway": "203.0.113.1", "mac": "bc:24:11:18:79:dc", "mtu": 1500},
			{"id": 2, "tagname": "wan2", "internet": "DHCP", "ip_mask": "198.51.100.22/24", "gateway": "198.51.100.1", "mac": "bc:24:11:a5:78:0a", "mtu": 1500},
		},
		"lan-config": {
			{"id": 1, "tagname": "lan1", "ip_mask": "192.168.9.1/24", "bandeth": "eth2", "dhcp_server": "yes", "vlan": "none"},
		},
		"wan-vlan": {
			{"id": 1, "vlan_name": "iptv_vlan", "interface": "wan2", "vlan_id": 85, "ip_mask": "10.85.1.10/24", "enabled": "yes"},
		},
		"physical-nics": {
			{"name": "eth0", "interface": "wan1", "link": 1, "speed": 2500, "mac": "bc:24:11:18:79:dc", "driver": "igc", "model": "Intel I225-V 2.5GbE"},
			{"name": "eth1", "interface": "wan2", "link": 1, "speed": 1000, "mac": "bc:24:11:a5:78:0a", "driver": "igc", "model": "Intel I225-V 2.5GbE"},
			{"name": "eth2", "interface": "lan1", "link": 1, "speed": 10000, "mac": "bc:24:11:a9:bc:78", "driver": "ixgbe", "model": "Intel X520 10GbE"},
		},
		"dns-proxy": {
			{"id": 1, "domain": "internal.corp", "dns_addr": "192.168.9.2", "parse_type": "ipv4", "comment": "内网域名解析", "enabled": "yes"},
			{"id": 2, "domain": "registry.dev", "dns_addr": "223.5.5.5", "parse_type": "ipv4", "comment": "镜像加速解析", "enabled": "yes"},
		},
		"dhcp-services": {
			{"id": 1, "tagname": "核心内网池", "interface": "lan1", "phy_ifnames": "all", "addr_pool": "192.168.9.100-192.168.9.230", "netmask": "255.255.255.0", "gateway": "192.168.9.1", "lease": 120, "enabled": "yes"},
		},
		"dhcp-static": {
			{"id": 1, "tagname": "synology-nas", "ip_addr": "192.168.9.10", "mac": "8E:F6:C5:99:8A:50", "interface": "lan1", "comment": "企业存储服务器", "enabled": "yes"},
			{"id": 2, "tagname": "pve-cluster-01", "ip_addr": "192.168.9.20", "mac": "EA:6C:25:F5:D7:43", "interface": "lan1", "comment": "虚拟化宿主机", "enabled": "yes"},
		},
		"dhcp-leases": {
			{"id": 1, "hostname": "workstation-01", "ip_addr": "192.168.9.101", "mac": "00:22:1F:97:F2:CD", "interface": "lan1", "timeout": "118m"},
			{"id": 2, "hostname": "macbook-pro", "ip_addr": "192.168.9.103", "mac": "00:70:FA:3D:30:10", "interface": "lan1", "timeout": "95m"},
		},
		"vlan": {
			{"id": 1, "vlan_name": "IoT_Devices", "vlan_id": 20, "interface": "lan1", "ip_addr": "10.20.0.1", "netmask": "255.255.255.0", "enabled": "yes"},
			{"id": 2, "vlan_name": "Guest_WiFi", "vlan_id": 30, "interface": "lan1", "ip_addr": "10.30.0.1", "netmask": "255.255.255.0", "enabled": "yes"},
		},
		"dnat-rules": {
			{"id": 1, "tagname": "NAS-HTTPS", "interface": "wan1", "protocol": "tcp", "wan_port": "5001", "lan_addr": "192.168.9.10", "lan_port": "5001", "enabled": "yes", "comment": "NAS 安全访问"},
			{"id": 2, "tagname": "WireGuard-UDP", "interface": "wan1", "protocol": "udp", "wan_port": "51820", "lan_addr": "192.168.9.1", "lan_port": "51820", "enabled": "yes", "comment": "VPN 入口"},
		},
		"nat-rules": {
			{"id": 1, "tagname": "LAN_SNAT", "action": "masquerade", "iinterface": "lan1", "ointerface": "wan1", "protocol": "any", "src_addr": map[string]any{"custom": []any{"192.168.9.0/24"}}, "enabled": "yes"},
		},
		"dmz-rules": {},
		"static-routes": {
			{"id": 1, "tagname": "lab-subnet", "dst_addr": "10.10.0.0", "netmask": "255.255.0.0", "gateway": "192.168.9.2", "interface": "lan1", "prio": 10, "enabled": "yes"},
		},
		"stream-domain": {
			{"id": 1, "tagname": "cdn-direct", "domain": map[string]any{"custom": []any{"example.com", "cdn.example.net"}}, "src_addr": map[string]any{"custom": []any{}}, "interface": "wan1", "prio": 30, "enabled": "yes"},
		},
		"stream-five-tuple": {
			{"id": 1, "tagname": "ssh-wan2", "protocol": "tcp", "src_addr": map[string]any{"custom": []any{}}, "dst_port": map[string]any{"custom": []any{"22"}}, "interface": "wan2", "prio": 20, "enabled": "yes"},
		},
		"stream-l7": {
			{"id": 1, "tagname": "dns-line1", "app_proto": map[string]any{"custom": []any{"DNS"}}, "interface": "wan1", "prio": 31, "enabled": "yes"},
		},
		"stream-load-balance": {
			{"id": 1, "tagname": "wan1-weight", "interface": "wan1", "mode": 1, "weight": "2", "isp_name": "电信", "enabled": "yes"},
			{"id": 2, "tagname": "wan2-weight", "interface": "wan2", "mode": 1, "weight": "1", "isp_name": "联通", "enabled": "yes"},
		},
		"stream-updown": {},
		"qos-ip": {
			{"id": 1, "tagname": "guest-limit", "ip_addr": map[string]any{"custom": []any{"10.30.0.0/24"}}, "interface": "wan1", "upload": 2000, "download": 10000, "enabled": "yes"},
		},
		"qos-mac": {},
		"acl-rules": {
			{"id": 1, "tagname": "block-telnet", "action": "drop", "dir": "forward", "protocol": "tcp", "dst_port": map[string]any{"custom": []any{"23"}}, "prio": 10, "enabled": "yes"},
		},
		"mac-rules": {
			{"id": 1, "tagname": "trusted-server", "mac": "8E:F6:C5:99:8A:50", "comment": "核心 NAS", "enabled": "yes"},
		},
		"l7-rules":         {},
		"url-black":        {{"id": 1, "tagname": "tracker-block", "domain": "ads.example.com", "mode": 0, "enabled": "yes"}},
		"domain-blacklist": {},
		"peerconn-rules": {
			{"id": 1, "tagname": "office-limit", "limits": 800, "protocol": "tcp", "src_addr": map[string]any{"custom": []any{"192.168.9.0/24"}}, "enabled": "yes"},
		},
		"terminals": {
			{"id": 1, "tagname": "群晖存储", "mac": "8E:F6:C5:99:8A:50", "comment": "机房机柜 A-02"},
		},
		"auth-online": {
			{"id": 1, "username": "dev_alice", "ppptype": "pppoe", "ip_addr": "192.168.9.140", "mac": "AA:11:22:33:44:55", "interface": "lan1", "auth_time": now - 3600},
		},
		"auth-accounts": {
			{"id": 1, "username": "dev_alice", "ppptype": "any", "upload": 0, "download": 0, "share": 2, "enabled": "yes"},
			{"id": 2, "username": "vpn_bob", "ppptype": "l2tp", "upload": 5000, "download": 10000, "share": 1, "enabled": "yes"},
		},
		"auth-packages": {
			{"id": 1, "packname": "研发千兆不限", "packtime": "1m", "price": 0, "up_speed": 100000, "down_speed": 100000, "comment": "内部员工套餐"},
		},
		"wireguard": {
			{"id": 1, "name": "wg_hq_mesh", "local_address": "10.9.0.1/24", "interface": "auto", "local_listenport": 51820, "mtu": 1420, "enabled": "yes"},
		},
		"ipsec-clients": {
			{"id": 1, "name": "ipsec_lab", "remote_addr": "198.51.100.88", "leftsubnet": "192.168.9.0/24", "rightsubnet": "10.10.0.0/24", "enabled": "yes"},
		},
		"openvpn-clients":   {},
		"l2tp-clients":      {},
		"pptp-clients":      {},
		"ikev2-clients":     {},
		"dhcp-access-rules": {},
		"url-keywords":      {},
		"url-redirect":      {},
		"url-replace":       {},
		"ftp-users":         {},
		"http-users":        {},
		"samba-users":       {},
		"wireless-aps": {
			{"id": 1, "tagname": "Office-AP-01", "mac": "C0:25:E9:11:22:33", "ip_addr": "192.168.9.31", "ssid1": "iKuai_5G", "status": "online"},
		},
		"wireless-blacklist": {},
		"wireless-vlan":      {},
		"ip-objects": {
			{"id": 1, "group_name": "core_servers", "group_value": []any{map[string]any{"ip": "192.168.9.10"}, map[string]any{"ip": "192.168.9.20"}}, "value": "192.168.9.10,192.168.9.20", "comment": "核心服务器组"},
		},
		"ip6-objects":    {},
		"mac-objects":    {},
		"port-objects":   {{"id": 1, "group_name": "web_ports", "group_value": []any{map[string]any{"port": "80,443"}}, "value": "80,443", "comment": "Web 标准端口"}},
		"domain-objects": {},
		"proto-objects":  {{"id": 1, "group_name": "web_protos", "group_value": []any{map[string]any{"proto": "tcp"}, map[string]any{"proto": "udp"}}, "value": "tcp,udp", "comment": "基础传输协议"}},
		"time-objects":   {{"id": 1, "group_name": "work_hours", "type": "weekly", "weekdays": "12345", "start_time": "09:00", "end_time": "18:00", "value": "weekly 12345 09:00~18:00", "comment": "工作日办公时段"}},
		"log-system": {
			{"id": 1, "timestamp": now - 120, "content": "系统运行正常：v4.0 API Token 服务就绪"},
			{"id": 2, "timestamp": now - 600, "content": "WAN1 链路检测通过 (RTT 4.4ms)"},
			{"id": 3, "timestamp": now - 1800, "content": "WireGuard 隧道 wg_hq_mesh 握手完成"},
		},
		"log-arp":        {{"id": 1, "timestamp": now - 300, "content": "ARP 绑定校验通过: 192.168.9.10 -> 8E:F6:C5:99:8A:50"}},
		"log-auth":       {{"id": 1, "timestamp": now - 900, "username": "dev_alice", "ppptype": "pppoe", "ip_addr": "192.168.9.140", "result": "success", "event": "登录成功"}},
		"log-dhcp":       {{"id": 1, "timestamp": now - 240, "msgtype": "DHCPACK", "interface": "lan1", "ip_addr": "192.168.9.103", "mac": "00:70:FA:3D:30:10", "hostname": "macbook-pro"}},
		"log-pppoe":      {{"id": 1, "timestamp": now - 1500, "interface": "wan1", "content": "LCP Echo 链路保活正常"}},
		"log-web":        {{"id": 1, "timestamp": now - 400, "username": "admin", "ip_addr": "192.168.9.101", "function": "API Token", "event": "读取监控概览"}},
		"log-ddns":       {},
		"log-notice":     {{"id": 1, "timestamp": now - 2000, "type": "system", "ip_addr": "192.168.9.1", "event": "多线负载均衡权重已同步"}},
		"log-wireless":   {{"id": 1, "timestamp": now - 500, "action": "assoc", "ssid": "iKuai_5G", "mac": "00:70:FA:3D:30:10", "signal": "-52"}},
		"log-url-visits": {{"id": 1, "timestamp": now - 100, "ip_addr": "192.168.9.101", "host": "github.com", "appname": "HTTPS", "uri": "/ikuaidev/ikuai-cli"}},
		"reboot-schedules": {
			{"id": 1, "tagname": "weekly_maintenance", "strategy": "week", "cycle_time": "7", "time": "04:00", "enabled": "no"},
		},
		"backups": {
			{"id": 1, "filename": "ikuai-v4-20261124.bak", "version": "4.0.314", "filesize": 184320, "timestamp": now - 86400},
		},
		"disks": {
			{"disk": "/dev/sda", "model": "QEMU HARDDISK", "size": "16 GB", "type": "SSD", "system": "1"},
		},
		"system-files": {
			{"f_name": "docker", "st_type": 0, "st_size": 4096, "st_mtime": now - 7200},
			{"f_name": "backup", "st_type": 0, "st_size": 4096, "st_mtime": now - 86400},
		},
		"web-admin-groups": {
			{"id": 1, "group_name": "Administrators", "ip_addr": "0.0.0.0/0", "perm_config": ""},
		},
		"web-admin-accounts": {
			{"id": 1, "username": "admin", "group_id": 1, "sesstimeout": 3600, "enabled": "yes", "comment": "超级管理员"},
		},
	}

	mockSingletons = map[string]map[string]any{
		"dns-config":        {"dns1": "223.5.5.5", "dns2": "119.29.29.29", "cache_ttl": 300, "proxy_force": "0", "enabled": "yes"},
		"dhcp-access-mode":  {"mode": 0},
		"pppoe-config":      {"enabled": "no", "server_name": "iKuai", "interface": "lan1", "mtu": 1480, "mru": 1480},
		"mac-mode":          {"acl_mac": 0},
		"security-advanced": {"noping_lan": "0", "noping_wan": "1", "notracert": "0", "hijack_ping": "0", "invalid": "0", "dos_lan": "1", "dos_lan_num": 1500, "tcp_mss": "0", "tcp_mss_num": 1400},
		"secondary-route":   {"nol2rt": 0, "ttl_num": 1, "nol2rt_ip": map[string]any{"custom": []any{}, "object": []any{}}},
		"auth-web-server":   {"enabled": "no", "interface": "lan1", "idle_time": 60, "max_time": 0, "user_auth": 0, "nopasswd": 0, "https_redirect": 0},
		"openvpn-service":   {"enabled": "no", "port": "1194", "proto": "udp", "subnet": "10.7.7.0", "mask": "255.255.255.0", "tun_mtu": "1400", "cipher": "BF-CBC"},
		"pptp-service":      {"enabled": "no", "server_ip": "10.0.0.1", "server_port": 1723, "addr_pool": "10.0.0.2-10.0.0.254", "dns1": "114.114.114.114", "dns2": "119.29.29.29", "open_mppe": 2, "mtu": 1400, "mru": 1400},
		"l2tp-service":      {"enabled": "no", "server_ip": "10.1.0.1", "server_port": 1701, "addr_pool": "10.1.0.2-10.1.0.254", "ipsec_secret": "", "dns1": "114.114.114.114", "dns2": "119.29.29.29", "mtu": 1400, "mru": 1400},
		"wireless-ac":       {"enabled": "yes", "version": "AC 4.0", "online_ap": 1},
		"ftp-config":        {"open_ftp": "0", "ftp_port": 21, "ftp_access": "0"},
		"samba-config":      {"enabled": "no", "workgroup": "WORKGROUP", "wsdd2": "1", "access": "0"},
		"snmpd-config":      {"enabled": "yes", "listen_port": 161, "version": "2", "community": "public", "rw": "ro", "sysname": "iKuai-V4", "syslocation": "Rack-A", "syscontact": "ops@local"},
		"system-basic":      {"hostname": "iKuai-HQ-Core", "time_zone": 8, "switch_ntp": "1", "ntpserver_list": "ntp.aliyun.com", "fast_nat": "1"},
		"remote-access":     {"open_sshd": "1", "sshd_port": 22, "open_wanweb": "0", "http_port": 80, "https_port": 443, "force_https": "1"},
		"kernel-params":     {"bbr": "1", "established_timeout": 3600, "time_wait_timeout": 120, "udp_timeout": 30},
		"cpufreq-mode":      {"cpufreq_support": 1, "mode": "ondemand", "current_cpufreq": "ondemand", "turbo": "1", "current_turbo": 1},
		"alg-config":        {"support_ftp": "1", "support_sip": "1", "support_tftp": "1", "support_h323": "0"},
		"vrrp-config":       {"enabled": "no", "type": "1", "prio": 150, "gateway": "192.168.9.1", "virtual_ips": "192.168.9.254"},
		"backup-auto":       {"id": 1, "enabled": "no", "strategy": "week", "cycle_time": "1234567", "time": "23:59", "valid_days": 30},
		"upgrade-info":      {"system_ver": "4.0.314-beta", "build_date": "202610131658", "new_system_ver": "4.0.314-beta", "firmware_channel": "beta"},
	}
}

var mockWGPeers = map[string][]map[string]any{
	"1": {
		{"id": 1, "interface": "wg_hq_mesh", "peer_publickey": "K9mockPeerPubKey1234567890abcdef=", "allowips": "10.9.0.2/32", "endpoint": "203.0.113.45", "endpoint_port": 51820, "keepalive": 25, "comment": "分支节点对端", "enabled": "yes"},
	},
}

func mockListOrGet(spec ResourceSpec, page, pageSize int, search string, extraQuery ...map[string]string) *ListResult {
	mockMu.Lock()
	defer mockMu.Unlock()
	ensureMockStore()

	if spec.Kind == "singleton" {
		cfg := mockSingletons[spec.Key]
		if cfg == nil {
			cfg = map[string]any{}
		}
		return &ListResult{
			Resource: spec,
			Rows:     []map[string]any{cfg},
			Total:    1,
			Config:   cfg,
		}
	}

	rows := mockCollections[spec.Key]
	filtered := make([]map[string]any, 0, len(rows))
	needle := strings.ToLower(strings.TrimSpace(search))
	for _, row := range rows {
		if needle != "" {
			matched := false
			for _, v := range row {
				if strings.Contains(strings.ToLower(fmt.Sprint(v)), needle) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		filtered = append(filtered, row)
	}

	return &ListResult{
		Resource: spec,
		Rows:     filtered,
		Total:    len(filtered),
	}
}

func mockCreate(spec ResourceSpec, payload map[string]any) map[string]any {
	mockMu.Lock()
	defer mockMu.Unlock()
	ensureMockStore()

	mockSeq++
	row := make(map[string]any, len(payload)+1)
	for k, v := range payload {
		row[k] = v
	}
	row["id"] = mockSeq
	if spec.ObjectValueKey != "" {
		row["value"] = formatMockObjValue(row["group_value"], spec.ObjectValueKey)
	}
	if spec.Key == "time-objects" {
		flattenTimeGroupValue(row)
	}
	mockCollections[spec.Key] = append([]map[string]any{row}, mockCollections[spec.Key]...)
	return map[string]any{"message": "created", "rowid": mockSeq}
}

func mockUpdatePreview(spec ResourceSpec, id string, rawInput, updates map[string]any) map[string]any {
	mockMu.Lock()
	defer mockMu.Unlock()
	ensureMockStore()

	var existing map[string]any
	if spec.Kind == "singleton" {
		existing = mockSingletons[spec.Key]
	} else {
		for _, r := range mockCollections[spec.Key] {
			if fmt.Sprint(r["id"]) == id {
				existing = r
				break
			}
		}
	}
	if existing == nil {
		existing = map[string]any{}
	}
	return buildMergedUpdateBody(spec, id, rawInput, updates, existing)
}

func mockUpdate(spec ResourceSpec, id string, rawInput, updates map[string]any) map[string]any {
	mockMu.Lock()
	defer mockMu.Unlock()
	ensureMockStore()

	if spec.Kind == "singleton" {
		cur := mockSingletons[spec.Key]
		if cur == nil {
			cur = map[string]any{}
		}
		mockSingletons[spec.Key] = buildMergedUpdateBody(spec, id, rawInput, updates, cur)
		return map[string]any{"message": "updated"}
	}

	rows := mockCollections[spec.Key]
	for i, r := range rows {
		if fmt.Sprint(r["id"]) == id {
			merged := buildMergedUpdateBody(spec, id, rawInput, updates, r)
			if r["id"] != nil {
				merged["id"] = r["id"]
			}
			if spec.ObjectValueKey != "" {
				merged["value"] = formatMockObjValue(merged["group_value"], spec.ObjectValueKey)
			}
			if spec.Key == "time-objects" {
				flattenTimeGroupValue(merged)
			}
			rows[i] = merged
			break
		}
	}
	return map[string]any{"message": "updated", "id": id}
}

func mockToggle(spec ResourceSpec, id, val string) map[string]any {
	mockMu.Lock()
	defer mockMu.Unlock()
	ensureMockStore()

	for _, r := range mockCollections[spec.Key] {
		if fmt.Sprint(r["id"]) == id {
			r["enabled"] = val
			break
		}
	}
	return map[string]any{"message": "toggled", "id": id, "enabled": val}
}

func mockDelete(spec ResourceSpec, id string) map[string]any {
	mockMu.Lock()
	defer mockMu.Unlock()
	ensureMockStore()

	if spec.DeleteMode == "clear" {
		mockCollections[spec.Key] = []map[string]any{}
		return map[string]any{"message": "cleared"}
	}

	rows := mockCollections[spec.Key]
	next := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		if spec.DeleteMode == "srcfile" {
			if fmt.Sprint(r["filename"]) == id {
				continue
			}
		} else if fmt.Sprint(r["id"]) == id {
			continue
		}
		next = append(next, r)
	}
	mockCollections[spec.Key] = next
	return map[string]any{"message": "deleted", "id": id}
}

func formatMockObjValue(raw any, key string) string {
	arr, ok := raw.([]map[string]any)
	if !ok {
		return ""
	}
	vals := make([]string, 0, len(arr))
	for _, item := range arr {
		if v, exists := item[key]; exists {
			vals = append(vals, fmt.Sprint(v))
		}
	}
	return strings.Join(vals, ",")
}

func mockObjectRefs(key, groupName string) map[string]any {
	rows := []map[string]any{}
	if groupName == "core_servers" {
		rows = append(rows, map[string]any{"module": "acl-rules", "rule": "allow_servers", "group": key})
	}
	return map[string]any{"group_name": groupName, "rows": rows, "total": len(rows)}
}

func mockListWGPeers(wgID string) map[string]any {
	mockMu.Lock()
	defer mockMu.Unlock()
	peers := mockWGPeers[wgID]
	if peers == nil {
		peers = []map[string]any{}
	}
	return map[string]any{"rows": peers, "total": len(peers)}
}

func mockCreateWGPeer(wgID string, payload map[string]any) map[string]any {
	mockMu.Lock()
	defer mockMu.Unlock()
	mockSeq++
	row := map[string]any{"id": mockSeq}
	for k, v := range payload {
		row[k] = v
	}
	mockWGPeers[wgID] = append([]map[string]any{row}, mockWGPeers[wgID]...)
	return map[string]any{"message": "created", "rowid": mockSeq}
}

func mockUpdateWGPeer(wgID, peerID string, updates map[string]any) map[string]any {
	mockMu.Lock()
	defer mockMu.Unlock()
	for _, r := range mockWGPeers[wgID] {
		if fmt.Sprint(r["id"]) == peerID {
			for k, v := range updates {
				r[k] = v
			}
			break
		}
	}
	return map[string]any{"message": "updated", "id": peerID}
}

func mockToggleWGPeer(wgID, peerID, val string) map[string]any {
	mockMu.Lock()
	defer mockMu.Unlock()
	for _, r := range mockWGPeers[wgID] {
		if fmt.Sprint(r["id"]) == peerID {
			r["enabled"] = val
			break
		}
	}
	return map[string]any{"message": "toggled", "id": peerID, "enabled": val}
}

func mockDeleteWGPeer(wgID, peerID string) map[string]any {
	mockMu.Lock()
	defer mockMu.Unlock()
	next := make([]map[string]any, 0, len(mockWGPeers[wgID]))
	for _, r := range mockWGPeers[wgID] {
		if fmt.Sprint(r["id"]) != peerID {
			next = append(next, r)
		}
	}
	mockWGPeers[wgID] = next
	return map[string]any{"message": "deleted", "id": peerID}
}

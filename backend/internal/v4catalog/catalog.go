package v4catalog

import "sync"

var (
	catalogOnce sync.Once
	catalogList []ResourceSpec
	catalogByID map[string]ResourceSpec
)

func weeklyTimeDefault() map[string]any {
	return map[string]any{
		"custom": []any{
			map[string]any{
				"type":       "weekly",
				"weekdays":   "1234567",
				"start_time": "00:00",
				"end_time":   "23:59",
				"comment":    "",
			},
		},
		"object": []any{},
	}
}

func emptyCustomObj() map[string]any {
	return map[string]any{
		"custom": []any{},
		"object": []any{},
	}
}

func enabledField() FieldSpec {
	return FieldSpec{
		Key: "enabled", Label: "启用状态", Type: "switch", Default: "yes",
		Options: []OptionItem{{Label: "启用", Value: "yes"}, {Label: "禁用", Value: "no"}},
	}
}

func commentField() FieldSpec {
	return FieldSpec{Key: "comment", Label: "备注", Type: "text", Placeholder: "可选说明"}
}

func initCatalog() {
	catalogOnce.Do(func() {
		var all []ResourceSpec
		all = append(all, networkSpecs()...)
		all = append(all, routingAndQosSpecs()...)
		all = append(all, securitySpecs()...)
		all = append(all, usersAndVpnSpecs()...)
		all = append(all, wirelessObjectsAdvancedSpecs()...)
		all = append(all, logAndSystemSpecs()...)

		catalogList = all
		catalogByID = make(map[string]ResourceSpec, len(all))
		for _, item := range all {
			catalogByID[item.Key] = item
		}
	})
}

// All returns all V4 resource specifications.
func All() []ResourceSpec {
	initCatalog()
	out := make([]ResourceSpec, len(catalogList))
	copy(out, catalogList)
	return out
}

// Lookup looks up a resource spec by key.
func Lookup(key string) (ResourceSpec, bool) {
	initCatalog()
	spec, ok := catalogByID[key]
	return spec, ok
}

func networkSpecs() []ResourceSpec {
	return []ResourceSpec{
		{
			Key: "wan-config", Group: "network", Label: "WAN 接口配置", Kind: "collection",
			V4Path: "/interfaces/wan-config", CLICommand: "ikuai-cli network wan list",
			Description: "查看各 WAN 口拨号方式、IP/掩码、网关、MAC 及 MTU 配置。",
			Operations:  []string{"list"},
			Columns: []ColumnSpec{
				{Key: "id", Label: "ID", Kind: "mono"},
				{Key: "tagname", Label: "接口", Kind: "badge"},
				{Key: "internet", Label: "接入方式"},
				{Key: "ip_mask", Label: "IP/掩码", Kind: "mono"},
				{Key: "gateway", Label: "网关", Kind: "mono"},
				{Key: "mac", Label: "MAC", Kind: "mono"},
				{Key: "mtu", Label: "MTU", Kind: "mono"},
			},
		},
		{
			Key: "lan-config", Group: "network", Label: "LAN 接口配置", Kind: "collection",
			V4Path: "/interfaces/lan-config", CLICommand: "ikuai-cli network lan list",
			Description: "查看内网 LAN 接口 IP 段、绑定网卡、DHCP 与 VLAN 状态。",
			Operations:  []string{"list"},
			Columns: []ColumnSpec{
				{Key: "id", Label: "ID", Kind: "mono"},
				{Key: "tagname", Label: "接口", Kind: "badge"},
				{Key: "ip_mask", Label: "IP/掩码", Kind: "mono"},
				{Key: "bandeth", Label: "绑定网口", Kind: "mono"},
				{Key: "dhcp_server", Label: "DHCP 服务"},
				{Key: "vlan", Label: "VLAN"},
			},
		},
		{
			Key: "wan-vlan", Group: "network", Label: "WAN VLAN 配置", Kind: "collection",
			V4Path: "/interfaces/wan-vlan-config", CLICommand: "ikuai-cli network wan-vlan list",
			Description: "查看 WAN 口子接口与多拨 VLAN 参数。",
			Operations:  []string{"list"},
			Columns: []ColumnSpec{
				{Key: "id", Label: "ID", Kind: "mono"},
				{Key: "vlan_name", Label: "名称"},
				{Key: "interface", Label: "接口", Kind: "badge"},
				{Key: "vlan_id", Label: "VLAN ID", Kind: "mono"},
				{Key: "ip_mask", Label: "IP/掩码", Kind: "mono"},
				{Key: "enabled", Label: "状态", Kind: "badge"},
			},
		},
		{
			Key: "physical-nics", Group: "network", Label: "物理网口信息", Kind: "collection",
			V4Path: "/interfaces/physical", CLICommand: "ikuai-cli network physical list",
			Description: "查看物理以太网口驱动、芯片型号、双工速率及绑定状态。",
			Operations:  []string{"list"},
			Columns: []ColumnSpec{
				{Key: "name", Label: "网口", Kind: "mono"},
				{Key: "interface", Label: "绑定角色", Kind: "badge"},
				{Key: "link", Label: "链路", Kind: "badge"},
				{Key: "speed", Label: "速率(Mbps)", Kind: "mono"},
				{Key: "mac", Label: "MAC", Kind: "mono"},
				{Key: "driver", Label: "驱动", Kind: "mono"},
				{Key: "model", Label: "型号"},
			},
		},
		{
			Key: "dns-config", Group: "network", Label: "DNS 全局配置", Kind: "singleton",
			V4Path: "/network/dns/config", CLICommand: "ikuai-cli network dns get",
			Description: "管理路由器上游 DNS、缓存模式、强制代理与 TTL。",
			Operations:  []string{"get", "update"},
			Columns: []ColumnSpec{
				{Key: "dns1", Label: "主 DNS", Kind: "mono"},
				{Key: "dns2", Label: "备 DNS", Kind: "mono"},
				{Key: "cachemode", Label: "缓存模式"},
				{Key: "cache_ttl", Label: "缓存 TTL"},
				{Key: "enabled", Label: "状态", Kind: "badge"},
			},
			Fields: []FieldSpec{
				{Key: "dns1", Label: "主 DNS", Type: "text", Required: true, Placeholder: "223.5.5.5"},
				{Key: "dns2", Label: "备用 DNS", Type: "text", Placeholder: "119.29.29.29"},
				{Key: "cache_ttl", Label: "缓存 TTL(秒)", Type: "number", Default: 300},
				{Key: "proxy_force", Label: "强制 DNS 代理", Type: "select", Options: []OptionItem{{Label: "关闭 (0)", Value: "0"}, {Label: "开启 (1)", Value: "1"}}},
				enabledField(),
			},
			UpdateFields: []string{"enabled", "forbid_dns_4a", "cache_ttl", "cachemode", "proxy_force", "proxy_force_dns", "dns1", "dns2", "query", "defense", "network", "query_args_ip", "query_head_ip"},
		},
		{
			Key: "dns-proxy", Group: "network", Label: "DNS 代理分流规则", Kind: "collection",
			V4Path: "/network/dns/proxy/rules", CLICommand: "ikuai-cli network dns proxy list",
			Description:    "指定域名走特定上游 DNS 服务器解析（支持 IPv4/IPv6）。",
			Operations:     []string{"list", "create", "update", "delete"},
			SupportsPaging: true, SearchKeys: "domain,dns_addr,comment",
			Columns: []ColumnSpec{
				{Key: "id", Label: "ID", Kind: "mono"},
				{Key: "domain", Label: "域名", Kind: "mono"},
				{Key: "dns_addr", Label: "指定 DNS", Kind: "mono"},
				{Key: "parse_type", Label: "解析类型", Kind: "badge"},
				{Key: "comment", Label: "备注"},
				{Key: "enabled", Label: "启用", Kind: "badge"},
			},
			Fields: []FieldSpec{
				{Key: "domain", Label: "域名", Type: "text", Required: true, Placeholder: "example.com"},
				{Key: "dns_addr", Label: "DNS 服务器", Type: "text", Required: true, Placeholder: "8.8.8.8"},
				{Key: "parse_type", Label: "解析类型", Type: "select", Default: "ipv4", Options: []OptionItem{
					{Label: "IPv4", Value: "ipv4"}, {Label: "IPv6", Value: "ipv6"}, {Label: "Proxy", Value: "proxy"}, {Label: "Proxy6", Value: "proxy6"},
				}},
				enabledField(),
				commentField(),
			},
			Defaults:     map[string]any{"enabled": "yes", "parse_type": "ipv4", "is_ipv6": 0, "src_addr": "", "comment": ""},
			UpdateFields: []string{"domain", "dns_addr", "src_addr", "parse_type", "is_ipv6", "comment", "enabled"},
		},
		{
			Key: "dhcp-services", Group: "network", Label: "DHCP 服务端地址池", Kind: "collection",
			V4Path: "/network/dhcp/services", CLICommand: "ikuai-cli network dhcp list",
			Description:    "管理各 LAN 口 DHCP 地址池、网关、租约及 DNS 下发。",
			Operations:     []string{"list", "create", "update", "toggle", "delete"},
			SupportsPaging: true, SearchKeys: "tagname,interface,addr_pool",
			Actions: []ActionSpec{
				{Key: "dhcp-start", Label: "启动 DHCP", Path: "/network/dhcp/services:start", Method: "POST"},
				{Key: "dhcp-restart", Label: "重启 DHCP 服务", Path: "/network/dhcp/services:restart", Method: "POST"},
				{Key: "dhcp-stop", Label: "停止 DHCP", Path: "/network/dhcp/services:stop", Method: "POST", Dangerous: true, ConfirmText: "确认停止 DHCP 地址池分配服务吗？"},
			},
			Columns: []ColumnSpec{
				{Key: "id", Label: "ID", Kind: "mono"},
				{Key: "tagname", Label: "名称"},
				{Key: "interface", Label: "接口", Kind: "badge"},
				{Key: "addr_pool", Label: "地址池", Kind: "mono"},
				{Key: "gateway", Label: "网关", Kind: "mono"},
				{Key: "lease", Label: "租期(分)", Kind: "mono"},
				{Key: "enabled", Label: "启用", Kind: "toggle"},
			},
			Fields: []FieldSpec{
				{Key: "tagname", Label: "名称", Type: "text", Required: true, Placeholder: "Office"},
				{Key: "interface", Label: "绑定接口", Type: "text", Required: true, Default: "lan1"},
				{Key: "phy_ifnames", Label: "物理接口", Type: "text", Default: "all", Placeholder: "all 或 eth2"},
				{Key: "addr_pool", Label: "地址池范围", Type: "text", Required: true, Placeholder: "192.168.9.100-192.168.9.200"},
				{Key: "netmask", Label: "子网掩码", Type: "text", Required: true, Default: "255.255.255.0"},
				{Key: "gateway", Label: "默认网关", Type: "text", Required: true, Placeholder: "192.168.9.1"},
				{Key: "dns1", Label: "主 DNS", Type: "text", Placeholder: "223.5.5.5"},
				{Key: "dns2", Label: "备 DNS", Type: "text", Placeholder: "119.29.29.29"},
				{Key: "lease", Label: "租期(分钟)", Type: "number", Default: 120},
				enabledField(),
			},
			Defaults: map[string]any{"enabled": "yes", "phy_ifnames": "all", "delay": 0, "check_addr_valid": 0, "check_relay_only": 0, "lease": 120},
			UpdateFields: []string{
				"enabled", "tagname", "interface", "phy_ifnames", "addr_pool", "exclude_pool",
				"netmask", "gateway", "dns1", "dns2", "wins1", "wins2", "domain", "next_server",
				"lease", "delay", "opt_type15", "opt15", "opt_type28", "opt28",
				"opt_type43", "opt43", "opt_type60", "opt60", "opt_type66", "opt66",
				"opt_type67", "opt67", "opt_type80", "opt80", "opt_type119", "opt119",
				"opt_type125", "opt125", "opt_type128", "opt128", "opt_type138", "opt138",
				"opt_type121", "opt121", "check_addr_valid", "check_relay_only",
			},
		},
		{
			Key: "dhcp-static", Group: "network", Label: "DHCP 静态分配", Kind: "collection",
			V4Path: "/network/dhcp/static", CLICommand: "ikuai-cli network dhcp static list",
			Description:    "为指定 MAC 地址设备绑定固定内网 IP。",
			Operations:     []string{"list", "create", "update", "toggle", "delete"},
			SupportsPaging: true, SearchKeys: "tagname,ip_addr,mac,comment",
			Columns: []ColumnSpec{
				{Key: "id", Label: "ID", Kind: "mono"},
				{Key: "tagname", Label: "名称"},
				{Key: "ip_addr", Label: "固定 IP", Kind: "mono"},
				{Key: "mac", Label: "MAC 地址", Kind: "mono"},
				{Key: "interface", Label: "接口", Kind: "badge"},
				{Key: "comment", Label: "备注"},
				{Key: "enabled", Label: "启用", Kind: "toggle"},
			},
			Fields: []FieldSpec{
				{Key: "tagname", Label: "设备名称", Type: "text", Required: true, Placeholder: "NAS"},
				{Key: "ip_addr", Label: "固定 IP", Type: "text", Required: true, Placeholder: "192.168.9.10"},
				{Key: "mac", Label: "MAC 地址", Type: "text", Required: true, Placeholder: "AA:BB:CC:DD:EE:FF"},
				{Key: "interface", Label: "所属接口", Type: "text", Required: true, Default: "lan1"},
				{Key: "gateway", Label: "网关", Type: "text", Placeholder: "192.168.9.1"},
				{Key: "dns1", Label: "主 DNS", Type: "text"},
				enabledField(),
				commentField(),
			},
			Defaults:     map[string]any{"enabled": "yes", "comment": ""},
			UpdateFields: []string{"enabled", "mac", "ip_addr", "interface", "gateway", "dns1", "dns2", "comment", "tagname", "hostname"},
		},
		{
			Key: "dhcp-leases", Group: "network", Label: "DHCP 动态租约", Kind: "collection",
			V4Path: "/network/dhcp/clients", CLICommand: "ikuai-cli network dhcp clients",
			Description:    "查看当前 DHCP 已分配 IP 租约和剩余过期时间，支持一键转为静态绑定。",
			Operations:     []string{"list"},
			SupportsPaging: true, SearchKeys: "hostname,ip_addr,mac",
			Columns: []ColumnSpec{
				{Key: "id", Label: "ID", Kind: "mono"},
				{Key: "hostname", Label: "主机名"},
				{Key: "ip_addr", Label: "IP 地址", Kind: "mono"},
				{Key: "mac", Label: "MAC 地址", Kind: "mono"},
				{Key: "interface", Label: "接口", Kind: "badge"},
				{Key: "timeout", Label: "剩余租期", Kind: "mono"},
			},
		},
		{
			Key: "dhcp-access-mode", Group: "network", Label: "DHCP 准入模式", Kind: "singleton",
			V4Path: "/network/dhcp/access-control/mode", CLICommand: "ikuai-cli network dhcp access-mode get",
			Description: "控制 DHCP 分配黑白名单模式（关闭 / 仅允许白名单 / 拒绝黑名单）。",
			Operations:  []string{"get", "update"},
			Fields: []FieldSpec{
				{Key: "mode", Label: "准入模式", Type: "select", Default: "0", Options: []OptionItem{
					{Label: "关闭黑白名单 (0)", Value: "0"},
					{Label: "白名单模式 (1)", Value: "1"},
					{Label: "黑名单模式 (2)", Value: "2"},
				}},
			},
		},
		{
			Key: "dhcp-access-rules", Group: "network", Label: "DHCP 准入控制规则", Kind: "collection",
			V4Path: "/network/dhcp/access-control/rules", CLICommand: "ikuai-cli network dhcp access-rule list",
			Description:    "配置允许或禁止获取 DHCP 地址的终端 MAC 规则。",
			Operations:     []string{"list", "create", "delete"},
			SupportsPaging: true, SearchKeys: "tagname,mac,comment",
			Columns: []ColumnSpec{
				{Key: "id", Label: "ID", Kind: "mono"},
				{Key: "tagname", Label: "规则名称"},
				{Key: "mac", Label: "MAC 地址", Kind: "mono"},
				{Key: "comment", Label: "备注"},
				{Key: "enabled", Label: "启用", Kind: "badge"},
			},
			Fields: []FieldSpec{
				{Key: "tagname", Label: "规则名称", Type: "text", Required: true, Placeholder: "Allowed-PC"},
				{Key: "mac", Label: "MAC 地址", Type: "text", Required: true, Placeholder: "AA:BB:CC:DD:EE:FF"},
				enabledField(),
				{Key: "comment", Label: "备注", Type: "text", Required: true, Placeholder: "设备说明"},
			},
			Defaults: map[string]any{"enabled": "yes", "comment": ""},
		},
		{
			Key: "vlan", Group: "network", Label: "VLAN 子网划分", Kind: "collection",
			V4Path: "/network/vlan", CLICommand: "ikuai-cli network vlan list",
			Description:    "管理 LAN 侧 IEEE 802.1Q VLAN 虚拟子网。",
			Operations:     []string{"list", "create", "update", "toggle", "delete"},
			SupportsPaging: true, SearchKeys: "vlan_name,interface,ip_addr",
			Columns: []ColumnSpec{
				{Key: "id", Label: "ID", Kind: "mono"},
				{Key: "vlan_name", Label: "VLAN 名称"},
				{Key: "vlan_id", Label: "VLAN ID", Kind: "mono"},
				{Key: "interface", Label: "父接口", Kind: "badge"},
				{Key: "ip_addr", Label: "网关 IP", Kind: "mono"},
				{Key: "netmask", Label: "掩码", Kind: "mono"},
				{Key: "enabled", Label: "启用", Kind: "toggle"},
			},
			Fields: []FieldSpec{
				{Key: "vlan_name", Label: "VLAN 名称", Type: "text", Required: true, Placeholder: "IoT"},
				{Key: "vlan_id", Label: "VLAN ID (1-4094)", Type: "number", Required: true, Placeholder: "100"},
				{Key: "interface", Label: "绑定接口", Type: "text", Required: true, Default: "lan1"},
				{Key: "ip_addr", Label: "IP 地址", Type: "text", Required: true, Placeholder: "10.0.100.1"},
				{Key: "netmask", Label: "子网掩码", Type: "text", Required: true, Default: "255.255.255.0"},
				enabledField(),
				commentField(),
			},
			Defaults:     map[string]any{"enabled": "yes", "mac": "", "ip_mask": "", "comment": ""},
			UpdateFields: []string{"vlan_id", "vlan_name", "tagname", "interface", "mac", "ip_addr", "netmask", "ip_mask", "enabled", "comment"},
		},
		{
			Key: "dnat-rules", Group: "network", Label: "端口映射 (DNAT)", Kind: "collection",
			V4Path: "/network/dnat/rules", CLICommand: "ikuai-cli network dnat list",
			Description:    "外网端口转发到内网服务主机，支持 TCP/UDP 与自定义端口范围。",
			Operations:     []string{"list", "create", "update", "toggle", "delete"},
			SupportsPaging: true, SearchKeys: "tagname,lan_addr,wan_port,comment",
			CompoundFields: []string{"src_addr"},
			Columns: []ColumnSpec{
				{Key: "id", Label: "ID", Kind: "mono"},
				{Key: "tagname", Label: "规则名"},
				{Key: "interface", Label: "外网接口", Kind: "badge"},
				{Key: "protocol", Label: "协议", Kind: "badge"},
				{Key: "wan_port", Label: "外网端口", Kind: "mono"},
				{Key: "lan_addr", Label: "内网 IP", Kind: "mono"},
				{Key: "lan_port", Label: "内网端口", Kind: "mono"},
				{Key: "enabled", Label: "启用", Kind: "toggle"},
			},
			Fields: []FieldSpec{
				{Key: "tagname", Label: "规则名称", Type: "text", Required: true, Placeholder: "NAS-Web"},
				{Key: "interface", Label: "外网线路", Type: "text", Required: true, Default: "all", Placeholder: "all 或 wan1"},
				{Key: "protocol", Label: "协议", Type: "select", Required: true, Default: "tcp", Options: []OptionItem{
					{Label: "TCP", Value: "tcp"}, {Label: "UDP", Value: "udp"}, {Label: "TCP+UDP", Value: "tcp+udp"},
				}},
				{Key: "wan_port", Label: "外网端口", Type: "text", Required: true, Placeholder: "8080 或 8000-8010"},
				{Key: "lan_addr", Label: "内网目标 IP", Type: "text", Required: true, Placeholder: "192.168.9.10"},
				{Key: "lan_port", Label: "内网端口", Type: "text", Required: true, Placeholder: "80"},
				{Key: "src_addr", Label: "允许来源 IP", Type: "csv", Placeholder: "留空不限，逗号分隔"},
				enabledField(),
				commentField(),
			},
			Defaults:     map[string]any{"enabled": "yes", "src_addr": emptyCustomObj(), "comment": ""},
			UpdateFields: []string{"id", "tagname", "enabled", "lan_addr", "lan_port", "protocol", "interface", "wan_port", "src_addr", "comment"},
		},
		{
			Key: "nat-rules", Group: "network", Label: "NAT 转发规则", Kind: "collection",
			V4Path: "/network/nat/rules", CLICommand: "ikuai-cli network nat list",
			Description:    "高级自定义源/目的 NAT 与地址转换策略。",
			Operations:     []string{"list", "create", "update", "toggle", "delete"},
			SupportsPaging: true, SearchKeys: "tagname,comment",
			CompoundFields: []string{"src_addr", "dst_addr", "src_port", "dst_port"},
			Columns: []ColumnSpec{
				{Key: "id", Label: "ID", Kind: "mono"},
				{Key: "tagname", Label: "规则名"},
				{Key: "action", Label: "动作", Kind: "badge"},
				{Key: "iinterface", Label: "入接口"},
				{Key: "ointerface", Label: "出接口"},
				{Key: "protocol", Label: "协议", Kind: "badge"},
				{Key: "src_addr", Label: "源地址", Kind: "compound"},
				{Key: "enabled", Label: "启用", Kind: "toggle"},
			},
			Fields: []FieldSpec{
				{Key: "tagname", Label: "规则名称", Type: "text", Required: true},
				{Key: "action", Label: "NAT 动作", Type: "select", Required: true, Default: "filter", Options: []OptionItem{
					{Label: "SNAT", Value: "snat"}, {Label: "MASQUERADE", Value: "masquerade"}, {Label: "Filter", Value: "filter"},
				}},
				{Key: "iinterface", Label: "入接口", Type: "text", Default: "any"},
				{Key: "ointerface", Label: "出接口", Type: "text", Default: "any"},
				{Key: "protocol", Label: "协议", Type: "select", Default: "any", Options: []OptionItem{
					{Label: "ANY", Value: "any"}, {Label: "TCP", Value: "tcp"}, {Label: "UDP", Value: "udp"},
				}},
				{Key: "src_addr", Label: "源地址", Type: "csv", Placeholder: "192.168.9.0/24"},
				{Key: "dst_addr", Label: "目的地址", Type: "csv"},
				enabledField(),
				commentField(),
			},
			Defaults:     map[string]any{"enabled": "yes", "src_addr": emptyCustomObj(), "dst_addr": emptyCustomObj(), "src_port": emptyCustomObj(), "dst_port": emptyCustomObj(), "comment": ""},
			UpdateFields: []string{"tagname", "enabled", "action", "iinterface", "ointerface", "src_addr", "dst_addr", "nat_addr", "nat_port", "protocol", "comment", "src_port", "dst_port", "src_addr_inv", "dst_addr_inv"},
		},
		{
			Key: "dmz-rules", Group: "network", Label: "DMZ 主机", Kind: "collection",
			V4Path: "/network/dmz/rules", CLICommand: "ikuai-cli network dmz list",
			Description:    "将外网未命中的流量整机映射至指定内网主机。",
			Operations:     []string{"list", "create", "update", "toggle", "delete"},
			SupportsPaging: true, SearchKeys: "tagname,lan_addr,interface",
			Columns: []ColumnSpec{
				{Key: "id", Label: "ID", Kind: "mono"},
				{Key: "tagname", Label: "名称"},
				{Key: "interface", Label: "外网接口/IP", Kind: "mono"},
				{Key: "lan_addr", Label: "DMZ 主机", Kind: "mono"},
				{Key: "protocol", Label: "协议", Kind: "badge"},
				{Key: "excl_port", Label: "排除端口", Kind: "mono"},
				{Key: "enabled", Label: "启用", Kind: "toggle"},
			},
			Fields: []FieldSpec{
				{Key: "tagname", Label: "规则名", Type: "text", Required: true},
				{Key: "interface", Label: "外网接口/IP", Type: "text", Required: true, Placeholder: "wan1"},
				{Key: "lan_addr", Label: "内网主机 IP", Type: "text", Required: true, Placeholder: "192.168.9.250"},
				{Key: "protocol", Label: "协议", Type: "select", Default: "any", Options: []OptionItem{
					{Label: "ANY", Value: "any"}, {Label: "TCP", Value: "tcp"}, {Label: "UDP", Value: "udp"},
				}},
				{Key: "excl_port", Label: "排除端口", Type: "text", Default: "80,443"},
				enabledField(),
				commentField(),
			},
			Defaults:     map[string]any{"enabled": "yes", "comment": ""},
			UpdateFields: []string{"id", "tagname", "enabled", "interface", "lan_addr", "protocol", "excl_port", "comment"},
		},
		{
			Key: "pppoe-config", Group: "network", Label: "PPPoE 服务端", Kind: "singleton",
			V4Path: "/network/pppoe/services", CLICommand: "ikuai-cli network pppoe get",
			Description: "配置本地 PPPoE 认证服务参数、MTU/MRU 与 RADIUS 对接。",
			Operations:  []string{"get", "update"},
			Columns: []ColumnSpec{
				{Key: "enabled", Label: "服务状态", Kind: "badge"},
				{Key: "server_name", Label: "服务名"},
				{Key: "server_ip", Label: "服务端 IP", Kind: "mono"},
				{Key: "addr_pool", Label: "地址池", Kind: "mono"},
				{Key: "mtu", Label: "MTU", Kind: "mono"},
				{Key: "mru", Label: "MRU", Kind: "mono"},
			},
			Fields: []FieldSpec{
				enabledField(),
				{Key: "server_name", Label: "服务名称", Type: "text", Default: "iKuai"},
				{Key: "interface", Label: "监听接口", Type: "text", Default: "lan1"},
				{Key: "server_ip", Label: "服务 IP", Type: "text"},
				{Key: "addr_pool", Label: "地址池", Type: "text"},
				{Key: "dns1", Label: "主 DNS", Type: "text"},
				{Key: "dns2", Label: "备 DNS", Type: "text"},
				{Key: "mtu", Label: "MTU", Type: "number", Default: 1480},
				{Key: "mru", Label: "MRU", Type: "number", Default: 1480},
				commentField(),
			},
			UpdateFields: []string{
				"enabled", "server_name", "force_verify_name", "server_ip", "dns1", "dns2",
				"authmode", "nas_identifier", "nas_ip_address", "radius_ip", "secret",
				"authport", "accountport", "addr_pool", "interface", "rate_limit_lan",
				"drop_client", "force_pppoe", "enhance_check", "share_deny",
				"bind_vlan", "verify_vlan", "bind_iface", "mtu", "mru",
				"lcp_echo_interval", "lcp_echo_failure", "maxconnect",
				"restart_timer", "restart_week", "restart_time", "comment",
			},
		},
	}
}

package service

import (
	"strings"
	"testing"
)

func TestBuildNetworkMapFromDataUsesInterfacesAndClients(t *testing.T) {
	data := buildNetworkMapFromData("Office Router", &InterfaceDataResult{
		WanStatus: []WanStatus{
			{Name: "wan1", IP: "203.0.113.10", Status: "success", Comment: "primary"},
		},
		TrafficDetails: []TrafficDetail{
			{Name: "lan1", IP: "192.168.10.1", Comment: "office lan"},
			{Name: "wan1", IP: "203.0.113.10", Comment: "primary"},
		},
	}, []ClientDTO{
		{MAC: "AA:BB:CC:DD:EE:FF", Hostname: "nas", IPs: []string{"192.168.10.20"}},
	})

	if len(data.Nodes) != 4 {
		t.Fatalf("node count = %d, want 4: %#v", len(data.Nodes), data.Nodes)
	}
	if !hasMapNode(data.Nodes, "router", "Office Router", "router") {
		t.Fatalf("router node missing: %#v", data.Nodes)
	}
	if !hasMapNode(data.Nodes, "wan:wan1", "wan1", "wan") {
		t.Fatalf("wan node missing: %#v", data.Nodes)
	}
	if !hasMapNode(data.Nodes, "lan:lan1", "lan1", "lan") {
		t.Fatalf("lan node missing: %#v", data.Nodes)
	}
	if !hasMapNode(data.Nodes, "device:AA:BB:CC:DD:EE:FF", "nas", "device") {
		t.Fatalf("client node missing: %#v", data.Nodes)
	}
	if !hasMapLink(data.Links, "wan:wan1", "router") || !hasMapLink(data.Links, "router", "lan:lan1") || !hasMapLink(data.Links, "lan:lan1", "device:AA:BB:CC:DD:EE:FF") {
		t.Fatalf("expected links missing: %#v", data.Links)
	}
}

func TestSecurityMappersFilterDNATAndAbnormalClients(t *testing.T) {
	ports := portMappingsFromRows([]map[string]any{
		{"id": 1, "enabled": "yes", "comment": "SSH", "protocol": "tcp", "wan_port": "2222", "lan_addr": "192.168.10.20", "lan_port": "22"},
		{"id": 2, "enabled": "no", "comment": "disabled RDP", "protocol": "tcp", "wan_port": "3389", "lan_addr": "192.168.10.21", "lan_port": "3389"},
		{"id": 3, "enabled": "yes", "comment": "game", "protocol": "udp", "wan_port": "25565", "lan_addr": "192.168.10.22", "lan_port": "25565"},
	})

	if len(ports) != 1 {
		t.Fatalf("port count = %d, want 1: %#v", len(ports), ports)
	}
	if ports[0].Name != "SSH" || ports[0].ExtPort != "2222" || ports[0].IntPort != "22" {
		t.Fatalf("unexpected port mapping: %#v", ports[0])
	}

	clients := abnormalClientsFromClients([]ClientDTO{
		{MAC: "normal", Connections: 12, UploadSpeed: 1024},
		{MAC: "busy", Connections: 180, UploadSpeed: 2048},
		{MAC: "upload", Connections: 5, UploadSpeed: 15 * 1024 * 1024},
	})
	if len(clients) != 2 {
		t.Fatalf("abnormal count = %d, want 2: %#v", len(clients), clients)
	}
}

func TestCommonResourceCatalogKeepsWritesToSafeIntersection(t *testing.T) {
	catalog := commonResourceCatalog()

	for _, name := range []string{"dhcp-static", "dns-static", "dnat-rules", "static-routes"} {
		def, ok := catalog.Lookup(name)
		if !ok {
			t.Fatalf("resource %q missing", name)
		}
		if !def.Writable {
			t.Fatalf("resource %q Writable = false, want true", name)
		}
	}

	for _, name := range []string{"vlan", "qos", "policy-routes", "domain-rules"} {
		def, ok := catalog.Lookup(name)
		if !ok {
			t.Fatalf("resource %q missing", name)
		}
		if def.Writable {
			t.Fatalf("resource %q Writable = true, want false", name)
		}
	}

	if _, ok := catalog.Lookup("wireguard"); ok {
		t.Fatalf("v4-only resource wireguard should not be in common catalog")
	}
}

func TestUnsupportedResourceErrorReturnsUnavailableResourceData(t *testing.T) {
	def, ok := commonResourceCatalog().Lookup("qos")
	if !ok {
		t.Fatal("qos resource missing")
	}

	data, handled := unavailableResourceData(def, "[SDK Error 3] Not found funcname(qos)")
	if !handled {
		t.Fatal("unsupported funcname error was not handled")
	}
	if len(data.Rows) != 0 {
		t.Fatalf("rows = %d, want 0", len(data.Rows))
	}
	if data.Resource.Available {
		t.Fatal("resource available = true, want false")
	}
	if data.Resource.Writable {
		t.Fatal("resource writable = true, want false")
	}
	if !strings.Contains(data.Error, "当前固件不支持") {
		t.Fatalf("error = %q, want unsupported firmware message", data.Error)
	}
}

func hasMapNode(nodes []MapNode, id, name, nodeType string) bool {
	for _, node := range nodes {
		if node.ID == id && node.Name == name && node.Type == nodeType {
			return true
		}
	}
	return false
}

func hasMapLink(links []MapLink, source, target string) bool {
	for _, link := range links {
		if link.Source == source && link.Target == target {
			return true
		}
	}
	return false
}

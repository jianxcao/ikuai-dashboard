package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	ikuaiapi "github.com/zy84338719/ikuai-api"
)

type CommonResourceDefinition struct {
	Name      string   `json:"name"`
	Label     string   `json:"label"`
	Group     string   `json:"group"`
	V3Name    string   `json:"v3_name"`
	V4Path    string   `json:"v4_path"`
	Writable  bool     `json:"writable"`
	Available bool     `json:"available"`
	Methods   []string `json:"methods"`
}

type CommonResourceCatalog []CommonResourceDefinition

func (c CommonResourceCatalog) Lookup(name string) (CommonResourceDefinition, bool) {
	for _, def := range c {
		if def.Name == name {
			return def, true
		}
	}
	return CommonResourceDefinition{}, false
}

type CommonResourceData struct {
	Resource CommonResourceDefinition `json:"resource"`
	Rows     []map[string]any         `json:"rows"`
	Error    string                   `json:"error,omitempty"`
}

type CommonResourceMutation struct {
	Resource string         `json:"resource"`
	Action   string         `json:"action"`
	Result   map[string]any `json:"result"`
}

func commonResourceCatalog() CommonResourceCatalog {
	return CommonResourceCatalog{
		{Name: "dhcp-services", Label: "DHCP 服务", Group: "network", V3Name: "dhcp-services", V4Path: "/network/dhcp/services", Writable: false, Available: true, Methods: []string{http.MethodGet}},
		{Name: "dhcp-static", Label: "DHCP 静态绑定", Group: "network", V3Name: "dhcp-static", V4Path: "/network/dhcp/static", Writable: true, Available: true, Methods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}},
		{Name: "dhcp-clients", Label: "DHCP 租约", Group: "network", V3Name: "dhcp-clients", V4Path: "/network/dhcp/clients", Writable: false, Available: true, Methods: []string{http.MethodGet}},
		{Name: "dns-forward", Label: "DNS 转发", Group: "network", V3Name: "dns-forward", V4Path: "/network/dns/config", Writable: false, Available: true, Methods: []string{http.MethodGet}},
		{Name: "dns-static", Label: "DNS 静态解析", Group: "network", V3Name: "dns-static", V4Path: "/network/dns/proxy/rules", Writable: true, Available: true, Methods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}},
		{Name: "vlan", Label: "VLAN", Group: "network", V3Name: "vlan", V4Path: "/network/vlan", Writable: false, Available: true, Methods: []string{http.MethodGet}},
		{Name: "dnat-rules", Label: "DNAT / 端口映射", Group: "security", V3Name: "dnat-rules", V4Path: "/network/dnat/rules", Writable: true, Available: true, Methods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}},
		{Name: "qos", Label: "QoS", Group: "network", V3Name: "qos", V4Path: "/network/qos/ip", Writable: false, Available: true, Methods: []string{http.MethodGet}},
		{Name: "bandwidth", Label: "带宽控制", Group: "network", V3Name: "bandwidth", V4Path: "/network/qos/mac", Writable: false, Available: true, Methods: []string{http.MethodGet}},
		{Name: "static-routes", Label: "静态路由", Group: "routing", V3Name: "static-routes", V4Path: "/routing/static-routes", Writable: true, Available: true, Methods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}},
		{Name: "policy-routes", Label: "策略路由", Group: "routing", V3Name: "policy-routes", V4Path: "/routing/five-tuple-rules", Writable: false, Available: true, Methods: []string{http.MethodGet}},
		{Name: "domain-rules", Label: "域名分流", Group: "routing", V3Name: "domain-rules", V4Path: "/routing/domain-rules", Writable: false, Available: true, Methods: []string{http.MethodGet}},
	}
}

func (s *MonitorService) ListCommonResources() []CommonResourceDefinition {
	defs := append([]CommonResourceDefinition(nil), commonResourceCatalog()...)
	sort.Slice(defs, func(i, j int) bool {
		if defs[i].Group != defs[j].Group {
			return defs[i].Group < defs[j].Group
		}
		return defs[i].Name < defs[j].Name
	})
	return defs
}

func (s *MonitorService) GetCommonResource(ctx context.Context, name string) (*CommonResourceData, error) {
	if s == nil {
		return nil, errors.New("服务未初始化")
	}
	if s.unconfigured {
		return nil, ErrUnconfigured
	}
	def, ok := commonResourceCatalog().Lookup(name)
	if !ok {
		return nil, fmt.Errorf("不支持的资源: %s", name)
	}
	if s.mockMode {
		return &CommonResourceData{Resource: def, Rows: mockCommonResourceRows(def.Name)}, nil
	}
	rows, err := s.getCommonResourceRows(ctx, def)
	if err != nil {
		if data, ok := unavailableResourceData(def, err.Error()); ok {
			return data, nil
		}
		return nil, err
	}
	return &CommonResourceData{Resource: def, Rows: rows}, nil
}

func (s *MonitorService) MutateCommonResource(ctx context.Context, name, method string, id int, body map[string]any) (*CommonResourceMutation, error) {
	if s == nil {
		return nil, errors.New("服务未初始化")
	}
	if s.unconfigured {
		return nil, ErrUnconfigured
	}
	def, ok := commonResourceCatalog().Lookup(name)
	if !ok {
		return nil, fmt.Errorf("不支持的资源: %s", name)
	}
	if !def.Writable {
		return nil, fmt.Errorf("资源 %s 暂未开放写操作", name)
	}
	if method == http.MethodPut || method == http.MethodDelete {
		if id <= 0 {
			return nil, errors.New("编辑或删除资源需要有效 ID")
		}
	}
	if s.mockMode {
		return &CommonResourceMutation{
			Resource: name,
			Action:   method,
			Result:   mapResult("mock mutation accepted", id, body),
		}, nil
	}

	var (
		result map[string]any
		err    error
	)
	if s.v4Client != nil {
		result, err = s.mutateV4CommonResource(ctx, def, method, id, body)
	} else {
		result, err = s.mutateV3CommonResource(ctx, def, method, id, body)
	}
	if err != nil {
		return nil, err
	}
	return &CommonResourceMutation{Resource: name, Action: method, Result: result}, nil
}

func (s *MonitorService) getCommonResourceRows(ctx context.Context, def CommonResourceDefinition) ([]map[string]any, error) {
	if s.v4Client != nil {
		payload, err := s.v4Get(ctx, def.V4Path)
		if err != nil {
			return nil, err
		}
		return rowsFromAny(payload), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var payload any
	action := s.client.V3Action()
	err := retryV3CallWithRelogin(ctx, func() error {
		return action.Show(ctx, def.V3Name, nil, &payload)
	}, s.reloginV3)
	if err != nil {
		return nil, err
	}
	return rowsFromAny(payload), nil
}

func (s *MonitorService) mutateV3CommonResource(ctx context.Context, def CommonResourceDefinition, method string, id int, body map[string]any) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	action := s.client.V3Action()
	if body == nil {
		body = map[string]any{}
	}
	if id > 0 {
		body["id"] = id
	}
	var payload any
	call := func() error {
		switch method {
		case http.MethodPost:
			return action.Add(ctx, def.V3Name, body, &payload)
		case http.MethodPut:
			return action.Edit(ctx, def.V3Name, body, &payload)
		case http.MethodDelete:
			return action.Delete(ctx, def.V3Name, id, &payload)
		default:
			return fmt.Errorf("不支持的写操作: %s", method)
		}
	}
	if err := retryV3CallWithRelogin(ctx, call, s.reloginV3); err != nil {
		return nil, err
	}
	return mapResult("ok", id, payload), nil
}

func (s *MonitorService) mutateV4CommonResource(ctx context.Context, def CommonResourceDefinition, method string, id int, body map[string]any) (map[string]any, error) {
	if body == nil {
		body = map[string]any{}
	}
	if id > 0 {
		body["id"] = id
	}
	var payload any
	switch method {
	case http.MethodPost:
		if err := s.v4Client.Post(ctx, def.V4Path, body, &payload); err != nil {
			return nil, err
		}
	case http.MethodPut:
		if err := s.v4Client.Put(ctx, def.V4Path, body, &payload); err != nil {
			return nil, err
		}
	case http.MethodDelete:
		if err := s.v4Client.Delete(ctx, fmt.Sprintf("%s/%d", def.V4Path, id), &payload); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("不支持的写操作: %s", method)
	}
	return mapResult("ok", id, payload), nil
}

func decodeResourceBody(raw []byte) (map[string]any, error) {
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, fmt.Errorf("JSON 请求体无效: %w", err)
	}
	return body, nil
}

func unavailableResourceData(def CommonResourceDefinition, message string) (*CommonResourceData, bool) {
	lowerMessage := strings.ToLower(message)
	if !strings.Contains(lowerMessage, "not found funcname") &&
		!strings.Contains(lowerMessage, "unknown v3 compatibility endpoint") &&
		!strings.Contains(lowerMessage, "does not support action") {
		return nil, false
	}
	def.Available = false
	def.Writable = false
	return &CommonResourceData{
		Resource: def,
		Rows:     []map[string]any{},
		Error:    "当前固件不支持该资源接口：" + message,
	}, true
}

func mockCommonResourceRows(name string) []map[string]any {
	switch name {
	case "dhcp-static":
		return []map[string]any{{"id": 1, "enabled": "yes", "tagname": "nas", "mac": "8E:F6:C5:99:8A:50", "ip_addr": "192.168.50.100", "comment": "NAS 静态绑定"}}
	case "dns-static":
		return []map[string]any{{"id": 1, "enabled": "yes", "domain": "nas.local", "ip_addr": "192.168.50.100", "comment": "本地 NAS"}}
	case "dnat-rules":
		return []map[string]any{{"id": 1, "enabled": "yes", "comment": "NAS SSH", "protocol": "tcp", "wan_port": "2222", "lan_addr": "192.168.50.100", "lan_port": "22"}}
	case "static-routes":
		return []map[string]any{{"id": 1, "enabled": "yes", "dst_addr": "10.20.0.0/16", "gateway": "192.168.50.2", "interface": "lan1", "metric": 10, "comment": "实验室网段"}}
	case "dhcp-services":
		return []map[string]any{{"id": 1, "enabled": "yes", "interface": "lan1", "ip_addr": "192.168.50.1", "start": "192.168.50.100", "end": "192.168.50.240"}}
	case "dhcp-clients":
		return []map[string]any{{"id": 1, "hostname": "macbook", "mac": "00:70:FA:3D:30:10", "ip_addr": "192.168.50.103", "expire": "12h"}}
	case "vlan":
		return []map[string]any{{"id": 1, "interface": "lan1", "vlan_id": 20, "comment": "IoT"}}
	case "qos":
		return []map[string]any{{"id": 1, "enabled": "yes", "src_addr": "192.168.50.0/24", "priority": 3, "comment": "办公网优先级"}}
	case "bandwidth":
		return []map[string]any{{"id": 1, "enabled": "yes", "src_addr": "192.168.50.100", "max_upload": 10485760, "max_download": 10485760, "comment": "NAS 限速"}}
	case "policy-routes":
		return []map[string]any{{"id": 1, "enabled": "yes", "src_addr": "192.168.50.0/24", "dst_addr": "0.0.0.0/0", "interface": "wan1", "comment": "办公出口"}}
	case "domain-rules":
		return []map[string]any{{"id": 1, "enabled": "yes", "domain": "example.com", "interface": "wan2", "comment": "测试分流"}}
	default:
		return nil
	}
}

var _ = ikuaiapi.V3Endpoint{}

package v4service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"ikuai-dashboard/backend/internal/config"
	"ikuai-dashboard/backend/internal/ikuaiv4"
)

var (
	ErrWanNotFound = errors.New("未找到指定的 WAN 口")
	ErrWanNoIP     = errors.New("WAN 口当前未获取到有效 IP")
)

// WanIPDTO represents the resolved IP of a router's WAN interface.
type WanIPDTO struct {
	RouterID   string `json:"router_id"`
	RouterName string `json:"router_name"`
	Interface  string `json:"interface"`
	IP         string `json:"ip"`
}

type wanCandidate struct {
	iface  string
	parent string
	ip     string
}

// GetWanIP resolves the IP address of the requested WAN interface (or the first active WAN if wanName is empty).
func GetWanIP(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client, wanName string) (*WanIPDTO, error) {
	wanName = strings.TrimSpace(wanName)

	if router.Mock || client == nil {
		ov := mockOverview(router)
		candidates := make([]wanCandidate, 0, len(ov.WanChecks))
		for _, w := range ov.WanChecks {
			candidates = appendCandidate(candidates, w.Interface, w.ParentInterface, w.IPAddr)
		}
		for _, t := range ov.InterfaceTraffic {
			if !isNonWanInterface(t.Interface) {
				candidates = appendCandidate(candidates, t.Interface, "", t.IPAddr)
			}
		}
		return selectWanCandidate(router, candidates, wanName)
	}

	statusRaw, err := client.Get(ctx, "/monitoring/interfaces-status", nil)
	if err != nil {
		return nil, err
	}

	var candidates []wanCandidate
	if statusAny, decErr := ikuaiv4.DecodeAny(statusRaw); decErr == nil {
		if statusMap, ok := statusAny.(map[string]any); ok {
			if checks, ok := statusMap["iface_check"].([]any); ok {
				for _, item := range checks {
					if row, ok := item.(map[string]any); ok {
						iface := ikuaiv4.Str(row, "interface", "name", "tagname")
						parent := ikuaiv4.Str(row, "parent_interface")
						ip := ikuaiv4.Str(row, "ip_addr", "ip", "ip_mask")
						candidates = appendCandidate(candidates, iface, parent, ip)
					}
				}
			}
			if streams, ok := statusMap["iface_stream"].([]any); ok {
				for _, item := range streams {
					if row, ok := item.(map[string]any); ok {
						iface := ikuaiv4.Str(row, "interface", "name", "tagname")
						if isNonWanInterface(iface) {
							continue
						}
						ip := ikuaiv4.Str(row, "ip_addr", "ip", "ip_mask")
						candidates = appendCandidate(candidates, iface, "", ip)
					}
				}
			}
		}
	}

	if res, selErr := selectWanCandidate(router, candidates, wanName); selErr == nil {
		return res, nil
	}

	// Fallback to /interfaces/wan-config (and /interfaces/wan-vlan-config when targeting a specific WAN).
	if wanRaw, wanErr := client.Get(ctx, "/interfaces/wan-config", nil); wanErr == nil {
		rows, _ := ikuaiv4.ExtractCollection(wanRaw)
		for _, row := range rows {
			iface := ikuaiv4.Str(row, "tagname", "interface", "name")
			if isNonWanInterface(iface) {
				continue
			}
			ip := ikuaiv4.Str(row, "ip_addr", "pppoe_ip_addr", "dhcp_ip_addr", "ip_mask", "ip")
			candidates = appendCandidate(candidates, iface, "", ip)
		}
	}

	if wanName != "" {
		if res, selErr := selectWanCandidate(router, candidates, wanName); selErr == nil {
			return res, nil
		}
		if vlanRaw, vlanErr := client.Get(ctx, "/interfaces/wan-vlan-config", nil); vlanErr == nil {
			rows, _ := ikuaiv4.ExtractCollection(vlanRaw)
			for _, row := range rows {
				iface := ikuaiv4.Str(row, "vlan_name", "interface", "name")
				parent := ikuaiv4.Str(row, "interface", "parent_interface")
				ip := ikuaiv4.Str(row, "ip_addr", "pppoe_ip_addr", "dhcp_ip_addr", "ip_mask", "ip")
				candidates = appendCandidate(candidates, iface, parent, ip)
			}
		}
	}

	return selectWanCandidate(router, candidates, wanName)
}

func appendCandidate(candidates []wanCandidate, iface, parent, rawIP string) []wanCandidate {
	iface = strings.TrimSpace(iface)
	if iface == "" {
		return candidates
	}
	parent = strings.TrimSpace(parent)
	ip := cleanWanIP(rawIP)
	for i := range candidates {
		if strings.EqualFold(candidates[i].iface, iface) {
			if candidates[i].ip == "" && ip != "" {
				candidates[i].ip = ip
			}
			if candidates[i].parent == "" && parent != "" {
				candidates[i].parent = parent
			}
			return candidates
		}
	}
	return append(candidates, wanCandidate{
		iface:  iface,
		parent: parent,
		ip:     ip,
	})
}

func selectWanCandidate(router config.RouterConfig, candidates []wanCandidate, wanName string) (*WanIPDTO, error) {
	if wanName != "" {
		matchedName := false
		for _, c := range candidates {
			if strings.EqualFold(c.iface, wanName) {
				matchedName = true
				if c.ip != "" {
					return &WanIPDTO{
						RouterID:   router.ID,
						RouterName: router.Name,
						Interface:  c.iface,
						IP:         c.ip,
					}, nil
				}
			}
		}
		for _, c := range candidates {
			if c.parent != "" && strings.EqualFold(c.parent, wanName) {
				matchedName = true
				if c.ip != "" {
					return &WanIPDTO{
						RouterID:   router.ID,
						RouterName: router.Name,
						Interface:  c.iface,
						IP:         c.ip,
					}, nil
				}
			}
		}
		if matchedName {
			return nil, fmt.Errorf("%w: %s", ErrWanNoIP, wanName)
		}
		return nil, fmt.Errorf("%w: %s", ErrWanNotFound, wanName)
	}

	for _, c := range candidates {
		if c.ip != "" {
			return &WanIPDTO{
				RouterID:   router.ID,
				RouterName: router.Name,
				Interface:  c.iface,
				IP:         c.ip,
			}, nil
		}
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("%w: 未检测到任何 WAN 口", ErrWanNotFound)
	}
	return nil, ErrWanNoIP
}

func cleanWanIP(raw string) string {
	for _, part := range strings.Split(raw, ",") {
		s := strings.TrimSpace(part)
		if s == "" {
			continue
		}
		if slash := strings.IndexByte(s, '/'); slash >= 0 {
			s = strings.TrimSpace(s[:slash])
		}
		if s == "" || s == "--" || s == "0.0.0.0" || s == "::" || strings.EqualFold(s, "nil") || strings.EqualFold(s, "null") {
			continue
		}
		return s
	}
	return ""
}

func isNonWanInterface(name string) bool {
	s := strings.ToLower(strings.TrimSpace(name))
	if s == "" || s == "lo" || s == "null" {
		return true
	}
	return strings.HasPrefix(s, "lan") || strings.HasPrefix(s, "br") || strings.HasPrefix(s, "docker")
}

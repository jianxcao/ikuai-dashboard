package v4catalog

import (
	"context"
	"crypto/ecdh"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"ikuai-dashboard/backend/internal/config"
	"ikuai-dashboard/backend/internal/ikuaiv4"
)

var integerAPIFields = map[string]bool{
	"id": true, "vlan_id": true, "vlanid": true, "lease": true, "delay": true,
	"check_addr_valid": true, "check_relay_only": true, "cache_ttl": true,
	"proxy_force": true, "is_ipv6": true, "prio": true, "priority": true,
	"mode": true, "method": true, "type": true, "iface_band": true,
	"src_addr_inv": true, "dst_addr_inv": true, "upload": true, "download": true,
	"limits": true, "noping_lan": true, "noping_wan": true, "notracert": true,
	"hijack_ping": true, "invalid": true, "acl_mac": true,
	"nol2rt": true, "ttl_num": true, "hit_rate": true,
	"dos_lan": true, "dos_lan_num": true, "tcp_mss": true, "tcp_mss_num": true,
	"packages": true, "share": true, "price": true, "up_speed": true, "down_speed": true,
	"idle_time": true, "max_time": true, "user_auth": true, "nopasswd": true, "https_redirect": true,
	"mtu": true, "mru": true, "lcp_echo_interval": true,
	"lcp_echo_failure": true, "maxconnect": true, "server_port": true,
	"local_listenport": true, "endpoint_port": true, "keepalive": true,
	"ikelifetime": true, "lifetime": true, "force_ipsec": true, "open_mppe": true,
	"open_ftp": true, "ftp_port": true, "ftp_access": true,
	"http_port": true, "ssl_on": true, "autoindex": true, "access": true,
	"wsdd2": true, "listen_port": true, "version": true,
	"time_zone": true, "switch_nat": true, "switch_ntp": true, "switch_ntpd": true,
	"switch_ntpserver": true, "ntp_sync_cycle": true, "lan_nat": true, "fast_nat": true,
	"open_telnetd": true, "open_sshd": true, "sshd_port": true, "open_wanweb": true,
	"https_port": true, "force_https": true, "turbo": true, "valid_days": true,
	"group_id": true, "sesstimeout": true, "interval": true, "force": true,
	"bbr": true, "established_timeout": true, "time_wait_timeout": true, "udp_timeout": true,
	"support_ftp": true, "support_sip": true, "support_tftp": true, "support_h323": true,
}

var listOnlyResources = map[string]bool{
	"dns-proxy":   true,
	"ftp-users":   true,
	"http-users":  true,
	"samba-users": true,
}

var nativeKeyPatternResources = map[string]bool{
	"vlan":            true,
	"wireguard":       true,
	"ipsec-clients":   true,
	"openvpn-clients": true,
	"pptp-clients":    true,
	"l2tp-clients":    true,
	"ikev2-clients":   true,
	"ftp-users":       true,
	"http-users":      true,
	"samba-users":     true,
}

func supportsNativeSearch(spec ResourceSpec) (mode string) {
	if spec.Key == "auth-online" {
		return "keywords"
	}
	if spec.Key == "log-url-visits" {
		return "pattern_only"
	}
	if nativeKeyPatternResources[spec.Key] || (strings.HasPrefix(spec.Key, "log-") && spec.Key != "log-url-visits") {
		return "key_pattern"
	}
	return ""
}

// ListOrGet fetches collection rows or singleton configuration for a catalog resource.
func ListOrGet(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client, key string, page, pageSize int, search, filter string, extraQuery ...map[string]string) (*ListResult, error) {
	spec, ok := Lookup(key)
	if !ok {
		return nil, fmt.Errorf("未注册的资源: %s", key)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	if router.Mock || client == nil {
		return mockListOrGet(spec, page, pageSize, search, extraQuery...), nil
	}

	params := map[string]string{}
	for _, q := range extraQuery {
		for k, v := range q {
			if strings.TrimSpace(v) != "" {
				params[k] = v
			}
		}
	}
	if spec.Key == "system-files" && params["path"] == "" {
		params["path"] = "/"
	}

	trimmedSearch := strings.TrimSpace(search)
	nativeSearchMode := supportsNativeSearch(spec)
	useLocalSearch := trimmedSearch != "" && (!spec.SupportsPaging || nativeSearchMode == "")

	if spec.Kind == "collection" && spec.SupportsPaging {
		if useLocalSearch {
			// Pull broader page from router so local substring matching works across pages.
			params["page"] = "1"
			params["limit"] = "500"
		} else {
			params["page"] = strconv.Itoa(page)
			params["limit"] = strconv.Itoa(pageSize)
		}
		if filter != "" {
			params["filter"] = filter
		}
		if trimmedSearch != "" && !useLocalSearch {
			switch nativeSearchMode {
			case "keywords":
				params["KEYWORDS"] = trimmedSearch
				if spec.SearchKeys != "" {
					params["FINDS"] = spec.SearchKeys
				}
			case "pattern_only":
				params["pattern"] = trimmedSearch
			case "key_pattern":
				if spec.SearchKeys != "" {
					params["key"] = spec.SearchKeys
				}
				params["pattern"] = trimmedSearch
			}
		}
	}

	raw, err := client.Get(ctx, spec.V4Path, params)
	if err != nil {
		return nil, err
	}

	if spec.Kind == "singleton" {
		obj := ikuaiv4.ExtractObject(raw)
		if spec.Key == "cpufreq-mode" {
			normalizeCPUFreqObject(obj)
		}
		decodeIKuaiPEMFields(obj)
		return &ListResult{
			Resource: spec,
			Rows:     []map[string]any{obj},
			Total:    1,
			Config:   obj,
		}, nil
	}

	rows, total := ikuaiv4.ExtractCollection(raw)
	for _, r := range rows {
		if spec.Key == "dhcp-leases" && r["timeout"] == nil && r["expire"] != nil {
			r["timeout"] = r["expire"]
		}
		if spec.Key == "wireless-aps" && strOr(r["tagname"], "") == "" && strOr(r["apname"], "") != "" {
			r["tagname"] = r["apname"]
		}
		if spec.ObjectValueKey != "" {
			if gv, ok := r["group_value"].([]any); ok {
				vals := make([]string, 0, len(gv))
				for _, item := range gv {
					if m, ok := item.(map[string]any); ok {
						if s := ikuaiv4.Str(m, spec.ObjectValueKey, "value"); s != "" {
							vals = append(vals, s)
						}
					}
				}
				r["value"] = strings.Join(vals, ",")
			}
		}
		if spec.Key == "time-objects" {
			flattenTimeGroupValue(r)
		}
		decodeIKuaiPEMFields(r)
	}

	if useLocalSearch {
		rows = filterRowsBySearch(rows, trimmedSearch)
		total = len(rows)
		if spec.SupportsPaging {
			rows = paginateRows(rows, page, pageSize)
		}
	}

	return &ListResult{
		Resource: spec,
		Rows:     rows,
		Total:    total,
	}, nil
}

// Create creates a new item in a collection resource (or previews when dryRun is true).
func Create(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client, key string, body map[string]any, dryRun bool) (map[string]any, error) {
	spec, ok := Lookup(key)
	if !ok {
		return nil, fmt.Errorf("未注册的资源: %s", key)
	}
	if !containsOp(spec.Operations, "create") {
		return nil, fmt.Errorf("资源 %s 不支持新建操作", spec.Label)
	}

	payload := preparePayload(spec, body, true)
	if dryRun {
		if client == nil {
			return map[string]any{"dry_run": true, "method": http.MethodPost, "url": "mock://api/v4.0" + spec.V4Path, "body": payload}, nil
		}
		return client.PreviewDryRun(http.MethodPost, spec.V4Path, nil, payload), nil
	}

	if router.Mock || client == nil {
		return mockCreate(spec, payload), nil
	}

	raw, err := client.Post(ctx, spec.V4Path, payload)
	if err != nil {
		return nil, err
	}
	return ikuaiv4.ExtractObject(raw), nil
}

// Update updates a collection item (by id) or singleton config using GET-then-merge-then-PUT.
func Update(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client, key, id string, body map[string]any, dryRun bool) (map[string]any, error) {
	spec, ok := Lookup(key)
	if !ok {
		return nil, fmt.Errorf("未注册的资源: %s", key)
	}
	if !containsOp(spec.Operations, "update") {
		return nil, fmt.Errorf("资源 %s 不支持更新操作", spec.Label)
	}

	targetPath := spec.V4Path
	if spec.Kind == "collection" {
		if strings.TrimSpace(id) == "" {
			return nil, fmt.Errorf("更新列表记录需要有效的 ID")
		}
		targetPath = spec.V4Path + "/" + strings.TrimSpace(id)
	}

	updates := preparePayload(spec, body, false)

	if router.Mock || client == nil {
		if dryRun {
			merged := mockUpdatePreview(spec, id, body, updates)
			return map[string]any{"dry_run": true, "method": http.MethodPut, "url": "mock://api/v4.0" + targetPath, "body": merged}, nil
		}
		return mockUpdate(spec, id, body, updates), nil
	}

	// Read current state first so firmware-required fields are never dropped.
	existing, fetchErr := fetchUpdateBaseline(ctx, client, spec, targetPath, id)
	if fetchErr != nil && len(existing) == 0 {
		return nil, fmt.Errorf("读取当前配置失败，无法安全执行更新: %w", fetchErr)
	}
	if spec.Kind == "collection" && len(existing) == 0 {
		return nil, fmt.Errorf("未找到 ID=%s 的现有配置，无法执行更新", id)
	}
	if spec.Key == "cpufreq-mode" {
		normalizeCPUFreqObject(existing)
	}

	merged := buildMergedUpdateBody(spec, id, body, updates, existing)

	if dryRun {
		return client.PreviewDryRun(http.MethodPut, targetPath, nil, merged), nil
	}

	raw, err := client.Put(ctx, targetPath, merged)
	if err != nil {
		return nil, err
	}
	return ikuaiv4.ExtractObject(raw), nil
}

// Toggle toggles enabled ("yes"|"no") via PATCH /path/<id>.
func Toggle(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client, key, id, enabled string, dryRun bool) (map[string]any, error) {
	spec, ok := Lookup(key)
	if !ok {
		return nil, fmt.Errorf("未注册的资源: %s", key)
	}
	if !containsOp(spec.Operations, "toggle") {
		return nil, fmt.Errorf("资源 %s 不支持启停切换", spec.Label)
	}
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("切换状态需要有效的 ID")
	}

	val := "yes"
	if strings.EqualFold(enabled, "no") || enabled == "0" || strings.EqualFold(enabled, "false") {
		val = "no"
	}
	body := map[string]any{"enabled": val}
	targetPath := spec.V4Path + "/" + strings.TrimSpace(id)

	if dryRun {
		if client == nil {
			return map[string]any{"dry_run": true, "method": http.MethodPatch, "url": "mock://api/v4.0" + targetPath, "body": body}, nil
		}
		return client.PreviewDryRun(http.MethodPatch, targetPath, nil, body), nil
	}
	if router.Mock || client == nil {
		return mockToggle(spec, id, val), nil
	}

	raw, err := client.Patch(ctx, targetPath, body)
	if err != nil {
		return nil, err
	}
	return ikuaiv4.ExtractObject(raw), nil
}

// Delete removes a resource item, clears logs, or deletes a backup snapshot.
func Delete(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client, key, id string, dryRun bool) (map[string]any, error) {
	spec, ok := Lookup(key)
	if !ok {
		return nil, fmt.Errorf("未注册的资源: %s", key)
	}
	if !containsOp(spec.Operations, "delete") {
		return nil, fmt.Errorf("资源 %s 不支持删除操作", spec.Label)
	}

	var targetPath string
	switch spec.DeleteMode {
	case "clear":
		targetPath = spec.V4Path
	case "srcfile":
		if strings.TrimSpace(id) == "" {
			return nil, fmt.Errorf("删除备份需要文件名")
		}
		targetPath = spec.V4Path + "?srcfile=" + url.QueryEscape(strings.TrimSpace(id))
	default:
		if strings.TrimSpace(id) == "" {
			return nil, fmt.Errorf("删除记录需要有效的 ID")
		}
		targetPath = spec.V4Path + "/" + strings.TrimSpace(id)
	}

	if dryRun {
		if client == nil {
			return map[string]any{"dry_run": true, "method": http.MethodDelete, "url": "mock://api/v4.0" + targetPath}, nil
		}
		return client.PreviewDryRun(http.MethodDelete, targetPath, nil, nil), nil
	}
	if router.Mock || client == nil {
		return mockDelete(spec, id), nil
	}

	raw, err := client.Delete(ctx, targetPath)
	if err != nil {
		return nil, err
	}
	return ikuaiv4.ExtractObject(raw), nil
}

// ExecuteAction runs a named one-click router action.
func ExecuteAction(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client, actionKey string, dryRun bool) (map[string]any, error) {
	var matched *ActionSpec
	for _, spec := range All() {
		for _, act := range spec.Actions {
			if act.Key == actionKey {
				a := act
				matched = &a
				break
			}
		}
		if matched != nil {
			break
		}
	}
	if matched == nil {
		return nil, fmt.Errorf("未知的系统动作: %s", actionKey)
	}

	if dryRun {
		if client == nil {
			return map[string]any{"dry_run": true, "method": matched.Method, "url": "mock://api/v4.0" + matched.Path}, nil
		}
		return client.PreviewDryRun(matched.Method, matched.Path, nil, map[string]any{}), nil
	}
	if router.Mock || client == nil {
		return map[string]any{"message": "模拟执行成功: " + matched.Label, "action": matched.Key}, nil
	}

	if actionKey == "upgrade-check" {
		raw, postErr := client.Post(ctx, matched.Path, map[string]any{})
		sysVer := ""
		verType := ""
		newVer := ""
		if infoRaw, infoErr := client.Get(ctx, "/system/upgrade", nil); infoErr == nil {
			infoObj := ikuaiv4.ExtractObject(infoRaw)
			sysVer = ikuaiv4.Str(infoObj, "system_ver")
			verType = ikuaiv4.Str(infoObj, "version_type")
			newVer = ikuaiv4.Str(infoObj, "new_system_ver")
		}
		if postErr != nil {
			var apiErr *ikuaiv4.APIError
			if errors.As(postErr, &apiErr) && apiErr.Code == 2019 {
				verLabel := sysVer
				if verLabel != "" && verType != "" {
					verLabel += " " + verType
				}
				msg := "控制台通信正常，但该路由器自身外网未连通，无法连接爱快云更新服务器"
				if verLabel != "" {
					msg = fmt.Sprintf("控制台通信正常（当前固件 v%s），但该路由器自身外网未连通，无法连接爱快云更新服务器", verLabel)
				}
				return map[string]any{
					"warning":       true,
					"upstream_code": 2019,
					"system_ver":    sysVer,
					"message":       msg,
				}, nil
			}
			return nil, postErr
		}
		res := ikuaiv4.ExtractObject(raw)
		if res == nil {
			res = map[string]any{}
		}
		res["system_ver"] = sysVer
		res["new_system_ver"] = newVer
		if newVer != "" {
			res["message"] = fmt.Sprintf("发现新固件版本 v%s（当前 v%s）", newVer, sysVer)
		} else if sysVer != "" {
			res["message"] = fmt.Sprintf("在线检查完成：当前固件 v%s 已是最新版本", sysVer)
		} else {
			res["message"] = "在线检查完成"
		}
		return res, nil
	}

	raw, err := client.Post(ctx, matched.Path, map[string]any{})
	if err != nil {
		return nil, err
	}
	return ikuaiv4.ExtractObject(raw), nil
}

var pemCertificateFields = []string{"ca", "cert", "key", "tls_auth", "leftcert", "rightcert"}

func preparePayload(spec ResourceSpec, input map[string]any, applyDefaults bool) map[string]any {
	out := make(map[string]any, len(input)+8)
	for k, v := range input {
		if integerAPIFields[k] {
			if v == nil {
				continue
			}
			if s, isStr := v.(string); isStr && strings.TrimSpace(s) == "" {
				continue
			}
		}
		out[k] = coerceValue(k, v)
	}

	for _, pemKey := range pemCertificateFields {
		if s, ok := out[pemKey].(string); ok {
			out[pemKey] = encodeIKuaiPEM(s)
		}
	}

	// Object groups convert "value": "1.1.1.1,2.2.2.2" -> group_value: [{key: "1.1.1.1"}, ...]
	if spec.ObjectValueKey != "" {
		if rawVal, ok := out["value"]; ok {
			if s, isStr := rawVal.(string); isStr {
				parts := splitCSV(s)
				gv := make([]map[string]any, 0, len(parts))
				for _, p := range parts {
					gv = append(gv, map[string]any{spec.ObjectValueKey: p})
				}
				out["group_value"] = gv
			}
		}
		delete(out, "value")
	}

	// Compound fields convert CSV strings into {"custom": [...], "object": []}
	for _, field := range spec.CompoundFields {
		if val, exists := out[field]; exists {
			if s, isStr := val.(string); isStr {
				out[field] = map[string]any{
					"custom": splitCSVAsAny(s),
					"object": []any{},
				}
			}
		}
	}

	if spec.Key == "dns-proxy" {
		if _, exists := out["is_ipv6"]; !exists {
			pt, _ := out["parse_type"].(string)
			if pt == "ipv6" || pt == "proxy6" {
				out["is_ipv6"] = 1
			} else {
				out["is_ipv6"] = 0
			}
		}
	}

	if spec.Key == "time-objects" {
		hasAnyTimeField := false
		for _, k := range []string{"type", "weekdays", "start_time", "end_time", "val_comment"} {
			if _, ok := out[k]; ok {
				hasAnyTimeField = true
				break
			}
		}
		if _, hasGV := out["group_value"]; !hasGV && (applyDefaults || hasAnyTimeField) {
			tVal := strOr(out["type"], "weekly")
			weekdays := strOr(out["weekdays"], "1234567")
			startTime := strOr(out["start_time"], "00:00")
			endTime := strOr(out["end_time"], "23:59")
			valComment := strOr(out["val_comment"], "")
			item := map[string]any{
				"type":       tVal,
				"start_time": startTime,
				"end_time":   endTime,
				"comment":    valComment,
			}
			if tVal == "weekly" {
				item["weekdays"] = weekdays
			}
			out["group_value"] = []any{item}
		}
		for _, tmpKey := range []string{"type", "weekdays", "start_time", "end_time", "val_comment", "value"} {
			delete(out, tmpKey)
		}
	}

	if spec.Key == "web-admin-accounts" {
		if plain, ok := out["passwd_plain"].(string); ok {
			plain = strings.TrimSpace(plain)
			if plain != "" {
				sum := md5.Sum([]byte(plain))
				out["passwd"] = hex.EncodeToString(sum[:])
			}
		}
		delete(out, "passwd_plain")
	}

	if applyDefaults && len(spec.Defaults) > 0 {
		for k, def := range spec.Defaults {
			if _, exists := out[k]; !exists {
				out[k] = def
			}
		}
	}
	return out
}

// GetObjectRefs queries rules referencing a given object group name.
func GetObjectRefs(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client, key, groupName string) (map[string]any, error) {
	spec, ok := Lookup(key)
	if !ok || spec.Group != "objects" {
		return nil, fmt.Errorf("未注册的对象组资源: %s", key)
	}
	if router.Mock || client == nil {
		return mockObjectRefs(key, groupName), nil
	}
	raw, err := client.Get(ctx, spec.V4Path+"/ref", map[string]string{"group_name": strings.TrimSpace(groupName)})
	if err != nil {
		return nil, err
	}
	rows, total := ikuaiv4.ExtractCollection(raw)
	return map[string]any{"group_name": groupName, "rows": rows, "total": total}, nil
}

// ListWGPeers lists peers of a WireGuard tunnel.
func ListWGPeers(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client, wgID string) (map[string]any, error) {
	if router.Mock || client == nil {
		return mockListWGPeers(wgID), nil
	}
	raw, err := client.Get(ctx, "/vpn/wireguard/"+url.PathEscape(wgID)+"/peers", nil)
	if err != nil {
		return nil, err
	}
	rows, total := ikuaiv4.ExtractCollection(raw)
	return map[string]any{"rows": rows, "total": total}, nil
}

// CreateWGPeer adds a peer to a WireGuard tunnel.
func CreateWGPeer(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client, wgID string, body map[string]any, dryRun bool) (map[string]any, error) {
	payload := prepareWGPeerPayload(body, true)
	path := "/vpn/wireguard/" + url.PathEscape(wgID) + "/peers"
	if dryRun {
		if client == nil {
			return map[string]any{"dry_run": true, "method": http.MethodPost, "url": "mock://api/v4.0" + path, "body": payload}, nil
		}
		return client.PreviewDryRun(http.MethodPost, path, nil, payload), nil
	}
	if router.Mock || client == nil {
		return mockCreateWGPeer(wgID, payload), nil
	}
	raw, err := client.Post(ctx, path, payload)
	if err != nil {
		return nil, err
	}
	return ikuaiv4.ExtractObject(raw), nil
}

var wireguardPeerInputFields = []string{
	"enabled", "comment", "interface", "peer_publickey", "presharedkey",
	"allowips", "endpoint", "endpoint_port", "keepalive",
}

// GenerateWGKeypair generates a valid RFC 7748 Curve25519 keypair for WireGuard tunnels.
func GenerateWGKeypair() (string, string, error) {
	privBytes := make([]byte, 32)
	if _, err := rand.Read(privBytes); err != nil {
		return "", "", err
	}
	privBytes[0] &= 248
	privBytes[31] &= 127
	privBytes[31] |= 64

	privKey, err := ecdh.X25519().NewPrivateKey(privBytes)
	if err != nil {
		return "", "", err
	}
	pubKey := privKey.PublicKey()
	return base64.StdEncoding.EncodeToString(privKey.Bytes()),
		base64.StdEncoding.EncodeToString(pubKey.Bytes()),
		nil
}

// UpdateWGPeer updates a peer on a WireGuard tunnel.
func UpdateWGPeer(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client, wgID, peerID string, body map[string]any, dryRun bool) (map[string]any, error) {
	updates := prepareWGPeerPayload(body, false)
	path := "/vpn/wireguard/" + url.PathEscape(wgID) + "/peers/" + url.PathEscape(peerID)
	if router.Mock || client == nil {
		if dryRun {
			return map[string]any{"dry_run": true, "method": http.MethodPut, "url": "mock://api/v4.0" + path, "body": updates}, nil
		}
		return mockUpdateWGPeer(wgID, peerID, updates), nil
	}
	existingRaw, err := client.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("读取 WireGuard Peer 配置失败: %w", err)
	}
	existing := ikuaiv4.ExtractObjectByID(existingRaw, peerID)
	if len(existing) == 0 {
		return nil, fmt.Errorf("未找到 ID=%s 的 WireGuard Peer", peerID)
	}
	merged := map[string]any{}
	for _, f := range wireguardPeerInputFields {
		if v, exists := existing[f]; exists {
			merged[f] = v
		}
	}
	for k, v := range updates {
		if containsOp(wireguardPeerInputFields, k) {
			merged[k] = v
		}
	}
	if dryRun {
		return client.PreviewDryRun(http.MethodPut, path, nil, merged), nil
	}
	raw, err := client.Put(ctx, path, merged)
	if err != nil {
		return nil, err
	}
	return ikuaiv4.ExtractObject(raw), nil
}

// ToggleWGPeer toggles enabled on a WireGuard peer.
func ToggleWGPeer(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client, wgID, peerID, enabled string, dryRun bool) (map[string]any, error) {
	val := "yes"
	if strings.EqualFold(enabled, "no") || enabled == "0" || strings.EqualFold(enabled, "false") {
		val = "no"
	}
	body := map[string]any{"enabled": val}
	path := "/vpn/wireguard/" + url.PathEscape(wgID) + "/peers/" + url.PathEscape(peerID)
	if dryRun {
		if client == nil {
			return map[string]any{"dry_run": true, "method": http.MethodPatch, "url": "mock://api/v4.0" + path, "body": body}, nil
		}
		return client.PreviewDryRun(http.MethodPatch, path, nil, body), nil
	}
	if router.Mock || client == nil {
		return mockToggleWGPeer(wgID, peerID, val), nil
	}
	raw, err := client.Patch(ctx, path, body)
	if err != nil {
		return nil, err
	}
	return ikuaiv4.ExtractObject(raw), nil
}

// DeleteWGPeer removes a WireGuard peer.
func DeleteWGPeer(ctx context.Context, router config.RouterConfig, client *ikuaiv4.Client, wgID, peerID string, dryRun bool) (map[string]any, error) {
	path := "/vpn/wireguard/" + url.PathEscape(wgID) + "/peers/" + url.PathEscape(peerID)
	if dryRun {
		if client == nil {
			return map[string]any{"dry_run": true, "method": http.MethodDelete, "url": "mock://api/v4.0" + path}, nil
		}
		return client.PreviewDryRun(http.MethodDelete, path, nil, nil), nil
	}
	if router.Mock || client == nil {
		return mockDeleteWGPeer(wgID, peerID), nil
	}
	raw, err := client.Delete(ctx, path)
	if err != nil {
		return nil, err
	}
	return ikuaiv4.ExtractObject(raw), nil
}

func prepareWGPeerPayload(body map[string]any, applyDefaults bool) map[string]any {
	out := make(map[string]any, len(body)+4)
	for k, v := range body {
		out[k] = coerceValue(k, v)
	}
	if applyDefaults {
		if _, exists := out["enabled"]; !exists {
			out["enabled"] = "yes"
		}
		if _, exists := out["keepalive"]; !exists {
			out["keepalive"] = 0
		}
		if _, exists := out["comment"]; !exists {
			out["comment"] = ""
		}
	}
	return out
}

func strOr(v any, fallback string) string {
	if v == nil {
		return fallback
	}
	s := strings.TrimSpace(fmt.Sprint(v))
	if s == "" || s == "<nil>" {
		return fallback
	}
	return s
}

func flattenTimeGroupValue(row map[string]any) {
	gv, ok := row["group_value"].([]any)
	if !ok || len(gv) == 0 {
		return
	}
	labels := make([]string, 0, len(gv))
	for i, rawItem := range gv {
		item, ok := rawItem.(map[string]any)
		if !ok {
			continue
		}
		t := ikuaiv4.Str(item, "type")
		w := ikuaiv4.Str(item, "weekdays")
		st := ikuaiv4.Str(item, "start_time")
		et := ikuaiv4.Str(item, "end_time")
		c := ikuaiv4.Str(item, "comment")
		if i == 0 {
			if t != "" {
				row["type"] = t
			}
			if w != "" {
				row["weekdays"] = w
			}
			if st != "" {
				row["start_time"] = st
			}
			if et != "" {
				row["end_time"] = et
			}
			if c != "" {
				row["val_comment"] = c
			}
		}
		if w != "" {
			labels = append(labels, fmt.Sprintf("%s %s %s~%s", t, w, st, et))
		} else {
			labels = append(labels, fmt.Sprintf("%s %s~%s", t, st, et))
		}
	}
	if len(labels) > 0 {
		row["value"] = strings.Join(labels, "; ")
	}
}

func fetchUpdateBaseline(ctx context.Context, client *ikuaiv4.Client, spec ResourceSpec, targetPath, id string) (map[string]any, error) {
	if spec.Kind == "collection" && listOnlyResources[spec.Key] {
		listRaw, err := client.Get(ctx, spec.V4Path, map[string]string{"page": "1", "limit": "500"})
		if err != nil {
			return nil, err
		}
		return ikuaiv4.ExtractObjectByID(listRaw, id), nil
	}

	existingRaw, err := client.Get(ctx, targetPath, nil)
	if err == nil {
		if existing := ikuaiv4.ExtractObjectByID(existingRaw, id); len(existing) > 0 {
			return existing, nil
		}
	}
	if spec.Kind != "collection" {
		if err != nil {
			return nil, err
		}
		return map[string]any{}, nil
	}

	listRaw, listErr := client.Get(ctx, spec.V4Path, map[string]string{"page": "1", "limit": "500"})
	if listErr != nil {
		if err != nil {
			return nil, err
		}
		return nil, listErr
	}
	return ikuaiv4.ExtractObjectByID(listRaw, id), nil
}

func buildMergedUpdateBody(spec ResourceSpec, id string, rawInput, updates, existing map[string]any) map[string]any {
	merged := make(map[string]any, len(existing)+len(updates))
	if len(spec.UpdateFields) > 0 {
		for _, f := range spec.UpdateFields {
			if v, exists := existing[f]; exists {
				if spec.Key == "dhcp-services" && shouldOmitEmptyDHCPOptType(f, v) {
					continue
				}
				merged[f] = v
			}
		}
	} else {
		for k, v := range existing {
			if k == "id" && spec.Kind == "singleton" {
				continue
			}
			if strings.HasSuffix(k, "_int") {
				continue
			}
			if (spec.ObjectValueKey != "" || spec.Key == "time-objects") && k == "value" {
				continue
			}
			if spec.Key == "time-objects" && (k == "type" || k == "weekdays" || k == "start_time" || k == "end_time" || k == "val_comment") {
				continue
			}
			merged[k] = v
		}
	}

	normalizeCompoundEmptyMaps(merged)

	for k, v := range updates {
		if len(spec.UpdateFields) > 0 && !containsOp(spec.UpdateFields, k) {
			continue
		}
		merged[k] = v
	}

	// Preserve existing object-group references when visual form submits CSV strings.
	for _, field := range spec.CompoundFields {
		if _, isStringEdit := rawInput[field].(string); !isStringEdit {
			continue
		}
		mergedCompound, okM := merged[field].(map[string]any)
		exCompound, okE := existing[field].(map[string]any)
		if okM && okE {
			if exObj, hasObj := exCompound["object"].([]any); hasObj && len(exObj) > 0 {
				mergedCompound["object"] = append([]any(nil), exObj...)
			}
		}
	}

	// Preserve trailing windows on multi-entry time-objects when edited in visual form.
	if spec.Key == "time-objects" {
		if _, explicitGV := rawInput["group_value"]; !explicitGV {
			if mergedGV, okM := merged["group_value"].([]any); okM && len(mergedGV) == 1 {
				if exGV, okE := existing["group_value"].([]any); okE && len(exGV) > 0 {
					if firstMerged, okFM := mergedGV[0].(map[string]any); okFM {
						if firstEx, okFE := exGV[0].(map[string]any); okFE {
							if _, has := rawInput["type"]; !has && firstEx["type"] != nil {
								firstMerged["type"] = firstEx["type"]
							}
							if _, has := rawInput["start_time"]; !has && firstEx["start_time"] != nil {
								firstMerged["start_time"] = firstEx["start_time"]
							}
							if _, has := rawInput["end_time"]; !has && firstEx["end_time"] != nil {
								firstMerged["end_time"] = firstEx["end_time"]
							}
							if _, has := rawInput["val_comment"]; !has && firstEx["comment"] != nil {
								firstMerged["comment"] = firstEx["comment"]
							}
							if strOr(firstMerged["type"], "weekly") == "weekly" {
								if _, has := rawInput["weekdays"]; !has && firstEx["weekdays"] != nil {
									firstMerged["weekdays"] = firstEx["weekdays"]
								}
							} else {
								delete(firstMerged, "weekdays")
							}
						}
					}
					if len(exGV) > 1 {
						merged["group_value"] = append(mergedGV, exGV[1:]...)
					}
				}
			}
		}
	}

	applyResourceUpdateInvariants(spec, merged)

	if spec.Kind == "collection" && id != "" {
		if _, has := merged["id"]; !has {
			merged["id"] = coerceValue("id", id)
		}
	}
	return merged
}

func applyResourceUpdateInvariants(spec ResourceSpec, merged map[string]any) {
	switch spec.Key {
	case "vlan":
		fillMissingDefaults(merged, map[string]any{
			"enabled": "yes",
			"mac":     "",
			"ip_mask": "",
			"comment": "",
		})
		if vlanName, ok := merged["vlan_name"]; ok && strOr(vlanName, "") != "" {
			merged["tagname"] = vlanName
		}
	case "dhcp-services":
		fillMissingDefaults(merged, map[string]any{
			"enabled":          "yes",
			"delay":            0,
			"check_addr_valid": 0,
			"check_relay_only": 0,
			"opt_type15":       0,
			"opt_type28":       0,
			"opt_type43":       0,
			"opt_type60":       0,
			"opt_type66":       0,
			"opt_type67":       0,
			"opt_type80":       0,
			"opt_type119":      0,
			"opt_type125":      0,
			"opt_type128":      0,
			"opt_type138":      0,
			"opt_type121":      2,
		})
	case "pppoe-config":
		fillMissingDefaults(merged, map[string]any{
			"server_name":       "iKuai",
			"force_verify_name": 0,
			"nas_identifier":    "iKuai",
			"nas_ip_address":    "",
			"authport":          1812,
			"accountport":       1813,
			"rate_limit_lan":    1,
			"drop_client":       1,
			"force_pppoe":       0,
			"enhance_check":     1,
			"share_deny":        0,
			"bind_vlan":         0,
			"verify_vlan":       1,
			"bind_iface":        0,
			"mtu":               1480,
			"mru":               1480,
			"lcp_echo_interval": 10,
			"lcp_echo_failure":  3,
			"maxconnect":        0,
			"restart_timer":     0,
			"comment":           "",
		})
	case "dns-proxy":
		pt := strings.TrimSpace(fmt.Sprint(merged["parse_type"]))
		if pt == "ipv6" || pt == "proxy6" {
			merged["is_ipv6"] = 1
		} else {
			merged["is_ipv6"] = 0
		}
	case "vrrp-config":
		fillMissingDefaults(merged, map[string]any{
			"enabled":          "no",
			"type":             1,
			"prio":             150,
			"method":           0,
			"domain":           "www.baidu.com",
			"dns":              "223.5.5.5",
			"gateway":          "www.baidu.com",
			"interval":         3,
			"ifnames":          "",
			"auto_sync":        0,
			"single_line":      0,
			"ignore_wanstatus": 1,
			"interfaces":       "",
			"virtual_ips":      "",
			"ht_iface":         "",
			"remote_addr":      "",
		})
	case "backup-auto":
		if strOr(merged["time"], "") == "" {
			merged["time"] = "23:59"
		}
		if strOr(merged["cycle_time"], "") == "" {
			switch strOr(merged["strategy"], "week") {
			case "week":
				merged["cycle_time"] = "7"
			case "month":
				merged["cycle_time"] = "1"
			default:
				merged["cycle_time"] = "2026-12-31"
			}
		}
	}
}

func encodeIKuaiPEM(raw string) string {
	if !strings.Contains(raw, "-----BEGIN") {
		return raw
	}
	s := strings.ReplaceAll(raw, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	if strings.Contains(s, "@") {
		s = strings.ReplaceAll(s, "@", "\n")
	}
	rawLines := strings.Split(s, "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "-----BEGIN") || strings.HasPrefix(trimmed, "-----END") {
			normalized := strings.Join(strings.Fields(strings.ReplaceAll(trimmed, "#", " ")), "#")
			lines = append(lines, normalized)
			continue
		}
		cleaned := strings.Map(func(r rune) rune {
			if r == ' ' || r == '\t' {
				return -1
			}
			return r
		}, trimmed)
		if cleaned != "" {
			lines = append(lines, cleaned)
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "@")
}

func decodeIKuaiPEM(raw string) string {
	if !strings.Contains(raw, "-----BEGIN") || !strings.Contains(raw, "@") {
		return raw
	}
	parts := strings.Split(raw, "@")
	lines := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "-----BEGIN") || strings.HasPrefix(trimmed, "-----END") {
			lines = append(lines, strings.ReplaceAll(trimmed, "#", " "))
		} else {
			lines = append(lines, trimmed)
		}
	}
	return strings.Join(lines, "\n")
}

func decodeIKuaiPEMFields(obj map[string]any) {
	if obj == nil {
		return
	}
	for _, k := range pemCertificateFields {
		if s, ok := obj[k].(string); ok && s != "" {
			obj[k] = decodeIKuaiPEM(s)
		}
	}
}

func fillMissingDefaults(target map[string]any, defaults map[string]any) {
	for k, v := range defaults {
		if _, exists := target[k]; !exists {
			target[k] = v
		}
	}
}

func shouldOmitEmptyDHCPOptType(key string, val any) bool {
	s, ok := val.(string)
	if !ok || s != "" {
		return false
	}
	return strings.HasPrefix(key, "opt_type")
}

func normalizeCompoundEmptyMaps(v any) {
	switch typed := v.(type) {
	case map[string]any:
		for k, child := range typed {
			if k == "custom" || k == "object" {
				if emptyMap, ok := child.(map[string]any); ok && len(emptyMap) == 0 {
					typed[k] = []any{}
					continue
				}
			}
			normalizeCompoundEmptyMaps(child)
		}
	case []any:
		for _, child := range typed {
			normalizeCompoundEmptyMaps(child)
		}
	}
}

func normalizeCPUFreqObject(obj map[string]any) {
	if obj == nil {
		return
	}
	if strOr(obj["mode"], "") == "" {
		if cur := ikuaiv4.Str(obj, "current_cpufreq"); cur != "" {
			obj["mode"] = cur
		}
	}
	if _, has := obj["turbo"]; !has {
		if _, hasCur := obj["current_turbo"]; hasCur {
			obj["turbo"] = ikuaiv4.Int(obj, "current_turbo")
		}
	}
}

func filterRowsBySearch(rows []map[string]any, search string) []map[string]any {
	needle := strings.ToLower(strings.TrimSpace(search))
	if needle == "" {
		return rows
	}
	filtered := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		if valueMatchesSearch(row, needle) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func valueMatchesSearch(val any, needle string) bool {
	if val == nil {
		return false
	}
	switch t := val.(type) {
	case map[string]any:
		for _, v := range t {
			if valueMatchesSearch(v, needle) {
				return true
			}
		}
		return false
	case []any:
		for _, v := range t {
			if valueMatchesSearch(v, needle) {
				return true
			}
		}
		return false
	case []map[string]any:
		for _, v := range t {
			if valueMatchesSearch(v, needle) {
				return true
			}
		}
		return false
	default:
		return strings.Contains(strings.ToLower(fmt.Sprint(val)), needle)
	}
}

func paginateRows(rows []map[string]any, page, pageSize int) []map[string]any {
	if len(rows) == 0 {
		return rows
	}
	start := (page - 1) * pageSize
	if start >= len(rows) {
		return []map[string]any{}
	}
	end := start + pageSize
	if end > len(rows) {
		end = len(rows)
	}
	return rows[start:end]
}

func coerceValue(key string, val any) any {
	if !integerAPIFields[key] {
		return val
	}
	if s, ok := val.(string); ok {
		s = strings.TrimSpace(s)
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return n
		}
	}
	return val
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func splitCSVAsAny(s string) []any {
	parts := splitCSV(s)
	out := make([]any, 0, len(parts))
	for _, p := range parts {
		out = append(out, p)
	}
	return out
}

func containsOp(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

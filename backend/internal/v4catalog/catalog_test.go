package v4catalog

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"ikuai-dashboard/backend/internal/config"
	"ikuai-dashboard/backend/internal/ikuaiv4"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestCatalogContainsAllModulesAndPaths(t *testing.T) {
	expected := map[string]string{
		"dnat-rules":         "/network/dnat/rules",
		"stream-l7":          "/routing/app-protocols",
		"l7-rules":           "/security/app-protocols/professional/rules",
		"mac-rules":          "/security/mac-rules",
		"wireguard":          "/vpn/wireguard",
		"wireless-aps":       "/network/ac/ap-config",
		"ip-objects":         "/ip-objects",
		"log-url-visits":     "/log/url-visits",
		"system-basic":       "/system/basic/config",
		"snmpd-config":       "/advanced-service/snmpd-config",
		"auth-online":        "/auth/online-users",
		"auth-web-server":    "/auth/web/services",
		"dhcp-access-mode":   "/network/dhcp/access-control/mode",
		"mac-mode":           "/security/mac-mode",
		"url-keywords":       "/security/url-keywords/rules",
		"pptp-service":       "/vpn/pptp/services",
		"l2tp-service":       "/vpn/l2tp/services",
		"proto-objects":      "/proto-objects",
		"ftp-users":          "/advanced-service/ftp-users",
		"backup-auto":        "/system/backup-auto",
		"system-files":       "/system/files",
		"web-admin-groups":   "/system/web-admin/groups",
		"qos-ip":             "/network/qos/ip",
		"wireless-blacklist": "/wireless/access-control/rules",
	}
	for key, wantPath := range expected {
		spec, ok := Lookup(key)
		if !ok {
			t.Fatalf("missing catalog key %q", key)
		}
		if spec.V4Path != wantPath {
			t.Fatalf("catalog %q path = %q, want %q", key, spec.V4Path, wantPath)
		}
	}
}

func TestPreparePayloadCompoundAndObjects(t *testing.T) {
	aclSpec, _ := Lookup("acl-rules")
	body := preparePayload(aclSpec, map[string]any{
		"tagname":  "block-ssh",
		"dst_port": "22, 2222",
		"prio":     "30",
	}, true)
	if prio, ok := body["prio"].(int64); !ok || prio != 30 {
		t.Fatalf("expected prio coerced to int64(30), got %T (%v)", body["prio"], body["prio"])
	}
	dstPort := body["dst_port"].(map[string]any)
	custom := dstPort["custom"].([]any)
	if len(custom) != 2 || custom[0] != "22" || custom[1] != "2222" {
		t.Fatalf("unexpected compound dst_port: %+v", dstPort)
	}

	objSpec, _ := Lookup("ip-objects")
	objBody := preparePayload(objSpec, map[string]any{
		"group_name": "servers",
		"value":      "192.168.9.10, 192.168.9.11",
	}, true)
	gv := objBody["group_value"].([]map[string]any)
	if len(gv) != 2 || gv[0]["ip"] != "192.168.9.10" {
		t.Fatalf("unexpected object group_value: %+v", gv)
	}

	timeSpec, _ := Lookup("time-objects")
	timeBody := preparePayload(timeSpec, map[string]any{
		"group_name": "work_hours",
		"type":       "weekly",
		"weekdays":   "12345",
		"start_time": "09:00",
		"end_time":   "18:00",
	}, true)
	tgv := timeBody["group_value"].([]any)
	if len(tgv) != 1 || tgv[0].(map[string]any)["weekdays"] != "12345" {
		t.Fatalf("unexpected time-object group_value: %+v", timeBody)
	}

	adminSpec, _ := Lookup("web-admin-accounts")
	adminBody := preparePayload(adminSpec, map[string]any{
		"username":     "ops",
		"passwd_plain": "secret123",
		"group_id":     "1",
	}, true)
	if _, hasPlain := adminBody["passwd_plain"]; hasPlain || adminBody["passwd"] == "" {
		t.Fatalf("expected passwd_plain hashed to passwd and stripped, got %+v", adminBody)
	}
}

func TestUpdateUsesGetThenMergeThenPut(t *testing.T) {
	var capturedMethod string
	var capturedBody map[string]any

	client := ikuaiv4.NewClientWithHTTP("https://192.0.2.1", "tok", &http.Client{
		Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method == http.MethodGet && r.URL.Path == "/api/v4.0/network/dns/config" {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(bytes.NewBufferString(`{"code":0,"message":"Success","data":{"dns1":"223.5.5.5","dns2":"119.29.29.29","cache_ttl":300,"enabled":"yes"}}`)),
				}, nil
			}
			capturedMethod = r.Method
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &capturedBody)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewBufferString(`{"code":0,"message":"Success"}`)),
			}, nil
		}),
	})

	router := config.RouterConfig{ID: "r1", Name: "R1", URL: "https://192.0.2.1", Token: "tok"}
	_, err := Update(context.Background(), router, client, "dns-config", "", map[string]any{"dns1": "1.1.1.1"}, false)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if capturedMethod != http.MethodPut || capturedBody["dns1"] != "1.1.1.1" || capturedBody["dns2"] != "119.29.29.29" {
		t.Fatalf("expected merged PUT preserving dns2, got method=%s body=%+v", capturedMethod, capturedBody)
	}
}

func TestV4CatalogContractsAndUpdatePreservation(t *testing.T) {
	// 1. auth-web-server must be read-only
	authSpec, ok := Lookup("auth-web-server")
	if !ok || len(authSpec.Operations) != 1 || authSpec.Operations[0] != "get" {
		t.Fatalf("expected auth-web-server operations=[get], got %+v", authSpec.Operations)
	}

	// 2. stream-load-balance keeps weight as string (not int)
	lbSpec, _ := Lookup("stream-load-balance")
	lbBody := preparePayload(lbSpec, map[string]any{"tagname": "lb1", "weight": "2"}, true)
	if w, ok := lbBody["weight"].(string); !ok || w != "2" {
		t.Fatalf("expected load-balance weight as string '2', got %T (%v)", lbBody["weight"], lbBody["weight"])
	}

	// 3. cpufreq-mode normalizes current_cpufreq/current_turbo on GET and PUT
	var cpuPut map[string]any
	client := ikuaiv4.NewClientWithHTTP("https://192.0.2.1", "tok", &http.Client{
		Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
			switch r.URL.Path {
			case "/api/v4.0/system/cpufreq/mode":
				if r.Method == http.MethodGet {
					return &http.Response{
						StatusCode: http.StatusOK,
						Header:     make(http.Header),
						Body:       io.NopCloser(bytes.NewBufferString(`{"code":0,"message":"Success","data":{"current_cpufreq":"ondemand","current_turbo":1,"cpufreq_support":["performance","ondemand"]}}`)),
					}, nil
				}
				raw, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(raw, &cpuPut)
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(bytes.NewBufferString(`{"code":0,"message":"Success"}`))}, nil

			case "/api/v4.0/security/acl-rules/7":
				if r.Method == http.MethodGet {
					return &http.Response{
						StatusCode: http.StatusOK,
						Header:     make(http.Header),
						Body:       io.NopCloser(bytes.NewBufferString(`{"code":0,"message":"Success","data":{"id":7,"tagname":"r7","action":"drop","dir":"forward","protocol":"tcp","src_addr":{"custom":["1.1.1.1"],"object":["office_ips"]},"dst_addr":{"custom":{},"object":{}},"src_port":{"custom":{},"object":{}},"dst_port":{"custom":["22"],"object":{}},"src_addr_int":0,"comment":"old"}}`)),
					}, nil
				}
				raw, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(raw, &aclPut)
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(bytes.NewBufferString(`{"code":0,"message":"Success"}`))}, nil

			case "/api/v4.0/advanced-service/ftp-users":
				if r.Method != http.MethodGet {
					t.Fatalf("expected GET list for ftp-users, got %s", r.Method)
				}
				if r.URL.Query().Get("limit") != "500" {
					t.Fatalf("expected limit=500 baseline fetch, got %s", r.URL.RawQuery)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(bytes.NewBufferString(`{"code":0,"message":"Success","data":{"items":[{"id":3,"username":"ftp1","passwd":"h","permission":"ro","home_dir":"/share","upload":0,"download":0,"enabled":"yes","runtime_stat":"idle"}]}}`)),
				}, nil

			case "/api/v4.0/advanced-service/ftp-users/3":
				raw, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(raw, &ftpPut)
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(bytes.NewBufferString(`{"code":0,"message":"Success"}`))}, nil

			default:
				return &http.Response{StatusCode: http.StatusNotFound, Header: make(http.Header), Body: io.NopCloser(bytes.NewBufferString(`{"code":404,"message":"not found"}`))}, nil
			}
		}),
	})

	router := config.RouterConfig{ID: "r1", Name: "R1", URL: "https://192.0.2.1", Token: "tok"}

	cpuRes, err := ListOrGet(context.Background(), router, client, "cpufreq-mode", 1, 20, "", "")
	if err != nil || cpuRes.Config["mode"] != "ondemand" || cpuRes.Config["turbo"] != 1 {
		t.Fatalf("expected cpufreq-mode GET normalized, got res=%+v err=%v", cpuRes.Config, err)
	}
	_, err = Update(context.Background(), router, client, "cpufreq-mode", "", map[string]any{"turbo": "0"}, false)
	if err != nil {
		t.Fatalf("cpufreq-mode update failed: %v", err)
	}
	if cpuPut["mode"] != "ondemand" || int(cpuPut["turbo"].(float64)) != 0 {
		t.Fatalf("expected cpufreq-mode PUT to retain mode and update turbo=0, got %+v", cpuPut)
	}

	// 4. Compound field edit preserves "object" slice and clears comment
	_, err = Update(context.Background(), router, client, "acl-rules", "7", map[string]any{
		"src_addr": "2.2.2.2",
		"comment":  "",
	}, false)
	if err != nil {
		t.Fatalf("acl-rules update failed: %v", err)
	}
	if aclPut["comment"] != "" {
		t.Fatalf("expected comment cleared to '', got %#v", aclPut["comment"])
	}
	if _, hasInt := aclPut["src_addr_int"]; hasInt {
		t.Fatalf("expected *_int field stripped, got %+v", aclPut)
	}
	srcAddr := aclPut["src_addr"].(map[string]any)
	customArr := srcAddr["custom"].([]any)
	objArr := srcAddr["object"].([]any)
	if len(customArr) != 1 || customArr[0] != "2.2.2.2" || len(objArr) != 1 || objArr[0] != "office_ips" {
		t.Fatalf("expected custom updated and object preserved, got %+v", srcAddr)
	}

	// 5. List-only resource Update uses ?page=1&limit=500 and filters via UpdateFields
	_, err = Update(context.Background(), router, client, "ftp-users", "3", map[string]any{"permission": "rw"}, false)
	if err != nil {
		t.Fatalf("ftp-users update failed: %v", err)
	}
	if _, hasRuntime := ftpPut["runtime_stat"]; hasRuntime || ftpPut["permission"] != "rw" {
		t.Fatalf("expected ftp-users PUT filtered via UpdateFields, got %+v", ftpPut)
	}
}

var aclPut map[string]any
var ftpPut map[string]any

func TestGenerateWGKeypairValidX25519(t *testing.T) {
	privB64, pubB64, err := GenerateWGKeypair()
	if err != nil {
		t.Fatalf("GenerateWGKeypair failed: %v", err)
	}
	privBytes, err := base64.StdEncoding.DecodeString(privB64)
	if err != nil || len(privBytes) != 32 {
		t.Fatalf("invalid private key %q: err=%v len=%d", privB64, err, len(privBytes))
	}
	pubBytes, err := base64.StdEncoding.DecodeString(pubB64)
	if err != nil || len(pubBytes) != 32 {
		t.Fatalf("invalid public key %q: err=%v len=%d", pubB64, err, len(pubBytes))
	}
	priv, err := ecdh.X25519().NewPrivateKey(privBytes)
	if err != nil {
		t.Fatalf("NewPrivateKey failed: %v", err)
	}
	if !bytes.Equal(priv.PublicKey().Bytes(), pubBytes) {
		t.Fatalf("public key mismatch: got %x want %x", pubBytes, priv.PublicKey().Bytes())
	}
}

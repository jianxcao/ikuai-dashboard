package ikuaiv4

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func makeResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

func TestNormalizeBaseURL(t *testing.T) {
	cases := map[string]string{
		"192.168.9.1":        "https://192.168.9.1",
		"https://10.0.0.1/":  "https://10.0.0.1",
		"http://router.lan/": "http://router.lan",
		"  ":                 "",
	}
	for input, want := range cases {
		if got := NormalizeBaseURL(input); got != want {
			t.Fatalf("NormalizeBaseURL(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSanitizeNilPreservesStrings(t *testing.T) {
	raw := []byte(`{"a":nil,"b":"keep nil inside string","c":[nil,1,"nil"],"d":null}`)
	cleaned := SanitizeNil(raw)
	var parsed map[string]any
	if err := json.Unmarshal(cleaned, &parsed); err != nil {
		t.Fatalf("unmarshal sanitized JSON failed: %v (%s)", err, string(cleaned))
	}
	if parsed["a"] != nil {
		t.Fatalf("expected a=nil, got %v", parsed["a"])
	}
	if parsed["b"] != "keep nil inside string" {
		t.Fatalf("string value was modified: %v", parsed["b"])
	}
	arr := parsed["c"].([]any)
	if arr[0] != nil || arr[2] != "nil" {
		t.Fatalf("unexpected array after sanitize: %#v", arr)
	}
}

func TestClientHandlesResultsAndAPIError(t *testing.T) {
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			return makeResponse(http.StatusUnauthorized, `{"code":3007,"message":"Invalid token"}`), nil
		}
		if r.URL.Path == "/api/v4.0/monitoring/system" {
			return makeResponse(http.StatusOK, `{"code":0,"message":"Success","results":{"sysinfo":{"hostname":"iKuai","cpu":["1.5%"],"note":nil}}}`), nil
		}
		return makeResponse(http.StatusOK, `{"code":0,"message":"Success","rowid":42}`), nil
	})

	client := NewClientWithHTTP("https://192.0.2.10", "test-token", &http.Client{Transport: rt})
	raw, err := client.Get(context.Background(), "/monitoring/system", nil)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	obj := ExtractObject(raw)
	if Str(obj, "hostname") != "iKuai" {
		t.Fatalf("expected hostname=iKuai, got %+v", obj)
	}

	created, err := client.Post(context.Background(), "/network/vlan", map[string]any{"vlan_id": 10})
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	createdObj := ExtractObject(created)
	if Int(createdObj, "rowid") != 42 {
		t.Fatalf("expected rowid=42, got %+v", createdObj)
	}

	badClient := NewClientWithHTTP("https://192.0.2.10", "wrong", &http.Client{Transport: rt})
	_, err = badClient.Get(context.Background(), "/monitoring/system", nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != 3007 {
		t.Fatalf("expected APIError code 3007, got %v", err)
	}

	offlineRT := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return makeResponse(http.StatusInternalServerError, `{"code":2019,"message":"网络连接失败","errors":"连接数据服务器失败","details":[]}`), nil
	})
	offlineClient := NewClientWithHTTP("https://192.0.2.10", "test-token", &http.Client{Transport: offlineRT})
	_, err = offlineClient.Post(context.Background(), "/system/upgrade:check", map[string]any{})
	if !errors.As(err, &apiErr) || apiErr.Code != 2019 || !strings.Contains(apiErr.Message, "连接数据服务器失败") {
		t.Fatalf("expected 2019 APIError with errors detail, got %v", err)
	}
}

func TestExtractCollectionShapes(t *testing.T) {
	rows, total := ExtractCollection(json.RawMessage(`{"data":[{"id":1,"name":"r1"}],"total":99}`))
	if total != 99 || len(rows) != 1 || Str(rows[0], "name") != "r1" {
		t.Fatalf("unexpected data+total extract: rows=%+v total=%d", rows, total)
	}

	rows, total = ExtractCollection(json.RawMessage(`{"ether_info":{"eth1":{"link":1},"eth0":{"link":1}}}`))
	if total != 2 || len(rows) != 2 || Str(rows[0], "name") != "eth0" || Str(rows[1], "name") != "eth1" {
		t.Fatalf("unexpected map-of-maps extract: rows=%+v total=%d", rows, total)
	}

	for _, emptyPayload := range []string{
		`{"data":{}}`,
		`{"data":null,"total":0}`,
		`{"proto3_day":{},"proto3_day_total":0,"proto3_day_total_flow":null}`,
	} {
		rows, total = ExtractCollection(json.RawMessage(emptyPayload))
		if len(rows) != 0 || total != 0 {
			t.Fatalf("expected empty extract for %s, got rows=%+v total=%d", emptyPayload, rows, total)
		}
	}
}

func TestPreviewDryRunRedactsToken(t *testing.T) {
	client := NewClient("192.0.2.10", "super-secret-token", true)
	preview := client.PreviewDryRun(http.MethodPost, "/network/dnat/rules", nil, map[string]any{"wan_port": "8080"})
	if preview["url"] != "https://192.0.2.10/api/v4.0/network/dnat/rules" {
		t.Fatalf("unexpected url: %v", preview["url"])
	}
	headers := preview["headers"].(map[string]string)
	if headers["Authorization"] == "Bearer super-secret-token" {
		t.Fatalf("token leaked into dry-run preview")
	}
}

func TestExtractObjectUnwrapsDataArrayAndHTML404(t *testing.T) {
	obj := ExtractObject(json.RawMessage(`{"data":[{"id":1,"hostname":"iKuai","dns1":"223.5.5.5"}]}`))
	if Str(obj, "hostname") != "iKuai" || Str(obj, "dns1") != "223.5.5.5" {
		t.Fatalf("expected 1-element data array unwrapped, got %+v", obj)
	}

	picked := ExtractObjectByID(json.RawMessage(`{"data":[{"id":1,"name":"first"},{"id":2,"name":"second"}]}`), "2")
	if Str(picked, "name") != "second" {
		t.Fatalf("expected id=2 matched inside data array, got %+v", picked)
	}

	staticPicked := ExtractObjectByID(json.RawMessage(`{"static_data":[{"id":1,"tagname":"nas","mac":"aa:bb:cc:11:22:33"}],"static_total":1}`), "1")
	if Str(staticPicked, "tagname") != "nas" {
		t.Fatalf("expected static_data unwrapped cleanly, got %+v", staticPicked)
	}

	customPicked := ExtractObjectByID(json.RawMessage(`{"custom_data":[{"id":9,"tagname":"edge"}]}`), "9")
	if Str(customPicked, "tagname") != "edge" {
		t.Fatalf("expected *_data suffix unwrapped cleanly, got %+v", customPicked)
	}

	empty := ExtractObject(json.RawMessage(`{"data":[]}`))
	if len(empty) != 0 {
		t.Fatalf("expected empty map for empty data array, got %+v", empty)
	}

	rows, total := ExtractCollection(json.RawMessage(`{"groups_data":[{"id":1,"group_name":"Administrators"}],"groups_total":1}`))
	if total != 1 || len(rows) != 1 || Str(rows[0], "group_name") != "Administrators" {
		t.Fatalf("expected groups_data unwrapped cleanly, got rows=%+v total=%d", rows, total)
	}

	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return makeResponse(http.StatusNotFound, `<html><body>404 Not Found</body></html>`), nil
	})
	client := NewClientWithHTTP("https://192.0.2.10", "tok", &http.Client{Transport: rt})
	_, err := client.Get(context.Background(), "/vpn/ikev2/services", nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.HTTPStatus != 404 {
		t.Fatalf("expected APIError 404 on HTML 404, got %v", err)
	}
}

func TestClientHandlesEmpty200Body(t *testing.T) {
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return makeResponse(http.StatusOK, "   "), nil
	})
	client := NewClientWithHTTP("https://192.0.2.10", "tok", &http.Client{Transport: rt})
	raw, err := client.Delete(context.Background(), "/wireless/access-control/rules/1")
	if err != nil {
		t.Fatalf("expected empty 200 body to succeed, got %v", err)
	}
	if obj := ExtractObject(raw); Str(obj, "message") != "Success" {
		t.Fatalf("unexpected object from empty 200: %+v", obj)
	}
}

package fleet

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"ikuai-dashboard/backend/internal/config"
	"ikuai-dashboard/backend/internal/ikuaiv4"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestProbeAllAndResolveRouter(t *testing.T) {
	orig := config.Snapshot()
	defer func() {
		if orig != nil {
			config.ReplaceGlobal(orig)
		}
	}()

	cfg := &config.AppConfig{
		Server:         config.ServerConfig{Port: "8080"},
		ActiveRouterID: "node-a",
		Routers: []config.RouterConfig{
			{ID: "node-a", Name: "Node A", URL: "https://192.0.2.1", Token: "tok-a", Source: "yaml"},
			{ID: "node-b", Name: "Node B", URL: "https://192.0.2.2", Mock: true, Source: "mock"},
		},
	}
	config.ReplaceGlobal(cfg)

	mgr := NewManager()
	mgr.SetClientFactory(func(r config.RouterConfig) *ikuaiv4.Client {
		return ikuaiv4.NewClientWithHTTP(r.URL, r.Token, &http.Client{
			Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
				body := `{"code":0,"message":"Success","results":{"sysinfo":{"hostname":"iKuai-A","uptime":999,"cpu":["12.5%"],"memory":{"used":"44%"},"stream":{"upload":1024,"download":4096,"connect_num":55},"online_user":{"count":8},"verinfo":{"version":"4.0.314","verstring":"4.0.314 x64","arch":"x86","sysbit":"x64"}}}}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(bytes.NewBufferString(body)),
				}, nil
			}),
		})
	})

	health := mgr.ProbeAll(context.Background())
	if len(health) != 2 {
		t.Fatalf("expected 2 health items, got %d", len(health))
	}
	if !health[0].Online || !health[0].Active || health[0].Hostname != "iKuai-A" || health[0].CPUPercent != 12.5 {
		t.Fatalf("unexpected real router health: %+v", health[0])
	}
	if !health[1].Online || !health[1].Mock {
		t.Fatalf("expected mock router online, got %+v", health[1])
	}

	router, client, err := mgr.ResolveRouter("node-b")
	if err != nil || router.ID != "node-b" || client != nil {
		t.Fatalf("unexpected override resolve: router=%+v client=%v err=%v", router, client, err)
	}

	var seenAuth string
	mgr.SetClientFactory(func(r config.RouterConfig) *ikuaiv4.Client {
		return ikuaiv4.NewClientWithHTTP(r.URL, r.Token, &http.Client{
			Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
				seenAuth = req.Header.Get("Authorization")
				body := `{"code":0,"message":"Success","results":{"sysinfo":{"hostname":"iKuai-A","uptime":999,"cpu":["12.5%"],"memory":{"used":"44%"},"stream":{"upload":1024,"download":4096,"connect_num":55},"online_user":{"count":8},"verinfo":{"version":"4.0.314","verstring":"4.0.314 x64","arch":"x86","sysbit":"x64"}}}}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(bytes.NewBufferString(body)),
				}, nil
			}),
		})
	})

	tested, err := mgr.TestRouter(context.Background(), RouterTestRequest{
		ID:         "node-a-renamed",
		OriginalID: "node-a",
		Name:       "Node A Renamed",
		URL:        "https://192.0.2.1",
	})
	if err != nil || !tested.Online || seenAuth != "Bearer tok-a" {
		t.Fatalf("expected TestRouter to preserve saved token via OriginalID, got tested=%+v err=%v auth=%q", tested, err, seenAuth)
	}
}

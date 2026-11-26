package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"ikuai-dashboard/backend/internal/config"

	"github.com/gin-gonic/gin"
)

func setupIPTestRouter(t *testing.T, readOnly bool) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	orig := config.Snapshot()
	t.Cleanup(func() {
		if orig != nil {
			config.ReplaceGlobal(orig)
		}
	})

	t.Setenv("DISABLE_AUTH", "false")
	if readOnly {
		t.Setenv("READ_ONLY", "true")
	} else {
		t.Setenv("READ_ONLY", "false")
	}

	cfg := &config.AppConfig{
		Server: config.ServerConfig{
			Port:        "8080",
			AccessToken: "test-access-token",
			ReadOnly:    readOnly,
		},
		ActiveRouterID: "main",
		Routers: []config.RouterConfig{
			{ID: "main", Name: "主路由节点", URL: "https://192.0.2.1", Mock: true, Source: "mock"},
			{ID: "lab-v4", Name: "EdgeLab", URL: "https://192.0.2.2", Mock: true, Source: "mock"},
		},
	}
	config.ReplaceGlobal(cfg)

	r := gin.New()
	api := r.Group("/api/v1", TokenAuthMiddleware(), ReadOnlyMiddleware())
	api.GET("/ip", GetWanIPHandler)
	return r
}

func TestGetWanIPHandler(t *testing.T) {
	r := setupIPTestRouter(t, false)

	t.Run("matches router by id and returns plain-text default wan1 IP", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/ip?name=main&token=test-access-token", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK || rec.Body.String() != "203.0.113.10" {
			t.Fatalf("got status=%d body=%q", rec.Code, rec.Body.String())
		}
	})

	t.Run("matches router by Chinese name and wan=wan2", func(t *testing.T) {
		q := url.Values{}
		q.Set("name", "主路由节点")
		q.Set("wan", "wan2")
		q.Set("token", "test-access-token")
		req := httptest.NewRequest(http.MethodGet, "/api/v1/ip?"+q.Encode(), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK || rec.Body.String() != "198.51.100.22" {
			t.Fatalf("got status=%d body=%q", rec.Code, rec.Body.String())
		}
	})

	t.Run("supports legacy lan=wan2 query param", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/ip?name=main&lan=wan2&token=test-access-token", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK || rec.Body.String() != "198.51.100.22" {
			t.Fatalf("got status=%d body=%q", rec.Code, rec.Body.String())
		}
	})

	t.Run("supports router_id and format=json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/ip?router_id=lab-v4&wan=wan1&format=json&token=test-access-token", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("got status=%d body=%q", rec.Code, rec.Body.String())
		}
		var payload struct {
			Code int `json:"code"`
			Data struct {
				RouterID  string `json:"router_id"`
				Interface string `json:"interface"`
				IP        string `json:"ip"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("invalid json: %v", err)
		}
		if payload.Code != 200 || payload.Data.RouterID != "lab-v4" || payload.Data.Interface != "wan1" || payload.Data.IP != "198.51.100.88" {
			t.Fatalf("unexpected json payload: %+v", payload)
		}
	})

	t.Run("rejects missing or invalid token with 401", func(t *testing.T) {
		for _, target := range []string{"/api/v1/ip?name=main", "/api/v1/ip?name=main&token=wrong"} {
			req := httptest.NewRequest(http.MethodGet, target, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("GET %s status=%d, want 401", target, rec.Code)
			}
		}
	})

	t.Run("returns 404 for unknown router or unknown wan", func(t *testing.T) {
		for _, target := range []string{
			"/api/v1/ip?name=missing&token=test-access-token",
			"/api/v1/ip?name=main&wan=wan9&token=test-access-token",
		} {
			req := httptest.NewRequest(http.MethodGet, target, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != http.StatusNotFound {
				t.Fatalf("GET %s status=%d, want 404", target, rec.Code)
			}
		}
	})

	t.Run("allows GET /api/v1/ip when READ_ONLY=true", func(t *testing.T) {
		roRouter := setupIPTestRouter(t, true)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/ip?name=main&wan=wan2&token=test-access-token", nil)
		rec := httptest.NewRecorder()
		roRouter.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK || rec.Body.String() != "198.51.100.22" {
			t.Fatalf("got status=%d body=%q", rec.Code, rec.Body.String())
		}
	})
}

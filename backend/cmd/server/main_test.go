package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterStaticRoutesCacheHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	assetsDir := filepath.Join(dir, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>v4</html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sw.js"), []byte("// sw"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "app.abc123.js"), []byte("console.log(1)"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	registerStaticRoutes(r, dir)

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantCache  string
	}{
		{
			name:       "root index has no-cache",
			path:       "/",
			wantStatus: http.StatusOK,
			wantCache:  "no-cache, no-store, must-revalidate",
		},
		{
			name:       "service worker has no-cache",
			path:       "/sw.js",
			wantStatus: http.StatusOK,
			wantCache:  "no-cache, no-store, must-revalidate",
		},
		{
			name:       "spa fallback route has no-cache",
			path:       "/fleet/settings",
			wantStatus: http.StatusOK,
			wantCache:  "no-cache, no-store, must-revalidate",
		},
		{
			name:       "hashed asset uses immutable cache",
			path:       "/assets/app.abc123.js",
			wantStatus: http.StatusOK,
			wantCache:  "public, max-age=31536000, immutable",
		},
		{
			name:       "unknown api returns 404 json",
			path:       "/api/unknown",
			wantStatus: http.StatusNotFound,
			wantCache:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("GET %s status = %d, want %d", tc.path, rec.Code, tc.wantStatus)
			}
			gotCache := rec.Header().Get("Cache-Control")
			if gotCache != tc.wantCache {
				t.Fatalf("GET %s Cache-Control = %q, want %q", tc.path, gotCache, tc.wantCache)
			}
		})
	}
}

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadParsesYAMLAndAppliesDefaults(t *testing.T) {
	cfg, err := Load([]byte(`
active_router_id: office
routers:
  - id: office
    name: 办公室爱快
    url: https://192.168.50.1:6443
    username: admin
    password: secret
    mock: false
    insecure_skip_verify: true
`))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Server.Port != "8080" {
		t.Fatalf("expected default port 8080, got %q", cfg.Server.Port)
	}
	if cfg.Server.StaticDir != "frontend/dist" {
		t.Fatalf("expected default static dir frontend/dist, got %q", cfg.Server.StaticDir)
	}

	router, ok := cfg.ActiveRouter()
	if !ok {
		t.Fatal("expected active router to resolve")
	}
	if router.ID != "office" || router.Password != "secret" || router.Version != RouterVersionV3 || router.Mock {
		t.Fatalf("unexpected active router: %+v", router)
	}
}

func TestDefaultConfigStartsWithoutMockRouter(t *testing.T) {
	cfg := DefaultConfig()

	if len(cfg.Routers) != 0 {
		t.Fatalf("expected empty router list for first-run config, got %+v", cfg.Routers)
	}
	if cfg.ActiveRouterID != "" {
		t.Fatalf("expected empty active router id, got %q", cfg.ActiveRouterID)
	}
	if cfg.Server.Port != defaultPort || cfg.Server.StaticDir != defaultStaticDir {
		t.Fatalf("unexpected default server config: %+v", cfg.Server)
	}
}

func TestLoadAllowsFirstRunEmptyRouterConfig(t *testing.T) {
	cfg, err := Load([]byte(`
server:
  port: "8080"
  static_dir: "frontend/dist"
active_router_id: ""
routers: []
`))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(cfg.Routers) != 0 || cfg.ActiveRouterID != "" {
		t.Fatalf("unexpected first-run config: %+v", cfg)
	}
}

func TestPublicFirstRunConfigReturnsEmptyRouterList(t *testing.T) {
	public := DefaultConfig().Public()

	if public.Routers == nil {
		t.Fatal("expected public routers to be an empty list, got nil")
	}
	if len(public.Routers) != 0 {
		t.Fatalf("expected no public routers, got %+v", public.Routers)
	}
}

func TestSaveToFileAllowsFirstRealRouterFromEmptyConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	cfg := &AppConfig{
		ActiveRouterID: "office",
		Routers: []RouterConfig{
			{
				ID:                 "office",
				Name:               "办公室爱快",
				URL:                "https://192.168.50.1:6443",
				Version:            RouterVersionV3,
				Username:           "admin",
				Password:           "secret",
				Mock:               false,
				InsecureSkipVerify: true,
			},
		},
	}

	if err := SaveToFile(path, cfg); err != nil {
		t.Fatalf("SaveToFile returned error: %v", err)
	}

	reloaded, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile returned error: %v", err)
	}
	if len(reloaded.Routers) != 1 || reloaded.ActiveRouterID != "office" {
		t.Fatalf("unexpected reloaded config: %+v", reloaded)
	}
}

func TestLoadParsesV4TokenRouter(t *testing.T) {
	cfg, err := Load([]byte(`
active_router_id: office-v4
routers:
  - id: office-v4
    name: 办公室爱快 v4
    url: https://192.168.50.1:6443
    version: v4
    token: token-secret
    mock: false
    insecure_skip_verify: true
`))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	router, ok := cfg.ActiveRouter()
	if !ok {
		t.Fatal("expected active router to resolve")
	}
	if router.Version != RouterVersionV4 || router.Token != "token-secret" || router.Username != "" {
		t.Fatalf("unexpected v4 router: %+v", router)
	}
}

func TestValidateRejectsV4RouterWithoutToken(t *testing.T) {
	_, err := Load([]byte(`
active_router_id: office-v4
routers:
  - id: office-v4
    name: 办公室爱快 v4
    url: https://192.168.50.1:6443
    version: v4
    mock: false
    insecure_skip_verify: true
`))
	if err == nil {
		t.Fatal("expected missing token error")
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	_, err := Load([]byte(`
active_router_id: office
unknown: value
routers:
  - id: office
    name: 办公室爱快
    url: https://192.168.50.1:6443
    username: admin
    mock: true
`))
	if err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestValidateRejectsDuplicateRouterIDs(t *testing.T) {
	_, err := Load([]byte(`
active_router_id: office
routers:
  - id: office
    name: 办公室爱快
    url: https://192.168.50.1:6443
    username: admin
    mock: true
  - id: office
    name: 备用爱快
    url: https://192.168.60.1:6443
    username: admin
    mock: true
`))
	if err == nil {
		t.Fatal("expected duplicate router id error")
	}
}

func TestMergeRouterSecretsKeepsExistingPasswordWhenIncomingBlank(t *testing.T) {
	existing := &AppConfig{
		ActiveRouterID: "office",
		Routers: []RouterConfig{
			{ID: "office", Name: "办公室爱快", URL: "https://192.168.50.1", Version: RouterVersionV3, Username: "admin", Password: "secret"},
		},
	}
	incoming := &AppConfig{
		ActiveRouterID: "office",
		Routers: []RouterConfig{
			{ID: "office", Name: "办公室爱快", URL: "https://192.168.50.1", Version: RouterVersionV3, Username: "admin", Password: ""},
		},
	}

	incoming.MergeRouterSecrets(existing)

	if got := incoming.Routers[0].Password; got != "secret" {
		t.Fatalf("expected preserved password, got %q", got)
	}
}

func TestMergeRouterSecretsKeepsExistingTokenWhenIncomingBlank(t *testing.T) {
	existing := &AppConfig{
		ActiveRouterID: "office-v4",
		Routers: []RouterConfig{
			{ID: "office-v4", Name: "办公室爱快 v4", URL: "https://192.168.50.1", Version: RouterVersionV4, Token: "token-secret"},
		},
	}
	incoming := &AppConfig{
		ActiveRouterID: "office-v4",
		Routers: []RouterConfig{
			{ID: "office-v4", Name: "办公室爱快 v4", URL: "https://192.168.50.1", Version: RouterVersionV4, Token: ""},
		},
	}

	incoming.MergeRouterSecrets(existing)

	if got := incoming.Routers[0].Token; got != "token-secret" {
		t.Fatalf("expected preserved token, got %q", got)
	}
}

func TestSaveToFileWritesValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	cfg := &AppConfig{
		Server:         ServerConfig{Port: "8081", StaticDir: "public"},
		ActiveRouterID: "office",
		Routers: []RouterConfig{
			{ID: "office", Name: "办公室爱快", URL: "https://192.168.50.1", Version: RouterVersionV3, Username: "admin", Password: "secret", Mock: false, InsecureSkipVerify: true},
		},
	}

	if err := SaveToFile(path, cfg); err != nil {
		t.Fatalf("SaveToFile returned error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file to exist: %v", err)
	}

	reloaded, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile returned error: %v", err)
	}
	if reloaded.Server.Port != "8081" || reloaded.Routers[0].Password != "secret" {
		t.Fatalf("unexpected reloaded config: %+v", reloaded)
	}
}

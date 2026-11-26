package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadParsesV4YAMLAndAppliesDefaults(t *testing.T) {
	cfg, err := Load([]byte(`
active_router_id: office-v4
routers:
  - id: office-v4
    name: 办公室爱快 V4
    url: 192.168.9.1
    token: token-secret
    tags:
      - 核心
      - 双线
    mock: false
    insecure_skip_verify: true
`))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Server.Port != "8080" || cfg.Server.StaticDir != "frontend/dist" {
		t.Fatalf("unexpected default server config: %+v", cfg.Server)
	}

	router, ok := cfg.ActiveRouter()
	if !ok {
		t.Fatal("expected active router to resolve")
	}
	if router.URL != "https://192.168.9.1" || router.Token != "token-secret" || len(router.Tags) != 2 {
		t.Fatalf("unexpected active router: %+v", router)
	}

	pub := cfg.Public()
	if len(pub.Routers) != 1 || !pub.Routers[0].HasToken || pub.Routers[0].Source != "yaml" {
		t.Fatalf("unexpected public config: %+v", pub)
	}
}

func TestValidateRejectsRealRouterWithoutToken(t *testing.T) {
	_, err := Load([]byte(`
active_router_id: office-v4
routers:
  - id: office-v4
    name: 办公室爱快 V4
    url: https://192.168.9.1
    mock: false
    insecure_skip_verify: true
`))
	if err == nil {
		t.Fatal("expected missing token error")
	}
}

func TestLoadRejectsUnknownLegacyFields(t *testing.T) {
	_, err := Load([]byte(`
active_router_id: office
routers:
  - id: office
    name: Legacy
    url: https://192.168.9.1
    version: v3
    mock: true
`))
	if err == nil {
		t.Fatal("expected unknown field error for removed version field")
	}
}

func TestMergeRouterSecretsPreservesToken(t *testing.T) {
	existing := &AppConfig{
		Server: ServerConfig{
			Port:        "9090",
			StaticDir:   "/srv/custom-dist",
			AccessToken: "dash-token",
			ReadOnly:    true,
			Username:    "ops",
			Password:    "secret",
		},
		ActiveRouterID: "office-v4",
		Routers: []RouterConfig{
			{ID: "office-v4", Name: "Office", URL: "https://192.168.9.1", Token: "router-secret"},
		},
	}
	incoming := &AppConfig{
		Server: ServerConfig{
			ReadOnly: false,
		},
		ActiveRouterID: "office-v4",
		Routers: []RouterConfig{
			{ID: "office-v4", Name: "Office", URL: "https://192.168.9.1", Token: ""},
		},
	}

	incoming.MergeRouterSecrets(existing)
	if incoming.Routers[0].Token != "router-secret" || incoming.Server.AccessToken != "dash-token" {
		t.Fatalf("expected secrets preserved, got %+v", incoming)
	}
	if incoming.Server.Port != "9090" || incoming.Server.StaticDir != "/srv/custom-dist" {
		t.Fatalf("expected custom Port/StaticDir preserved, got %+v", incoming.Server)
	}
	if incoming.Server.Username != "ops" || incoming.Server.Password != "secret" {
		t.Fatalf("expected username/password preserved, got %+v", incoming.Server)
	}
	if incoming.Server.ReadOnly {
		t.Fatalf("expected explicit ReadOnly=false not forced back to true, got %+v", incoming.Server)
	}

	renamed := &AppConfig{
		ActiveRouterID: "home-gw",
		Routers: []RouterConfig{
			{ID: "home-gw", OriginalID: "office-v4", Name: "Home GW", URL: "https://192.168.9.1", Token: ""},
		},
	}
	renamed.MergeRouterSecrets(existing)
	if renamed.Routers[0].Token != "router-secret" {
		t.Fatalf("expected renamed router token preserved via OriginalID, got %+v", renamed.Routers[0])
	}
}

func TestLoadRuntimeConfigEnvInjectionAndMockMode(t *testing.T) {
	dir := t.TempDir()
	missingPath := filepath.Join(dir, "config.yaml")

	t.Setenv("MOCK_MODE", "true")
	cfg, _, err := LoadRuntimeConfig(DefaultConfigPath)
	if err != nil {
		t.Fatalf("LoadRuntimeConfig mock failed: %v", err)
	}
	if len(cfg.Routers) != 2 || cfg.ActiveRouterID != "hq-v4" {
		t.Fatalf("expected 2 mock routers, got %+v", cfg.Routers)
	}

	t.Setenv("MOCK_MODE", "false")
	t.Setenv("IKUAI_CLI_BASE_URL", "192.0.2.10")
	t.Setenv("IKUAI_CLI_TOKEN", "ephemeral-token")
	cfg, _, err = LoadRuntimeConfig(DefaultConfigPath)
	if err != nil {
		// If local config.yaml happened not to exist, this tests default fallback.
		t.Fatalf("LoadRuntimeConfig env failed: %v", err)
	}
	if len(cfg.Routers) == 0 || cfg.Routers[0].Token != "ephemeral-token" || cfg.Routers[0].Source != "env" {
		t.Fatalf("expected env-v4 router, got %+v", cfg.Routers)
	}

	if err := SaveToFile(missingPath, cfg); err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}
	if _, err := os.Stat(missingPath); err != nil {
		t.Fatalf("expected file to be written: %v", err)
	}
}

func TestReadOnlyModeEnvAndYAML(t *testing.T) {
	t.Setenv("MOCK_MODE", "true")
	t.Setenv("READ_ONLY", "true")
	cfg, _, err := LoadRuntimeConfig(DefaultConfigPath)
	if err != nil {
		t.Fatalf("LoadRuntimeConfig failed: %v", err)
	}
	if !cfg.Server.ReadOnly || !cfg.Public().Server.ReadOnly || !IsReadOnly() {
		t.Fatalf("expected read_only=true when READ_ONLY=true, got %+v", cfg.Public().Server)
	}

	t.Setenv("READ_ONLY", "false")
	if IsReadOnly() {
		t.Fatal("expected IsReadOnly()=false when READ_ONLY=false")
	}
}

func TestFindRouterByIDOrName(t *testing.T) {
	cfg := &AppConfig{
		ActiveRouterID: "main",
		Routers: []RouterConfig{
			{ID: "main", Name: "总部核心网关", URL: "https://192.168.1.1", Mock: true},
			{ID: "edge-lab", Name: "EdgeLab", URL: "https://10.10.0.1", Mock: true},
		},
	}

	if r, ok := cfg.FindRouterByIDOrName("main"); !ok || r.ID != "main" {
		t.Fatalf("expected exact ID match for main, got ok=%v r=%+v", ok, r)
	}
	if r, ok := cfg.FindRouterByIDOrName("总部核心网关"); !ok || r.ID != "main" {
		t.Fatalf("expected exact Name match, got ok=%v r=%+v", ok, r)
	}
	if r, ok := cfg.FindRouterByIDOrName("MAIN"); !ok || r.ID != "main" {
		t.Fatalf("expected case-insensitive ID match, got ok=%v r=%+v", ok, r)
	}
	if r, ok := cfg.FindRouterByIDOrName("edgelab"); !ok || r.ID != "edge-lab" {
		t.Fatalf("expected case-insensitive Name match, got ok=%v r=%+v", ok, r)
	}
	if _, ok := cfg.FindRouterByIDOrName("nonexistent"); ok {
		t.Fatal("expected unknown router key to return false")
	}
	if _, ok := cfg.FindRouterByIDOrName("   "); ok {
		t.Fatal("expected blank router key to return false")
	}
}

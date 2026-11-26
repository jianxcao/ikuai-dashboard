package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"ikuai-dashboard/backend/internal/ikuaiv4"

	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigPath = "config.yaml"
	defaultPort       = "8080"
	defaultStaticDir  = "frontend/dist"
)

// ServerConfig 保存后端运行参数。
type ServerConfig struct {
	Port        string `yaml:"port" json:"port"`
	StaticDir   string `yaml:"static_dir" json:"static_dir"`
	AccessToken string `yaml:"access_token,omitempty" json:"access_token,omitempty"`
	ReadOnly    bool   `yaml:"read_only,omitempty" json:"read_only"`
	Username    string `yaml:"username,omitempty" json:"-"`
	Password    string `yaml:"password,omitempty" json:"-"`
}

// PublicServerConfig 是返回给前端的脱敏服务端配置。
type PublicServerConfig struct {
	Port         string `json:"port"`
	StaticDir    string `json:"static_dir"`
	TokenEnabled bool   `json:"token_enabled"`
	ReadOnly     bool   `json:"read_only"`
}

// RouterConfig 描述一台爱快 V4 路由器节点。
type RouterConfig struct {
	ID                 string   `yaml:"id" json:"id"`
	OriginalID         string   `yaml:"-" json:"original_id,omitempty"`
	Name               string   `yaml:"name" json:"name"`
	URL                string   `yaml:"url" json:"url"`
	Token              string   `yaml:"token,omitempty" json:"token,omitempty"`
	Tags               []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	Note               string   `yaml:"note,omitempty" json:"note,omitempty"`
	Mock               bool     `yaml:"mock" json:"mock"`
	InsecureSkipVerify bool     `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
	Source             string   `yaml:"-" json:"source,omitempty"`
}

// PublicRouterConfig 是脱敏后的节点配置（不含 Token 原文）。
type PublicRouterConfig struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	URL                string   `json:"url"`
	HasToken           bool     `json:"has_token"`
	Tags               []string `json:"tags"`
	Note               string   `json:"note"`
	Mock               bool     `json:"mock"`
	InsecureSkipVerify bool     `json:"insecure_skip_verify"`
	Source             string   `json:"source"`
}

// AppConfig 是后端完整 YAML 配置。
type AppConfig struct {
	Server         ServerConfig   `yaml:"server" json:"server"`
	ActiveRouterID string         `yaml:"active_router_id" json:"active_router_id"`
	Routers        []RouterConfig `yaml:"routers" json:"routers"`
}

// PublicAppConfig 是返回给前端的脱敏配置。
type PublicAppConfig struct {
	Server         PublicServerConfig   `json:"server"`
	ActiveRouterID string               `json:"active_router_id"`
	Routers        []PublicRouterConfig `json:"routers"`
}

var (
	GlobalConfig *AppConfig
	GlobalPath   string
	mu           sync.RWMutex
)

// InitConfig 初始化全局配置。
func InitConfig(path string) error {
	cfg, resolvedPath, err := LoadRuntimeConfig(path)
	if err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()
	GlobalConfig = cfg
	GlobalPath = resolvedPath
	return nil
}

// LoadRuntimeConfig 读取运行配置；支持 MOCK_MODE=true 和环境变量单节点注入。
func LoadRuntimeConfig(path string) (*AppConfig, string, error) {
	resolvedPath := ResolvePath(path)
	cfg, err := LoadFromFile(resolvedPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && resolvedPath == DefaultConfigPath {
			cfg = DefaultConfig()
			cfg.applyDefaults()
		} else {
			return nil, resolvedPath, err
		}
	}

	if envPort := strings.TrimSpace(os.Getenv("PORT")); envPort != "" {
		cfg.Server.Port = envPort
	}

	switch strings.ToLower(strings.TrimSpace(os.Getenv("READ_ONLY"))) {
	case "true", "1", "yes":
		cfg.Server.ReadOnly = true
	case "false", "0", "no":
		cfg.Server.ReadOnly = false
	}

	mockEnv := strings.EqualFold(strings.TrimSpace(os.Getenv("MOCK_MODE")), "true")
	if mockEnv {
		if len(cfg.Routers) == 0 {
			cfg.Routers = defaultMockRouters()
			cfg.ActiveRouterID = cfg.Routers[0].ID
		} else {
			for i := range cfg.Routers {
				cfg.Routers[i].Mock = true
				if cfg.Routers[i].Source == "" {
					cfg.Routers[i].Source = "mock"
				}
			}
		}
	} else if len(cfg.Routers) == 0 {
		if envRouter, ok := routerFromEnv(); ok {
			cfg.Routers = []RouterConfig{envRouter}
			cfg.ActiveRouterID = envRouter.ID
		}
	}

	cfg.applyDefaults()
	return cfg, resolvedPath, nil
}

func defaultMockRouters() []RouterConfig {
	return []RouterConfig{
		{
			ID:                 "hq-v4",
			Name:               "总部核心网关 · x86",
			URL:                "https://192.168.9.1",
			Tags:               []string{"主路由", "双线", "x86"},
			Note:               "核心出口节点 · 模拟高保真遥测",
			Mock:               true,
			InsecureSkipVerify: true,
			Source:             "mock",
		},
		{
			ID:                 "lab-v4",
			Name:               "研发边缘路由 · ARM",
			URL:                "https://10.10.0.1",
			Tags:               []string{"实验室", "WireGuard", "ARM"},
			Note:               "分支机构互联节点 · 模拟高保真遥测",
			Mock:               true,
			InsecureSkipVerify: true,
			Source:             "mock",
		},
	}
}

func routerFromEnv() (RouterConfig, bool) {
	baseURL := firstNonEmpty(os.Getenv("IKUAI_CLI_BASE_URL"), os.Getenv("IKUAI_URL"))
	token := firstNonEmpty(os.Getenv("IKUAI_CLI_TOKEN"), os.Getenv("IKUAI_TOKEN"))
	baseURL = ikuaiv4.NormalizeBaseURL(baseURL)
	token = strings.TrimSpace(token)
	if baseURL == "" || token == "" {
		return RouterConfig{}, false
	}
	name := strings.TrimSpace(os.Getenv("IKUAI_NAME"))
	if name == "" {
		name = "环境变量节点 (V4)"
	}
	return RouterConfig{
		ID:                 "env-v4",
		Name:               name,
		URL:                baseURL,
		Token:              token,
		Tags:               []string{"env", "v4"},
		Note:               "通过进程环境变量加载的节点（未保存前仅驻留内存）",
		Mock:               false,
		InsecureSkipVerify: true,
		Source:             "env",
	}, true
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// ResolvePath 确定配置文件路径。
func ResolvePath(path string) string {
	if path != "" {
		return path
	}
	if envPath := os.Getenv("IKUAI_CONFIG"); envPath != "" {
		return envPath
	}
	return DefaultConfigPath
}

// Snapshot 返回当前配置副本。
func Snapshot() *AppConfig {
	mu.RLock()
	defer mu.RUnlock()
	if GlobalConfig == nil {
		return nil
	}
	return GlobalConfig.Clone()
}

// IsReadOnly 返回当前是否处于只读保护模式（支持 READ_ONLY 环境变量覆盖）。
func IsReadOnly() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("READ_ONLY"))) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	}
	mu.RLock()
	defer mu.RUnlock()
	if GlobalConfig != nil {
		return GlobalConfig.Server.ReadOnly
	}
	return false
}

// ReplaceGlobal 替换全局配置。
func ReplaceGlobal(cfg *AppConfig) {
	mu.Lock()
	defer mu.Unlock()
	GlobalConfig = cfg.Clone()
}

// LoadFromFile 从文件读取配置。
func LoadFromFile(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Load(data)
}

// Load 从 YAML 字节读取配置，并启用未知字段检查。
func Load(data []byte) (*AppConfig, error) {
	var cfg AppConfig
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("解析 YAML 配置失败: %w", err)
	}
	for i := range cfg.Routers {
		cfg.Routers[i].Source = "yaml"
	}
	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SaveToFile 校验并原子写入 YAML 配置文件。
func SaveToFile(path string, cfg *AppConfig) error {
	next := cfg.Clone()
	next.applyDefaults()
	if err := next.Validate(); err != nil {
		return err
	}

	data, err := yaml.Marshal(next)
	if err != nil {
		return fmt.Errorf("序列化 YAML 配置失败: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".config-*.yaml")
	if err != nil {
		return fmt.Errorf("创建临时配置文件失败: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("写入临时配置失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("关闭临时配置失败: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("替换配置文件失败: %w", err)
	}
	return nil
}

// DefaultConfig 返回空节点首配配置。
func DefaultConfig() *AppConfig {
	return &AppConfig{
		Server:  ServerConfig{Port: defaultPort, StaticDir: defaultStaticDir},
		Routers: []RouterConfig{},
	}
}

// Validate 校验纯 V4 配置一致性。
func (c *AppConfig) Validate() error {
	if c == nil {
		return errors.New("配置不能为空")
	}
	if strings.TrimSpace(c.Server.Port) == "" {
		return errors.New("server.port 不能为空")
	}
	if len(c.Routers) == 0 {
		if strings.TrimSpace(c.ActiveRouterID) != "" {
			return errors.New("active_router_id 未配置路由器时必须为空")
		}
		return nil
	}

	seen := make(map[string]struct{}, len(c.Routers))
	hasActive := false
	for i, router := range c.Routers {
		prefix := fmt.Sprintf("routers[%d]", i)
		id := strings.TrimSpace(router.ID)
		if id == "" {
			return fmt.Errorf("%s.id 不能为空", prefix)
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("routers.id %q 重复", id)
		}
		seen[id] = struct{}{}
		if strings.TrimSpace(router.Name) == "" {
			return fmt.Errorf("%s.name 不能为空", prefix)
		}
		if strings.TrimSpace(router.URL) == "" {
			return fmt.Errorf("%s.url 不能为空", prefix)
		}
		if !router.Mock && strings.TrimSpace(router.Token) == "" {
			return fmt.Errorf("%s.token 不能为空（爱快 V4 连接需要 API Token，除非 mock 为 true）", prefix)
		}
		if id == c.ActiveRouterID {
			hasActive = true
		}
	}
	if strings.TrimSpace(c.ActiveRouterID) == "" {
		return errors.New("active_router_id 不能为空")
	}
	if !hasActive {
		return fmt.Errorf("active_router_id %q 未在 routers 中定义", c.ActiveRouterID)
	}
	return nil
}

// ActiveRouter 返回当前激活的爱快 V4 节点。
func (c *AppConfig) ActiveRouter() (RouterConfig, bool) {
	if c == nil {
		return RouterConfig{}, false
	}
	for _, router := range c.Routers {
		if router.ID == c.ActiveRouterID {
			return router, true
		}
	}
	return RouterConfig{}, false
}

// FindRouter 按 ID 查找节点。
func (c *AppConfig) FindRouter(id string) (RouterConfig, bool) {
	if c == nil {
		return RouterConfig{}, false
	}
	for _, router := range c.Routers {
		if router.ID == id {
			return router, true
		}
	}
	return RouterConfig{}, false
}

// FindRouterByIDOrName 按节点 ID 或名称查找节点（先精确匹配，再忽略大小写匹配）。
func (c *AppConfig) FindRouterByIDOrName(key string) (RouterConfig, bool) {
	if c == nil {
		return RouterConfig{}, false
	}
	needle := strings.TrimSpace(key)
	if needle == "" {
		return RouterConfig{}, false
	}
	for _, router := range c.Routers {
		if router.ID == needle || router.Name == needle {
			return router, true
		}
	}
	for _, router := range c.Routers {
		if strings.EqualFold(router.ID, needle) || strings.EqualFold(router.Name, needle) {
			return router, true
		}
	}
	return RouterConfig{}, false
}

// Clone 深拷贝配置。
func (c *AppConfig) Clone() *AppConfig {
	if c == nil {
		return nil
	}
	next := *c
	next.Routers = make([]RouterConfig, len(c.Routers))
	for i, r := range c.Routers {
		next.Routers[i] = r
		if r.Tags != nil {
			next.Routers[i].Tags = append([]string(nil), r.Tags...)
		}
	}
	return &next
}

// Public 返回脱敏后的前端配置（绝不返回 Token 原文）。
func (c *AppConfig) Public() PublicAppConfig {
	if c == nil {
		return PublicAppConfig{
			Routers: []PublicRouterConfig{},
		}
	}
	routers := make([]PublicRouterConfig, 0, len(c.Routers))
	for _, r := range c.Routers {
		tags := []string{}
		if len(r.Tags) > 0 {
			tags = append([]string(nil), r.Tags...)
		}
		source := r.Source
		if source == "" {
			if r.Mock {
				source = "mock"
			} else {
				source = "yaml"
			}
		}
		routers = append(routers, PublicRouterConfig{
			ID:                 r.ID,
			Name:               r.Name,
			URL:                r.URL,
			HasToken:           strings.TrimSpace(r.Token) != "",
			Tags:               tags,
			Note:               r.Note,
			Mock:               r.Mock,
			InsecureSkipVerify: r.InsecureSkipVerify,
			Source:             source,
		})
	}
	return PublicAppConfig{
		Server: PublicServerConfig{
			Port:         c.Server.Port,
			StaticDir:    c.Server.StaticDir,
			TokenEnabled: strings.TrimSpace(c.Server.AccessToken) != "",
			ReadOnly:     IsReadOnly() || c.Server.ReadOnly,
		},
		ActiveRouterID: c.ActiveRouterID,
		Routers:        routers,
	}
}

// MergeRouterSecrets 在前端提交空 Token 时保留已有令牌，并保留未传的服务端配置与凭据。
func (c *AppConfig) MergeRouterSecrets(existing *AppConfig) {
	if c == nil || existing == nil {
		return
	}
	if strings.TrimSpace(c.Server.Port) == "" {
		c.Server.Port = existing.Server.Port
	}
	if strings.TrimSpace(c.Server.StaticDir) == "" {
		c.Server.StaticDir = existing.Server.StaticDir
	}
	if strings.TrimSpace(c.Server.AccessToken) == "" {
		c.Server.AccessToken = existing.Server.AccessToken
	}
	if strings.TrimSpace(c.Server.Username) == "" {
		c.Server.Username = existing.Server.Username
	}
	if strings.TrimSpace(c.Server.Password) == "" {
		c.Server.Password = existing.Server.Password
	}
	tokens := make(map[string]string, len(existing.Routers))
	for _, router := range existing.Routers {
		tokens[router.ID] = router.Token
	}
	for i := range c.Routers {
		if strings.TrimSpace(c.Routers[i].Token) != "" {
			continue
		}
		if tok := tokens[strings.TrimSpace(c.Routers[i].ID)]; tok != "" {
			c.Routers[i].Token = tok
		} else if orig := strings.TrimSpace(c.Routers[i].OriginalID); orig != "" {
			c.Routers[i].Token = tokens[orig]
		}
	}
}

func (c *AppConfig) applyDefaults() {
	if strings.TrimSpace(c.Server.Port) == "" {
		c.Server.Port = defaultPort
	}
	if strings.TrimSpace(c.Server.StaticDir) == "" {
		c.Server.StaticDir = defaultStaticDir
	}
	for i := range c.Routers {
		c.Routers[i].ID = strings.TrimSpace(c.Routers[i].ID)
		c.Routers[i].Name = strings.TrimSpace(c.Routers[i].Name)
		c.Routers[i].URL = ikuaiv4.NormalizeBaseURL(c.Routers[i].URL)
		c.Routers[i].Token = strings.TrimSpace(c.Routers[i].Token)
		if c.Routers[i].ID == "" {
			if c.Routers[i].Name != "" {
				slug := strings.ToLower(strings.ReplaceAll(c.Routers[i].Name, " ", "-"))
				c.Routers[i].ID = slug
			} else {
				c.Routers[i].ID = fmt.Sprintf("router-%d", i+1)
			}
		}
		if c.Routers[i].Source == "" {
			if c.Routers[i].Mock {
				c.Routers[i].Source = "mock"
			} else {
				c.Routers[i].Source = "yaml"
			}
		}
	}
	if strings.TrimSpace(c.ActiveRouterID) == "" && len(c.Routers) > 0 {
		c.ActiveRouterID = c.Routers[0].ID
	}
}

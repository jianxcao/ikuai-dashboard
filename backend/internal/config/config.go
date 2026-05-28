package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigPath = "config.yaml"
	defaultPort       = "8080"
	defaultStaticDir  = "frontend/dist"
	RouterVersionV3   = "v3"
	RouterVersionV4   = "v4"
)

// ServerConfig 保存后端运行参数。
type ServerConfig struct {
	Port        string `yaml:"port" json:"port"`
	StaticDir   string `yaml:"static_dir" json:"static_dir"`
	AccessToken string `yaml:"access_token" json:"access_token,omitempty"`
}

// PublicServerConfig 是返回给前端的脱敏服务端配置（不含 AccessToken 内容）。
type PublicServerConfig struct {
	Port         string `json:"port"`
	StaticDir    string `json:"static_dir"`
	TokenEnabled bool   `json:"token_enabled"`
}

// RouterConfig 描述一台爱快服务器。
type RouterConfig struct {
	ID                 string `yaml:"id" json:"id"`
	Name               string `yaml:"name" json:"name"`
	URL                string `yaml:"url" json:"url"`
	Version            string `yaml:"version" json:"version"`
	Username           string `yaml:"username" json:"username"`
	Password           string `yaml:"password" json:"password,omitempty"`
	Token              string `yaml:"token" json:"token,omitempty"`
	Mock               bool   `yaml:"mock" json:"mock"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
}

// AppConfig 是后端完整 YAML 配置。
type AppConfig struct {
	Server         ServerConfig   `yaml:"server" json:"server"`
	ActiveRouterID string         `yaml:"active_router_id" json:"active_router_id"`
	Routers        []RouterConfig `yaml:"routers" json:"routers"`
}

// PublicAppConfig 是返回给前端的脱敏配置。
type PublicAppConfig struct {
	Server         PublicServerConfig `json:"server"`
	ActiveRouterID string             `json:"active_router_id"`
	Routers        []RouterConfig     `json:"routers"`
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

// LoadRuntimeConfig 读取运行配置；缺失默认配置时使用空配置启动首配流程。
func LoadRuntimeConfig(path string) (*AppConfig, string, error) {
	resolvedPath := ResolvePath(path)
	cfg, err := LoadFromFile(resolvedPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && resolvedPath == DefaultConfigPath {
			cfg = DefaultConfig()
			cfg.applyDefaults()
			return cfg, resolvedPath, nil
		}
		return nil, resolvedPath, err
	}
	return cfg, resolvedPath, nil
}

// ResolvePath 确定配置文件路径。环境变量仅用于定位配置文件。
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
		tmp.Close()
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

// DefaultConfig 返回无真实凭据的首配配置。
func DefaultConfig() *AppConfig {
	return &AppConfig{
		Server:  ServerConfig{Port: defaultPort, StaticDir: defaultStaticDir},
		Routers: []RouterConfig{},
	}
}

// Validate 校验配置一致性。
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
		switch router.Version {
		case RouterVersionV3:
			if strings.TrimSpace(router.Username) == "" {
				return fmt.Errorf("%s.username 不能为空", prefix)
			}
			if !router.Mock && strings.TrimSpace(router.Password) == "" {
				return fmt.Errorf("%s.password 不能为空，除非 mock 为 true", prefix)
			}
		case RouterVersionV4:
			if !router.Mock && strings.TrimSpace(router.Token) == "" {
				return fmt.Errorf("%s.token 不能为空，v4 真实连接需要 API Token", prefix)
			}
		default:
			return fmt.Errorf("%s.version 只支持 v3 或 v4", prefix)
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

// ActiveRouter 返回当前激活的爱快服务器配置。
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

// Clone 深拷贝配置。
func (c *AppConfig) Clone() *AppConfig {
	if c == nil {
		return nil
	}
	next := *c
	next.Routers = append([]RouterConfig(nil), c.Routers...)
	return &next
}

// Public 返回脱敏后的前端配置。
func (c *AppConfig) Public() PublicAppConfig {
	if c == nil {
		return PublicAppConfig{}
	}
	routers := append([]RouterConfig{}, c.Routers...)
	for i := range routers {
		routers[i].Password = ""
		routers[i].Token = ""
	}
	return PublicAppConfig{
		Server: PublicServerConfig{
			Port:         c.Server.Port,
			StaticDir:    c.Server.StaticDir,
			TokenEnabled: strings.TrimSpace(c.Server.AccessToken) != "",
		},
		ActiveRouterID: c.ActiveRouterID,
		Routers:        routers,
	}
}

// MergeRouterSecrets 在前端提交空密码或空 Token 时保留已有密钥。
func (c *AppConfig) MergeRouterSecrets(existing *AppConfig) {
	if c == nil || existing == nil {
		return
	}
	passwords := make(map[string]string, len(existing.Routers))
	tokens := make(map[string]string, len(existing.Routers))
	for _, router := range existing.Routers {
		passwords[router.ID] = router.Password
		tokens[router.ID] = router.Token
	}
	for i := range c.Routers {
		if c.Routers[i].Password == "" {
			c.Routers[i].Password = passwords[c.Routers[i].ID]
		}
		if c.Routers[i].Token == "" {
			c.Routers[i].Token = tokens[c.Routers[i].ID]
		}
	}
}

func (c *AppConfig) applyDefaults() {
	if c.Server.Port == "" {
		c.Server.Port = defaultPort
	}
	if c.Server.StaticDir == "" {
		c.Server.StaticDir = defaultStaticDir
	}
	for i := range c.Routers {
		if c.Routers[i].ID == "" && c.Routers[i].Name != "" {
			c.Routers[i].ID = strings.ToLower(strings.ReplaceAll(c.Routers[i].Name, " ", "-"))
		}
		if strings.TrimSpace(c.Routers[i].Version) == "" {
			c.Routers[i].Version = RouterVersionV3
		}
	}
	if c.ActiveRouterID == "" && len(c.Routers) > 0 {
		c.ActiveRouterID = c.Routers[0].ID
	}
}

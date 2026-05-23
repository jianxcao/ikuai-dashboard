package config

import "os"

// Config 后端核心配置
type Config struct {
	Port      string // Go 后端服务监听端口，默认 8080
	MockMode  bool   // 是否开启高保真 Mock 模拟模式，若无法连接真实爱快则自动开启
	IKuaiURL  string // 爱快路由器 Web 地址，如 http://192.168.1.1
	Username  string // 爱快登录账号
	Password  string // 爱快登录密码
	APIToken  string // 爱快 V4 API Token (若提供则优先使用 Token 鉴权)
}

// GlobalConfig 全局配置单例
var GlobalConfig *Config

// InitConfig 初始化全局配置
func InitConfig() {
	GlobalConfig = &Config{
		Port:     getEnv("PORT", "8080"),
		MockMode: getEnv("MOCK_MODE", "true") == "true", // 默认开启高保真仿真模式
		IKuaiURL: getEnv("IKUAI_URL", "http://192.168.9.1"),
		Username: getEnv("IKUAI_USERNAME", "admin"),
		Password: getEnv("IKUAI_PASSWORD", ""),
		APIToken: getEnv("IKUAI_API_TOKEN", ""),
	}

	// 如果配置了真实的 URL 且提供了密码或 Token，我们将尝试在服务启动时验证爱快连接
	if GlobalConfig.IKuaiURL != "" && (GlobalConfig.Password != "" || GlobalConfig.APIToken != "") {
		GlobalConfig.MockMode = false
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

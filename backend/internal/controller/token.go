package controller

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"

	"ikuai-dashboard/backend/internal/config"
	"ikuai-dashboard/backend/internal/logger"

	"github.com/gin-gonic/gin"
)

type tokenStatusResponse struct {
	TokenEnabled bool `json:"token_enabled"`
}

type saveTokenRequest struct {
	Token    string `json:"token"`
	Generate bool   `json:"generate"`
}

type saveTokenResponse struct {
	Token        string `json:"token"`
	TokenEnabled bool   `json:"token_enabled"`
}

// GetTokenStatusHandler GET /api/v1/config/token
// 返回 Dashboard 访问 Token 的启用状态（不含 Token 实际值）。
// 此端点不受 Token 保护，供前端判断是否需要认证。
func GetTokenStatusHandler(c *gin.Context) {
	cfg := config.Snapshot()
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "配置未初始化"})
		return
	}
	enabled := strings.TrimSpace(cfg.Server.AccessToken) != ""
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": tokenStatusResponse{TokenEnabled: enabled},
	})
}

// SaveTokenHandler PUT /api/v1/config/token
// 保存或生成 Dashboard 访问 Token。
//   - 传入 {"token": "xxx"} 使用指定 Token
//   - 传入 {"generate": true} 自动生成随机 Token
//   - 传入 {"token": ""} 清除 Token（关闭保护）
func SaveTokenHandler(c *gin.Context) {
	var req saveTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	token := strings.TrimSpace(req.Token)
	if req.Generate {
		var err error
		token, err = generateToken()
		if err != nil {
			logger.Log.Errorf("[token] 生成随机 Token 失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成 Token 失败"})
			return
		}
	}

	cfg := config.Snapshot()
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "配置未初始化"})
		return
	}

	next := cfg.Clone()
	next.Server.AccessToken = token
	if err := config.SaveToFile(config.GlobalPath, next); err != nil {
		logger.Log.Errorf("[token] 保存 Token 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	config.ReplaceGlobal(next)
	logger.Log.Infof("[token] Dashboard 访问 Token 已%s", map[bool]string{true: "生成", false: func() string {
		if token == "" {
			return "清除"
		}
		return "更新"
	}()}[req.Generate])

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": saveTokenResponse{
			Token:        token,
			TokenEnabled: token != "",
		},
	})
}

// generateToken 使用 crypto/rand 生成 32 字节的 URL 安全 Base64 Token。
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

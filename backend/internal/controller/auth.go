package controller

import (
	"net/http"

	"ikuai-dashboard/backend/internal/auth"
	"ikuai-dashboard/backend/internal/logger"

	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
}

type authStatusResponse struct {
	AuthEnabled bool `json:"auth_enabled"`
}

// GetAuthStatusHandler GET /api/v1/auth/status
// 返回是否启用了登录认证，前端据此决定是否显示登录页。
// 此端点始终公开，不受任何中间件保护。
func GetAuthStatusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": authStatusResponse{AuthEnabled: auth.IsAuthEnabled()},
	})
}

// LoginHandler POST /api/v1/auth/login
// 验证用户名密码，成功后返回 Session Token。
func LoginHandler(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供用户名和密码"})
		return
	}

	if !auth.ValidateCredentials(req.Username, req.Password) {
		logger.Log.Warnf("[auth] 登录失败：用户 %q 密码错误，来源 IP: %s", req.Username, c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户名或密码错误"})
		return
	}

	token := auth.GenerateSessionToken(req.Username)
	logger.Log.Infof("[auth] 用户 %q 登录成功，来源 IP: %s", req.Username, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": loginResponse{
			Token:     token,
			ExpiresIn: auth.SessionTTLSeconds(),
		},
	})
}

// LogoutHandler POST /api/v1/auth/logout
// Session Token 无状态，客户端清除 localStorage 即完成登出。
// 此端点仅供前端规范调用，后端记录日志并返回 200。
func LogoutHandler(c *gin.Context) {
	logger.Log.Infof("[auth] 用户登出，来源 IP: %s", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已登出"})
}

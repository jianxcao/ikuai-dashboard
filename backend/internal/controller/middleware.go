package controller

import (
	"net/http"
	"strings"

	"ikuai-dashboard/backend/internal/auth"
	"ikuai-dashboard/backend/internal/config"

	"github.com/gin-gonic/gin"
)

// TokenAuthMiddleware 验证请求认证。
//
// 验证优先级：
//  1. DISABLE_AUTH=true → 直接放行
//  2. Authorization: Bearer <token>
//     a. 先验证是否为有效 Session Token（用户名密码登录颁发）
//     b. 再验证是否为 config.yaml 的 access_token（API 直接调用）
//  3. ?token=<value> → 仅与 config.yaml access_token 比较
//  4. 都不匹配 → 401
func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// DISABLE_AUTH=true：跳过所有认证
		if !auth.IsAuthEnabled() {
			c.Next()
			return
		}

		cfg := config.Snapshot()
		staticToken := ""
		if cfg != nil {
			staticToken = strings.TrimSpace(cfg.Server.AccessToken)
		}

		// 从 Authorization: Bearer <token> 中提取
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			bearer := strings.TrimPrefix(authHeader, "Bearer ")

			// 优先验证 Session Token（用户登录颁发）
			if auth.ValidateSessionToken(bearer) {
				c.Next()
				return
			}

			// 兼容 config.yaml access_token（API 直接调用场景）
			if staticToken != "" && bearer == staticToken {
				c.Next()
				return
			}
		}

		// 从 ?token=<value> 查询参数中提取（仅 access_token，不用于 Session）
		if queryToken := c.Query("token"); staticToken != "" && queryToken == staticToken {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权：请先登录或提供有效的访问 Token",
		})
	}
}

// ReadOnlyMiddleware 拦截服务端只读保护模式 (READ_ONLY=true) 下的写入操作。
// 放行只读查询、节点连接测试与 Dry-Run 请求预览。
func ReadOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.IsReadOnly() {
			c.Next()
			return
		}
		method := c.Request.Method
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			c.Next()
			return
		}
		if c.Request.URL.Path == "/api/v1/routers/test" {
			c.Next()
			return
		}
		if strings.EqualFold(c.Query("dry_run"), "true") || c.Query("dry_run") == "1" {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "当前处于服务端只读保护模式 (READ_ONLY=true)，禁止执行写入操作；仍可使用 Dry-Run 预览",
		})
	}
}

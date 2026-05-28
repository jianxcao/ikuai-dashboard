package controller

import (
	"net/http"
	"strings"

	"ikuai-dashboard/backend/internal/config"

	"github.com/gin-gonic/gin"
)

// TokenAuthMiddleware 验证 Dashboard 访问 Token。
// 仅当 config.yaml 中 server.access_token 不为空时才启用认证。
// Token 可通过以下方式传递：
//   - Authorization: Bearer <token>  请求头
//   - ?token=<token>                 查询参数
func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Snapshot()
		if cfg == nil {
			c.Next()
			return
		}

		required := strings.TrimSpace(cfg.Server.AccessToken)
		// Token 未配置时，跳过认证（保持向后兼容）
		if required == "" {
			c.Next()
			return
		}

		// 从 Authorization: Bearer <token> 头中提取
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			if strings.TrimPrefix(authHeader, "Bearer ") == required {
				c.Next()
				return
			}
		}

		// 从 ?token=<token> 查询参数中提取
		if queryToken := c.Query("token"); queryToken == required {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权：请提供有效的访问 Token",
		})
	}
}

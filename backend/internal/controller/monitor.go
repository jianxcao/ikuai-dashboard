package controller

import (
	"net/http"

	"ikuai4-backend/backend/internal/logger"
	"ikuai4-backend/backend/internal/service"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware 处理前后端分离的跨域请求
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// GetInterfaceDataHandler 获取首页流量看板数据
func GetInterfaceDataHandler(c *gin.Context) {
	ctx := c.Request.Context()
	data, err := service.GlobalMonitorService.GetInterfaceData(ctx)
	if err != nil {
		logger.Log.Errorf("Controller 获取首页流量数据发生异常: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取数据发生异常",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": data,
	})
}

// GetLanClientsHandler 获取聚合去重后的 LAN 客户端数据列表
func GetLanClientsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search") // 备注模糊过滤

	clients, err := service.GlobalMonitorService.GetLanClients(ctx, search)
	if err != nil {
		logger.Log.Errorf("Controller 获取局域网在线终端发生异常: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "拉取局域网终端发生异常",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": clients,
	})
}

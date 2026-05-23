package controller

import (
	"errors"
	"net/http"

	"ikuai-dashboard/backend/internal/logger"
	"ikuai-dashboard/backend/internal/service"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware 处理跨域请求（前后端分离必需）
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// GetInterfaceDataHandler GET /api/v1/monitor/interface
// 返回首页看板所需数据：汇总卡片 + WAN 状态 + 接口流量明细
func GetInterfaceDataHandler(c *gin.Context) {
	data, err := service.CurrentMonitorService().GetInterfaceData(c.Request.Context())
	if err != nil {
		if errors.Is(err, service.ErrUnconfigured) {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": err.Error()})
			return
		}
		logger.Log.Errorf("[interface] 获取失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// GetLanClientsHandler GET /api/v1/monitor/lan?search=关键词
// 返回按 MAC 聚合去重的局域网客户端列表，支持按备注 / hostname 搜索
func GetLanClientsHandler(c *gin.Context) {
	search := c.Query("search")
	clients, err := service.CurrentMonitorService().GetLanClients(c.Request.Context(), search)
	if err != nil {
		if errors.Is(err, service.ErrUnconfigured) {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": err.Error()})
			return
		}
		logger.Log.Errorf("[lan] 获取失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": clients})
}

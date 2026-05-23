package controller

import (
	"net/http"

	"ikuai-dashboard/backend/internal/config"
	"ikuai-dashboard/backend/internal/logger"
	"ikuai-dashboard/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type activeRouterRequest struct {
	ID string `json:"id" binding:"required"`
}

type routersConfigRequest struct {
	Server         config.ServerConfig   `json:"server"`
	ActiveRouterID string                `json:"active_router_id"`
	Routers        []config.RouterConfig `json:"routers"`
}

// GetRoutersConfigHandler GET /api/v1/config/routers
func GetRoutersConfigHandler(c *gin.Context) {
	cfg := config.Snapshot()
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "配置未初始化"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   200,
		"data":   cfg.Public(),
		"status": service.CurrentMonitorService().Status(),
	})
}

// SwitchActiveRouterHandler PUT /api/v1/config/active-router
func SwitchActiveRouterHandler(c *gin.Context) {
	var req activeRouterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	cfg := config.Snapshot()
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "配置未初始化"})
		return
	}

	next := cfg.Clone()
	next.ActiveRouterID = req.ID
	if err := next.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	if err := config.SaveToFile(config.GlobalPath, next); err != nil {
		logger.Log.Errorf("[config] 保存激活路由器失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	config.ReplaceGlobal(next)
	service.ReloadMonitorService()

	c.JSON(http.StatusOK, gin.H{
		"code":   200,
		"data":   next.Public(),
		"status": service.CurrentMonitorService().Status(),
	})
}

// SaveRoutersConfigHandler PUT /api/v1/config/routers
func SaveRoutersConfigHandler(c *gin.Context) {
	var req routersConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	current := config.Snapshot()
	next := &config.AppConfig{
		Server:         req.Server,
		ActiveRouterID: req.ActiveRouterID,
		Routers:        req.Routers,
	}
	next.MergeRouterSecrets(current)
	if err := next.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	if err := config.SaveToFile(config.GlobalPath, next); err != nil {
		logger.Log.Errorf("[config] 保存路由器配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	config.ReplaceGlobal(next)
	service.ReloadMonitorService()

	c.JSON(http.StatusOK, gin.H{
		"code":   200,
		"data":   next.Public(),
		"status": service.CurrentMonitorService().Status(),
	})
}

package controller

import (
	"net/http"
	"strings"

	"ikuai-dashboard/backend/internal/config"
	"ikuai-dashboard/backend/internal/fleet"
	"ikuai-dashboard/backend/internal/logger"

	"github.com/gin-gonic/gin"
)

type activeRouterRequest struct {
	ID string `json:"id" binding:"required"`
}

type routersConfigRequest struct {
	Server         *config.ServerConfig  `json:"server,omitempty"`
	ActiveRouterID string                `json:"active_router_id"`
	Routers        []config.RouterConfig `json:"routers"`
}

// GetRoutersConfigHandler GET /api/v1/routers (and /api/v1/config/routers)
func GetRoutersConfigHandler(c *gin.Context) {
	cfg := config.Snapshot()
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "配置未初始化"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   200,
		"data":   cfg.Public(),
		"status": fleet.GlobalManager.Status(),
	})
}

// ProbeRoutersHandler GET /api/v1/routers/probe
func ProbeRoutersHandler(c *gin.Context) {
	items := fleet.GlobalManager.ProbeAll(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": items,
	})
}

// TestRouterHandler POST /api/v1/routers/test
func TestRouterHandler(c *gin.Context) {
	var req fleet.RouterTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	res, err := fleet.GlobalManager.TestRouter(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": res})
}

// SwitchActiveRouterHandler PUT /api/v1/routers/active (and /api/v1/config/active-router)
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

	// Only persist to YAML if the node list is YAML-managed (do not write mock/env ephemeral tokens to disk).
	if !hasEphemeralOnlyRouters(next.Routers) {
		if err := config.SaveToFile(config.GlobalPath, next); err != nil {
			logger.Log.Errorf("[config] 保存激活路由器失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
	}

	config.ReplaceGlobal(next)
	c.JSON(http.StatusOK, gin.H{
		"code":   200,
		"data":   next.Public(),
		"status": fleet.GlobalManager.Status(),
	})
}

// SaveRoutersConfigHandler PUT /api/v1/routers (and /api/v1/config/routers)
func SaveRoutersConfigHandler(c *gin.Context) {
	var req routersConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	current := config.Snapshot()
	routers := make([]config.RouterConfig, len(req.Routers))
	copy(routers, req.Routers)
	for i := range routers {
		routers[i].ID = strings.TrimSpace(routers[i].ID)
		routers[i].OriginalID = strings.TrimSpace(routers[i].OriginalID)
		routers[i].Source = "yaml"
	}

	activeID := strings.TrimSpace(req.ActiveRouterID)
	if len(routers) > 0 {
		hasActive := false
		for _, r := range routers {
			if r.ID != "" && r.ID == activeID {
				hasActive = true
				break
			}
		}
		if !hasActive {
			if activeID == "" {
				activeID = routers[0].ID
			} else {
				for _, r := range routers {
					if r.OriginalID != "" && r.OriginalID == activeID && r.ID != "" {
						activeID = r.ID
						break
					}
				}
			}
		}
	}

	var nextServer config.ServerConfig
	if current != nil {
		nextServer = current.Server
	}
	if req.Server != nil {
		nextServer = *req.Server
	}

	next := &config.AppConfig{
		Server:         nextServer,
		ActiveRouterID: activeID,
		Routers:        routers,
	}
	next.MergeRouterSecrets(current)
	for i := range next.Routers {
		next.Routers[i].OriginalID = ""
	}
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
	fleet.GlobalManager.InvalidateCache()

	c.JSON(http.StatusOK, gin.H{
		"code":   200,
		"data":   next.Public(),
		"status": fleet.GlobalManager.Status(),
	})
}

func hasEphemeralOnlyRouters(routers []config.RouterConfig) bool {
	if len(routers) == 0 {
		return false
	}
	for _, r := range routers {
		if r.Source == "env" || r.Source == "mock" {
			return true
		}
	}
	return false
}

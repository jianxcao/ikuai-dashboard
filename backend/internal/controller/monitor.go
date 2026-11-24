package controller

import (
	"errors"
	"net/http"

	"ikuai-dashboard/backend/internal/config"
	"ikuai-dashboard/backend/internal/fleet"
	"ikuai-dashboard/backend/internal/ikuaiv4"
	"ikuai-dashboard/backend/internal/logger"
	"ikuai-dashboard/backend/internal/v4service"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware 处理跨域请求。
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Router-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func resolveRouterForRequest(c *gin.Context) (config.RouterConfig, *ikuaiv4.Client, error) {
	overrideID := c.Query("router_id")
	if overrideID == "" {
		overrideID = c.GetHeader("X-Router-ID")
	}
	return fleet.GlobalManager.ResolveRouter(overrideID)
}

func respondV4Error(c *gin.Context, tag string, err error) {
	if errors.Is(err, fleet.ErrUnconfigured) {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": err.Error()})
		return
	}
	if errors.Is(err, fleet.ErrRouterNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": err.Error()})
		return
	}
	var apiErr *ikuaiv4.APIError
	if errors.As(err, &apiErr) {
		status := http.StatusBadGateway
		if apiErr.HTTPStatus == http.StatusUnauthorized || apiErr.Code == 3007 || apiErr.Code == 1008 {
			status = http.StatusBadGateway
		} else if apiErr.Code == 3001 {
			status = http.StatusBadRequest
		}
		logger.Log.Warnf("[%s] 上游 v4 接口返回错误: %v", tag, err)
		c.JSON(status, gin.H{"code": status, "message": err.Error(), "upstream_code": apiErr.Code})
		return
	}
	logger.Log.Errorf("[%s] 请求失败: %v", tag, err)
	c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
}

// GetInterfaceDataHandler GET /api/v1/monitor/interface
func GetInterfaceDataHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "interface", err)
		return
	}
	ov, err := v4service.GetOverview(c.Request.Context(), router, client)
	if err != nil {
		respondV4Error(c, "interface", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": v4service.ToLegacyInterfaceData(ov)})
}

// GetLanClientsHandler GET /api/v1/monitor/lan?search=
func GetLanClientsHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "lan", err)
		return
	}
	clients, err := v4service.GetMergedLanClients(c.Request.Context(), router, client, c.Query("search"))
	if err != nil {
		respondV4Error(c, "lan", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": clients})
}

// GetNetworkMapHandler GET /api/v1/monitor/network-map
func GetNetworkMapHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "network-map", err)
		return
	}
	data, err := v4service.GetNetworkMap(c.Request.Context(), router, client)
	if err != nil {
		respondV4Error(c, "network-map", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// GetSecurityHubHandler GET /api/v1/monitor/security-hub
func GetSecurityHubHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "security-hub", err)
		return
	}
	data, err := v4service.GetSecurityHub(c.Request.Context(), router, client)
	if err != nil {
		respondV4Error(c, "security-hub", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// GetMultiWanHandler GET /api/v1/monitor/multi-wan
func GetMultiWanHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "multi-wan", err)
		return
	}
	data, err := v4service.GetMultiWan(c.Request.Context(), router, client)
	if err != nil {
		respondV4Error(c, "multi-wan", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

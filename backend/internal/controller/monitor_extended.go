package controller

import (
	"net/http"

	"ikuai-dashboard/backend/internal/logger"
	"ikuai-dashboard/backend/internal/service"

	"github.com/gin-gonic/gin"
)

func GetNetworkMapHandler(c *gin.Context) {
	data, err := service.CurrentMonitorService().GetNetworkMap(c.Request.Context())
	if err != nil {
		logger.Log.Errorf("[network-map] 获取失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

func GetSecurityHubHandler(c *gin.Context) {
	data, err := service.CurrentMonitorService().GetSecurityHub(c.Request.Context())
	if err != nil {
		logger.Log.Errorf("[security-hub] 获取失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

func GetMultiWanHandler(c *gin.Context) {
	data, err := service.CurrentMonitorService().GetMultiWan(c.Request.Context())
	if err != nil {
		logger.Log.Errorf("[multi-wan] 获取失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

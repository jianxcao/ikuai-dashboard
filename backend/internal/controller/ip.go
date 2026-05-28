package controller

import (
	"errors"
	"net/http"

	"ikuai-dashboard/backend/internal/logger"
	"ikuai-dashboard/backend/internal/service"

	"github.com/gin-gonic/gin"
)

func GetWanIPHandler(c *gin.Context) {
	name := c.Query("name")
	lan := c.Query("lan")

	data, err := service.GetWanIPByRouterName(c.Request.Context(), name, lan)
	if err != nil {
		if errors.Is(err, service.ErrUnconfigured) {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": err.Error()})
			return
		}
		logger.Log.Errorf("[ip] 获取 WAN IP 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	// c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
	c.String(http.StatusOK, data.IP)
}

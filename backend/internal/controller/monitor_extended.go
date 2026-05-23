package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"ikuai-dashboard/backend/internal/logger"
	"ikuai-dashboard/backend/internal/service"

	"github.com/gin-gonic/gin"
)

func GetNetworkMapHandler(c *gin.Context) {
	data, err := service.CurrentMonitorService().GetNetworkMap(c.Request.Context())
	if err != nil {
		if errors.Is(err, service.ErrUnconfigured) {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": err.Error()})
			return
		}
		logger.Log.Errorf("[network-map] 获取失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

func GetSecurityHubHandler(c *gin.Context) {
	data, err := service.CurrentMonitorService().GetSecurityHub(c.Request.Context())
	if err != nil {
		if errors.Is(err, service.ErrUnconfigured) {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": err.Error()})
			return
		}
		logger.Log.Errorf("[security-hub] 获取失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

func GetMultiWanHandler(c *gin.Context) {
	data, err := service.CurrentMonitorService().GetMultiWan(c.Request.Context())
	if err != nil {
		if errors.Is(err, service.ErrUnconfigured) {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": err.Error()})
			return
		}
		logger.Log.Errorf("[multi-wan] 获取失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

func GetMonitorInsightsHandler(c *gin.Context) {
	data, err := service.CurrentMonitorService().GetMonitorInsights(c.Request.Context())
	if err != nil {
		if errors.Is(err, service.ErrUnconfigured) {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": err.Error()})
			return
		}
		logger.Log.Errorf("[monitor-insights] 获取失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

func ListCommonResourcesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": service.CurrentMonitorService().ListCommonResources()})
}

func GetCommonResourceHandler(c *gin.Context) {
	data, err := service.CurrentMonitorService().GetCommonResource(c.Request.Context(), c.Param("name"))
	if err != nil {
		writeResourceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

func CreateCommonResourceHandler(c *gin.Context) {
	body, err := readResourceBody(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	data, err := service.CurrentMonitorService().MutateCommonResource(c.Request.Context(), c.Param("name"), http.MethodPost, 0, body)
	if err != nil {
		writeResourceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

func UpdateCommonResourceHandler(c *gin.Context) {
	id, err := serviceIDFromParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	body, err := readResourceBody(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	data, err := service.CurrentMonitorService().MutateCommonResource(c.Request.Context(), c.Param("name"), http.MethodPut, id, body)
	if err != nil {
		writeResourceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

func DeleteCommonResourceHandler(c *gin.Context) {
	id, err := serviceIDFromParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	data, err := service.CurrentMonitorService().MutateCommonResource(c.Request.Context(), c.Param("name"), http.MethodDelete, id, nil)
	if err != nil {
		writeResourceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

func readResourceBody(c *gin.Context) (map[string]any, error) {
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, err
	}
	return body, nil
}

func serviceIDFromParam(value string) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || id <= 0 {
		return 0, errors.New("无效的资源 ID")
	}
	return id, nil
}

func writeResourceError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrUnconfigured) {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": err.Error()})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
}

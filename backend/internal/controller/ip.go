package controller

import (
	"errors"
	"net/http"
	"strings"

	"ikuai-dashboard/backend/internal/fleet"
	"ikuai-dashboard/backend/internal/v4service"

	"github.com/gin-gonic/gin"
)

// GetWanIPHandler GET /api/v1/ip
// 支持按节点 ID/名称（?name= / ?router_id= / ?id= / X-Router-ID）与线路名（?wan= / ?interface= / ?iface= / ?lan=）
// 获取指定 WAN 口 IP，默认返回纯文本 IP，传入 ?format=json 返回 JSON 结构。
func GetWanIPHandler(c *gin.Context) {
	routerKey := firstNonEmptyQuery(c, "name", "router_id", "id")
	if routerKey == "" {
		routerKey = strings.TrimSpace(c.GetHeader("X-Router-ID"))
	}
	wanName := firstNonEmptyQuery(c, "wan", "interface", "iface", "lan")

	router, client, err := fleet.GlobalManager.ResolveRouterByKey(routerKey)
	if err != nil {
		respondV4Error(c, "ip", err)
		return
	}

	res, err := v4service.GetWanIP(c.Request.Context(), router, client, wanName)
	if err != nil {
		if errors.Is(err, v4service.ErrWanNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": err.Error()})
			return
		}
		if errors.Is(err, v4service.ErrWanNoIP) {
			c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": err.Error()})
			return
		}
		respondV4Error(c, "ip", err)
		return
	}

	if strings.EqualFold(strings.TrimSpace(c.Query("format")), "json") {
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": res})
		return
	}
	c.String(http.StatusOK, res.IP)
}

func firstNonEmptyQuery(c *gin.Context, keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(c.Query(k)); v != "" {
			return v
		}
	}
	return ""
}

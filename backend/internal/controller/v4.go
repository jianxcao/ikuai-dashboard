package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"ikuai-dashboard/backend/internal/v4catalog"
	"ikuai-dashboard/backend/internal/v4service"

	"github.com/gin-gonic/gin"
)

// GetV4OverviewHandler GET /api/v1/v4/overview
func GetV4OverviewHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "v4-overview", err)
		return
	}
	data, err := v4service.GetOverview(c.Request.Context(), router, client)
	if err != nil {
		respondV4Error(c, "v4-overview", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// GetV4LoadHistoryHandler GET /api/v1/v4/monitor/load
func GetV4LoadHistoryHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "v4-load", err)
		return
	}
	data, err := v4service.GetLoadHistory(
		c.Request.Context(),
		router,
		client,
		c.DefaultQuery("metric", "cpu"),
		c.DefaultQuery("range", "hour"),
		c.DefaultQuery("aggregate", "avg"),
	)
	if err != nil {
		respondV4Error(c, "v4-load", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// GetV4ClientsHandler GET /api/v1/v4/clients
func GetV4ClientsHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "v4-clients", err)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	data, err := v4service.GetClients(
		c.Request.Context(),
		router,
		client,
		c.DefaultQuery("mode", "online"),
		page,
		pageSize,
		c.Query("search"),
	)
	if err != nil {
		respondV4Error(c, "v4-clients", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// GetV4ClientDetailHandler GET /api/v1/v4/clients/detail
func GetV4ClientDetailHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "v4-client-detail", err)
		return
	}
	data, err := v4service.GetClientDetail(c.Request.Context(), router, client, c.Query("ip"), c.Query("mac"))
	if err != nil {
		respondV4Error(c, "v4-client-detail", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// KickV4OnlineUserHandler DELETE /api/v1/v4/users/online/:id
func KickV4OnlineUserHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "v4-kick-user", err)
		return
	}
	dryRun := isDryRunQuery(c)
	res, err := v4service.KickOnlineSession(c.Request.Context(), router, client, c.Param("id"), dryRun)
	if err != nil {
		respondV4Error(c, "v4-kick-user", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": res})
}

// GetV4AnalyticsHandler GET /api/v1/v4/traffic/analytics
func GetV4AnalyticsHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "v4-analytics", err)
		return
	}
	data, err := v4service.GetAppProtocolAnalytics(c.Request.Context(), router, client)
	if err != nil {
		respondV4Error(c, "v4-analytics", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// GetV4CatalogHandler GET /api/v1/v4/catalog
func GetV4CatalogHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": v4catalog.All()})
}

// ListV4ResourceHandler GET /api/v1/v4/resources/:key
func ListV4ResourceHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "v4-resource-list", err)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	extra := map[string]string{}
	if p := strings.TrimSpace(c.Query("path")); p != "" {
		extra["path"] = p
	}
	res, err := v4catalog.ListOrGet(
		c.Request.Context(),
		router,
		client,
		c.Param("key"),
		page,
		pageSize,
		c.Query("search"),
		c.Query("filter"),
		extra,
	)
	if err != nil {
		respondV4Error(c, "v4-resource-list", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": res})
}

// CreateV4ResourceHandler POST /api/v1/v4/resources/:key
func CreateV4ResourceHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "v4-resource-create", err)
		return
	}
	body, err := readJSONMap(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	res, err := v4catalog.Create(c.Request.Context(), router, client, c.Param("key"), body, isDryRunQuery(c))
	if err != nil {
		respondV4Error(c, "v4-resource-create", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": res})
}

// UpdateV4ResourceHandler PUT /api/v1/v4/resources/:key (and /:id)
func UpdateV4ResourceHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "v4-resource-update", err)
		return
	}
	body, err := readJSONMap(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	res, err := v4catalog.Update(c.Request.Context(), router, client, c.Param("key"), c.Param("id"), body, isDryRunQuery(c))
	if err != nil {
		respondV4Error(c, "v4-resource-update", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": res})
}

// ToggleV4ResourceHandler PATCH /api/v1/v4/resources/:key/:id/toggle
func ToggleV4ResourceHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "v4-resource-toggle", err)
		return
	}
	body, err := readJSONMap(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	enabled := "yes"
	if v, ok := body["enabled"]; ok {
		enabled = strings.TrimSpace(toString(v))
	}
	res, err := v4catalog.Toggle(c.Request.Context(), router, client, c.Param("key"), c.Param("id"), enabled, isDryRunQuery(c))
	if err != nil {
		respondV4Error(c, "v4-resource-toggle", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": res})
}

// DeleteV4ResourceHandler DELETE /api/v1/v4/resources/:key (and /:id)
func DeleteV4ResourceHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "v4-resource-delete", err)
		return
	}
	id := c.Param("id")
	if id == "" {
		id = c.Query("srcfile")
	}
	res, err := v4catalog.Delete(c.Request.Context(), router, client, c.Param("key"), id, isDryRunQuery(c))
	if err != nil {
		respondV4Error(c, "v4-resource-delete", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": res})
}

// ExecuteV4ActionHandler POST /api/v1/v4/actions/:action
func ExecuteV4ActionHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "v4-action", err)
		return
	}
	res, err := v4catalog.ExecuteAction(c.Request.Context(), router, client, c.Param("action"), isDryRunQuery(c))
	if err != nil {
		respondV4Error(c, "v4-action", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": res})
}

// GetObjectRefsHandler GET /api/v1/v4/objects/:key/refs?group_name=...
func GetObjectRefsHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "v4-object-refs", err)
		return
	}
	res, err := v4catalog.GetObjectRefs(c.Request.Context(), router, client, c.Param("key"), c.Query("group_name"))
	if err != nil {
		respondV4Error(c, "v4-object-refs", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": res})
}

// GenerateWGKeypairHandler GET /api/v1/v4/wireguard/keypair
func GenerateWGKeypairHandler(c *gin.Context) {
	priv, pub, err := v4catalog.GenerateWGKeypair()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成 WireGuard 密钥对失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"private_key": priv,
			"public_key":  pub,
		},
	})
}

// ListWGPeersHandler GET /api/v1/v4/wireguard/:id/peers
func ListWGPeersHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "v4-wg-peers", err)
		return
	}
	res, err := v4catalog.ListWGPeers(c.Request.Context(), router, client, c.Param("id"))
	if err != nil {
		respondV4Error(c, "v4-wg-peers", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": res})
}

// CreateWGPeerHandler POST /api/v1/v4/wireguard/:id/peers
func CreateWGPeerHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "v4-wg-peer-create", err)
		return
	}
	body, err := readJSONMap(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	res, err := v4catalog.CreateWGPeer(c.Request.Context(), router, client, c.Param("id"), body, isDryRunQuery(c))
	if err != nil {
		respondV4Error(c, "v4-wg-peer-create", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": res})
}

// UpdateWGPeerHandler PUT /api/v1/v4/wireguard/:id/peers/:peer_id
func UpdateWGPeerHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "v4-wg-peer-update", err)
		return
	}
	body, err := readJSONMap(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	res, err := v4catalog.UpdateWGPeer(c.Request.Context(), router, client, c.Param("id"), c.Param("peer_id"), body, isDryRunQuery(c))
	if err != nil {
		respondV4Error(c, "v4-wg-peer-update", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": res})
}

// ToggleWGPeerHandler PATCH /api/v1/v4/wireguard/:id/peers/:peer_id/toggle
func ToggleWGPeerHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "v4-wg-peer-toggle", err)
		return
	}
	body, err := readJSONMap(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	enabled := "yes"
	if v, ok := body["enabled"]; ok {
		enabled = strings.TrimSpace(toString(v))
	}
	res, err := v4catalog.ToggleWGPeer(c.Request.Context(), router, client, c.Param("id"), c.Param("peer_id"), enabled, isDryRunQuery(c))
	if err != nil {
		respondV4Error(c, "v4-wg-peer-toggle", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": res})
}

// DeleteWGPeerHandler DELETE /api/v1/v4/wireguard/:id/peers/:peer_id
func DeleteWGPeerHandler(c *gin.Context) {
	router, client, err := resolveRouterForRequest(c)
	if err != nil {
		respondV4Error(c, "v4-wg-peer-delete", err)
		return
	}
	res, err := v4catalog.DeleteWGPeer(c.Request.Context(), router, client, c.Param("id"), c.Param("peer_id"), isDryRunQuery(c))
	if err != nil {
		respondV4Error(c, "v4-wg-peer-delete", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": res})
}

func isDryRunQuery(c *gin.Context) bool {
	return strings.EqualFold(c.Query("dry_run"), "true") || c.Query("dry_run") == "1"
}

func readJSONMap(c *gin.Context) (map[string]any, error) {
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		return map[string]any{}, nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if m == nil {
		m = map[string]any{}
	}
	return m, nil
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "yes"
		}
		return "no"
	default:
		b, _ := json.Marshal(t)
		return strings.Trim(string(b), `"`)
	}
}

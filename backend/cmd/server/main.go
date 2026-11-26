package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"ikuai-dashboard/backend/internal/config"
	"ikuai-dashboard/backend/internal/controller"
	"ikuai-dashboard/backend/internal/fleet"
	"ikuai-dashboard/backend/internal/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	configPath := flag.String("config", "", "YAML 配置文件路径")
	flag.Parse()

	logger.InitLogger()
	logger.Log.Info("══════════════════════════════════════")
	logger.Log.Info("  iKuai Console V4 · 后端服务启动中")
	logger.Log.Info("══════════════════════════════════════")

	if err := config.InitConfig(*configPath); err != nil {
		logger.Log.Fatalf("加载配置失败: %v", err)
	}
	activeRouter, ok := config.GlobalConfig.ActiveRouter()
	if ok {
		logger.Log.Infof("配置: %s | 端口: %s | 当前爱快 V4: %s (%s) | Mock: %v",
			config.GlobalPath, config.GlobalConfig.Server.Port, activeRouter.Name, activeRouter.Source, activeRouter.Mock)
	} else {
		logger.Log.Infof("配置: %s | 端口: %s | 当前爱快 V4: 未配置",
			config.GlobalPath, config.GlobalConfig.Server.Port)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(controller.CORSMiddleware())

	// Public status & auth endpoints
	r.GET("/api/v1/config/token", controller.GetTokenStatusHandler)
	r.GET("/api/v1/auth/status", controller.GetAuthStatusHandler)
	r.POST("/api/v1/auth/login", controller.LoginHandler)
	r.POST("/api/v1/auth/logout", controller.LogoutHandler)

	// Protected API endpoints
	api := r.Group("/api/v1", controller.TokenAuthMiddleware(), controller.ReadOnlyMiddleware())
	{
		// Multi-router V4 fleet management
		api.GET("/routers", controller.GetRoutersConfigHandler)
		api.GET("/routers/probe", controller.ProbeRoutersHandler)
		api.POST("/routers/test", controller.TestRouterHandler)
		api.PUT("/routers", controller.SaveRoutersConfigHandler)
		api.PUT("/routers/active", controller.SwitchActiveRouterHandler)

		// Config compatibility aliases
		api.GET("/config/routers", controller.GetRoutersConfigHandler)
		api.PUT("/config/routers", controller.SaveRoutersConfigHandler)
		api.PUT("/config/active-router", controller.SwitchActiveRouterHandler)
		api.PUT("/config/token", controller.SaveTokenHandler)

		// Pure V4 Monitoring, Clients, Analytics & CLI Resource Catalog
		v4 := api.Group("/v4")
		{
			v4.GET("/overview", controller.GetV4OverviewHandler)
			v4.GET("/interfaces/options", controller.GetV4InterfaceOptionsHandler)
			v4.GET("/monitor/load", controller.GetV4LoadHistoryHandler)
			v4.GET("/clients", controller.GetV4ClientsHandler)
			v4.GET("/clients/detail", controller.GetV4ClientDetailHandler)
			v4.DELETE("/users/online/:id", controller.KickV4OnlineUserHandler)
			v4.GET("/traffic/analytics", controller.GetV4AnalyticsHandler)

			v4.GET("/catalog", controller.GetV4CatalogHandler)
			v4.GET("/resources/:key", controller.ListV4ResourceHandler)
			v4.POST("/resources/:key", controller.CreateV4ResourceHandler)
			v4.PUT("/resources/:key", controller.UpdateV4ResourceHandler)
			v4.PUT("/resources/:key/:id", controller.UpdateV4ResourceHandler)
			v4.PATCH("/resources/:key/:id/toggle", controller.ToggleV4ResourceHandler)
			v4.DELETE("/resources/:key", controller.DeleteV4ResourceHandler)
			v4.DELETE("/resources/:key/:id", controller.DeleteV4ResourceHandler)
			v4.POST("/actions/:action", controller.ExecuteV4ActionHandler)
			v4.GET("/objects/:key/refs", controller.GetObjectRefsHandler)
			v4.GET("/wireguard/keypair", controller.GenerateWGKeypairHandler)
			v4.GET("/wireguard/:id/peers", controller.ListWGPeersHandler)
			v4.POST("/wireguard/:id/peers", controller.CreateWGPeerHandler)
			v4.PUT("/wireguard/:id/peers/:peer_id", controller.UpdateWGPeerHandler)
			v4.PATCH("/wireguard/:id/peers/:peer_id/toggle", controller.ToggleWGPeerHandler)
			v4.DELETE("/wireguard/:id/peers/:peer_id", controller.DeleteWGPeerHandler)
		}

		// Preserved monitor endpoints
		api.GET("/ip", controller.GetWanIPHandler)
		api.GET("/monitor/interface", controller.GetInterfaceDataHandler)
		api.GET("/monitor/lan", controller.GetLanClientsHandler)
		api.GET("/monitor/network-map", controller.GetNetworkMapHandler)
		api.GET("/monitor/security-hub", controller.GetSecurityHubHandler)
		api.GET("/monitor/multi-wan", controller.GetMultiWanHandler)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": "v4",
			"config":  config.GlobalPath,
			"service": fleet.GlobalManager.Status(),
		})
	})

	registerStaticRoutes(r, config.GlobalConfig.Server.StaticDir)

	addr := fmt.Sprintf(":%s", config.GlobalConfig.Server.Port)
	logger.Log.Infof("V4 控制台服务就绪 → http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		logger.Log.Fatalf("服务启动失败: %v", err)
	}
}

func setStaticCacheHeaders(c *gin.Context, requestPath string) {
	normalized := filepath.ToSlash(requestPath)
	if strings.HasPrefix(normalized, "assets/") {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
		return
	}
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
}

func registerStaticRoutes(r *gin.Engine, staticDir string) {
	indexPath := filepath.Join(staticDir, "index.html")
	staticRoot, _ := filepath.Abs(staticDir)
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "not found"})
			return
		}
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Status(http.StatusNotFound)
			return
		}

		requestPath := filepath.Clean(strings.TrimPrefix(c.Request.URL.Path, "/"))
		if requestPath != "." {
			filePath := filepath.Join(staticDir, requestPath)
			absFilePath, _ := filepath.Abs(filePath)
			if rel, err := filepath.Rel(staticRoot, absFilePath); err == nil && !strings.HasPrefix(rel, "..") {
				if stat, err := os.Stat(filePath); err == nil && !stat.IsDir() {
					setStaticCacheHeaders(c, rel)
					c.File(filePath)
					return
				}
			}
		}
		if _, err := os.Stat(indexPath); err == nil {
			setStaticCacheHeaders(c, "index.html")
			c.File(indexPath)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "frontend assets not found"})
	})
}

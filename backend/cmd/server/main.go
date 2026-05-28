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
	"ikuai-dashboard/backend/internal/logger"
	"ikuai-dashboard/backend/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	configPath := flag.String("config", "", "YAML 配置文件路径")
	flag.Parse()

	// 1. 初始化 Zap 高性能结构化日志
	logger.InitLogger()
	logger.Log.Info("══════════════════════════════════════")
	logger.Log.Info("  iKuai 流量监控系统 · 后端服务启动中")
	logger.Log.Info("══════════════════════════════════════")

	// 2. 加载全局配置
	if err := config.InitConfig(*configPath); err != nil {
		logger.Log.Fatalf("加载配置失败: %v", err)
	}
	activeRouter, ok := config.GlobalConfig.ActiveRouter()
	if ok {
		logger.Log.Infof("配置: %s | 端口: %s | 当前爱快: %s | Mock: %v",
			config.GlobalPath, config.GlobalConfig.Server.Port, activeRouter.Name, activeRouter.Mock)
	} else {
		logger.Log.Infof("配置: %s | 端口: %s | 当前爱快: 未配置",
			config.GlobalPath, config.GlobalConfig.Server.Port)
	}

	// 3. 初始化爱快连接 / Mock 数据源
	service.InitMonitorService()

	// 4. 初始化 Gin 路由引擎
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(controller.CORSMiddleware())

	// 5. 注册 API 路由
	// 公开端点：不受 Token 保护（供前端查询是否需要认证）
	r.GET("/api/v1/config/token", controller.GetTokenStatusHandler)
	// 登录相关公开端点
	r.GET("/api/v1/auth/status", controller.GetAuthStatusHandler)
	r.POST("/api/v1/auth/login", controller.LoginHandler)
	r.POST("/api/v1/auth/logout", controller.LogoutHandler)

	// 受 Token 保护的路由组
	api := r.Group("/api/v1", controller.TokenAuthMiddleware())
	{
		api.GET("/monitor/interface", controller.GetInterfaceDataHandler)
		api.GET("/monitor/lan", controller.GetLanClientsHandler)
		api.GET("/monitor/network-map", controller.GetNetworkMapHandler)
		api.GET("/monitor/security-hub", controller.GetSecurityHubHandler)
		api.GET("/monitor/multi-wan", controller.GetMultiWanHandler)
		api.GET("/monitor/insights", controller.GetMonitorInsightsHandler)
		api.GET("/ip", controller.GetWanIPHandler)
		api.GET("/router/resources", controller.ListCommonResourcesHandler)
		api.GET("/router/resources/:name", controller.GetCommonResourceHandler)
		api.POST("/router/resources/:name", controller.CreateCommonResourceHandler)
		api.PUT("/router/resources/:name/:id", controller.UpdateCommonResourceHandler)
		api.DELETE("/router/resources/:name/:id", controller.DeleteCommonResourceHandler)
		api.GET("/config/routers", controller.GetRoutersConfigHandler)
		api.PUT("/config/active-router", controller.SwitchActiveRouterHandler)
		api.PUT("/config/routers", controller.SaveRoutersConfigHandler)
		api.PUT("/config/token", controller.SaveTokenHandler)
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"config":  config.GlobalPath,
			"service": service.CurrentMonitorService().Status(),
		})
	})

	registerStaticRoutes(r, config.GlobalConfig.Server.StaticDir)

	// 6. 启动监听
	addr := fmt.Sprintf(":%s", config.GlobalConfig.Server.Port)
	logger.Log.Infof("服务就绪 → http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		logger.Log.Fatalf("服务启动失败: %v", err)
	}
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
					c.File(filePath)
					return
				}
			}
		}
		if _, err := os.Stat(indexPath); err == nil {
			c.File(indexPath)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "frontend assets not found"})
	})
}

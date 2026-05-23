package main

import (
	"fmt"
	"net/http"

	"ikuai4-backend/backend/internal/config"
	"ikuai4-backend/backend/internal/controller"
	"ikuai4-backend/backend/internal/logger"
	"ikuai4-backend/backend/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 初始化 Zap 高性能日志
	logger.InitLogger()
	logger.Log.Info("==================================================")
	logger.Log.Info("  iKuai 流量监控后台服务端启动中...")
	logger.Log.Info("==================================================")

	// 2. 初始化全局配置
	config.InitConfig()
	logger.Log.Infof("配置参数载入成功！监听端口: %s", config.GlobalConfig.Port)

	// 3. 建立爱快物理连接 / 切换仿真数据源
	service.InitMonitorService()

	if config.GlobalConfig.MockMode {
		logger.Log.Warn("【警告】服务已切换为 [高保真模拟模式]。您可以在没有连接物理爱快路由器的情况下运行本系统。")
	} else {
		logger.Log.Info("【信息】系统已连接到真实爱快物理路由器！")
	}

	// 4. 初始化 Gin 引擎
	gin.SetMode(gin.ReleaseMode) // 发布模式，减少无关控制台输出
	r := gin.New()

	// 使用 Zap 接管 Gin 内部日志
	r.Use(gin.Recovery())
	r.Use(controller.CORSMiddleware())

	// 5. 注册路由组
	api := r.Group("/api/v1")
	{
		// 首页看板与网口流量
		api.GET("/monitor/interface", controller.GetInterfaceDataHandler)
		// 局域网在线客户端列表 (包含 MAC 聚合去重与备注搜索)
		api.GET("/monitor/lan", controller.GetLanClientsHandler)
	}

	// 默认健康检查路由
	r.GET("/health", func(c *gin.Context) {
		modeStr := "real"
		if config.GlobalConfig.MockMode {
			modeStr = "mock"
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"mode":   modeStr,
		})
	})

	// 6. 启动 HTTP 端口监听
	addr := fmt.Sprintf(":%s", config.GlobalConfig.Port)
	logger.Log.Infof("Go Web 服务已成功开启！准备监听接口: http://localhost%s", addr)

	if err := r.Run(addr); err != nil {
		logger.Log.Fatalf("Web 服务运行发生致命故障: %v", err)
	}
}

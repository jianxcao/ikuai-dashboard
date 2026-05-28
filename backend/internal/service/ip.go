package service

import (
	"context"
	"errors"
	"strings"

	"ikuai-dashboard/backend/internal/config"
	"ikuai-dashboard/backend/internal/logger"
)

// WanIPResult WAN 口 IP 查询结果
type WanIPResult struct {
	RouterName string `json:"router_name"`
	WanName    string `json:"wan_name"`
	IP         string `json:"ip"`
}

// GetWanIP 获取指定路由器的 WAN 口 IP
func (s *MonitorService) GetWanIP(ctx context.Context, routerName string, wanName string) (*WanIPResult, error) {
	if s == nil {
		return nil, errors.New("服务未初始化")
	}
	if s.unconfigured {
		return nil, ErrUnconfigured
	}

	iface, err := s.GetInterfaceData(ctx)
	if err != nil {
		return nil, err
	}

	if len(iface.WanStatus) == 0 {
		return nil, errors.New("未找到 WAN 口信息")
	}

	if wanName != "" {
		for _, wan := range iface.WanStatus {
			if strings.EqualFold(wan.Name, wanName) {
				return &WanIPResult{
					RouterName: s.router.Name,
					WanName:    wan.Name,
					IP:         wan.IP,
				}, nil
			}
		}
		return nil, errors.New("未找到指定的 WAN 口: " + wanName)
	}

	wan := iface.WanStatus[0]
	return &WanIPResult{
		RouterName: s.router.Name,
		WanName:    wan.Name,
		IP:         wan.IP,
	}, nil
}

// GetWanIPByRouterName 根据路由器名称获取 WAN 口 IP
func GetWanIPByRouterName(ctx context.Context, routerName string, wanName string) (*WanIPResult, error) {
	cfg := config.Snapshot()
	if cfg == nil {
		return nil, ErrUnconfigured
	}

	var targetRouter *config.RouterConfig
	for i, router := range cfg.Routers {
		if router.ID == routerName || router.Name == routerName {
			targetRouter = &cfg.Routers[i]
			break
		}
	}

	if targetRouter == nil {
		return nil, errors.New("未找到路由器: " + routerName)
	}

	activeRouter, ok := cfg.ActiveRouter()
	if ok && activeRouter.ID == targetRouter.ID {
		return CurrentMonitorService().GetWanIP(ctx, routerName, wanName)
	}

	logger.Log.Infof("为路由器 [%s] 创建临时服务实例", targetRouter.Name)
	tempSvc := NewMonitorService(*targetRouter)
	return tempSvc.GetWanIP(ctx, routerName, wanName)
}

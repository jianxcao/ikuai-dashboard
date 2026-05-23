# Docker、YAML 配置与多爱快服务器设计

## 背景

当前项目是一个 Go 后端 + Vue/Vite 前端的爱快监控面板。后端入口在 `backend/cmd/server/main.go`，配置由 `backend/internal/config/config.go` 从环境变量读取，并初始化单个全局监控服务。前端使用顶部导航在多个监控页面之间切换，PWA 已接入，但浏览器图标、PWA 图标和页面品牌标识还没有统一到项目自己的视觉资产。

本次改造目标是让项目可用 Docker 打包部署，用 YAML 文件管理运行配置，支持多台爱快服务器并能在页面切换，同时让导航结构和图标资产适合继续扩展。

## 目标

- 根目录提供可构建运行的 Dockerfile。
- 后端运行配置从 `gopkg.in/yaml.v3` 读取，不再以环境变量作为主要配置来源。
- 配置文件支持多台爱快服务器，并记录当前激活服务器。
- 页面支持查看、编辑、保存服务器配置，并切换当前激活服务器。
- 桌面端导航从顶部标签改为左侧导航；移动端提供底部主导航，并保留进入设置的入口。
- favicon、PWA 图标和页面品牌标识来自同一套爱快监控图标设计。

## 非目标

- 不实现同时展示多台爱快服务器数据。
- 不为配置页面增加登录鉴权。
- 不改现有监控接口响应结构。
- 不引入大型 UI 框架。

## 配置文件

新增根目录 `config.example.yaml`，后端默认读取 `config.yaml`。容器环境中默认路径为 `/app/config.yaml`，可以通过命令行参数 `-config` 或环境变量 `IKUAI_CONFIG` 指定其它路径。环境变量只用于定位配置文件，不再承载业务配置。

配置结构：

```yaml
server:
  port: "8080"
  static_dir: "frontend/dist"

active_router_id: "office"

routers:
  - id: "office"
    name: "办公室爱快"
    url: "https://192.168.50.1:6443"
    username: "admin"
    password: ""
    mock: true
    insecure_skip_verify: true
  - id: "home"
    name: "家用爱快"
    url: "https://192.168.1.1"
    username: "admin"
    password: ""
    mock: true
    insecure_skip_verify: true
```

`routers[].mock` 为 true 时该路由器使用模拟数据。为空密码不会自动写入任何默认真实密码。`insecure_skip_verify` 默认 true，以兼容爱快常见的自签证书。

## 后端架构

`backend/internal/config` 负责：

- 用 `yaml.v3` 读取和写入配置。
- 校验端口、激活路由器、路由器 ID 唯一性、URL、用户名和模拟模式约束。
- 对返回前端的配置做密码脱敏。

`backend/internal/service` 继续保留一个全局监控服务，但服务初始化基于当前激活路由器。切换激活路由器时，后端更新配置文件并重新初始化监控服务。所有现有监控 API 继续读取当前全局服务，因此前端监控页面不需要传 `router_id`。

新增配置 API：

- `GET /api/v1/config/routers`：返回服务器配置列表、当前激活 ID 和服务连接状态，密码字段只返回空字符串。
- `PUT /api/v1/config/active-router`：请求体 `{ "id": "office" }`，切换激活服务器并重新初始化监控服务。
- `PUT /api/v1/config/routers`：保存完整配置中的服务器列表与激活 ID，写回 YAML，再重新初始化监控服务。

写配置使用原子替换：写入临时文件后 rename，避免中途失败损坏配置。

## 前端架构

新增 `RouterSettings.vue` 页面和 `useRouterConfig` 组合函数。页面展示服务器列表、当前激活状态、模拟模式、地址、用户名和密码输入。保存时把完整列表发到后端；切换时调用激活 API 并触发现有监控页面下一次轮询读取新服务器数据。

导航改造：

- 桌面端：`AppLayout` 使用左侧栏 + 主内容区。左侧栏包含品牌、监控页面入口、设置入口、连接状态和主题切换。
- 移动端：底部主导航展示首页、局域网、网络、安全、多 WAN；设置入口放在底部的设置按钮或同一导航项中。页面内容增加底部安全间距，避免被底部导航遮挡。
- 页面切换仍使用当前本地状态，不引入路由库。

## 图标设计

新增 `frontend/public/ikuai-icon.svg` 作为统一源。视觉方向：

- 圆角方形深色底，适合浏览器标签和手机桌面。
- 中央是抽象路由器核心和闪电链路，表达爱快、网络、速度和监控。
- 颜色使用当前主题中的青蓝和绿色，避免通用 Vite 紫色图标。

`favicon.svg` 直接复用该设计；PWA 192 和 512 PNG 由同一 SVG 渲染生成。页面品牌图标也从该 SVG 或等价内联标识派生。

## Docker 架构

根目录新增多阶段 Dockerfile：

1. Node/pnpm 阶段安装前端依赖并执行 `pnpm run build`。
2. Go 阶段执行 `go build -o /out/ikuai-dashboard ./backend/cmd/server`。
3. 运行阶段复制 Go 二进制、前端 `dist`、`config.example.yaml`，默认监听 8080。

Go 后端负责托管前端静态文件：未命中 `/api` 和 `/health` 的 GET 请求返回 `index.html`，保证 Docker 容器单进程即可运行完整应用。

## 验证

- `go test ./...`
- `cd frontend && pnpm run build`
- `docker build -t ikuai-dashboard:local .`
- 运行容器后访问 `/health`、`/api/v1/config/routers` 和前端页面。
- 浏览器移动宽度下确认左侧/底部导航没有遮挡内容。

## 风险

- 写回 YAML 会包含密码。配置页面需要清晰区分“留空表示不修改原密码”和“删除密码”。第一版采用完整配置写回，并在编辑表单中保留用户输入的密码；从 API 读回时密码为空，用户不改密码则后端保留旧值。
- 切换服务器会重建全局监控服务。切换期间已有轮询请求可能短暂失败；前端按现有错误展示处理。
- Docker 内静态目录与本地开发目录不同，需要通过配置默认值和构建复制保证两种模式都能运行。

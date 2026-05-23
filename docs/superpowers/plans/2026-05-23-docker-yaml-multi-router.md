# Docker、YAML 配置与多爱快服务器实现计划

> **给自动化执行代理：** 必须使用子技能：优先使用 `superpowers:subagent-driven-development`，也可以使用 `superpowers:executing-plans` 按任务实现本计划。步骤使用复选框（`- [ ]`）语法跟踪。

**目标：** 让项目支持 Docker 单镜像部署、YAML 文件配置、多台爱快服务器页面配置与切换、扩展型导航，以及统一的爱快监控品牌图标。

**架构：** 后端新增 YAML 配置读写层和配置 API，监控服务基于当前激活路由器重建。前端新增路由器设置页和响应式导航骨架，保留现有本地页面状态切换。Docker 镜像由多阶段构建生成前端静态资源和 Go 二进制，最终由 Go 服务托管前端。

**技术栈：** Go、Gin、`gopkg.in/yaml.v3`、Vue 3、Vite、TypeScript、pnpm、Docker。

---

## 文件结构

- 新建 `config.example.yaml`：示例配置。
- 新建 `.dockerignore`：排除构建无关文件。
- 新建 `Dockerfile`：多阶段构建前端和后端。
- 修改 `go.mod`、`go.sum`：加入 `gopkg.in/yaml.v3`。
- 修改 `backend/internal/config/config.go`：YAML 配置模型、加载、校验、写回、脱敏。
- 新建 `backend/internal/config/config_test.go`：配置加载、校验、密码保留测试。
- 修改 `backend/internal/service/monitor.go`、`monitor_extended.go`：服务基于激活路由器配置工作。
- 新建 `backend/internal/controller/config.go`：配置 API。
- 修改 `backend/cmd/server/main.go`：配置路径、静态文件托管、配置 API 注册。
- 修改 `frontend/src/api/monitor.ts`：增加配置 API 类型和函数。
- 新建 `frontend/src/composables/useRouterConfig.ts`：配置页面状态与请求封装。
- 新建 `frontend/src/views/RouterSettings.vue`：多路由器配置与切换页面。
- 修改 `frontend/src/components/layout/AppLayout.vue`、`AppHeader.vue`：桌面左侧导航、移动底部导航。
- 新建 `frontend/public/ikuai-icon.svg`，修改 `favicon.svg`、PWA PNG 和 manifest 图标引用。

## 任务 1：YAML 配置测试与模型

**文件：**
- 新建：`backend/internal/config/config_test.go`
- 修改：`backend/internal/config/config.go`
- 修改：`go.mod`

- [ ] **步骤 1：写失败测试**

测试覆盖：读取 YAML、多服务器、激活服务器、默认值、未知字段报错、保存配置时保留未填写密码。

- [ ] **步骤 2：运行测试确认失败**

运行：

```sh
go test ./backend/internal/config
```

预期：失败，原因是还没有 YAML 加载 API。

- [ ] **步骤 3：实现配置模型**

实现 `AppConfig`、`RouterConfig`、`LoadFromFile`、`SaveToFile`、`ActiveRouter`、`PublicConfig`、`MergeRouterSecrets`。

- [ ] **步骤 4：运行配置测试**

运行：

```sh
go test ./backend/internal/config
```

预期：通过。

## 任务 2：后端服务和配置 API

**文件：**
- 修改：`backend/internal/service/monitor.go`
- 修改：`backend/internal/service/monitor_extended.go`
- 新建：`backend/internal/controller/config.go`
- 修改：`backend/cmd/server/main.go`

- [ ] **步骤 1：写控制器测试或服务级测试**

至少验证切换激活路由器会更新配置对象，并让当前服务使用新激活项。

- [ ] **步骤 2：实现服务重载**

把 `InitMonitorService()` 调整为读取 `config.GlobalConfig.ActiveRouter()`，并新增线程安全的 `ReloadMonitorService()`。

- [ ] **步骤 3：实现配置 API**

新增：

```text
GET /api/v1/config/routers
PUT /api/v1/config/active-router
PUT /api/v1/config/routers
```

- [ ] **步骤 4：注册路由并更新健康检查**

`/health` 返回当前路由器 ID、模式和状态。

- [ ] **步骤 5：运行后端测试**

运行：

```sh
go test ./...
```

预期：通过。

## 任务 3：前端路由器配置页面

**文件：**
- 修改：`frontend/src/api/monitor.ts`
- 新建：`frontend/src/composables/useRouterConfig.ts`
- 新建：`frontend/src/views/RouterSettings.vue`
- 修改：`frontend/src/components/layout/AppLayout.vue`

- [ ] **步骤 1：增加 API 类型和函数**

增加路由器配置的 TypeScript 类型，封装获取、保存、切换 API。

- [ ] **步骤 2：实现组合函数**

维护列表、激活 ID、加载、保存、切换、错误和提交状态。

- [ ] **步骤 3：实现设置页面**

提供新增、删除、编辑、保存和切换按钮。密码字段为空时表示保留原密码。

- [ ] **步骤 4：接入页面切换**

`AppLayout` 增加 `settings` 页面。

- [ ] **步骤 5：运行前端类型检查**

运行：

```sh
cd frontend && pnpm run type-check
```

预期：通过。

## 任务 4：响应式导航改造

**文件：**
- 修改：`frontend/src/components/layout/AppLayout.vue`
- 修改：`frontend/src/components/layout/AppHeader.vue`
- 修改：`frontend/src/assets/main.css`

- [ ] **步骤 1：把顶部导航改成侧栏组件语义**

复用 `AppHeader.vue`，但渲染为桌面侧栏和移动底部导航。

- [ ] **步骤 2：移动端底部导航**

移动端固定底部，主内容增加底部 padding，确保不遮挡。

- [ ] **步骤 3：设置入口**

桌面侧栏直接展示设置入口；移动端底部也展示设置入口，避免隐藏关键配置能力。

- [ ] **步骤 4：运行构建**

运行：

```sh
cd frontend && pnpm run build
```

预期：通过。

## 任务 5：统一图标资产

**文件：**
- 新建：`frontend/public/ikuai-icon.svg`
- 修改：`frontend/public/favicon.svg`
- 修改：`frontend/public/pwa-192x192.png`
- 修改：`frontend/public/pwa-512x512.png`
- 修改：`frontend/vite.config.ts`
- 修改：`frontend/vite.config.js`
- 修改：`frontend/src/components/layout/AppHeader.vue`

- [ ] **步骤 1：创建 SVG 源图标**

图标为深色圆角底、青蓝/绿色网络链路和速度闪电。

- [ ] **步骤 2：生成 PWA PNG**

用 Node 脚本或现有可用工具从同一设计生成 192 和 512 PNG。

- [ ] **步骤 3：更新 favicon 和 manifest**

`favicon.svg` 与 PWA PNG 统一来源，manifest 图标路径保持稳定。

- [ ] **步骤 4：运行构建并检查 manifest**

运行：

```sh
cd frontend && pnpm run build
sed -n '1,120p' dist/manifest.webmanifest
```

预期：manifest 图标路径和 PWA 产物存在。

## 任务 6：Docker 打包

**文件：**
- 新建：`Dockerfile`
- 新建：`.dockerignore`
- 新建：`config.example.yaml`
- 修改：`backend/cmd/server/main.go`

- [ ] **步骤 1：让 Go 服务托管前端 dist**

未命中 API 的 GET 请求返回静态文件或 `index.html`。

- [ ] **步骤 2：编写 Dockerfile**

多阶段构建前端和后端，运行阶段复制二进制、前端 dist 和示例配置。

- [ ] **步骤 3：编写 .dockerignore**

排除 `node_modules`、`dist`、`.git`、临时文件和本地配置。

- [ ] **步骤 4：构建镜像**

运行：

```sh
docker build -t ikuai-dashboard:local .
```

预期：镜像构建成功。

- [ ] **步骤 5：容器 smoke test**

运行：

```sh
docker run --rm -p 18080:8080 ikuai-dashboard:local
curl http://localhost:18080/health
curl http://localhost:18080/api/v1/config/routers
```

预期：健康检查和配置 API 返回 200。

## 最终验证

运行：

```sh
go test ./...
cd frontend && pnpm run build
docker build -t ikuai-dashboard:local .
```

并进行一次本地后端启动 smoke：

```sh
go run ./backend/cmd/server -config config.example.yaml
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/config/routers
```

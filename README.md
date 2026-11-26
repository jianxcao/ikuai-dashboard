# iKuai Console V4 (`V4` 主分支)

面向 **爱快 OS v4.0** 的纯原生 REST + API Token 监控与全功能运维控制台（Go + Gin 后端，Vue 3 + Vite + ECharts 前端）。

## 核心特性

- **纯 V4 原生接入**：彻底移除历史 v3 `/Action/call` 依赖，原生对接 `/api/v4.0/*` 与 Bearer Token 鉴权，内置固件 `nil` 容错、map-of-maps 与 `*_data` 集合解包、`GET -> Merge -> PUT` 安全更新与 `?dry_run=true` 请求预览。
- **多爱快 V4 节点矩阵**：支持配置文件与界面同时维护多台爱快 4.x 路由器，并发探测在线状态、延迟、固件版本、CPU/内存与速率，支持保存前「测试连接」及顶部一键切换当前受管节点。
- **6 套独立设计主题（苹果液态玻璃领衔）**：
  - `apple-liquid-glass`：液态玻璃 · 昼光（Apple Liquid Glass 昼间通透悬浮玻璃与霓虹弥散背景）
  - `vision-glass-dark`：液态暗晶 · 星夜（Dark Spatial Vision Glass 暗黑景深玻璃与星云光晕）
  - `cupertino-studio`：极简云白 · 工作室（高对比无纹理现代云白）
  - `nordic-paper`：暖砂纸感 · 护眼（低蓝光温润羊皮纸色调）
  - `linear-graphite`：钛黑星轨 · 极客（极致暗黑高精度工程桌面）
  - `te-industrial`：机械橙工控 · 硬件（Hardware Precision 实体仪表锐利轮廓）
- **安全鉴权与写保护模式**：
  - 支持控制台管理员登录（HMAC-SHA256 Session Token），可通过 `USER_NAME` / `PASSWORD` 或 `config.yaml` 配置凭据，开发态默认 `admin` / `admin`（设置 `DISABLE_AUTH=true` 可跳过登录）。
  - 支持 **外部 API Token 与 `/api/v1/ip`**：可在界面生成静态 `Access Token`，供 DDNS 或自动化脚本直接获取指定节点、指定 WAN 口的纯文本出口 IP。
  - 支持 **只读保护双防线**：后端设置 `READ_ONLY=true` 或 `server.read_only: true` 拒绝除登录、测活与 `?dry_run=true` 外的所有写入接口（返回 HTTP `403`）；前端顶部栏内置「写保护锁」可随时本地锁定危险写入操作。
- **全模块交互式管理（对齐 `ikuai-cli`）**：
  1. **总览驾驶舱**：硬件负载、`ether_info` 网口面板、WAN 健康检查、双时序图表与系统资源历史曲线，外网总计吞吐与接口 WAN/LAN 角色明细对照，支持指标卡与模块拖拽互换并固定布局，以及一键「同步 NTP / 立即备份 / 检查更新 / 重启 DHCP」。
  2. **终端与认证用户**：IPv4/IPv6 在线、离线历史、单终端 5 分钟画像、快捷操作（标注、转静态 DHCP、限速、MAC 阻断）、在线踢下线、认证账号/套餐、PPPoE 与 Web 认证服务。
  3. **流量与拓扑**：应用与传输协议排行及可交互网络拓扑图（节点点击联动规则入口）。
  4. **网络与接口**：WAN/LAN/物理网口（表单内置接口下拉选择）、DHCP 服务/静态/租约（含一键转静态）、访问控制模式/规则、DNS/VLAN/DNAT/NAT/DMZ、PPPoE 拨号。
  5. **多线分流路由与 QoS**：静态路由、域名/五元组/L7 分流、负载均衡、上下行分离、IP/MAC 限速。
  6. **防火墙与安全中心**：暴露面态势卡片与跳转、ACL、MAC 过滤模式与规则、L7 管控、URL 黑名单/关键字/重定向/重写、域名黑名单、连接数限制、终端标注、高级安全防护与二级路由检测。
  7. **策略对象分组**：7 类 IP/IPv6/MAC/端口/协议/域名/时间段对象分组、规则表内 `@组名` 药丸弹窗快改、动态跳转高亮与策略引用反查。
  8. **VPN 隧道**：WireGuard（含随机 X25519 密钥生成与 Peer 抽屉管理）、IPSec 站点隧道、OpenVPN/PPTP/L2TP 服务与客户端、IKEv2。
  9. **无线网络**：无线 AC 控制器、AP 设备列表、无线黑白名单与无线终端 VLAN 绑定。
  10. **日志审计中心**：10 类系统/认证/DHCP/URL 日志检索、JSON 详情查看与清空。
  11. **系统维护与高级服务**：基础与 NTP、远程访问、BBR 内核调优、CPU 调频、ALG、定时重启、VRRP、备份快照/自动备份、升级检查、磁盘、文件目录浏览、FTP/HTTP/Samba/SNMPD、Web 管理角色组与账号。
  12. **全站智能排序与 `⌘K / Ctrl+K` 命令面板**：支持全站规则/终端/接口/应用负载表头智能前端排序（自然 IP、带单位速率/时长等），按中文模块名或 `ikuai-cli` 命令即时检索直达，表格与配置卡片均支持完整 JSON 检查器与可编辑 Raw JSON 视图。

## Docker Compose 部署（推荐）

本仓库根目录已提供 [`docker-compose.yml`](docker-compose.yml)，支持 `linux/amd64` 与 `linux/arm64` 双架构镜像 `jianxcao/ikuai-dashboard-v4`：

```yaml
services:
  ikuai-dashboard-v4:
    image: jianxcao/ikuai-dashboard-v4:latest
    container_name: ikuai-dashboard-v4
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      # 注意：请使用全新目录，切勿直接挂载旧版 (main/V3) 的 config.yaml
      - ./config:/config
    environment:
      TZ: Asia/Shanghai
      PUID: "1000"
      PGID: "1000"
      USER_NAME: admin
      PASSWORD: change_me_please
      # SESSION_SECRET: "replace-with-a-random-secret"
      # READ_ONLY: "false"
```

启动服务：

```sh
docker compose up -d
```

或使用 `docker run` 单命令启动：

```sh
docker run -d \
  --name ikuai-dashboard-v4 \
  --restart unless-stopped \
  -p 8080:8080 \
  -v $(pwd)/config:/config \
  -e TZ=Asia/Shanghai \
  -e USER_NAME=admin \
  -e PASSWORD=change_me_please \
  jianxcao/ikuai-dashboard-v4:latest
```

首次启动如果 `/config/config.yaml` 不存在，容器会自动复制默认模板；随后打开 `http://<宿主机IP>:8080`，在 **「多节点与主题 → 新增 V4 路由器」** 中填入爱快 4.x 的 URL 与 API Token 即可（Token 生成位置：爱快后台「高级设置 → 系统设置 → API Token」）。

### 从旧版（`main` / V3）迁移注意事项

1. **镜像名已独立**：V4 使用 **`jianxcao/ikuai-dashboard-v4`**，与旧版 `jianxcao/ikuai-dashboard` 完全隔离，可同时部署或独立升级。
2. **切勿直接挂载旧版 `config.yaml`**：容器内挂载路径仍为 `/config`、端口仍为 `8080`，但 V4 配置启用了严格的未知字段检查并仅支持纯 V4 API Token（`routers[].token`）；旧版 `routers[].username`、`password`、`version` 等字段会导致启动报错。请挂载新的宿主机目录（例如 `./config:/config`）并在界面中重新添加 V4 节点。
3. **环境变量变更**：如不通过配置文件而是通过环境变量接入单台路由器，需使用 `IKUAI_URL`（或 `IKUAI_CLI_BASE_URL`）+ `IKUAI_TOKEN`（或 `IKUAI_CLI_TOKEN`），原旧版的 `IKUAI_USERNAME` / `IKUAI_PASSWORD` 已不再使用。

## 环境变量说明

| 变量名 | 默认值 | 说明 |
|---|---|---|
| `USER_NAME` | `admin` | 控制台登录用户名（优先级高于 `config.yaml` 中的 `server.username`） |
| `PASSWORD` | `admin` | 控制台登录密码（优先级高于 `config.yaml` 中的 `server.password`） |
| `SESSION_SECRET` | 进程随机盐 | 登录 Session 签名密钥；设置固定值后容器重启不会使已登录浏览器失效 |
| `SESSION_TTL_DAYS` | `365` | 登录 Session 有效期（天） |
| `DISABLE_AUTH` | `false` | 设为 `true` 跳过所有接口认证（仅建议受信任内网使用） |
| `READ_ONLY` | `false` | 设为 `true` 开启服务端强制只读保护（拒绝所有非 `GET` 写入，仍放行测试与 `?dry_run=true`） |
| `PORT` | `8080` | 后端 HTTP 监听端口 |
| `IKUAI_CONFIG` | `config.yaml`（容器内 `/config/config.yaml`） | 配置文件路径 |
| `PUID` / `PGID` / `UMASK` / `TZ` | `1000` / `1000` / `022` / UTC | 容器内运行用户 UID/GID、掩码与时区 |
| `IKUAI_URL` / `IKUAI_CLI_BASE_URL` | 空 | 单节点环境变量接入时的爱快 V4 地址（当 `routers` 为空时生效，仅驻留内存） |
| `IKUAI_TOKEN` / `IKUAI_CLI_TOKEN` | 空 | 单节点环境变量接入时的爱快 V4 API Token |
| `IKUAI_NAME` | `环境变量节点 (V4)` | 单节点环境变量接入时的显示名称 |
| `MOCK_MODE` | `false` | 设为 `true` 启用内置双节点高保真模拟模式（无需真机） |

## 外部 API：获取指定节点与 WAN 口出口 IP（`/api/v1/ip`）

V4 保留了旧版的纯文本出口 IP 查询接口，适合 DDNS 或自动化脚本调用。先在控制台 **「多节点与主题 → 读写安全锁与访问认证保护」** 中生成并复制 **访问 Token**（对应 `config.yaml` 中的 `server.access_token`）：

```sh
# 获取指定节点默认 WAN 口纯文本 IP（name 可传节点 ID 或节点名称）
curl "http://localhost:8080/api/v1/ip?name=main&token=<ACCESS_TOKEN>"

# 获取指定节点 wan2 口纯文本 IP（也兼容旧版 lan=wan2 参数）
curl "http://localhost:8080/api/v1/ip?name=main&wan=wan2&token=<ACCESS_TOKEN>"

# 返回结构化 JSON 详情
curl "http://localhost:8080/api/v1/ip?router_id=main&wan=wan1&format=json&token=<ACCESS_TOKEN>"
```

- **节点参数**：支持 `name`、`router_id`、`id` 或请求头 `X-Router-ID`（均支持节点 ID 或名称匹配；不传时默认取当前激活节点）。
- **线路参数**：支持 `wan`、`interface`、`iface`、`lan`（不传时默认取第一个有有效 IP 的 WAN 口）。
- **认证方式**：支持 `?token=<ACCESS_TOKEN>` 或请求头 `Authorization: Bearer <ACCESS_TOKEN>`。

## 本地开发与快速启动

### 1. 模拟模式（无需真机，内置双 V4 节点）

```sh
MOCK_MODE=true go run ./backend/cmd/server
```

另一个终端启动前端开发服务：

```sh
cd frontend
pnpm install
pnpm dev
```

### 2. 使用环境变量接入单台 V4 真机（不写入磁盘）

```sh
export IKUAI_CLI_BASE_URL="https://192.168.9.1"
export IKUAI_CLI_TOKEN="<YOUR_API_TOKEN>"
# 可选：自定义控制台登录账号与只读保护模式
export USER_NAME="admin"
export PASSWORD="<CONSOLE_PASSWORD>"
export READ_ONLY=false
go run ./backend/cmd/server
```

### 3. 多节点 YAML 配置

参考 [`config.example.yaml`](config.example.yaml) 配置多台 V4 路由器、控制台账号与 `active_router_id`。

## 验证命令

```sh
go test -v ./...
cd frontend && pnpm run build
```

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
  - 支持 **只读保护双防线**：后端设置 `READ_ONLY=true` 或 `server.read_only: true` 拒绝除登录、测活与 `?dry_run=true` 外的所有写入接口（返回 HTTP `403`）；前端顶部栏内置「写保护锁」可随时本地锁定危险写入操作。
- **全模块交互式管理（对齐 `ikuai-cli`）**：
  1. **总览驾驶舱**：硬件负载、`ether_info` 网口面板、WAN 健康检查、双时序图表，以及一键「同步 NTP / 立即备份 / 检查更新 / 重启 DHCP」。
  2. **终端与认证用户**：IPv4/IPv6 在线、离线历史、单终端 5 分钟画像、快捷操作（标注、转静态 DHCP、限速、MAC 阻断）、在线踢下线、认证账号/套餐、PPPoE 与 Web 认证服务。
  3. **流量与拓扑**：应用与传输协议排行及可交互网络拓扑图（节点点击联动规则入口）。
  4. **网络与接口**：WAN/LAN/物理网口、DHCP 服务/静态/租约（含一键转静态）、访问控制模式/规则、DNS/VLAN/DNAT/NAT/DMZ、PPPoE 拨号。
  5. **多线分流路由与 QoS**：静态路由、域名/五元组/L7 分流、负载均衡、上下行分离、IP/MAC 限速。
  6. **防火墙与安全中心**：暴露面态势卡片与跳转、ACL、MAC 过滤模式与规则、L7 管控、URL 黑名单/关键字/重定向/重写、域名黑名单、连接数限制、终端标注、高级安全防护与二级路由检测。
  7. **VPN · 无线 · 对象组**：WireGuard（含随机 X25519 密钥生成与 Peer 抽屉管理）、IPSec、OpenVPN/PPTP/L2TP 服务与客户端、IKEv2、无线 AC/AP/黑白名单/VLAN、7 类对象组及引用关系反查。
  8. **日志审计中心**：10 类系统/认证/DHCP/URL 日志检索、JSON 详情查看与清空。
  9. **系统维护与高级服务**：基础与 NTP、远程访问、BBR 内核调优、CPU 调频、ALG、定时重启、VRRP、备份快照/自动备份、升级检查、磁盘、文件目录浏览、FTP/HTTP/Samba/SNMPD、Web 管理角色组与账号。
  10. **`⌘K / Ctrl+K` 命令面板**：按中文模块名或 `ikuai-cli` 命令即时检索直达，表格与配置卡片均支持完整 JSON 检查器与可编辑 Raw JSON 视图。

## 快速启动

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

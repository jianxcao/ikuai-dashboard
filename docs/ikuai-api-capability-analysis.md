# iKuai API 能力与页面规划

> 更新日期：2026-05-23  
> 目标：分析 `github.com/zy84338719/ikuai-api` 与 `ikuaidev/ikuai-cli` 暴露出的 iKuai API 能力，并判断本项目后续可以实现哪些页面。

## 1. 资料来源

本次分析基于以下源码快照：

| 来源 | 快照 | 说明 |
| --- | --- | --- |
| [`zy84338719/ikuai-api`](https://github.com/zy84338719/ikuai-api) | `21fc6b0d08e27c08ea5f20c98f8e207090052a01`，2026-05-05 | Go SDK，支持 v3 `Action/call` 与 v4 REST |
| [`ikuaidev/ikuai-cli`](https://github.com/ikuaidev/ikuai-cli) | `68955d81a52d70e027ee2a02bde335ae4760377e`，2026-05-11 | v4 Token CLI，覆盖监控、网络、安全、路由、VPN、系统等管理场景 |
| 本项目 | `go.mod` 当前依赖 `github.com/zy84338719/ikuai-api v0.2.0` | 已实现监控首页、LAN 客户端、配置页；网络拓扑、安全中心、多 WAN 目前主要是 mock |

补充说明：`ikuai-cli` 运行时帮助命令在本机因 Go 模块代理连接失败未跑通，本分析使用其源码和 `docs/cli-reference.md` 静态内容。

## 2. 协议模型

### 2.1 iKuai v3：`Action/call`

v3 使用单入口 RPC 风格接口：

```json
{
  "func_name": "monitor_lanip",
  "action": "show",
  "param": {}
}
```

关键点：

- 登录入口是 `/Action/login`，业务入口是 `/Action/call`。
- SDK 的 `Client.Call(ctx, funcName, action, param, result)` 会构造上述请求。
- SDK 新增了 `V3ActionClient` 和 `V3EndpointCatalog`，可以按兼容 endpoint 名称调用，而不是直接硬编码 `func_name`。
- 成功码在 SDK 中按 `Result == 10000` 或 `Result == 30000` 处理。

### 2.2 iKuai v4：REST + Token

v4 使用 REST 风格接口：

- API base：`/api/v4.0`
- 鉴权：`Authorization: Bearer <token>`
- 响应 envelope：常见成功码为 `code == 0` 或 `code == 20000`
- 数据字段：多数接口返回 `data`，部分监控/load 接口返回 `results`，新增类响应可能只返回 `rowid`

本项目当前 `backend/internal/service/monitor_v4.go` 已直接使用 SDK 的 `V4Client.Get(ctx, path, params, &payload)` 访问：

- `/monitoring/system`
- `/monitoring/interfaces-status`
- `/monitoring/interfaces-traffic`
- `/monitoring/clients-online`
- `/monitoring/clients-ip6-online`

## 3. `ikuai-api` 能力梳理

### 3.1 SDK typed service

`ikuai-api/service/interface.go` 当前定义了这些服务：

| 服务 | 能力范围 | 对本项目价值 |
| --- | --- | --- |
| `Monitor()` | LAN IPv4/IPv6 在线终端、接口状态/流量、系统负载、ARP | 已用于首页和 LAN 客户端；可扩展拓扑、ARP、实时趋势 |
| `System()` | 首页状态、升级信息、备份列表、Web 用户 | 可做系统状态、升级、备份、管理员审计 |
| `Network()` | WAN/LAN/VLAN/IPv6/IPTV/DDNS/DHCP/DNS/静态路由/策略路由/QoS | 可做网络配置、DHCP、DNS、VLAN、QoS、路由页面 |
| `Firewall()` | ACL、DNAT、连接数限制、域名对象、自定义 ISP、域名分流 | 可做安全中心、端口映射、访问控制、分流策略 |
| `VPN()` | PPTP/L2TP 客户端 CRUD | 可做基础 VPN 页面；v4 可进一步覆盖 OpenVPN/IKEv2/IPSec/WireGuard |
| `Log()` | notice、PPPoE、DHCP、ARP、DDNS、Web admin、系统事件 | 可做日志中心和异常时间线 |
| `Docker()` | 镜像、容器、网络、Compose | 可做高级服务/容器状态页面 |
| `VM()` | QEMU VM 列表、增删改、启停、重启 | 可做虚拟机管理页面 |
| `UPnP()` | UPnP 规则 CRUD | 可放在安全中心或网络服务页面 |
| `Traffic()` | 实时流量、历史流量 | 可增强趋势图、Top 客户端、历史分析 |
| `AppControl()` | 应用控制 CRUD | 可做应用管控页面 |
| `UserManage()` | 用户账号 CRUD | 可做认证用户/套餐页面 |
| `OnlineMonitor()` | 在线用户 | 可做在线用户详情、踢下线入口 |

### 3.2 v3 兼容目录

SDK 的 `V3EndpointCatalog` 覆盖 53 个 endpoint。核心映射如下：

| 分组 | endpoint | v3 `func_name` | action |
| --- | --- | --- | --- |
| monitoring | `homepage` / `system` / `clients-online` / `clients-ip6-online` / `interfaces` / `arp` / `traffic-realtime` / `traffic-history` | `homepage`、`monitor_system`、`monitor_lanip`、`monitor_lanipv6`、`monitor_iface`、`arp`、`traffic_realtime`、`traffic_history` | show |
| network | `wan` / `lan` / `vlan` / `ipv6` / `iptv` / `ddns` / `dhcp-services` / `dhcp-static` / `dhcp-clients` / `dns-forward` / `dns-static` / `pppoe-services` / `qos` / `bandwidth` / `flow-control` | `wan`、`lan`、`vlan`、`ipv6`、`iptv`、`ddns`、`dhcpd`、`dhcp_static`、`dhcp_lease`、`dns_forward`、`dns_static`、`qos`、`bandwidth`、`flow_control` | 多数 show，部分 add/edit/del |
| routing | `static-routes` / `policy-routes` / `domain-rules` | `route_static`、`route_policy`、`stream_domain` | show，静态路由支持 add/edit/del |
| security | `acl-rules` / `dnat-rules` / `peerconn-rules` / `domain-groups` / `custom-isp` / `app-control` | `acl`、`dnat`、`conn_limit`、`domain_group`、`custom_isp`、`app_control` | ACL/DNAT/peerconn/app-control 支持 CRUD |
| vpn | `pptp-clients` / `l2tp-clients` | `pptp_client`、`l2tp_client` | CRUD |
| system | `basic-config` / `upgrade` / `backup` / `web-users` / `upnp` | `homepage`、`upgrade`、`backup`、`webuser`、`upnp` | show，UPnP 支持 CRUD |
| users | `accounts` / `online` | `user_manage`、`online_monitor` | accounts CRUD，online show |
| log | `notice` / `pppoe` / `dhcp` / `arp` / `ddns` / `web-activity` / `system` | `syslog-*` | show |
| advanced | `docker-*` / `vm` | `docker_image`、`docker_container`、`docker_network`、`docker_compose`、`qemu` | Docker show，VM 支持启停和 CRUD |

v3 明确标注不支持或未知的能力：

- OpenVPN / IKEv2 / IPSec / WireGuard
- wireless access-control-rules

这些能力如果要做完整页面，应优先走 v4 REST。

### 3.3 v4 REST 目录

SDK 的 `V4EndpointCatalog` 当前有 134 个 endpoint，按分组：

| 分组 | 数量 | 代表能力 |
| --- | ---: | --- |
| monitoring | 37 | 系统概览、CPU/内存/磁盘/温度历史、终端数、连接数、接口状态/流量、在线/离线客户端、协议流量、应用流量、无线质量、摄像头、交换机、流控分流 |
| network | 25 | AC/AP、DHCP/DHCPv6、DNS、DNS 代理、NAT/DNAT/DMZ、PPPoE、QoS、VLAN |
| security | 13 | ACL、应用协议管控、域名黑名单、MAC 黑白名单、连接数限制、终端管控、URL 黑/关键字/重定向/替换、高级安全、二级路由 |
| system | 11 | 基础配置、NTP 同步、重启计划、远程访问、VRRP、ALG、内核参数、CPU 频率 |
| vpn | 10 | PPTP/L2TP/OpenVPN/IKEv2/IPSec/WireGuard 服务和客户端 |
| log | 9 | ARP、认证、DDNS、DHCP、通知、PPPoE、系统、Web 活动、无线日志，支持清理 |
| objects | 7 | IP、IPv6、MAC、端口、协议、域名、时间对象 |
| advanced | 6 | FTP/HTTP/Samba/SNMPD 高级服务 |
| routing | 6 | 静态路由、域名分流、五元组、负载均衡、上下行分离、应用协议分流 |
| interfaces | 4 | WAN、LAN、物理网口、WAN VLAN 只读配置 |
| auth | 4 | 认证用户、套餐、在线认证用户、Web 服务 |
| wireless | 2 | 无线接入控制、无线 VLAN |

## 4. `ikuai-cli` 能力梳理

`ikuai-cli` 是 v4 REST 的高覆盖度参考实现。它的核心特征：

- 只需要 `IKUAI_CLI_BASE_URL` 和 `IKUAI_CLI_TOKEN`，适合 v4 Token 模式。
- 输出支持 table/json/yaml/raw，列表命令普遍支持分页、过滤、排序。
- 代码中对不少复杂写接口做了默认字段补齐，可以反推页面表单需要哪些字段。
- 它还使用了一些未列入 `ikuai-api` 当前 `V4EndpointCatalog` 的 v4 路径，例如 `/system/backup`、`/system/upgrade`、`/system/web-admin/*`、`/system/disks`、`/system/files`。后续如果要做这些页面，可以参考 CLI 源码直接用 SDK 的 generic `V4Client`。

主要命令分组与页面价值：

| CLI 分组 | 覆盖能力 | 可反推页面 |
| --- | --- | --- |
| `monitor` | 系统概览、负载历史、接口、客户端、协议/应用流量、无线、摄像头、交换机、分流 | 总览、趋势、终端、应用流量、无线、拓扑 |
| `network` | DNS、WAN/LAN、物理网口、DHCP、NAT/DNAT、VLAN | 网络配置、地址池、端口映射、VLAN、DNS |
| `users` | 认证账号、在线用户、套餐、踢下线 | 认证用户管理、在线会话 |
| `system` | 基础配置、计划重启、远程访问、VRRP、ALG、内核、CPU 频率、磁盘、文件、备份、升级、Web admin | 系统设置、备份升级、管理员、运行参数 |
| `security` | ACL、MAC、L7、URL、域名黑名单、peerconn、终端、二级路由、高级安全 | 安全中心、访问控制、终端准入、URL 管控 |
| `vpn` | PPTP/L2TP/OpenVPN/IKEv2/IPSec/WireGuard | VPN 管理中心 |
| `routing` | 静态路由、域名分流、五元组、L7、负载均衡、上下行分离 | 多 WAN/分流策略 |
| `qos` | IP/MAC 限速规则 | QoS 限速页面 |
| `log` | 系统、ARP、认证、DHCP、PPPoE、Web、DDNS、通知、无线 | 日志审计中心 |
| `wireless` | AC、AP、无线黑名单、无线 VLAN | 无线/AP 管理 |
| `advanced` | FTP、HTTP、Samba、SNMPD | 高级服务管理 |
| `objects` | IP/MAC/端口/协议/域名/时间对象 | 对象库，供 ACL/分流/安全规则复用 |

## 5. 本项目当前状态

当前已注册后端接口：

- `GET /health`
- `GET /api/v1/monitor/interface`
- `GET /api/v1/monitor/lan?search=...`
- `GET /api/v1/monitor/network-map`
- `GET /api/v1/monitor/security-hub`
- `GET /api/v1/monitor/multi-wan`
- `GET /api/v1/config/routers`
- `PUT /api/v1/config/active-router`
- `PUT /api/v1/config/routers`

当前真实数据接入情况：

| 页面/接口 | v3 真实数据 | v4 真实数据 | 备注 |
| --- | --- | --- | --- |
| 首页看板 | `homepage`、`monitor_iface` | `/monitoring/system`、`/monitoring/interfaces-status`、`/monitoring/interfaces-traffic` | 已接真实接口 |
| LAN 客户端 | `monitor_lanip`、`monitor_lanipv6` | `/monitoring/clients-online`、`/monitoring/clients-ip6-online` | 已接真实接口 |
| 网络拓扑 | 未接 | 未接 | 当前从 mock LAN 客户端生成 |
| 安全中心 | 未接 | 未接 | 当前 mock 高危端口和异常设备 |
| 多 WAN | 部分复用 WAN 状态 mock | 未接分流规则 | 当前路由规则 mock |
| 路由器配置 | 本项目 YAML | 本项目 YAML | 管理本应用连接配置，不是爱快系统配置 |

## 6. 可实现页面规划

### P0：优先补齐当前已有页面的真实数据

这些页面已经有 UI 入口，改造成本低，且直接提升可用性。

| 页面 | 可以接入的 API | 后端聚合建议 |
| --- | --- | --- |
| 网络拓扑 | v3: `arp`、`monitor_lanip`、`monitor_lanipv6`、`wan`、`lan`；v4: `/monitoring/clients-online`、`/monitoring/clients-ip6-online`、`/monitoring/interfaces-status`、`/interfaces/lan-config`、`/interfaces/wan-config` | 以 router 为中心，WAN/LAN/interface 为二级节点，客户端按 MAC 聚合；v3 可用 ARP 补充 IP/MAC 映射 |
| 安全中心 | v3: `dnat`、`acl`、`conn_limit`、`upnp`、`syslog-*`；v4: `/network/dnat/rules`、`/security/acl-rules`、`/security/peerconn/rules`、`/log/*` | 端口暴露、规则风险、连接数异常、近 24h 安全日志聚合；UPnP 在当前资料中主要走 v3 SDK |
| 多 WAN / 分流 | v3: `wan`、`route_policy`、`stream_domain`、`qos`、`flow_control`；v4: `/interfaces/wan-config`、`/routing/domain-rules`、`/routing/five-tuple-rules`、`/routing/load-balance-rules`、`/routing/updown`、`/monitoring/flow-shunting` | WAN 状态 + 分流策略 + 命中/流量统计 |

### P1：监控类增强页面

| 页面 | API | 页面内容 |
| --- | --- | --- |
| 系统运行状态 | v3: `monitor_system`、`homepage`；v4: `/monitoring/system`、`/monitoring/cpu`、`/monitoring/memory`、`/monitoring/disk`、`/monitoring/cputemp` | CPU/内存/磁盘/温度、运行时间、版本、历史曲线 |
| 流量分析 | v3: `traffic_realtime`、`traffic_history`；v4: `/monitoring/clients-traffic-summary`、`/monitoring/clients-traffic-load`、`/monitoring/protocols`、`/monitoring/app-traffic-summary` | Top 客户端、协议占比、应用占比、历史趋势 |
| 在线/离线终端 | v3: `monitor_lanip`、`monitor_lanipv6`；v4: `/monitoring/clients-online`、`/monitoring/clients-offline`、`/monitoring/clients-ip6-online`、`/monitoring/clients-ip6-offline` | 在线、离线、双栈、备注、连接数、历史流量、搜索过滤 |
| 下游设备 | v4: `/monitoring/downstream`、`/monitoring/cameras`、`/monitoring/switch` | 摄像头、交换机、下游网络设备列表和状态 |
| 无线监控 | v4: `/monitoring/wireless-statistics`、`/monitoring/wireless-score`、`/monitoring/wireless-traffic`、`/monitoring/ssid-clients`、`/monitoring/channel-clients` | SSID、信道、AP 质量、无线终端和流量 |

### P2：配置与管理页面

这些页面涉及写操作，应优先做“只读 + 变更预览”，再逐步开放编辑。

| 页面 | API | 页面内容 |
| --- | --- | --- |
| DHCP / 地址管理 | v3: `dhcpd`、`dhcp_static`、`dhcp_lease`；v4: `/network/dhcp/services`、`/network/dhcp/static`、`/network/dhcp/clients`、`/network/dhcp6/*` | 地址池、租约、静态绑定、IPv6 DHCP |
| DNS 管理 | v3: `dns_forward`、`dns_static`；v4: `/network/dns/config`、`/network/dns/proxy/rules`、`/network/dns/stats` | DNS 配置、域名代理、查询统计 |
| VLAN / 接口 | v3: `vlan`、`lan`、`wan`；v4: `/network/vlan`、`/interfaces/*`、`/network/pppoe/services` | VLAN、物理口、WAN/LAN、PPPoE |
| NAT / 端口映射 | v3: `dnat`；v4: `/network/nat/rules`、`/network/dnat/rules`、`/network/dmz/rules` | NAT/DNAT/DMZ 规则、风险标识 |
| QoS 限速 | v3: `qos`、`bandwidth`；v4: `/network/qos/ip`、`/network/qos/mac` | IP/MAC 限速、启停、命中对象 |
| 路由策略 | v3: `route_static`、`route_policy`、`stream_domain`；v4: `/routing/static-routes`、`/routing/domain-rules`、`/routing/five-tuple-rules`、`/routing/load-balance-rules`、`/routing/updown` | 静态路由、域名分流、五元组、负载均衡、上下行分离 |
| 对象库 | v4: `/ip-objects`、`/mac-objects`、`/port-objects`、`/protocol-objects`、`/domain-objects`、`/time-objects` | 安全/路由规则复用对象 |

### P3：高级与运维页面

| 页面 | API | 页面内容 |
| --- | --- | --- |
| 日志审计 | v3: `syslog-*`；v4: `/log/system`、`/log/arp`、`/log/auth`、`/log/dhcp`、`/log/pppoe`、`/log/web_activity`、`/log/wireless` | 多类型日志、时间线、筛选、导出、清理 |
| 系统设置 | v3: `upgrade`、`backup`、`webuser`；v4: `/system/basic/config`、`/system/reboot-schedules`、`/system/remote-access`、`/system/vrrp/config`、`/system/alg`、`/system/kernel-params`、`/system/cpufreq` | 基础配置、计划重启、远程访问、VRRP、ALG、内核、CPU 频率 |
| 备份与升级 | v3: `backup`、`upgrade`; v4: `/system/backup`、`/system/backup-auto`、`/system/upgrade`、`/system/upgrade:*` | 备份列表、自动备份、升级检查、升级状态 |
| VPN 管理 | v3: `pptp_client`、`l2tp_client`; v4: `/vpn/pptp`、`/vpn/l2tp`、`/vpn/openvpn`、`/vpn/ikev2`、`/vpn/ipsec/clients`、`/vpn/wireguard` | 各类 VPN 服务和客户端 |
| 认证用户 | v3: `user_manage`、`online_monitor`; v4: `/auth/users`、`/auth/packages`、`/auth/online-users` | 认证账号、套餐、在线用户、踢下线 |
| 高级服务 | v3: Docker/VM；v4: `/advanced-service/ftp-*`、`/advanced-service/http-users`、`/advanced-service/samba-*`、`/advanced-service/snmpd-config` | FTP/HTTP/Samba/SNMPD；v3 可做 Docker/VM |

## 7. 推荐落地顺序

1. **先补真实数据，不先做写操作。**  
   优先把 `NetworkMap`、`SecurityHub`、`MultiWan` 从 mock 切到真实数据。它们已有 UI，且多数只需要 `show/GET`。

2. **抽象一个统一的上游调用层。**  
   当前 v3 走 typed SDK，v4 走手写 path。建议在 service 层增加内部方法，例如 `callEndpoint(ctx, capability, params)`，按 router version 分派到 v3 endpoint 或 v4 path，页面 DTO 不直接暴露上游原始结构。

3. **每个新页面先定义本项目 DTO。**  
   上游字段在 v3/v4 之间命名差异大，应保持前端只消费稳定字段，例如 `enabled`、`name`、`interface`、`comment`、`hit_count`、`risk_level`。

4. **写操作默认二次确认，并先从低风险接口开始。**  
   推荐先开放 DNS proxy、DHCP static、端口映射、QoS toggle 这类可回滚规则；系统升级、备份恢复、VPN 密钥、Web admin 密码这类高风险操作应后置。

5. **mock 数据跟真实 DTO 一起维护。**  
   本项目依赖 mock 支持本地 UI 开发，新增页面时应同步补齐高保真 mock，避免真实路由不可用时前端开发停滞。

## 8. 近期最值得实现的页面包

### 8.1 “监控增强包”

适合作为下一个迭代：

- 系统运行状态：CPU、内存、磁盘、温度、连接数趋势
- 流量分析：客户端 TopN、协议占比、应用占比
- 在线/离线终端：现有 LAN 页面扩展离线、IPv6、排序、详情抽屉

收益：全部偏只读，数据来自 monitoring，风险低。

### 8.2 “安全中心真实化”

替换当前 mock：

- DNAT / UPnP 暴露面
- ACL / peerconn 风险规则
- 终端异常：高连接数、高上行、未知 MAC、长时间在线
- 日志异常：Web 登录失败、DHCP/ARP 异常、PPPoE 掉线

收益：符合当前 `SecurityHub.vue` 的产品方向，且能复用已有客户端 DTO。

### 8.3 “多 WAN 与策略路由”

替换当前 mock routes：

- WAN 状态与配置
- 域名分流
- 五元组分流
- 负载均衡
- 上下行分离
- 分流流量统计

收益：与现有 `MultiWan.vue` 高度匹配，能形成独立高级页面。

## 9. 注意事项

- 不要把爱快账号、密码、Token 暴露到前端；所有真实调用必须留在 Go 后端。
- v4 API 的 endpoint 很多，但不同固件版本可能字段和可用性不同；后端需要容错 alias 和空数组降级。
- v3 的 typed service 比 v4 typed 能力更窄，但 `V3ActionClient` 和 `V3EndpointCatalog` 可以减少硬编码。
- `ikuai-cli` 对写接口补了很多默认字段，后续做写页面时应参考对应 command 的 body 构造逻辑，不能只按 UI 表单字段提交。
- 涉及删除、升级、重启、恢复备份、踢用户、改 Web admin 的操作应单独加权限、确认、审计日志。

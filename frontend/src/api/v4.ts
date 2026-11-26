import { http } from './monitor'

// ─── Multi-Router V4 Fleet Types ────────────────────────────────────────────

export interface AppServerConfig {
  port: string
  static_dir: string
  token_enabled: boolean
  read_only?: boolean
}

export interface PublicRouterConfig {
  id: string
  name: string
  url: string
  has_token: boolean
  tags: string[]
  note: string
  mock: boolean
  insecure_skip_verify: boolean
  source: string
}

export interface EditableRouterConfig {
  id: string
  original_id?: string
  name: string
  url: string
  token?: string
  has_token?: boolean
  tags: string[]
  note: string
  mock: boolean
  insecure_skip_verify: boolean
  source?: string
}

export interface PublicAppConfig {
  server: AppServerConfig
  active_router_id: string
  routers: PublicRouterConfig[]
}

export interface MonitorServiceStatus {
  mode: 'mock' | 'real' | 'unconfigured'
  version: 'v4'
  router_id: string
  router_name: string
  source?: string
  router_count: number
  error?: string
}

export interface RouterConfigPayload {
  config: PublicAppConfig
  status: MonitorServiceStatus
}

export interface RouterHealthDTO {
  id: string
  name: string
  url: string
  tags: string[]
  note: string
  mock: boolean
  source: string
  active: boolean
  online: boolean
  latency_ms: number
  hostname: string
  version: string
  verstring: string
  arch: string
  uptime: number
  cpu_percent: number
  memory_percent: number
  upload_bps: number
  download_bps: number
  online_users: number
  connections: number
  checked_at: number
  error?: string
}

export interface RouterTestPayload {
  id?: string
  original_id?: string
  name?: string
  url: string
  token: string
  insecure_skip_verify: boolean
  mock: boolean
}

// ─── V4 Telemetry Types ─────────────────────────────────────────────────────

export interface OverviewDTO {
  router: {
    id: string
    name: string
    hostname: string
    gwid: string
    lan_ip: string
    link_status: number
    uptime: number
    version: string
    verstring: string
    arch: string
    build_date: string
    mock: boolean
  }
  stream: {
    upload_bps: number
    download_bps: number
    total_up: number
    total_down: number
  }
  cpu: {
    avg_percent: number
    cores: number[]
    freq_mhz: number[]
    temps: number[]
  }
  memory: {
    total_kb: number
    available_kb: number
    free_kb: number
    cached_kb: number
    buffers_kb: number
    used_percent: number
  }
  connections: {
    total: number
    tcp: number
    udp: number
    icmp: number
  }
  online_users: {
    total: number
    wired: number
    wireless: number
    wifi_2g: number
    wifi_5g: number
  }
  wan_checks: Array<{
    id: number
    interface: string
    parent_interface: string
    ip_addr: string
    gateway: string
    internet: string
    result: string
    errmsg: string
    auto_switch: string
    updatetime: string
    comment: string
    avg_rtt_ms: number
    drop_rate: number
  }>
  interface_traffic: Array<{
    interface: string
    ip_addr: string
    comment: string
    upload_bps: number
    download_bps: number
    total_up: number
    total_down: number
    connections: number
    updropped: number
    downdropped: number
    uppacked: number
    downpacked: number
  }>
  physical_ports: Array<{
    name: string
    interface: string
    link: number
    speed: number
    duplex: number
    mac: string
    type: string
    driver: string
    model: string
    lock: number
  }>
  wan_history: Array<{
    timestamp: number
    interface: string
    avg_upload_kbps: number
    avg_download_kbps: number
    max_upload_kbps: number
    max_download_kbps: number
    avg_rtt_ms: number
    max_rtt_ms: number
    min_rtt_ms: number
    drop_rate: number
  }>
  collected_at: number
}

export interface LoadHistoryDTO {
  metric: string
  time_range: string
  aggregate: string
  series: Array<Record<string, unknown>>
  raw?: unknown
}

export interface ClientRecordDTO {
  id: number
  mac: string
  hostname: string
  ip_addr: string
  ips: string[]
  interface: string
  upload_speed: number
  download_speed: number
  total_up: number
  total_down: number
  connections: number
  comment: string
  client_type: string
  client_model: string
  uptime: string
  downtime: string
  auth_type: string
  username: string
  ssid: string
  signal: string
}

export interface ClientPageDTO {
  mode: string
  page: number
  page_size: number
  total: number
  items: ClientRecordDTO[]
}

export interface ClientDetailDTO {
  ip: string
  mac: string
  traffic_load: Array<Record<string, unknown>>
  app_protocols: Array<Record<string, unknown>>
}

export interface AppProtocolAnalyticsDTO {
  protocols: Array<Record<string, unknown>>
  app_summary: Array<Record<string, unknown>>
  app_load: Array<Record<string, unknown>>
}

// ─── V4 CLI Resource Catalog Types ──────────────────────────────────────────

export interface OptionItem {
  label: string
  value: string
}

export type InterfaceScope = 'wan' | 'wan-all' | 'wan-auto' | 'wan-lan' | 'lan' | 'any' | 'physical-all'

export interface InterfaceOptionItem {
  name: string
  kind: 'wan' | 'lan' | 'vlan' | 'physical'
  ip?: string
  proto?: string
  comment?: string
  detail?: string
}

export interface InterfaceOptionsDTO {
  wan: InterfaceOptionItem[]
  lan: InterfaceOptionItem[]
  physical: InterfaceOptionItem[]
}

export interface FieldSpec {
  key: string
  label: string
  type: 'text' | 'number' | 'select' | 'switch' | 'textarea' | 'csv' | 'password'
  interface_scope?: InterfaceScope
  required?: boolean
  default?: unknown
  placeholder?: string
  options?: OptionItem[]
  help?: string
}

export interface ColumnSpec {
  key: string
  label: string
  kind?: 'text' | 'mono' | 'badge' | 'toggle' | 'time' | 'bytes' | 'speed' | 'compound'
}

export interface ActionSpec {
  key: string
  label: string
  path: string
  method: string
  dangerous?: boolean
  confirm_text?: string
}

export interface ResourceSpec {
  key: string
  group: string
  label: string
  kind: 'collection' | 'singleton'
  v4_path: string
  cli_command: string
  description: string
  operations: string[]
  columns: ColumnSpec[]
  fields: FieldSpec[]
  actions?: ActionSpec[]
  object_value_key?: string
  delete_mode?: 'id' | 'clear' | 'srcfile'
  supports_paging: boolean
}

export interface ResourceListResult {
  resource: ResourceSpec
  rows: Array<Record<string, unknown>>
  total: number
  config?: Record<string, unknown>
}

// ─── API Functions ──────────────────────────────────────────────────────────

export async function fetchRouters(): Promise<RouterConfigPayload> {
  const res = (await http.get('/routers')) as unknown as {
    data: PublicAppConfig
    status: MonitorServiceStatus
  }
  return { config: res.data, status: res.status }
}

export async function probeRouters(): Promise<RouterHealthDTO[]> {
  const res = await http.get('/routers/probe')
  return res.data
}

export async function testRouterConnection(payload: RouterTestPayload): Promise<RouterHealthDTO> {
  const res = await http.post('/routers/test', payload)
  return res.data
}

export async function activateRouterApi(id: string): Promise<RouterConfigPayload> {
  const res = (await http.put('/routers/active', { id })) as unknown as {
    data: PublicAppConfig
    status: MonitorServiceStatus
  }
  return { config: res.data, status: res.status }
}

export async function saveRoutersApi(payload: {
  server?: AppServerConfig
  active_router_id: string
  routers: EditableRouterConfig[]
}): Promise<RouterConfigPayload> {
  const res = (await http.put('/routers', payload)) as unknown as {
    data: PublicAppConfig
    status: MonitorServiceStatus
  }
  return { config: res.data, status: res.status }
}

export async function fetchV4Overview(): Promise<OverviewDTO> {
  const res = await http.get('/v4/overview')
  return res.data
}

export async function fetchV4InterfaceOptions(): Promise<InterfaceOptionsDTO> {
  const res = await http.get('/v4/interfaces/options')
  return res.data
}

export async function fetchV4LoadHistory(
  metric: string,
  range: string = 'hour',
  aggregate: string = 'avg'
): Promise<LoadHistoryDTO> {
  const res = await http.get('/v4/monitor/load', { params: { metric, range, aggregate } })
  return res.data
}

export async function fetchV4Clients(params: {
  mode: string
  page?: number
  page_size?: number
  search?: string
}): Promise<ClientPageDTO> {
  const res = await http.get('/v4/clients', { params })
  return res.data
}

export async function fetchV4ClientDetail(ip: string, mac: string): Promise<ClientDetailDTO> {
  const res = await http.get('/v4/clients/detail', { params: { ip, mac } })
  return res.data
}

export async function kickV4OnlineUser(id: string | number, dryRun = false): Promise<Record<string, unknown>> {
  const res = await http.delete(`/v4/users/online/${id}`, { params: { dry_run: dryRun } })
  return res.data
}

export async function fetchV4Analytics(): Promise<AppProtocolAnalyticsDTO> {
  const res = await http.get('/v4/traffic/analytics')
  return res.data
}

export async function fetchV4Catalog(): Promise<ResourceSpec[]> {
  const res = await http.get('/v4/catalog')
  return res.data
}

export async function fetchV4Resource(
  key: string,
  params?: { page?: number; page_size?: number; search?: string; filter?: string; path?: string }
): Promise<ResourceListResult> {
  const res = await http.get(`/v4/resources/${key}`, { params })
  return res.data
}

export async function fetchObjectRefs(
  key: string,
  groupName: string
): Promise<{ group_name: string; rows: Array<Record<string, unknown>>; total: number }> {
  const res = await http.get(`/v4/objects/${key}/refs`, { params: { group_name: groupName } })
  return res.data
}

export async function fetchWGPeers(
  wgID: string | number
): Promise<{ rows: Array<Record<string, unknown>>; total: number }> {
  const res = await http.get(`/v4/wireguard/${wgID}/peers`)
  return res.data
}

export async function generateWGKeypair(): Promise<{ private_key: string; public_key: string }> {
  const res = await http.get('/v4/wireguard/keypair')
  return res.data
}

export async function createWGPeer(
  wgID: string | number,
  body: Record<string, unknown>,
  dryRun = false
): Promise<Record<string, unknown>> {
  const res = await http.post(`/v4/wireguard/${wgID}/peers`, body, { params: { dry_run: dryRun } })
  return res.data
}

export async function updateWGPeer(
  wgID: string | number,
  peerID: string | number,
  body: Record<string, unknown>,
  dryRun = false
): Promise<Record<string, unknown>> {
  const res = await http.put(`/v4/wireguard/${wgID}/peers/${peerID}`, body, { params: { dry_run: dryRun } })
  return res.data
}

export async function toggleWGPeer(
  wgID: string | number,
  peerID: string | number,
  enabled: 'yes' | 'no',
  dryRun = false
): Promise<Record<string, unknown>> {
  const res = await http.patch(
    `/v4/wireguard/${wgID}/peers/${peerID}/toggle`,
    { enabled },
    { params: { dry_run: dryRun } }
  )
  return res.data
}

export async function deleteWGPeer(
  wgID: string | number,
  peerID: string | number,
  dryRun = false
): Promise<Record<string, unknown>> {
  const res = await http.delete(`/v4/wireguard/${wgID}/peers/${peerID}`, { params: { dry_run: dryRun } })
  return res.data
}

export async function createV4Resource(
  key: string,
  body: Record<string, unknown>,
  dryRun = false
): Promise<Record<string, unknown>> {
  const res = await http.post(`/v4/resources/${key}`, body, { params: { dry_run: dryRun } })
  return res.data
}

export async function updateV4Resource(
  key: string,
  id: string | number | null,
  body: Record<string, unknown>,
  dryRun = false
): Promise<Record<string, unknown>> {
  const url = id !== null && id !== '' ? `/v4/resources/${key}/${id}` : `/v4/resources/${key}`
  const res = await http.put(url, body, { params: { dry_run: dryRun } })
  return res.data
}

export async function toggleV4Resource(
  key: string,
  id: string | number,
  enabled: 'yes' | 'no',
  dryRun = false
): Promise<Record<string, unknown>> {
  const res = await http.patch(
    `/v4/resources/${key}/${id}/toggle`,
    { enabled },
    { params: { dry_run: dryRun } }
  )
  return res.data
}

export async function deleteV4Resource(
  key: string,
  id?: string | number,
  dryRun = false,
  srcfile?: string
): Promise<Record<string, unknown>> {
  const url = id !== undefined && id !== '' ? `/v4/resources/${key}/${id}` : `/v4/resources/${key}`
  const params: Record<string, unknown> = { dry_run: dryRun }
  if (srcfile) params.srcfile = srcfile
  const res = await http.delete(url, { params })
  return res.data
}

export async function executeV4Action(action: string, dryRun = false): Promise<Record<string, unknown>> {
  const res = await http.post(`/v4/actions/${action}`, {}, { params: { dry_run: dryRun } })
  return res.data
}

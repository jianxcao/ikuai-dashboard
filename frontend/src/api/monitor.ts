import axios from 'axios'

export interface IKuaiSummary {
  upload_speed: number
  download_speed: number
  total_connections: number
  online_users: number
}

export interface IKuaiWanStatus {
  name: string
  ip: string
  proto: string
  status: string
  comment: string
}

export interface IKuaiTraffic {
  name: string
  ip: string
  upload_speed: number
  download_speed: number
  total_up: number
  total_down: number
  connections: number
  comment: string
}

export interface IKuaiInterfaceData {
  summary: IKuaiSummary
  wan_status: IKuaiWanStatus[]
  traffic_details: IKuaiTraffic[]
}

export interface NetworkMapNode {
  id: string
  name: string
  type: 'router' | 'wan' | 'lan' | 'device' | string
  ip: string
  category: number
}

export interface NetworkMapLink {
  source: string
  target: string
}

export interface NetworkMapData {
  nodes: NetworkMapNode[]
  links: NetworkMapLink[]
}

export interface ClientDTO {
  mac: string
  hostname: string
  ips: string[]
  upload_speed: number
  download_speed: number
  total_up: number
  total_down: number
  connections: number
  comment: string
  client_type: string
  uptime: string
}

export interface PortMapping {
  name: string
  ext_port: string
  int_ip: string
  int_port: string
  protocol: string
}

export interface SecurityHubData {
  high_risk_ports: PortMapping[]
  abnormal_devices: ClientDTO[]
}

export interface RouteRule {
  type: string
  interface: string
  target: string
  enabled: boolean
}

export interface MultiWanData {
  wan_status: IKuaiWanStatus[]
  routes: RouteRule[]
}

export interface MonitorInsightsData {
  summary: IKuaiSummary
  top_clients: ClientDTO[]
  top_interfaces: IKuaiTraffic[]
  abnormal_clients: ClientDTO[]
  high_risk_mappings: PortMapping[]
}

export interface CommonResourceDefinition {
  name: string
  label: string
  group: string
  v3_name: string
  v4_path: string
  writable: boolean
  available: boolean
  methods: string[]
}

export interface CommonResourceData {
  resource: CommonResourceDefinition
  rows: Record<string, unknown>[]
  error?: string
}

export interface CommonResourceMutation {
  resource: string
  action: string
  result: Record<string, unknown>
}

export interface AppServerConfig {
  port: string
  static_dir: string
  token_enabled: boolean
}

export interface RouterConfig {
  id: string
  name: string
  url: string
  version: 'v3' | 'v4'
  username: string
  password?: string
  token?: string
  mock: boolean
  insecure_skip_verify: boolean
}

export interface PublicAppConfig {
  server: AppServerConfig
  active_router_id: string
  routers: RouterConfig[]
}

export interface MonitorServiceStatus {
  mode: 'mock' | 'real' | 'unconfigured'
  router_id: string
  router_name: string
  version?: 'v3' | 'v4' | ''
  error?: string
}

export interface RouterConfigPayload {
  config: PublicAppConfig
  status: MonitorServiceStatus
}

interface RouterConfigEnvelope {
  code: number
  data: PublicAppConfig
  status: MonitorServiceStatus
}

const API_BASE_URL = '/api/v1'

const http = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
})

// ─── Dashboard 访问 Token 工具函数 ───────────────────────────────────────

const DASHBOARD_TOKEN_KEY = 'dashboard_access_token'

export function getLocalToken(): string {
  return localStorage.getItem(DASHBOARD_TOKEN_KEY) || ''
}

export function setLocalToken(token: string): void {
  if (token) {
    localStorage.setItem(DASHBOARD_TOKEN_KEY, token)
  } else {
    localStorage.removeItem(DASHBOARD_TOKEN_KEY)
  }
}

// 统一响应拦截：直接返回 data 字段
http.interceptors.response.use(
  (res) => res.data,
  (err) => {
    // 401 时派发全局未授权事件，供登录状态管理者响应
    if (err.response?.status === 401) {
      window.dispatchEvent(new CustomEvent('auth:unauthorized'))
    }
    console.error('[API Error]', err.message)
    return Promise.reject(err)
  }
)

// 请求拦截器：自动携带 Dashboard 访问 Token
http.interceptors.request.use((config) => {
  const token = getLocalToken()
  if (token) {
    config.headers = config.headers ?? {}
    config.headers['Authorization'] = `Bearer ${token}`
  }
  return config
})

export async function fetchInterfaceData(): Promise<IKuaiInterfaceData> {
  const res = await http.get('/monitor/interface')
  return res.data
}

export async function fetchLanClients(search: string = ''): Promise<ClientDTO[]> {
  const res = await http.get('/monitor/lan', { params: { search } })
  return res.data
}

export async function fetchNetworkMap(): Promise<NetworkMapData> {
  const res = await http.get('/monitor/network-map')
  return res.data
}

export async function fetchSecurityHub(): Promise<SecurityHubData> {
  const res = await http.get('/monitor/security-hub')
  return res.data
}

export async function fetchMultiWan(): Promise<MultiWanData> {
  const res = await http.get('/monitor/multi-wan')
  return res.data
}

export async function fetchMonitorInsights(): Promise<MonitorInsightsData> {
  const res = await http.get('/monitor/insights')
  return res.data
}

export async function fetchCommonResources(): Promise<CommonResourceDefinition[]> {
  const res = await http.get('/router/resources')
  return res.data
}

export async function fetchCommonResource(name: string): Promise<CommonResourceData> {
  const res = await http.get(`/router/resources/${name}`)
  return res.data
}

export async function createCommonResource(
  name: string,
  payload: Record<string, unknown>
): Promise<CommonResourceMutation> {
  const res = await http.post(`/router/resources/${name}`, payload)
  return res.data
}

export async function updateCommonResource(
  name: string,
  id: string | number,
  payload: Record<string, unknown>
): Promise<CommonResourceMutation> {
  const res = await http.put(`/router/resources/${name}/${id}`, payload)
  return res.data
}

export async function deleteCommonResource(
  name: string,
  id: string | number
): Promise<CommonResourceMutation> {
  const res = await http.delete(`/router/resources/${name}/${id}`)
  return res.data
}

export async function fetchRouterConfig(): Promise<RouterConfigPayload> {
  const res = await http.get<unknown, RouterConfigEnvelope>('/config/routers')
  return { config: res.data, status: res.status }
}

export async function switchActiveRouter(id: string): Promise<RouterConfigPayload> {
  const res = await http.put<unknown, RouterConfigEnvelope>('/config/active-router', { id })
  return { config: res.data, status: res.status }
}

export async function saveRouterConfig(config: PublicAppConfig): Promise<RouterConfigPayload> {
  const res = await http.put<unknown, RouterConfigEnvelope>('/config/routers', config)
  return { config: res.data, status: res.status }
}

// ─── 用户登录认证 API ─────────────────────────────────────────────────

export interface AuthStatus {
  auth_enabled: boolean
}

export interface LoginResult {
  token: string
  expires_in: number
}

/**
 * 查询后端是否启用登录认证（公开接口，不需要 Token）。
 */
export async function fetchAuthStatus(): Promise<AuthStatus> {
  const res = await http.get<unknown, { code: number; data: AuthStatus }>('/auth/status')
  return res.data
}

/**
 * 使用用户名密码登录。
 */
export async function login(username: string, password: string): Promise<LoginResult> {
  const res = await http.post<unknown, { code: number; data: LoginResult }>('/auth/login', {
    username,
    password,
  })
  return res.data
}

/**
 * 登出（尞层清除 localStorage，后端仅做日志）。
 */
export async function logout(): Promise<void> {
  await http.post('/auth/logout').catch(() => {}) // 登出不需要处理错误
}

export interface TokenStatus {
  token_enabled: boolean
}

export interface SaveTokenResult {
  token: string
  token_enabled: boolean
}

/**
 * 查询 Dashboard 访问 Token 状态（不受 Token 保护，公开接口）。
 */
export async function fetchTokenStatus(): Promise<TokenStatus> {
  // 直接使用原始 axios ，避免 http 实例的拦截器层干扰
  const res = await http.get<unknown, { code: number; data: TokenStatus }>('/config/token')
  return res.data
}

/**
 * 保存 Dashboard 访问 Token。
 * @param token - 指定 Token。传入空字符串清除 Token。
 * @param generate - 传 true 则让后端自动生成随机 Token。
 */
export async function saveToken(
  token: string,
  generate = false
): Promise<SaveTokenResult> {
  const res = await http.put<unknown, { code: number; data: SaveTokenResult }>('/config/token', {
    token,
    generate,
  })
  return res.data
}

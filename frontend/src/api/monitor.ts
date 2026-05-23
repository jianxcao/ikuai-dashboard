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

export interface AppServerConfig {
  port: string
  static_dir: string
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

// 统一响应拦截：直接返回 data 字段
http.interceptors.response.use(
  (res) => res.data,
  (err) => {
    console.error('[API Error]', err.message)
    return Promise.reject(err)
  }
)

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

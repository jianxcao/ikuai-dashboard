import axios from 'axios'

// ─── Shared HTTP Instance & Dashboard Auth / Token Helpers ──────────────────

export const http = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

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

http.interceptors.response.use(
  (res) => res.data,
  (err) => {
    if (err.response?.status === 401) {
      window.dispatchEvent(new CustomEvent('auth:unauthorized'))
    }
    return Promise.reject(err)
  }
)

http.interceptors.request.use((config) => {
  const token = getLocalToken()
  if (token) {
    config.headers = config.headers ?? {}
    config.headers['Authorization'] = `Bearer ${token}`
  }
  return config
})

// ─── Auth & Token API ───────────────────────────────────────────────────────

export interface AuthStatus {
  auth_enabled: boolean
  read_only?: boolean
}

export interface LoginResult {
  token: string
  expires_in: number
}

export async function fetchAuthStatus(): Promise<AuthStatus> {
  const res = await http.get<unknown, { code: number; data: AuthStatus }>('/auth/status')
  return res.data
}

export async function login(username: string, password: string): Promise<LoginResult> {
  const res = await http.post<unknown, { code: number; data: LoginResult }>('/auth/login', {
    username,
    password,
  })
  return res.data
}

export async function logout(): Promise<void> {
  await http.post('/auth/logout').catch(() => {})
}

export interface TokenStatus {
  token_enabled: boolean
}

export interface SaveTokenResult {
  token: string
  token_enabled: boolean
}

export async function fetchTokenStatus(): Promise<TokenStatus> {
  const res = await http.get<unknown, { code: number; data: TokenStatus }>('/config/token')
  return res.data
}

export async function saveToken(token: string, generate = false): Promise<SaveTokenResult> {
  const res = await http.put<unknown, { code: number; data: SaveTokenResult }>('/config/token', {
    token,
    generate,
  })
  return res.data
}

// ─── Preserved Compatibility Interfaces ─────────────────────────────────────

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

export interface NetworkMapNode {
  id: string
  name: string
  type: 'router' | 'wan' | 'lan' | 'device' | string
  ip: string
  category: number
  status?: string
}

export interface NetworkMapLink {
  source: string
  target: string
}

export interface NetworkMapData {
  nodes: NetworkMapNode[]
  links: NetworkMapLink[]
}

export interface PortMapping {
  name: string
  ext_port: string
  int_ip: string
  int_port: string
  protocol: string
  enabled?: boolean
}

export interface SecurityHubData {
  high_risk_ports: PortMapping[]
  abnormal_devices: ClientDTO[]
  acl_count: number
  peerconn_count: number
  url_black_count: number
}

export async function fetchNetworkMap(): Promise<NetworkMapData> {
  const res = await http.get('/monitor/network-map')
  return res.data
}

export async function fetchSecurityHub(): Promise<SecurityHubData> {
  const res = await http.get('/monitor/security-hub')
  return res.data
}

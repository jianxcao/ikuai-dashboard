import { ref, readonly } from 'vue'
import {
  fetchAuthStatus,
  login as apiLogin,
  logout as apiLogout,
  getLocalToken,
  setLocalToken,
} from '@/api/monitor'

// ─── 常量 ────────────────────────────────────────────────────────────────────

const SESSION_TOKEN_KEY = 'dashboard_session_token'
const SESSION_EXPIRY_KEY = 'dashboard_session_expiry'

// ─── 模块级响应式状态（单例，跨组件共享）──────────────────────────────────────

const _authEnabled = ref(false)
const _isAuthenticated = ref(false)
const _loading = ref(true) // 初始化中

// ─── localStorage 工具 ───────────────────────────────────────────────────────

function saveSession(token: string, expiresIn: number) {
  localStorage.setItem(SESSION_TOKEN_KEY, token)
  const expiry = Date.now() + expiresIn * 1000
  localStorage.setItem(SESSION_EXPIRY_KEY, String(expiry))
  // 同步到 monitor.ts 的 axios 拦截器（Bearer token）
  setLocalToken(token)
}

function clearSession() {
  localStorage.removeItem(SESSION_TOKEN_KEY)
  localStorage.removeItem(SESSION_EXPIRY_KEY)
  setLocalToken('')
}

function isSessionValid(): boolean {
  const token = localStorage.getItem(SESSION_TOKEN_KEY)
  const expiry = localStorage.getItem(SESSION_EXPIRY_KEY)
  if (!token || !expiry) return false
  return Date.now() < parseInt(expiry, 10)
}

// 启动时恢复会话（如果 localStorage 中有未过期的 token）
function restoreSession() {
  if (isSessionValid()) {
    const token = localStorage.getItem(SESSION_TOKEN_KEY)!
    setLocalToken(token) // 确保 axios 拦截器中有 token
    return true
  }
  clearSession()
  return false
}

// ─── 初始化（应用启动时调用一次）────────────────────────────────────────────

let initialized = false

async function init() {
  if (initialized) return
  initialized = true

  try {
    const status = await fetchAuthStatus()
    _authEnabled.value = status.auth_enabled

    if (!status.auth_enabled) {
      // 认证未启用，直接标记为已认证
      _isAuthenticated.value = true
    } else {
      // 尝试从 localStorage 恢复会话
      _isAuthenticated.value = restoreSession()
    }
  } catch {
    // 网络错误等，假设认证已启用，让用户登录
    _authEnabled.value = true
    _isAuthenticated.value = false
  } finally {
    _loading.value = false
  }
}

// ─── 监听 401 全局事件（任何 API 请求 401 → 自动登出）────────────────────────

window.addEventListener('auth:unauthorized', () => {
  if (_authEnabled.value && _isAuthenticated.value) {
    clearSession()
    _isAuthenticated.value = false
  }
})

// ─── Composable 导出 ─────────────────────────────────────────────────────────

export function useAuth() {
  // 触发初始化（幂等）
  void init()

  /**
   * 登录：验证用户名密码 → 保存 Session Token → 更新认证状态
   * @throws Error 登录失败时抛出（由调用方展示错误信息）
   */
  async function login(username: string, password: string) {
    const result = await apiLogin(username, password)
    saveSession(result.token, result.expires_in)
    _isAuthenticated.value = true
  }

  /**
   * 登出：清除 Session → 触发 UI 跳转到登录页
   */
  async function logout() {
    await apiLogout()
    clearSession()
    _isAuthenticated.value = false
  }

  return {
    authEnabled: readonly(_authEnabled),
    isAuthenticated: readonly(_isAuthenticated),
    loading: readonly(_loading),
    login,
    logout,
  }
}

import { computed, readonly, ref } from 'vue'
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
const SAFETY_LOCK_KEY = 'ikuai_v4_safety_lock'

// ─── 模块级响应式状态（单例，跨组件共享）──────────────────────────────────────

const _authEnabled = ref(false)
const _isAuthenticated = ref(false)
const _loading = ref(true)
const _serverReadOnly = ref(false)
const _uiSafetyLock = ref(
  typeof window !== 'undefined' && window.localStorage.getItem(SAFETY_LOCK_KEY) === 'true'
)

// ─── localStorage 工具 ───────────────────────────────────────────────────────

function saveSession(token: string, expiresIn: number) {
  localStorage.setItem(SESSION_TOKEN_KEY, token)
  const expiry = Date.now() + expiresIn * 1000
  localStorage.setItem(SESSION_EXPIRY_KEY, String(expiry))
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

function restoreSession() {
  if (isSessionValid()) {
    const token = localStorage.getItem(SESSION_TOKEN_KEY)!
    setLocalToken(token)
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
    _serverReadOnly.value = Boolean(status.read_only)

    if (!status.auth_enabled) {
      _isAuthenticated.value = true
    } else {
      _isAuthenticated.value = restoreSession()
    }
  } catch {
    _authEnabled.value = true
    _isAuthenticated.value = false
  } finally {
    _loading.value = false
  }
}

window.addEventListener('auth:unauthorized', () => {
  if (_authEnabled.value && _isAuthenticated.value) {
    clearSession()
    _isAuthenticated.value = false
  }
})

// ─── Composable 导出 ─────────────────────────────────────────────────────────

export function useAuth() {
  void init()

  async function login(username: string, password: string) {
    const result = await apiLogin(username, password)
    saveSession(result.token, result.expires_in)
    _isAuthenticated.value = true
  }

  async function logout() {
    await apiLogout()
    clearSession()
    _isAuthenticated.value = false
  }

  function setServerReadOnly(val: boolean) {
    _serverReadOnly.value = val
  }

  function toggleSafetyLock() {
    if (_serverReadOnly.value) return
    _uiSafetyLock.value = !_uiSafetyLock.value
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(SAFETY_LOCK_KEY, String(_uiSafetyLock.value))
    }
  }

  const readOnly = computed(() => _serverReadOnly.value || _uiSafetyLock.value)

  return {
    authEnabled: readonly(_authEnabled),
    isAuthenticated: readonly(_isAuthenticated),
    loading: readonly(_loading),
    serverReadOnly: readonly(_serverReadOnly),
    uiSafetyLock: readonly(_uiSafetyLock),
    readOnly,
    login,
    logout,
    setServerReadOnly,
    toggleSafetyLock,
    getLocalToken,
  }
}

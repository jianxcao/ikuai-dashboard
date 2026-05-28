import { ref, onMounted } from 'vue'
import {
  fetchTokenStatus,
  saveToken,
  getLocalToken,
  setLocalToken,
  type SaveTokenResult,
} from '@/api/monitor'

/**
 * useTokenConfig — 管理 Dashboard 访问 Token 的 Composable。
 *
 * 职责：
 * - 查询后端当前 Token 状态（是否启用）
 * - 生成随机 Token（通过后端 generate=true）
 * - 清除 Token
 * - 保持 localStorage 与后端配置同步
 */
export function useTokenConfig() {
  const tokenEnabled = ref(false)
  const currentToken = ref<string>('')
  const saving = ref(false)
  const loading = ref(true)
  const error = ref<string | null>(null)
  const successMsg = ref<string | null>(null)

  async function loadStatus() {
    loading.value = true
    error.value = null
    try {
      const status = await fetchTokenStatus()
      tokenEnabled.value = status.token_enabled
      // 若本地有 Token，显示脱敏形式（供用户确认）
      currentToken.value = getLocalToken()
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : '获取 Token 状态失败'
    } finally {
      loading.value = false
    }
  }

  async function applyToken(result: SaveTokenResult) {
    tokenEnabled.value = result.token_enabled
    currentToken.value = result.token
    setLocalToken(result.token)
    successMsg.value = result.token ? 'Token 已保存' : 'Token 已清除'
    setTimeout(() => {
      successMsg.value = null
    }, 3000)
  }

  /** 通过后端生成随机 Token 并保存 */
  async function generateToken() {
    saving.value = true
    error.value = null
    try {
      const result = await saveToken('', true)
      await applyToken(result)
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : '生成 Token 失败'
    } finally {
      saving.value = false
    }
  }

  /** 清除 Token（关闭访问保护） */
  async function clearToken() {
    saving.value = true
    error.value = null
    try {
      const result = await saveToken('', false)
      await applyToken(result)
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : '清除 Token 失败'
    } finally {
      saving.value = false
    }
  }

  /** 复制 Token 到剪贴板 */
  async function copyToken(): Promise<boolean> {
    if (!currentToken.value) return false
    try {
      await navigator.clipboard.writeText(currentToken.value)
      return true
    } catch {
      return false
    }
  }

  onMounted(() => {
    void loadStatus()
  })

  return {
    tokenEnabled,
    currentToken,
    saving,
    loading,
    error,
    successMsg,
    loadStatus,
    generateToken,
    clearToken,
    copyToken,
  }
}

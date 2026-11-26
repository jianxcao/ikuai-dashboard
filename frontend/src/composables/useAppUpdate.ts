import { computed, ref } from 'vue'
import { registerSW } from 'virtual:pwa-register'

export function normalizeVersion(raw?: string | null): string {
  const trimmed = (raw ?? '').trim()
  if (!trimmed) return ''
  if (/^\d+\.\d+/.test(trimmed)) return `v${trimmed}`
  if (/^V\d+/.test(trimmed)) return `v${trimmed.slice(1)}`
  return trimmed
}

export function isConcreteVersion(v: string): boolean {
  return Boolean(v && v !== 'dev' && v !== 'v4-dev' && v !== 'latest')
}

const buildVersion = normalizeVersion(import.meta.env.VITE_APP_VERSION) || 'v4-dev'
const serverVersion = ref('')
const swReady = ref(false)
const swNeedRefresh = ref(false)
const httpVersionMismatch = ref(false)
const promptDismissed = ref(false)
const updating = ref(false)
const checking = ref(false)
const lastCheckMessage = ref<string | null>(null)

let initialized = false
let lastCheckedAt = 0
let lastMismatchVersion = ''
let clearMessageTimer: number | null = null
let swRegistration: ServiceWorkerRegistration | undefined
let updateSWFn: ((reloadPage?: boolean) => Promise<void>) | undefined

function hasRegisteredOrActiveSw(): boolean {
  return (
    typeof navigator !== 'undefined' &&
    'serviceWorker' in navigator &&
    Boolean(
      navigator.serviceWorker.controller ||
        swRegistration?.active ||
        swRegistration?.waiting ||
        swRegistration?.installing,
    )
  )
}

const displayVersion = computed(() => {
  if (isConcreteVersion(buildVersion)) return buildVersion
  if (isConcreteVersion(serverVersion.value)) return serverVersion.value
  return buildVersion
})

const remoteVersionLabel = computed(() => {
  if (isConcreteVersion(serverVersion.value) && serverVersion.value !== buildVersion) {
    return serverVersion.value
  }
  return ''
})

const updateAvailable = computed(() => swNeedRefresh.value || httpVersionMismatch.value)
const promptVisible = computed(() => updateAvailable.value && !promptDismissed.value)

async function fetchServerVersion(triggerSwUpdate = true) {
  if (typeof fetch === 'undefined') return
  try {
    const resp = await fetch('/health', {
      cache: 'no-store',
      headers: { Accept: 'application/json' },
    })
    if (!resp.ok) return
    const data = (await resp.json()) as { app_version?: string }
    const normalized = normalizeVersion(data?.app_version)
    if (!normalized) return
    serverVersion.value = normalized

    if (isConcreteVersion(buildVersion) && isConcreteVersion(normalized) && normalized !== buildVersion) {
      if (hasRegisteredOrActiveSw()) {
        if (triggerSwUpdate && !swNeedRefresh.value && !swRegistration?.waiting) {
          await swRegistration?.update()
        }
        if (swRegistration?.waiting) {
          swNeedRefresh.value = true
          promptDismissed.value = false
        }
      } else {
        httpVersionMismatch.value = true
        if (lastMismatchVersion !== normalized) {
          lastMismatchVersion = normalized
          promptDismissed.value = false
        }
      }
    }
  } catch {
    // Ignore transient network errors when probing /health
  }
}

export function initPwaRegistration() {
  if (initialized) return
  initialized = true

  updateSWFn = registerSW({
    immediate: true,
    onNeedRefresh() {
      swNeedRefresh.value = true
      promptDismissed.value = false
      void fetchServerVersion(false)
    },
    onOfflineReady() {
      swReady.value = true
    },
    onRegisteredSW(_swUrl, reg) {
      swRegistration = reg
      swReady.value = Boolean(reg)
      if (reg?.waiting) {
        swNeedRefresh.value = true
        promptDismissed.value = false
      }
    },
  })
}

export async function checkForUpdates(manual = false) {
  const now = Date.now()
  if (!manual && now - lastCheckedAt < 60_000) return
  lastCheckedAt = now
  checking.value = true

  if (clearMessageTimer !== null) {
    window.clearTimeout(clearMessageTimer)
    clearMessageTimer = null
  }

  try {
    await Promise.allSettled([swRegistration?.update(), fetchServerVersion(false)])
    if (swRegistration?.waiting) {
      swNeedRefresh.value = true
      promptDismissed.value = false
    }

    if (manual) {
      if (updateAvailable.value) {
        promptDismissed.value = false
        lastCheckMessage.value = '发现新版本界面，可立即更新'
      } else {
        lastCheckMessage.value = `当前已是最新版本 (${displayVersion.value})`
        clearMessageTimer = window.setTimeout(() => {
          lastCheckMessage.value = null
          clearMessageTimer = null
        }, 4000)
      }
    }
  } finally {
    checking.value = false
  }
}

export async function applyUpdate() {
  updating.value = true
  let reloading = false
  const reloadOnce = () => {
    if (reloading) return
    reloading = true
    window.location.reload()
  }

  if (typeof navigator !== 'undefined' && 'serviceWorker' in navigator) {
    navigator.serviceWorker.addEventListener('controllerchange', reloadOnce, { once: true })
  }

  try {
    if (swNeedRefresh.value && updateSWFn) {
      window.setTimeout(reloadOnce, 1800)
      await updateSWFn(true)
      const reg = swRegistration ?? (await navigator.serviceWorker?.getRegistration?.())
      if (reg?.waiting) {
        reg.waiting.postMessage({ type: 'SKIP_WAITING' })
      }
      return
    }
    reloadOnce()
  } catch {
    reloadOnce()
  }
}

export function dismissPrompt() {
  promptDismissed.value = true
}

export function reopenPrompt() {
  promptDismissed.value = false
}

export function useAppUpdate() {
  return {
    buildVersion,
    serverVersion,
    displayVersion,
    remoteVersionLabel,
    swReady,
    swNeedRefresh,
    httpVersionMismatch,
    updateAvailable,
    promptVisible,
    updating,
    checking,
    lastCheckMessage,
    initPwaRegistration,
    checkForUpdates,
    applyUpdate,
    dismissPrompt,
    reopenPrompt,
  }
}

import { computed, onMounted, ref } from 'vue'
import {
  fetchRouterConfig,
  saveRouterConfig,
  switchActiveRouter,
  type MonitorServiceStatus,
  type PublicAppConfig,
  type RouterConfig,
} from '@/api/monitor'

function createBlankRouter(index: number): RouterConfig {
  return {
    id: `router-${index}`,
    name: '新爱快服务器',
    url: 'https://192.168.1.1',
    version: 'v3',
    username: 'admin',
    password: '',
    token: '',
    mock: false,
    insecure_skip_verify: true,
  }
}

function normalizeConfig(payload: PublicAppConfig): PublicAppConfig {
  return {
    ...payload,
    routers: (payload.routers || []).map((router) => ({
      ...router,
      version: router.version || 'v3',
      username: router.username || '',
      password: router.password || '',
      token: router.token || '',
    })),
  }
}

const loading = ref(true)
const saving = ref(false)
const switching = ref<string | null>(null)
const error = ref<string | null>(null)
const saved = ref(false)
const config = ref<PublicAppConfig>({
  server: { port: '8080', static_dir: 'frontend/dist', token_enabled: false },
  active_router_id: '',
  routers: [],
})
const status = ref<MonitorServiceStatus | null>(null)
const routerRevision = ref(0)
let loaded = false

const activeRouter = computed(() => {
  return config.value.routers.find((router) => router.id === config.value.active_router_id) ?? null
})

async function load() {
  loading.value = true
  error.value = null
  try {
    const payload = await fetchRouterConfig()
    config.value = normalizeConfig(payload.config)
    status.value = payload.status
    loaded = true
  } catch (e: any) {
    error.value = e.message || String(e)
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  saved.value = false
  error.value = null
  try {
    const payload = await saveRouterConfig(config.value)
    config.value = normalizeConfig(payload.config)
    status.value = payload.status
    routerRevision.value += 1
    saved.value = true
  } catch (e: any) {
    error.value = e.response?.data?.message || e.message || String(e)
  } finally {
    saving.value = false
  }
}

async function activate(id: string) {
  switching.value = id
  error.value = null
  try {
    const payload = await switchActiveRouter(id)
    config.value = normalizeConfig(payload.config)
    status.value = payload.status
    routerRevision.value += 1
  } catch (e: any) {
    error.value = e.response?.data?.message || e.message || String(e)
  } finally {
    switching.value = null
  }
}

function addRouter() {
  const next = createBlankRouter(config.value.routers.length + 1)
  config.value.routers.push(next)
  if (!config.value.active_router_id) {
    config.value.active_router_id = next.id
  }
}

function removeRouter(id: string) {
  config.value.routers = config.value.routers.filter((router) => router.id !== id)
  if (config.value.active_router_id === id) {
    config.value.active_router_id = config.value.routers[0]?.id ?? ''
  }
}

export function useRouterConfig() {
  onMounted(() => {
    if (!loaded) {
      load()
    }
  })

  return {
    loading,
    saving,
    switching,
    error,
    saved,
    config,
    status,
    activeRouter,
    routerRevision,
    load,
    save,
    activate,
    addRouter,
    removeRouter,
  }
}

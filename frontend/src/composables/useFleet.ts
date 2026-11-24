import { computed, readonly, ref } from 'vue'
import {
  activateRouterApi,
  fetchRouters,
  probeRouters,
  saveRoutersApi,
  testRouterConnection,
  type EditableRouterConfig,
  type MonitorServiceStatus,
  type PublicAppConfig,
  type RouterHealthDTO,
  type RouterTestPayload,
} from '@/api/v4'

const loading = ref(true)
const probing = ref(false)
const saving = ref(false)
const switchingId = ref<string | null>(null)
const error = ref<string | null>(null)
const routerEpoch = ref(0)

const config = ref<PublicAppConfig>({
  server: { port: '8080', static_dir: 'frontend/dist', token_enabled: false },
  active_router_id: '',
  routers: [],
})

const status = ref<MonitorServiceStatus>({
  mode: 'unconfigured',
  version: 'v4',
  router_id: '',
  router_name: '',
  router_count: 0,
})

const probeItems = ref<RouterHealthDTO[]>([])
let loadedOnce = false

async function loadFleet(alsoProbe = true) {
  loading.value = true
  error.value = null
  try {
    const payload = await fetchRouters()
    config.value = payload.config
    status.value = payload.status
    loadedOnce = true
    if (alsoProbe && payload.config.routers.length > 0) {
      void refreshProbe()
    }
  } catch (e: unknown) {
    error.value = extractErrorMessage(e)
  } finally {
    loading.value = false
  }
}

async function refreshProbe() {
  probing.value = true
  try {
    probeItems.value = await probeRouters()
  } catch {
    // Keep last probe snapshot if one probe attempt fails
  } finally {
    probing.value = false
  }
}

async function activateRouter(id: string) {
  if (!id || id === config.value.active_router_id) return
  switchingId.value = id
  error.value = null
  try {
    const payload = await activateRouterApi(id)
    config.value = payload.config
    status.value = payload.status
    routerEpoch.value += 1
    void refreshProbe()
  } catch (e: unknown) {
    error.value = extractErrorMessage(e)
  } finally {
    switchingId.value = null
  }
}

async function saveRouters(activeRouterId: string, routers: EditableRouterConfig[]) {
  saving.value = true
  error.value = null
  try {
    const payload = await saveRoutersApi({
      active_router_id: activeRouterId,
      routers,
    })
    config.value = payload.config
    status.value = payload.status
    routerEpoch.value += 1
    await refreshProbe()
    return true
  } catch (e: unknown) {
    error.value = extractErrorMessage(e)
    return false
  } finally {
    saving.value = false
  }
}

async function testRouter(payload: RouterTestPayload): Promise<RouterHealthDTO> {
  return testRouterConnection(payload)
}

function extractErrorMessage(e: unknown): string {
  if (e && typeof e === 'object' && 'response' in e) {
    const resp = (e as { response?: { data?: { message?: string } } }).response
    if (resp?.data?.message) return resp.data.message
  }
  return e instanceof Error ? e.message : String(e)
}

export function useFleet() {
  if (!loadedOnce) {
    loadedOnce = true
    void loadFleet(true)
  }

  const activeRouter = computed(
    () => config.value.routers.find((r) => r.id === config.value.active_router_id) ?? null
  )
  const activeProbe = computed(
    () => probeItems.value.find((p) => p.id === config.value.active_router_id) ?? null
  )
  const isUnconfigured = computed(() => !loading.value && config.value.routers.length === 0)

  return {
    loading: readonly(loading),
    probing: readonly(probing),
    saving: readonly(saving),
    switchingId: readonly(switchingId),
    error,
    config,
    status,
    probeItems,
    activeRouter,
    activeProbe,
    isUnconfigured,
    routerEpoch: readonly(routerEpoch),
    loadFleet,
    refreshProbe,
    activateRouter,
    saveRouters,
    testRouter,
  }
}

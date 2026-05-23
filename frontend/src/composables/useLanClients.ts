import { ref, computed, onActivated, onDeactivated, onMounted, onUnmounted, watch } from 'vue'
import { fetchLanClients, type ClientDTO } from '@/api/monitor'
import { getIPv4 } from '@/composables/useFormatters'
import { useRouterConfig } from '@/composables/useRouterConfig'

function ipToNumber(ip: string): number {
  if (!ip || ip === '—') return 0
  const parts = ip.split('.')
  if (parts.length !== 4) return 0
  // Handle unsigned 32-bit int properly in JS bitwise operations
  return parts.reduce((acc, part) => (acc * 256) + parseInt(part, 10), 0)
}

export function useLanClients() {
  const { activeRouter, routerRevision } = useRouterConfig()
  const loading = ref<boolean>(true)
  const error = ref<string | null>(null)
  const rawClients = ref<ClientDTO[]>([])
  const search = ref<string>('')
  
  // 排序规则: 'default' | 'upload_desc' | 'download_desc' | 'connections_desc'
  const sortBy = ref<string>(localStorage.getItem('ikuai_lan_sort') || 'default')

  let timer: ReturnType<typeof setInterval> | null = null
  let debounceTimer: ReturnType<typeof setTimeout> | null = null

  // 监听排序规则变化并持久化
  watch(sortBy, (newVal) => {
    localStorage.setItem('ikuai_lan_sort', newVal)
  })

  // 计算属性：对列表进行搜索和排序
  const clients = computed<ClientDTO[]>(() => {
    const list = [...rawClients.value]
    
    if (sortBy.value === 'upload_desc') {
      list.sort((a, b) => (b.upload_speed || 0) - (a.upload_speed || 0))
    } else if (sortBy.value === 'upload_asc') {
      list.sort((a, b) => (a.upload_speed || 0) - (b.upload_speed || 0))
    } else if (sortBy.value === 'download_desc') {
      list.sort((a, b) => (b.download_speed || 0) - (a.download_speed || 0))
    } else if (sortBy.value === 'download_asc') {
      list.sort((a, b) => (a.download_speed || 0) - (b.download_speed || 0))
    } else if (sortBy.value === 'connections_desc') {
      list.sort((a, b) => (b.connections || 0) - (a.connections || 0))
    } else if (sortBy.value === 'connections_asc') {
      list.sort((a, b) => (a.connections || 0) - (b.connections || 0))
    } else if (sortBy.value === 'total_desc') {
      list.sort((a, b) => ((b.total_up || 0) + (b.total_down || 0)) - ((a.total_up || 0) + (a.total_down || 0)))
    } else if (sortBy.value === 'total_asc') {
      list.sort((a, b) => ((a.total_up || 0) + (a.total_down || 0)) - ((b.total_up || 0) + (b.total_down || 0)))
    } else if (sortBy.value === 'ip_desc') {
      list.sort((a, b) => ipToNumber(getIPv4(b.ips)) - ipToNumber(getIPv4(a.ips)))
    } else if (sortBy.value === 'ip_asc') {
      list.sort((a, b) => ipToNumber(getIPv4(a.ips)) - ipToNumber(getIPv4(b.ips)))
    }
    return list
  })

  async function fetchData() {
    if (!activeRouter.value) {
      rawClients.value = []
      error.value = null
      loading.value = false
      return
    }
    try {
      const data = await fetchLanClients(search.value)
      rawClients.value = data || []
      error.value = null
    } catch (e: any) {
      error.value = e.message || String(e)
    } finally {
      loading.value = false
    }
  }

  function startPolling() {
    if (timer || !activeRouter.value) {
      if (!activeRouter.value) {
        rawClients.value = []
        error.value = null
        loading.value = false
      }
      return
    }
    void fetchData()
    timer = setInterval(fetchData, 5000)
  }

  function stopPolling() {
    if (timer) {
      clearInterval(timer)
      timer = null
    }
  }

  function onSearch(val: string) {
    search.value = val
    if (debounceTimer) clearTimeout(debounceTimer)
    debounceTimer = setTimeout(() => {
      loading.value = true
      fetchData()
    }, 300)
  }

  onMounted(startPolling)
  onActivated(startPolling)
  onDeactivated(stopPolling)
  onUnmounted(() => {
    stopPolling()
    if (debounceTimer) clearTimeout(debounceTimer)
  })
  watch(activeRouter, (router) => {
    if (!router) {
      stopPolling()
      rawClients.value = []
      error.value = null
      loading.value = false
      return
    }
    loading.value = true
    startPolling()
  })
  watch(routerRevision, () => {
    if (!activeRouter.value) return
    rawClients.value = []
    loading.value = true
    void fetchData()
  })

  return {
    loading,
    error,
    clients,
    search,
    sortBy,
    onSearch,
  }
}

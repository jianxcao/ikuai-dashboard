import { ref, onActivated, onDeactivated, onMounted, onUnmounted, watch } from 'vue'
import { fetchInterfaceData, type IKuaiSummary, type IKuaiWanStatus, type IKuaiTraffic } from '@/api/monitor'
import { useRouterConfig } from '@/composables/useRouterConfig'

const CHART_MAX_POINTS = 60 // 60 个数据点 = 3 分钟滚动窗口

export function useMonitor() {
  const { activeRouter, routerRevision } = useRouterConfig()
  const loading = ref<boolean>(true)
  const error = ref<string | null>(null)
  const summary = ref<IKuaiSummary>({ upload_speed: 0, download_speed: 0, total_connections: 0, online_users: 0 })
  const wanStatus = ref<IKuaiWanStatus[]>([])
  const trafficDetails = ref<IKuaiTraffic[]>([])

  // ECharts 数据：最近 60 个时间点的速率历史
  const chartLabels = ref<string[]>([])
  const chartUpload = ref<number[]>([])
  const chartDownload = ref<number[]>([])

  let timer: ReturnType<typeof setInterval> | null = null

  function resetData() {
    summary.value = { upload_speed: 0, download_speed: 0, total_connections: 0, online_users: 0 }
    wanStatus.value = []
    trafficDetails.value = []
    chartLabels.value = []
    chartUpload.value = []
    chartDownload.value = []
  }

  function addChartPoint(up: number, down: number) {
    const now = new Date()
    const label = `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}:${String(now.getSeconds()).padStart(2, '0')}`

    chartLabels.value.push(label)
    chartUpload.value.push(up)
    chartDownload.value.push(down)

    if (chartLabels.value.length > CHART_MAX_POINTS) {
      chartLabels.value.shift()
      chartUpload.value.shift()
      chartDownload.value.shift()
    }
  }

  async function fetchData() {
    if (!activeRouter.value) {
      resetData()
      loading.value = false
      error.value = null
      return
    }
    try {
      const data = await fetchInterfaceData()
      summary.value = data.summary || summary.value
      wanStatus.value = data.wan_status || []
      trafficDetails.value = data.traffic_details || []
      addChartPoint(data.summary?.upload_speed || 0, data.summary?.download_speed || 0)
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
        resetData()
        loading.value = false
        error.value = null
      }
      return
    }
    void fetchData()
    timer = setInterval(fetchData, 3000)
  }

  function stopPolling() {
    if (timer) {
      clearInterval(timer)
      timer = null
    }
  }

  onMounted(startPolling)
  onActivated(startPolling)
  onDeactivated(stopPolling)
  onUnmounted(stopPolling)
  watch(activeRouter, (router) => {
    if (!router) {
      stopPolling()
      resetData()
      loading.value = false
      error.value = null
      return
    }
    loading.value = true
    startPolling()
  })
  watch(routerRevision, () => {
    if (!activeRouter.value) return
    resetData()
    loading.value = true
    void fetchData()
  })

  return {
    loading,
    error,
    summary,
    wanStatus,
    trafficDetails,
    chartLabels,
    chartUpload,
    chartDownload,
  }
}

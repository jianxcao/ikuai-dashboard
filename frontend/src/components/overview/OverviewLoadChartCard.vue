<template>
  <OverviewSectionShell
    title="系统资源历史曲线"
    :layoutLocked="layoutLocked"
    :canMovePrev="canMovePrev"
    :canMoveNext="canMoveNext"
    @move="emit('move', $event)"
  >
    <template #icon>
      <TrendingUp :size="16" />
    </template>
    <template #actions>
      <div class="pill-group">
        <select v-model="loadMetric" class="mini-select" aria-label="监控指标">
          <option value="cpu">CPU %</option>
          <option value="memory">内存 %</option>
          <option value="connections">连接数</option>
          <option value="terminals">终端数</option>
          <option value="network">网络负载</option>
        </select>
        <select v-model="loadRange" class="mini-select" aria-label="时间范围">
          <option value="hour">1 小时</option>
          <option value="day">24 小时</option>
          <option value="week">7 天</option>
        </select>
      </div>
    </template>

    <div class="chart-stage">
      <div :ref="setChartEl" class="chart-box" />
      <div v-if="loading && !loadData?.series?.length" class="chart-overlay font-mono">
        正在读取历史曲线...
      </div>
      <div v-else-if="!loading && loadData && loadData.series.length === 0" class="chart-overlay font-mono">
        当前时间范围暂无历史采样数据
      </div>
    </div>
  </OverviewSectionShell>
</template>

<script setup lang="ts">
import { nextTick, onActivated, onMounted, onUnmounted, ref, watch } from 'vue'
import { TrendingUp } from 'lucide-vue-next'
import { echarts } from '@/lib/echarts'
import { fetchV4LoadHistory, type LoadHistoryDTO } from '@/api/v4'
import { useChartTheme } from '@/composables/useChartTheme'
import OverviewSectionShell from './OverviewSectionShell.vue'

const props = defineProps<{
  routerKey: string
  pollTick: number
  layoutLocked: boolean
  canMovePrev: boolean
  canMoveNext: boolean
  orderIndex: number
}>()

const emit = defineEmits<{
  (e: 'move', delta: -1 | 1): void
}>()

const { palette } = useChartTheme()

const loadMetric = ref('cpu')
const loadRange = ref('hour')
const loading = ref(false)
const loadData = ref<LoadHistoryDTO | null>(null)
const chartEl = ref<HTMLElement | null>(null)

let chart: echarts.ECharts | null = null
let resizeObserver: ResizeObserver | null = null
let resizeRaf = 0
let reqSeq = 0
let activeReqs = 0
let lastSeriesCount = 0

function setChartEl(el: unknown) {
  const dom = el instanceof HTMLElement ? el : null
  if (chartEl.value !== dom) {
    if (chartEl.value) resizeObserver?.unobserve(chartEl.value)
    chartEl.value = dom
    if (dom) resizeObserver?.observe(dom)
  }
}

async function refresh() {
  const seq = ++reqSeq
  activeReqs += 1
  loading.value = true
  try {
    const data = await fetchV4LoadHistory(loadMetric.value, loadRange.value, 'avg')
    if (seq !== reqSeq) return
    loadData.value = data
    await nextTick()
    if (seq !== reqSeq) return
    renderChart()
  } catch {
    // 瞬态超时保留旧图，等待轮询兜底重试
  } finally {
    activeReqs = Math.max(0, activeReqs - 1)
    if (seq === reqSeq) {
      loading.value = false
    }
  }
}

function renderChart() {
  if (!chartEl.value || !loadData.value) return
  if (chart && chart.getDom() !== chartEl.value) {
    chart.dispose()
    chart = null
    lastSeriesCount = 0
  }
  if (!chart) {
    chart = echarts.init(chartEl.value)
    resizeObserver?.observe(chartEl.value)
  }

  const pal = palette.value
  const rows = loadData.value.series || []
  const labels = rows.map((r) => {
    const ts = Number(r.timestamp ?? r.time ?? 0)
    if (ts > 0) {
      const d = new Date(ts * 1000)
      return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
    }
    return String(r.label ?? '')
  })

  const parsePoint = (raw: unknown): number => parseFloat(String(raw ?? 0).replace('%', '')) || 0
  const hasDualNetwork =
    loadMetric.value === 'network' &&
    rows.some((r) => r.upload !== undefined || r.avg_upload !== undefined || r.download !== undefined || r.avg_download !== undefined)

  const series = hasDualNetwork
    ? [
        {
          id: 'load-primary',
          name: '下行负载',
          type: 'line',
          smooth: true,
          showSymbol: false,
          data: rows.map((r) => parsePoint(r.download ?? r.avg_download ?? 0)),
          lineStyle: { width: 2.2, color: pal.downColor },
          itemStyle: { color: pal.downColor },
          areaStyle: { opacity: 0.18, color: pal.downColor },
        },
        {
          id: 'load-secondary',
          name: '上行负载',
          type: 'line',
          smooth: true,
          showSymbol: false,
          data: rows.map((r) => parsePoint(r.upload ?? r.avg_upload ?? 0)),
          lineStyle: { width: 2, color: pal.upColor },
          itemStyle: { color: pal.upColor },
          areaStyle: { opacity: 0.14, color: pal.upColor },
        },
      ]
    : [
        {
          id: 'load-primary',
          name: loadMetric.value.toUpperCase(),
          type: 'line',
          smooth: true,
          showSymbol: false,
          data: rows.map((r) =>
            parsePoint(
              r.value ??
                r.avg ??
                r.cpu ??
                r.memory ??
                r.connect_num ??
                r.terminals ??
                r.user_num ??
                r.online ??
                r.download ??
                r.count ??
                0
            )
          ),
          lineStyle: { width: 2.2, color: pal.downColor },
          itemStyle: { color: pal.downColor },
          areaStyle: { opacity: 0.18, color: pal.downColor },
        },
      ]

  const countChanged = lastSeriesCount !== 0 && lastSeriesCount !== series.length
  lastSeriesCount = series.length

  chart.setOption(
    {
      animationDuration: 260,
      animationDurationUpdate: 180,
      backgroundColor: 'transparent',
      tooltip: {
        trigger: 'axis',
        backgroundColor: pal.tooltipBg,
        borderColor: pal.tooltipBorder,
        textStyle: { color: pal.textStrong, fontSize: 12 },
      },
      grid: { left: 8, right: 16, top: hasDualNetwork ? 30 : 20, bottom: 6, containLabel: true },
      legend: hasDualNetwork ? { top: 0, textStyle: { color: pal.textMuted, fontSize: 11 } } : undefined,
      xAxis: {
        type: 'category',
        data: labels,
        axisLine: { lineStyle: { color: pal.gridLine } },
        axisLabel: { color: pal.textMuted, fontSize: 11 },
      },
      yAxis: {
        type: 'value',
        splitLine: { lineStyle: { color: pal.gridLine } },
        axisLabel: { color: pal.textMuted, fontSize: 11 },
      },
      series,
    },
    countChanged ? { notMerge: true } : undefined
  )
}

watch([loadMetric, loadRange], () => void refresh())

watch([chartEl, loadData, palette], () => {
  if (chartEl.value && loadData.value) renderChart()
})

watch(
  () => props.routerKey,
  () => {
    loadData.value = null
    lastSeriesCount = 0
    void refresh()
  },
  { immediate: true }
)

watch(
  () => props.pollTick,
  () => {
    if (!loadData.value && activeReqs === 0) {
      void refresh()
    }
  }
)

watch(
  () => props.orderIndex,
  async () => {
    await nextTick()
    chart?.resize()
  }
)

onMounted(() => {
  resizeObserver = new ResizeObserver(() => {
    if (resizeRaf) cancelAnimationFrame(resizeRaf)
    resizeRaf = requestAnimationFrame(() => {
      chart?.resize()
    })
  })
  if (chartEl.value) {
    resizeObserver.observe(chartEl.value)
    if (loadData.value) renderChart()
  }
})

onActivated(() => {
  void nextTick(() => chart?.resize())
})

onUnmounted(() => {
  if (resizeRaf) cancelAnimationFrame(resizeRaf)
  resizeObserver?.disconnect()
  chart?.dispose()
})

defineExpose({ refresh })
</script>

<style scoped>
.pill-group {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.mini-select {
  min-height: 28px;
  padding: 0 10px;
  border-radius: 8px;
  border: 1px solid var(--divider-line);
  background: var(--surface-btn);
  color: var(--text-normal);
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
}

.chart-stage {
  position: relative;
  width: 100%;
  height: 240px;
}

.chart-box {
  width: 100%;
  height: 240px;
}

.chart-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: var(--text-muted);
  pointer-events: none;
}

@media (max-width: 600px) {
  .chart-stage,
  .chart-box {
    height: 200px;
  }
}
</style>

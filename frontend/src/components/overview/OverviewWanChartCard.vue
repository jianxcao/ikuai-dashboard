<template>
  <OverviewSectionShell
    title="WAN 历史速率与链路延迟"
    :layoutLocked="layoutLocked"
    :canMovePrev="canMovePrev"
    :canMoveNext="canMoveNext"
    @move="emit('move', $event)"
  >
    <template #icon>
      <Gauge :size="16" />
    </template>
    <template #actions>
      <div class="v4-seg-bar" role="tablist" aria-label="WAN 历史速率单位与模式">
        <button
          type="button"
          role="tab"
          class="v4-seg-btn"
          :class="{ active: wanChartMode === 'speed' && wanSpeedUnit === 'KB' }"
          :aria-selected="wanChartMode === 'speed' && wanSpeedUnit === 'KB'"
          @click="selectWanSpeedUnit('KB')"
        >
          速率 KB/s
        </button>
        <button
          type="button"
          role="tab"
          class="v4-seg-btn"
          :class="{ active: wanChartMode === 'speed' && wanSpeedUnit === 'MB' }"
          :aria-selected="wanChartMode === 'speed' && wanSpeedUnit === 'MB'"
          @click="selectWanSpeedUnit('MB')"
        >
          速率 MB/s
        </button>
        <button
          type="button"
          role="tab"
          class="v4-seg-btn"
          :class="{ active: wanChartMode === 'rtt' }"
          :aria-selected="wanChartMode === 'rtt'"
          @click="wanChartMode = 'rtt'"
        >
          延迟 RTT ms
        </button>
      </div>
    </template>

    <div :ref="setChartEl" class="chart-box" />
  </OverviewSectionShell>
</template>

<script setup lang="ts">
import { nextTick, onActivated, onMounted, onUnmounted, ref, watch } from 'vue'
import { Gauge } from 'lucide-vue-next'
import { echarts } from '@/lib/echarts'
import type { OverviewDTO } from '@/api/v4'
import { useChartTheme } from '@/composables/useChartTheme'
import { useWanHistoryBuffer } from '@/composables/useWanHistoryBuffer'
import OverviewSectionShell from './OverviewSectionShell.vue'

type WanSpeedUnit = 'KB' | 'MB'
const WAN_SPEED_UNIT_STORAGE_KEY = 'ikuai_v4_wan_speed_unit'

function loadStoredWanSpeedUnit(): WanSpeedUnit {
  if (typeof window === 'undefined') return 'KB'
  try {
    return localStorage.getItem(WAN_SPEED_UNIT_STORAGE_KEY) === 'MB' ? 'MB' : 'KB'
  } catch {
    return 'KB'
  }
}

function convertWanSpeed(kbps: number, unit: WanSpeedUnit): number {
  if (!Number.isFinite(kbps)) return 0
  if (unit === 'MB') {
    const mb = kbps / 1024
    return mb > 0 && mb < 0.1 ? Math.round(mb * 1000) / 1000 : Math.round(mb * 100) / 100
  }
  return Math.round(kbps * 100) / 100
}

const props = defineProps<{
  overview: OverviewDTO
  routerKey: string
  layoutLocked: boolean
  canMovePrev: boolean
  canMoveNext: boolean
  orderIndex: number
}>()

const emit = defineEmits<{
  (e: 'move', delta: -1 | 1): void
}>()

const { palette } = useChartTheme()
const { seriesPoints, ingest } = useWanHistoryBuffer()

const wanChartMode = ref<'speed' | 'rtt'>('speed')
const wanSpeedUnit = ref<WanSpeedUnit>(loadStoredWanSpeedUnit())
const chartEl = ref<HTMLElement | null>(null)

let chart: echarts.ECharts | null = null
let resizeObserver: ResizeObserver | null = null
let resizeRaf = 0
let lastSig = ''
let lastModeKey = ''

function selectWanSpeedUnit(unit: WanSpeedUnit) {
  wanSpeedUnit.value = unit
  wanChartMode.value = 'speed'
  if (typeof window !== 'undefined') {
    try {
      localStorage.setItem(WAN_SPEED_UNIT_STORAGE_KEY, unit)
    } catch {
      // Ignore storage quota errors
    }
  }
}

function setChartEl(el: unknown) {
  const dom = el instanceof HTMLElement ? el : null
  if (chartEl.value !== dom) {
    if (chartEl.value) resizeObserver?.unobserve(chartEl.value)
    chartEl.value = dom
    if (dom) resizeObserver?.observe(dom)
  }
}

function renderChart(force = false) {
  if (!chartEl.value) return
  if (chart && chart.getDom() !== chartEl.value) {
    chart.dispose()
    chart = null
    lastSig = ''
    lastModeKey = ''
  }
  if (!chart) {
    chart = echarts.init(chartEl.value)
    resizeObserver?.observe(chartEl.value)
  }

  const list = seriesPoints.value
  const pal = palette.value
  const speedUnit = wanSpeedUnit.value
  const firstPt = list[0]
  const lastPt = list[list.length - 1]
  const sig = `${props.routerKey}|${wanChartMode.value}|${speedUnit}|${pal.downColor}|${pal.upColor}|${list.length}|${firstPt?.timestamp ?? 0}|${lastPt?.timestamp ?? 0}|${lastPt?.avg_download_kbps ?? 0}|${lastPt?.avg_upload_kbps ?? 0}|${lastPt?.avg_rtt_ms ?? 0}`
  if (!force && sig === lastSig) return
  lastSig = sig

  const labels = list.map((p) => {
    const d = new Date(p.timestamp * 1000)
    return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  })

  const isSpeed = wanChartMode.value === 'speed'
  const speedUnitLabel = speedUnit === 'MB' ? 'MB/s' : 'KB/s'
  const modeKey = isSpeed ? `speed:${speedUnit}` : 'rtt'
  const modeChanged = lastModeKey !== '' && modeKey !== lastModeKey
  lastModeKey = modeKey

  // 固定两条折线稳定的 series id，常规 4 秒轮询增量更新不触发 replaceMerge，避免整图清空重画闪烁
  chart.setOption(
    {
      animationDuration: 240,
      animationDurationUpdate: 220,
      animationEasingUpdate: 'linear',
      backgroundColor: 'transparent',
      tooltip: {
        trigger: 'axis',
        backgroundColor: pal.tooltipBg,
        borderColor: pal.tooltipBorder,
        textStyle: { color: pal.textStrong, fontSize: 12 },
      },
      legend: {
        top: 0,
        textStyle: { color: pal.textMuted, fontSize: 11 },
      },
      grid: { left: 8, right: 16, top: 32, bottom: 6, containLabel: true },
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
      series: isSpeed
        ? [
            {
              id: 'wan-primary',
              name: `平均下行 (${speedUnitLabel})`,
              type: 'line',
              smooth: true,
              showSymbol: false,
              data: list.map((p) => convertWanSpeed(p.avg_download_kbps, speedUnit)),
              lineStyle: { width: 2, type: 'solid', color: pal.downColor },
              itemStyle: { color: pal.downColor },
              areaStyle: { opacity: 0.16, color: pal.downColor },
            },
            {
              id: 'wan-secondary',
              name: `平均上行 (${speedUnitLabel})`,
              type: 'line',
              smooth: true,
              showSymbol: false,
              data: list.map((p) => convertWanSpeed(p.avg_upload_kbps, speedUnit)),
              lineStyle: { width: 2, type: 'solid', color: pal.upColor },
              itemStyle: { color: pal.upColor },
              areaStyle: { opacity: 0.14, color: pal.upColor },
            },
          ]
        : [
            {
              id: 'wan-primary',
              name: '平均延迟 (ms)',
              type: 'line',
              smooth: true,
              showSymbol: false,
              data: list.map((p) => p.avg_rtt_ms),
              lineStyle: { width: 2, type: 'solid', color: pal.accent },
              itemStyle: { color: pal.accent },
              areaStyle: { opacity: 0.16, color: pal.accent },
            },
            {
              id: 'wan-secondary',
              name: '最大延迟 (ms)',
              type: 'line',
              smooth: true,
              showSymbol: false,
              data: list.map((p) => p.max_rtt_ms),
              lineStyle: { width: 1.5, type: 'dashed', color: pal.warn },
              itemStyle: { color: pal.warn },
              areaStyle: { opacity: 0 },
            },
          ],
    },
    modeChanged ? { notMerge: true } : undefined
  )
}

watch(
  () => [props.routerKey, props.overview] as const,
  ([routerKey, ov]) => {
    ingest(routerKey, ov)
    if (chartEl.value) renderChart(false)
  },
  { immediate: true }
)

watch([wanChartMode, wanSpeedUnit, palette], () => {
  if (chartEl.value) renderChart(true)
})

watch(chartEl, (dom) => {
  if (dom) renderChart(true)
})

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
    renderChart(true)
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
</script>

<style scoped>
.chart-box {
  width: 100%;
  height: 240px;
}

@media (max-width: 600px) {
  .chart-box {
    height: 200px;
  }
}
</style>

<template>
  <div class="chart-section glass-card">
    <div class="section-header">
      <div class="header-left">
        <span class="section-title">
          <span class="title-dot"></span>
          实时流量趋势
        </span>
      </div>
      <div class="header-legend">
        <div class="legend-item">
          <span class="legend-color" style="background: var(--system-green)"></span>
          上传速率
        </div>
        <div class="legend-item">
          <span class="legend-color" style="background: var(--system-blue)"></span>
          下载速率
        </div>
      </div>
    </div>

    <div class="chart-container">
      <div ref="chartRef" class="echarts-instance"></div>

      <!-- Loading Overlay -->
      <div v-if="loading && chartLabels.length === 0" class="chart-loading">
        <div class="spinner"></div>
        <span>初始化数据收集中...</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, markRaw } from 'vue'
import * as echarts from 'echarts'
import { formatSpeed } from '@/composables/useFormatters'

const props = withDefaults(
  defineProps<{
    chartLabels?: string[]
    chartUpload?: number[]
    chartDownload?: number[]
    loading?: boolean
  }>(),
  {
    chartLabels: () => [],
    chartUpload: () => [],
    chartDownload: () => [],
    loading: false
  }
)

const chartRef = ref<HTMLElement | null>(null)
let chartInstance: echarts.ECharts | null = null
let resizeObserver: ResizeObserver | null = null
let themeObserver: MutationObserver | null = null

function cssVar(name: string) {
  if (typeof document === 'undefined') return ''
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}

function chartTokens() {
  return {
    textPrimary: cssVar('--text-primary') || 'rgba(248, 253, 255, 0.96)',
    textSecondary: cssVar('--text-secondary') || 'rgba(222, 239, 245, 0.72)',
    glassStrong: cssVar('--glass-bg-strong') || 'rgba(30, 30, 35, 0.75)',
    glassBorder: cssVar('--glass-border') || 'rgba(255, 255, 255, 0.15)',
    controlBorder: cssVar('--control-border') || 'rgba(255, 255, 255, 0.12)',
    green: cssVar('--system-green') || '#62f4ad',
    greenDim: cssVar('--system-green-dim') || 'rgba(98, 244, 173, 0.18)',
    blue: cssVar('--system-blue') || '#5ecbff',
    blueDim: cssVar('--system-blue-dim') || 'rgba(94, 203, 255, 0.18)'
  }
}

function buildOption() {
  const tokens = chartTokens()

  return {
    grid: {
      top: 30,
      left: 20,
      right: 20,
      bottom: 20,
      containLabel: true
    },
    tooltip: {
      trigger: 'axis',
      backgroundColor: tokens.glassStrong,
      borderColor: tokens.glassBorder,
      textStyle: { color: tokens.textPrimary },
      extraCssText:
        'backdrop-filter: blur(18px); border-radius: 14px; box-shadow: 0 18px 48px rgba(0,0,0,0.24);',
      axisPointer: {
        type: 'line',
        lineStyle: { color: tokens.controlBorder, type: 'dashed' }
      },
      formatter: (params: any) => {
        let res = `<div style="font-family: inherit; font-size: 13px; font-weight: 600; margin-bottom: 6px; color: ${tokens.textSecondary}">${params[0].axisValue}</div>`
        params.forEach((p: any) => {
          const color = p.seriesName === '上传' ? tokens.green : tokens.blue
          res += `
            <div style="display: flex; align-items: center; justify-content: space-between; gap: 24px; font-family: inherit; font-size: 14px; margin-bottom: 4px; color: ${tokens.textPrimary};">
              <span style="display: flex; align-items: center; gap: 6px;">
                <span style="display:inline-block; border-radius:50%; width:8px; height:8px; background-color:${color};"></span>
                ${p.seriesName}
              </span>
              <span style="font-weight: 700; font-family: ui-monospace, SFMono-Regular, monospace; color: ${color}">${formatSpeed(p.value)}</span>
            </div>
          `
        })
        return res
      }
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: props.chartLabels,
      axisLine: { lineStyle: { color: tokens.controlBorder } },
      axisLabel: { color: tokens.textSecondary, fontFamily: 'ui-monospace, monospace' },
      splitLine: { show: false }
    },
    yAxis: {
      type: 'value',
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: {
        color: tokens.textSecondary,
        fontFamily: 'ui-monospace, monospace',
        formatter: (val: number) => {
          if (val >= 1024 ** 3) return (val / 1024 ** 3).toFixed(1) + ' G/s'
          if (val >= 1024 ** 2) return (val / 1024 ** 2).toFixed(1) + ' M/s'
          if (val >= 1024) return (val / 1024).toFixed(0) + ' K/s'
          return val + ' B/s'
        }
      },
      splitLine: {
        lineStyle: { color: tokens.controlBorder, type: 'solid' }
      }
    },
    series: [
      {
        name: '上传',
        type: 'line',
        smooth: true,
        symbol: 'none',
        lineStyle: { width: 3, color: tokens.green },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: tokens.greenDim },
            { offset: 1, color: 'rgba(98, 244, 173, 0.02)' }
          ])
        },
        data: props.chartUpload
      },
      {
        name: '下载',
        type: 'line',
        smooth: true,
        symbol: 'none',
        lineStyle: { width: 3, color: tokens.blue },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: tokens.blueDim },
            { offset: 1, color: 'rgba(94, 203, 255, 0.02)' }
          ])
        },
        data: props.chartDownload
      }
    ]
  }
}

function refreshChartOptions() {
  if (!chartInstance) return
  chartInstance.setOption(buildOption(), true)
}

// 初始化 ECharts
function initChart() {
  if (!chartRef.value) return

  chartInstance = markRaw(echarts.init(chartRef.value))
  chartInstance.setOption(buildOption())
}

// 监听数据变化，增量更新图表
watch(
  () => props.chartLabels,
  (newLabels) => {
    if (!chartInstance) return
    chartInstance.setOption({
      xAxis: { data: newLabels },
      series: [{ data: props.chartUpload }, { data: props.chartDownload }]
    })
  },
  { deep: true }
)

onMounted(() => {
  initChart()

  // 响应式 Resize
  resizeObserver = new ResizeObserver(() => {
    if (chartInstance) chartInstance.resize()
  })
  if (chartRef.value) resizeObserver.observe(chartRef.value)

  themeObserver = new MutationObserver(refreshChartOptions)
  themeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['data-theme']
  })
})

onUnmounted(() => {
  if (resizeObserver) resizeObserver.disconnect()
  if (themeObserver) themeObserver.disconnect()
  if (chartInstance) chartInstance.dispose()
})
</script>

<style scoped>
.chart-section {
  padding: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.section-header {
  padding: 18px 24px;
  border-bottom: 1px solid var(--glass-border);
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 0.04em;
  color: var(--text-primary);
}

.title-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--system-blue);
  box-shadow: 0 0 12px var(--system-blue-dim);
}

.header-legend {
  display: flex;
  gap: 20px;
  flex-wrap: wrap;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
}

.legend-color {
  width: 12px;
  height: 12px;
  border-radius: 4px;
}

.chart-container {
  height: clamp(260px, 38vw, 360px);
  position: relative;
  width: 100%;
}

.echarts-instance {
  width: 100%;
  height: 100%;
}

.chart-loading {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  background: var(--glass-bg-strong);
  backdrop-filter: blur(16px);
  color: var(--text-secondary);
  font-size: 14px;
  z-index: 10;
}

@media (max-width: 560px) {
  .section-header {
    align-items: flex-start;
    flex-direction: column;
    padding: 16px;
  }

  .header-legend {
    gap: 12px;
  }

  .chart-container {
    height: 248px;
  }
}
</style>

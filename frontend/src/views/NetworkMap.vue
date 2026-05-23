<template>
  <div class="network-map-container glass-panel">
    <div class="panel-header">
      <div class="panel-heading">
        <h2 class="panel-title">网络拓扑</h2>
        <div v-if="hasData" class="map-stats" aria-label="拓扑统计">
          <span>{{ nodeCount }} 节点</span>
          <span>{{ linkCount }} 链路</span>
        </div>
      </div>
      <div class="panel-actions">
        <button class="btn-refresh liquid-button" @click="fetchData" :disabled="loading">
          <span v-if="loading" class="spinner"></span>
          刷新
        </button>
      </div>
    </div>
    
    <div class="map-content">
      <div v-if="loading && !chartInstance" class="loading-state">
        <div class="loader"></div>
        <p>正在分析网络结构...</p>
      </div>
      <div v-else-if="error && !hasData" class="error-state">
        <p>{{ error }}</p>
        <button @click="fetchData" class="btn-retry liquid-button">重试</button>
      </div>
      <div ref="chartRef" class="chart-container" v-show="hasData || loading"></div>
      <div v-if="error && hasData" class="map-notice">
        刷新失败：{{ error }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onActivated, onMounted, onUnmounted, ref, shallowRef } from 'vue'
import * as echarts from 'echarts'
import { fetchNetworkMap, type NetworkMapData } from '@/api/monitor'

const chartRef = ref<HTMLElement | null>(null)
const chartInstance = shallowRef<echarts.ECharts | null>(null)
const lastData = shallowRef<NetworkMapData | null>(null)
const loading = ref(false)
const error = ref('')
const hasData = ref(false)
const nodeCount = ref(0)
const linkCount = ref(0)
let themeObserver: MutationObserver | null = null

function cssVar(name: string) {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}

const initChart = (data: NetworkMapData) => {
  if (!chartRef.value) return
  const nodeIds = new Set(data.nodes.map((node) => node.id).filter(Boolean))
  const nodes = data.nodes.map((n) => ({
    id: n.id,
    name: n.name || n.id,
    value: n.ip || '',
    category: n.category,
    symbolSize: n.category === 0 ? 60 : (n.category === 3 ? 30 : 45),
  }))
  const links = data.links.filter((link) => nodeIds.has(link.source) && nodeIds.has(link.target))

  nodeCount.value = nodes.length
  linkCount.value = links.length
  hasData.value = nodes.length > 0
  if (!hasData.value) return

  if (!chartInstance.value) {
    chartInstance.value = echarts.init(chartRef.value)
  }

  const textPrimary = cssVar('--text-primary') || '#f8fdff'
  const textSecondary = cssVar('--text-secondary') || 'rgba(222, 239, 245, 0.72)'
  const glassStrong = cssVar('--glass-bg-strong') || 'rgba(20, 30, 36, 0.78)'
  const glassBorder = cssVar('--glass-border') || 'rgba(196, 245, 255, 0.18)'
  const blue = cssVar('--system-blue') || '#5ecbff'
  const red = cssVar('--system-red') || '#ff6f7d'
  const green = cssVar('--system-green') || '#62f4ad'
  const orange = cssVar('--system-orange') || '#ffc46b'

  const option = {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'item',
      backgroundColor: glassStrong,
      borderColor: glassBorder,
      textStyle: { color: textPrimary },
      extraCssText: 'backdrop-filter: blur(18px); border-radius: 14px;',
      formatter: '{b}: {c}'
    },
    series: [
      {
        type: 'graph',
        layout: 'force',
        force: {
          repulsion: 1000,
          edgeLength: 150
        },
        roam: true,
        label: {
          show: true,
          position: 'right',
          formatter: '{b}',
          color: textSecondary,
          fontWeight: 600
        },
        lineStyle: {
          color: textSecondary,
          curveness: 0.3
        },
        emphasis: {
          focus: 'adjacency',
          lineStyle: { width: 4 }
        },
        categories: [
          { name: 'Router', itemStyle: { color: blue } },
          { name: 'WAN', itemStyle: { color: red } },
          { name: 'LAN', itemStyle: { color: green } },
          { name: 'Device', itemStyle: { color: orange } }
        ],
        data: nodes,
        links
      }
    ]
  }

  chartInstance.value.setOption(option, true)
  chartInstance.value.resize()
}

const fetchData = async () => {
  loading.value = true
  error.value = ''
  try {
    const data = await fetchNetworkMap()
    if (!data?.nodes?.length) {
      throw new Error('暂无可展示的拓扑数据')
    }
    lastData.value = data
    await nextTick()
    initChart(data)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '网络请求错误'
  } finally {
    loading.value = false
  }
}

const handleResize = () => {
  chartInstance.value?.resize()
}

const refreshTheme = () => {
  if (lastData.value) {
    initChart(lastData.value)
  }
}

onMounted(() => {
  fetchData()
  window.addEventListener('resize', handleResize)
  themeObserver = new MutationObserver(refreshTheme)
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] })
})

onActivated(() => {
  void nextTick(handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  themeObserver?.disconnect()
  chartInstance.value?.dispose()
})
</script>

<style scoped>
.network-map-container {
  display: flex;
  flex-direction: column;
  min-height: min(720px, calc(100vh - 120px));
  height: calc(100vh - 120px);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 20px;
  border-bottom: 1px solid var(--glass-border);
}

.panel-heading {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
}

.panel-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-primary);
}

.map-stats {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.map-stats span {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  padding: 0 9px;
  border: 1px solid var(--control-border);
  border-radius: 999px;
  background: var(--control-bg);
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 700;
}

.map-content {
  flex: 1;
  position: relative;
  min-height: 400px;
}

.chart-container {
  width: 100%;
  height: 100%;
  min-height: inherit;
}

.loading-state, .error-state {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  color: var(--text-secondary);
}

.map-notice {
  position: absolute;
  right: 18px;
  bottom: 18px;
  max-width: min(420px, calc(100% - 36px));
  padding: 10px 12px;
  border: 1px solid var(--system-red-dim);
  border-radius: var(--radius-sm);
  background: var(--glass-bg-strong);
  color: var(--system-red);
  box-shadow: var(--glass-shadow);
  font-size: 13px;
}

.btn-refresh, .btn-retry {
  min-height: 34px;
  padding: 0 16px;
  font-weight: 500;
}

.btn-refresh:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-refresh .spinner {
  display: inline-block;
  width: 14px;
  height: 14px;
  margin-right: 6px;
  vertical-align: text-bottom;
}

.loader {
  width: 40px;
  height: 40px;
  border: 3px solid var(--glass-border);
  border-top-color: var(--system-blue);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 760px) {
  .network-map-container {
    height: auto;
    min-height: 560px;
  }

  .panel-header {
    align-items: flex-start;
    flex-direction: column;
    padding: 16px;
  }

  .panel-actions,
  .btn-refresh {
    width: 100%;
  }

  .map-content {
    min-height: 480px;
  }
}

@media (max-width: 420px) {
  .map-content {
    min-height: 420px;
  }
}
</style>

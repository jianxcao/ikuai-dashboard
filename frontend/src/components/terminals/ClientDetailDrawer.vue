<template>
  <teleport to="body">
    <div v-if="client" class="drawer-backdrop" @click.self="emit('close')">
      <div class="drawer-panel v4-panel" role="dialog" aria-label="终端流量画像">
        <div class="drawer-head">
          <div>
            <strong>{{ client.comment || client.hostname || client.ip_addr }}</strong>
            <span class="font-mono">{{ client.ip_addr }} · {{ client.mac }}</span>
          </div>
          <button type="button" class="v4-btn btn-xs" aria-label="关闭抽屉" @click="emit('close')">
            <X :size="15" />
          </button>
        </div>

        <div class="drawer-body">
          <div v-if="loading" class="loading-state"><div class="spinner" /></div>
          <div v-else-if="error" class="error-banner">{{ error }}</div>
          <template v-else-if="detail">
            <div class="detail-section">
              <h4>5 分钟速率趋势</h4>
              <div ref="chartEl" class="detail-chart" />
            </div>
            <div class="detail-section">
              <h4>Top L7 应用协议分布</h4>
              <div v-if="detail.app_protocols.length === 0" class="empty-state">暂无应用层统计数据</div>
              <div v-else class="app-list">
                <div v-for="(app, idx) in detail.app_protocols" :key="idx" class="app-row">
                  <strong>{{ String(app.appname ?? app.app_name ?? '未知应用') }}</strong>
                  <span class="font-mono">↓ {{ formatBytes(Number(app.total_down ?? 0)) }} · ↑ {{ formatBytes(Number(app.total_up ?? 0)) }}</span>
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { nextTick, onUnmounted, ref, watch } from 'vue'
import { X } from 'lucide-vue-next'
import { echarts } from '@/lib/echarts'
import { fetchV4ClientDetail, type ClientDetailDTO, type ClientRecordDTO } from '@/api/v4'
import { useChartTheme } from '@/composables/useChartTheme'
import { formatBytes } from '@/composables/useFormatters'

const props = defineProps<{
  client: ClientRecordDTO | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const { palette } = useChartTheme()

const loading = ref(false)
const error = ref<string | null>(null)
const detail = ref<ClientDetailDTO | null>(null)
const chartEl = ref<HTMLElement | null>(null)

let chart: echarts.ECharts | null = null
let reqSeq = 0

function disposeChart() {
  chart?.dispose()
  chart = null
}

function renderChart() {
  if (!chartEl.value || !detail.value) return
  if (chart && chart.getDom() !== chartEl.value) {
    disposeChart()
  }
  if (!chart) chart = echarts.init(chartEl.value)
  const pal = palette.value
  const list = detail.value.traffic_load || []
  chart.setOption(
    {
      backgroundColor: 'transparent',
      tooltip: {
        trigger: 'axis',
        backgroundColor: pal.tooltipBg,
        borderColor: pal.tooltipBorder,
        textStyle: { color: pal.textStrong },
      },
      grid: { left: 48, right: 12, top: 18, bottom: 24 },
      xAxis: {
        type: 'category',
        data: list.map((item) => {
          const ts = Number(item.timestamp ?? 0)
          if (!ts) return ''
          const d = new Date(ts * 1000)
          return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
        }),
        axisLabel: { color: pal.textMuted, fontSize: 10 },
      },
      yAxis: {
        type: 'value',
        splitLine: { lineStyle: { color: pal.gridLine } },
        axisLabel: { color: pal.textMuted, fontSize: 10 },
      },
      series: [
        {
          name: '下行',
          type: 'line',
          smooth: true,
          showSymbol: false,
          data: list.map((x) => Number(x.download ?? 0)),
          itemStyle: { color: pal.downColor },
        },
        {
          name: '上行',
          type: 'line',
          smooth: true,
          showSymbol: false,
          data: list.map((x) => Number(x.upload ?? 0)),
          itemStyle: { color: pal.upColor },
        },
      ],
    },
    true
  )
}

async function loadDetail(client: ClientRecordDTO) {
  const seq = ++reqSeq
  loading.value = true
  error.value = null
  detail.value = null
  disposeChart()
  try {
    const res = await fetchV4ClientDetail(client.ip_addr, client.mac)
    if (seq !== reqSeq) return
    detail.value = res
    loading.value = false
    await nextTick()
    if (seq !== reqSeq) return
    renderChart()
  } catch (e: unknown) {
    if (seq !== reqSeq) return
    if (e && typeof e === 'object' && 'response' in e) {
      const resp = (e as { response?: { data?: { message?: string } } }).response
      if (resp?.data?.message) {
        error.value = resp.data.message
        return
      }
    }
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    if (seq === reqSeq) loading.value = false
  }
}

watch(
  () => props.client,
  (c) => {
    if (c) {
      void loadDetail(c)
    } else {
      reqSeq += 1
      detail.value = null
      disposeChart()
    }
  },
  { immediate: true }
)

watch([palette, chartEl], () => {
  if (detail.value && chartEl.value) renderChart()
})

onUnmounted(() => {
  disposeChart()
})
</script>

<style scoped>
.drawer-backdrop {
  position: fixed;
  inset: 0;
  z-index: 250;
  background: rgba(6, 10, 20, 0.45);
  -webkit-backdrop-filter: blur(8px);
  backdrop-filter: blur(8px);
  display: flex;
  justify-content: flex-end;
}

.drawer-panel {
  width: min(480px, 100vw);
  height: 100%;
  border-radius: var(--radius-xl) 0 0 var(--radius-xl);
  display: flex;
  flex-direction: column;
  background: var(--surface-popover);
  border-color: var(--chrome-border);
  box-shadow: var(--chrome-shadow);
  -webkit-backdrop-filter: var(--backdrop-chrome);
  backdrop-filter: var(--backdrop-chrome);
}

.drawer-head {
  padding: 16px 20px;
  border-bottom: 1px solid var(--divider-line);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.drawer-head > div > strong {
  display: block;
  font-size: 15px;
}

.drawer-head > div > span {
  font-size: 11px;
  color: var(--text-muted);
}

.drawer-body {
  padding: 18px 20px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.detail-section h4 {
  margin: 0 0 10px;
  font-size: 13px;
  color: var(--text-strong);
}

.detail-chart {
  width: 100%;
  height: 220px;
}

.app-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.app-row {
  display: flex;
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: var(--radius-md);
  border: 1px solid var(--divider-line);
  background: var(--surface-input);
  font-size: 12px;
}

.empty-state {
  padding: 28px 16px;
  text-align: center;
  color: var(--text-muted);
}

@media (max-width: 640px) {
  .drawer-backdrop {
    align-items: flex-end;
  }

  .drawer-panel {
    width: 100%;
    height: auto;
    max-height: 88dvh;
    border-radius: var(--radius-xl) var(--radius-xl) 0 0;
  }

  .drawer-body {
    padding: 14px 16px;
  }
}
</style>

<template>
  <section class="traffic-page">
    <header class="ws-head v4-panel">
      <div>
        <span class="ws-eyebrow font-mono">TRAFFIC ANALYTICS & TOPOLOGY</span>
        <h1>应用协议画像与实时拓扑</h1>
        <p>透明解析路由器底层传输协议占比、L7 应用流量排名及多线全网节点拓扑关系。</p>
      </div>
      <div class="v4-seg-bar" role="tablist">
        <button type="button" class="v4-seg-btn" :class="{ active: activeSub === 'analytics' }" @click="selectSub('analytics')">
          <PieChart :size="14" />
          <span>应用与协议洞察</span>
        </button>
        <button type="button" class="v4-seg-btn" :class="{ active: activeSub === 'topology' }" @click="selectSub('topology')">
          <Network :size="14" />
          <span>实时网络拓扑图</span>
        </button>
      </div>
    </header>

    <div v-if="error" class="error-banner">{{ error }}</div>

    <!-- Analytics Tab -->
    <template v-if="activeSub === 'analytics'">
      <div class="two-col">
        <article class="v4-card panel-box">
          <div class="box-head">
            <h3>传输协议分布</h3>
            <button type="button" class="v4-btn btn-xs" @click="loadAnalytics">
              <RefreshCw :size="13" />
              <span>刷新</span>
            </button>
          </div>
          <div ref="protoChartEl" class="chart-area" />
        </article>

        <article class="v4-card panel-box">
          <div class="box-head">
            <h3>24h 应用层流量排行榜</h3>
          </div>
          <div v-if="!analytics?.app_summary?.length" class="empty-box">暂无应用层统计数据</div>
          <div v-else class="rank-list">
            <div v-for="(item, idx) in analytics.app_summary" :key="idx" class="rank-item">
              <div class="rank-top">
                <span class="rank-name">
                  <em class="font-mono">#{{ idx + 1 }}</em>
                  {{ String(item.appname ?? item.app_name ?? item.appname_level1 ?? '未知应用') }}
                </span>
                <span class="font-mono">↓ {{ formatBytes(Number(item.total_down ?? 0)) }} / ↑ {{ formatBytes(Number(item.total_up ?? 0)) }}</span>
              </div>
              <div class="bar-bg">
                <div class="bar-fill" :style="{ width: `${appPercent(item)}%` }" />
              </div>
            </div>
          </div>
        </article>
      </div>

      <article class="v4-card panel-box">
        <div class="box-head">
          <h3>当前活跃应用实时负载</h3>
        </div>
        <div v-if="!analytics?.app_load?.length" class="empty-box">当前暂无活跃应用负载明细</div>
        <div v-else class="table-wrap">
          <table class="v4-table">
            <thead>
              <tr>
                <th
                  v-for="col in appLoadColumns"
                  :key="col.key"
                  class="sortable-th"
                  :class="{ 'is-sorted': appLoadSort.sortKey.value === col.key }"
                  :aria-sort="appLoadSort.ariaSort(col.key)"
                  tabindex="0"
                  @click="appLoadSort.toggleSort(col.key)"
                  @keydown.enter.prevent="appLoadSort.toggleSort(col.key)"
                  @keydown.space.prevent="appLoadSort.toggleSort(col.key)"
                >
                  <span class="sort-th-inner">
                    <span>{{ col.label }}</span>
                    <SortIndicator :active="appLoadSort.sortKey.value === col.key" :dir="appLoadSort.sortDir.value" />
                  </span>
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(row, idx) in sortedAppLoad" :key="String(row.appname ?? row.app_name ?? idx)">
                <td><strong>{{ String(row.appname ?? row.app_name ?? '未知') }}</strong></td>
                <td class="font-mono">{{ row.conn_cnt ?? row.connect_num ?? 0 }}</td>
                <td class="font-mono ic-up">{{ formatSpeed(Number(row.upload ?? 0)) }}</td>
                <td class="font-mono ic-down">{{ formatSpeed(Number(row.download ?? 0)) }}</td>
                <td class="font-mono">{{ formatBytes(Number(row.total_down ?? 0)) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="quick-links-bar">
          <button type="button" class="v4-btn btn-xs" @click="emit('navigate', { tab: 'routing', subResource: 'stream-l7' })">
            配置 L7 应用分流规则 →
          </button>
          <button type="button" class="v4-btn btn-xs" @click="emit('navigate', { tab: 'security', subResource: 'l7-rules' })">
            配置 L7 应用协议管控/阻断 →
          </button>
        </div>
      </article>
    </template>

    <!-- Topology Tab -->
    <template v-else>
      <article class="v4-card panel-box">
        <div class="box-head">
          <div class="legend-row">
            <span class="v4-badge v4-badge-up">WAN 出口</span>
            <span class="v4-badge v4-badge-accent">iKuai V4 网关</span>
            <span class="v4-badge">LAN / AP</span>
            <span class="v4-badge v4-badge-warn">在线终端</span>
          </div>
          <button type="button" class="v4-btn btn-xs" @click="loadTopology">
            <RefreshCw :size="13" />
            <span>重新发现拓扑</span>
          </button>
        </div>
        <div v-if="selectedNode" class="node-inspector">
          <div class="node-info">
            <span class="v4-badge v4-badge-accent">{{ selectedNode.type }}</span>
            <strong>{{ selectedNode.name }}</strong>
            <span class="font-mono">{{ selectedNode.ip || '—' }}</span>
          </div>
          <div class="node-actions">
            <button type="button" class="v4-btn btn-xs" @click="emit('navigate', { tab: 'network', subResource: 'wan-config' })">接口配置</button>
            <button type="button" class="v4-btn btn-xs" @click="emit('navigate', { tab: 'terminals', subResource: 'online' })">在线终端</button>
          </div>
        </div>
        <div ref="topoChartEl" class="topo-canvas" />
      </article>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onActivated, onMounted, onUnmounted, ref, watch } from 'vue'
import { echarts } from '@/lib/echarts'
import { Network, PieChart, RefreshCw } from 'lucide-vue-next'
import { fetchNetworkMap, type NetworkMapData } from '@/api/monitor'
import { fetchV4Analytics, type AppProtocolAnalyticsDTO } from '@/api/v4'
import { useAppRoute } from '@/composables/useAppRoute'
import { useFleet } from '@/composables/useFleet'
import { useChartTheme } from '@/composables/useChartTheme'
import { formatBytes, formatSpeed } from '@/composables/useFormatters'
import SortIndicator from '@/components/common/SortIndicator.vue'
import { useTableSort, type SortValueHint } from '@/composables/useTableSort'

const props = defineProps<{
  initialSubResource?: string
  navSeq?: number
}>()

const emit = defineEmits<{
  (e: 'navigate', payload: { tab: string; subResource?: string }): void
}>()

const { currentTab, syncSubTab } = useAppRoute()
const { routerEpoch } = useFleet()
const { palette } = useChartTheme()

const activeSub = ref<'analytics' | 'topology'>(
  props.initialSubResource === 'topology' ? 'topology' : 'analytics'
)

if (currentTab.value === 'traffic-topology' && !props.initialSubResource) {
  syncSubTab('traffic-topology', activeSub.value, { replace: true })
}

function selectSub(sub: 'analytics' | 'topology') {
  activeSub.value = sub
  syncSubTab('traffic-topology', sub, { replace: false })
}
const error = ref<string | null>(null)
const analytics = ref<AppProtocolAnalyticsDTO | null>(null)
const topoData = ref<NetworkMapData | null>(null)
const selectedNode = ref<{ name: string; type: string; ip: string } | null>(null)

const protoChartEl = ref<HTMLElement | null>(null)
const topoChartEl = ref<HTMLElement | null>(null)
let protoChart: echarts.ECharts | null = null
let topoChart: echarts.ECharts | null = null
let resizeObs: ResizeObserver | null = null
let analyticsSeq = 0
let topoSeq = 0

function extractErr(e: unknown): string {
  if (e && typeof e === 'object' && 'response' in e) {
    const resp = (e as { response?: { data?: { message?: string } } }).response
    if (resp?.data?.message) return resp.data.message
  }
  return e instanceof Error ? e.message : String(e)
}

const maxAppBytes = computed(() => {
  const list = analytics.value?.app_summary || []
  return Math.max(1, ...list.map((i) => Number(i.total_down ?? 0)))
})

type AppLoadSortKey = 'name' | 'connections' | 'upload' | 'download' | 'total_down'
const appLoadColumns: Array<{ key: AppLoadSortKey; label: string; hint: SortValueHint }> = [
  { key: 'name', label: '应用名称', hint: 'text' },
  { key: 'connections', label: '连接数', hint: 'number' },
  { key: 'upload', label: '实时上行', hint: 'number' },
  { key: 'download', label: '实时下行', hint: 'number' },
  { key: 'total_down', label: '累计下行', hint: 'number' },
]

const appLoadSort = useTableSort<AppLoadSortKey>({
  defaultDirForKey: (key) => (key === 'name' ? 'asc' : 'desc'),
})

const sortedAppLoad = computed(() =>
  appLoadSort.sortRows(
    analytics.value?.app_load || [],
    (row, key) => {
      switch (key) {
        case 'name':
          return String(row.appname ?? row.app_name ?? '')
        case 'connections':
          return Number(row.conn_cnt ?? row.connect_num ?? 0)
        case 'upload':
          return Number(row.upload ?? 0)
        case 'download':
          return Number(row.download ?? 0)
        case 'total_down':
          return Number(row.total_down ?? 0)
      }
    },
    (key) => appLoadColumns.find((c) => c.key === key)?.hint ?? 'auto'
  )
)

function appPercent(item: Record<string, unknown>): number {
  return Math.max(6, Math.round((Number(item.total_down ?? 0) / maxAppBytes.value) * 100))
}

async function loadAnalytics() {
  const seq = ++analyticsSeq
  error.value = null
  try {
    const res = await fetchV4Analytics()
    if (seq !== analyticsSeq) return
    analytics.value = res
    await nextTick()
    renderProtoChart()
  } catch (e: unknown) {
    if (seq !== analyticsSeq) return
    error.value = extractErr(e)
  }
}

async function loadTopology() {
  const seq = ++topoSeq
  error.value = null
  try {
    const res = await fetchNetworkMap()
    if (seq !== topoSeq) return
    topoData.value = res
    await nextTick()
    renderTopoChart()
  } catch (e: unknown) {
    if (seq !== topoSeq) return
    error.value = extractErr(e)
  }
}

function renderProtoChart() {
  if (!protoChartEl.value || !analytics.value) return
  if (protoChart && protoChart.getDom() !== protoChartEl.value) {
    protoChart.dispose()
    protoChart = null
  }
  if (!protoChart) {
    protoChart = echarts.init(protoChartEl.value)
    resizeObs?.observe(protoChartEl.value)
  }
  const pal = palette.value
  const data = (analytics.value.protocols || []).map((p, idx) => ({
    name: String(p.proto_name ?? p.name ?? `Proto-${idx}`),
    value: Number(p.total_down ?? p.percent ?? p.total ?? 0),
  }))

  protoChart.setOption({
    backgroundColor: 'transparent',
    color: pal.seriesPalette,
    tooltip: {
      trigger: 'item',
      backgroundColor: pal.tooltipBg,
      borderColor: pal.tooltipBorder,
      textStyle: { color: pal.textStrong },
    },
    legend: { bottom: 0, textStyle: { color: pal.textMuted, fontSize: 11 } },
    series: [
      {
        type: 'pie',
        radius: ['46%', '72%'],
        center: ['50%', '44%'],
        itemStyle: { borderRadius: 6 },
        label: { color: pal.textMuted, fontSize: 11 },
        data: data.length ? data : [{ name: '暂无流量数据', value: 1 }],
      },
    ],
  }, true)
}

function renderTopoChart() {
  if (!topoChartEl.value || !topoData.value) return
  if (topoChart && topoChart.getDom() !== topoChartEl.value) {
    topoChart.dispose()
    topoChart = null
  }
  if (!topoChart) {
    topoChart = echarts.init(topoChartEl.value)
    resizeObs?.observe(topoChartEl.value)
    topoChart.on('click', (params) => {
      const d = (params?.data ?? null) as { name?: string; ip?: string; type?: string } | null
      if (d?.name) {
        selectedNode.value = {
          name: d.name,
          type: d.type || 'node',
          ip: d.ip || '',
        }
      }
    })
  }
  const pal = palette.value
  const catColors = [pal.downColor, pal.upColor, pal.accent, pal.warn]

  topoChart.setOption({
    backgroundColor: 'transparent',
    tooltip: {
      backgroundColor: pal.tooltipBg,
      borderColor: pal.tooltipBorder,
      textStyle: { color: pal.textStrong },
      formatter: (params: { data?: { name?: string; ip?: string; type?: string } }) => {
        const d = params.data
        if (!d) return ''
        return `<strong>${d.name || ''}</strong><br/>类型: ${d.type || ''}<br/>IP: ${d.ip || '—'}`
      },
    },
    series: [
      {
        type: 'graph',
        layout: 'force',
        roam: true,
        draggable: true,
        force: { repulsion: 280, edgeLength: [85, 140] },
        label: { show: true, position: 'bottom', color: pal.textStrong, fontSize: 11 },
        lineStyle: { color: pal.textMuted, width: 1.4, opacity: 0.6, curveness: 0.08 },
        nodes: topoData.value.nodes.map((n) => ({
          ...n,
          symbolSize: n.type === 'router' ? 48 : n.type === 'wan' || n.type === 'lan' ? 34 : 22,
          itemStyle: { color: catColors[n.category % catColors.length] },
        })),
        links: topoData.value.links,
      },
    ],
  }, true)
}

let resizeRaf = 0

watch(
  () => [props.initialSubResource, props.navSeq] as const,
  ([next]) => {
    if (currentTab.value !== 'traffic-topology') return
    if (next === 'topology' || next === 'analytics') {
      activeSub.value = next
    } else if (!next) {
      syncSubTab('traffic-topology', activeSub.value, { replace: true })
    }
  }
)

watch(activeSub, (tab) => {
  if (tab === 'analytics') void loadAnalytics()
  else void loadTopology()
})
watch(palette, () => {
  renderProtoChart()
  renderTopoChart()
})
watch(routerEpoch, () => {
  if (activeSub.value === 'analytics') void loadAnalytics()
  else void loadTopology()
})

onMounted(async () => {
  resizeObs = new ResizeObserver(() => {
    if (resizeRaf) cancelAnimationFrame(resizeRaf)
    resizeRaf = requestAnimationFrame(() => {
      protoChart?.resize()
      topoChart?.resize()
    })
  })
  if (activeSub.value === 'topology') {
    await loadTopology()
  } else {
    await loadAnalytics()
  }
})

onActivated(() => {
  if (currentTab.value === 'traffic-topology' && !props.initialSubResource) {
    syncSubTab('traffic-topology', activeSub.value, { replace: true })
  }
  void nextTick(() => {
    protoChart?.resize()
    topoChart?.resize()
  })
})

onUnmounted(() => {
  if (resizeRaf) cancelAnimationFrame(resizeRaf)
  resizeObs?.disconnect()
  protoChart?.dispose()
  topoChart?.dispose()
})
</script>

<style scoped>
.traffic-page { display: flex; flex-direction: column; gap: 16px; }
.ws-head { padding: 18px 20px; display: flex; flex-wrap: wrap; justify-content: space-between; align-items: center; gap: 12px; }
.ws-eyebrow { font-size: 11px; font-weight: 600; color: var(--text-faint); letter-spacing: 0.06em; }
.ws-head h1 { margin: 4px 0 0; font-size: 21px; font-weight: 800; }
.ws-head p { margin: 4px 0 0; font-size: 13px; color: var(--text-muted); }

.two-col { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 14px; }
.panel-box { padding: 16px; display: flex; flex-direction: column; gap: 14px; }
.box-head { display: flex; flex-wrap: wrap; justify-content: space-between; align-items: center; gap: 10px; }
.box-head h3 { margin: 0; font-size: 15px; font-weight: 700; }
.legend-row { display: flex; flex-wrap: wrap; gap: 6px; }

.chart-area { width: 100%; height: 290px; }
.topo-canvas { width: 100%; height: 520px; }

.rank-list { display: flex; flex-direction: column; gap: 12px; }
.rank-item { display: flex; flex-direction: column; gap: 5px; font-size: 12px; }
.rank-top { display: flex; justify-content: space-between; gap: 10px; }
.rank-name em { font-style: normal; color: var(--text-faint); margin-right: 6px; }
.bar-bg { height: 7px; border-radius: 999px; background: var(--surface-input); overflow: hidden; }
.bar-fill { height: 100%; border-radius: 999px; background: var(--accent-primary); }

.table-wrap { overflow-x: auto; border: 1px solid var(--divider-line); border-radius: 12px; }
.v4-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.v4-table th { text-align: left; padding: 10px 12px; font-size: 11px; font-weight: 600; color: var(--text-muted); background: var(--surface-table-head); border-bottom: 1px solid var(--divider-line); }
.v4-table td { padding: 10px 12px; border-bottom: 1px solid var(--divider-line); }
.ic-up { color: var(--chart-up); }
.ic-down { color: var(--chart-down); }
.btn-xs { min-height: 30px; padding: 0 10px; font-size: 12px; }
.empty-box { padding: 40px 12px; text-align: center; color: var(--text-muted); }
.quick-links-bar { display: flex; flex-wrap: wrap; gap: 8px; padding-top: 6px; }
.node-inspector {
  display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 10px;
  padding: 10px 14px; border-radius: var(--radius-md); border: 1px solid var(--border-subtle); background: var(--surface-input);
}
.node-info { display: flex; align-items: center; gap: 8px; font-size: 13px; }
.node-actions { display: flex; gap: 6px; }

@media (max-width: 960px) {
  .two-col { grid-template-columns: 1fr; }
}

@media (max-width: 640px) {
  .ws-head { padding: 14px; }
  .ws-head h1 { font-size: 18px; }
  .chart-area { height: 230px; }
  .topo-canvas { height: 320px; }
}
</style>

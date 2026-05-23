<template>
  <div class="insights-page">
    <div class="glass-panel page-head">
      <div>
        <h2 class="panel-title">监控分析</h2>
        <p class="panel-subtitle">系统状态、流量 Top 与风险摘要</p>
      </div>
      <button class="liquid-button refresh-button" type="button" :disabled="loading" @click="loadData">
        <RefreshCw :size="15" :class="{ spinning: loading }" />
        <span>刷新</span>
      </button>
    </div>

    <div v-if="error" class="error-banner">
      <span class="error-icon">!</span>
      <span>获取数据失败：{{ error }}</span>
    </div>

    <div class="metric-grid">
      <div class="glass-card metric-card">
        <Activity :size="19" />
        <span class="metric-label">当前上传</span>
        <strong class="font-mono">{{ formatSpeed(data?.summary.upload_speed || 0) }}</strong>
      </div>
      <div class="glass-card metric-card">
        <Gauge :size="19" />
        <span class="metric-label">当前下载</span>
        <strong class="font-mono">{{ formatSpeed(data?.summary.download_speed || 0) }}</strong>
      </div>
      <div class="glass-card metric-card">
        <Cable :size="19" />
        <span class="metric-label">连接数</span>
        <strong class="font-mono">{{ formatCountCompact(data?.summary.total_connections || 0) }}</strong>
      </div>
      <div class="glass-card metric-card">
        <Users :size="19" />
        <span class="metric-label">在线终端</span>
        <strong class="font-mono">{{ data?.summary.online_users || 0 }}</strong>
      </div>
    </div>

    <div class="analysis-grid">
      <section class="glass-panel section-panel">
        <div class="section-header">
          <h3>终端实时流量 Top</h3>
        </div>
        <div class="table-wrap">
          <table class="data-table">
            <thead>
              <tr>
                <th>终端</th>
                <th>IP</th>
                <th>上传</th>
                <th>下载</th>
                <th>连接</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="client in data?.top_clients" :key="client.mac">
                <td>
                  <strong>{{ client.hostname || client.comment || client.mac }}</strong>
                  <span class="muted font-mono">{{ client.mac }}</span>
                </td>
                <td class="font-mono">{{ client.ips[0] || '—' }}</td>
                <td class="font-mono metric-up">{{ formatSpeed(client.upload_speed) }}</td>
                <td class="font-mono metric-down">{{ formatSpeed(client.download_speed) }}</td>
                <td class="font-mono">{{ client.connections }}</td>
              </tr>
              <tr v-if="!data?.top_clients.length">
                <td colspan="5" class="empty-row">{{ loading ? '加载中...' : '暂无数据' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="glass-panel section-panel">
        <div class="section-header">
          <h3>接口实时流量 Top</h3>
        </div>
        <div class="table-wrap">
          <table class="data-table">
            <thead>
              <tr>
                <th>接口</th>
                <th>IP</th>
                <th>上传</th>
                <th>下载</th>
                <th>连接</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="iface in data?.top_interfaces" :key="iface.name">
                <td>
                  <strong class="font-mono">{{ iface.name }}</strong>
                  <span class="muted">{{ iface.comment || '—' }}</span>
                </td>
                <td class="font-mono">{{ iface.ip || '—' }}</td>
                <td class="font-mono metric-up">{{ formatSpeed(iface.upload_speed) }}</td>
                <td class="font-mono metric-down">{{ formatSpeed(iface.download_speed) }}</td>
                <td class="font-mono">{{ formatCountCompact(iface.connections) }}</td>
              </tr>
              <tr v-if="!data?.top_interfaces.length">
                <td colspan="5" class="empty-row">{{ loading ? '加载中...' : '暂无数据' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>

    <section class="glass-panel section-panel">
      <div class="section-header">
        <h3>风险摘要</h3>
      </div>
      <div class="risk-grid">
        <div class="risk-block">
          <div class="risk-title">
            <ShieldAlert :size="16" />
            <span>高风险端口映射</span>
          </div>
          <div v-if="!data?.high_risk_mappings.length" class="empty-inline">暂无高风险端口</div>
          <div v-for="item in data?.high_risk_mappings" :key="`${item.int_ip}:${item.int_port}:${item.ext_port}`" class="risk-item">
            <span>{{ item.name }}</span>
            <strong class="font-mono">{{ item.ext_port }} → {{ item.int_ip }}:{{ item.int_port }}</strong>
          </div>
        </div>
        <div class="risk-block">
          <div class="risk-title">
            <TriangleAlert :size="16" />
            <span>异常终端</span>
          </div>
          <div v-if="!data?.abnormal_clients.length" class="empty-inline">暂无异常终端</div>
          <div v-for="item in data?.abnormal_clients" :key="item.mac" class="risk-item">
            <span>{{ item.hostname || item.comment || item.mac }}</span>
            <strong class="font-mono">{{ item.connections }} 连接 / {{ formatSpeed(item.upload_speed) }}</strong>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onActivated, onMounted, ref } from 'vue'
import { Activity, Cable, Gauge, RefreshCw, ShieldAlert, TriangleAlert, Users } from 'lucide-vue-next'
import { fetchMonitorInsights, type MonitorInsightsData } from '@/api/monitor'
import { formatCountCompact, formatSpeed } from '@/composables/useFormatters'

const data = ref<MonitorInsightsData | null>(null)
const loading = ref(false)
const error = ref('')

async function loadData() {
  loading.value = true
  error.value = ''
  try {
    data.value = await fetchMonitorInsights()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '请求错误'
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
onActivated(loadData)
</script>

<style scoped>
.insights-page {
  display: flex;
  flex-direction: column;
  gap: 22px;
}

.page-head,
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 18px 20px;
}

.panel-title {
  color: var(--text-primary);
  font-size: 1.25rem;
  font-weight: 720;
}

.panel-subtitle,
.muted {
  color: var(--text-secondary);
  font-size: 12px;
}

.refresh-button {
  min-height: 38px;
  padding: 0 14px;
}

.spinning {
  animation: spin 0.8s linear infinite;
}

.metric-grid,
.analysis-grid,
.risk-grid {
  display: grid;
  gap: 16px;
}

.metric-grid {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.analysis-grid,
.risk-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.metric-card {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 9px;
  padding: 18px;
}

.metric-card svg {
  color: var(--system-blue);
}

.metric-label {
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 700;
}

.metric-card strong {
  color: var(--text-primary);
  font-size: 24px;
  font-weight: 560;
}

.section-panel {
  overflow: hidden;
}

.section-header {
  border-bottom: 1px solid var(--glass-border);
}

.section-header h3 {
  color: var(--text-primary);
  font-size: 15px;
  font-weight: 720;
}

.table-wrap {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 620px;
}

.data-table th,
.data-table td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid var(--control-border);
  vertical-align: top;
}

.data-table th {
  color: var(--text-secondary);
  background: var(--table-header-bg);
  font-size: 12px;
  font-weight: 620;
}

.data-table td {
  color: var(--text-primary);
  font-size: 13px;
}

.data-table td:first-child {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.metric-up {
  color: var(--system-green);
}

.metric-down {
  color: var(--system-blue);
}

.risk-block {
  padding: 18px;
}

.risk-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  color: var(--text-primary);
  font-weight: 720;
}

.risk-title svg {
  color: var(--system-orange);
}

.risk-item {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 11px 0;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--control-border);
}

.risk-item strong {
  color: var(--text-primary);
  text-align: right;
}

.empty-row,
.empty-inline {
  color: var(--text-tertiary);
  text-align: center;
}

.empty-row {
  padding: 32px !important;
}

.empty-inline {
  padding: 18px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 960px) {
  .metric-grid,
  .analysis-grid,
  .risk-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 560px) {
  .page-head {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>

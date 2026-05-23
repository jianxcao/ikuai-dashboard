<template>
  <div class="tables-container">
    <!-- WAN 接口状态 -->
    <div class="table-section glass-card">
      <div class="section-header">
        <span class="section-title">
          <span class="title-dot"></span>
          WAN 接口状态
        </span>
      </div>

      <table class="data-table">
        <thead>
          <tr>
            <th>连接名称</th>
            <th>IP 地址</th>
            <th>联网方式</th>
            <th>状态</th>
            <th>备注</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="wan in wanStatus" :key="wan.name" class="table-row">
            <td class="font-mono table-name">{{ wan.name }}</td>
            <td class="font-mono metric-up">{{ wan.ip || '—' }}</td>
            <td>
              <span class="proto-badge">{{ wan.proto || '—' }}</span>
            </td>
            <td>
              <span :class="['status-badge', mapStatus(wan.status)]">
                <span class="status-dot"></span>
                {{ wan.status }}
              </span>
            </td>
            <td class="table-muted">{{ wan.comment || '—' }}</td>
          </tr>
          <tr v-if="!wanStatus.length">
            <td colspan="5" class="empty-row">暂无数据</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 接口流量详情 -->
    <div class="table-section glass-card traffic-section">
      <div class="section-header">
        <span class="section-title">
          <span class="title-dot title-dot-blue"></span>
          接口流量详情
        </span>
      </div>

      <table class="data-table">
        <thead>
          <tr>
            <th>连接名称</th>
            <th>IP 地址</th>
            <th>当前上传</th>
            <th>当前下载</th>
            <th>总上传</th>
            <th>总下载</th>
            <th>连接数</th>
            <th>备注</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="iface in trafficDetails" :key="iface.name" class="table-row">
            <td class="font-mono table-name">{{ iface.name }}</td>
            <td class="font-mono metric-up">{{ iface.ip || '—' }}</td>
            <td class="font-mono metric-up">
              {{ iface.upload_speed > 0 ? formatSpeed(iface.upload_speed) : '0' }}
            </td>
            <td class="font-mono metric-down">
              {{ iface.download_speed > 0 ? formatSpeed(iface.download_speed) : '0' }}
            </td>
            <td class="font-mono table-muted">{{ formatBytes(iface.total_up) }}</td>
            <td class="font-mono table-muted">{{ formatBytes(iface.total_down) }}</td>
            <td class="font-mono table-soft">{{ formatCountCompact(iface.connections) }}</td>
            <td class="table-muted">{{ iface.comment || '—' }}</td>
          </tr>
          <tr v-if="!trafficDetails.length">
            <td colspan="8" class="empty-row">暂无数据</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatSpeed, formatBytes, formatCountCompact, mapStatus } from '@/composables/useFormatters'
import type { IKuaiWanStatus, IKuaiTraffic } from '@/api/monitor'

withDefaults(defineProps<{
  wanStatus?: IKuaiWanStatus[]
  trafficDetails?: IKuaiTraffic[]
}>(), {
  wanStatus: () => [],
  trafficDetails: () => []
})
</script>

<style scoped>
.table-section {
  padding: 0;
  overflow: hidden;
}

.traffic-section {
  margin-top: 24px;
}

.section-header {
  padding: 18px 24px;
  border-bottom: 1px solid var(--glass-border);
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
  background: var(--system-green);
  box-shadow: 0 0 12px var(--system-green-dim);
}

.title-dot-blue {
  background: var(--system-blue);
  box-shadow: 0 0 12px var(--system-blue-dim);
}

/* ── 表格 ── */
.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th {
  padding: 10px 24px;
  text-align: left;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  background: var(--table-header-bg);
  border-bottom: 1px solid var(--glass-border);
}

.table-row td {
  padding: 12px 24px;
  font-size: 13px;
  border-bottom: 1px solid var(--control-border);
  vertical-align: middle;
  transition: all 0.2s ease;
}

.table-row:last-child td {
  border-bottom: none;
}

.table-row:hover td {
  background: var(--table-row-hover);
}

.table-name {
  color: var(--text-primary);
  font-weight: 650;
}

.metric-up {
  color: var(--system-green);
}

.metric-down {
  color: var(--system-blue);
}

.table-muted {
  color: var(--text-secondary);
}

.table-soft {
  color: var(--text-tertiary);
}

/* 状态徽章 */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
  padding: 4px 10px;
  border-radius: 20px;
}

.status-badge.success {
  color: var(--system-green);
  background: var(--system-green-dim);
}

.status-badge.error {
  color: var(--system-red);
  background: var(--system-red-dim);
}

.status-badge.warning {
  color: var(--system-orange);
  background: var(--system-orange-dim);
}

.status-badge .status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

/* 协议标签 */
.proto-badge {
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 6px;
  background: var(--system-blue-dim);
  color: var(--system-blue);
  font-weight: 500;
}

.empty-row {
  text-align: center;
  padding: 40px !important;
  color: var(--text-tertiary);
  font-size: 14px;
}

@media (max-width: 900px) {
  .table-section {
    overflow-x: auto;
  }

  .data-table {
    min-width: 760px;
  }
}

@media (max-width: 560px) {
  .section-header {
    padding: 16px;
  }

  .data-table th,
  .table-row td {
    padding: 10px 14px;
  }
}
</style>

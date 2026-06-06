<template>
  <div class="tables-stack">
    <section class="table-section glass-card">
      <div class="section-header">
        <span class="section-title">
          <span class="title-dot"></span>
          WAN 接口状态
        </span>
      </div>

      <div class="desktop-table">
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

      <div class="mobile-cards">
        <article v-for="wan in wanStatus" :key="`wan-${wan.name}`" class="compact-card">
          <div class="compact-head">
            <strong class="font-mono">{{ wan.name }}</strong>
            <span :class="['status-badge', mapStatus(wan.status)]">
              <span class="status-dot"></span>
              {{ wan.status }}
            </span>
          </div>
          <dl class="compact-grid">
            <div>
              <dt>IP</dt>
              <dd class="font-mono">{{ wan.ip || '—' }}</dd>
            </div>
            <div>
              <dt>协议</dt>
              <dd>
                <span class="proto-badge">{{ wan.proto || '—' }}</span>
              </dd>
            </div>
            <div class="full">
              <dt>备注</dt>
              <dd>{{ wan.comment || '—' }}</dd>
            </div>
          </dl>
        </article>
        <div v-if="!wanStatus.length" class="empty-row mobile-empty">暂无数据</div>
      </div>
    </section>

    <section class="table-section glass-card">
      <div class="section-header">
        <span class="section-title">
          <span class="title-dot title-dot-blue"></span>
          接口流量详情
        </span>
      </div>

      <div class="desktop-table">
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

      <div class="mobile-cards">
        <article
          v-for="iface in trafficDetails"
          :key="`traffic-${iface.name}`"
          class="compact-card"
        >
          <div class="compact-head">
            <strong class="font-mono">{{ iface.name }}</strong>
            <span class="font-mono compact-pill">{{ iface.ip || '—' }}</span>
          </div>
          <dl class="compact-grid">
            <div>
              <dt>当前上传</dt>
              <dd class="font-mono metric-up">{{ formatSpeed(iface.upload_speed) }}</dd>
            </div>
            <div>
              <dt>当前下载</dt>
              <dd class="font-mono metric-down">{{ formatSpeed(iface.download_speed) }}</dd>
            </div>
            <div>
              <dt>总上传</dt>
              <dd class="font-mono">{{ formatBytes(iface.total_up) }}</dd>
            </div>
            <div>
              <dt>总下载</dt>
              <dd class="font-mono">{{ formatBytes(iface.total_down) }}</dd>
            </div>
            <div>
              <dt>连接数</dt>
              <dd class="font-mono">{{ formatCountCompact(iface.connections) }}</dd>
            </div>
            <div class="full">
              <dt>备注</dt>
              <dd>{{ iface.comment || '—' }}</dd>
            </div>
          </dl>
        </article>
        <div v-if="!trafficDetails.length" class="empty-row mobile-empty">暂无数据</div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import {
  formatSpeed,
  formatBytes,
  formatCountCompact,
  mapStatus
} from '@/composables/useFormatters'
import type { IKuaiWanStatus, IKuaiTraffic } from '@/api/monitor'

withDefaults(
  defineProps<{
    wanStatus?: IKuaiWanStatus[]
    trafficDetails?: IKuaiTraffic[]
  }>(),
  {
    wanStatus: () => [],
    trafficDetails: () => []
  }
)
</script>

<style scoped>
.tables-stack {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.table-section {
  padding: 0;
  overflow: hidden;
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
  font-weight: 700;
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

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th {
  padding: 10px 24px;
  text-align: left;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  background: var(--table-header-bg);
  border-bottom: 1px solid var(--glass-border);
}

.table-row td {
  padding: 13px 24px;
  font-size: 13px;
  border-bottom: 1px solid var(--control-border);
  vertical-align: middle;
  transition: background 0.2s ease;
}

.table-row:last-child td {
  border-bottom: none;
}

.table-row:hover td {
  background: var(--table-row-hover);
}

.table-name {
  color: var(--text-primary);
  font-weight: 700;
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

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 600;
  padding: 5px 10px;
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
  box-shadow: none;
}

.proto-badge,
.compact-pill {
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 999px;
  background: var(--system-blue-dim);
  color: var(--system-blue);
  font-weight: 600;
}

.empty-row {
  text-align: center;
  padding: 40px !important;
  color: var(--text-tertiary);
  font-size: 14px;
}

.mobile-cards {
  display: none;
}

@media (max-width: 760px) {
  .desktop-table {
    display: none;
  }

  .mobile-cards {
    display: grid;
    gap: 12px;
    padding: 14px;
  }

  .compact-card {
    padding: 16px;
    border: 1px solid var(--control-border);
    border-radius: 18px;
    background: var(--control-bg);
  }

  .compact-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 10px;
    margin-bottom: 14px;
  }

  .compact-head strong {
    color: var(--text-primary);
    font-size: 14px;
  }

  .compact-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px 14px;
  }

  .compact-grid .full {
    grid-column: 1 / -1;
  }

  .compact-grid dt {
    margin-bottom: 5px;
    color: var(--text-tertiary);
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.04em;
  }

  .compact-grid dd {
    margin: 0;
    color: var(--text-primary);
    font-size: 13px;
    line-height: 1.5;
    overflow-wrap: anywhere;
  }

  .mobile-empty {
    padding: 24px !important;
  }
}

@media (max-width: 560px) {
  .section-header {
    padding: 16px 18px;
  }

  .compact-grid {
    grid-template-columns: 1fr;
  }
}
</style>

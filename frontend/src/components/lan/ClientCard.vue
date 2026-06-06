<template>
  <div class="client-card client-grid">
    <div class="card-left">
      <div class="device-info">
        <div class="device-name">{{ client.comment || client.hostname || '未知设备' }}</div>
        <div class="device-meta">
          <button
            type="button"
            class="mac-chip"
            @click="copyToClipboard(client.mac)"
            title="点击复制 MAC"
          >
            <Cpu :size="13" class="mac-icon" />
            <span class="font-mono">{{ client.mac }}</span>
          </button>
          <span class="meta-tag">{{ client.client_type || 'Unknown' }}</span>
          <span class="meta-tag">在线: {{ formatDurationCompact(client.uptime) }}</span>
        </div>
      </div>
    </div>

    <div class="card-ips">
      <div class="ip-row v4">
        <span class="ip-label">IPv4</span>
        <span class="font-mono ip-address">{{ ipv4 }}</span>
      </div>
      <div v-if="ipv6" class="ip-row v6">
        <span class="ip-label">IPv6</span>
        <span class="font-mono ip-address">{{ ipv6 }}</span>
      </div>
    </div>

    <div class="metric-cell">
      <span class="metric-label">上传</span>
      <span class="font-mono speed up">{{ formatSpeed(client.upload_speed) }}</span>
    </div>

    <div class="metric-cell">
      <span class="metric-label">下载</span>
      <span class="font-mono speed down">{{ formatSpeed(client.download_speed) }}</span>
    </div>

    <div class="metric-cell">
      <span class="metric-label">连接数</span>
      <span class="font-mono conns">{{ formatCountCompact(client.connections) }}</span>
    </div>

    <div class="card-totals">
      <div class="font-mono total-sum">{{ formatBytes(totalBytes) }}</div>
      <div class="total-breakdown">
        <span class="font-mono"
          ><span class="total-label">↑</span>{{ formatBytes(client.total_up) }}</span
        >
        <span class="font-mono"
          ><span class="total-label">↓</span>{{ formatBytes(client.total_down) }}</span
        >
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Cpu } from 'lucide-vue-next'
import {
  formatBytes,
  formatCountCompact,
  formatDurationCompact,
  formatSpeed,
  getIPv4,
  getIPv6
} from '@/composables/useFormatters'
import type { ClientDTO } from '@/api/monitor'

const props = defineProps<{ client: ClientDTO }>()

const ipv4 = computed(() => getIPv4(props.client.ips))
const ipv6 = computed(() => getIPv6(props.client.ips))
const totalBytes = computed(() => (props.client.total_up || 0) + (props.client.total_down || 0))

function copyToClipboard(text: string) {
  void navigator.clipboard.writeText(text)
}
</script>

<style scoped>
.client-card {
  padding: 16px 24px;
  position: relative;
  overflow: hidden;
  border-bottom: 1px solid var(--control-border);
  transition: background 0.2s ease;
  flex-shrink: 0;
}

.client-card:last-child {
  border-bottom: none;
}

.client-card:hover {
  background: var(--table-row-hover);
}

.card-left {
  min-width: 0;
}

.mac-chip {
  appearance: none;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  max-width: 100%;
  min-height: 26px;
  padding: 0 9px;
  background: var(--control-bg);
  border: 1px solid var(--control-border);
  border-radius: 999px;
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.mac-chip:hover {
  background: var(--control-bg-hover);
  color: var(--text-primary);
  border-color: var(--glass-border-light);
}

.device-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.device-name {
  font-size: 15px;
  font-weight: 720;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  letter-spacing: 0;
}

.device-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  overflow: hidden;
  margin-top: 8px;
}

.meta-tag {
  font-size: 11px;
  min-height: 24px;
  padding: 0 8px;
  background: var(--control-bg);
  border: 1px solid var(--control-border);
  color: var(--text-secondary);
  border-radius: 999px;
  font-weight: 650;
  line-height: 22px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-ips {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.ip-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  min-width: 0;
  color: var(--text-primary);
  font-size: 13px;
  line-height: 1.35;
}

.ip-label {
  font-size: 11px;
  min-width: 40px;
  padding: 3px 6px;
  border-radius: 4px;
  font-weight: 600;
  flex-shrink: 0;
}

.ip-address {
  min-width: 0;
  overflow-wrap: anywhere;
  word-break: break-word;
}

.ip-row.v4 .ip-label {
  background: var(--system-blue-dim);
  color: var(--system-blue);
}
.ip-row.v6 .ip-label {
  background: var(--system-purple-dim);
  color: var(--system-purple);
}
.ip-row.v6 .ip-address {
  color: var(--text-secondary);
  font-size: 12px;
}

.metric-cell {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  min-width: 0;
  text-align: right;
}

.metric-label {
  display: none;
}

.metric-cell .speed,
.metric-cell .conns {
  font-size: 15px;
  font-weight: 720;
  white-space: nowrap;
}

.metric-cell .speed.up {
  color: var(--system-green);
}
.metric-cell .speed.down {
  color: var(--system-blue);
}
.metric-cell .conns {
  color: var(--text-primary);
}

.card-totals {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
  text-align: right;
}

.total-sum {
  color: var(--text-primary);
  font-size: 15px;
  font-weight: 720;
  white-space: nowrap;
}

.total-breakdown {
  display: grid;
  gap: 3px;
  font-size: 11px;
  color: var(--text-primary);
}

.total-label {
  margin-right: 4px;
  color: var(--text-tertiary);
}

@media (max-width: 920px) {
  .client-card {
    display: grid;
    grid-template-columns: 1fr;
    gap: 14px;
    padding: 18px;
    border: 1px solid var(--control-border);
    border-bottom: 1px solid var(--control-border);
    border-radius: var(--radius-md);
    background: var(--control-bg);
  }

  .card-left,
  .card-ips,
  .metric-cell,
  .card-totals {
    width: 100%;
  }

  .metric-cell {
    justify-content: space-between;
    min-height: 28px;
    text-align: left;
  }

  .metric-label {
    display: inline;
    color: var(--text-tertiary);
    font-size: 12px;
    font-weight: 700;
  }

  .card-totals {
    padding-top: 12px;
    border-top: 1px solid var(--control-border);
    text-align: left;
  }

  .total-breakdown {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 460px) {
  .device-name {
    white-space: normal;
  }
}
</style>

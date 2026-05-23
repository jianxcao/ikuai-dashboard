<template>
  <div class="client-card glass-card">
    <div class="card-left">
      <!-- MAC 芯片样式 -->
      <div class="mac-chip" @click="copyToClipboard(client.mac)" title="点击复制 MAC">
        <Cpu :size="14" class="mac-icon" />
        <span class="font-mono">{{ client.mac }}</span>
      </div>

      <!-- 设备标识信息 -->
      <div class="device-info">
        <div class="device-name">{{ client.comment || client.hostname || '未知设备' }}</div>
        <div class="device-meta">
          <span class="meta-tag">{{ client.client_type || 'Unknown' }}</span>
          <span class="meta-tag">在线: {{ formatDurationCompact(client.uptime) }}</span>
        </div>
      </div>
    </div>

    <!-- IP 双栈信息 -->
    <div class="card-ips">
      <div class="ip-row v4">
        <span class="ip-label">IPv4</span>
        <span class="font-mono">{{ ipv4 }}</span>
      </div>
      <div v-if="ipv6" class="ip-row v6">
        <span class="ip-label">IPv6</span>
        <span class="font-mono">{{ ipv6 }}</span>
      </div>
    </div>

    <!-- 实时速率与连接数 -->
    <div class="card-metrics">
      <div class="metric-col" style="width: 100px">
        <span class="font-mono speed up">{{ formatSpeed(client.upload_speed) }}</span>
      </div>
      <div class="metric-col" style="width: 100px">
        <span class="font-mono speed down">{{ formatSpeed(client.download_speed) }}</span>
      </div>
      <div class="metric-col" style="width: 80px">
        <span class="font-mono conns">{{ formatCountCompact(client.connections) }}</span>
      </div>
    </div>

    <!-- 流量消耗进度条形式展示 (可选，这里用数字) -->
    <div class="card-totals">
      <div class="total-row">
        <span class="total-label">↑</span>
        <span class="font-mono">{{ formatBytes(client.total_up) }}</span>
      </div>
      <div class="total-row">
        <span class="total-label">↓</span>
        <span class="font-mono">{{ formatBytes(client.total_down) }}</span>
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
  getIPv6,
} from '@/composables/useFormatters'
import type { ClientDTO } from '@/api/monitor'

const props = defineProps<{ client: ClientDTO }>()

const ipv4 = computed(() => getIPv4(props.client.ips))
const ipv6 = computed(() => getIPv6(props.client.ips))

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
}
</script>

<style scoped>
.client-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  gap: 24px;
  position: relative;
  overflow: hidden;
  border-bottom: 1px solid var(--control-border);
  transition: all 0.2s ease;
  flex-shrink: 0; /* 防止行在列表中被挤压 */
}

.client-card:last-child {
  border-bottom: none;
}

.client-card:hover {
  background: var(--table-row-hover);
}

/* 左侧：MAC 与 标识 */
.card-left {
  display: flex;
  align-items: center;
  gap: 16px;
  width: 260px;
  flex-shrink: 0;
}

.mac-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: var(--control-bg);
  border: 1px solid var(--control-border);
  border-radius: 8px;
  color: var(--text-secondary);
  font-size: 13px;
  font-family: ui-monospace, SFMono-Regular, monospace;
  cursor: pointer;
  transition: all 0.2s ease;
}

.mac-chip:hover {
  background: var(--control-bg-hover);
  color: var(--text-primary);
  border-color: var(--glass-border-light);
}

.device-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.device-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  letter-spacing: 0.01em;
}

.device-meta {
  display: flex;
  gap: 6px;
  flex-wrap: nowrap;
  overflow: hidden;
}

.meta-tag {
  font-size: 11px;
  padding: 3px 8px;
  background: var(--control-bg);
  border: 1px solid var(--control-border);
  color: var(--text-secondary);
  border-radius: 6px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex-shrink: 0;
}

/* 中间：IP 信息 */
.card-ips {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 280px;
  flex-shrink: 0;
  overflow: hidden;
}

.ip-row {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  font-family: ui-monospace, SFMono-Regular, monospace;
  overflow: hidden;
}

.ip-label {
  font-size: 11px;
  padding: 3px 6px;
  border-radius: 4px;
  font-weight: 600;
  flex-shrink: 0;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
}

.ip-row.v4 span.font-mono { color: var(--text-primary); }

.ip-row.v4 .ip-label { background: var(--system-blue-dim); color: var(--system-blue); }
.ip-row.v6 .ip-label { background: var(--system-purple-dim); color: var(--system-purple); }
.ip-row.v6 span.font-mono { 
  color: var(--text-secondary); 
  font-size: 12px; 
  white-space: nowrap; 
  overflow: hidden; 
  text-overflow: ellipsis; 
}

/* 速率与连接 */
.card-metrics {
  display: flex;
  gap: 16px;
  flex: 1;
  align-items: center;
}

.metric-col {
  display: flex;
  align-items: center;
}

.metric-col .speed {
  font-size: 15px;
  font-weight: 600;
  font-family: ui-monospace, SFMono-Regular, monospace;
}
.metric-col .speed.up { color: var(--system-green); }
.metric-col .speed.down { color: var(--system-blue); }
.metric-col .conns { color: var(--text-primary); font-family: ui-monospace, SFMono-Regular, monospace; font-size: 15px; font-weight: 600;}

/* 总计消耗 */
.card-totals {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 120px;
  flex-shrink: 0;
  text-align: right;
}

.total-row {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  font-size: 13px;
  color: var(--text-primary);
  font-family: ui-monospace, SFMono-Regular, monospace;
}

.total-label {
  color: var(--text-tertiary);
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
}

@media (max-width: 920px) {
  .client-card {
    align-items: flex-start;
    flex-direction: column;
    gap: 14px;
    padding: 18px;
  }

  .card-left,
  .card-ips,
  .card-metrics,
  .card-totals {
    width: 100%;
  }

  .card-left {
    align-items: flex-start;
    flex-direction: column;
  }

  .card-metrics {
    justify-content: space-between;
  }

  .card-totals {
    text-align: left;
  }

  .total-row {
    justify-content: flex-start;
  }
}
</style>

<template>
  <OverviewSectionShell
    title="WAN 线路健康检查"
    :layoutLocked="layoutLocked"
    :canMovePrev="canMovePrev"
    :canMoveNext="canMoveNext"
    @move="emit('move', $event)"
  >
    <template #icon>
      <Globe :size="16" />
    </template>
    <template #actions>
      <div class="pill-group">
        <button
          type="button"
          class="mini-pill"
          @click="emit('navigate', { tab: 'routing', subResource: 'stream-domain' })"
        >
          分流策略
        </button>
        <button
          type="button"
          class="mini-pill"
          @click="emit('navigate', { tab: 'network', subResource: 'dnat-rules' })"
        >
          DNAT 映射
        </button>
      </div>
    </template>

    <div class="wan-list">
      <div v-for="w in wanChecks" :key="w.interface" class="wan-row">
        <div class="wan-head">
          <div class="wan-title">
            <strong class="font-mono">{{ w.interface }}</strong>
            <span class="v4-badge">{{ w.internet }}</span>
            <span v-if="w.comment" class="wan-comment">{{ w.comment }}</span>
          </div>
          <span class="v4-badge" :class="w.result === 'success' ? 'v4-badge-up' : 'v4-badge-down'">
            {{ w.errmsg || w.result }}
          </span>
        </div>
        <div class="wan-meta font-mono">
          <span>IP {{ w.ip_addr || '—' }}</span>
          <span>GW {{ w.gateway || '—' }}</span>
          <span>RTT {{ w.avg_rtt_ms.toFixed(2) }}ms</span>
          <span>丢包 {{ w.drop_rate.toFixed(2) }}%</span>
        </div>
      </div>
    </div>
  </OverviewSectionShell>
</template>

<script setup lang="ts">
import { Globe } from 'lucide-vue-next'
import type { OverviewDTO } from '@/api/v4'
import OverviewSectionShell from './OverviewSectionShell.vue'

defineProps<{
  wanChecks: OverviewDTO['wan_checks']
  layoutLocked: boolean
  canMovePrev: boolean
  canMoveNext: boolean
}>()

const emit = defineEmits<{
  (e: 'move', delta: -1 | 1): void
  (e: 'navigate', payload: { tab: string; subResource?: string }): void
}>()
</script>

<style scoped>
.pill-group {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.mini-pill {
  min-height: 28px;
  padding: 0 10px;
  border-radius: 8px;
  border: 1px solid var(--divider-line);
  background: var(--surface-btn);
  color: var(--text-normal);
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
}

.mini-pill:hover {
  border-color: var(--accent-primary);
  color: var(--text-strong);
}

.wan-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.wan-row {
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid var(--divider-line);
  background: var(--surface-input);
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.wan-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.wan-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.wan-comment {
  font-size: 12px;
  color: var(--text-muted);
}

.wan-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  font-size: 11px;
  color: var(--text-muted);
}
</style>

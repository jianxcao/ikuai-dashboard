<template>
  <OverviewSectionShell
    title="计算与内存负载"
    :layoutLocked="layoutLocked"
    :canMovePrev="canMovePrev"
    :canMoveNext="canMoveNext"
    @move="emit('move', $event)"
  >
    <template #icon>
      <Cpu :size="16" />
    </template>
    <template #actions>
      <span v-if="overview.cpu?.freq_mhz?.length" class="v4-badge font-mono">
        {{ overview.cpu.freq_mhz[0] }} MHz
      </span>
    </template>

    <div class="hw-split">
      <div class="hw-block">
        <div class="hw-label-row">
          <span>CPU 综合负载</span>
          <strong class="font-mono">{{ (overview.cpu?.avg_percent ?? 0).toFixed(2) }}%</strong>
        </div>
        <div class="core-bars">
          <div v-for="(val, idx) in overview.cpu?.cores || []" :key="idx" class="core-row">
            <span class="font-mono core-idx">Core {{ idx }}</span>
            <div class="bar-track">
              <div class="bar-fill accent" :style="{ width: `${Math.min(100, Math.max(2, val))}%` }" />
            </div>
            <span class="font-mono core-val">{{ val.toFixed(1) }}%</span>
          </div>
        </div>
      </div>

      <div class="hw-block">
        <div class="hw-label-row">
          <span>物理内存占用</span>
          <strong class="font-mono">{{ overview.memory.used_percent.toFixed(0) }}%</strong>
        </div>
        <div class="bar-track lg">
          <div class="bar-fill up" :style="{ width: `${Math.min(100, overview.memory.used_percent)}%` }" />
        </div>
        <div class="mem-details font-mono">
          <span>总容量 {{ formatBytes(overview.memory.total_kb * 1024) }}</span>
          <span>可用 {{ formatBytes(overview.memory.available_kb * 1024) }}</span>
          <span>缓存 {{ formatBytes(overview.memory.cached_kb * 1024) }}</span>
          <span>缓冲 {{ formatBytes(overview.memory.buffers_kb * 1024) }}</span>
        </div>
      </div>
    </div>
  </OverviewSectionShell>
</template>

<script setup lang="ts">
import { Cpu } from 'lucide-vue-next'
import type { OverviewDTO } from '@/api/v4'
import { formatBytes } from '@/composables/useFormatters'
import OverviewSectionShell from './OverviewSectionShell.vue'

defineProps<{
  overview: OverviewDTO
  layoutLocked: boolean
  canMovePrev: boolean
  canMoveNext: boolean
}>()

const emit = defineEmits<{
  (e: 'move', delta: -1 | 1): void
}>()
</script>

<style scoped>
.hw-split {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.hw-block {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.hw-label-row {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--text-normal);
}

.core-bars {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.core-row {
  display: grid;
  grid-template-columns: 48px 1fr 44px;
  align-items: center;
  gap: 8px;
  font-size: 11px;
}

.core-idx,
.core-val {
  color: var(--text-muted);
}

.bar-track {
  height: 7px;
  border-radius: 999px;
  background: var(--surface-input);
  overflow: hidden;
}

.bar-track.lg {
  height: 11px;
}

.bar-fill {
  height: 100%;
  border-radius: 999px;
  transition: width 0.3s ease;
}

.bar-fill.accent { background: var(--accent-primary); }
.bar-fill.up { background: var(--status-up); }

.mem-details {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
  font-size: 11px;
  color: var(--text-muted);
}

@media (max-width: 1180px) {
  .hw-split {
    grid-template-columns: 1fr;
  }
}
</style>

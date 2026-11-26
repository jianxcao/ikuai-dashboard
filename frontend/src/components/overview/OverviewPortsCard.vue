<template>
  <OverviewSectionShell
    title="物理网口面板 (ether_info)"
    :layoutLocked="layoutLocked"
    :canMovePrev="canMovePrev"
    :canMoveNext="canMoveNext"
    @move="emit('move', $event)"
  >
    <template #icon>
      <Cable :size="16" />
    </template>
    <template #actions>
      <div class="pill-group">
        <span class="v4-badge font-mono">{{ ports.length }} Ports</span>
        <button
          type="button"
          class="mini-pill"
          @click="emit('navigate', { tab: 'network', subResource: 'physical-nics' })"
        >
          网口配置
        </button>
      </div>
    </template>

    <div class="port-matrix">
      <div
        v-for="port in ports"
        :key="port.name"
        class="port-slot"
        :class="{ linked: port.link === 1 }"
      >
        <div class="port-top">
          <span class="port-led" :class="{ on: port.link === 1 }" />
          <strong class="font-mono">{{ port.name }}</strong>
          <span class="v4-badge" :class="port.interface && port.interface !== 'Null' ? 'v4-badge-accent' : ''">
            {{ port.interface && port.interface !== 'Null' ? port.interface : '未分配' }}
          </span>
        </div>
        <div class="port-mid font-mono">
          <span>{{ port.speed > 0 ? `${port.speed} Mbps` : port.link === 1 ? '已连接' : '未连接' }}</span>
          <span>{{ port.type || 'NIC' }}</span>
        </div>
        <div class="port-sub font-mono" :title="port.model">
          {{ port.mac }} · {{ port.driver }}
        </div>
      </div>
    </div>
  </OverviewSectionShell>
</template>

<script setup lang="ts">
import { Cable } from 'lucide-vue-next'
import type { OverviewDTO } from '@/api/v4'
import OverviewSectionShell from './OverviewSectionShell.vue'

defineProps<{
  ports: OverviewDTO['physical_ports']
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

.port-matrix {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(185px, 1fr));
  gap: 10px;
}

.port-slot {
  padding: 10px 12px;
  border-radius: 12px;
  border: 1px solid var(--divider-line);
  background: var(--surface-input);
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.port-slot.linked {
  border-color: color-mix(in srgb, var(--status-up) 30%, var(--divider-line));
}

.port-top {
  display: flex;
  align-items: center;
  gap: 7px;
}

.port-led {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--text-faint);
}

.port-led.on {
  background: var(--status-up);
  box-shadow: 0 0 6px color-mix(in srgb, var(--status-up) 55%, transparent);
}

.port-mid {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: var(--text-normal);
}

.port-sub {
  font-size: 10px;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

@media (max-width: 600px) {
  .port-matrix {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
  }
}
</style>

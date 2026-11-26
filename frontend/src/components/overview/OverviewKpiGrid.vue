<template>
  <div class="kpi-grid">
    <article
      v-for="(kpiId, kpiIdx) in kpiOrder"
      :key="kpiId"
      class="kpi-card v4-card"
      :class="{
        clickable: layoutLocked && (kpiId === 'connections' || kpiId === 'users'),
        'layout-editable': !layoutLocked,
        'is-dragging': draggingId === kpiId,
        'is-drop-target': overId === kpiId && draggingId !== kpiId,
      }"
      :draggable="!layoutLocked"
      :title="kpiCardTitle(kpiId)"
      @click="handleKpiClick(kpiId)"
      @dragstart="emit('drag-start', { id: kpiId, event: $event })"
      @dragover.prevent
      @dragenter.prevent="emit('drag-enter', kpiId)"
      @dragleave="emit('drag-leave', { id: kpiId, event: $event })"
      @drop.prevent="emit('drop', kpiId)"
      @dragend="emit('drag-end')"
    >
      <div v-if="!layoutLocked" class="card-drag-bar">
        <span class="drag-handle-pill">
          <GripVertical :size="12" />
          <span>拖动互换</span>
        </span>
        <div class="step-btns">
          <button
            type="button"
            class="drag-step-btn"
            title="左移"
            :disabled="kpiIdx === 0"
            @click.stop="emit('move', { id: kpiId, delta: -1 })"
          >
            <ArrowLeft :size="11" />
          </button>
          <button
            type="button"
            class="drag-step-btn"
            title="右移"
            :disabled="kpiIdx === kpiOrder.length - 1"
            @click.stop="emit('move', { id: kpiId, delta: 1 })"
          >
            <ArrowRight :size="11" />
          </button>
        </div>
      </div>

      <template v-if="kpiId === 'upload'">
        <div class="kpi-head">
          <span>实时上行速率</span>
          <ArrowUpRight :size="16" class="ic-up" />
        </div>
        <div class="kpi-value font-mono">
          <strong>{{ upParts.value }}</strong>
          <span>{{ upParts.unit }}</span>
        </div>
        <div class="kpi-sub font-mono">外网总计 · 累计上传 {{ formatBytes(overview.stream.total_up) }}</div>
      </template>

      <template v-else-if="kpiId === 'download'">
        <div class="kpi-head">
          <span>实时下行速率</span>
          <ArrowDownLeft :size="16" class="ic-down" />
        </div>
        <div class="kpi-value font-mono">
          <strong>{{ downParts.value }}</strong>
          <span>{{ downParts.unit }}</span>
        </div>
        <div class="kpi-sub font-mono">外网总计 · 累计下载 {{ formatBytes(overview.stream.total_down) }}</div>
      </template>

      <template v-else-if="kpiId === 'connections'">
        <div class="kpi-head">
          <span>活跃 NAT 会话</span>
          <Activity :size="16" class="ic-accent" />
        </div>
        <div class="kpi-value font-mono">
          <strong>{{ overview.connections.total.toLocaleString() }}</strong>
          <span>Conns</span>
        </div>
        <div class="kpi-sub font-mono">
          TCP {{ overview.connections.tcp }} · UDP {{ overview.connections.udp }} · ICMP {{ overview.connections.icmp }}
        </div>
      </template>

      <template v-else-if="kpiId === 'users'">
        <div class="kpi-head">
          <span>在线终端设备</span>
          <Users :size="16" class="ic-up" />
        </div>
        <div class="kpi-value font-mono">
          <strong>{{ overview.online_users.total }}</strong>
          <span>台在线</span>
        </div>
        <div class="kpi-sub font-mono">
          有线 {{ overview.online_users.wired }} · 2.4G {{ overview.online_users.wifi_2g }} · 5G {{ overview.online_users.wifi_5g }}
        </div>
      </template>
    </article>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Activity, ArrowDownLeft, ArrowLeft, ArrowRight, ArrowUpRight, GripVertical, Users } from 'lucide-vue-next'
import type { OverviewDTO } from '@/api/v4'
import { formatBytes, formatSpeedParts } from '@/composables/useFormatters'
import type { KpiCardId } from '@/composables/useOverviewLayout'

const props = defineProps<{
  overview: OverviewDTO
  kpiOrder: KpiCardId[]
  layoutLocked: boolean
  draggingId: string | null
  overId: string | null
}>()

const emit = defineEmits<{
  (e: 'navigate', payload: { tab: string; subResource?: string }): void
  (e: 'move', payload: { id: KpiCardId; delta: -1 | 1 }): void
  (e: 'drag-start', payload: { id: KpiCardId; event: DragEvent }): void
  (e: 'drag-enter', id: KpiCardId): void
  (e: 'drag-leave', payload: { id: KpiCardId; event: DragEvent }): void
  (e: 'drop', id: KpiCardId): void
  (e: 'drag-end'): void
}>()

const upParts = computed(() => formatSpeedParts(props.overview.stream.upload_bps))
const downParts = computed(() => formatSpeedParts(props.overview.stream.download_bps))

function kpiCardTitle(id: KpiCardId): string {
  if (!props.layoutLocked) return '按住拖动与其他指标卡互换位置'
  if (id === 'upload' || id === 'download') {
    return '直接取自 /api/v4.0/monitoring/system 的 sysinfo.stream（全部 WAN 外网出口总吞吐）'
  }
  return id === 'connections' ? '点击查看流量与协议画像' : '点击查看在线终端列表'
}

function handleKpiClick(id: KpiCardId) {
  if (!props.layoutLocked) return
  if (id === 'connections') emit('navigate', { tab: 'traffic-topology' })
  else if (id === 'users') emit('navigate', { tab: 'terminals', subResource: 'online' })
}
</script>

<style scoped>
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.kpi-card {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.kpi-card.clickable {
  cursor: pointer;
}

.kpi-card.clickable:hover {
  transform: translateY(-2px);
}

.kpi-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
}

.kpi-value {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.kpi-value strong {
  font-size: 26px;
  font-weight: 800;
  color: var(--text-strong);
}

.kpi-value span {
  font-size: 12px;
  color: var(--text-muted);
}

.kpi-sub {
  font-size: 11px;
  color: var(--text-muted);
}

.ic-up { color: var(--chart-up); }
.ic-down { color: var(--chart-down); }
.ic-accent { color: var(--accent-text); }

.card-drag-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  padding-bottom: 6px;
  border-bottom: 1px dashed var(--divider-line);
}

.step-btns {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.drag-handle-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 999px;
  border: 1px dashed color-mix(in srgb, var(--accent-primary) 42%, var(--border-control));
  background: var(--accent-primary-soft);
  color: var(--accent-text);
  font-size: 11px;
  font-weight: 600;
  cursor: grab;
  user-select: none;
}

.drag-handle-pill:active {
  cursor: grabbing;
}

.drag-step-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 7px;
  border: 1px solid var(--border-control);
  background: var(--surface-btn);
  color: var(--text-normal);
  cursor: pointer;
}

.drag-step-btn:hover:not(:disabled) {
  border-color: var(--accent-primary);
  color: var(--accent-text);
}

.drag-step-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.layout-editable {
  border-style: dashed;
  border-color: color-mix(in srgb, var(--accent-primary) 34%, var(--border-subtle));
  cursor: grab;
}

.layout-editable.is-dragging {
  opacity: 0.45;
  transform: scale(0.985);
}

.layout-editable.is-drop-target {
  border-style: solid;
  border-color: var(--accent-primary);
  box-shadow:
    0 0 0 2px var(--accent-primary-soft),
    var(--shadow-elevated);
}

@media (max-width: 1180px) {
  .kpi-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 600px) {
  .kpi-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
  }

  .kpi-card {
    padding: 12px;
    gap: 6px;
  }

  .kpi-value strong {
    font-size: 20px;
  }
}
</style>

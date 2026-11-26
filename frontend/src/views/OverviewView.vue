<template>
  <section class="overview-page">
    <OverviewIdentityBanner
      :overview="overview"
      :routerName="activeRouter?.name"
      :readOnly="readOnly"
      :actionBusy="actionBusy"
      :loading="loading"
      :layoutLocked="layoutLocked"
      @quick-action="runQuickAction($event.key, $event.okText)"
      @toggle-layout="layoutLocked ? unlockLayout() : finishLayoutEdit()"
      @refresh="refreshAll"
    />

    <div v-if="!layoutLocked" class="layout-edit-bar v4-card" role="status">
      <div class="layout-edit-copy">
        <Pin :size="14" />
        <span>布局调整中：按住拖动上方指标卡或下方任意模块即可互换位置（手机端可点 ← / →），完成后点击「固定布局」锁定。</span>
      </div>
      <div class="layout-edit-actions">
        <button type="button" class="v4-btn btn-xs" :disabled="!isCustomLayout" @click="handleResetLayout">
          <RotateCcw :size="13" />
          <span>恢复默认</span>
        </button>
        <button type="button" class="v4-btn v4-btn-primary btn-xs" @click="finishLayoutEdit">
          <Check :size="13" />
          <span>固定布局</span>
        </button>
      </div>
    </div>

    <div
      v-if="actionNotice"
      class="notice-row"
      :class="actionNotice.tone === 'warn' ? 'warn-banner' : actionNotice.tone === 'err' ? 'error-banner' : 'toast-banner'"
    >
      <span>{{ actionNotice.text }}</span>
      <button type="button" class="notice-dismiss" aria-label="关闭提示" @click="actionNotice = null">×</button>
    </div>
    <div v-if="error" class="error-banner">{{ error }}</div>

    <div v-if="loading && !overview" class="loading-state">
      <div class="spinner" />
      <span class="font-mono">正在连接 iKuai V4 遥测总线...</span>
    </div>

    <template v-else-if="overview">
      <OverviewKpiGrid
        :overview="overview"
        :kpiOrder="kpiOrder"
        :layoutLocked="layoutLocked"
        :draggingId="dragState.group === 'kpi' ? dragState.draggingId : null"
        :overId="dragState.group === 'kpi' ? dragState.overId : null"
        @navigate="emit('navigate', $event)"
        @move="moveKpi($event.id, $event.delta)"
        @drag-start="onDragStart('kpi', $event.id, $event.event)"
        @drag-enter="onDragEnter('kpi', $event)"
        @drag-leave="onDragLeave('kpi', $event.id, $event.event)"
        @drop="onDrop('kpi', $event)"
        @drag-end="onDragEnd"
      />

      <div class="two-col-grid sections-grid">
        <article
          v-for="(secId, secIdx) in sectionOrder"
          :key="secId"
          class="v4-card section-card"
          :class="{
            'layout-editable': !layoutLocked,
            'is-dragging': dragState.group === 'section' && dragState.draggingId === secId,
            'is-drop-target':
              dragState.group === 'section' && dragState.overId === secId && dragState.draggingId !== secId,
          }"
          :draggable="!layoutLocked"
          @dragstart="onDragStart('section', secId, $event)"
          @dragover.prevent
          @dragenter.prevent="onDragEnter('section', secId)"
          @dragleave="onDragLeave('section', secId, $event)"
          @drop.prevent="onDrop('section', secId)"
          @dragend="onDragEnd"
        >
          <OverviewComputeCard
            v-if="secId === 'cpu-mem'"
            :overview="overview"
            :layoutLocked="layoutLocked"
            :canMovePrev="secIdx > 0"
            :canMoveNext="secIdx < sectionOrder.length - 1"
            @move="moveSection(secId, $event)"
          />

          <OverviewPortsCard
            v-else-if="secId === 'physical-ports'"
            :ports="overview.physical_ports || []"
            :layoutLocked="layoutLocked"
            :canMovePrev="secIdx > 0"
            :canMoveNext="secIdx < sectionOrder.length - 1"
            @move="moveSection(secId, $event)"
            @navigate="emit('navigate', $event)"
          />

          <OverviewWanChartCard
            v-else-if="secId === 'wan-chart'"
            :overview="overview"
            :routerKey="routerBufferKey"
            :layoutLocked="layoutLocked"
            :canMovePrev="secIdx > 0"
            :canMoveNext="secIdx < sectionOrder.length - 1"
            :orderIndex="secIdx"
            @move="moveSection(secId, $event)"
          />

          <OverviewLoadChartCard
            v-else-if="secId === 'load-chart'"
            :ref="setLoadChartCardRef"
            :routerKey="`${routerBufferKey}:${routerEpoch}`"
            :pollTick="loadPollTick"
            :layoutLocked="layoutLocked"
            :canMovePrev="secIdx > 0"
            :canMoveNext="secIdx < sectionOrder.length - 1"
            :orderIndex="secIdx"
            @move="moveSection(secId, $event)"
          />

          <OverviewWanHealthCard
            v-else-if="secId === 'wan-checks'"
            :wanChecks="overview.wan_checks || []"
            :layoutLocked="layoutLocked"
            :canMovePrev="secIdx > 0"
            :canMoveNext="secIdx < sectionOrder.length - 1"
            @move="moveSection(secId, $event)"
            @navigate="emit('navigate', $event)"
          />

          <OverviewIfaceTrafficCard
            v-else-if="secId === 'iface-traffic'"
            :interfaces="overview.interface_traffic || []"
            :wanChecks="overview.wan_checks || []"
            :layoutLocked="layoutLocked"
            :canMovePrev="secIdx > 0"
            :canMoveNext="secIdx < sectionOrder.length - 1"
            @move="moveSection(secId, $event)"
          />
        </article>
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, onActivated, onDeactivated, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { Check, Pin, RotateCcw } from 'lucide-vue-next'
import { executeV4Action, fetchV4Overview, type OverviewDTO } from '@/api/v4'
import OverviewComputeCard from '@/components/overview/OverviewComputeCard.vue'
import OverviewIdentityBanner from '@/components/overview/OverviewIdentityBanner.vue'
import OverviewIfaceTrafficCard from '@/components/overview/OverviewIfaceTrafficCard.vue'
import OverviewKpiGrid from '@/components/overview/OverviewKpiGrid.vue'
import OverviewLoadChartCard from '@/components/overview/OverviewLoadChartCard.vue'
import OverviewPortsCard from '@/components/overview/OverviewPortsCard.vue'
import OverviewWanChartCard from '@/components/overview/OverviewWanChartCard.vue'
import OverviewWanHealthCard from '@/components/overview/OverviewWanHealthCard.vue'
import { useAuth } from '@/composables/useAuth'
import { useFleet } from '@/composables/useFleet'
import { useOverviewLayout, type KpiCardId, type SectionCardId } from '@/composables/useOverviewLayout'

const emit = defineEmits<{
  (e: 'navigate', payload: { tab: string; subResource?: string }): void
}>()

const { readOnly } = useAuth()
const { activeRouter, routerEpoch } = useFleet()
const {
  kpiOrder,
  sectionOrder,
  layoutLocked,
  isCustomLayout,
  swapKpi,
  swapSection,
  moveKpi,
  moveSection,
  unlockLayout,
  lockLayout,
  resetLayout,
} = useOverviewLayout()

const loading = ref(true)
const actionBusy = ref(false)
const actionNotice = ref<{ text: string; tone: 'ok' | 'warn' | 'err' } | null>(null)
const error = ref<string | null>(null)
const overview = ref<OverviewDTO | null>(null)
const loadPollTick = ref(0)

const loadChartCardRef = ref<{ refresh: () => Promise<void> } | null>(null)

function setLoadChartCardRef(el: unknown) {
  if (el && typeof el === 'object' && 'refresh' in el) {
    loadChartCardRef.value = el as { refresh: () => Promise<void> }
  } else {
    loadChartCardRef.value = null
  }
}

const routerBufferKey = computed(() => activeRouter.value?.id || 'default')

const dragState = reactive<{
  group: 'kpi' | 'section' | null
  draggingId: string | null
  overId: string | null
}>({ group: null, draggingId: null, overId: null })

let pollTimer: ReturnType<typeof setInterval> | null = null
let overviewReqSeq = 0
let activeOverviewReqs = 0

function onDragStart(group: 'kpi' | 'section', id: string, event: DragEvent) {
  if (layoutLocked.value) {
    event.preventDefault()
    return
  }
  const target = event.target as HTMLElement | null
  if (target?.closest('button, select, input, th') && !target.closest('.drag-handle-pill')) {
    event.preventDefault()
    return
  }
  dragState.group = group
  dragState.draggingId = id
  dragState.overId = id
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', `${group}:${id}`)
  }
}

function onDragEnter(group: 'kpi' | 'section', id: string) {
  if (layoutLocked.value || dragState.group !== group || !dragState.draggingId) return
  dragState.overId = id
}

function onDragLeave(group: 'kpi' | 'section', id: string, event: DragEvent) {
  const current = event.currentTarget as HTMLElement | null
  const related = event.relatedTarget as Node | null
  if (current && related && current.contains(related)) return
  if (dragState.group === group && dragState.overId === id && dragState.draggingId !== id) {
    dragState.overId = null
  }
}

function onDrop(group: 'kpi' | 'section', targetId: string) {
  if (layoutLocked.value || dragState.group !== group || !dragState.draggingId) return
  const fromId = dragState.draggingId
  if (fromId !== targetId) {
    if (group === 'kpi') swapKpi(fromId as KpiCardId, targetId as KpiCardId)
    else swapSection(fromId as SectionCardId, targetId as SectionCardId)
  }
  dragState.group = null
  dragState.draggingId = null
  dragState.overId = null
}

function onDragEnd() {
  dragState.group = null
  dragState.draggingId = null
  dragState.overId = null
}

function finishLayoutEdit() {
  onDragEnd()
  lockLayout()
  showToast('首页模块布局已固定保存', 'ok')
}

function handleResetLayout() {
  onDragEnd()
  resetLayout()
  showToast('已恢复默认模块排序', 'ok')
}

function showToast(text: string, tone: 'ok' | 'warn' | 'err', duration = 4200) {
  actionNotice.value = { text, tone }
  setTimeout(() => {
    if (actionNotice.value?.text === text) actionNotice.value = null
  }, duration)
}

async function fetchOverviewOnly(clearLoading = true, background = false) {
  if (background && activeOverviewReqs > 0) return
  const seq = ++overviewReqSeq
  activeOverviewReqs += 1
  try {
    const data = await fetchV4Overview()
    if (seq !== overviewReqSeq) return
    overview.value = data
    error.value = null
  } catch (e: unknown) {
    if (seq !== overviewReqSeq) return
    error.value = extractErr(e)
  } finally {
    activeOverviewReqs = Math.max(0, activeOverviewReqs - 1)
    if (clearLoading && seq === overviewReqSeq) {
      loading.value = false
    }
  }
}

async function refreshAll() {
  loading.value = true
  try {
    await Promise.all([
      fetchOverviewOnly(false),
      loadChartCardRef.value ? loadChartCardRef.value.refresh() : Promise.resolve(),
    ])
  } finally {
    loading.value = false
  }
}

async function runQuickAction(actionKey: string, okText: string) {
  if (readOnly.value) return
  actionBusy.value = true
  try {
    const res = await executeV4Action(actionKey, false)
    const msg = String(res.message || okText)
    const tone: 'ok' | 'warn' = res.warning ? 'warn' : 'ok'
    showToast(msg, tone, 5500)
    await fetchOverviewOnly()
  } catch (e: unknown) {
    showToast(extractErr(e), 'err', 5500)
  } finally {
    actionBusy.value = false
  }
}

function extractErr(e: unknown): string {
  if (e && typeof e === 'object' && 'response' in e) {
    const resp = (e as { response?: { data?: { message?: string } } }).response
    if (resp?.data?.message) return resp.data.message
  }
  return e instanceof Error ? e.message : String(e)
}

watch(routerEpoch, () => {
  void fetchOverviewOnly(false)
})

let hasMounted = false

function startPolling() {
  if (pollTimer) return
  pollTimer = setInterval(() => {
    if (typeof document !== 'undefined' && document.hidden) return
    void fetchOverviewOnly(true, true)
    loadPollTick.value += 1
  }, 4000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

function handleVisibilityChange() {
  if (typeof document !== 'undefined' && !document.hidden) {
    void fetchOverviewOnly(false)
  }
}

onMounted(async () => {
  hasMounted = true
  await fetchOverviewOnly()
  startPolling()
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

onActivated(() => {
  if (hasMounted) {
    void fetchOverviewOnly()
    startPolling()
    document.addEventListener('visibilitychange', handleVisibilityChange)
  }
})

onDeactivated(() => {
  stopPolling()
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})

onUnmounted(() => {
  stopPolling()
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})
</script>

<style scoped>
.overview-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.two-col-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.section-card {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-width: 0;
}

.layout-edit-bar {
  padding: 10px 14px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  border-color: color-mix(in srgb, var(--accent-primary) 38%, transparent);
  background: var(--accent-primary-soft);
}

.layout-edit-copy {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12.5px;
  font-weight: 600;
  color: var(--text-strong);
}

.layout-edit-copy svg {
  color: var(--accent-text);
  flex-shrink: 0;
}

.layout-edit-actions {
  display: flex;
  align-items: center;
  gap: 8px;
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

.notice-row {
  justify-content: space-between;
}

.notice-dismiss {
  flex-shrink: 0;
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 999px;
  background: transparent;
  color: inherit;
  font-size: 16px;
  line-height: 1;
  cursor: pointer;
  opacity: 0.78;
}

.notice-dismiss:hover {
  opacity: 1;
  background: rgba(128, 128, 128, 0.14);
}

@media (max-width: 1180px) {
  .two-col-grid {
    grid-template-columns: 1fr;
  }
}
</style>

<template>
  <div class="fleet-switcher-wrap" ref="rootEl">
    <button
      type="button"
      class="fleet-trigger"
      :aria-expanded="open"
      @click="toggleOpen"
    >
      <span
        class="status-led"
        :class="statusClass"
      />
      <div class="trigger-meta">
        <span class="trigger-name">{{ activeRouter?.name || '未配置 V4 节点' }}</span>
        <span class="trigger-sub font-mono">
          {{ activeProbe?.version ? `v${activeProbe.version}` : 'V4 REST' }}
          <template v-if="activeProbe?.online && activeProbe.latency_ms >= 0">
            · {{ activeProbe.latency_ms }}ms
          </template>
          <template v-else-if="status.mode === 'mock'"> · 模拟模式</template>
        </span>
      </div>
      <ChevronDown :size="15" class="chevron" :class="{ rotated: open }" />
    </button>

    <transition name="pop">
      <div v-if="open" class="fleet-popover v4-panel" role="dialog" aria-label="多爱快 V4 节点切换">
        <div class="popover-head">
          <div>
            <strong>爱快 V4 节点矩阵</strong>
            <span>共 {{ config.routers.length }} 台路由器 · 实时状态</span>
          </div>
          <div class="popover-actions">
            <button type="button" class="v4-btn icon-only" :disabled="probing" title="探测全部节点" @click="refreshProbe">
              <RefreshCw :size="14" :class="{ spin: probing }" />
            </button>
            <button type="button" class="v4-btn v4-btn-primary btn-sm" @click="goManage">
              <SlidersHorizontal :size="13" />
              <span>管理</span>
            </button>
          </div>
        </div>

        <div v-if="config.routers.length === 0" class="empty-fleet">
          <p>暂无 V4 节点，点击管理添加爱快路由器。</p>
        </div>

        <div v-else class="fleet-list">
          <button
            v-for="item in mergedList"
            :key="item.id"
            type="button"
            class="fleet-item"
            :class="{ active: item.id === config.active_router_id }"
            :disabled="switchingId !== null"
            @click="selectRouter(item.id)"
          >
            <div class="item-top">
              <div class="item-title-row">
                <span class="status-led" :class="item.online ? 'up' : item.checked ? 'down' : 'warn'" />
                <strong class="item-name">{{ item.name }}</strong>
                <span v-if="item.mock" class="v4-badge v4-badge-accent">MOCK</span>
                <span v-else-if="item.source === 'env'" class="v4-badge">ENV</span>
              </div>
              <span v-if="item.id === config.active_router_id" class="v4-badge v4-badge-up">当前</span>
              <span v-else-if="switchingId === item.id" class="v4-badge">切换中...</span>
              <span v-else-if="item.online" class="latency font-mono">{{ item.latency_ms }}ms</span>
            </div>

            <div class="item-mid font-mono">
              <span>{{ item.url }}</span>
              <span v-if="item.verstring">{{ item.version }}</span>
            </div>

            <div v-if="item.online" class="item-telemetry font-mono">
              <span>CPU {{ item.cpu_percent.toFixed(1) }}%</span>
              <span>MEM {{ item.memory_percent.toFixed(0) }}%</span>
              <span>↓ {{ formatSpeed(item.download_bps) }}</span>
              <span>↑ {{ formatSpeed(item.upload_bps) }}</span>
            </div>
            <div v-else-if="item.error" class="item-error">
              {{ item.error }}
            </div>
          </button>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { ChevronDown, RefreshCw, SlidersHorizontal } from 'lucide-vue-next'
import { useFleet } from '@/composables/useFleet'
import { formatSpeed } from '@/composables/useFormatters'

const emit = defineEmits<{
  (e: 'manage'): void
}>()

const {
  config,
  status,
  probeItems,
  probing,
  switchingId,
  activeRouter,
  activeProbe,
  refreshProbe,
  activateRouter,
} = useFleet()

const open = ref(false)
const rootEl = ref<HTMLElement | null>(null)

const statusClass = computed(() => {
  if (status.value.mode === 'unconfigured') return 'warn'
  if (activeProbe.value) return activeProbe.value.online ? 'up' : 'down'
  return 'up'
})

const mergedList = computed(() => {
  return config.value.routers.map((r) => {
    const p = probeItems.value.find((item) => item.id === r.id)
    return {
      id: r.id,
      name: r.name,
      url: r.url,
      mock: r.mock,
      source: r.source,
      checked: Boolean(p),
      online: p?.online ?? false,
      latency_ms: p?.latency_ms ?? 0,
      version: p?.version ?? '',
      verstring: p?.verstring ?? '',
      cpu_percent: p?.cpu_percent ?? 0,
      memory_percent: p?.memory_percent ?? 0,
      upload_bps: p?.upload_bps ?? 0,
      download_bps: p?.download_bps ?? 0,
      error: p?.error ?? '',
    }
  })
})

function toggleOpen() {
  open.value = !open.value
  if (open.value) {
    void refreshProbe()
  }
}

async function selectRouter(id: string) {
  await activateRouter(id)
  open.value = false
}

function goManage() {
  open.value = false
  emit('manage')
}

function handleDocClick(event: MouseEvent) {
  if (open.value && rootEl.value && !rootEl.value.contains(event.target as Node)) {
    open.value = false
  }
}

onMounted(() => document.addEventListener('click', handleDocClick))
onUnmounted(() => document.removeEventListener('click', handleDocClick))
</script>

<style scoped>
.fleet-switcher-wrap {
  position: relative;
  min-width: 0;
  max-width: 100%;
}

.fleet-trigger {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-height: 38px;
  padding: 4px 12px;
  border-radius: var(--radius-control);
  border: 1px solid var(--border-control);
  background: var(--surface-btn);
  color: var(--text-strong);
  box-shadow: var(--inner-sheen);
  cursor: pointer;
  transition: all 0.18s ease;
  max-width: min(280px, 100%);
  min-width: 0;
}

.fleet-trigger:hover {
  border-color: var(--border-strong);
  background: var(--surface-hover);
}

.trigger-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  min-width: 0;
  flex: 1;
  text-align: left;
}

.trigger-name {
  font-size: 13px;
  font-weight: 700;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}

.trigger-sub {
  font-size: 11px;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}

.status-led {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.status-led.up {
  background: var(--status-up);
  box-shadow: 0 0 10px var(--status-up);
}
.status-led.warn {
  background: var(--status-warn);
  box-shadow: 0 0 10px var(--status-warn);
}
.status-led.down {
  background: var(--status-down);
  box-shadow: 0 0 10px var(--status-down);
}

.chevron {
  color: var(--text-muted);
  transition: transform 0.2s ease;
  flex-shrink: 0;
}
.chevron.rotated {
  transform: rotate(180deg);
}

.fleet-popover {
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  width: min(400px, calc(100vw - 24px));
  padding: 14px;
  z-index: 220;
  background: var(--surface-popover);
  border: 1px solid var(--border-control);
  -webkit-backdrop-filter: var(--backdrop-glass);
  backdrop-filter: var(--backdrop-glass);
}

.popover-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-subtle);
}

.popover-head > div:first-child > strong {
  display: block;
  font-size: 14px;
  color: var(--text-strong);
}

.popover-head > div:first-child > span {
  font-size: 11px;
  color: var(--text-muted);
}

.popover-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.icon-only {
  width: 32px;
  min-height: 30px;
  padding: 0;
}

.btn-sm {
  min-height: 30px;
  padding: 0 10px;
  font-size: 12px;
}

.fleet-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 12px;
  max-height: 360px;
  overflow-y: auto;
}

.fleet-item {
  width: 100%;
  text-align: left;
  padding: 10px 12px;
  border-radius: 12px;
  border: 1px solid var(--border-subtle);
  background: var(--surface-input);
  color: var(--text-strong);
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 6px;
  transition: all 0.16s ease;
}

.fleet-item:hover {
  border-color: var(--border-strong);
  background: var(--surface-hover);
}

.fleet-item.active {
  border-color: var(--accent-primary);
  background: var(--accent-primary-soft);
}

.item-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.item-title-row {
  display: flex;
  align-items: center;
  gap: 7px;
  min-width: 0;
}

.item-name {
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.latency {
  font-size: 11px;
  color: var(--status-up);
}

.item-mid {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: var(--text-muted);
}

.item-telemetry {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  font-size: 11px;
  color: var(--text-normal);
}

.item-error {
  font-size: 11px;
  color: var(--status-down);
}

.empty-fleet {
  padding: 18px 8px;
  text-align: center;
  color: var(--text-muted);
  font-size: 13px;
}

.spin {
  animation: spin 0.9s linear infinite;
}

.pop-enter-active,
.pop-leave-active {
  transition: opacity 0.16s ease, transform 0.16s ease;
}
.pop-enter-from,
.pop-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

@media (max-width: 640px) {
  .fleet-switcher-wrap {
    flex: 1;
    display: flex;
  }

  .fleet-trigger {
    width: 100%;
    max-width: 100%;
    min-height: 34px;
    padding: 3px 9px;
    gap: 7px;
  }

  .trigger-name {
    font-size: 12.5px;
  }

  .trigger-sub {
    font-size: 10px;
  }

  .fleet-popover {
    position: fixed;
    top: 58px;
    left: 10px;
    right: 10px;
    width: auto;
    max-height: calc(100dvh - 140px);
    display: flex;
    flex-direction: column;
  }

  .fleet-list {
    flex: 1;
    max-height: none;
  }
}
</style>

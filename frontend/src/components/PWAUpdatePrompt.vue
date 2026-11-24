<template>
  <Transition name="pwa-toast">
    <div v-if="visible" class="pwa-update v4-panel" role="status" aria-live="polite">
      <div class="pwa-copy">
        <span class="pwa-kicker">{{ needRefresh ? '发现新版本' : '离线模式已就绪' }}</span>
        <strong>{{ needRefresh ? '刷新后即可使用最新面板' : '已可添加到桌面使用' }}</strong>
      </div>
      <div class="pwa-actions">
        <button v-if="needRefresh" type="button" class="v4-btn v4-btn-primary btn-xs" @click="refreshApp">
          立即更新
        </button>
        <button type="button" class="v4-btn btn-xs" @click="dismiss">
          稍后
        </button>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'
import { registerSW } from 'virtual:pwa-register'

const needRefresh = ref(false)
const offlineReady = ref(false)
let autoHideTimer: ReturnType<typeof setTimeout> | null = null

const updateSW = registerSW({
  onNeedRefresh() {
    needRefresh.value = true
  },
  onOfflineReady() {
    offlineReady.value = true
    if (autoHideTimer) clearTimeout(autoHideTimer)
    autoHideTimer = setTimeout(() => {
      offlineReady.value = false
    }, 3500)
  },
})

const visible = computed(() => needRefresh.value || offlineReady.value)

function refreshApp() {
  updateSW(true)
}

function dismiss() {
  needRefresh.value = false
  offlineReady.value = false
}

onUnmounted(() => {
  if (autoHideTimer) clearTimeout(autoHideTimer)
})
</script>

<style scoped>
.pwa-update {
  position: fixed;
  right: max(18px, env(safe-area-inset-right));
  bottom: max(18px, env(safe-area-inset-bottom));
  z-index: 80;
  display: flex;
  align-items: center;
  gap: 16px;
  max-width: min(440px, calc(100vw - 32px));
  padding: 14px 16px;
}

.pwa-copy {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 0;
}

.pwa-kicker {
  color: var(--accent-text);
  font-size: 11.5px;
  font-weight: 800;
}

.pwa-copy strong {
  color: var(--text-strong);
  font-size: 13.5px;
  line-height: 1.35;
}

.pwa-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.pwa-toast-enter-active,
.pwa-toast-leave-active {
  transition: opacity 0.22s ease, transform 0.22s ease;
}

.pwa-toast-enter-from,
.pwa-toast-leave-to {
  opacity: 0;
  transform: translateY(12px);
}

@media (max-width: 1120px) {
  .pwa-update {
    bottom: calc(82px + max(10px, env(safe-area-inset-bottom)));
  }
}

@media (max-width: 560px) {
  .pwa-update {
    left: max(12px, env(safe-area-inset-left));
    right: max(12px, env(safe-area-inset-right));
    flex-direction: column;
    align-items: stretch;
    border-radius: var(--radius-lg);
  }

  .pwa-actions {
    justify-content: flex-end;
  }
}
</style>

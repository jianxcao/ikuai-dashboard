<template>
  <Transition name="pwa-toast">
    <div v-if="visible" class="pwa-update glass-panel" role="status" aria-live="polite">
      <div class="pwa-copy">
        <span class="pwa-kicker">{{ needRefresh ? '发现新版本' : '离线模式已就绪' }}</span>
        <strong>{{ needRefresh ? '刷新后即可使用最新面板' : '已可添加到桌面使用' }}</strong>
      </div>
      <div class="pwa-actions">
        <button v-if="needRefresh" type="button" class="liquid-button pwa-primary" @click="refreshApp">
          立即更新
        </button>
        <button type="button" class="liquid-button" @click="dismiss">
          稍后
        </button>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { registerSW } from 'virtual:pwa-register'

const needRefresh = ref(false)
const offlineReady = ref(false)

const updateSW = registerSW({
  onNeedRefresh() {
    needRefresh.value = true
  },
  onOfflineReady() {
    offlineReady.value = true
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
</script>

<style scoped>
.pwa-update {
  position: fixed;
  right: max(18px, env(safe-area-inset-right));
  bottom: max(18px, env(safe-area-inset-bottom));
  z-index: 80;
  display: flex;
  align-items: center;
  gap: 18px;
  max-width: min(460px, calc(100vw - 32px));
  padding: 16px;
}

.pwa-copy {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.pwa-kicker {
  color: var(--system-blue);
  font-size: 12px;
  font-weight: 800;
}

.pwa-copy strong {
  color: var(--text-primary);
  font-size: 14px;
  line-height: 1.35;
}

.pwa-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.pwa-actions .liquid-button {
  min-height: 34px;
  padding: 0 14px;
  font-size: 13px;
  font-weight: 700;
}

.pwa-primary {
  color: var(--text-primary);
  border-color: var(--system-blue);
  background: var(--control-active);
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

@media (max-width: 920px) {
  .pwa-update {
    bottom: calc(82px + max(10px, env(safe-area-inset-bottom)));
  }
}

@media (max-width: 640px) {
  .pwa-update {
    bottom: calc(136px + max(10px, env(safe-area-inset-bottom)));
  }
}

@media (max-width: 380px) {
  .pwa-update {
    bottom: calc(190px + max(10px, env(safe-area-inset-bottom)));
  }
}
</style>

<template>
  <Transition name="pwa-toast">
    <aside
      v-if="promptVisible"
      class="pwa-update-toast v4-panel v4-chrome"
      role="status"
      aria-live="polite"
    >
      <div class="pwa-toast-head">
        <div class="pwa-toast-title">
          <span class="pwa-icon-badge" aria-hidden="true">
            <Sparkles :size="15" />
          </span>
          <div>
            <strong>发现新版本控制台界面</strong>
            <span class="pwa-ver-pill font-mono">
              {{ remoteVersionLabel ? `${displayVersion} → ${remoteVersionLabel}` : displayVersion }}
            </span>
          </div>
        </div>
        <button
          type="button"
          class="pwa-close-btn"
          title="稍后提醒"
          :disabled="updating"
          @click="dismissPrompt"
        >
          <X :size="14" />
        </button>
      </div>

      <p class="pwa-toast-desc">
        新版本前端资源已准备就绪，刷新后即可应用更新。
      </p>

      <div class="pwa-toast-actions">
        <button
          type="button"
          class="v4-btn btn-xs"
          :disabled="updating"
          @click="dismissPrompt"
        >
          稍后提醒
        </button>
        <button
          type="button"
          class="v4-btn v4-btn-primary btn-xs"
          :disabled="updating"
          @click="applyUpdate"
        >
          <RefreshCw :size="13" :class="{ spin: updating }" />
          <span>{{ updating ? '正在刷新...' : '立即更新' }}</span>
        </button>
      </div>
    </aside>
  </Transition>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { RefreshCw, Sparkles, X } from 'lucide-vue-next'
import { useAppUpdate } from '@/composables/useAppUpdate'

const CHECK_INTERVAL_MS = 10 * 60 * 1000

const {
  displayVersion,
  remoteVersionLabel,
  promptVisible,
  updating,
  initPwaRegistration,
  checkForUpdates,
  applyUpdate,
  dismissPrompt,
} = useAppUpdate()

initPwaRegistration()

let pollTimer: number | null = null

function handleVisibilityChange() {
  if (document.visibilityState === 'visible') {
    void checkForUpdates(false)
  }
}

onMounted(() => {
  void checkForUpdates(false)
  pollTimer = window.setInterval(() => {
    void checkForUpdates(false)
  }, CHECK_INTERVAL_MS)
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

onUnmounted(() => {
  if (pollTimer !== null) {
    window.clearInterval(pollTimer)
    pollTimer = null
  }
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})
</script>

<style scoped>
.pwa-update-toast {
  position: fixed;
  right: 20px;
  bottom: 20px;
  z-index: 320;
  width: min(388px, calc(100vw - 28px));
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  background: var(--surface-popover);
  border-color: color-mix(in srgb, var(--accent-primary) 34%, var(--chrome-border));
  box-shadow: var(--chrome-shadow);
}

.pwa-toast-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.pwa-toast-title {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.pwa-toast-title strong {
  display: block;
  font-size: 13.5px;
  font-weight: 750;
  color: var(--text-strong);
}

.pwa-icon-badge {
  width: 30px;
  height: 30px;
  border-radius: 9px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: var(--accent-primary-soft);
  color: var(--accent-text);
  border: 1px solid color-mix(in srgb, var(--accent-primary) 26%, transparent);
}

.pwa-ver-pill {
  display: inline-flex;
  align-items: center;
  margin-top: 3px;
  padding: 1px 7px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 600;
  background: var(--surface-input);
  border: 1px solid var(--divider-line);
  color: var(--text-muted);
}

.pwa-close-btn {
  border: none;
  background: transparent;
  color: var(--text-muted);
  padding: 4px;
  border-radius: 6px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: color 0.14s var(--ease-ui), background-color 0.14s var(--ease-ui);
}

.pwa-close-btn:hover {
  color: var(--text-strong);
  background: var(--surface-hover);
}

.pwa-toast-desc {
  margin: 0;
  font-size: 12.5px;
  line-height: 1.5;
  color: var(--text-normal);
}

.pwa-toast-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.spin {
  animation: spin 0.9s linear infinite;
}

.pwa-toast-enter-active,
.pwa-toast-leave-active {
  transition: opacity 0.2s var(--ease-ui), transform 0.22s var(--ease-ui);
}

.pwa-toast-enter-from,
.pwa-toast-leave-to {
  opacity: 0;
  transform: translateY(10px) scale(0.98);
}

@media (max-width: 1120px) {
  .pwa-update-toast {
    left: 12px;
    right: 12px;
    width: auto;
    bottom: calc(76px + env(safe-area-inset-bottom));
  }
}
</style>

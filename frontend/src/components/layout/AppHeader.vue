<template>
  <header class="app-shell-nav">
    <nav class="side-nav glass-panel" aria-label="主导航">
      <div class="brand">
        <img class="brand-mark" src="/ikuai-icon.svg" alt="" />
        <div class="brand-copy">
          <span class="brand-text">iKuai Dashboard</span>
          <span class="brand-subtitle">Fluid Network Console</span>
          <span class="brand-version"> {{ buildVersion }}</span>
        </div>
      </div>

      <div v-if="!configurationOnly" class="router-switcher" aria-label="切换爱快服务器">
        <div class="router-switcher-head">
          <Server :size="14" />
          <span>当前服务器</span>
        </div>
        <select
          class="router-select"
          :value="config.active_router_id"
          :disabled="switching !== null || config.routers.length <= 1"
          @change="handleRouterChange"
        >
          <option v-for="router in config.routers" :key="router.id" :value="router.id">
            {{ router.name }}
          </option>
        </select>
      </div>

      <div class="nav-group">
        <p class="nav-group-title">{{ configurationOnly ? '设置' : '核心视图' }}</p>
        <div class="nav-tabs" role="tablist" aria-label="主导航">
          <button
            v-for="tab in primaryTabs"
            :key="tab.id"
            type="button"
            :class="['nav-tab', { active: currentTab === tab.id }]"
            :aria-selected="currentTab === tab.id"
            role="tab"
            @click="handleTabChange(tab.id)"
          >
            <component :is="tab.icon" :size="15" />
            <span>{{ tab.label }}</span>
          </button>
        </div>
      </div>

      <div v-if="secondaryTabs.length" class="nav-group secondary-group">
        <p class="nav-group-title">扩展模块</p>
        <div class="nav-tabs" role="tablist" aria-label="扩展模块">
          <button
            v-for="tab in secondaryTabs"
            :key="tab.id"
            type="button"
            :class="['nav-tab', { active: currentTab === tab.id }]"
            :aria-selected="currentTab === tab.id"
            role="tab"
            @click="handleTabChange(tab.id)"
          >
            <component :is="tab.icon" :size="15" />
            <span>{{ tab.label }}</span>
          </button>
        </div>
      </div>

      <div class="header-actions">
        <div v-if="!configurationOnly" class="status-area" :aria-label="statusLabel">
          <span class="status-dot"></span>
          <span class="status-text">{{ statusLabel }}</span>
        </div>

        <div class="theme-switcher" aria-label="主题切换">
          <button
            v-for="item in themes"
            :key="item.id"
            type="button"
            class="theme-chip"
            :aria-pressed="theme === item.id"
            :title="item.label"
            @click="emit('theme-change', item.id)"
          >
            <component :is="item.id === 'liquid-dark' ? Moon : SunMedium" :size="14" />
            <span>{{ item.shortLabel }}</span>
          </button>
        </div>

        <button
          v-if="authEnabled"
          type="button"
          class="logout-btn"
          title="登出"
          @click="handleLogout"
        >
          <LogOut :size="14" />
          <span>登出</span>
        </button>
      </div>
    </nav>

    <div class="mobile-topbar glass-panel">
      <div class="mobile-brand">
        <img class="mobile-brand-mark" src="/ikuai-icon.svg" alt="" />
        <div class="mobile-brand-copy">
          <span class="mobile-title">{{ activeTab?.mobileLabel || '首页' }}</span>
          <span class="mobile-subtitle">
            {{ configurationOnly ? '配置中心' : activeRouter?.name || statusLabel }}
          </span>
        </div>
      </div>

      <div class="mobile-topbar-actions">
        <!-- <button
          type="button"
          class="icon-pill"
          :aria-expanded="moreOpen"
          aria-controls="mobile-more-sheet"
          @click="moreOpen = !moreOpen"
        >
          <PanelsTopLeft :size="16" />
        </button> -->

        <button
          v-if="authEnabled"
          type="button"
          class="icon-pill"
          aria-label="登出"
          @click="handleLogout"
        >
          <LogOut :size="16" />
        </button>
      </div>
    </div>

    <transition name="sheet-fade">
      <div
        v-if="moreOpen && !configurationOnly"
        class="mobile-sheet-backdrop"
        @click="moreOpen = false"
      />
    </transition>

    <transition name="sheet-rise">
      <section
        v-if="moreOpen"
        id="mobile-more-sheet"
        class="mobile-more-sheet glass-panel"
        aria-label="更多导航"
      >
        <div class="mobile-sheet-head">
          <div>
            <p class="sheet-title">更多模块</p>
            <span class="sheet-subtitle">服务器切换、主题和扩展视图</span>
          </div>
          <button type="button" class="icon-pill" aria-label="关闭" @click="moreOpen = false">
            <X :size="16" />
          </button>
        </div>

        <div v-if="!configurationOnly" class="mobile-sheet-block">
          <span class="mobile-sheet-label">切换服务器</span>
          <select
            class="mobile-router-select"
            :value="config.active_router_id"
            :disabled="switching !== null || config.routers.length <= 1"
            aria-label="切换爱快服务器"
            @change="handleRouterChange"
          >
            <option v-for="router in config.routers" :key="router.id" :value="router.id">
              {{ router.name }}
            </option>
          </select>
        </div>

        <div class="mobile-sheet-block">
          <span class="mobile-sheet-label">扩展视图</span>
          <div class="mobile-secondary-grid">
            <button
              v-for="tab in mobileSheetTabs"
              :key="tab.id"
              type="button"
              :class="['mobile-secondary-tab', { active: currentTab === tab.id }]"
              @click="handleTabChange(tab.id)"
            >
              <component :is="tab.icon" :size="16" />
              <span>{{ tab.label }}</span>
            </button>
          </div>
        </div>

        <div class="mobile-sheet-block">
          <span class="mobile-sheet-label">主题切换</span>
          <div class="theme-switcher mobile-theme-switcher" aria-label="主题切换">
            <button
              v-for="item in themes"
              :key="item.id"
              type="button"
              class="theme-chip"
              :aria-pressed="theme === item.id"
              :title="item.label"
              @click="emit('theme-change', item.id)"
            >
              <component :is="item.id === 'liquid-dark' ? Moon : SunMedium" :size="14" />
              <span>{{ item.shortLabel }}</span>
            </button>
          </div>
        </div>
      </section>
    </transition>

    <nav v-if="!configurationOnly" class="bottom-nav glass-panel" aria-label="移动主导航">
      <button
        v-for="tab in primaryTabs"
        :key="tab.id"
        type="button"
        :class="['bottom-tab', { active: currentTab === tab.id }]"
        :aria-current="currentTab === tab.id ? 'page' : undefined"
        @click="handleTabChange(tab.id)"
      >
        <component :is="tab.icon" :size="18" />
        <span>{{ tab.mobileLabel }}</span>
      </button>
      <button
        type="button"
        :class="['bottom-tab', 'more-tab', { active: moreOpen || isSecondaryActive }]"
        :aria-expanded="moreOpen"
        aria-controls="mobile-more-sheet"
        @click="moreOpen = !moreOpen"
      >
        <Ellipsis :size="18" />
        <span>更多</span>
      </button>
    </nav>
  </header>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Ellipsis, LogOut, Moon, Server, SunMedium, X } from 'lucide-vue-next'
import { useRouterConfig } from '@/composables/useRouterConfig'
import { useAuth } from '@/composables/useAuth'
import type { AppTheme } from '@/composables/useTheme'

type NavTab = {
  id: string
  label: string
  mobileLabel: string
  icon: unknown
  primary?: boolean
}

const props = withDefaults(
  defineProps<{
    currentTab?: string
    connected?: boolean
    configurationOnly?: boolean
    theme: AppTheme
    themes: Array<{ id: AppTheme; label: string; shortLabel: string }>
    tabs: NavTab[]
  }>(),
  {
    currentTab: 'interface',
    connected: false,
    configurationOnly: false
  }
)

const emit = defineEmits<{
  (event: 'tab-change', value: string): void
  (event: 'theme-change', value: AppTheme): void
}>()

const { config, activeRouter, switching, status, activate } = useRouterConfig()
const { authEnabled, logout } = useAuth()
const buildVersion = import.meta.env.VITE_APP_VERSION || 'dev'
const moreOpen = ref(false)

const primaryTabs = computed(() => props.tabs.filter((tab) => tab.primary))
const secondaryTabs = computed(() => props.tabs.filter((tab) => !tab.primary))
const mobileSheetTabs = computed(() => {
  if (props.configurationOnly) {
    return props.tabs.filter((tab) => tab.id === 'settings')
  }
  return secondaryTabs.value
})

const activeTab = computed(() => props.tabs.find((tab) => tab.id === props.currentTab))
const isSecondaryActive = computed(() =>
  secondaryTabs.value.some((tab) => tab.id === props.currentTab)
)

watch(
  () => props.currentTab,
  () => {
    moreOpen.value = false
  }
)

async function handleLogout() {
  await logout()
}

const statusLabel = computed(() => {
  if (status.value?.mode === 'unconfigured') return '未配置'
  return status.value?.mode === 'mock' ? '模拟数据' : '实时连接'
})

function handleRouterChange(event: Event) {
  const id = (event.target as HTMLSelectElement).value
  if (id && id !== config.value.active_router_id) {
    void activate(id)
  }
}

function handleTabChange(tabId: string) {
  emit('tab-change', tabId)
  moreOpen.value = false
}
</script>

<style scoped>
.app-shell-nav {
  z-index: 50;
}

.side-nav {
  position: fixed;
  inset: 16px auto 16px 16px;
  display: flex;
  flex-direction: column;
  gap: 18px;
  width: 264px;
  padding: 18px;
}

.brand,
.mobile-brand {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.brand-copy,
.mobile-brand-copy {
  min-width: 0;
}

.brand-mark {
  width: 44px;
  height: 44px;
  flex: 0 0 auto;
}

.brand-text,
.mobile-title {
  display: block;
  color: var(--text-primary);
  font-size: 15px;
  font-weight: 760;
  white-space: nowrap;
}

.brand-subtitle,
.mobile-subtitle {
  display: block;
  margin-top: 3px;
  color: var(--text-tertiary);
  font-size: 12px;
  font-weight: 640;
}

.brand-version {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  max-width: 100%;
  min-height: 22px;
  margin-top: 8px;
  padding: 0 8px;
  overflow: hidden;
  color: var(--text-secondary);
  border: 1px solid var(--control-border);
  border-radius: 999px;
  background: var(--control-bg);
  font-size: 11px;
  font-weight: 760;
  line-height: 1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.router-switcher {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  border: 1px solid var(--control-border);
  border-radius: 16px;
  background: var(--control-bg);
}

.router-switcher-head {
  display: flex;
  align-items: center;
  gap: 7px;
  color: var(--text-tertiary);
  font-size: 12px;
  font-weight: 760;
}

.router-select,
.mobile-router-select {
  width: 100%;
  min-height: 40px;
  padding: 0 40px 0 12px;
  color: var(--text-primary);
  border: 1px solid var(--control-border);
  border-radius: 12px;
  background: var(--glass-bg-strong);
  font-size: 13px;
  font-weight: 720;
  appearance: none;
  background-image:
    linear-gradient(45deg, transparent 50%, currentColor 50%),
    linear-gradient(135deg, currentColor 50%, transparent 50%), var(--glass-bg-strong);
  background-position:
    calc(100% - 23px) 50%,
    calc(100% - 16px) 50%,
    0 0;
  background-size:
    7px 7px,
    7px 7px,
    auto;
  background-repeat: no-repeat;
}

.nav-group {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.nav-group-title,
.mobile-sheet-label,
.sheet-title {
  color: var(--text-tertiary);
  font-size: 11px;
  font-weight: 760;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.nav-tabs {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.nav-tab {
  display: inline-flex;
  align-items: center;
  justify-content: flex-start;
  gap: 8px;
  min-height: 44px;
  width: 100%;
  padding: 0 12px;
  border: 1px solid transparent;
  border-radius: 14px;
  background: transparent;
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 650;
  cursor: pointer;
  white-space: nowrap;
  transition:
    color 0.2s ease,
    background 0.2s ease,
    border-color 0.2s ease,
    transform 0.2s ease;
}

.nav-tab:hover {
  color: var(--text-primary);
  background: var(--control-bg);
}

.nav-tab.active {
  color: var(--text-primary);
  background: var(--control-active);
  border-color: var(--system-blue);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.24),
    0 0 24px var(--system-blue-dim);
}

.secondary-group {
  padding-top: 4px;
  border-top: 1px solid var(--control-border);
}

.header-actions {
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  gap: 10px;
  margin-top: auto;
}

.status-area {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 40px;
  padding: 0 12px;
  border: 1px solid var(--control-border);
  border-radius: 999px;
  background: var(--control-bg);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: var(--system-green);
  box-shadow: 0 0 0 4px var(--system-green-dim);
}

.status-text {
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 700;
}

.theme-switcher {
  display: inline-flex;
  gap: 4px;
  padding: 4px;
  border: 1px solid var(--control-border);
  border-radius: 999px;
  background: rgba(0, 0, 0, 0.06);
}

.theme-chip {
  min-height: 30px;
  padding: 0 10px;
  font-size: 12px;
  font-weight: 700;
}

.logout-btn,
.icon-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  min-height: 34px;
  padding: 0 12px;
  border: 1px solid var(--control-border);
  border-radius: 999px;
  background: var(--control-bg);
  color: var(--text-tertiary);
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  transition:
    color 0.18s ease,
    background 0.18s ease,
    border-color 0.18s ease;
}

.logout-btn:hover,
.icon-pill:hover {
  color: var(--text-primary);
  border-color: var(--glass-border-light);
  background: var(--control-bg-hover);
}

.mobile-topbar,
.bottom-nav,
.mobile-more-sheet {
  display: none;
}

@media (max-width: 1100px) {
  .side-nav {
    display: none;
  }

  .mobile-topbar {
    position: fixed;
    top: max(10px, env(safe-area-inset-top));
    right: max(10px, env(safe-area-inset-right));
    left: max(10px, env(safe-area-inset-left));
    z-index: 80;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    padding: 10px 12px;
    border-radius: 22px;
  }

  .mobile-brand-mark {
    width: 34px;
    height: 34px;
    flex: 0 0 auto;
  }

  .mobile-title {
    font-size: 14px;
  }

  .mobile-subtitle {
    max-width: 42vw;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .mobile-topbar-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .icon-pill {
    width: 38px;
    min-height: 38px;
    padding: 0;
  }

  .bottom-nav {
    position: fixed;
    right: max(10px, env(safe-area-inset-right));
    bottom: max(10px, env(safe-area-inset-bottom));
    left: max(10px, env(safe-area-inset-left));
    z-index: 80;
    display: grid;
    grid-template-columns: repeat(5, minmax(0, 1fr));
    gap: 6px;
    padding: 8px;
    border-radius: 24px;
  }

  .bottom-tab {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
    min-height: 56px;
    padding: 0 4px;
    color: var(--text-tertiary);
    border: 0;
    border-radius: 16px;
    background: transparent;
    cursor: pointer;
    transition:
      color 0.18s ease,
      background 0.18s ease;
  }

  .bottom-tab span {
    max-width: 100%;
    overflow: hidden;
    font-size: 11px;
    font-weight: 760;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .bottom-tab.active {
    color: var(--text-primary);
    background: var(--control-active);
    box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.18);
  }

  .mobile-sheet-backdrop {
    position: fixed;
    inset: 0;
    z-index: 84;
    background: rgba(3, 10, 14, 0.34);
    backdrop-filter: blur(6px);
  }

  .mobile-more-sheet {
    position: fixed;
    right: max(10px, env(safe-area-inset-right));
    bottom: calc(88px + max(10px, env(safe-area-inset-bottom)));
    left: max(10px, env(safe-area-inset-left));
    z-index: 85;
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 16px;
    border-radius: 24px;
  }

  .mobile-sheet-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
  }

  .sheet-subtitle {
    display: block;
    margin-top: 6px;
    color: var(--text-secondary);
    font-size: 12px;
  }

  .mobile-sheet-block {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .mobile-secondary-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
  }

  .mobile-secondary-tab {
    display: inline-flex;
    align-items: center;
    justify-content: flex-start;
    gap: 8px;
    min-height: 46px;
    padding: 0 14px;
    border: 1px solid var(--control-border);
    border-radius: 16px;
    background: var(--control-bg);
    color: var(--text-secondary);
    font-size: 13px;
    font-weight: 700;
    cursor: pointer;
  }

  .mobile-secondary-tab.active {
    color: var(--text-primary);
    border-color: var(--system-blue);
    background: var(--control-active);
  }

  .mobile-theme-switcher {
    width: fit-content;
    max-width: 100%;
    flex-wrap: wrap;
  }
}

@media (max-width: 640px) {
  .mobile-subtitle {
    max-width: 34vw;
  }

  .bottom-nav {
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }

  .bottom-tab {
    min-height: 54px;
  }

  .mobile-secondary-grid {
    grid-template-columns: 1fr;
  }
}

.sheet-rise-enter-active,
.sheet-rise-leave-active {
  transition:
    transform 0.22s ease,
    opacity 0.22s ease;
}

.sheet-rise-enter-from,
.sheet-rise-leave-to {
  transform: translateY(12px);
  opacity: 0;
}

.sheet-fade-enter-active,
.sheet-fade-leave-active {
  transition: opacity 0.22s ease;
}

.sheet-fade-enter-from,
.sheet-fade-leave-to {
  opacity: 0;
}
</style>

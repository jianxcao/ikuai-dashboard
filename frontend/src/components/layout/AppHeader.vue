<template>
  <header class="app-shell-nav">
    <nav class="side-nav glass-panel" aria-label="主导航">
      <div class="brand">
        <img class="brand-mark" src="/ikuai-icon.svg" alt="" />
        <div class="brand-copy">
          <span class="brand-text">iKuai Dashboard</span>
          <span class="brand-subtitle">爱快实时监控</span>
          <span class="brand-version">构建 {{ buildVersion }}</span>
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

      <div v-if="!configurationOnly" class="nav-tabs" role="tablist" aria-label="监控视图">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          type="button"
          :class="['nav-tab', { active: currentTab === tab.id }]"
          :aria-selected="currentTab === tab.id"
          role="tab"
          @click="emit('tab-change', tab.id)"
        >
          <component :is="tab.icon" :size="15" />
          <span>{{ tab.label }}</span>
        </button>
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

    <div v-if="!configurationOnly" class="mobile-router-bar glass-panel" aria-label="移动端服务器切换">
      <div class="mobile-router-meta">
        <div class="mobile-router-name">
          <Server :size="15" />
          <span>{{ activeRouter?.name || '未选择服务器' }}</span>
        </div>
        <span class="mobile-build-version">构建 {{ buildVersion }}</span>
      </div>
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

    <nav v-if="!configurationOnly" class="bottom-nav glass-panel" aria-label="移动主导航">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        type="button"
        :class="['bottom-tab', { active: currentTab === tab.id }]"
        :aria-current="currentTab === tab.id ? 'page' : undefined"
        @click="emit('tab-change', tab.id)"
      >
        <component :is="tab.icon" :size="19" />
        <span>{{ tab.mobileLabel }}</span>
      </button>
    </nav>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Activity, BarChart3, LogOut, Moon, Network, Server, Settings, ShieldCheck, SlidersHorizontal, SunMedium, Waypoints, Wifi } from 'lucide-vue-next'
import { useRouterConfig } from '@/composables/useRouterConfig'
import { useAuth } from '@/composables/useAuth'
import type { AppTheme } from '@/composables/useTheme'

withDefaults(defineProps<{
  currentTab?: string
  connected?: boolean
  configurationOnly?: boolean
  theme: AppTheme
  themes: Array<{ id: AppTheme; label: string; shortLabel: string }>
}>(), {
  currentTab: 'interface',
  connected: false,
  configurationOnly: false,
})

const emit = defineEmits<{
  (event: 'tab-change', value: string): void
  (event: 'theme-change', value: AppTheme): void
}>()

const { config, activeRouter, switching, status, activate } = useRouterConfig()
const { authEnabled, logout } = useAuth()
const buildVersion = import.meta.env.VITE_APP_VERSION || 'dev'

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

const tabs = [
  { id: 'interface', label: '首页看板', mobileLabel: '首页', icon: Activity },
  { id: 'lan', label: '局域网客户端', mobileLabel: '终端', icon: Wifi },
  { id: 'network-map', label: '网络拓扑', mobileLabel: '拓扑', icon: Network },
  { id: 'security-hub', label: '安全中心', mobileLabel: '安全', icon: ShieldCheck },
  { id: 'multi-wan', label: '多 WAN', mobileLabel: 'WAN', icon: Waypoints },
  { id: 'insights', label: '监控分析', mobileLabel: '分析', icon: BarChart3 },
  { id: 'resources', label: '配置管理', mobileLabel: '配置', icon: SlidersHorizontal },
  { id: 'settings', label: '路由器设置', mobileLabel: '设置', icon: Settings },
]
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
  width: 248px;
  padding: 16px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.brand-copy {
  min-width: 0;
}

.brand-mark {
  width: 42px;
  height: 42px;
  flex: 0 0 auto;
}

.brand-text {
  display: block;
  color: var(--text-primary);
  font-size: 15px;
  font-weight: 760;
  letter-spacing: 0;
  white-space: nowrap;
}

.brand-subtitle {
  display: block;
  margin-top: 3px;
  color: var(--text-tertiary);
  font-size: 12px;
  font-weight: 650;
}

.brand-version,
.mobile-build-version {
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
  border-radius: 14px;
  background: var(--control-bg);
}

.router-switcher-head,
.mobile-router-name {
  display: flex;
  align-items: center;
  gap: 7px;
  min-width: 0;
  color: var(--text-tertiary);
  font-size: 12px;
  font-weight: 760;
}

.mobile-router-meta {
  min-width: 0;
}

.router-select,
.mobile-router-select {
  width: 100%;
  min-height: 38px;
  padding: 0 40px 0 10px;
  color: var(--text-primary);
  border: 1px solid var(--control-border);
  border-radius: 10px;
  background: var(--glass-bg-strong);
  font-size: 13px;
  font-weight: 720;
  appearance: none;
  background-image:
    linear-gradient(45deg, transparent 50%, currentColor 50%),
    linear-gradient(135deg, currentColor 50%, transparent 50%),
    var(--glass-bg-strong);
  background-position:
    calc(100% - 23px) 50%,
    calc(100% - 16px) 50%,
    0 0;
  background-size: 7px 7px, 7px 7px, auto;
  background-repeat: no-repeat;
}

.router-select:disabled,
.mobile-router-select:disabled {
  opacity: 0.62;
}

.nav-tabs {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 8px;
}

.nav-tab {
  display: inline-flex;
  align-items: center;
  justify-content: flex-start;
  gap: 7px;
  min-height: 42px;
  width: 100%;
  padding: 0 12px;
  border: 1px solid transparent;
  border-radius: 12px;
  background: transparent;
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 620;
  cursor: pointer;
  white-space: nowrap;
  transition: color 0.2s ease, background 0.2s ease, border-color 0.2s ease, transform 0.2s ease;
}

.nav-tab:hover {
  color: var(--text-primary);
  background: var(--control-bg);
}

.nav-tab.active {
  color: var(--text-primary);
  background: var(--control-active);
  border-color: var(--system-blue);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.24), 0 0 24px var(--system-blue-dim);
}

.header-actions {
  display: flex;
  align-items: stretch;
  flex-direction: column;
  justify-content: flex-end;
  gap: 10px;
  margin-top: auto;
  min-width: 0;
}

.status-area {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 38px;
  padding: 0 12px;
  border: 1px solid var(--control-border);
  border-radius: 999px;
  background: var(--control-bg);
}

.status-text {
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 650;
  white-space: nowrap;
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

.logout-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  min-height: 32px;
  padding: 0 12px;
  border: 1px solid var(--control-border);
  border-radius: 999px;
  background: var(--control-bg);
  color: var(--text-tertiary);
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  transition: color 0.18s ease, background 0.18s ease, border-color 0.18s ease;
  white-space: nowrap;
}

.logout-btn:hover {
  color: var(--system-red);
  border-color: color-mix(in srgb, var(--system-red) 42%, var(--control-border));
  background: var(--system-red-dim);
}

.bottom-nav {
  display: none;
}

.mobile-router-bar {
  display: none;
}

@media (max-width: 920px) {
  .side-nav {
    display: none;
  }

  .mobile-router-bar {
    position: fixed;
    top: max(10px, env(safe-area-inset-top));
    right: max(10px, env(safe-area-inset-right));
    left: max(10px, env(safe-area-inset-left));
    z-index: 60;
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(132px, 42%);
    gap: 10px;
    align-items: center;
    padding: 10px;
    border-radius: 18px;
  }

  .mobile-router-name {
    color: var(--text-secondary);
  }

  .mobile-router-name span {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .mobile-build-version {
    min-height: 20px;
    margin-top: 4px;
    padding: 0 7px;
    font-size: 10px;
  }

  .bottom-nav {
    position: fixed;
    right: max(10px, env(safe-area-inset-right));
    bottom: max(10px, env(safe-area-inset-bottom));
    left: max(10px, env(safe-area-inset-left));
    z-index: 60;
    display: grid;
    grid-template-columns: repeat(8, minmax(0, 1fr));
    gap: 4px;
    padding: 8px;
    border-radius: 24px;
  }

  .bottom-tab {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-direction: column;
    gap: 3px;
    min-width: 0;
    min-height: 50px;
    color: var(--text-tertiary);
    border: 0;
    border-radius: 16px;
    background: transparent;
    cursor: pointer;
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
}

@media (max-width: 640px) {
  .bottom-nav {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@media (max-width: 380px) {
  .mobile-router-bar {
    grid-template-columns: 1fr;
  }

  .bottom-nav {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
</style>

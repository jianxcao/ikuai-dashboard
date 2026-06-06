<template>
  <div class="app-layout">
    <AppHeader
      :currentTab="currentTab"
      :connected="true"
      :theme="theme"
      :themes="themes"
      :configurationOnly="configurationOnly"
      :tabs="tabs"
      @tab-change="currentTab = $event"
      @theme-change="setTheme"
    />

    <main class="main-content">
      <section class="content-shell">
        <KeepAlive>
          <component :is="currentComponent" />
        </KeepAlive>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, defineAsyncComponent, watch } from 'vue'
import {
  Activity,
  BarChart3,
  Network,
  Settings,
  ShieldCheck,
  SlidersHorizontal,
  Waypoints,
  Wifi
} from 'lucide-vue-next'
import AppHeader from './AppHeader.vue'
import MonitorInterface from '@/views/MonitorInterface.vue'
import MonitorLan from '@/views/MonitorLan.vue'
import { useTheme } from '@/composables/useTheme'
import { useRouterConfig } from '@/composables/useRouterConfig'

const NetworkMap = defineAsyncComponent(() => import('@/views/NetworkMap.vue'))
const SecurityHub = defineAsyncComponent(() => import('@/views/SecurityHub.vue'))
const MultiWan = defineAsyncComponent(() => import('@/views/MultiWan.vue'))
const MonitorInsights = defineAsyncComponent(() => import('@/views/MonitorInsights.vue'))
const ConfigResources = defineAsyncComponent(() => import('@/views/ConfigResources.vue'))
const RouterSettings = defineAsyncComponent(() => import('@/views/RouterSettings.vue'))

type AppTab = {
  id: string
  label: string
  mobileLabel: string
  icon: unknown
  primary?: boolean
}

const tabs: AppTab[] = [
  {
    id: 'interface',
    label: '首页看板',
    mobileLabel: '首页',
    icon: Activity,
    primary: true
  },
  {
    id: 'lan',
    label: '局域网客户端',
    mobileLabel: '终端',
    icon: Wifi,
    primary: true
  },
  {
    id: 'network-map',
    label: '网络拓扑',
    mobileLabel: '拓扑',
    icon: Network
  },
  {
    id: 'security-hub',
    label: '安全中心',
    mobileLabel: '安全',
    icon: ShieldCheck
  },
  {
    id: 'multi-wan',
    label: '多 WAN',
    mobileLabel: 'WAN',
    icon: Waypoints
  },
  {
    id: 'insights',
    label: '监控分析',
    mobileLabel: '分析',
    icon: BarChart3,
    primary: true
  },
  {
    id: 'resources',
    label: '配置管理',
    mobileLabel: '配置',
    icon: SlidersHorizontal
  },
  {
    id: 'settings',
    label: '路由器设置',
    mobileLabel: '设置',
    icon: Settings,
    primary: true
  }
]

const currentTab = ref('interface')
const { theme, themes, setTheme } = useTheme()
const { loading: configLoading, config } = useRouterConfig()

const configurationOnly = computed(() => !configLoading.value && config.value.routers.length === 0)

watch(
  configurationOnly,
  (enabled) => {
    if (enabled) currentTab.value = 'settings'
  },
  { immediate: true }
)

const currentComponent = computed(() => {
  if (configurationOnly.value) return RouterSettings
  switch (currentTab.value) {
    case 'interface':
      return MonitorInterface
    case 'lan':
      return MonitorLan
    case 'network-map':
      return NetworkMap
    case 'security-hub':
      return SecurityHub
    case 'multi-wan':
      return MultiWan
    case 'insights':
      return MonitorInsights
    case 'resources':
      return ConfigResources
    case 'settings':
      return RouterSettings
    default:
      return MonitorInterface
  }
})
</script>

<style scoped>
.app-layout {
  min-height: 100vh;
  display: flex;
}

.main-content {
  flex: 1;
  width: 100%;
  max-width: 1580px;
  margin: 0 auto;
  padding: 28px 28px 36px;
}

.content-shell {
  min-width: 0;
}

@media (min-width: 1101px) {
  .main-content {
    margin-left: 300px;
    width: calc(100% - 300px);
  }
}

@media (max-width: 1100px) {
  .main-content {
    padding: calc(92px + max(10px, env(safe-area-inset-top))) max(18px, env(safe-area-inset-right))
      calc(112px + max(14px, env(safe-area-inset-bottom))) max(18px, env(safe-area-inset-left));
  }
}

@media (max-width: 760px) {
  .main-content {
    padding: calc(94px + max(10px, env(safe-area-inset-top))) max(14px, env(safe-area-inset-right))
      calc(112px + max(14px, env(safe-area-inset-bottom))) max(14px, env(safe-area-inset-left));
  }
}
</style>

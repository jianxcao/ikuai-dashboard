<template>
  <div class="app-layout">
    <AppHeader
      :currentTab="currentTab"
      :connected="true"
      :theme="theme"
      :themes="themes"
      :configurationOnly="configurationOnly"
      @tab-change="currentTab = $event"
      @theme-change="setTheme"
    />

    <main class="main-content">
      <KeepAlive>
        <component :is="currentComponent" />
      </KeepAlive>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, defineAsyncComponent, watch } from 'vue'
import AppHeader from './AppHeader.vue'
import MonitorInterface from '@/views/MonitorInterface.vue'
import MonitorLan from '@/views/MonitorLan.vue'
import { useTheme } from '@/composables/useTheme'
import { useRouterConfig } from '@/composables/useRouterConfig'

// Lazy load new components to improve performance
const NetworkMap = defineAsyncComponent(() => import('@/views/NetworkMap.vue'))
const SecurityHub = defineAsyncComponent(() => import('@/views/SecurityHub.vue'))
const MultiWan = defineAsyncComponent(() => import('@/views/MultiWan.vue'))
const MonitorInsights = defineAsyncComponent(() => import('@/views/MonitorInsights.vue'))
const ConfigResources = defineAsyncComponent(() => import('@/views/ConfigResources.vue'))
const RouterSettings = defineAsyncComponent(() => import('@/views/RouterSettings.vue'))

const currentTab = ref('interface')
const { theme, themes, setTheme } = useTheme()
const { loading: configLoading, config } = useRouterConfig()

const configurationOnly = computed(() => !configLoading.value && config.value.routers.length === 0)

watch(configurationOnly, (enabled) => {
  if (enabled) {
    currentTab.value = 'settings'
  }
}, { immediate: true })

const currentComponent = computed(() => {
  if (configurationOnly.value) return RouterSettings
  switch (currentTab.value) {
    case 'interface': return MonitorInterface
    case 'lan': return MonitorLan
    case 'network-map': return NetworkMap
    case 'security-hub': return SecurityHub
    case 'multi-wan': return MultiWan
    case 'insights': return MonitorInsights
    case 'resources': return ConfigResources
    case 'settings': return RouterSettings
    default: return MonitorInterface
  }
})
</script>

<style scoped>
.app-layout {
  min-height: 100vh;
  display: flex;
  align-items: stretch;
}

.main-content {
  flex: 1;
  max-width: 1400px;
  width: 100%;
  margin: 0 auto 0 0;
  padding: 24px;
}

@media (min-width: 921px) {
  .main-content {
    margin-left: 280px;
    width: calc(100% - 280px);
  }
}

@media (max-width: 920px) {
  .main-content {
    padding:
      calc(72px + max(10px, env(safe-area-inset-top)))
      max(16px, env(safe-area-inset-right))
      calc(96px + max(10px, env(safe-area-inset-bottom)))
      max(16px, env(safe-area-inset-left));
  }
}

@media (max-width: 640px) {
  .main-content {
    padding-bottom: calc(138px + max(10px, env(safe-area-inset-bottom)));
  }
}

@media (max-width: 420px) {
  .main-content {
    padding:
      calc(72px + max(10px, env(safe-area-inset-top)))
      max(12px, env(safe-area-inset-right))
      calc(138px + max(10px, env(safe-area-inset-bottom)))
      max(12px, env(safe-area-inset-left));
  }
}

@media (max-width: 380px) {
  .main-content {
    padding-bottom: calc(190px + max(10px, env(safe-area-inset-bottom)));
  }
}
</style>

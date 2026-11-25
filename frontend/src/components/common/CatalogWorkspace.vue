<template>
  <section class="workspace-page">
    <header class="ws-head v4-panel">
      <div>
        <span class="ws-eyebrow font-mono">{{ eyebrow }}</span>
        <h1>{{ title }}</h1>
        <p>{{ subtitle }}</p>
      </div>
      <slot name="header-extra" />
    </header>

    <slot name="banner" />

    <div ref="subtabBarEl" class="subtab-bar" role="tablist">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        type="button"
        role="tab"
        class="subtab-pill"
        :class="{ active: activeKey === tab.key }"
        :aria-selected="activeKey === tab.key"
        @click="selectSubTab(tab.key)"
      >
        <span v-if="tab.groupLabel" class="group-tag">{{ tab.groupLabel }}</span>
        <span>{{ tab.label }}</span>
      </button>
    </div>

    <SingletonConfigPanel
      v-if="activeSpec?.kind === 'singleton' || activeTab?.kind === 'singleton'"
      :key="`singleton-${activeKey}`"
      :resourceKey="activeKey"
    />
    <ResourceCollectionPanel v-else :key="`collection-${activeKey}`" :resourceKey="activeKey" />
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onActivated, ref, watch } from 'vue'
import ResourceCollectionPanel from './ResourceCollectionPanel.vue'
import SingletonConfigPanel from './SingletonConfigPanel.vue'
import { useAppRoute } from '@/composables/useAppRoute'
import { useCatalog } from '@/composables/useCatalog'

export interface WorkspaceTab {
  key: string
  label: string
  kind?: 'collection' | 'singleton'
  groupLabel?: string
}

const props = defineProps<{
  eyebrow: string
  title: string
  subtitle: string
  tabs: WorkspaceTab[]
  routeTab?: string
  initialSubResource?: string
  navSeq?: number
}>()

const { currentTab, syncSubTab } = useAppRoute()
const ownerTab = props.routeTab || currentTab.value
const { getSpec } = useCatalog()
const initialValid =
  props.initialSubResource && props.tabs.some((t) => t.key === props.initialSubResource)
    ? props.initialSubResource
    : props.tabs[0]?.key || ''
const activeKey = ref(initialValid)
const subtabBarEl = ref<HTMLElement | null>(null)

if (currentTab.value === ownerTab && activeKey.value && !props.initialSubResource) {
  syncSubTab(ownerTab, activeKey.value, { replace: true })
}

function selectSubTab(key: string) {
  if (!props.tabs.some((t) => t.key === key)) return
  activeKey.value = key
  syncSubTab(ownerTab, key, { replace: false })
}

watch(activeKey, async () => {
  await nextTick()
  const activeEl = subtabBarEl.value?.querySelector('.subtab-pill.active') as HTMLElement | null
  activeEl?.scrollIntoView({ block: 'nearest', inline: 'center', behavior: 'smooth' })
})

watch(
  () => [props.initialSubResource, props.navSeq] as const,
  ([next]) => {
    if (currentTab.value !== ownerTab) return
    if (next && props.tabs.some((t) => t.key === next)) {
      activeKey.value = next
    } else if (!next && activeKey.value) {
      syncSubTab(ownerTab, activeKey.value, { replace: true })
    }
  }
)

onActivated(() => {
  if (currentTab.value === ownerTab && !props.initialSubResource && activeKey.value) {
    syncSubTab(ownerTab, activeKey.value, { replace: true })
  }
})

const activeTab = computed(() => props.tabs.find((t) => t.key === activeKey.value))
const activeSpec = computed(() => getSpec(activeKey.value))

defineExpose({
  setTab(key: string) {
    selectSubTab(key)
  },
})
</script>

<style scoped>
.workspace-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.ws-head {
  padding: 18px 20px;
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  align-items: center;
  gap: 14px;
}

.ws-eyebrow {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--text-faint);
}

.ws-head h1 {
  margin: 4px 0 0;
  font-size: 21px;
  font-weight: 800;
  color: var(--text-strong);
}

.ws-head p {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--text-muted);
}

.subtab-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding: 5px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--seg-track-border);
  background: var(--seg-track-bg);
}

.subtab-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-height: 32px;
  padding: 0 12px;
  border-radius: var(--radius-control);
  border: 1px solid transparent;
  background: transparent;
  color: var(--text-muted);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: color 0.14s var(--ease-ui), background-color 0.14s var(--ease-ui), box-shadow 0.14s var(--ease-ui);
}

.subtab-pill:hover {
  color: var(--text-strong);
}

.subtab-pill.active {
  color: var(--seg-active-fg);
  border-color: var(--seg-active-border);
  background: var(--seg-active-bg);
  box-shadow: var(--seg-active-shadow);
}

.group-tag {
  padding: 1px 6px;
  border-radius: 999px;
  background: var(--surface-input);
  border: 1px solid var(--divider-line);
  font-size: 10px;
  color: var(--text-muted);
}

.subtab-pill.active .group-tag {
  background: var(--accent-primary-soft);
  border-color: transparent;
  color: var(--accent-text);
}

@media (max-width: 768px) {
  .ws-head {
    padding: 14px;
    gap: 10px;
  }

  .ws-head h1 {
    font-size: 18px;
  }

  .ws-head p {
    font-size: 12px;
  }

  .subtab-bar {
    flex-wrap: nowrap;
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
    scrollbar-width: none;
    padding: 5px;
  }

  .subtab-bar::-webkit-scrollbar {
    display: none;
  }

  .subtab-pill {
    flex-shrink: 0;
    white-space: nowrap;
    min-height: 32px;
    padding: 0 11px;
  }
}
</style>

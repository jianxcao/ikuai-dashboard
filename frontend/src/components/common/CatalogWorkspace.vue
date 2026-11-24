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
        @click="activeKey = tab.key"
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
import { computed, nextTick, ref, watch } from 'vue'
import ResourceCollectionPanel from './ResourceCollectionPanel.vue'
import SingletonConfigPanel from './SingletonConfigPanel.vue'
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
  initialSubResource?: string
  navSeq?: number
}>()

const { getSpec } = useCatalog()
const activeKey = ref(props.tabs[0]?.key || '')
const subtabBarEl = ref<HTMLElement | null>(null)

watch(activeKey, async () => {
  await nextTick()
  const activeEl = subtabBarEl.value?.querySelector('.subtab-pill.active') as HTMLElement | null
  activeEl?.scrollIntoView({ block: 'nearest', inline: 'center', behavior: 'smooth' })
})

watch(
  () => [props.initialSubResource, props.navSeq] as const,
  ([next]) => {
    if (next && props.tabs.some((t) => t.key === next)) {
      activeKey.value = next
    }
  },
  { immediate: true }
)

const activeTab = computed(() => props.tabs.find((t) => t.key === activeKey.value))
const activeSpec = computed(() => getSpec(activeKey.value))

defineExpose({
  setTab(key: string) {
    if (props.tabs.some((t) => t.key === key)) {
      activeKey.value = key
    }
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
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--accent-text);
}

.ws-head h1 {
  margin: 4px 0 0;
  font-size: 22px;
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
  gap: 6px;
  padding: 6px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  background: var(--surface-input);
  box-shadow: var(--inner-sheen);
}

.subtab-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-height: 34px;
  padding: 0 13px;
  border-radius: var(--radius-control);
  border: 1px solid transparent;
  background: transparent;
  color: var(--text-normal);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.16s ease;
}

.subtab-pill:hover {
  color: var(--text-strong);
  background: var(--surface-hover);
}

.subtab-pill.active {
  color: var(--btn-primary-fg);
  border-color: var(--btn-primary-border);
  background: var(--btn-primary-bg);
  box-shadow: var(--btn-primary-shadow);
}

.subtab-pill.active > * {
  color: var(--btn-primary-fg);
}

.group-tag {
  padding: 1px 6px;
  border-radius: 999px;
  background: var(--surface-input);
  border: 1px solid var(--border-control);
  font-size: 10px;
  color: var(--text-muted);
}

.subtab-pill.active .group-tag {
  background: rgba(255, 255, 255, 0.22);
  border-color: rgba(255, 255, 255, 0.40);
  color: var(--btn-primary-fg);
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

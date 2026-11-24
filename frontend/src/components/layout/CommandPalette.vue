<template>
  <teleport to="body">
    <div v-if="open" class="palette-backdrop" @click.self="emit('close')">
      <div class="palette-dialog v4-panel" role="dialog" aria-label="快速指令面板">
        <div class="palette-search">
          <Search :size="16" class="search-icon" />
          <input
            ref="inputRef"
            v-model.trim="query"
            type="text"
            placeholder="搜索功能模块、ikuai-cli 命令、V4 节点或主题..."
            @keydown.esc="emit('close')"
            @keydown.enter="runFirst"
          />
          <kbd class="kbd">ESC</kbd>
        </div>

        <div class="palette-results">
          <div v-if="filteredEntries.length === 0" class="palette-empty">
            未找到匹配 "{{ query }}" 的页面、CLI 命令或节点
          </div>

          <button
            v-for="entry in filteredEntries"
            :key="entry.id"
            type="button"
            class="palette-item"
            @click="execute(entry)"
          >
            <div class="entry-left">
              <span class="v4-badge" :class="entry.badgeClass">{{ entry.category }}</span>
              <div class="entry-copy">
                <strong>{{ entry.title }}</strong>
                <span class="font-mono">{{ entry.subtitle }}</span>
              </div>
            </div>
            <ArrowRight :size="14" class="entry-arrow" />
          </button>
        </div>
      </div>
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { ArrowRight, Search } from 'lucide-vue-next'
import { useCatalog } from '@/composables/useCatalog'
import { useFleet } from '@/composables/useFleet'
import { useTheme } from '@/composables/useTheme'

const props = defineProps<{
  open: boolean
  navTabs: Array<{ id: string; label: string; desc: string }>
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'navigate', payload: { tab: string; subResource?: string }): void
}>()

const { catalog } = useCatalog()
const { config, activateRouter } = useFleet()
const { themes, setTheme } = useTheme()

const query = ref('')
const inputRef = ref<HTMLInputElement | null>(null)

watch(
  () => props.open,
  async (val) => {
    if (val) {
      query.value = ''
      await nextTick()
      inputRef.value?.focus()
    }
  }
)

const groupToTab: Record<string, string> = {
  network: 'network',
  routing: 'routing',
  qos: 'routing',
  security: 'security',
  users: 'terminals',
  vpn: 'vpn-wireless',
  wireless: 'vpn-wireless',
  objects: 'vpn-wireless',
  advanced: 'system',
  log: 'logs',
  system: 'system',
}

interface PaletteEntry {
  id: string
  category: string
  badgeClass: string
  title: string
  subtitle: string
  action: () => void
}

const allEntries = computed<PaletteEntry[]>(() => {
  const entries: PaletteEntry[] = []

  for (const tab of props.navTabs) {
    entries.push({
      id: `tab:${tab.id}`,
      category: '工作区',
      badgeClass: 'v4-badge-accent',
      title: tab.label,
      subtitle: tab.desc,
      action: () => emit('navigate', { tab: tab.id }),
    })
  }

  for (const spec of catalog.value) {
    const tab = groupToTab[spec.group] || 'overview'
    entries.push({
      id: `res:${spec.key}`,
      category: 'CLI 资源',
      badgeClass: '',
      title: spec.label,
      subtitle: spec.cli_command,
      action: () => emit('navigate', { tab, subResource: spec.key }),
    })
  }

  for (const r of config.value.routers) {
    entries.push({
      id: `router:${r.id}`,
      category: '切换 V4 节点',
      badgeClass: 'v4-badge-up',
      title: r.name,
      subtitle: r.url,
      action: () => {
        void activateRouter(r.id)
      },
    })
  }

  for (const t of themes) {
    entries.push({
      id: `theme:${t.id}`,
      category: '主题皮肤',
      badgeClass: 'v4-badge-warn',
      title: `${t.label} (${t.englishName})`,
      subtitle: t.description,
      action: () => setTheme(t.id),
    })
  }

  return entries
})

const filteredEntries = computed(() => {
  const q = query.value.toLowerCase()
  if (!q) return allEntries.value.slice(0, 18)
  return allEntries.value
    .filter((e) => (e.title + ' ' + e.subtitle + ' ' + e.category).toLowerCase().includes(q))
    .slice(0, 24)
})

function execute(entry: PaletteEntry) {
  entry.action()
  emit('close')
}

function runFirst() {
  const first = filteredEntries.value[0]
  if (first) execute(first)
}
</script>

<style scoped>
.palette-backdrop {
  position: fixed;
  inset: 0;
  z-index: 300;
  background: rgba(2, 6, 14, 0.62);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: min(14vh, 110px);
}

.palette-dialog {
  width: min(640px, calc(100vw - 24px));
  overflow: hidden;
  background: var(--surface-popover);
}

.palette-search {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.search-icon {
  color: var(--accent-text);
  flex-shrink: 0;
}

.palette-search input {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--text-strong);
  font-size: 14px;
  outline: none;
}

.kbd {
  padding: 2px 7px;
  border-radius: 6px;
  border: 1px solid var(--border-subtle);
  background: var(--surface-input);
  color: var(--text-muted);
  font-size: 11px;
}

.palette-results {
  max-height: 420px;
  overflow-y: auto;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.palette-empty {
  padding: 28px;
  text-align: center;
  color: var(--text-muted);
}

.palette-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  width: 100%;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid transparent;
  background: transparent;
  color: var(--text-strong);
  cursor: pointer;
  text-align: left;
}

.palette-item:hover {
  border-color: var(--border-subtle);
  background: var(--accent-primary-soft);
}

.entry-left {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.entry-copy {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.entry-copy strong {
  font-size: 13px;
}

.entry-copy span {
  font-size: 11px;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.entry-arrow {
  color: var(--text-muted);
  flex-shrink: 0;
}

@media (max-width: 640px) {
  .palette-backdrop {
    padding: 10px;
    padding-top: max(14px, env(safe-area-inset-top));
  }

  .palette-dialog {
    width: 100%;
    max-height: 82dvh;
    display: flex;
    flex-direction: column;
  }

  .palette-results {
    flex: 1;
    max-height: none;
  }

  .kbd {
    display: none;
  }
}
</style>

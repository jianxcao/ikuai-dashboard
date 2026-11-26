<template>
  <teleport to="body">
    <div v-if="open" class="drawer-backdrop center" @click.self="emit('close')">
      <div class="confirm-dialog v4-panel modal-wide" role="dialog" :aria-label="`对象组引用关系 · ${groupName}`">
        <div class="modal-top">
          <h4>对象组引用关系 · {{ groupName }}</h4>
          <button type="button" class="icon-btn" aria-label="关闭弹窗" @click="emit('close')">
            <X :size="14" />
          </button>
        </div>

        <div v-if="loading" class="loading-state"><div class="spinner" /></div>
        <div v-else-if="error" class="error-banner">{{ error }}</div>
        <div v-else-if="rows.length === 0" class="empty-state">当前未被任何策略或分流规则引用</div>
        <template v-else>
          <div v-if="refJumpTargets.length > 0" class="ref-jump-bar">
            <button
              v-for="target in refJumpTargets"
              :key="target.key"
              type="button"
              class="v4-btn btn-xs"
              @click="handleJump(target.tab, target.key)"
            >
              <ExternalLink :size="12" />
              <span>前往「{{ target.label }}」</span>
            </button>
          </div>
          <pre class="json-pre font-mono">{{ JSON.stringify(rows, null, 2) }}</pre>
        </template>
      </div>
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ExternalLink, X } from 'lucide-vue-next'
import { fetchObjectRefs } from '@/api/v4'
import { resolveResourceTab } from '@/composables/useAppRoute'
import { useCatalog } from '@/composables/useCatalog'

const props = defineProps<{
  open: boolean
  resourceKey: string
  groupName: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'jump', payload: { tab: string; key: string }): void
}>()

const { getSpec } = useCatalog()

const loading = ref(false)
const error = ref<string | null>(null)
const rows = ref<Array<Record<string, unknown>>>([])
let reqSeq = 0

const refJumpTargets = computed(() => {
  const seen = new Set<string>()
  const out: Array<{ key: string; label: string; tab: string }> = []
  for (const r of rows.value) {
    const modKey = String(r.module ?? '').trim()
    if (!modKey || seen.has(modKey)) continue
    const targetSpec = getSpec(modKey)
    if (!targetSpec) continue
    seen.add(modKey)
    out.push({ key: modKey, label: targetSpec.label, tab: resolveResourceTab(targetSpec.group) })
  }
  return out
})

async function loadRefs() {
  if (!props.open || !props.resourceKey || !props.groupName) return
  const seq = ++reqSeq
  loading.value = true
  error.value = null
  rows.value = []
  try {
    const res = await fetchObjectRefs(props.resourceKey, props.groupName)
    if (seq !== reqSeq) return
    rows.value = res.rows || []
  } catch (e: unknown) {
    if (seq !== reqSeq) return
    if (e && typeof e === 'object' && 'response' in e) {
      const resp = (e as { response?: { data?: { message?: string } } }).response
      if (resp?.data?.message) {
        error.value = resp.data.message
        return
      }
    }
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    if (seq === reqSeq) loading.value = false
  }
}

function handleJump(tab: string, key: string) {
  emit('jump', { tab, key })
  emit('close')
}

watch(
  () => [props.open, props.resourceKey, props.groupName] as const,
  ([isOpen]) => {
    if (isOpen) void loadRefs()
    else reqSeq += 1
  },
  { immediate: true }
)
</script>

<style scoped>
.drawer-backdrop {
  position: fixed;
  inset: 0;
  z-index: 255;
  background: rgba(6, 10, 20, 0.42);
  -webkit-backdrop-filter: blur(8px);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.confirm-dialog {
  width: min(560px, calc(100vw - 32px));
  max-height: 86dvh;
  overflow-y: auto;
  padding: 20px;
  background: var(--surface-popover);
  border-color: var(--chrome-border);
  box-shadow: var(--chrome-shadow);
  -webkit-backdrop-filter: var(--backdrop-chrome);
  backdrop-filter: var(--backdrop-chrome);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.modal-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-top h4 {
  margin: 0;
  font-size: 15px;
  color: var(--text-strong);
}

.icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-control);
  background: var(--surface-btn);
  color: var(--text-normal);
  cursor: pointer;
}

.icon-btn:hover {
  border-color: var(--accent-primary);
  color: var(--text-strong);
}

.ref-jump-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.json-pre {
  margin: 0;
  font-size: 11.5px;
  line-height: 1.55;
  overflow-x: auto;
  color: var(--text-normal);
  padding: 12px;
  border-radius: var(--radius-md);
  background: var(--surface-input);
  border: 1px solid var(--divider-line);
}

.empty-state {
  padding: 28px 16px;
  text-align: center;
  color: var(--text-muted);
}

@media (max-width: 640px) {
  .confirm-dialog {
    width: calc(100vw - 20px);
    padding: 16px;
  }
}
</style>

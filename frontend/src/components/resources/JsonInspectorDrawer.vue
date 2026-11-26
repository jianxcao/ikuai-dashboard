<template>
  <teleport to="body">
    <div v-if="row" class="drawer-backdrop" @click.self="emit('close')">
      <div class="drawer-panel v4-panel" role="dialog" aria-label="记录详情 JSON">
        <div class="drawer-head">
          <div>
            <strong>{{ label }} · 记录完整字段</strong>
            <span class="font-mono">{{ v4Path }}</span>
          </div>
          <button type="button" class="icon-btn" aria-label="关闭弹窗" @click="emit('close')">
            <X :size="16" />
          </button>
        </div>
        <div class="drawer-body">
          <pre class="json-pre font-mono">{{ JSON.stringify(row, null, 2) }}</pre>
        </div>
        <div class="drawer-foot">
          <button type="button" class="v4-btn" @click="copyJson">
            <Copy :size="14" />
            <span>复制 JSON</span>
          </button>
          <button
            v-if="canEdit"
            type="button"
            class="v4-btn v4-btn-primary"
            @click="emit('edit', row)"
          >
            <Pencil :size="14" />
            <span>编辑此记录</span>
          </button>
        </div>
      </div>
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { Copy, Pencil, X } from 'lucide-vue-next'

const props = defineProps<{
  row: Record<string, unknown> | null
  label?: string
  v4Path?: string
  canEdit?: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'edit', row: Record<string, unknown>): void
  (e: 'copied'): void
}>()

async function copyJson() {
  if (!props.row) return
  try {
    await navigator.clipboard.writeText(JSON.stringify(props.row, null, 2))
    emit('copied')
  } catch {
    // ignore clipboard errors
  }
}
</script>

<style scoped>
.drawer-backdrop {
  position: fixed;
  inset: 0;
  z-index: 250;
  background: rgba(6, 10, 20, 0.42);
  -webkit-backdrop-filter: blur(8px);
  backdrop-filter: blur(8px);
  display: flex;
  justify-content: flex-end;
}

.drawer-panel {
  width: min(540px, 100vw);
  height: 100%;
  border-radius: var(--radius-xl) 0 0 var(--radius-xl);
  display: flex;
  flex-direction: column;
  background: var(--surface-popover);
  border-color: var(--chrome-border);
  box-shadow: var(--chrome-shadow);
  -webkit-backdrop-filter: var(--backdrop-chrome);
  backdrop-filter: var(--backdrop-chrome);
}

.drawer-head {
  padding: 16px 20px;
  border-bottom: 1px solid var(--divider-line);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.drawer-head > div > strong {
  display: block;
  font-size: 15px;
  color: var(--text-strong);
}

.drawer-head > div > span {
  font-size: 11px;
  color: var(--text-muted);
}

.drawer-body {
  flex: 1;
  overflow-y: auto;
  padding: 18px 20px;
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

.drawer-foot {
  padding: 14px 20px;
  border-top: 1px solid var(--divider-line);
  display: flex;
  justify-content: flex-end;
  gap: 10px;
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

@media (max-width: 640px) {
  .drawer-backdrop {
    align-items: flex-end;
  }

  .drawer-panel {
    width: 100%;
    height: auto;
    max-height: 88dvh;
    border-radius: var(--radius-xl) var(--radius-xl) 0 0;
  }

  .drawer-body {
    padding: 14px 16px;
  }

  .drawer-foot {
    padding: 12px 16px max(12px, env(safe-area-inset-bottom));
  }
}
</style>

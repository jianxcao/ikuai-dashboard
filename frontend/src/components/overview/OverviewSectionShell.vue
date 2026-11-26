<template>
  <div class="card-title-row">
    <div class="card-title">
      <slot name="icon" />
      <h3>{{ title }}</h3>
    </div>
    <div class="title-right-group">
      <slot name="actions" />
      <div v-if="!layoutLocked" class="drag-ctl-group">
        <span class="drag-handle-pill" title="按住拖动与其他模块互换位置">
          <GripVertical :size="13" />
          <span>拖动</span>
        </span>
        <button
          type="button"
          class="drag-step-btn"
          title="前移"
          :disabled="!canMovePrev"
          @click.stop="emit('move', -1)"
        >
          <ArrowLeft :size="12" />
        </button>
        <button
          type="button"
          class="drag-step-btn"
          title="后移"
          :disabled="!canMoveNext"
          @click.stop="emit('move', 1)"
        >
          <ArrowRight :size="12" />
        </button>
      </div>
    </div>
  </div>
  <slot />
</template>

<script setup lang="ts">
import { ArrowLeft, ArrowRight, GripVertical } from 'lucide-vue-next'

defineProps<{
  title: string
  layoutLocked: boolean
  canMovePrev: boolean
  canMoveNext: boolean
}>()

const emit = defineEmits<{
  (e: 'move', delta: -1 | 1): void
}>()
</script>

<style scoped>
.card-title-row {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-muted);
}

.card-title h3 {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
  color: var(--text-strong);
}

.title-right-group {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.drag-ctl-group {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.drag-handle-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 999px;
  border: 1px dashed color-mix(in srgb, var(--accent-primary) 42%, var(--border-control));
  background: var(--accent-primary-soft);
  color: var(--accent-text);
  font-size: 11px;
  font-weight: 600;
  cursor: grab;
  user-select: none;
}

.drag-handle-pill:active {
  cursor: grabbing;
}

.drag-step-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 7px;
  border: 1px solid var(--border-control);
  background: var(--surface-btn);
  color: var(--text-normal);
  cursor: pointer;
}

.drag-step-btn:hover:not(:disabled) {
  border-color: var(--accent-primary);
  color: var(--accent-text);
}

.drag-step-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}
</style>

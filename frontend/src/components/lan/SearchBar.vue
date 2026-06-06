<template>
  <div class="search-container glass-card">
    <div class="search-icon">
      <Search :size="16" />
    </div>
    <input
      type="text"
      class="search-input"
      placeholder="输入备注或主机名搜索局域网设备..."
      :value="modelValue"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
    <div class="search-hint"><kbd>CTRL</kbd> + <kbd>K</kbd></div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { Search } from 'lucide-vue-next'

defineProps<{ modelValue: string }>()
defineEmits<{ (e: 'update:modelValue', val: string): void }>()

function handleGlobalKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault()
    ;(document.querySelector('.search-input') as HTMLElement)?.focus()
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleGlobalKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleGlobalKeydown)
})
</script>

<style scoped>
.search-container {
  display: flex;
  align-items: center;
  padding: 0 20px;
  height: 56px;
  border-radius: 20px;
  position: relative;
  overflow: hidden;
}

.search-container:focus-within {
  border-color: var(--system-blue);
  box-shadow:
    0 0 0 3px var(--system-blue-dim),
    var(--glass-shadow),
    var(--glass-inner-shadow);
  background: var(--glass-bg-hover);
}

.search-icon {
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 14px;
  transition: color 0.3s;
}

.search-container:focus-within .search-icon {
  color: var(--system-blue);
}

.search-input {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--text-primary);
  font-size: 15px;
  outline: none;
  font-family: inherit;
  font-weight: 500;
  letter-spacing: 0.02em;
}

.search-input::placeholder {
  color: var(--text-tertiary);
  font-weight: 400;
}

.search-hint {
  display: flex;
  align-items: center;
  gap: 4px;
  opacity: 0.8;
}

kbd {
  background: var(--control-bg);
  border: 1px solid var(--control-border);
  border-radius: 6px;
  padding: 4px 8px;
  font-size: 11px;
  color: var(--text-secondary);
  font-family: ui-monospace, SFMono-Regular, monospace;
  box-shadow: var(--glass-inner-shadow);
}

@media (max-width: 560px) {
  .search-container {
    min-height: 56px;
    border-radius: 18px;
    align-items: center;
    padding: 16px;
  }

  .search-hint {
    display: none;
  }
}
</style>

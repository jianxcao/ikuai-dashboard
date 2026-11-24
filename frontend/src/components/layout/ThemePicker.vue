<template>
  <div v-if="!compact" class="theme-pills-wrap" role="group" aria-label="切换控制台主题">
    <button
      v-for="item in themes"
      :key="item.id"
      type="button"
      class="theme-pill"
      :class="{ active: theme === item.id }"
      :aria-pressed="theme === item.id"
      :title="`${item.label} (${item.englishName}) — ${item.description}`"
      @click="setTheme(item.id)"
    >
      <span class="swatch-card" :style="{ background: item.previewBg }">
        <i :style="{ background: item.swatchPrimary }" />
        <i :style="{ background: item.swatchSecondary }" />
      </span>
      <span class="theme-name">{{ item.label }}</span>
    </button>
  </div>

  <div v-else class="theme-dropdown-wrap" ref="dropEl">
    <button
      type="button"
      class="theme-trigger"
      :aria-expanded="open"
      aria-label="切换主题皮肤"
      @click="open = !open"
    >
      <span class="swatch-card" :style="{ background: currentThemeMeta.previewBg }">
        <i :style="{ background: currentThemeMeta.swatchPrimary }" />
        <i :style="{ background: currentThemeMeta.swatchSecondary }" />
      </span>
      <span class="trigger-label">{{ currentThemeMeta.label }}</span>
      <ChevronDown :size="14" class="chevron" :class="{ rotated: open }" />
    </button>

    <transition name="pop">
      <div v-if="open" class="theme-popover v4-panel" role="dialog" aria-label="选择视觉主题">
        <div class="pop-head">
          <div>
            <strong>视觉主题系统</strong>
            <span>6 套独立美学（支持苹果液态玻璃与高对比工作室）</span>
          </div>
          <Sparkles :size="15" class="pop-ic" />
        </div>

        <div class="pop-list">
          <button
            v-for="item in themes"
            :key="item.id"
            type="button"
            class="pop-item"
            :class="{ active: theme === item.id }"
            @click="choose(item.id)"
          >
            <div class="mini-ui" :style="{ background: item.previewBg, borderRadius: item.cornerRadius }">
              <span
                class="mini-glass"
                :class="{ glass: item.isGlass, dark: item.isDark }"
              >
                <b :style="{ background: item.swatchPrimary }" />
                <em :style="{ background: item.swatchSecondary }" />
              </span>
            </div>

            <div class="pop-meta">
              <div class="pop-title-row">
                <strong>{{ item.label }}</strong>
                <span class="v4-badge" :class="item.isGlass ? 'v4-badge-accent' : ''">{{ item.tag }}</span>
              </div>
              <span class="pop-desc">{{ item.description }}</span>
            </div>

            <Check v-if="theme === item.id" :size="15" class="active-check" />
          </button>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { Check, ChevronDown, Sparkles } from 'lucide-vue-next'
import { useTheme, type AppTheme } from '@/composables/useTheme'

withDefaults(
  defineProps<{
    compact?: boolean
  }>(),
  { compact: false }
)

const { theme, currentThemeMeta, themes, setTheme } = useTheme()
const open = ref(false)
const dropEl = ref<HTMLElement | null>(null)

function choose(id: AppTheme) {
  setTheme(id)
  open.value = false
}

function onDocClick(e: MouseEvent) {
  if (open.value && dropEl.value && !dropEl.value.contains(e.target as Node)) {
    open.value = false
  }
}

onMounted(() => document.addEventListener('click', onDocClick))
onUnmounted(() => document.removeEventListener('click', onDocClick))
</script>

<style scoped>
.theme-pills-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.theme-pill {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  min-height: 32px;
  padding: 4px 10px 4px 6px;
  border-radius: var(--radius-control);
  border: 1px solid var(--border-control);
  background: var(--surface-btn);
  color: var(--text-normal);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.16s ease;
}

.theme-pill:hover {
  color: var(--text-strong);
  border-color: var(--border-strong);
  background: var(--surface-hover);
}

.theme-pill.active {
  color: var(--text-strong);
  border-color: var(--accent-primary);
  background: var(--accent-primary-soft);
}

.swatch-card {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 4px 6px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.55);
  box-shadow: inset 0 1px 1px rgba(255, 255, 255, 0.7);
}

.swatch-card i {
  display: block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
}

.theme-dropdown-wrap {
  position: relative;
}

.theme-trigger {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 38px;
  padding: 0 11px 0 8px;
  border-radius: var(--radius-control);
  border: 1px solid var(--border-control);
  background: var(--surface-btn);
  color: var(--text-strong);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  box-shadow: var(--inner-sheen);
  transition: all 0.16s ease;
}

.theme-trigger:hover {
  border-color: var(--border-strong);
  background: var(--surface-hover);
}

.trigger-label {
  white-space: nowrap;
}

.chevron {
  color: var(--text-muted);
  transition: transform 0.18s ease;
}

.chevron.rotated {
  transform: rotate(180deg);
}

.theme-popover {
  position: absolute;
  top: calc(100% + 10px);
  right: 0;
  width: min(410px, calc(100vw - 24px));
  padding: 14px;
  background: var(--surface-popover);
  border: 1px solid var(--border-control);
  display: flex;
  flex-direction: column;
  gap: 10px;
  z-index: 95;
}

.pop-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--border-subtle);
}

.pop-head > div > strong {
  display: block;
  font-size: 13px;
  color: var(--text-strong);
}

.pop-head > div > span {
  display: block;
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 2px;
}

.pop-ic {
  color: var(--accent-text);
}

.pop-list {
  display: flex;
  flex-direction: column;
  gap: 7px;
  max-height: 64vh;
  overflow-y: auto;
}

.pop-item {
  display: flex;
  align-items: center;
  gap: 11px;
  width: 100%;
  padding: 8px 10px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
  background: var(--surface-input);
  text-align: left;
  cursor: pointer;
  transition: all 0.16s ease;
}

.pop-item:hover {
  border-color: var(--border-strong);
  background: var(--surface-hover);
}

.pop-item.active {
  border-color: var(--accent-primary);
  background: var(--accent-primary-soft);
}

.mini-ui {
  width: 48px;
  height: 36px;
  padding: 5px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border: 1px solid rgba(148, 163, 184, 0.35);
}

.mini-glass {
  width: 100%;
  height: 100%;
  border-radius: 6px;
  padding: 3px;
  background: rgba(255, 255, 255, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 3px;
}

.mini-glass.dark {
  background: rgba(18, 22, 34, 0.85);
}

.mini-glass.glass {
  background: rgba(255, 255, 255, 0.45);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(255, 255, 255, 0.72);
}

.mini-glass.glass.dark {
  background: rgba(255, 255, 255, 0.16);
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.mini-glass b,
.mini-glass em {
  display: block;
  height: 7px;
  border-radius: 999px;
}

.mini-glass b {
  width: 14px;
}

.mini-glass em {
  width: 8px;
}

.pop-meta {
  flex: 1;
  min-width: 0;
}

.pop-title-row {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.pop-title-row strong {
  font-size: 12.5px;
  color: var(--text-strong);
}

.pop-desc {
  display: block;
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 2px;
  line-height: 1.35;
}

.active-check {
  color: var(--accent-primary);
  flex-shrink: 0;
}

.pop-enter-active,
.pop-leave-active {
  transition: opacity 0.16s ease, transform 0.16s ease;
}

.pop-enter-from,
.pop-leave-to {
  opacity: 0;
  transform: translateY(-6px) scale(0.98);
}

@media (max-width: 640px) {
  .theme-pills-wrap {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 6px;
  }

  .theme-pill {
    min-width: 0;
    width: 100%;
    justify-content: flex-start;
  }

  .theme-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}
</style>

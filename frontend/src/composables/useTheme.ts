import { computed, readonly, ref, watchEffect } from 'vue'

export type AppTheme = 'liquid-dark' | 'liquid-light'

export const themes: Array<{ id: AppTheme; label: string; shortLabel: string }> = [
  { id: 'liquid-dark', label: 'Liquid Dark', shortLabel: 'Dark' },
  { id: 'liquid-light', label: 'Liquid Light', shortLabel: 'Light' },
]

const STORAGE_KEY = 'ikuai_theme'
const fallbackTheme: AppTheme = 'liquid-dark'
const theme = ref<AppTheme>(readInitialTheme())

function isTheme(value: string | null): value is AppTheme {
  return value === 'liquid-dark' || value === 'liquid-light'
}

function readInitialTheme(): AppTheme {
  if (typeof window === 'undefined') return fallbackTheme
  const saved = window.localStorage.getItem(STORAGE_KEY)
  return isTheme(saved) ? saved : fallbackTheme
}

function applyTheme(value: AppTheme) {
  if (typeof document === 'undefined') return
  document.documentElement.dataset.theme = value
}

function setTheme(value: AppTheme) {
  theme.value = value
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(STORAGE_KEY, value)
  }
}

function toggleTheme() {
  setTheme(theme.value === 'liquid-dark' ? 'liquid-light' : 'liquid-dark')
}

watchEffect(() => applyTheme(theme.value))

export function useTheme() {
  return {
    theme: readonly(theme),
    themeLabel: computed(() => themes.find((item) => item.id === theme.value)?.label ?? 'Liquid Dark'),
    themes,
    setTheme,
    toggleTheme,
  }
}

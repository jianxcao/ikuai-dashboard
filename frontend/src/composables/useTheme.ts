import { computed, readonly, ref, watchEffect } from 'vue'

export type AppTheme =
  | 'apple-liquid-glass'
  | 'vision-glass-dark'
  | 'cupertino-studio'
  | 'nordic-paper'
  | 'linear-graphite'
  | 'te-industrial'

export interface ThemeMeta {
  id: AppTheme
  label: string
  englishName: string
  tag: string
  description: string
  isDark: boolean
  isGlass: boolean
  bgColor: string
  previewBg: string
  swatchPrimary: string
  swatchSecondary: string
  cornerRadius: string
}

export const themes: ThemeMeta[] = [
  {
    id: 'apple-liquid-glass',
    label: '液态玻璃 · 浅色',
    englishName: 'Apple Liquid Glass',
    tag: 'iOS 光学玻璃',
    description: '银雾大气景深底图 + 光学折射胶囊导航 + 双层高光玻璃边框',
    isDark: false,
    isGlass: true,
    bgColor: '#e8eef6',
    previewBg: 'linear-gradient(155deg, #dce6f4 0%, #eff4fb 55%, #e3ecf7 100%)',
    swatchPrimary: '#0066dc',
    swatchSecondary: '#047857',
    cornerRadius: '18px',
  },
  {
    id: 'vision-glass-dark',
    label: '液态玻璃 · 深空',
    englishName: 'Liquid Glass Dark',
    tag: '暗色光学玻璃',
    description: '深海岩灰空间底景搭配烟熏透射玻璃，悬浮棱光护眼通透',
    isDark: true,
    isGlass: true,
    bgColor: '#090d14',
    previewBg: 'linear-gradient(155deg, #0b1322 0%, #112036 55%, #0c1724 100%)',
    swatchPrimary: '#2990ff',
    swatchSecondary: '#34d399',
    cornerRadius: '18px',
  },
  {
    id: 'cupertino-studio',
    label: '纯净云白 · 工作室',
    englishName: 'Studio Light',
    tag: '平面高对比',
    description: '中性纯白零模糊平面表层 + 1px 细腻发丝线，极致清晰效率',
    isDark: false,
    isGlass: false,
    bgColor: '#f4f5f7',
    previewBg: 'linear-gradient(155deg, #f4f5f7 0%, #eceff4 100%)',
    swatchPrimary: '#0062cc',
    swatchSecondary: '#047857',
    cornerRadius: '12px',
  },
  {
    id: 'linear-graphite',
    label: '曜石黑 · 极简',
    englishName: 'Graphite Dark',
    tag: '中性暗黑',
    description: '纯净低眩光石墨深色层阶，克制中性色与单一精准蓝主控',
    isDark: true,
    isGlass: false,
    bgColor: '#0a0b0d',
    previewBg: 'linear-gradient(155deg, #101216 0%, #191d24 100%)',
    swatchPrimary: '#3b82f6',
    swatchSecondary: '#10b981',
    cornerRadius: '10px',
  },
  {
    id: 'nordic-paper',
    label: '暖石纸感 · 护眼',
    englishName: 'Warm Stone Paper',
    tag: '温润阅读',
    description: '哑光石灰纸基底色搭配深松石绿强调色，长时间运维不刺眼',
    isDark: false,
    isGlass: false,
    bgColor: '#f2efe9',
    previewBg: 'linear-gradient(155deg, #f2efe9 0%, #e7e1d6 100%)',
    swatchPrimary: '#1f5c4b',
    swatchSecondary: '#0f766e',
    cornerRadius: '10px',
  },
  {
    id: 'te-industrial',
    label: '精密暗架 · 机房',
    englishName: 'Rack Console',
    tag: '工控机架',
    description: '哑光阳极炭黑机身与克制信号橙指示，干练硬朗机房控制台',
    isDark: true,
    isGlass: false,
    bgColor: '#111215',
    previewBg: 'linear-gradient(155deg, #16181c 0%, #22252c 100%)',
    swatchPrimary: '#b93809',
    swatchSecondary: '#4ade80',
    cornerRadius: '7px',
  },
]

const STORAGE_KEY = 'ikuai_v4_theme_v2'
const fallbackTheme: AppTheme = 'apple-liquid-glass'
const theme = ref<AppTheme>(readInitialTheme())

function isTheme(value: string | null): value is AppTheme {
  return themes.some((item) => item.id === value)
}

function readInitialTheme(): AppTheme {
  if (typeof window === 'undefined') return fallbackTheme
  const saved = window.localStorage.getItem(STORAGE_KEY)
  return isTheme(saved) ? saved : fallbackTheme
}

function applyTheme(value: AppTheme) {
  if (typeof document === 'undefined') return
  document.documentElement.dataset.theme = value
  const meta = themes.find((item) => item.id === value)
  if (meta) {
    const themeMetaEl = document.querySelector('meta[name="theme-color"]')
    if (themeMetaEl) {
      themeMetaEl.setAttribute('content', meta.bgColor)
    }
  }
}

function setTheme(value: AppTheme) {
  theme.value = value
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(STORAGE_KEY, value)
  }
}

function cycleTheme() {
  const idx = themes.findIndex((item) => item.id === theme.value)
  const next = themes[(idx + 1) % themes.length]
  if (next) setTheme(next.id)
}

watchEffect(() => applyTheme(theme.value))

export function useTheme() {
  return {
    theme: readonly(theme),
    currentThemeMeta: computed(() => themes.find((t) => t.id === theme.value) ?? themes[0]!),
    themes,
    setTheme,
    cycleTheme,
  }
}

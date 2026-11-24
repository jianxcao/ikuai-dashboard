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
    label: '液态玻璃 · 昼光',
    englishName: 'Apple Liquid Glass',
    tag: '推荐 · 苹果光学玻璃',
    description: '高透射双层高光光学玻璃 + 灵动胶囊控件 + 柔和多彩流体底图',
    isDark: false,
    isGlass: true,
    bgColor: '#e8f1fd',
    previewBg: 'linear-gradient(135deg, #c2e2ff 0%, #e4d9ff 50%, #ffd8d0 100%)',
    swatchPrimary: '#007aff',
    swatchSecondary: '#34c759',
    cornerRadius: '20px',
  },
  {
    id: 'vision-glass-dark',
    label: '液态暗晶 · 星夜',
    englishName: 'visionOS Spatial Glass',
    tag: '暗色光学玻璃',
    description: 'visionOS 深空烟熏光学玻璃，星云极光折射配合立体边框高光',
    isDark: true,
    isGlass: true,
    bgColor: '#080d1a',
    previewBg: 'linear-gradient(135deg, #162a5c 0%, #3b1c69 52%, #0c3b4a 100%)',
    swatchPrimary: '#0a84ff',
    swatchSecondary: '#30d158',
    cornerRadius: '20px',
  },
  {
    id: 'cupertino-studio',
    label: '极简云白 · 工作室',
    englishName: 'Cupertino Studio',
    tag: '纯净高对比日间',
    description: '纯白零模糊平面卡片 + 1px 发丝灰边框 + 钴蓝高对比按钮',
    isDark: false,
    isGlass: false,
    bgColor: '#f6f8fb',
    previewBg: 'linear-gradient(135deg, #f8fafc 0%, #eff6ff 100%)',
    swatchPrimary: '#2563eb',
    swatchSecondary: '#0d9488',
    cornerRadius: '12px',
  },
  {
    id: 'nordic-paper',
    label: '暖砂纸感 · 护眼',
    englishName: 'Nordic Warm Paper',
    tag: '温润编辑风',
    description: '亚麻米白底色搭配赤陶橙与鼠尾草绿，低眩光久看舒适',
    isDark: false,
    isGlass: false,
    bgColor: '#f4f0e8',
    previewBg: 'linear-gradient(135deg, #f5f0e6 0%, #ebe3d3 100%)',
    swatchPrimary: '#c85a32',
    swatchSecondary: '#2f6f5e',
    cornerRadius: '10px',
  },
  {
    id: 'linear-graphite',
    label: '钛黑星轨 · 极客',
    englishName: 'Linear Precision',
    tag: '经典极客暗黑',
    description: '中性纯黑微光表层 + 顶部 1px 棱光，精密开发者工具质感',
    isDark: true,
    isGlass: false,
    bgColor: '#09090b',
    previewBg: 'linear-gradient(135deg, #141418 0%, #1c1d2b 100%)',
    swatchPrimary: '#5e6ad2',
    swatchSecondary: '#38bdf8',
    cornerRadius: '10px',
  },
  {
    id: 'te-industrial',
    label: '机械橙工控 · 硬件',
    englishName: 'Field Rack Hardware',
    tag: '硬核硬件风',
    description: '阳极炭黑机身 + 硬朗机械倒角 + 国际救援橙实体按键',
    isDark: true,
    isGlass: false,
    bgColor: '#17171a',
    previewBg: 'linear-gradient(135deg, #1f1f24 0%, #2a221d 100%)',
    swatchPrimary: '#ff5500',
    swatchSecondary: '#ffb800',
    cornerRadius: '6px',
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

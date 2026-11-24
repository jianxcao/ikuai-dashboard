import { computed } from 'vue'
import { useTheme } from './useTheme'

export interface ChartPalette {
  isDark: boolean
  textStrong: string
  textMuted: string
  gridLine: string
  tooltipBg: string
  tooltipBorder: string
  upColor: string
  downColor: string
  accent: string
  warn: string
  danger: string
  seriesPalette: string[]
}

const themePalettes: Record<string, ChartPalette> = {
  'apple-liquid-glass': {
    isDark: false,
    textStrong: '#0b1526',
    textMuted: '#4c5f7d',
    gridLine: 'rgba(15, 23, 42, 0.08)',
    tooltipBg: 'rgba(255, 255, 255, 0.90)',
    tooltipBorder: 'rgba(0, 122, 255, 0.25)',
    upColor: '#10b981',
    downColor: '#007aff',
    accent: '#5856d6',
    warn: '#f59e0b',
    danger: '#ff3b30',
    seriesPalette: ['#007aff', '#10b981', '#5856d6', '#ff9500', '#ff3b30', '#06b6d4'],
  },
  'vision-glass-dark': {
    isDark: true,
    textStrong: '#f5f8ff',
    textMuted: '#9db4d8',
    gridLine: 'rgba(255, 255, 255, 0.09)',
    tooltipBg: 'rgba(14, 21, 40, 0.92)',
    tooltipBorder: 'rgba(255, 255, 255, 0.22)',
    upColor: '#30d158',
    downColor: '#0a84ff',
    accent: '#bf5af2',
    warn: '#ffd60a',
    danger: '#ff453a',
    seriesPalette: ['#0a84ff', '#30d158', '#bf5af2', '#ffd60a', '#64d2ff', '#ff453a'],
  },
  'cupertino-studio': {
    isDark: false,
    textStrong: '#0f172a',
    textMuted: '#64748b',
    gridLine: 'rgba(15, 23, 42, 0.07)',
    tooltipBg: '#ffffff',
    tooltipBorder: '#cbd5e1',
    upColor: '#059669',
    downColor: '#2563eb',
    accent: '#4f46e5',
    warn: '#d97706',
    danger: '#dc2626',
    seriesPalette: ['#2563eb', '#059669', '#4f46e5', '#d97706', '#dc2626', '#0d9488'],
  },
  'nordic-paper': {
    isDark: false,
    textStrong: '#241e18',
    textMuted: '#6e6559',
    gridLine: 'rgba(62, 48, 34, 0.08)',
    tooltipBg: '#fffefa',
    tooltipBorder: '#ded5c2',
    upColor: '#2f6f5e',
    downColor: '#c85a32',
    accent: '#b45309',
    warn: '#d97706',
    danger: '#b91c1c',
    seriesPalette: ['#c85a32', '#2f6f5e', '#b45309', '#3b82f6', '#7c3aed', '#b91c1c'],
  },
  'linear-graphite': {
    isDark: true,
    textStrong: '#f4f4f6',
    textMuted: '#9898a6',
    gridLine: 'rgba(255, 255, 255, 0.07)',
    tooltipBg: '#16161b',
    tooltipBorder: 'rgba(255, 255, 255, 0.14)',
    upColor: '#34d399',
    downColor: '#5e6ad2',
    accent: '#38bdf8',
    warn: '#fbbf24',
    danger: '#f87171',
    seriesPalette: ['#5e6ad2', '#38bdf8', '#34d399', '#fbbf24', '#a78bfa', '#f87171'],
  },
  'te-industrial': {
    isDark: true,
    textStrong: '#f5f5f0',
    textMuted: '#a1a1aa',
    gridLine: 'rgba(255, 255, 255, 0.08)',
    tooltipBg: '#232328',
    tooltipBorder: '#3e3e48',
    upColor: '#22c55e',
    downColor: '#ff5500',
    accent: '#ffb800',
    warn: '#ffb800',
    danger: '#ef4444',
    seriesPalette: ['#ff5500', '#22c55e', '#ffb800', '#38bdf8', '#e4e4e7', '#ef4444'],
  },
}

export function useChartTheme() {
  const { theme } = useTheme()
  const palette = computed<ChartPalette>(
    () => themePalettes[theme.value] ?? themePalettes['apple-liquid-glass']!
  )
  return { palette }
}

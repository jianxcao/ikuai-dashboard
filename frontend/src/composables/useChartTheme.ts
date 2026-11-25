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
    textStrong: '#0b1320',
    textMuted: '#475569',
    gridLine: 'rgba(15, 23, 42, 0.07)',
    tooltipBg: 'rgba(255, 255, 255, 0.92)',
    tooltipBorder: 'rgba(15, 23, 42, 0.12)',
    upColor: '#059669',
    downColor: '#0066dc',
    accent: '#0284c7',
    warn: '#d97706',
    danger: '#be123c',
    seriesPalette: ['#0066dc', '#059669', '#0284c7', '#d97706', '#64748b', '#0d9488'],
  },
  'vision-glass-dark': {
    isDark: true,
    textStrong: '#f6f9ff',
    textMuted: '#9eb0cc',
    gridLine: 'rgba(255, 255, 255, 0.08)',
    tooltipBg: 'rgba(18, 25, 39, 0.94)',
    tooltipBorder: 'rgba(255, 255, 255, 0.18)',
    upColor: '#34d399',
    downColor: '#3b9eff',
    accent: '#38bdf8',
    warn: '#fbbf24',
    danger: '#fb7185',
    seriesPalette: ['#3b9eff', '#34d399', '#38bdf8', '#fbbf24', '#94a3b8', '#2dd4bf'],
  },
  'cupertino-studio': {
    isDark: false,
    textStrong: '#0f172a',
    textMuted: '#475569',
    gridLine: 'rgba(15, 23, 42, 0.06)',
    tooltipBg: '#ffffff',
    tooltipBorder: '#d8dde6',
    upColor: '#047857',
    downColor: '#0062cc',
    accent: '#0284c7',
    warn: '#d97706',
    danger: '#b91c1c',
    seriesPalette: ['#0062cc', '#047857', '#0284c7', '#d97706', '#64748b', '#0d9488'],
  },
  'linear-graphite': {
    isDark: true,
    textStrong: '#f4f6fa',
    textMuted: '#9aa3b2',
    gridLine: 'rgba(255, 255, 255, 0.06)',
    tooltipBg: '#161920',
    tooltipBorder: 'rgba(255, 255, 255, 0.14)',
    upColor: '#34d399',
    downColor: '#3b82f6',
    accent: '#38bdf8',
    warn: '#fbbf24',
    danger: '#fb7185',
    seriesPalette: ['#3b82f6', '#34d399', '#38bdf8', '#fbbf24', '#94a3b8', '#2dd4bf'],
  },
  'nordic-paper': {
    isDark: false,
    textStrong: '#1c1917',
    textMuted: '#575047',
    gridLine: 'rgba(46, 41, 36, 0.07)',
    tooltipBg: '#fbf9f5',
    tooltipBorder: '#d6cfc2',
    upColor: '#1f5c4b',
    downColor: '#0284c7',
    accent: '#0f766e',
    warn: '#b45309',
    danger: '#b91c1c',
    seriesPalette: ['#1f5c4b', '#0284c7', '#0f766e', '#b45309', '#575047', '#047857'],
  },
  'te-industrial': {
    isDark: true,
    textStrong: '#fafaf9',
    textMuted: '#a1a4b0',
    gridLine: 'rgba(255, 255, 255, 0.06)',
    tooltipBg: '#1c1e24',
    tooltipBorder: '#3e424f',
    upColor: '#4ade80',
    downColor: '#fb923c',
    accent: '#38bdf8',
    warn: '#fbbf24',
    danger: '#f87185',
    seriesPalette: ['#fb923c', '#4ade80', '#38bdf8', '#fbbf24', '#a1a4b0', '#2dd4bf'],
  },
}

export function useChartTheme() {
  const { theme } = useTheme()
  const palette = computed<ChartPalette>(
    () => themePalettes[theme.value] ?? themePalettes['apple-liquid-glass']!
  )
  return { palette }
}

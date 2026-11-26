import { computed, ref, type Ref } from 'vue'
import { formatCompoundValue } from './useFormatters'

export type SortDirection = 'asc' | 'desc'
export type SortValueHint = 'auto' | 'ip' | 'number' | 'boolean' | 'duration' | 'compound' | 'text'

const localeCollator = new Intl.Collator('zh-CN', {
  numeric: true,
  sensitivity: 'base',
})

const UNIT_MULTIPLIERS: Record<string, number> = {
  b: 1,
  bytes: 1,
  kb: 1024,
  kib: 1024,
  mb: 1024 ** 2,
  mib: 1024 ** 2,
  gb: 1024 ** 3,
  gib: 1024 ** 3,
  tb: 1024 ** 4,
  tib: 1024 ** 4,
  'b/s': 1,
  bps: 1,
  'kb/s': 1024,
  kbps: 1024,
  'mb/s': 1024 ** 2,
  mbps: 1024 ** 2,
  'gb/s': 1024 ** 3,
  gbps: 1024 ** 3,
  ms: 1,
  s: 1,
  '%': 1,
}

export function isSortEmpty(val: unknown): boolean {
  if (val === null || val === undefined) return true
  if (typeof val === 'number') return Number.isNaN(val)
  if (typeof val === 'string') {
    const t = val.trim()
    return t === '' || t === '—' || t === '--'
  }
  return false
}

function isTruthyVal(val: unknown): boolean {
  if (typeof val === 'boolean') return val
  if (typeof val === 'number') return val !== 0
  const s = String(val ?? '').trim().toLowerCase()
  return s === 'yes' || s === '1' || s === 'true' || s === 'on'
}

export function parseIPv4SortKey(val: unknown): { num: number; suffix: string } | null {
  if (typeof val !== 'string') return null
  const s = val.trim()
  const m = /^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})(.*)$/.exec(s)
  if (!m) return null
  const o0 = Number(m[1])
  const o1 = Number(m[2])
  const o2 = Number(m[3])
  const o3 = Number(m[4])
  if (o0 > 255 || o1 > 255 || o2 > 255 || o3 > 255) return null
  const num = ((o0 * 256 + o1) * 256 + o2) * 256 + o3
  return { num, suffix: (m[5] ?? '').trim() }
}

const DURATION_FULL_RE =
  /^(?:\d+(?:\.\d+)?\s*(?:d|days?|天|h|hrs?|hours?|时|小时|m|mins?|minutes?|分|分钟|s|secs?|seconds?|秒)[\s,]*)+(?:ago|前)?$/i

export function parseDurationSeconds(val: unknown): number | null {
  if (typeof val === 'number' && !Number.isNaN(val)) return val
  if (typeof val !== 'string') return null
  const s = val.trim().toLowerCase()
  if (!s) return null

  const hms = /^(\d+):(\d{1,2}):(\d{1,2})$/.exec(s)
  if (hms) {
    return Number(hms[1]) * 3600 + Number(hms[2]) * 60 + Number(hms[3])
  }

  if (!DURATION_FULL_RE.test(s)) return null

  let matched = false
  let total = 0
  const tokenRe = /(\d+(?:\.\d+)?)\s*(d|days?|天|h|hrs?|hours?|时|小时|m|mins?|minutes?|分|分钟|s|secs?|seconds?|秒)/g
  let match: RegExpExecArray | null
  while ((match = tokenRe.exec(s)) !== null) {
    matched = true
    const num = Number(match[1])
    const unit = match[2] ?? ''
    if (unit === 'd' || unit.startsWith('day') || unit === '天') total += num * 86400
    else if (unit.startsWith('h') || unit === '时' || unit === '小时') total += num * 3600
    else if (unit.startsWith('m') || unit === '分' || unit === '分钟') total += num * 60
    else total += num
  }
  return matched ? total : null
}

function parseNumericSortKey(val: unknown): { primary: number; secondary: number } | null {
  if (typeof val === 'number') {
    return Number.isNaN(val) ? null : { primary: val, secondary: 0 }
  }
  if (typeof val !== 'string') return null
  const s = val.trim()
  if (!s) return null

  if (/^-?\d+(?:\.\d+)?$/.test(s)) {
    const n = Number(s)
    return Number.isNaN(n) ? null : { primary: n, secondary: 0 }
  }

  const rangeMatch = /^(\d+)\s*-\s*(\d+)$/.exec(s)
  if (rangeMatch) {
    return { primary: Number(rangeMatch[1]), secondary: Number(rangeMatch[2]) }
  }

  const unitMatch = /^(-?\d+(?:\.\d+)?)\s*([a-zA-Z/%]+)$/.exec(s)
  if (unitMatch) {
    const unit = (unitMatch[2] ?? '').toLowerCase()
    const mul = UNIT_MULTIPLIERS[unit]
    if (mul !== undefined) {
      return { primary: Number(unitMatch[1]) * mul, secondary: 0 }
    }
  }
  return null
}

export function compareSortValues(
  a: unknown,
  b: unknown,
  dir: SortDirection,
  hint: SortValueHint = 'auto'
): number {
  const aEmpty = isSortEmpty(a)
  const bEmpty = isSortEmpty(b)
  if (aEmpty && bEmpty) return 0
  if (aEmpty) return 1
  if (bEmpty) return -1

  const factor = dir === 'asc' ? 1 : -1

  if (hint === 'boolean') {
    const av = isTruthyVal(a) ? 1 : 0
    const bv = isTruthyVal(b) ? 1 : 0
    if (av !== bv) return (av - bv) * factor
    return localeCollator.compare(String(a), String(b)) * factor
  }

  const aVal = hint === 'compound' || (a && typeof a === 'object') ? formatCompoundValue(a) : a
  const bVal = hint === 'compound' || (b && typeof b === 'object') ? formatCompoundValue(b) : b

  if (hint === 'ip' || hint === 'auto') {
    const ipA = parseIPv4SortKey(aVal)
    const ipB = parseIPv4SortKey(bVal)
    if (ipA && ipB) {
      if (ipA.num !== ipB.num) return (ipA.num - ipB.num) * factor
      return localeCollator.compare(ipA.suffix, ipB.suffix) * factor
    }
  }

  if (hint === 'duration' || hint === 'auto') {
    const durA = parseDurationSeconds(aVal)
    const durB = parseDurationSeconds(bVal)
    if (durA !== null && durB !== null && durA !== durB) {
      return (durA - durB) * factor
    }
  }

  if (hint === 'number' || hint === 'auto') {
    const numA = parseNumericSortKey(aVal)
    const numB = parseNumericSortKey(bVal)
    if (numA && numB) {
      if (numA.primary !== numB.primary) return (numA.primary - numB.primary) * factor
      if (numA.secondary !== numB.secondary) return (numA.secondary - numB.secondary) * factor
    }
  }

  return localeCollator.compare(String(aVal), String(bVal)) * factor
}

export interface UseTableSortOptions<K extends string = string> {
  initialKey?: K | null
  initialDir?: SortDirection
  allowUnsorted?: boolean
  defaultDirForKey?: (key: K) => SortDirection
}

export function useTableSort<K extends string = string>(options?: UseTableSortOptions<K>) {
  const allowUnsorted = options?.allowUnsorted ?? true
  const sortKey = ref(options?.initialKey ?? null) as Ref<K | null>
  const sortDir = ref<SortDirection>(options?.initialDir ?? 'asc')

  const isSorted = computed(() => sortKey.value !== null)

  function toggleSort(key: K) {
    if (sortKey.value !== key) {
      sortKey.value = key
      sortDir.value = options?.defaultDirForKey ? options.defaultDirForKey(key) : 'asc'
      return
    }
    const firstDir = options?.defaultDirForKey ? options.defaultDirForKey(key) : 'asc'
    const secondDir: SortDirection = firstDir === 'asc' ? 'desc' : 'asc'
    if (sortDir.value === firstDir) {
      sortDir.value = secondDir
    } else if (allowUnsorted) {
      sortKey.value = null
      sortDir.value = 'asc'
    } else {
      sortDir.value = firstDir
    }
  }

  function setSort(key: K | null, dir: SortDirection = 'asc') {
    sortKey.value = key
    sortDir.value = dir
  }

  function resetSort() {
    sortKey.value = options?.initialKey ?? null
    sortDir.value = options?.initialDir ?? 'asc'
  }

  function ariaSort(key: K): 'ascending' | 'descending' | 'none' {
    if (sortKey.value !== key) return 'none'
    return sortDir.value === 'asc' ? 'ascending' : 'descending'
  }

  function sortRows<T>(
    list: readonly T[],
    getValue: (item: T, key: K) => unknown,
    getHint?: (key: K) => SortValueHint
  ): T[] {
    const key = sortKey.value
    if (!key) return [...list]
    const dir = sortDir.value
    const hint = getHint ? getHint(key) : 'auto'
    return list
      .map((item, idx) => ({ item, idx, value: getValue(item, key) }))
      .sort((a, b) => {
        const cmp = compareSortValues(a.value, b.value, dir, hint)
        return cmp !== 0 ? cmp : a.idx - b.idx
      })
      .map((entry) => entry.item)
  }

  return {
    sortKey,
    sortDir,
    isSorted,
    toggleSort,
    setSort,
    resetSort,
    ariaSort,
    sortRows,
  }
}

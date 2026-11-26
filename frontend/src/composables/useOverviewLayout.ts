import { computed, ref } from 'vue'

export type KpiCardId = 'upload' | 'download' | 'connections' | 'users'
export type SectionCardId =
  | 'cpu-mem'
  | 'physical-ports'
  | 'wan-chart'
  | 'load-chart'
  | 'wan-checks'
  | 'iface-traffic'

export const DEFAULT_KPI_ORDER: readonly KpiCardId[] = ['upload', 'download', 'connections', 'users']
export const DEFAULT_SECTION_ORDER: readonly SectionCardId[] = [
  'cpu-mem',
  'physical-ports',
  'wan-chart',
  'load-chart',
  'wan-checks',
  'iface-traffic',
]

const STORAGE_KEY = 'ikuai_v4_overview_layout_v1'

interface StoredLayout {
  kpiOrder?: unknown
  sectionOrder?: unknown
}

function normalizeOrder<T extends string>(raw: unknown, defaults: readonly T[]): T[] {
  const valid = new Set<T>(defaults)
  const seen = new Set<T>()
  const out: T[] = []
  if (Array.isArray(raw)) {
    for (const item of raw) {
      if (typeof item === 'string' && valid.has(item as T) && !seen.has(item as T)) {
        seen.add(item as T)
        out.push(item as T)
      }
    }
  }
  for (const def of defaults) {
    if (!seen.has(def)) {
      seen.add(def)
      out.push(def)
    }
  }
  return out
}

function loadStoredLayout(): { kpi: KpiCardId[]; section: SectionCardId[] } {
  if (typeof window === 'undefined') {
    return {
      kpi: [...DEFAULT_KPI_ORDER],
      section: [...DEFAULT_SECTION_ORDER],
    }
  }
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) {
      return {
        kpi: [...DEFAULT_KPI_ORDER],
        section: [...DEFAULT_SECTION_ORDER],
      }
    }
    const parsed = JSON.parse(raw) as StoredLayout
    return {
      kpi: normalizeOrder(parsed.kpiOrder, DEFAULT_KPI_ORDER),
      section: normalizeOrder(parsed.sectionOrder, DEFAULT_SECTION_ORDER),
    }
  } catch {
    return {
      kpi: [...DEFAULT_KPI_ORDER],
      section: [...DEFAULT_SECTION_ORDER],
    }
  }
}

const initial = loadStoredLayout()
const kpiOrder = ref<KpiCardId[]>(initial.kpi)
const sectionOrder = ref<SectionCardId[]>(initial.section)
const layoutLocked = ref(true)

function persistLayout() {
  if (typeof window === 'undefined') return
  try {
    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify({
        kpiOrder: kpiOrder.value,
        sectionOrder: sectionOrder.value,
      })
    )
  } catch {
    // Ignore storage quota errors
  }
}

function swapItems<T extends string>(list: T[], from: T, to: T): T[] {
  if (from === to) return list
  const fromIdx = list.indexOf(from)
  const toIdx = list.indexOf(to)
  if (fromIdx < 0 || toIdx < 0) return list
  const next = [...list]
  const fromVal = next[fromIdx]!
  const toVal = next[toIdx]!
  next[fromIdx] = toVal
  next[toIdx] = fromVal
  return next
}

function moveItem<T extends string>(list: T[], id: T, delta: -1 | 1): T[] {
  const idx = list.indexOf(id)
  const targetIdx = idx + delta
  if (idx < 0 || targetIdx < 0 || targetIdx >= list.length) return list
  const targetId = list[targetIdx]!
  return swapItems(list, id, targetId)
}

export function useOverviewLayout() {
  const isCustomLayout = computed(
    () =>
      kpiOrder.value.some((id, idx) => id !== DEFAULT_KPI_ORDER[idx]) ||
      sectionOrder.value.some((id, idx) => id !== DEFAULT_SECTION_ORDER[idx])
  )

  function swapKpi(from: KpiCardId, to: KpiCardId) {
    kpiOrder.value = swapItems(kpiOrder.value, from, to)
    persistLayout()
  }

  function swapSection(from: SectionCardId, to: SectionCardId) {
    sectionOrder.value = swapItems(sectionOrder.value, from, to)
    persistLayout()
  }

  function moveKpi(id: KpiCardId, delta: -1 | 1) {
    kpiOrder.value = moveItem(kpiOrder.value, id, delta)
    persistLayout()
  }

  function moveSection(id: SectionCardId, delta: -1 | 1) {
    sectionOrder.value = moveItem(sectionOrder.value, id, delta)
    persistLayout()
  }

  function unlockLayout() {
    layoutLocked.value = false
  }

  function lockLayout() {
    persistLayout()
    layoutLocked.value = true
  }

  function resetLayout() {
    kpiOrder.value = [...DEFAULT_KPI_ORDER]
    sectionOrder.value = [...DEFAULT_SECTION_ORDER]
    persistLayout()
  }

  return {
    kpiOrder,
    sectionOrder,
    layoutLocked,
    isCustomLayout,
    swapKpi,
    swapSection,
    moveKpi,
    moveSection,
    unlockLayout,
    lockLayout,
    resetLayout,
  }
}

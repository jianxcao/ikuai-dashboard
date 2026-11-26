import { readonly, ref } from 'vue'

export const ROUTE_TAB_TITLES: Record<string, string> = {
  overview: '总览驾驶舱',
  terminals: '终端与用户',
  'traffic-topology': '流量与拓扑',
  network: '网络与接口',
  routing: '路由与 QoS',
  security: '安全防护',
  objects: '对象分组',
  vpn: 'VPN 隧道',
  wireless: '无线网络',
  logs: '日志审计',
  system: '系统维护',
  fleet: '节点与主题',
}

const RESOURCE_GROUP_TAB_MAP: Record<string, string> = {
  network: 'network',
  routing: 'routing',
  qos: 'routing',
  security: 'security',
  users: 'terminals',
  vpn: 'vpn',
  wireless: 'wireless',
  objects: 'objects',
  advanced: 'system',
  log: 'logs',
  system: 'system',
}

export function resolveResourceTab(group: string): string {
  return RESOURCE_GROUP_TAB_MAP[group] || 'overview'
}

interface ParsedRoute {
  tab: string
  sub?: string
  normalized: boolean
}

function parsePathname(pathname: string): ParsedRoute {
  const parts = pathname
    .split('/')
    .map((s) => s.trim())
    .filter(Boolean)

  if (parts.length === 0) {
    return { tab: 'overview', sub: undefined, normalized: true }
  }

  const candidateTab = decodeSegment(parts[0]!)
  const candidateSub = parts[1] ? decodeSegment(parts[1]) : undefined

  if (candidateTab === 'vpn-wireless') {
    if (candidateSub?.startsWith('wireless-')) {
      return { tab: 'wireless', sub: candidateSub, normalized: true }
    }
    if (candidateSub?.endsWith('-objects')) {
      return { tab: 'objects', sub: candidateSub, normalized: true }
    }
    return { tab: 'vpn', sub: candidateSub || undefined, normalized: true }
  }

  if (!Object.prototype.hasOwnProperty.call(ROUTE_TAB_TITLES, candidateTab)) {
    return { tab: 'overview', sub: undefined, normalized: true }
  }

  return {
    tab: candidateTab,
    sub: candidateSub || undefined,
    normalized: parts.length > 2,
  }
}

function decodeSegment(raw: string): string {
  try {
    return decodeURIComponent(raw)
  } catch {
    return raw
  }
}

function buildPath(tab: string, sub?: string): string {
  if (!Object.prototype.hasOwnProperty.call(ROUTE_TAB_TITLES, tab)) {
    return '/overview'
  }
  if (sub) {
    return `/${tab}/${encodeURIComponent(sub)}`
  }
  return `/${tab}`
}

function updateDocumentTitle(tab: string) {
  if (typeof document === 'undefined') return
  const label = ROUTE_TAB_TITLES[tab] || '总览驾驶舱'
  document.title = `${label} · iKuai Console V4`
}

const initialParsed =
  typeof window !== 'undefined'
    ? parsePathname(window.location.pathname)
    : { tab: 'overview', sub: undefined, normalized: false }

const startedExplicitNonOverview =
  typeof window !== 'undefined' &&
  initialParsed.tab !== 'overview' &&
  !initialParsed.normalized

const currentTab = ref(initialParsed.tab)
const currentSub = ref<string | undefined>(initialParsed.sub)
const navSeq = ref(0)

if (typeof window !== 'undefined') {
  updateDocumentTitle(currentTab.value)
  if (initialParsed.normalized || window.location.pathname === '/') {
    const clean = buildPath(currentTab.value, currentSub.value)
    if (window.location.pathname !== clean) {
      window.history.replaceState(null, '', clean + window.location.search + window.location.hash)
    }
  }

  window.addEventListener('popstate', () => {
    const parsed = parsePathname(window.location.pathname)
    currentTab.value = parsed.tab
    currentSub.value = parsed.sub
    navSeq.value += 1
    updateDocumentTitle(parsed.tab)
  })
}

function navigate(tab: string, sub?: string, opts?: { replace?: boolean }) {
  const safeTab = Object.prototype.hasOwnProperty.call(ROUTE_TAB_TITLES, tab) ? tab : 'overview'
  const cleanSub = sub?.trim() || undefined
  currentTab.value = safeTab
  currentSub.value = cleanSub
  navSeq.value += 1
  updateDocumentTitle(safeTab)

  if (typeof window !== 'undefined') {
    const targetPath = buildPath(safeTab, cleanSub)
    if (window.location.pathname !== targetPath) {
      if (opts?.replace) {
        window.history.replaceState(null, '', targetPath)
      } else {
        window.history.pushState(null, '', targetPath)
      }
    }
  }
}

function syncSubTab(tab: string, sub: string, opts?: { replace?: boolean }) {
  if (currentTab.value !== tab) return
  const cleanSub = sub.trim()
  if (!cleanSub) return
  currentSub.value = cleanSub

  if (typeof window !== 'undefined') {
    const targetPath = buildPath(tab, cleanSub)
    if (window.location.pathname !== targetPath) {
      if (opts?.replace) {
        window.history.replaceState(null, '', targetPath)
      } else {
        window.history.pushState(null, '', targetPath)
      }
    }
  }
}

export function useAppRoute() {
  return {
    currentTab: readonly(currentTab),
    currentSub: readonly(currentSub),
    navSeq: readonly(navSeq),
    startedExplicitNonOverview,
    navigate,
    syncSubTab,
    buildPath,
  }
}

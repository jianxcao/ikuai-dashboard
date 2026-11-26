import { reactive, ref } from 'vue'
import { fetchV4Resource } from '@/api/v4'
import { useAppRoute } from '@/composables/useAppRoute'
import { useFleet } from '@/composables/useFleet'
import { formatObjectRef } from '@/composables/useFormatters'

export type ObjectGroupKey =
  | 'ip-objects'
  | 'ip6-objects'
  | 'mac-objects'
  | 'port-objects'
  | 'proto-objects'
  | 'domain-objects'
  | 'time-objects'

export const OBJECT_GROUP_META: Record<
  ObjectGroupKey,
  { label: string; valueKey?: string; placeholder?: string }
> = {
  'ip-objects': {
    label: 'IP 对象组',
    valueKey: 'ip',
    placeholder: '每行或逗号分隔多个 IP / CIDR，如 192.168.9.10, 10.0.0.0/24',
  },
  'ip6-objects': {
    label: 'IPv6 对象组',
    valueKey: 'ipv6',
    placeholder: '每行或逗号分隔 IPv6 地址或前缀，如 fd00::/64',
  },
  'mac-objects': {
    label: 'MAC 对象组',
    valueKey: 'mac',
    placeholder: '每行或逗号分隔 MAC 地址，如 AA:BB:CC:DD:EE:FF',
  },
  'port-objects': {
    label: '端口对象组',
    valueKey: 'port',
    placeholder: '每行或逗号分隔端口或端口段，如 80, 443, 8080-8090',
  },
  'proto-objects': {
    label: '协议对象组',
    valueKey: 'proto',
    placeholder: '每行或逗号分隔协议名，如 tcp, udp',
  },
  'domain-objects': {
    label: '域名对象组',
    valueKey: 'domain',
    placeholder: '每行或逗号分隔域名，如 example.com, api.example.com',
  },
  'time-objects': {
    label: '时间段对象组',
  },
}

const COMPOUND_FIELD_MAP: Record<string, ObjectGroupKey> = {
  src_addr: 'ip-objects',
  dst_addr: 'ip-objects',
  ip_addr: 'ip-objects',
  nol2rt_ip: 'ip-objects',
  src6_addr: 'ip6-objects',
  dst6_addr: 'ip6-objects',
  src_port: 'port-objects',
  dst_port: 'port-objects',
  domain: 'domain-objects',
  mac_addr: 'mac-objects',
  lmac: 'mac-objects',
  time: 'time-objects',
}

export function getFieldObjectGroup(resourceKey: string, fieldKey: string): ObjectGroupKey | null {
  if (resourceKey.endsWith('-objects')) return null
  if (resourceKey === 'domain-blacklist' && fieldKey === 'domain_group') {
    return 'domain-objects'
  }
  return COMPOUND_FIELD_MAP[fieldKey] ?? null
}

export function extractCompoundParts(val: unknown): { custom: string[]; objects: string[] } {
  if (val === null || val === undefined || val === '') {
    return { custom: [], objects: [] }
  }
  if (typeof val === 'string' || typeof val === 'number') {
    const items = String(val)
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
    return { custom: items, objects: [] }
  }
  if (Array.isArray(val)) {
    const custom = val
      .map((item) => {
        if (item && typeof item === 'object') {
          return Object.values(item as Record<string, unknown>).join(':')
        }
        return String(item)
      })
      .map((s) => s.trim())
      .filter(Boolean)
    return { custom, objects: [] }
  }
  if (typeof val === 'object') {
    const obj = val as Record<string, unknown>
    const custom = Array.isArray(obj.custom)
      ? obj.custom.map((x) => String(x ?? '').trim()).filter(Boolean)
      : []
    const objects = Array.isArray(obj.object)
      ? obj.object.map((x) => formatObjectRef(x)).filter(Boolean)
      : []
    return { custom, objects }
  }
  return { custom: [String(val)], objects: [] }
}

const groupCache = reactive<Partial<Record<ObjectGroupKey, Array<Record<string, unknown>>>>>({})
const groupLoading = reactive<Partial<Record<ObjectGroupKey, boolean>>>({})
const inFlight = new Map<ObjectGroupKey, Promise<Array<Record<string, unknown>>>>()
let cachedEpoch = -1

const pendingGroupFilter = ref<{ resourceKey: string; query: string } | null>(null)

function ensureCacheEpoch() {
  const { routerEpoch } = useFleet()
  if (cachedEpoch !== routerEpoch.value) {
    cachedEpoch = routerEpoch.value
    for (const key of Object.keys(groupCache) as ObjectGroupKey[]) {
      delete groupCache[key]
      delete groupLoading[key]
    }
    inFlight.clear()
  }
}

export async function loadObjectGroupList(
  groupKey: ObjectGroupKey,
  force = false
): Promise<Array<Record<string, unknown>>> {
  ensureCacheEpoch()
  if (!force && groupCache[groupKey]) {
    return groupCache[groupKey] || []
  }
  const existing = inFlight.get(groupKey)
  if (!force && existing) {
    return existing
  }
  groupLoading[groupKey] = true
  const task = fetchV4Resource(groupKey, { page: 1, page_size: 200 })
    .then((res) => {
      const list = res.rows || []
      groupCache[groupKey] = list
      return list
    })
    .catch(() => groupCache[groupKey] || [])
    .finally(() => {
      groupLoading[groupKey] = false
      inFlight.delete(groupKey)
    })
  inFlight.set(groupKey, task)
  return task
}

export function invalidateObjectGroupCache(groupKey?: ObjectGroupKey) {
  if (groupKey) {
    delete groupCache[groupKey]
    inFlight.delete(groupKey)
    return
  }
  for (const key of Object.keys(groupCache) as ObjectGroupKey[]) {
    delete groupCache[key]
  }
  inFlight.clear()
}

export function jumpToObjectGroup(groupKey: ObjectGroupKey, groupName?: string) {
  const cleanName = groupName?.trim() || ''
  pendingGroupFilter.value = cleanName ? { resourceKey: groupKey, query: cleanName } : null
  const { navigate } = useAppRoute()
  navigate('objects', groupKey)
}

export function consumePendingGroupFilter(resourceKey: string): string | null {
  const pending = pendingGroupFilter.value
  if (!pending || pending.resourceKey !== resourceKey) return null
  pendingGroupFilter.value = null
  return pending.query
}

export function useObjectGroups() {
  ensureCacheEpoch()
  return {
    groupCache,
    groupLoading,
    loadObjectGroupList,
    invalidateObjectGroupCache,
    jumpToObjectGroup,
    consumePendingGroupFilter,
  }
}

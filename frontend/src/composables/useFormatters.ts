/**
 * 数值、时间与 V4 复合结构格式化工具集
 */

export function formatSpeed(bps: number | null | undefined): string {
  if (bps == null || isNaN(bps)) return '0 B/s'
  const abs = Math.abs(bps)
  if (abs >= 1024 ** 3) return (bps / 1024 ** 3).toFixed(2) + ' GB/s'
  if (abs >= 1024 ** 2) return (bps / 1024 ** 2).toFixed(2) + ' MB/s'
  if (abs >= 1024) return (bps / 1024).toFixed(2) + ' KB/s'
  return Math.round(bps) + ' B/s'
}

export function formatSpeedParts(bps: number | null | undefined): { value: string; unit: string } {
  const text = formatSpeed(bps)
  const [value, ...unit] = text.split(' ')
  return { value: value || '0', unit: unit.join(' ') || 'B/s' }
}

export function formatBytes(bytes: number | null | undefined): string {
  if (bytes == null || isNaN(bytes)) return '0 B'
  const abs = Math.abs(bytes)
  if (abs >= 1024 ** 4) return (bytes / 1024 ** 4).toFixed(2) + ' TB'
  if (abs >= 1024 ** 3) return (bytes / 1024 ** 3).toFixed(2) + ' GB'
  if (abs >= 1024 ** 2) return (bytes / 1024 ** 2).toFixed(2) + ' MB'
  if (abs >= 1024) return (bytes / 1024).toFixed(2) + ' KB'
  return Math.round(bytes) + ' B'
}

export function formatUptime(seconds: number | null | undefined): string {
  if (!seconds || seconds <= 0) return '—'
  const d = Math.floor(seconds / 86400)
  const h = Math.floor((seconds % 86400) / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (d > 0) return `${d}天 ${h}时 ${m}分`
  if (h > 0) return `${h}时 ${m}分`
  return `${m}分钟`
}

export function formatTimestamp(ts: unknown): string {
  if (ts === null || ts === undefined || ts === '' || ts === '--') return '—'
  const num = Number(ts)
  if (Number.isNaN(num) || num <= 0) return String(ts)
  const ms = num > 1e12 ? num : num * 1000
  const d = new Date(ms)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

export function formatObjectRef(item: unknown): string {
  if (item === null || item === undefined) return ''
  if (typeof item === 'string' || typeof item === 'number') return String(item)
  if (typeof item === 'object') {
    const obj = item as Record<string, unknown>
    for (const key of ['group_name', 'name', 'tagname', 'label', 'value', 'id']) {
      const val = obj[key]
      if (val !== null && val !== undefined && String(val).trim() !== '') {
        return String(val).trim()
      }
    }
    return JSON.stringify(item)
  }
  return String(item)
}

export function formatCompoundValue(val: unknown): string {
  if (val === null || val === undefined || val === '') return '全部 / 默认'
  if (typeof val === 'string' || typeof val === 'number') return String(val)
  if (Array.isArray(val)) {
    if (val.length === 0) return '全部'
    return val
      .map((item) => {
        if (item && typeof item === 'object') {
          const entries = Object.values(item as Record<string, unknown>)
          return entries.join(':')
        }
        return String(item)
      })
      .join(', ')
  }
  if (typeof val === 'object') {
    const obj = val as Record<string, unknown>
    const custom = Array.isArray(obj.custom) ? obj.custom : []
    const objectRefs = Array.isArray(obj.object) ? obj.object : []
    const parts = [
      ...custom.map(String),
      ...objectRefs.map((x) => {
        const refLabel = formatObjectRef(x)
        return refLabel ? `@${refLabel}` : ''
      }),
    ].filter(Boolean)
    return parts.length > 0 ? parts.join(', ') : '全部'
  }
  return String(val)
}

export function compoundToCSV(val: unknown): string {
  if (val === null || val === undefined || val === '') return ''
  if (typeof val === 'string' || typeof val === 'number') return String(val)
  if (Array.isArray(val)) {
    return val.map(String).filter(Boolean).join(', ')
  }
  if (typeof val === 'object') {
    const obj = val as Record<string, unknown>
    const custom = Array.isArray(obj.custom) ? obj.custom : []
    return custom.map(String).filter(Boolean).join(', ')
  }
  return String(val)
}

export function formatCountCompact(value: number | null | undefined): string {
  if (value == null || isNaN(value)) return '0'
  const abs = Math.abs(value)
  if (abs >= 1_000_000) return `${(value / 1_000_000).toFixed(1).replace(/\.0$/, '')}M`
  if (abs >= 10_000) return `${(value / 1_000).toFixed(1).replace(/\.0$/, '')}k`
  return value.toLocaleString()
}

export function formatDurationCompact(value: string | null | undefined): string {
  if (!value) return '—'
  return value
    .replace(/\s*hours?\s*/gi, 'h ')
    .replace(/\s*minutes?\s*/gi, 'm ')
    .replace(/\s*seconds?\s*/gi, 's')
    .replace(/\s+/g, ' ')
    .trim()
}

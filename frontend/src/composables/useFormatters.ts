/**
 * 数值格式化工具集
 * 所有速率/流量值来自后端，单位均为字节（Bytes）
 */

export function formatSpeed(bps: number | null | undefined): string {
  if (bps == null || isNaN(bps)) return '0 B/s'
  const abs = Math.abs(bps)
  if (abs >= 1024 ** 3) return (bps / 1024 ** 3).toFixed(2) + ' GB/s'
  if (abs >= 1024 ** 2) return (bps / 1024 ** 2).toFixed(2) + ' MB/s'
  if (abs >= 1024)      return (bps / 1024).toFixed(2) + ' KB/s'
  return bps + ' B/s'
}

export function formatSpeedParts(bps: number | null | undefined): { value: string; unit: string } {
  const text = formatSpeed(bps)
  const [value, ...unit] = text.split(' ')
  return { value, unit: unit.join(' ') }
}

export function formatBytes(bytes: number | null | undefined): string {
  if (bytes == null || isNaN(bytes)) return '0 B'
  const abs = Math.abs(bytes)
  if (abs >= 1024 ** 4) return (bytes / 1024 ** 4).toFixed(2) + ' TB'
  if (abs >= 1024 ** 3) return (bytes / 1024 ** 3).toFixed(2) + ' GB'
  if (abs >= 1024 ** 2) return (bytes / 1024 ** 2).toFixed(2) + ' MB'
  if (abs >= 1024)      return (bytes / 1024).toFixed(2) + ' KB'
  return bytes + ' B'
}

export function formatCountCompact(value: number | null | undefined): string {
  if (value == null || isNaN(value)) return '0'
  const abs = Math.abs(value)
  if (abs >= 1_000_000) return `${trimFixed(value / 1_000_000)}m`
  if (abs >= 10_000) return `${trimFixed(value / 1_000)}k`
  return value.toLocaleString()
}

export function formatDurationCompact(value: string | null | undefined): string {
  if (!value) return '—'
  return value
    .replace(/\s*hours?\s*/gi, 'h ')
    .replace(/\s*minutes?\s*/gi, 'm ')
    .replace(/\s*seconds?\s*/gi, 's')
    .replace(/\s*小时\s*/g, 'h ')
    .replace(/\s*分钟\s*/g, 'm ')
    .replace(/\s*秒\s*/g, 's')
    .replace(/\s+/g, ' ')
    .trim()
}

function trimFixed(value: number): string {
  return value.toFixed(1).replace(/\.0$/, '')
}

export function getIPv4(ips: string[] | null | undefined): string {
  if (!ips || !ips.length) return '—'
  return ips.find(ip => !ip.includes(':')) || ips[0]
}

export function getIPv6(ips: string[] | null | undefined): string | null {
  if (!ips || !ips.length) return null
  return ips.find(ip => ip.includes(':')) || null
}

export function mapStatus(status: string | null | undefined): string {
  if (!status) return 'unknown'
  const s = status.toLowerCase()
  if (s === 'success' || s === 'ok' || s === 'connected') return 'success'
  if (s === 'fail' || s === 'error' || s === 'disconnected') return 'error'
  return 'warning'
}

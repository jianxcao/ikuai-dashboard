import { ref } from 'vue'
import type { OverviewDTO } from '@/api/v4'

export type WanHistoryPoint = OverviewDTO['wan_history'][number]

const DEFAULT_BUCKET_SEC = 60
const DEFAULT_WINDOW_SEC = 3600
const MAX_WINDOW_SEC = 10800
const MAX_POINTS = 180

interface LiveBucket {
  ts: number
  upSum: number
  downSum: number
  maxUp: number
  maxDown: number
  rttSum: number
  rttCount: number
  maxRtt: number
  samples: number
}

interface RouterWanBuffer {
  points: Map<number, WanHistoryPoint>
  bucketSec: number
  windowSec: number
  lastOfficialTs: number
  anchorLocalMs: number
  liveBucket: LiveBucket | null
}

/**
 * 将底层接口的 wan_history 规范化为按时间戳升序的全外网聚合序列。
 * 当固件仅按单个 WAN 口返回明细而非 `interface === "all"` 时，自动按时间戳合并各 WAN 口流量并平均 RTT。
 */
export function normalizeWanHistoryRows(history: WanHistoryPoint[]): WanHistoryPoint[] {
  if (!history || history.length === 0) return []
  const allRows = history.filter((p) => p.interface === 'all')
  if (allRows.length > 0) {
    const dedup = new Map<number, WanHistoryPoint>()
    for (const p of allRows) {
      if (p.timestamp > 0) dedup.set(p.timestamp, p)
    }
    return Array.from(dedup.values()).sort((a, b) => a.timestamp - b.timestamp)
  }

  type AccPoint = WanHistoryPoint & { rttSum: number; rttCount: number }
  const byTs = new Map<number, AccPoint>()
  for (const p of history) {
    if (!p.timestamp || p.timestamp <= 0) continue
    const existing = byTs.get(p.timestamp)
    if (!existing) {
      byTs.set(p.timestamp, {
        ...p,
        interface: 'all',
        rttSum: p.avg_rtt_ms > 0 ? p.avg_rtt_ms : 0,
        rttCount: p.avg_rtt_ms > 0 ? 1 : 0,
      })
    } else {
      existing.avg_upload_kbps += p.avg_upload_kbps
      existing.avg_download_kbps += p.avg_download_kbps
      existing.max_upload_kbps += p.max_upload_kbps
      existing.max_download_kbps += p.max_download_kbps
      if (p.avg_rtt_ms > 0) {
        existing.rttSum += p.avg_rtt_ms
        existing.rttCount += 1
      }
      if (p.max_rtt_ms > existing.max_rtt_ms) {
        existing.max_rtt_ms = p.max_rtt_ms
      }
      if (p.min_rtt_ms > 0 && (existing.min_rtt_ms <= 0 || p.min_rtt_ms < existing.min_rtt_ms)) {
        existing.min_rtt_ms = p.min_rtt_ms
      }
    }
  }

  return Array.from(byTs.values())
    .sort((a, b) => a.timestamp - b.timestamp)
    .map((item) => ({
      timestamp: item.timestamp,
      interface: 'all',
      avg_upload_kbps: Math.round(item.avg_upload_kbps * 100) / 100,
      avg_download_kbps: Math.round(item.avg_download_kbps * 100) / 100,
      max_upload_kbps: Math.round(item.max_upload_kbps * 100) / 100,
      max_download_kbps: Math.round(item.max_download_kbps * 100) / 100,
      avg_rtt_ms: item.rttCount > 0 ? Math.round((item.rttSum / item.rttCount) * 100) / 100 : 0,
      max_rtt_ms: Math.round(item.max_rtt_ms * 100) / 100,
      min_rtt_ms: Math.round(item.min_rtt_ms * 100) / 100,
      drop_rate: item.drop_rate,
    }))
}

function detectBucketSeconds(sorted: WanHistoryPoint[]): number {
  if (sorted.length < 2) return DEFAULT_BUCKET_SEC
  const deltas: number[] = []
  for (let i = 1; i < sorted.length; i += 1) {
    const diff = sorted[i]!.timestamp - sorted[i - 1]!.timestamp
    if (diff >= 10 && diff <= 900) deltas.push(diff)
  }
  if (deltas.length === 0) return DEFAULT_BUCKET_SEC
  deltas.sort((a, b) => a - b)
  return deltas[Math.floor(deltas.length / 2)] || DEFAULT_BUCKET_SEC
}

const routerBuffers = new Map<string, RouterWanBuffer>()

function getOrCreateBuffer(routerKey: string): RouterWanBuffer {
  let buf = routerBuffers.get(routerKey)
  if (!buf) {
    buf = {
      points: new Map(),
      bucketSec: DEFAULT_BUCKET_SEC,
      windowSec: DEFAULT_WINDOW_SEC,
      lastOfficialTs: 0,
      anchorLocalMs: 0,
      liveBucket: null,
    }
    routerBuffers.set(routerKey, buf)
  }
  return buf
}

export function ingestWanOverview(routerKey: string, ov: OverviewDTO): WanHistoryPoint[] {
  const key = routerKey || ov.router?.id || 'default'
  const buf = getOrCreateBuffer(key)
  const official = normalizeWanHistoryRows(ov.wan_history || [])

  if (official.length >= 2) {
    buf.bucketSec = detectBucketSeconds(official)
    const span = official[official.length - 1]!.timestamp - official[0]!.timestamp
    if (span > DEFAULT_WINDOW_SEC) {
      buf.windowSec = Math.min(MAX_WINDOW_SEC, span)
    }
  }

  let maxOffTs = buf.lastOfficialTs
  for (const pt of official) {
    buf.points.set(pt.timestamp, pt)
    if (pt.timestamp > maxOffTs) maxOffTs = pt.timestamp
  }

  const nowMs = Date.now()
  if (maxOffTs > buf.lastOfficialTs) {
    buf.lastOfficialTs = maxOffTs
    buf.anchorLocalMs = nowMs
    if (buf.liveBucket && buf.liveBucket.ts <= maxOffTs) {
      buf.liveBucket = null
    }
  } else if (buf.lastOfficialTs > 0 && buf.anchorLocalMs === 0) {
    buf.anchorLocalMs = nowMs
  }

  // 把 4 秒实时外网吞吐与链路延迟增量合并到最新时间桶，避免等待路由器历史整分采样时曲线静止或闪烁
  const bucketSec = buf.bucketSec
  const liveTs =
    buf.lastOfficialTs > 0
      ? buf.lastOfficialTs +
        Math.max(1, Math.floor(Math.max(0, nowMs - buf.anchorLocalMs) / (bucketSec * 1000)) + 1) * bucketSec
      : Math.floor((ov.collected_at || Math.floor(nowMs / 1000)) / bucketSec) * bucketSec

  const upKBps = Math.max(0, (ov.stream?.upload_bps ?? 0) / 1024)
  const downKBps = Math.max(0, (ov.stream?.download_bps ?? 0) / 1024)
  const validChecks = (ov.wan_checks || []).filter((w) => w.avg_rtt_ms > 0)
  const avgRtt =
    validChecks.length > 0
      ? validChecks.reduce((acc, item) => acc + item.avg_rtt_ms, 0) / validChecks.length
      : 0
  const maxRtt = validChecks.reduce((acc, item) => Math.max(acc, item.avg_rtt_ms), 0)

  if (!buf.liveBucket || buf.liveBucket.ts !== liveTs) {
    buf.liveBucket = {
      ts: liveTs,
      upSum: upKBps,
      downSum: downKBps,
      maxUp: upKBps,
      maxDown: downKBps,
      rttSum: avgRtt,
      rttCount: avgRtt > 0 ? 1 : 0,
      maxRtt,
      samples: 1,
    }
  } else {
    buf.liveBucket.upSum += upKBps
    buf.liveBucket.downSum += downKBps
    buf.liveBucket.maxUp = Math.max(buf.liveBucket.maxUp, upKBps)
    buf.liveBucket.maxDown = Math.max(buf.liveBucket.maxDown, downKBps)
    if (avgRtt > 0) {
      buf.liveBucket.rttSum += avgRtt
      buf.liveBucket.rttCount += 1
      buf.liveBucket.maxRtt = Math.max(buf.liveBucket.maxRtt, maxRtt)
    }
    buf.liveBucket.samples += 1
  }

  const lb = buf.liveBucket
  buf.points.set(liveTs, {
    timestamp: liveTs,
    interface: 'all',
    avg_upload_kbps: Math.round((lb.upSum / lb.samples) * 100) / 100,
    avg_download_kbps: Math.round((lb.downSum / lb.samples) * 100) / 100,
    max_upload_kbps: Math.round(lb.maxUp * 100) / 100,
    max_download_kbps: Math.round(lb.maxDown * 100) / 100,
    avg_rtt_ms: lb.rttCount > 0 ? Math.round((lb.rttSum / lb.rttCount) * 100) / 100 : 0,
    max_rtt_ms: Math.round(lb.maxRtt * 100) / 100,
    min_rtt_ms: 0,
    drop_rate: 0,
  })

  // 按滑动时间窗和最大点数自动淘汰过期旧数据
  let latestTs = 0
  for (const ts of buf.points.keys()) {
    if (ts > latestTs) latestTs = ts
  }
  const cutoff = latestTs - buf.windowSec
  for (const ts of buf.points.keys()) {
    if (ts < cutoff) {
      buf.points.delete(ts)
    }
  }

  const sorted = Array.from(buf.points.values()).sort((a, b) => a.timestamp - b.timestamp)
  if (sorted.length > MAX_POINTS) {
    const dropCount = sorted.length - MAX_POINTS
    for (let i = 0; i < dropCount; i += 1) {
      buf.points.delete(sorted[i]!.timestamp)
    }
    return sorted.slice(dropCount)
  }
  return sorted
}

export function useWanHistoryBuffer() {
  const seriesPoints = ref<WanHistoryPoint[]>([])

  function ingest(routerKey: string, ov: OverviewDTO): WanHistoryPoint[] {
    const next = ingestWanOverview(routerKey, ov)
    seriesPoints.value = next
    return next
  }

  function clearRouter(routerKey: string) {
    routerBuffers.delete(routerKey)
    seriesPoints.value = []
  }

  return {
    seriesPoints,
    ingest,
    clearRouter,
  }
}

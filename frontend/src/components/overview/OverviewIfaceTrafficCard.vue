<template>
  <OverviewSectionShell
    title="接口实时吞吐明细"
    :layoutLocked="layoutLocked"
    :canMovePrev="canMovePrev"
    :canMoveNext="canMoveNext"
    @move="emit('move', $event)"
  >
    <template #icon>
      <Network :size="16" />
    </template>
    <template #actions>
      <span
        v-if="wanTrafficSummary.count > 0"
        class="v4-badge font-mono"
        title="接口明细来自 /monitoring/interfaces-status；此处汇总全部 WAN 口实时速率（不包含内网 LAN 口），可与顶部外网总速率对照"
      >
        WAN 合计 ↑{{ formatSpeed(wanTrafficSummary.up) }} · ↓{{ formatSpeed(wanTrafficSummary.down) }}
      </span>
    </template>

    <div class="table-scroll">
      <table class="mini-table">
        <thead>
          <tr>
            <th
              v-for="col in ifaceColumns"
              :key="col.key"
              class="sortable-th"
              :class="{ 'is-sorted': ifaceSort.sortKey.value === col.key }"
              :aria-sort="ifaceSort.ariaSort(col.key)"
              tabindex="0"
              @click="ifaceSort.toggleSort(col.key)"
              @keydown.enter.prevent="ifaceSort.toggleSort(col.key)"
              @keydown.space.prevent="ifaceSort.toggleSort(col.key)"
            >
              <span class="sort-th-inner">
                <span>{{ col.label }}</span>
                <SortIndicator :active="ifaceSort.sortKey.value === col.key" :dir="ifaceSort.sortDir.value" />
              </span>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="iface in sortedInterfaceTraffic" :key="iface.interface">
            <td class="font-mono">
              <div class="iface-name-cell">
                <strong>{{ iface.interface }}</strong>
                <span class="iface-role-tag" :class="isWanInterface(iface.interface) ? 'wan' : 'lan'">
                  {{ isWanInterface(iface.interface) ? 'WAN' : 'LAN' }}
                </span>
              </div>
            </td>
            <td class="font-mono">{{ iface.ip_addr || '—' }}</td>
            <td class="font-mono ic-up">{{ formatSpeed(iface.upload_bps) }}</td>
            <td class="font-mono ic-down">{{ formatSpeed(iface.download_bps) }}</td>
            <td class="font-mono">↑{{ formatBytes(iface.total_up) }} / ↓{{ formatBytes(iface.total_down) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </OverviewSectionShell>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Network } from 'lucide-vue-next'
import type { OverviewDTO } from '@/api/v4'
import { formatBytes, formatSpeed } from '@/composables/useFormatters'
import SortIndicator from '@/components/common/SortIndicator.vue'
import { type SortValueHint, useTableSort } from '@/composables/useTableSort'
import OverviewSectionShell from './OverviewSectionShell.vue'

const props = defineProps<{
  interfaces: OverviewDTO['interface_traffic']
  wanChecks: OverviewDTO['wan_checks']
  layoutLocked: boolean
  canMovePrev: boolean
  canMoveNext: boolean
}>()

const emit = defineEmits<{
  (e: 'move', delta: -1 | 1): void
}>()

function isWanInterface(name: string): boolean {
  const clean = name.trim().toLowerCase()
  if (!clean) return false
  if (props.wanChecks.some((w) => w.interface.trim().toLowerCase() === clean)) {
    return true
  }
  return !(clean.startsWith('lan') || clean.startsWith('vlan') || clean.startsWith('br'))
}

const wanTrafficSummary = computed(() => {
  let up = 0
  let down = 0
  let count = 0
  for (const item of props.interfaces) {
    if (isWanInterface(item.interface)) {
      up += item.upload_bps || 0
      down += item.download_bps || 0
      count += 1
    }
  }
  return { up, down, count }
})

type IfaceSortKey = 'interface' | 'ip_addr' | 'upload_bps' | 'download_bps' | 'total_bytes'
const ifaceColumns: Array<{ key: IfaceSortKey; label: string; hint: SortValueHint }> = [
  { key: 'interface', label: '接口', hint: 'auto' },
  { key: 'ip_addr', label: 'IP', hint: 'ip' },
  { key: 'upload_bps', label: '实时上行', hint: 'number' },
  { key: 'download_bps', label: '实时下行', hint: 'number' },
  { key: 'total_bytes', label: '累计总量', hint: 'number' },
]

const ifaceSort = useTableSort<IfaceSortKey>({
  defaultDirForKey: (key) => (key === 'interface' || key === 'ip_addr' ? 'asc' : 'desc'),
})

const sortedInterfaceTraffic = computed(() =>
  ifaceSort.sortRows(
    props.interfaces,
    (iface, key) => {
      if (key === 'total_bytes') return (iface.total_up || 0) + (iface.total_down || 0)
      return iface[key]
    },
    (key) => ifaceColumns.find((c) => c.key === key)?.hint ?? 'auto'
  )
)
</script>

<style scoped>
.table-scroll {
  overflow-x: auto;
}

.mini-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.mini-table th {
  text-align: left;
  padding: 8px;
  color: var(--text-muted);
  font-size: 11px;
  border-bottom: 1px solid var(--divider-line);
  white-space: nowrap;
}

.mini-table td {
  padding: 9px 8px;
  border-bottom: 1px solid var(--divider-line);
  white-space: nowrap;
}

.iface-name-cell {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.iface-role-tag {
  padding: 1px 6px;
  border-radius: 999px;
  font-size: 10px;
  font-weight: 700;
  border: 1px solid var(--divider-line);
}

.iface-role-tag.wan {
  color: var(--accent-text);
  background: var(--accent-primary-soft);
  border-color: color-mix(in srgb, var(--accent-primary) 26%, transparent);
}

.iface-role-tag.lan {
  color: var(--text-muted);
  background: var(--surface-input);
}

.ic-up { color: var(--chart-up); }
.ic-down { color: var(--chart-down); }
</style>

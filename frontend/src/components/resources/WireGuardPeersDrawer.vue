<template>
  <teleport to="body">
    <div v-if="open" class="drawer-backdrop" @click.self="emit('close')">
      <div class="drawer-panel v4-panel wide-drawer" role="dialog" aria-label="WireGuard Peers 管理">
        <div class="drawer-head">
          <div>
            <strong>WireGuard Peers · {{ wgName }} (#{{ wgId }})</strong>
            <span class="font-mono">/vpn/wireguard/{{ wgId }}/peers</span>
          </div>
          <button type="button" class="icon-btn" aria-label="关闭抽屉" @click="emit('close')">
            <X :size="16" />
          </button>
        </div>

        <div class="drawer-body">
          <div v-if="error" class="error-banner">{{ error }}</div>

          <div class="peer-form v4-card">
            <div class="form-grid two-cols">
              <label class="form-field wide-col">
                <span class="field-label">对端公钥 (peer_publickey) *</span>
                <input v-model.trim="form.peer_publickey" class="v4-input font-mono" placeholder="Base64 公钥" />
              </label>
              <label class="form-field">
                <span class="field-label">允许子网 (allowips) *</span>
                <input v-model.trim="form.allowips" class="v4-input font-mono" placeholder="10.9.0.2/32" />
              </label>
              <label class="form-field">
                <span class="field-label">对端 Endpoint</span>
                <input v-model.trim="form.endpoint" class="v4-input font-mono" placeholder="可选 IP/域名" />
              </label>
              <label class="form-field">
                <span class="field-label">对端端口</span>
                <input v-model.number="form.endpoint_port" type="number" class="v4-input font-mono" />
              </label>
              <label class="form-field">
                <span class="field-label">备注</span>
                <input v-model.trim="form.comment" class="v4-input" placeholder="分支节点" />
              </label>
            </div>
            <div class="confirm-actions">
              <button type="button" class="v4-btn v4-btn-primary btn-xs" :disabled="readOnly || busy" @click="addPeer">
                <Plus :size="13" />
                <span>添加 Peer</span>
              </button>
            </div>
          </div>

          <div v-if="loading" class="loading-state"><div class="spinner" /></div>
          <div v-else-if="rows.length === 0" class="empty-state">该隧道暂无 Peer 对端</div>
          <div v-else class="table-wrap">
            <table class="v4-table">
              <thead>
                <tr>
                  <th
                    class="sortable-th"
                    :class="{ 'is-sorted': peerSort.sortKey.value === 'id' }"
                    :aria-sort="peerSort.ariaSort('id')"
                    tabindex="0"
                    @click="peerSort.toggleSort('id')"
                    @keydown.enter.prevent="peerSort.toggleSort('id')"
                    @keydown.space.prevent="peerSort.toggleSort('id')"
                  >
                    <span class="sort-th-inner">
                      <span>ID</span>
                      <SortIndicator :active="peerSort.sortKey.value === 'id'" :dir="peerSort.sortDir.value" />
                    </span>
                  </th>
                  <th
                    class="sortable-th"
                    :class="{ 'is-sorted': peerSort.sortKey.value === 'allowips' }"
                    :aria-sort="peerSort.ariaSort('allowips')"
                    tabindex="0"
                    @click="peerSort.toggleSort('allowips')"
                    @keydown.enter.prevent="peerSort.toggleSort('allowips')"
                    @keydown.space.prevent="peerSort.toggleSort('allowips')"
                  >
                    <span class="sort-th-inner">
                      <span>允许子网</span>
                      <SortIndicator :active="peerSort.sortKey.value === 'allowips'" :dir="peerSort.sortDir.value" />
                    </span>
                  </th>
                  <th
                    class="sortable-th"
                    :class="{ 'is-sorted': peerSort.sortKey.value === 'endpoint' }"
                    :aria-sort="peerSort.ariaSort('endpoint')"
                    tabindex="0"
                    @click="peerSort.toggleSort('endpoint')"
                    @keydown.enter.prevent="peerSort.toggleSort('endpoint')"
                    @keydown.space.prevent="peerSort.toggleSort('endpoint')"
                  >
                    <span class="sort-th-inner">
                      <span>Endpoint</span>
                      <SortIndicator :active="peerSort.sortKey.value === 'endpoint'" :dir="peerSort.sortDir.value" />
                    </span>
                  </th>
                  <th
                    class="sortable-th"
                    :class="{ 'is-sorted': peerSort.sortKey.value === 'enabled' }"
                    :aria-sort="peerSort.ariaSort('enabled')"
                    tabindex="0"
                    @click="peerSort.toggleSort('enabled')"
                    @keydown.enter.prevent="peerSort.toggleSort('enabled')"
                    @keydown.space.prevent="peerSort.toggleSort('enabled')"
                  >
                    <span class="sort-th-inner">
                      <span>启用</span>
                      <SortIndicator :active="peerSort.sortKey.value === 'enabled'" :dir="peerSort.sortDir.value" />
                    </span>
                  </th>
                  <th>删除</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="peer in sortedPeers" :key="String(peer.id)">
                  <td class="font-mono">{{ peer.id }}</td>
                  <td class="font-mono">{{ peer.allowips }}</td>
                  <td class="font-mono">{{ peer.endpoint ? `${peer.endpoint}:${peer.endpoint_port || 51820}` : '动态' }}</td>
                  <td>
                    <button
                      type="button"
                      class="toggle-switch"
                      :class="{ on: isTruthy(peer.enabled) }"
                      :disabled="readOnly"
                      @click="flipPeer(peer)"
                    >
                      <span class="knob" />
                    </button>
                  </td>
                  <td>
                    <button type="button" class="icon-btn danger" :disabled="readOnly" @click="removePeer(peer)">
                      <Trash2 :size="13" />
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { Plus, Trash2, X } from 'lucide-vue-next'
import { createWGPeer, deleteWGPeer, fetchWGPeers, toggleWGPeer } from '@/api/v4'
import SortIndicator from '@/components/common/SortIndicator.vue'
import { useTableSort } from '@/composables/useTableSort'

const props = defineProps<{
  open: boolean
  wgRow: Record<string, unknown> | null
  readOnly: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const loading = ref(false)
const busy = ref(false)
const error = ref<string | null>(null)
const rows = ref<Array<Record<string, unknown>>>([])

const form = reactive({
  peer_publickey: '',
  allowips: '10.9.0.2/32',
  endpoint: '',
  endpoint_port: 51820,
  comment: '',
})

const wgId = computed(() => String(props.wgRow?.id ?? ''))
const wgName = computed(() => String(props.wgRow?.name ?? wgId.value))

const peerSort = useTableSort<'id' | 'allowips' | 'endpoint' | 'enabled'>({
  defaultDirForKey: (key) => (key === 'enabled' ? 'desc' : 'asc'),
})

const sortedPeers = computed(() =>
  peerSort.sortRows(
    rows.value,
    (peer, key) => {
      if (key === 'endpoint') {
        return peer.endpoint ? `${peer.endpoint}:${peer.endpoint_port || 51820}` : ''
      }
      return peer[key]
    },
    (key) => (key === 'enabled' ? 'boolean' : key === 'id' ? 'number' : 'auto')
  )
)

function isTruthy(val: unknown): boolean {
  if (typeof val === 'boolean') return val
  const s = String(val ?? '').toLowerCase()
  return s === 'yes' || s === '1' || s === 'true' || s === 'on'
}

function extractError(e: unknown): string {
  if (e && typeof e === 'object' && 'response' in e) {
    const resp = (e as { response?: { data?: { message?: string } } }).response
    if (resp?.data?.message) return resp.data.message
  }
  return e instanceof Error ? e.message : String(e)
}

async function reloadPeers() {
  if (!wgId.value) return
  loading.value = true
  error.value = null
  try {
    const res = await fetchWGPeers(wgId.value)
    rows.value = res.rows || []
  } catch (e: unknown) {
    error.value = extractError(e)
  } finally {
    loading.value = false
  }
}

async function addPeer() {
  if (!form.peer_publickey || !form.allowips) {
    error.value = '请填写对端公钥和允许子网'
    return
  }
  busy.value = true
  error.value = null
  try {
    await createWGPeer(wgId.value, {
      ...form,
      interface: wgName.value,
      enabled: 'yes',
    })
    form.peer_publickey = ''
    await reloadPeers()
  } catch (e: unknown) {
    error.value = extractError(e)
  } finally {
    busy.value = false
  }
}

async function flipPeer(peer: Record<string, unknown>) {
  try {
    const next = isTruthy(peer.enabled) ? 'no' : 'yes'
    await toggleWGPeer(wgId.value, String(peer.id), next)
    peer.enabled = next
  } catch (e: unknown) {
    error.value = extractError(e)
  }
}

async function removePeer(peer: Record<string, unknown>) {
  try {
    await deleteWGPeer(wgId.value, String(peer.id))
    await reloadPeers()
  } catch (e: unknown) {
    error.value = extractError(e)
  }
}

watch(
  () => [props.open, props.wgRow] as const,
  ([isOpen, row]) => {
    if (isOpen && row) {
      error.value = null
      peerSort.resetSort()
      void reloadPeers()
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.drawer-backdrop {
  position: fixed;
  inset: 0;
  z-index: 250;
  background: rgba(6, 10, 20, 0.42);
  -webkit-backdrop-filter: blur(8px);
  backdrop-filter: blur(8px);
  display: flex;
  justify-content: flex-end;
}

.drawer-panel {
  width: min(680px, 100vw);
  height: 100%;
  border-radius: var(--radius-xl) 0 0 var(--radius-xl);
  display: flex;
  flex-direction: column;
  background: var(--surface-popover);
  border-color: var(--chrome-border);
  box-shadow: var(--chrome-shadow);
  -webkit-backdrop-filter: var(--backdrop-chrome);
  backdrop-filter: var(--backdrop-chrome);
}

.drawer-head {
  padding: 16px 20px;
  border-bottom: 1px solid var(--divider-line);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.drawer-head > div > strong {
  display: block;
  font-size: 15px;
  color: var(--text-strong);
}

.drawer-head > div > span {
  font-size: 11px;
  color: var(--text-muted);
}

.drawer-body {
  flex: 1;
  overflow-y: auto;
  padding: 18px 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.peer-form {
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.form-grid.two-cols {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.wide-col {
  grid-column: 1 / -1;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-normal);
}

.confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.table-wrap {
  overflow-x: auto;
  border: 1px solid var(--divider-line);
  border-radius: var(--radius-md);
}

.v4-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.v4-table th {
  text-align: left;
  padding: 10px 12px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  background: var(--surface-table-head);
  border-bottom: 1px solid var(--divider-line);
  white-space: nowrap;
}

.v4-table td {
  padding: 10px 12px;
  border-bottom: 1px solid var(--divider-line);
  color: var(--text-normal);
  vertical-align: middle;
  white-space: nowrap;
}

.v4-table tbody tr:last-child td {
  border-bottom: none;
}

.icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-control);
  background: var(--surface-btn);
  color: var(--text-normal);
  cursor: pointer;
}

.icon-btn:hover:not(:disabled) {
  border-color: var(--accent-primary);
  color: var(--text-strong);
}

.icon-btn.danger:hover:not(:disabled) {
  border-color: var(--status-down);
  color: var(--status-down);
  background: var(--status-down-soft);
}

.icon-btn:disabled {
  opacity: 0.42;
  cursor: not-allowed;
}

.toggle-switch {
  width: 40px;
  height: 22px;
  padding: 2px;
  border-radius: 999px;
  border: 1px solid var(--border-control);
  background: var(--surface-input);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  transition: background 0.2s ease, border-color 0.2s ease;
}

.toggle-switch:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.toggle-switch .knob {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: var(--text-muted);
  transition: transform 0.2s ease, background 0.2s ease;
}

.toggle-switch.on {
  border-color: var(--status-up);
  background: var(--status-up-soft);
}

.toggle-switch.on .knob {
  transform: translateX(18px);
  background: var(--status-up);
}

.empty-state {
  padding: 28px 16px;
  text-align: center;
  color: var(--text-muted);
}

@media (max-width: 640px) {
  .drawer-backdrop {
    align-items: flex-end;
  }

  .drawer-panel {
    width: 100%;
    height: auto;
    max-height: 88dvh;
    border-radius: var(--radius-xl) var(--radius-xl) 0 0;
  }

  .drawer-body {
    padding: 14px 16px;
  }

  .form-grid.two-cols {
    grid-template-columns: 1fr;
  }
}
</style>

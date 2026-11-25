<template>
  <section class="resource-panel v4-card">
    <header class="panel-head">
      <div class="head-copy">
        <div class="title-row">
          <h3>{{ spec?.label || resourceKey }}</h3>
          <span v-if="spec" class="v4-badge font-mono">{{ spec.v4_path }}</span>
          <span class="v4-badge">共 {{ total }} 条</span>
          <span
            v-if="isUpstreamReadOnly"
            class="v4-badge"
            title="爱快 V4.0 OpenAPI 与 ikuai-cli 当前未开放此资源的写接口"
          >
            V4 接口只读
          </span>
          <span v-if="readOnly" class="v4-badge v4-badge-warn">只读锁定</span>
        </div>
        <p v-if="spec?.description" class="desc">{{ spec.description }}</p>
      </div>

      <div class="head-actions">
        <button
          v-if="spec?.cli_command"
          type="button"
          class="cli-pill font-mono"
          :title="'点击复制命令：' + spec.cli_command"
          @click="copyCLI"
        >
          <Terminal :size="13" />
          <span>{{ spec.cli_command }}</span>
          <Check v-if="copiedCli" :size="13" class="ok-icon" />
          <Copy v-else :size="12" />
        </button>

        <button
          v-for="act in spec?.actions || []"
          :key="act.key"
          type="button"
          class="v4-btn"
          :class="act.dangerous ? 'v4-btn-danger' : 'v4-btn-primary'"
          :disabled="actionBusy || readOnly || loadFailed"
          :title="readOnly ? '只读模式下禁止写入动作' : act.label"
          @click="handleAction(act)"
        >
          <Zap :size="14" />
          <span>{{ act.label }}</span>
        </button>

        <button
          v-if="canCreate"
          type="button"
          class="v4-btn v4-btn-primary"
          :disabled="readOnly || loadFailed"
          :title="readOnly ? `只读模式下仍可在列表/表单点击 Dry-Run 预览，禁止直接${createButtonLabel}` : createButtonLabel"
          @click="openCreateDrawer"
        >
          <Plus :size="15" />
          <span>{{ createButtonLabel }}</span>
        </button>

        <button
          v-if="canClearAll"
          type="button"
          class="v4-btn v4-btn-danger"
          :disabled="readOnly || loadFailed"
          @click="confirmClearAll"
        >
          <Trash2 :size="14" />
          <span>清空全部</span>
        </button>
      </div>
    </header>

    <!-- Directory Navigation for system-files -->
    <div v-if="resourceKey === 'system-files'" class="path-bar">
      <button type="button" class="v4-btn btn-xs" @click="goPath('/')">根目录 /</button>
      <button type="button" class="v4-btn btn-xs" :disabled="filePath === '/'" @click="goParentPath">上级</button>
      <input
        v-model.trim="filePath"
        type="text"
        class="v4-input font-mono path-input"
        aria-label="目录路径"
        @keyup.enter="reload"
      />
      <button type="button" class="v4-btn btn-xs v4-btn-primary" @click="reload">跳转</button>
    </div>

    <div class="toolbar">
      <div class="search-box">
        <Search :size="14" class="search-ic" />
        <input
          v-model.trim="search"
          type="text"
          class="v4-input search-input"
          placeholder="搜索关键字、IP、MAC 或备注..."
          @keyup.enter="reload"
        />
      </div>
      <div class="toolbar-right">
        <button type="button" class="v4-btn" :disabled="loading" @click="reload">
          <RefreshCw :size="14" :class="{ spin: loading }" />
          <span>刷新</span>
        </button>
      </div>
    </div>

    <div v-if="error" class="error-banner">{{ error }}</div>
    <div v-if="toast" class="toast-banner">{{ toast }}</div>

    <div class="table-wrap">
      <div v-if="loading && rows.length === 0" class="loading-state">
        <div class="spinner" />
        <span class="font-mono">正在读取 {{ spec?.v4_path || resourceKey }}...</span>
      </div>

      <div v-else-if="rows.length === 0" class="empty-state">
        暂无数据记录
      </div>

      <table v-else class="v4-table">
        <thead>
          <tr>
            <th v-for="col in displayColumns" :key="col.key">{{ col.label }}</th>
            <th class="actions-col">{{ isUpstreamReadOnly ? '详情' : '操作' }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, idx) in rows" :key="rowId(row, idx)">
            <td v-for="col in displayColumns" :key="col.key" :class="{ 'font-mono': col.kind === 'mono' || col.kind === 'time' }">
              <template v-if="col.kind === 'toggle'">
                <button
                  type="button"
                  class="toggle-switch"
                  :class="{ on: isTruthy(row[col.key]) }"
                  :disabled="!canToggle || readOnly || busyId === rowId(row, idx)"
                  :aria-pressed="isTruthy(row[col.key])"
                  :title="readOnly ? '只读保护中' : isTruthy(row[col.key]) ? '点击禁用' : '点击启用'"
                  @click="handleToggle(row)"
                >
                  <span class="knob" />
                </button>
              </template>
              <template v-else-if="col.kind === 'badge'">
                <span class="v4-badge" :class="badgeTone(row[col.key])">{{ formatCell(row[col.key], col.kind, col.key) }}</span>
              </template>
              <template v-else-if="col.kind === 'time'">
                {{ formatTimestamp(row[col.key]) }}
              </template>
              <template v-else-if="col.kind === 'bytes'">
                {{ formatBytes(Number(row[col.key])) }}
              </template>
              <template v-else-if="col.kind === 'compound'">
                <span class="compound-chip">{{ formatCompoundValue(row[col.key]) }}</span>
              </template>
              <template v-else>
                {{ formatCell(row[col.key], col.kind, col.key) }}
              </template>
            </td>

            <td class="actions-col">
              <div class="row-actions">
                <button
                  v-if="resourceKey === 'system-files' && Number(row.st_type) === 0"
                  type="button"
                  class="v4-btn btn-xs"
                  title="进入子目录"
                  @click="enterFolder(String(row.f_name ?? ''))"
                >
                  <FolderOpen :size="13" />
                  <span>打开</span>
                </button>

                <button
                  v-if="resourceKey === 'dhcp-leases' && !readOnly"
                  type="button"
                  class="v4-btn btn-xs"
                  title="转为 DHCP 静态绑定"
                  @click="openBindStaticModal(row)"
                >
                  <Pin :size="13" />
                  <span>转静态</span>
                </button>

                <button
                  v-if="spec?.group === 'objects'"
                  type="button"
                  class="v4-btn btn-xs"
                  title="查询策略引用"
                  @click="openRefsModal(String(row.group_name ?? ''))"
                >
                  <Link2 :size="13" />
                  <span>引用</span>
                </button>

                <button
                  v-if="resourceKey === 'wireguard'"
                  type="button"
                  class="v4-btn btn-xs"
                  title="管理 WireGuard Peers"
                  @click="openPeersModal(row)"
                >
                  <Network :size="13" />
                  <span>Peers</span>
                </button>

                <button
                  type="button"
                  class="icon-btn"
                  title="查看完整 JSON"
                  @click="inspectRow = row"
                >
                  <Braces :size="14" />
                </button>

                <button
                  v-if="canUpdate"
                  type="button"
                  class="icon-btn"
                  :disabled="readOnly"
                  :title="readOnly ? '只读保护中' : '编辑'"
                  @click="openEditDrawer(row)"
                >
                  <Pencil :size="14" />
                </button>

                <button
                  v-if="canDeleteRow"
                  type="button"
                  class="icon-btn danger"
                  :disabled="readOnly"
                  :title="readOnly ? '只读保护中' : '删除'"
                  @click="askDeleteRow(row)"
                >
                  <Trash2 :size="14" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <footer v-if="spec?.supports_paging && total > pageSize" class="pagination-bar">
      <span class="font-mono">第 {{ page }} 页 / 共 {{ Math.ceil(total / pageSize) }} 页</span>
      <div class="pager-btns">
        <button type="button" class="v4-btn" :disabled="page <= 1" @click="changePage(page - 1)">上一页</button>
        <button type="button" class="v4-btn" :disabled="page * pageSize >= total" @click="changePage(page + 1)">下一页</button>
      </div>
    </footer>

    <!-- Create / Edit Drawer -->
    <teleport to="body">
      <div v-if="drawerOpen" class="drawer-backdrop" @click.self="drawerOpen = false">
        <div class="drawer-panel v4-panel" role="dialog" :aria-label="editingId ? `编辑${itemNoun}` : createButtonLabel">
          <div class="drawer-head">
            <div>
              <strong>{{ editingId ? `编辑 ${spec?.label} #${editingId}` : `新建 ${spec?.label}` }}</strong>
              <span class="font-mono">{{ spec?.v4_path }}</span>
            </div>
            <button type="button" class="icon-btn" @click="drawerOpen = false">
              <X :size="16" />
            </button>
          </div>

          <div class="drawer-body">
            <div class="mode-switch">
              <div class="v4-seg-bar" role="tablist">
                <button type="button" class="v4-seg-btn" :class="{ active: !rawJsonMode }" @click="exitRawJson">可视化表单</button>
                <button type="button" class="v4-seg-btn" :class="{ active: rawJsonMode }" @click="switchToRawJson">高级 JSON</button>
              </div>
              <button
                v-if="resourceKey === 'wireguard'"
                type="button"
                class="v4-btn btn-xs"
                :disabled="generatingKeys"
                @click="generateWGKeys"
              >
                <KeyRound :size="13" />
                <span>{{ generatingKeys ? '生成中...' : '生成随机密钥对' }}</span>
              </button>
            </div>

            <div v-if="drawerError" class="error-banner">{{ drawerError }}</div>

            <template v-if="!rawJsonMode">
              <div class="form-grid">
                <label v-for="field in spec?.fields || []" :key="field.key" class="form-field">
                  <span class="field-label">
                    {{ field.label }}
                    <em v-if="field.required">*</em>
                  </span>
                  <select v-if="field.type === 'select' || field.type === 'switch'" v-model="formState[field.key]" class="v4-select">
                    <option v-for="opt in field.options" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
                  </select>
                  <input
                    v-else-if="field.type === 'number'"
                    v-model.number="formState[field.key]"
                    type="number"
                    class="v4-input"
                    :placeholder="field.placeholder"
                  />
                  <input
                    v-else-if="field.type === 'password'"
                    v-model="formState[field.key]"
                    type="password"
                    class="v4-input"
                    :placeholder="field.placeholder"
                  />
                  <textarea
                    v-else-if="field.type === 'textarea'"
                    v-model="formState[field.key] as string"
                    class="v4-textarea font-mono"
                    rows="4"
                    :placeholder="field.placeholder"
                  />
                  <input
                    v-else
                    v-model="formState[field.key]"
                    type="text"
                    class="v4-input"
                    :placeholder="field.placeholder"
                  />
                  <span
                    v-if="field.type === 'csv' && compoundObjectRefs(field.key).length > 0"
                    class="obj-ref-hint font-mono"
                  >
                    已关联对象组: {{ compoundObjectRefs(field.key).join(', ') }}（可视化保存将自动保留）
                  </span>
                </label>
              </div>
            </template>

            <div v-else class="raw-json-wrap">
              <textarea v-model="rawJsonText" class="v4-textarea font-mono" rows="14" />
            </div>

            <div v-if="dryRunPreview" class="dryrun-box">
              <div class="dryrun-head">
                <strong>Dry-Run 请求预览（未写入路由器）</strong>
                <button type="button" class="icon-btn" @click="dryRunPreview = null"><X :size="13" /></button>
              </div>
              <pre class="font-mono">{{ JSON.stringify(dryRunPreview, null, 2) }}</pre>
            </div>
          </div>

          <div class="drawer-foot">
            <button type="button" class="v4-btn" :disabled="submitting" @click="submitForm(true)">
              <Eye :size="14" />
              <span>Dry-Run 预览</span>
            </button>
            <button type="button" class="v4-btn v4-btn-primary" :disabled="submitting || readOnly" @click="submitForm(false)">
              <Save :size="14" />
              <span>{{ submitting ? '提交中...' : '确认提交' }}</span>
            </button>
          </div>
        </div>
      </div>
    </teleport>

    <!-- Row JSON Inspector Drawer -->
    <teleport to="body">
      <div v-if="inspectRow" class="drawer-backdrop" @click.self="inspectRow = null">
        <div class="drawer-panel v4-panel" role="dialog" aria-label="记录详情 JSON">
          <div class="drawer-head">
            <div>
              <strong>{{ spec?.label }} · 记录完整字段</strong>
              <span class="font-mono">{{ spec?.v4_path }}</span>
            </div>
            <button type="button" class="icon-btn" @click="inspectRow = null"><X :size="16" /></button>
          </div>
          <div class="drawer-body">
            <pre class="json-pre font-mono">{{ JSON.stringify(inspectRow, null, 2) }}</pre>
          </div>
          <div class="drawer-foot">
            <button type="button" class="v4-btn" @click="copyInspectJson">
              <Copy :size="14" />
              <span>复制 JSON</span>
            </button>
            <button
              v-if="canUpdate && !readOnly"
              type="button"
              class="v4-btn v4-btn-primary"
              @click="openEditFromInspector"
            >
              <Pencil :size="14" />
              <span>编辑此记录</span>
            </button>
          </div>
        </div>
      </div>
    </teleport>

    <!-- Object References Modal -->
    <teleport to="body">
      <div v-if="refsModal.open" class="drawer-backdrop center" @click.self="refsModal.open = false">
        <div class="confirm-dialog v4-panel modal-wide">
          <div class="modal-top">
            <h4>对象组引用关系 · {{ refsModal.groupName }}</h4>
            <button type="button" class="icon-btn" @click="refsModal.open = false"><X :size="14" /></button>
          </div>
          <div v-if="refsModal.loading" class="loading-state"><div class="spinner" /></div>
          <div v-else-if="refsModal.error" class="error-banner">{{ refsModal.error }}</div>
          <div v-else-if="refsModal.rows.length === 0" class="empty-state">当前未被任何策略或分流规则引用</div>
          <pre v-else class="json-pre font-mono">{{ JSON.stringify(refsModal.rows, null, 2) }}</pre>
        </div>
      </div>
    </teleport>

    <!-- DHCP Lease -> Static Quick Bind Modal -->
    <teleport to="body">
      <div v-if="bindStatic.open" class="drawer-backdrop center" @click.self="bindStatic.open = false">
        <div class="confirm-dialog v4-panel modal-wide">
          <div class="modal-top">
            <h4>将动态租约转为 DHCP 静态绑定</h4>
            <button type="button" class="icon-btn" @click="bindStatic.open = false"><X :size="14" /></button>
          </div>
          <div v-if="bindStatic.error" class="error-banner">{{ bindStatic.error }}</div>
          <div class="form-grid">
            <label class="form-field">
              <span class="field-label">设备名称</span>
              <input v-model="bindStatic.tagname" class="v4-input" />
            </label>
            <label class="form-field">
              <span class="field-label">固定 IP</span>
              <input v-model="bindStatic.ip_addr" class="v4-input font-mono" />
            </label>
            <label class="form-field">
              <span class="field-label">MAC 地址</span>
              <input v-model="bindStatic.mac" class="v4-input font-mono" />
            </label>
            <label class="form-field">
              <span class="field-label">所属接口</span>
              <input v-model="bindStatic.interface" class="v4-input font-mono" />
            </label>
          </div>
          <div class="confirm-actions">
            <button type="button" class="v4-btn" @click="bindStatic.open = false">取消</button>
            <button type="button" class="v4-btn v4-btn-primary" :disabled="bindStatic.busy" @click="submitBindStatic">
              {{ bindStatic.busy ? '绑定中...' : '确认绑定' }}
            </button>
          </div>
        </div>
      </div>
    </teleport>

    <!-- WireGuard Peers Manager Drawer -->
    <teleport to="body">
      <div v-if="peersModal.open" class="drawer-backdrop" @click.self="peersModal.open = false">
        <div class="drawer-panel v4-panel wide-drawer" role="dialog" aria-label="WireGuard Peers 管理">
          <div class="drawer-head">
            <div>
              <strong>WireGuard Peers · {{ peersModal.wgName }} (#{{ peersModal.wgId }})</strong>
              <span class="font-mono">/vpn/wireguard/{{ peersModal.wgId }}/peers</span>
            </div>
            <button type="button" class="icon-btn" @click="peersModal.open = false"><X :size="16" /></button>
          </div>

          <div class="drawer-body">
            <div v-if="peersModal.error" class="error-banner">{{ peersModal.error }}</div>

            <div class="peer-form v4-card">
              <div class="form-grid two-cols">
                <label class="form-field wide-col">
                  <span class="field-label">对端公钥 (peer_publickey) *</span>
                  <input v-model.trim="peersModal.form.peer_publickey" class="v4-input font-mono" placeholder="Base64 公钥" />
                </label>
                <label class="form-field">
                  <span class="field-label">允许子网 (allowips) *</span>
                  <input v-model.trim="peersModal.form.allowips" class="v4-input font-mono" placeholder="10.9.0.2/32" />
                </label>
                <label class="form-field">
                  <span class="field-label">对端 Endpoint</span>
                  <input v-model.trim="peersModal.form.endpoint" class="v4-input font-mono" placeholder="可选 IP/域名" />
                </label>
                <label class="form-field">
                  <span class="field-label">对端端口</span>
                  <input v-model.number="peersModal.form.endpoint_port" type="number" class="v4-input font-mono" />
                </label>
                <label class="form-field">
                  <span class="field-label">备注</span>
                  <input v-model.trim="peersModal.form.comment" class="v4-input" placeholder="分支节点" />
                </label>
              </div>
              <div class="confirm-actions">
                <button type="button" class="v4-btn v4-btn-primary btn-xs" :disabled="readOnly || peersModal.busy" @click="addPeer">
                  <Plus :size="13" />
                  <span>添加 Peer</span>
                </button>
              </div>
            </div>

            <div v-if="peersModal.loading" class="loading-state"><div class="spinner" /></div>
            <div v-else-if="peersModal.rows.length === 0" class="empty-state">该隧道暂无 Peer 对端</div>
            <table v-else class="v4-table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>允许子网</th>
                  <th>Endpoint</th>
                  <th>启用</th>
                  <th>删除</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="peer in peersModal.rows" :key="String(peer.id)">
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
    </teleport>

    <!-- Delete Confirmation Modal -->
    <teleport to="body">
      <div v-if="deleteTarget" class="drawer-backdrop center" @click.self="deleteTarget = null">
        <div class="confirm-dialog v4-panel">
          <h4>危险操作确认</h4>
          <p>{{ deleteTarget.label }}</p>
          <div class="confirm-actions">
            <button type="button" class="v4-btn" @click="deleteTarget = null">取消</button>
            <button type="button" class="v4-btn v4-btn-danger" :disabled="submitting" @click="executeDelete">
              {{ submitting ? '执行中...' : '确认删除' }}
            </button>
          </div>
        </div>
      </div>
    </teleport>
  </section>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import {
  Braces,
  Check,
  Copy,
  Eye,
  FolderOpen,
  KeyRound,
  Link2,
  Network,
  Pencil,
  Pin,
  Plus,
  RefreshCw,
  Save,
  Search,
  Terminal,
  Trash2,
  X,
  Zap,
} from 'lucide-vue-next'
import {
  createV4Resource,
  createWGPeer,
  deleteV4Resource,
  deleteWGPeer,
  executeV4Action,
  fetchObjectRefs,
  fetchV4Resource,
  fetchWGPeers,
  generateWGKeypair,
  toggleV4Resource,
  toggleWGPeer,
  updateV4Resource,
  type ActionSpec,
  type ColumnSpec,
  type ResourceSpec,
} from '@/api/v4'
import { useAuth } from '@/composables/useAuth'
import { useCatalog } from '@/composables/useCatalog'
import { useFleet } from '@/composables/useFleet'
import { compoundToCSV, formatBytes, formatCompoundValue, formatObjectRef, formatTimestamp } from '@/composables/useFormatters'

const props = defineProps<{
  resourceKey: string
}>()

const { readOnly } = useAuth()
const { getSpec } = useCatalog()
const { routerEpoch } = useFleet()

const loading = ref(false)
const loadFailed = ref(false)
let reqSeq = 0
const actionBusy = ref(false)
const submitting = ref(false)
const busyId = ref<string | null>(null)
const error = ref<string | null>(null)
const toast = ref<string | null>(null)
const spec = ref<ResourceSpec | null>(null)
const rows = ref<Array<Record<string, unknown>>>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const search = ref('')
const filePath = ref('/')
const copiedCli = ref(false)

const drawerOpen = ref(false)
const editingId = ref<string | null>(null)
const editingRow = ref<Record<string, unknown> | null>(null)
const generatingKeys = ref(false)
const rawJsonMode = ref(false)
const rawJsonText = ref('{}')
const formState = ref<Record<string, unknown>>({})
const drawerError = ref<string | null>(null)
const dryRunPreview = ref<Record<string, unknown> | null>(null)

const inspectRow = ref<Record<string, unknown> | null>(null)
const deleteTarget = ref<{ id?: string; srcfile?: string; label: string } | null>(null)

const refsModal = reactive<{
  open: boolean
  loading: boolean
  groupName: string
  rows: Array<Record<string, unknown>>
  error: string | null
}>({ open: false, loading: false, groupName: '', rows: [], error: null })

const bindStatic = reactive({
  open: false,
  busy: false,
  error: null as string | null,
  tagname: '',
  ip_addr: '',
  mac: '',
  interface: 'lan1',
})

const peersModal = reactive({
  open: false,
  loading: false,
  busy: false,
  error: null as string | null,
  wgId: '',
  wgName: '',
  rows: [] as Array<Record<string, unknown>>,
  form: {
    peer_publickey: '',
    allowips: '10.9.0.2/32',
    endpoint: '',
    endpoint_port: 51820,
    comment: '',
  },
})

const RESOURCE_COPY_MAP: Record<string, { createLabel: string; noun: string }> = {
  'dhcp-services': { createLabel: '新建地址池', noun: '地址池' },
  'dhcp-static': { createLabel: '新增静态绑定', noun: '静态绑定' },
  vlan: { createLabel: '新建 VLAN', noun: 'VLAN' },
  'dmz-rules': { createLabel: '新增 DMZ 主机', noun: 'DMZ 主机' },
  'static-routes': { createLabel: '新建静态路由', noun: '静态路由' },
  terminals: { createLabel: '新增终端标注', noun: '终端标注' },
  'auth-accounts': { createLabel: '新增认证账号', noun: '认证账号' },
  'auth-packages': { createLabel: '新建套餐', noun: '套餐' },
  wireguard: { createLabel: '新建隧道', noun: '隧道' },
  'ipsec-clients': { createLabel: '新建站点隧道', noun: '站点隧道' },
  'openvpn-clients': { createLabel: '新建客户端', noun: '客户端连接' },
  'pptp-clients': { createLabel: '新建客户端', noun: '客户端连接' },
  'l2tp-clients': { createLabel: '新建客户端', noun: '客户端连接' },
  'ikev2-clients': { createLabel: '新建客户端', noun: '客户端连接' },
  'wireless-blacklist': { createLabel: '新增黑白名单', noun: '无线黑白名单' },
  'wireless-vlan': { createLabel: '新增 VLAN 绑定', noun: '无线 VLAN 绑定' },
  'ftp-users': { createLabel: '新增 FTP 账号', noun: 'FTP 账号' },
  'http-users': { createLabel: '新建 HTTP 站点', noun: 'HTTP 站点' },
  'samba-users': { createLabel: '新建共享目录', noun: '共享目录' },
  'reboot-schedules': { createLabel: '新建定时任务', noun: '定时任务' },
  'web-admin-groups': { createLabel: '新建权限组', noun: '权限组' },
  'web-admin-accounts': { createLabel: '新增管理员', noun: '管理员账号' },
}

const resourceCopy = computed(() => {
  const mapped = RESOURCE_COPY_MAP[props.resourceKey]
  if (mapped) return mapped
  if (spec.value?.group === 'objects') {
    return { createLabel: '新建对象组', noun: '对象组' }
  }
  return { createLabel: '新建规则', noun: '规则' }
})
const createButtonLabel = computed(() => resourceCopy.value.createLabel)
const itemNoun = computed(() => resourceCopy.value.noun)

const isUpstreamReadOnly = computed(() => {
  if (!spec.value) return false
  const hasWriteOp = (spec.value.operations || []).some((op) => ['create', 'update', 'toggle', 'delete'].includes(op))
  const hasActions = (spec.value.actions?.length ?? 0) > 0
  return !hasWriteOp && !hasActions && props.resourceKey !== 'dhcp-leases'
})

const canCreate = computed(() => spec.value?.operations?.includes('create') ?? false)
const canUpdate = computed(() => spec.value?.operations?.includes('update') ?? false)
const canToggle = computed(() => spec.value?.operations?.includes('toggle') ?? false)
const canDeleteRow = computed(
  () => (spec.value?.operations?.includes('delete') ?? false) && spec.value?.delete_mode !== 'clear'
)
const canClearAll = computed(
  () => (spec.value?.operations?.includes('delete') ?? false) && spec.value?.delete_mode === 'clear'
)

const displayColumns = computed<ColumnSpec[]>(() => {
  if (spec.value?.columns?.length) return spec.value.columns
  const first = rows.value[0]
  if (!first) return []
  return Object.keys(first)
    .slice(0, 7)
    .map((key) => ({ key, label: key }))
})

function rowId(row: Record<string, unknown>, idx: number): string {
  if (row.id !== undefined && row.id !== null) return String(row.id)
  if (row.filename !== undefined) return String(row.filename)
  if (row.f_name !== undefined) return String(row.f_name)
  return String(idx)
}

function isTruthy(val: unknown): boolean {
  if (typeof val === 'boolean') return val
  const s = String(val ?? '').toLowerCase()
  return s === 'yes' || s === '1' || s === 'true' || s === 'on'
}

function badgeTone(val: unknown): string {
  const s = String(val ?? '').toLowerCase()
  if (['yes', '1', 'true', 'ok', 'success', 'online', 'accept'].includes(s)) return 'v4-badge-up'
  if (['no', '0', 'drop', 'reject', 'fail', 'error', 'offline'].includes(s)) return 'v4-badge-down'
  return 'v4-badge-accent'
}

function formatCell(val: unknown, kind?: string, colKey?: string): string {
  if (val === null || val === undefined || val === '') return '—'
  if (typeof val === 'object') return formatCompoundValue(val)
  if (colKey === 'st_type') return String(val) === '0' ? '目录' : '文件'
  if (kind === 'badge' && (val === 1 || val === '1')) return '在线 / 正常'
  if (kind === 'badge' && (val === 0 || val === '0')) return '离线 / 关闭'
  return String(val)
}

async function reload() {
  const seq = ++reqSeq
  loading.value = true
  error.value = null
  spec.value = getSpec(props.resourceKey) || spec.value
  try {
    const res = await fetchV4Resource(props.resourceKey, {
      page: page.value,
      page_size: pageSize.value,
      search: search.value,
      path: props.resourceKey === 'system-files' ? filePath.value || '/' : undefined,
    })
    if (seq !== reqSeq) return
    loadFailed.value = false
    spec.value = res.resource
    rows.value = res.rows || []
    total.value = res.total || rows.value.length
  } catch (e: unknown) {
    if (seq !== reqSeq) return
    loadFailed.value = true
    spec.value = getSpec(props.resourceKey) || null
    rows.value = []
    total.value = 0
    error.value = extractError(e)
  } finally {
    if (seq === reqSeq) {
      loading.value = false
    }
  }
}

function goPath(p: string) {
  filePath.value = p || '/'
  void reload()
}

function goParentPath() {
  if (!filePath.value || filePath.value === '/') return
  const parts = filePath.value.split('/').filter(Boolean)
  parts.pop()
  filePath.value = parts.length ? '/' + parts.join('/') : '/'
  void reload()
}

function enterFolder(name: string) {
  if (!name) return
  const base = filePath.value === '/' ? '' : filePath.value.replace(/\/+$/, '')
  filePath.value = `${base}/${name}`
  void reload()
}

function changePage(next: number) {
  page.value = next
  void reload()
}

async function copyCLI() {
  if (!spec.value?.cli_command) return
  try {
    await navigator.clipboard.writeText(spec.value.cli_command)
    copiedCli.value = true
    setTimeout(() => {
      copiedCli.value = false
    }, 1800)
  } catch {
    // ignore clipboard errors
  }
}

async function copyInspectJson() {
  if (!inspectRow.value) return
  try {
    await navigator.clipboard.writeText(JSON.stringify(inspectRow.value, null, 2))
    showToast('已复制完整 JSON 到剪贴板')
  } catch {
    // ignore
  }
}

function openEditFromInspector() {
  if (!inspectRow.value) return
  const target = inspectRow.value
  inspectRow.value = null
  openEditDrawer(target)
}

async function handleToggle(row: Record<string, unknown>) {
  if (readOnly.value) return
  const id = String(row.id ?? '')
  if (!id) return
  const next: 'yes' | 'no' = isTruthy(row.enabled) ? 'no' : 'yes'
  busyId.value = id
  error.value = null
  try {
    await toggleV4Resource(props.resourceKey, id, next)
    row.enabled = next
    showToast(`${itemNoun.value} #${id} 已${next === 'yes' ? '启用' : '禁用'}`)
  } catch (e: unknown) {
    error.value = extractError(e)
  } finally {
    busyId.value = null
  }
}

function compoundObjectRefs(fieldKey: string): string[] {
  const val = editingRow.value?.[fieldKey]
  if (!val || typeof val !== 'object' || Array.isArray(val)) return []
  const objs = (val as { object?: unknown[] }).object
  return Array.isArray(objs) ? objs.map(formatObjectRef).filter(Boolean) : []
}

function openCreateDrawer() {
  editingId.value = null
  editingRow.value = null
  rawJsonMode.value = false
  drawerError.value = null
  dryRunPreview.value = null
  const init: Record<string, unknown> = {}
  for (const field of spec.value?.fields || []) {
    init[field.key] = field.default ?? ''
  }
  formState.value = init
  rawJsonText.value = JSON.stringify(buildRawPayloadFromForm(init), null, 2)
  drawerOpen.value = true
}

function mapRowToFormState(row: Record<string, unknown>): Record<string, unknown> {
  const init: Record<string, unknown> = {}
  for (const field of spec.value?.fields || []) {
    const val = row[field.key]
    if (field.type === 'csv' && val && typeof val === 'object') {
      init[field.key] = compoundToCSV(val)
    } else if (field.type === 'password') {
      init[field.key] = ''
    } else if (field.type === 'select' || field.type === 'switch') {
      init[field.key] = val !== undefined && val !== null ? String(val) : (field.default ?? '')
    } else {
      init[field.key] = val ?? field.default ?? ''
    }
  }
  return init
}

function openEditDrawer(row: Record<string, unknown>) {
  editingId.value = String(row.id ?? '')
  editingRow.value = row
  rawJsonMode.value = false
  drawerError.value = null
  dryRunPreview.value = null
  const init = mapRowToFormState(row)
  formState.value = init
  rawJsonText.value = JSON.stringify(buildRawPayloadFromForm(init), null, 2)
  drawerOpen.value = true
}

function buildRawPayloadFromForm(edits: Record<string, unknown>): Record<string, unknown> {
  const base: Record<string, unknown> = editingRow.value ? { ...editingRow.value } : {}
  if (editingId.value) {
    delete base.id
  }
  for (const field of spec.value?.fields || []) {
    const val = edits[field.key]
    if (field.type === 'password' && (val === '' || val === undefined)) {
      continue
    }
    const orig = editingRow.value?.[field.key]
    if (field.type === 'csv' && orig && typeof orig === 'object' && !Array.isArray(orig)) {
      const hasCompoundShape = 'custom' in orig || 'object' in orig
      if (hasCompoundShape) {
        const existingObj = Array.isArray((orig as { object?: unknown }).object)
          ? (orig as { object: unknown[] }).object
          : []
        const customList = String(val ?? '')
          .split(',')
          .map((s) => s.trim())
          .filter(Boolean)
        base[field.key] = { custom: customList, object: existingObj }
        continue
      }
    }
    if (val !== undefined) {
      base[field.key] = val
    }
  }
  const objKey = spec.value?.object_value_key
  if (objKey) {
    const parts = String(edits.value ?? '')
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
    base.group_value = parts.map((p) => ({ [objKey]: p }))
    delete base.value
  }
  if (props.resourceKey === 'time-objects') {
    const tVal = String(edits.type || 'weekly')
    const firstWindow: Record<string, unknown> = {
      type: tVal,
      start_time: String(edits.start_time || '00:00'),
      end_time: String(edits.end_time || '23:59'),
      comment: String(edits.val_comment ?? ''),
    }
    if (tVal === 'weekly') {
      firstWindow.weekdays = String(edits.weekdays || '1234567')
    }
    const prevGV = Array.isArray(editingRow.value?.group_value) ? editingRow.value.group_value : []
    base.group_value = [firstWindow, ...prevGV.slice(1)]
    for (const k of ['type', 'weekdays', 'start_time', 'end_time', 'val_comment', 'value']) {
      delete base[k]
    }
  }
  return base
}

function switchToRawJson() {
  rawJsonText.value = JSON.stringify(buildRawPayloadFromForm(formState.value), null, 2)
  rawJsonMode.value = true
}

function exitRawJson() {
  if (!rawJsonMode.value) return
  let parsed: unknown
  try {
    parsed = JSON.parse(rawJsonText.value)
  } catch {
    drawerError.value = 'JSON 格式无效，请先修正语法再切回可视化表单'
    return
  }
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
    drawerError.value = 'JSON 根节点必须是对象格式 ({ ... })'
    return
  }
  drawerError.value = null
  const prevRow = editingRow.value
  const parsedObj = parsed as Record<string, unknown>
  const mergedRow: Record<string, unknown> = prevRow ? { ...prevRow, ...parsedObj } : { ...parsedObj }
  const objKey = spec.value?.object_value_key
  if (objKey && Array.isArray(parsedObj.group_value) && parsedObj.value === undefined) {
    mergedRow.value = parsedObj.group_value
      .filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === 'object' && !Array.isArray(item))
      .map((item) => String(item[objKey] ?? item.value ?? '').trim())
      .filter(Boolean)
      .join(',')
  }
  if (props.resourceKey === 'time-objects' && Array.isArray(parsedObj.group_value) && parsedObj.group_value.length > 0) {
    const firstRaw = parsedObj.group_value[0]
    if (firstRaw && typeof firstRaw === 'object' && !Array.isArray(firstRaw)) {
      const first = firstRaw as Record<string, unknown>
      if (parsedObj.type === undefined && first.type !== undefined) {
        mergedRow.type = String(first.type)
      }
      if (parsedObj.weekdays === undefined) {
        mergedRow.weekdays = first.weekdays !== undefined ? String(first.weekdays) : ''
      }
      if (parsedObj.start_time === undefined && first.start_time !== undefined) {
        mergedRow.start_time = String(first.start_time)
      }
      if (parsedObj.end_time === undefined && first.end_time !== undefined) {
        mergedRow.end_time = String(first.end_time)
      }
      if (parsedObj.val_comment === undefined) {
        mergedRow.val_comment = first.comment !== undefined ? String(first.comment) : ''
      }
    }
  }
  if (prevRow) {
    editingRow.value = mergedRow
  }
  const nextForm = mapRowToFormState(mergedRow)
  for (const field of spec.value?.fields || []) {
    if (
      field.type === 'password' &&
      typeof parsedObj[field.key] === 'string' &&
      parsedObj[field.key] !== '' &&
      parsedObj[field.key] !== prevRow?.[field.key]
    ) {
      nextForm[field.key] = parsedObj[field.key]
    }
  }
  formState.value = nextForm
  rawJsonMode.value = false
}

async function generateWGKeys() {
  generatingKeys.value = true
  drawerError.value = null
  try {
    const pair = await generateWGKeypair()
    formState.value.local_privatekey = pair.private_key
    formState.value.local_publickey = pair.public_key
    if (rawJsonMode.value) {
      rawJsonText.value = JSON.stringify(buildRawPayloadFromForm(formState.value), null, 2)
    }
  } catch (e: unknown) {
    drawerError.value = extractError(e)
  } finally {
    generatingKeys.value = false
  }
}

async function submitForm(dryRun: boolean) {
  if (!dryRun && readOnly.value) {
    drawerError.value = '当前处于只读保护模式，请使用 Dry-Run 预览或解除写保护'
    return
  }
  drawerError.value = null
  let payload: Record<string, unknown>
  if (rawJsonMode.value) {
    let parsed: unknown
    try {
      parsed = JSON.parse(rawJsonText.value)
    } catch {
      drawerError.value = 'JSON 格式无效，请检查语法'
      return
    }
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
      drawerError.value = 'JSON 根节点必须是对象格式 ({ ... })'
      return
    }
    payload = parsed as Record<string, unknown>
  } else {
    payload = { ...formState.value }
    for (const field of spec.value?.fields || []) {
      const val = payload[field.key]
      if (field.type === 'password' && editingId.value && val === '') {
        delete payload[field.key]
      } else if (field.type === 'number' && (val === '' || val === null || Number.isNaN(val))) {
        delete payload[field.key]
      }
    }
  }

  submitting.value = true
  try {
    const res = editingId.value
      ? await updateV4Resource(props.resourceKey, editingId.value, payload, dryRun)
      : await createV4Resource(props.resourceKey, payload, dryRun)

    if (dryRun) {
      dryRunPreview.value = res
    } else {
      drawerOpen.value = false
      showToast(editingId.value ? `${itemNoun.value}已更新` : `${itemNoun.value}已创建`)
      await reload()
    }
  } catch (e: unknown) {
    drawerError.value = extractError(e)
  } finally {
    submitting.value = false
  }
}

function askDeleteRow(row: Record<string, unknown>) {
  if (readOnly.value) return
  if (spec.value?.delete_mode === 'srcfile') {
    const filename = String(row.filename ?? '')
    deleteTarget.value = { srcfile: filename, label: `确认永久删除备份文件 ${filename} 吗？` }
    return
  }
  const id = String(row.id ?? '')
  const name = String(row.tagname ?? row.name ?? row.username ?? row.group_name ?? id)
  deleteTarget.value = { id, label: `确认删除 "${name}" (#${id}) 吗？此操作会立即同步到路由器。` }
}

function confirmClearAll() {
  if (readOnly.value) return
  deleteTarget.value = { label: `确认清空所有 ${spec.value?.label || ''} 记录吗？此操作不可撤销。` }
}

async function executeDelete() {
  if (!deleteTarget.value || readOnly.value) return
  submitting.value = true
  error.value = null
  try {
    await deleteV4Resource(props.resourceKey, deleteTarget.value.id, false, deleteTarget.value.srcfile)
    deleteTarget.value = null
    showToast('删除操作已完成')
    await reload()
  } catch (e: unknown) {
    error.value = extractError(e)
    deleteTarget.value = null
  } finally {
    submitting.value = false
  }
}

async function handleAction(act: ActionSpec) {
  if (readOnly.value) return
  if (act.dangerous && !window.confirm(act.confirm_text || `确认执行 ${act.label} 吗？`)) return
  actionBusy.value = true
  error.value = null
  try {
    const res = await executeV4Action(act.key, false)
    showToast(String(res.message || `动作「${act.label}」执行成功`))
    await reload()
  } catch (e: unknown) {
    error.value = extractError(e)
  } finally {
    actionBusy.value = false
  }
}

async function openRefsModal(groupName: string) {
  refsModal.open = true
  refsModal.loading = true
  refsModal.groupName = groupName
  refsModal.rows = []
  refsModal.error = null
  try {
    const res = await fetchObjectRefs(props.resourceKey, groupName)
    refsModal.rows = res.rows || []
  } catch (e: unknown) {
    refsModal.error = extractError(e)
  } finally {
    refsModal.loading = false
  }
}

function openBindStaticModal(row: Record<string, unknown>) {
  bindStatic.open = true
  bindStatic.error = null
  bindStatic.tagname = String(row.hostname || 'StaticDevice')
  bindStatic.ip_addr = String(row.ip_addr || '')
  bindStatic.mac = String(row.mac || '')
  bindStatic.interface = String(row.interface || 'lan1')
}

async function submitBindStatic() {
  bindStatic.busy = true
  bindStatic.error = null
  try {
    await createV4Resource('dhcp-static', {
      tagname: bindStatic.tagname,
      ip_addr: bindStatic.ip_addr,
      mac: bindStatic.mac,
      interface: bindStatic.interface,
      enabled: 'yes',
    })
    bindStatic.open = false
    showToast(`已将 ${bindStatic.ip_addr} 加入 DHCP 静态绑定`)
  } catch (e: unknown) {
    bindStatic.error = extractError(e)
  } finally {
    bindStatic.busy = false
  }
}

async function openPeersModal(row: Record<string, unknown>) {
  peersModal.open = true
  peersModal.wgId = String(row.id ?? '')
  peersModal.wgName = String(row.name ?? peersModal.wgId)
  peersModal.error = null
  await reloadPeers()
}

async function reloadPeers() {
  peersModal.loading = true
  peersModal.error = null
  try {
    const res = await fetchWGPeers(peersModal.wgId)
    peersModal.rows = res.rows || []
  } catch (e: unknown) {
    peersModal.error = extractError(e)
  } finally {
    peersModal.loading = false
  }
}

async function addPeer() {
  if (!peersModal.form.peer_publickey || !peersModal.form.allowips) {
    peersModal.error = '请填写对端公钥和允许子网'
    return
  }
  peersModal.busy = true
  peersModal.error = null
  try {
    await createWGPeer(peersModal.wgId, {
      ...peersModal.form,
      interface: peersModal.wgName,
      enabled: 'yes',
    })
    peersModal.form.peer_publickey = ''
    await reloadPeers()
  } catch (e: unknown) {
    peersModal.error = extractError(e)
  } finally {
    peersModal.busy = false
  }
}

async function flipPeer(peer: Record<string, unknown>) {
  try {
    const next = isTruthy(peer.enabled) ? 'no' : 'yes'
    await toggleWGPeer(peersModal.wgId, String(peer.id), next)
    peer.enabled = next
  } catch (e: unknown) {
    peersModal.error = extractError(e)
  }
}

async function removePeer(peer: Record<string, unknown>) {
  try {
    await deleteWGPeer(peersModal.wgId, String(peer.id))
    await reloadPeers()
  } catch (e: unknown) {
    peersModal.error = extractError(e)
  }
}

function showToast(msg: string) {
  toast.value = msg
  setTimeout(() => {
    if (toast.value === msg) toast.value = null
  }, 2600)
}

function extractError(e: unknown): string {
  if (e && typeof e === 'object' && 'response' in e) {
    const resp = (e as { response?: { data?: { message?: string } } }).response
    if (resp?.data?.message) return resp.data.message
  }
  return e instanceof Error ? e.message : String(e)
}

watch(
  [() => props.resourceKey, routerEpoch],
  () => {
    page.value = 1
    search.value = ''
    filePath.value = '/'
    spec.value = getSpec(props.resourceKey) || null
    rows.value = []
    total.value = 0
    loadFailed.value = false
    error.value = null
    toast.value = null
    drawerOpen.value = false
    inspectRow.value = null
    deleteTarget.value = null
    refsModal.open = false
    peersModal.open = false
    bindStatic.open = false
    void reload()
  },
  { immediate: true }
)
</script>

<style scoped>
.resource-panel {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.panel-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.title-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.title-row h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 750;
  color: var(--text-strong);
}

.desc {
  margin: 5px 0 0;
  font-size: 12px;
  color: var(--text-muted);
}

.head-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.cli-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-height: 32px;
  padding: 0 11px;
  max-width: 100%;
  overflow: hidden;
  border-radius: var(--radius-control);
  border: 1px solid var(--border-control);
  background: var(--surface-btn);
  color: var(--text-normal);
  font-size: 11px;
  cursor: pointer;
}

.cli-pill span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cli-pill:hover {
  border-color: var(--border-strong);
  color: var(--text-strong);
}

.ok-icon {
  color: var(--status-up);
}

.path-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.path-input {
  flex: 1;
  min-width: 180px;
  min-height: 32px;
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.search-box {
  position: relative;
  flex: 1;
  min-width: 220px;
  max-width: 380px;
}

.search-ic {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-muted);
}

.search-input {
  padding-left: 34px;
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
  max-width: 260px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.v4-table td.actions-col {
  max-width: none;
  overflow: visible;
}

.v4-table tbody tr:last-child td {
  border-bottom: none;
}

.v4-table tbody tr:hover {
  background: var(--surface-input);
}

.actions-col {
  text-align: right !important;
  white-space: nowrap;
}

.row-actions {
  display: inline-flex;
  align-items: center;
  gap: 6px;
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

.compound-chip {
  font-size: 12px;
  color: var(--text-normal);
}

.empty-state {
  padding: 36px 16px;
  text-align: center;
  color: var(--text-muted);
}

.pagination-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 12px;
  color: var(--text-muted);
}

.pager-btns {
  display: flex;
  gap: 8px;
}

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

.drawer-backdrop.center {
  align-items: center;
  justify-content: center;
}

.drawer-panel {
  width: min(540px, 100vw);
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

.wide-drawer {
  width: min(680px, 100vw);
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

.mode-switch {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
}

.form-grid.two-cols {
  grid-template-columns: repeat(2, minmax(0, 1fr));
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

.field-label em {
  color: var(--status-down);
  font-style: normal;
}

.dryrun-box {
  padding: 12px;
  border-radius: var(--radius-md);
  border: 1px solid var(--accent-primary);
  background: var(--surface-input);
}

.dryrun-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 12px;
}

.dryrun-box pre,
.json-pre {
  margin: 0;
  font-size: 11.5px;
  line-height: 1.55;
  overflow-x: auto;
  color: var(--text-normal);
  padding: 12px;
  border-radius: var(--radius-md);
  background: var(--surface-input);
  border: 1px solid var(--divider-line);
}

.drawer-foot {
  padding: 14px 20px;
  border-top: 1px solid var(--divider-line);
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.confirm-dialog {
  width: min(440px, calc(100vw - 32px));
  padding: 20px;
  background: var(--surface-popover);
  border-color: var(--chrome-border);
  box-shadow: var(--chrome-shadow);
  -webkit-backdrop-filter: var(--backdrop-chrome);
  backdrop-filter: var(--backdrop-chrome);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.modal-wide {
  width: min(560px, calc(100vw - 32px));
}

.modal-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.confirm-dialog h4,
.modal-top h4 {
  margin: 0;
  font-size: 15px;
  color: var(--text-strong);
}

.confirm-dialog p {
  margin: 0;
  color: var(--text-normal);
  font-size: 13px;
}

.confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 4px;
}

.peer-form {
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.spin {
  animation: spin 0.9s linear infinite;
}

.obj-ref-hint {
  font-size: 11px;
  color: var(--accent-text);
  margin-top: 2px;
}

@media (max-width: 640px) {
  .resource-panel {
    padding: 14px;
    gap: 12px;
  }

  .head-actions {
    width: 100%;
  }

  .search-box {
    min-width: 100%;
    max-width: none;
  }

  .form-grid.two-cols {
    grid-template-columns: 1fr;
  }

  .pagination-bar {
    flex-wrap: wrap;
    gap: 8px;
  }

  .drawer-backdrop {
    align-items: flex-end;
  }

  .drawer-panel,
  .wide-drawer {
    width: 100%;
    height: auto;
    max-height: 88dvh;
    border-radius: var(--radius-xl) var(--radius-xl) 0 0;
  }

  .drawer-body {
    padding: 14px 16px;
  }

  .drawer-foot {
    padding: 12px 16px max(12px, env(safe-area-inset-bottom));
  }

  .confirm-dialog,
  .modal-wide {
    width: calc(100vw - 20px);
    max-height: 86dvh;
    overflow-y: auto;
    padding: 16px;
  }
}
</style>

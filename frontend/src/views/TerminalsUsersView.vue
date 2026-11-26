<template>
  <section class="terminals-page">
    <header class="ws-head v4-panel">
      <div>
        <span class="ws-eyebrow font-mono">TERMINALS & AUTH USERS</span>
        <h1>终端设备与认证用户</h1>
        <p>实时查看局域网双栈终端速率，支持终端标注、静态 DHCP、一键限速、MAC 阻断与 Web 认证服务管理。</p>
      </div>
    </header>

    <div ref="subtabBarEl" class="subtab-bar" role="tablist">
      <button
        v-for="tab in subTabs"
        :key="tab.key"
        type="button"
        class="subtab-pill"
        :class="{ active: activeTab === tab.key }"
        @click="selectSubTab(tab.key)"
      >
        {{ tab.label }}
      </button>
    </div>

    <div v-if="actionToast" class="toast-banner">{{ actionToast }}</div>

    <!-- Client Modes -->
    <template v-if="isClientMode">
      <div class="v4-card clients-card">
        <div class="toolbar">
          <div class="search-box">
            <Search :size="14" class="search-ic" />
            <input
              v-model.trim="search"
              type="text"
              class="v4-input search-input"
              placeholder="搜索主机名、IP、MAC 或备注..."
              @keyup.enter="loadClients"
            />
          </div>

          <div class="toolbar-right">
            <select v-model="selectedClientSortKey" class="v4-select sort-select" aria-label="排序方式">
              <option value="download_speed">按实时下行排序</option>
              <option value="upload_speed">按实时上行排序</option>
              <option value="total_down">按累计下行排序</option>
              <option value="total_up">按累计上行排序</option>
              <option value="connections">按连接数排序</option>
              <option value="uptime">按在线时长排序</option>
              <option value="ip_addr">按 IP 地址排序</option>
              <option value="mac">按 MAC 地址排序</option>
              <option value="name">按终端名称排序</option>
            </select>
            <button
              type="button"
              class="v4-btn sort-dir-btn"
              :title="clientSort.sortDir.value === 'asc' ? '当前升序，点击切换降序' : '当前降序，点击切换升序'"
              @click="toggleClientSortDirection"
            >
              <ArrowUp v-if="clientSort.sortDir.value === 'asc'" :size="14" />
              <ArrowDown v-else :size="14" />
              <span>{{ clientSort.sortDir.value === 'asc' ? '升序' : '降序' }}</span>
            </button>
            <button type="button" class="v4-btn" :disabled="loading" @click="loadClients">
              <RefreshCw :size="14" :class="{ spin: loading }" />
              <span>刷新</span>
            </button>
          </div>
        </div>

        <div v-if="error" class="error-banner">{{ error }}</div>

        <div class="table-wrap">
          <div v-if="loading && sortedClients.length === 0" class="loading-state">
            <div class="spinner" />
            <span class="font-mono">正在读取终端列表...</span>
          </div>
          <div v-else-if="sortedClients.length === 0" class="empty-state">暂无匹配终端记录</div>
          <table v-else class="v4-table">
            <thead>
              <tr>
                <th
                  class="sortable-th"
                  :class="{ 'is-sorted': clientSort.sortKey.value === 'name' }"
                  :aria-sort="clientSort.ariaSort('name')"
                  tabindex="0"
                  @click="clientSort.toggleSort('name')"
                  @keydown.enter.prevent="clientSort.toggleSort('name')"
                  @keydown.space.prevent="clientSort.toggleSort('name')"
                >
                  <span class="sort-th-inner">
                    <span>终端名称 / 备注</span>
                    <SortIndicator :active="clientSort.sortKey.value === 'name'" :dir="clientSort.sortDir.value" />
                  </span>
                </th>
                <th
                  class="sortable-th"
                  :class="{ 'is-sorted': clientSort.sortKey.value === 'ip_addr' }"
                  :aria-sort="clientSort.ariaSort('ip_addr')"
                  tabindex="0"
                  @click="clientSort.toggleSort('ip_addr')"
                  @keydown.enter.prevent="clientSort.toggleSort('ip_addr')"
                  @keydown.space.prevent="clientSort.toggleSort('ip_addr')"
                >
                  <span class="sort-th-inner">
                    <span>IP 地址</span>
                    <SortIndicator :active="clientSort.sortKey.value === 'ip_addr'" :dir="clientSort.sortDir.value" />
                  </span>
                </th>
                <th
                  class="sortable-th"
                  :class="{ 'is-sorted': clientSort.sortKey.value === 'mac' }"
                  :aria-sort="clientSort.ariaSort('mac')"
                  tabindex="0"
                  @click="clientSort.toggleSort('mac')"
                  @keydown.enter.prevent="clientSort.toggleSort('mac')"
                  @keydown.space.prevent="clientSort.toggleSort('mac')"
                >
                  <span class="sort-th-inner">
                    <span>MAC 地址</span>
                    <SortIndicator :active="clientSort.sortKey.value === 'mac'" :dir="clientSort.sortDir.value" />
                  </span>
                </th>
                <th
                  :aria-sort="
                    clientSort.sortKey.value === 'upload_speed' || clientSort.sortKey.value === 'download_speed'
                      ? clientSort.sortDir.value === 'asc'
                        ? 'ascending'
                        : 'descending'
                      : 'none'
                  "
                >
                  <div class="dual-sort-th">
                    <span>实时</span>
                    <button
                      type="button"
                      class="sort-sub-btn"
                      :class="{ active: clientSort.sortKey.value === 'upload_speed' }"
                      @click="clientSort.toggleSort('upload_speed')"
                    >
                      <span>上行</span>
                      <SortIndicator :active="clientSort.sortKey.value === 'upload_speed'" :dir="clientSort.sortDir.value" />
                    </button>
                    <span>/</span>
                    <button
                      type="button"
                      class="sort-sub-btn"
                      :class="{ active: clientSort.sortKey.value === 'download_speed' }"
                      @click="clientSort.toggleSort('download_speed')"
                    >
                      <span>下行</span>
                      <SortIndicator :active="clientSort.sortKey.value === 'download_speed'" :dir="clientSort.sortDir.value" />
                    </button>
                  </div>
                </th>
                <th
                  :aria-sort="
                    clientSort.sortKey.value === 'total_up' || clientSort.sortKey.value === 'total_down'
                      ? clientSort.sortDir.value === 'asc'
                        ? 'ascending'
                        : 'descending'
                      : 'none'
                  "
                >
                  <div class="dual-sort-th">
                    <span>累计</span>
                    <button
                      type="button"
                      class="sort-sub-btn"
                      :class="{ active: clientSort.sortKey.value === 'total_up' }"
                      @click="clientSort.toggleSort('total_up')"
                    >
                      <span>上行</span>
                      <SortIndicator :active="clientSort.sortKey.value === 'total_up'" :dir="clientSort.sortDir.value" />
                    </button>
                    <span>/</span>
                    <button
                      type="button"
                      class="sort-sub-btn"
                      :class="{ active: clientSort.sortKey.value === 'total_down' }"
                      @click="clientSort.toggleSort('total_down')"
                    >
                      <span>下行</span>
                      <SortIndicator :active="clientSort.sortKey.value === 'total_down'" :dir="clientSort.sortDir.value" />
                    </button>
                  </div>
                </th>
                <th
                  class="sortable-th"
                  :class="{ 'is-sorted': clientSort.sortKey.value === 'connections' }"
                  :aria-sort="clientSort.ariaSort('connections')"
                  tabindex="0"
                  @click="clientSort.toggleSort('connections')"
                  @keydown.enter.prevent="clientSort.toggleSort('connections')"
                  @keydown.space.prevent="clientSort.toggleSort('connections')"
                >
                  <span class="sort-th-inner">
                    <span>连接数</span>
                    <SortIndicator :active="clientSort.sortKey.value === 'connections'" :dir="clientSort.sortDir.value" />
                  </span>
                </th>
                <th
                  class="sortable-th"
                  :class="{ 'is-sorted': clientSort.sortKey.value === 'uptime' }"
                  :aria-sort="clientSort.ariaSort('uptime')"
                  tabindex="0"
                  @click="clientSort.toggleSort('uptime')"
                  @keydown.enter.prevent="clientSort.toggleSort('uptime')"
                  @keydown.space.prevent="clientSort.toggleSort('uptime')"
                >
                  <span class="sort-th-inner">
                    <span>在线时长</span>
                    <SortIndicator :active="clientSort.sortKey.value === 'uptime'" :dir="clientSort.sortDir.value" />
                  </span>
                </th>
                <th>快捷管控与画像</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="c in sortedClients" :key="c.mac + c.ip_addr">
                <td>
                  <div class="host-cell">
                    <strong>{{ c.comment || c.hostname || '未命名终端' }}</strong>
                    <span v-if="c.comment && c.hostname" class="sub-muted font-mono">{{ c.hostname }}</span>
                  </div>
                </td>
                <td class="font-mono">{{ c.ip_addr || '—' }}</td>
                <td class="font-mono">{{ c.mac || '—' }}</td>
                <td class="font-mono">
                  <span class="ic-up">↑ {{ formatSpeed(c.upload_speed) }}</span>
                  <span class="sep">/</span>
                  <span class="ic-down">↓ {{ formatSpeed(c.download_speed) }}</span>
                </td>
                <td class="font-mono">{{ formatBytes(c.total_up) }} / {{ formatBytes(c.total_down) }}</td>
                <td class="font-mono">{{ c.connections }}</td>
                <td class="font-mono">{{ c.uptime || c.downtime || '—' }}</td>
                <td>
                  <div class="client-actions">
                    <button
                      v-if="c.ip_addr && c.mac"
                      type="button"
                      class="v4-btn btn-xs"
                      @click="openDetail(c)"
                    >
                      <BarChart3 :size="13" />
                      <span>画像</span>
                    </button>
                    <button
                      v-if="c.mac"
                      type="button"
                      class="v4-btn btn-xs"
                      :disabled="readOnly"
                      title="为该 MAC 添加终端备注"
                      @click="openQuickModal('terminal', c)"
                    >
                      <Tag :size="12" />
                      <span>标注</span>
                    </button>
                    <button
                      v-if="c.ip_addr && c.mac"
                      type="button"
                      class="v4-btn btn-xs"
                      :disabled="readOnly"
                      title="转为 DHCP 静态绑定"
                      @click="openQuickModal('static', c)"
                    >
                      <Pin :size="12" />
                      <span>静态</span>
                    </button>
                    <button
                      v-if="c.ip_addr"
                      type="button"
                      class="v4-btn btn-xs"
                      :disabled="readOnly"
                      title="设置 IP 限速"
                      @click="openQuickModal('qos', c)"
                    >
                      <Gauge :size="12" />
                      <span>限速</span>
                    </button>
                    <button
                      v-if="c.mac"
                      type="button"
                      class="v4-btn v4-btn-danger btn-xs"
                      :disabled="readOnly"
                      title="添加到 MAC 访问控制"
                      @click="openQuickModal('mac-block', c)"
                    >
                      <Ban :size="12" />
                      <span>阻断</span>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>

    <!-- Auth Online Sessions with 1-Click Kick -->
    <template v-else-if="activeTab === 'auth-online'">
      <div class="v4-card clients-card">
        <div class="toolbar">
          <span class="v4-badge font-mono">ikuai-cli users online</span>
          <button type="button" class="v4-btn" @click="reloadAuthOnline">
            <RefreshCw :size="14" />
            <span>刷新会话</span>
          </button>
        </div>
        <div v-if="error" class="error-banner">{{ error }}</div>
        <div class="table-wrap">
          <div v-if="authRows.length === 0" class="empty-state">当前无在线认证会话</div>
          <table v-else class="v4-table">
            <thead>
              <tr>
                <th
                  v-for="col in authColumns"
                  :key="col.key"
                  class="sortable-th"
                  :class="{ 'is-sorted': authSort.sortKey.value === col.key }"
                  :aria-sort="authSort.ariaSort(col.key)"
                  tabindex="0"
                  @click="authSort.toggleSort(col.key)"
                  @keydown.enter.prevent="authSort.toggleSort(col.key)"
                  @keydown.space.prevent="authSort.toggleSort(col.key)"
                >
                  <span class="sort-th-inner">
                    <span>{{ col.label }}</span>
                    <SortIndicator :active="authSort.sortKey.value === col.key" :dir="authSort.sortDir.value" />
                  </span>
                </th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in sortedAuthRows" :key="String(row.id)">
                <td class="font-mono">{{ row.id }}</td>
                <td><strong>{{ row.username }}</strong></td>
                <td><span class="v4-badge">{{ row.ppptype }}</span></td>
                <td class="font-mono">{{ row.ip_addr }}</td>
                <td class="font-mono">{{ row.mac }}</td>
                <td><span class="v4-badge v4-badge-accent">{{ row.interface }}</span></td>
                <td class="font-mono">{{ formatTimestamp(row.auth_time) }}</td>
                <td>
                  <button
                    type="button"
                    class="v4-btn v4-btn-danger btn-xs"
                    :disabled="readOnly"
                    @click="kickUser(row)"
                  >
                    <UserX :size="13" />
                    <span>踢下线</span>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>

    <SingletonConfigPanel v-else-if="activeTab === 'auth-web-server'" resourceKey="auth-web-server" />

    <!-- Accounts & Packages via ResourceCollectionPanel -->
    <ResourceCollectionPanel v-else :key="activeTab" :resourceKey="activeTab" />

    <!-- Client Detail Drawer -->
    <teleport to="body">
      <div v-if="selectedClient" class="drawer-backdrop" @click.self="closeDetail">
        <div class="drawer-panel v4-panel">
          <div class="drawer-head">
            <div>
              <strong>{{ selectedClient.comment || selectedClient.hostname || selectedClient.ip_addr }}</strong>
              <span class="font-mono">{{ selectedClient.ip_addr }} · {{ selectedClient.mac }}</span>
            </div>
            <button type="button" class="v4-btn btn-xs" @click="closeDetail">
              <X :size="15" />
            </button>
          </div>
          <div class="drawer-body">
            <div v-if="detailLoading" class="loading-state"><div class="spinner" /></div>
            <div v-else-if="detailError" class="error-banner">{{ detailError }}</div>
            <template v-else-if="clientDetail">
              <div class="detail-section">
                <h4>5 分钟速率趋势</h4>
                <div ref="detailChartEl" class="detail-chart" />
              </div>
              <div class="detail-section">
                <h4>Top L7 应用协议分布</h4>
                <div v-if="clientDetail.app_protocols.length === 0" class="empty-state">暂无应用层统计数据</div>
                <div v-else class="app-list">
                  <div v-for="(app, idx) in clientDetail.app_protocols" :key="idx" class="app-row">
                    <strong>{{ String(app.appname ?? app.app_name ?? '未知应用') }}</strong>
                    <span class="font-mono">↓ {{ formatBytes(Number(app.total_down ?? 0)) }} · ↑ {{ formatBytes(Number(app.total_up ?? 0)) }}</span>
                  </div>
                </div>
              </div>
            </template>
          </div>
        </div>
      </div>
    </teleport>

    <!-- Quick Control Modal -->
    <teleport to="body">
      <div v-if="quickModal.open" class="drawer-backdrop center" @click.self="quickModal.open = false">
        <div class="confirm-dialog v4-panel">
          <div class="modal-top">
            <h4>{{ quickModalTitle }}</h4>
            <button type="button" class="v4-btn btn-xs" @click="quickModal.open = false"><X :size="14" /></button>
          </div>
          <div v-if="quickModal.error" class="error-banner">{{ quickModal.error }}</div>
          <div class="modal-form">
            <label class="field-item">
              <span>名称 / 标签</span>
              <input v-model.trim="quickModal.tagname" class="v4-input" />
            </label>
            <label v-if="quickModal.kind === 'static' || quickModal.kind === 'qos'" class="field-item">
              <span>IP 地址</span>
              <input v-model.trim="quickModal.ip_addr" class="v4-input font-mono" />
            </label>
            <label v-if="quickModal.kind !== 'qos'" class="field-item">
              <span>MAC 地址</span>
              <input v-model.trim="quickModal.mac" class="v4-input font-mono" />
            </label>
            <div v-if="quickModal.kind === 'static' || quickModal.kind === 'qos'" class="field-item">
              <span>生效接口</span>
              <InterfaceSelect
                v-model="quickModal.iface"
                :scope="quickModal.kind === 'qos' ? 'wan-all' : 'lan'"
              />
            </div>
            <template v-if="quickModal.kind === 'qos'">
              <label class="field-item">
                <span>上行限速 (Kbps)</span>
                <input v-model.number="quickModal.upload" type="number" class="v4-input font-mono" />
              </label>
              <label class="field-item">
                <span>下行限速 (Kbps)</span>
                <input v-model.number="quickModal.download" type="number" class="v4-input font-mono" />
              </label>
            </template>
            <label v-if="quickModal.kind === 'mac-block'" class="field-item">
              <span>规则状态</span>
              <select v-model="quickModal.enabled" class="v4-select">
                <option value="yes">立即生效 (enabled=yes)</option>
                <option value="no">先创建暂不生效 (enabled=no)</option>
              </select>
            </label>
            <label v-if="quickModal.kind !== 'qos'" class="field-item">
              <span>备注</span>
              <input v-model.trim="quickModal.comment" class="v4-input" />
            </label>
          </div>
          <div class="confirm-actions">
            <button type="button" class="v4-btn" @click="quickModal.open = false">取消</button>
            <button type="button" class="v4-btn v4-btn-primary" :disabled="quickModal.busy" @click="submitQuickModal">
              {{ quickModal.busy ? '执行中...' : '确认应用' }}
            </button>
          </div>
        </div>
      </div>
    </teleport>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onActivated, onDeactivated, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { echarts } from '@/lib/echarts'
import { ArrowDown, ArrowUp, Ban, BarChart3, Gauge, Pin, RefreshCw, Search, Tag, UserX, X } from 'lucide-vue-next'
import ResourceCollectionPanel from '@/components/common/ResourceCollectionPanel.vue'
import SingletonConfigPanel from '@/components/common/SingletonConfigPanel.vue'
import InterfaceSelect from '@/components/common/InterfaceSelect.vue'
import SortIndicator from '@/components/common/SortIndicator.vue'
import { useTableSort, type SortValueHint } from '@/composables/useTableSort'
import {
  createV4Resource,
  fetchV4ClientDetail,
  fetchV4Clients,
  fetchV4Resource,
  kickV4OnlineUser,
  type ClientDetailDTO,
  type ClientRecordDTO,
} from '@/api/v4'
import { useAppRoute } from '@/composables/useAppRoute'
import { useAuth } from '@/composables/useAuth'
import { useFleet } from '@/composables/useFleet'
import { useChartTheme } from '@/composables/useChartTheme'
import { formatBytes, formatSpeed, formatTimestamp } from '@/composables/useFormatters'

const props = defineProps<{
  initialSubResource?: string
  navSeq?: number
}>()

const { currentTab, syncSubTab } = useAppRoute()
const { readOnly } = useAuth()
const { routerEpoch } = useFleet()
const { palette } = useChartTheme()

const subTabs = [
  { key: 'online', label: 'IPv4 在线终端' },
  { key: 'ipv6-online', label: 'IPv6 在线终端' },
  { key: 'offline', label: '历史离线终端' },
  { key: 'summary', label: '终端流量汇总' },
  { key: 'auth-online', label: '在线认证会话' },
  { key: 'auth-accounts', label: '认证账号管理' },
  { key: 'auth-packages', label: '计费带宽套餐' },
  { key: 'auth-web-server', label: 'Web 认证服务' },
]

const initialTab =
  props.initialSubResource && subTabs.some((t) => t.key === props.initialSubResource)
    ? props.initialSubResource
    : 'online'
const activeTab = ref(initialTab)

if (currentTab.value === 'terminals' && !props.initialSubResource) {
  syncSubTab('terminals', activeTab.value, { replace: true })
}

function selectSubTab(key: string) {
  if (!subTabs.some((t) => t.key === key)) return
  activeTab.value = key
  syncSubTab('terminals', key, { replace: false })
}
const subtabBarEl = ref<HTMLElement | null>(null)
const search = ref('')
type ClientSortKey =
  | 'name'
  | 'ip_addr'
  | 'mac'
  | 'upload_speed'
  | 'download_speed'
  | 'total_up'
  | 'total_down'
  | 'connections'
  | 'uptime'

const TEXT_ASC_CLIENT_KEYS = new Set<ClientSortKey>(['name', 'ip_addr', 'mac'])
const clientSort = useTableSort<ClientSortKey>({
  initialKey: 'download_speed',
  initialDir: 'desc',
  allowUnsorted: false,
  defaultDirForKey: (key) => (TEXT_ASC_CLIENT_KEYS.has(key) ? 'asc' : 'desc'),
})

const selectedClientSortKey = computed<ClientSortKey>({
  get: () => clientSort.sortKey.value || 'download_speed',
  set: (key) => {
    const defaultDir = TEXT_ASC_CLIENT_KEYS.has(key) ? 'asc' : 'desc'
    clientSort.setSort(key, defaultDir)
  },
})

function toggleClientSortDirection() {
  const key = selectedClientSortKey.value
  clientSort.setSort(key, clientSort.sortDir.value === 'asc' ? 'desc' : 'asc')
}

type AuthSortKey = 'id' | 'username' | 'ppptype' | 'ip_addr' | 'mac' | 'interface' | 'auth_time'
const authColumns: Array<{ key: AuthSortKey; label: string; hint: SortValueHint }> = [
  { key: 'id', label: 'ID', hint: 'number' },
  { key: 'username', label: '账号', hint: 'auto' },
  { key: 'ppptype', label: '认证类型', hint: 'auto' },
  { key: 'ip_addr', label: 'IP 地址', hint: 'ip' },
  { key: 'mac', label: 'MAC', hint: 'auto' },
  { key: 'interface', label: '接口', hint: 'auto' },
  { key: 'auth_time', label: '上线时间', hint: 'number' },
]

const authSort = useTableSort<AuthSortKey>({
  defaultDirForKey: (key) => (key === 'auth_time' ? 'desc' : 'asc'),
})
const loading = ref(false)
const error = ref<string | null>(null)
const actionToast = ref<string | null>(null)
const clients = ref<ClientRecordDTO[]>([])
const authRows = ref<Array<Record<string, unknown>>>([])

const selectedClient = ref<ClientRecordDTO | null>(null)
const detailLoading = ref(false)
const detailError = ref<string | null>(null)
const clientDetail = ref<ClientDetailDTO | null>(null)
const detailChartEl = ref<HTMLElement | null>(null)
let detailChart: echarts.ECharts | null = null
let pollTimer: ReturnType<typeof setInterval> | null = null
let clientReqSeq = 0
let activeClientReqs = 0
let authReqSeq = 0
let detailReqSeq = 0

const quickModal = reactive({
  open: false,
  busy: false,
  error: null as string | null,
  kind: 'terminal' as 'terminal' | 'static' | 'qos' | 'mac-block',
  tagname: '',
  ip_addr: '',
  mac: '',
  iface: 'lan1',
  upload: 1000,
  download: 5000,
  enabled: 'yes',
  comment: '',
})

const quickModalTitle = computed(() => {
  switch (quickModal.kind) {
    case 'terminal':
      return '快捷添加终端设备标注'
    case 'static':
      return '快捷创建 DHCP 静态地址绑定'
    case 'qos':
      return '快捷创建 IP 限速流控规则'
    case 'mac-block':
      return '快捷添加 MAC 访问控制规则'
  }
})

const isClientMode = computed(() => ['online', 'ipv6-online', 'offline', 'summary'].includes(activeTab.value))

const sortedClients = computed(() =>
  clientSort.sortRows(
    clients.value,
    (c, key) => {
      switch (key) {
        case 'name':
          return c.comment || c.hostname || ''
        case 'uptime':
          return c.uptime || c.downtime || ''
        default:
          return c[key]
      }
    },
    (key) => {
      if (key === 'ip_addr') return 'ip'
      if (key === 'uptime') return 'duration'
      if (key === 'name' || key === 'mac') return 'auto'
      return 'number'
    }
  )
)

const sortedAuthRows = computed(() =>
  authSort.sortRows(
    authRows.value,
    (row, key) => row[key],
    (key) => authColumns.find((c) => c.key === key)?.hint ?? 'auto'
  )
)

function openQuickModal(kind: 'terminal' | 'static' | 'qos' | 'mac-block', c: ClientRecordDTO) {
  quickModal.open = true
  quickModal.busy = false
  quickModal.error = null
  quickModal.kind = kind
  quickModal.tagname = c.comment || c.hostname || c.ip_addr || 'Device'
  quickModal.ip_addr = c.ip_addr || ''
  quickModal.mac = c.mac || ''
  quickModal.iface = kind === 'qos' ? 'wan1' : 'lan1'
  quickModal.upload = 1000
  quickModal.download = 5000
  quickModal.enabled = 'yes'
  quickModal.comment = c.comment || ''
}

async function submitQuickModal() {
  quickModal.busy = true
  quickModal.error = null
  try {
    if (quickModal.kind === 'terminal') {
      await createV4Resource('terminals', { tagname: quickModal.tagname, mac: quickModal.mac, comment: quickModal.comment })
      showToast(`已标注终端 ${quickModal.tagname}`)
    } else if (quickModal.kind === 'static') {
      await createV4Resource('dhcp-static', {
        tagname: quickModal.tagname,
        ip_addr: quickModal.ip_addr,
        mac: quickModal.mac,
        interface: quickModal.iface,
        enabled: 'yes',
        comment: quickModal.comment,
      })
      showToast(`已绑定静态 IP ${quickModal.ip_addr}`)
    } else if (quickModal.kind === 'qos') {
      await createV4Resource('qos-ip', {
        tagname: quickModal.tagname,
        ip_addr: quickModal.ip_addr,
        interface: quickModal.iface,
        upload: quickModal.upload,
        download: quickModal.download,
        protocol: 'any',
        enabled: 'yes',
      })
      showToast(`已为 ${quickModal.ip_addr} 应用 QoS 限速`)
    } else if (quickModal.kind === 'mac-block') {
      await createV4Resource('mac-rules', {
        tagname: quickModal.tagname,
        mac: quickModal.mac,
        enabled: quickModal.enabled,
        comment: quickModal.comment,
      })
      showToast(`已写入 MAC 控制规则 ${quickModal.mac}`)
    }
    quickModal.open = false
  } catch (e: unknown) {
    quickModal.error = extractErr(e)
  } finally {
    quickModal.busy = false
  }
}

function extractErr(e: unknown): string {
  if (e && typeof e === 'object' && 'response' in e) {
    const resp = (e as { response?: { data?: { message?: string } } }).response
    if (resp?.data?.message) return resp.data.message
  }
  return e instanceof Error ? e.message : String(e)
}

function showToast(msg: string) {
  actionToast.value = msg
  setTimeout(() => {
    if (actionToast.value === msg) actionToast.value = null
  }, 2800)
}

async function loadClients() {
  if (!isClientMode.value) return
  const seq = ++clientReqSeq
  activeClientReqs += 1
  loading.value = true
  error.value = null
  try {
    const res = await fetchV4Clients({ mode: activeTab.value, page: 1, page_size: 200, search: search.value })
    if (seq !== clientReqSeq) return
    clients.value = res.items || []
  } catch (e: unknown) {
    if (seq !== clientReqSeq) return
    error.value = extractErr(e)
  } finally {
    activeClientReqs = Math.max(0, activeClientReqs - 1)
    if (seq === clientReqSeq) {
      loading.value = false
    }
  }
}

function pollClients() {
  if (activeClientReqs > 0) return
  void loadClients()
}

async function reloadAuthOnline() {
  const seq = ++authReqSeq
  error.value = null
  try {
    const res = await fetchV4Resource('auth-online', { page: 1, page_size: 100 })
    if (seq !== authReqSeq) return
    authRows.value = res.rows || []
  } catch (e: unknown) {
    if (seq !== authReqSeq) return
    error.value = extractErr(e)
  }
}

async function kickUser(row: Record<string, unknown>) {
  if (readOnly.value) return
  const id = String(row.id ?? '')
  if (!id || !window.confirm(`确认强制下线认证用户 ${String(row.username ?? id)} 吗？`)) return
  try {
    await kickV4OnlineUser(id)
    await reloadAuthOnline()
  } catch (e: unknown) {
    error.value = extractErr(e)
  }
}

async function openDetail(client: ClientRecordDTO) {
  const seq = ++detailReqSeq
  selectedClient.value = client
  detailLoading.value = true
  detailError.value = null
  clientDetail.value = null
  detailChart?.dispose()
  detailChart = null
  try {
    const detail = await fetchV4ClientDetail(client.ip_addr, client.mac)
    if (seq !== detailReqSeq) return
    clientDetail.value = detail
    await nextTick()
    if (seq !== detailReqSeq) return
    renderDetailChart()
  } catch (e: unknown) {
    if (seq !== detailReqSeq) return
    detailError.value = extractErr(e)
  } finally {
    if (seq === detailReqSeq) {
      detailLoading.value = false
    }
  }
}

function closeDetail() {
  detailReqSeq += 1
  selectedClient.value = null
  detailChart?.dispose()
  detailChart = null
}

function renderDetailChart() {
  if (!detailChartEl.value || !clientDetail.value) return
  if (detailChart && detailChart.getDom() !== detailChartEl.value) {
    detailChart.dispose()
    detailChart = null
  }
  if (!detailChart) detailChart = echarts.init(detailChartEl.value)
  const pal = palette.value
  const list = clientDetail.value.traffic_load || []
  detailChart.setOption({
    backgroundColor: 'transparent',
    tooltip: { trigger: 'axis', backgroundColor: pal.tooltipBg, borderColor: pal.tooltipBorder, textStyle: { color: pal.textStrong } },
    grid: { left: 48, right: 12, top: 18, bottom: 24 },
    xAxis: {
      type: 'category',
      data: list.map((item) => {
        const ts = Number(item.timestamp ?? 0)
        if (!ts) return ''
        const d = new Date(ts * 1000)
        return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
      }),
      axisLabel: { color: pal.textMuted, fontSize: 10 },
    },
    yAxis: { type: 'value', splitLine: { lineStyle: { color: pal.gridLine } }, axisLabel: { color: pal.textMuted, fontSize: 10 } },
    series: [
      { name: '下行', type: 'line', smooth: true, showSymbol: false, data: list.map((x) => Number(x.download ?? 0)), itemStyle: { color: pal.downColor } },
      { name: '上行', type: 'line', smooth: true, showSymbol: false, data: list.map((x) => Number(x.upload ?? 0)), itemStyle: { color: pal.upColor } },
    ],
  }, true)
}

watch(
  () => [props.initialSubResource, props.navSeq] as const,
  ([next]) => {
    if (currentTab.value !== 'terminals') return
    if (next && subTabs.some((t) => t.key === next)) {
      activeTab.value = next
    } else if (!next) {
      syncSubTab('terminals', activeTab.value, { replace: true })
    }
  }
)

watch(activeTab, async () => {
  await nextTick()
  const activeEl = subtabBarEl.value?.querySelector('.subtab-pill.active') as HTMLElement | null
  activeEl?.scrollIntoView({ block: 'nearest', inline: 'center', behavior: 'smooth' })
})

watch([activeTab, routerEpoch], () => {
  if (isClientMode.value) void loadClients()
  else if (activeTab.value === 'auth-online') void reloadAuthOnline()
}, { immediate: true })

let hasMounted = false

function startPolling() {
  if (pollTimer) return
  pollTimer = setInterval(() => {
    if (typeof document !== 'undefined' && document.hidden) return
    if (activeTab.value === 'online' || activeTab.value === 'ipv6-online') {
      pollClients()
    }
  }, 6000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

function handleVisibilityChange() {
  if (typeof document !== 'undefined' && !document.hidden && (activeTab.value === 'online' || activeTab.value === 'ipv6-online')) {
    void loadClients()
  }
}

onMounted(() => {
  hasMounted = true
  startPolling()
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

onActivated(() => {
  if (currentTab.value === 'terminals' && !props.initialSubResource) {
    syncSubTab('terminals', activeTab.value, { replace: true })
  }
  if (hasMounted && (activeTab.value === 'online' || activeTab.value === 'ipv6-online')) {
    void loadClients()
  }
  startPolling()
  document.addEventListener('visibilitychange', handleVisibilityChange)
  void nextTick(() => detailChart?.resize())
})

onDeactivated(() => {
  stopPolling()
  document.removeEventListener('visibilitychange', handleVisibilityChange)
  closeDetail()
  quickModal.open = false
})

onUnmounted(() => {
  stopPolling()
  document.removeEventListener('visibilitychange', handleVisibilityChange)
  detailChart?.dispose()
})
</script>

<style scoped>
.terminals-page { display: flex; flex-direction: column; gap: 16px; }
.ws-head { padding: 18px 20px; }
.ws-eyebrow { font-size: 11px; font-weight: 600; color: var(--text-faint); letter-spacing: 0.06em; }
.ws-head h1 { margin: 4px 0 0; font-size: 21px; font-weight: 800; }
.ws-head p { margin: 4px 0 0; font-size: 13px; color: var(--text-muted); }

.subtab-bar {
  display: flex; flex-wrap: wrap; gap: 4px; padding: 5px;
  border-radius: var(--radius-lg); border: 1px solid var(--seg-track-border); background: var(--seg-track-bg);
}
.subtab-pill {
  min-height: 32px; padding: 0 12px; border-radius: var(--radius-control); border: 1px solid transparent;
  background: transparent; color: var(--text-muted); font-size: 12px; font-weight: 600; cursor: pointer;
  transition: color 0.14s var(--ease-ui), background-color 0.14s var(--ease-ui), box-shadow 0.14s var(--ease-ui);
}
.subtab-pill:hover {
  color: var(--text-strong);
}
.subtab-pill.active {
  color: var(--seg-active-fg);
  border-color: var(--seg-active-border);
  background: var(--seg-active-bg);
  box-shadow: var(--seg-active-shadow);
}

.clients-card { padding: 18px; display: flex; flex-direction: column; gap: 14px; }
.toolbar { display: flex; flex-wrap: wrap; justify-content: space-between; align-items: center; gap: 10px; }
.search-box { position: relative; flex: 1; min-width: 220px; max-width: 380px; }
.search-ic { position: absolute; left: 11px; top: 50%; transform: translateY(-50%); color: var(--text-muted); }
.search-input { padding-left: 32px; }
.toolbar-right { display: flex; gap: 8px; align-items: center; }
.sort-select { width: auto; min-width: 150px; }
.sort-dir-btn { white-space: nowrap; }

.table-wrap { overflow-x: auto; border: 1px solid var(--divider-line); border-radius: var(--radius-md); }
.v4-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.v4-table th { text-align: left; padding: 10px 12px; font-size: 11px; font-weight: 600; color: var(--text-muted); background: var(--surface-table-head); border-bottom: 1px solid var(--divider-line); white-space: nowrap; }
.v4-table td { padding: 10px 12px; border-bottom: 1px solid var(--divider-line); color: var(--text-normal); white-space: nowrap; }
.v4-table tbody tr:last-child td { border-bottom: none; }

.host-cell { display: flex; flex-direction: column; gap: 2px; }
.sub-muted { font-size: 11px; color: var(--text-muted); }
.ic-up { color: var(--chart-up); }
.ic-down { color: var(--chart-down); }
.sep { margin: 0 4px; color: var(--text-faint); }
.client-actions { display: flex; flex-wrap: wrap; gap: 5px; }
.empty-state { padding: 36px 16px; text-align: center; color: var(--text-muted); }

.drawer-backdrop { position: fixed; inset: 0; z-index: 250; background: rgba(6, 10, 20, 0.45); -webkit-backdrop-filter: blur(8px); backdrop-filter: blur(8px); display: flex; justify-content: flex-end; }
.drawer-backdrop.center { align-items: center; justify-content: center; }
.drawer-panel { width: min(480px, 100vw); height: 100%; border-radius: var(--radius-xl) 0 0 var(--radius-xl); display: flex; flex-direction: column; background: var(--surface-popover); border-color: var(--chrome-border); box-shadow: var(--chrome-shadow); -webkit-backdrop-filter: var(--backdrop-chrome); backdrop-filter: var(--backdrop-chrome); }
.drawer-head { padding: 16px 20px; border-bottom: 1px solid var(--divider-line); display: flex; justify-content: space-between; align-items: center; }
.drawer-head > div > strong { display: block; font-size: 15px; }
.drawer-head > div > span { font-size: 11px; color: var(--text-muted); }
.drawer-body { padding: 18px 20px; overflow-y: auto; display: flex; flex-direction: column; gap: 18px; }
.detail-section h4 { margin: 0 0 10px; font-size: 13px; color: var(--text-strong); }
.detail-chart { width: 100%; height: 220px; }
.app-list { display: flex; flex-direction: column; gap: 8px; }
.app-row { display: flex; justify-content: space-between; padding: 10px 12px; border-radius: var(--radius-md); border: 1px solid var(--divider-line); background: var(--surface-input); font-size: 12px; }

.confirm-dialog { width: min(440px, calc(100vw - 32px)); padding: 20px; background: var(--surface-popover); border-color: var(--chrome-border); box-shadow: var(--chrome-shadow); -webkit-backdrop-filter: var(--backdrop-chrome); backdrop-filter: var(--backdrop-chrome); display: flex; flex-direction: column; gap: 12px; }
.modal-top { display: flex; justify-content: space-between; align-items: center; }
.modal-top h4 { margin: 0; font-size: 15px; }
.modal-form { display: flex; flex-direction: column; gap: 10px; }
.field-item { display: flex; flex-direction: column; gap: 5px; font-size: 12px; font-weight: 600; color: var(--text-normal); }
.confirm-actions { display: flex; justify-content: flex-end; gap: 10px; }
.spin { animation: spin 0.9s linear infinite; }

@media (max-width: 768px) {
  .ws-head { padding: 14px; }
  .ws-head h1 { font-size: 18px; }
  .subtab-bar { flex-wrap: nowrap; overflow-x: auto; -webkit-overflow-scrolling: touch; scrollbar-width: none; padding: 5px; }
  .subtab-bar::-webkit-scrollbar { display: none; }
  .subtab-pill { flex-shrink: 0; white-space: nowrap; min-height: 32px; padding: 0 11px; }
}

@media (max-width: 640px) {
  .clients-card { padding: 14px; }
  .search-box { min-width: 100%; max-width: none; }
  .toolbar-right { width: 100%; }
  .sort-select { flex: 1; min-width: 0; }
  .drawer-backdrop { align-items: flex-end; }
  .drawer-panel { width: 100%; height: auto; max-height: 88dvh; border-radius: var(--radius-xl) var(--radius-xl) 0 0; }
  .drawer-body { padding: 14px 16px; }
}
</style>

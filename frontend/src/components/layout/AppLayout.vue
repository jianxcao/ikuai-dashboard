<template>
  <div class="console-shell">
    <!-- Desktop Left Navigation Rail -->
    <aside class="side-rail v4-panel" aria-label="主导航">
      <div class="brand-row">
        <img src="/ikuai-icon.svg" alt="iKuai V4" class="brand-icon" />
        <div class="brand-text">
          <div class="brand-title-line">
            <strong>iKuai Console</strong>
            <span class="v4-pill font-mono">V4.0</span>
          </div>
          <span class="brand-sub">Multi-Router · Multi-Theme</span>
        </div>
      </div>

      <div class="rail-groups">
        <div v-for="group in navGroups" :key="group.title" class="rail-group">
          <p class="rail-group-title">{{ group.title }}</p>
          <button
            v-for="tab in group.items"
            :key="tab.id"
            type="button"
            class="rail-item"
            :class="{ active: currentTab === tab.id }"
            @click="selectTab(tab.id)"
          >
            <component :is="tab.icon" :size="16" />
            <span>{{ tab.label }}</span>
          </button>
        </div>
      </div>

      <div class="rail-footer">
        <div class="rail-theme-info">
          <span class="dot-pair">
            <i :style="{ background: currentThemeMeta?.swatchPrimary }" />
            <i :style="{ background: currentThemeMeta?.swatchSecondary }" />
          </span>
          <span>{{ currentThemeMeta?.label }}</span>
        </div>
        <button
          v-if="authEnabled"
          type="button"
          class="v4-btn logout-btn"
          @click="logout"
        >
          <LogOut :size="14" />
          <span>退出登录</span>
        </button>
      </div>
    </aside>

    <!-- Main Column -->
    <div class="workspace-col">
      <!-- Top Command Header -->
      <header class="top-bar v4-panel">
        <div class="top-left">
          <img src="/ikuai-icon.svg" alt="" class="mobile-logo" />
          <RouterFleetSwitcher @manage="selectTab('fleet')" />
        </div>

        <div class="top-right">
          <button
            type="button"
            class="cmd-trigger"
            title="快速搜索模块、ikuai-cli 命令或节点 (⌘K / Ctrl+K)"
            @click="paletteOpen = true"
          >
            <Search :size="14" />
            <span class="cmd-label">搜索功能或 CLI 命令...</span>
            <kbd class="font-mono">⌘K</kbd>
          </button>

          <button
            type="button"
            class="lock-toggle"
            :class="{ locked: readOnly, server: serverReadOnly }"
            :disabled="serverReadOnly"
            :title="
              serverReadOnly
                ? '服务端已开启 READ_ONLY 只读保护，禁止写操作'
                : readOnly
                  ? '写保护已开启（点击切换为可写模式）'
                  : '当前允许管理写入操作（点击开启只读保护锁）'
            "
            @click="toggleSafetyLock"
          >
            <Lock v-if="readOnly" :size="14" />
            <Unlock v-else :size="14" />
            <span class="lock-label-full">{{ serverReadOnly ? '服务端只读' : readOnly ? '写保护开启' : '读写就绪' }}</span>
            <span class="lock-label-short">{{ serverReadOnly ? '只读' : readOnly ? '锁定' : '可写' }}</span>
          </button>

          <ThemePicker compact class="desktop-themes" />

          <button
            v-if="authEnabled"
            type="button"
            class="v4-btn mobile-logout"
            title="退出登录"
            @click="logout"
          >
            <LogOut :size="15" />
          </button>
        </div>
      </header>

      <!-- Global Read-Only Banner -->
      <div v-if="readOnly" class="readonly-banner v4-card" role="status">
        <div class="ro-left">
          <Lock :size="15" />
          <span>
            {{
              serverReadOnly
                ? '当前服务端处于只读保护模式 (READ_ONLY=true)：所有写入操作已锁定，查询与 Dry-Run 预览不受影响。'
                : '前端写保护锁已启用：新建 / 编辑 / 启停 / 删除已锁定以防误操作，Dry-Run 预览仍可使用。'
            }}
          </span>
        </div>
        <button v-if="!serverReadOnly" type="button" class="v4-btn btn-xs" @click="toggleSafetyLock">
          解锁写入
        </button>
      </div>

      <!-- Active View -->
      <main class="main-stage">
        <KeepAlive>
          <component
            :is="currentComponent"
            :initialSubResource="activeSubResource"
            :navSeq="navSeq"
            @navigate="handleNavigate"
          />
        </KeepAlive>
      </main>
    </div>

    <!-- Mobile Bottom Navigation -->
    <nav class="mobile-bottom v4-panel" aria-label="移动导航">
      <button
        v-for="tab in mobilePrimaryTabs"
        :key="tab.id"
        type="button"
        class="m-tab"
        :class="{ active: currentTab === tab.id }"
        @click="selectTab(tab.id)"
      >
        <component :is="tab.icon" :size="18" />
        <span>{{ tab.shortLabel }}</span>
      </button>
      <button
        type="button"
        class="m-tab"
        :class="{ active: mobileMoreOpen || isSecondaryActive }"
        @click="mobileMoreOpen = !mobileMoreOpen"
      >
        <Ellipsis :size="18" />
        <span>更多</span>
      </button>
    </nav>

    <!-- Mobile More Drawer -->
    <div v-if="mobileMoreOpen" class="mobile-sheet-backdrop" @click.self="mobileMoreOpen = false">
      <div class="mobile-sheet v4-panel">
        <div class="sheet-head">
          <strong>全部 V4 模块与主题</strong>
          <button type="button" class="v4-btn" @click="mobileMoreOpen = false">
            <X :size="15" />
          </button>
        </div>
        <div class="sheet-grid">
          <button
            v-for="tab in allTabs"
            :key="tab.id"
            type="button"
            class="v4-btn"
            :class="{ 'v4-btn-primary': currentTab === tab.id }"
            @click="selectTab(tab.id)"
          >
            <component :is="tab.icon" :size="15" />
            <span>{{ tab.label }}</span>
          </button>
        </div>
        <div class="sheet-themes">
          <span class="sheet-label">主题皮肤</span>
          <ThemePicker />
        </div>
      </div>
    </div>

    <CommandPalette
      :open="paletteOpen"
      :navTabs="paletteTabs"
      @close="paletteOpen = false"
      @navigate="handleNavigate"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, onUnmounted, ref, watch } from 'vue'
import {
  Activity,
  Ellipsis,
  Globe,
  Laptop,
  Lock,
  LogOut,
  Network,
  ScrollText,
  Search,
  ServerCog,
  ShieldCheck,
  Unlock,
  Waypoints,
  Wrench,
  X,
} from 'lucide-vue-next'
import RouterFleetSwitcher from './RouterFleetSwitcher.vue'
import ThemePicker from './ThemePicker.vue'
import CommandPalette from './CommandPalette.vue'
import OverviewView from '@/views/OverviewView.vue'
import { useAuth } from '@/composables/useAuth'
import { useFleet } from '@/composables/useFleet'
import { useTheme } from '@/composables/useTheme'

const TerminalsUsersView = defineAsyncComponent(() => import('@/views/TerminalsUsersView.vue'))
const TrafficTopologyView = defineAsyncComponent(() => import('@/views/TrafficTopologyView.vue'))
const NetworkView = defineAsyncComponent(() => import('@/views/NetworkView.vue'))
const RoutingQosView = defineAsyncComponent(() => import('@/views/RoutingQosView.vue'))
const SecurityView = defineAsyncComponent(() => import('@/views/SecurityView.vue'))
const VpnWirelessObjectsView = defineAsyncComponent(() => import('@/views/VpnWirelessObjectsView.vue'))
const LogsAuditView = defineAsyncComponent(() => import('@/views/LogsAuditView.vue'))
const SystemMaintenanceView = defineAsyncComponent(() => import('@/views/SystemMaintenanceView.vue'))
const FleetThemeCenterView = defineAsyncComponent(() => import('@/views/FleetThemeCenterView.vue'))

interface NavTab {
  id: string
  label: string
  shortLabel: string
  desc: string
  icon: unknown
  mobilePrimary?: boolean
}

const navGroups: Array<{ title: string; items: NavTab[] }> = [
  {
    title: '实时与洞察',
    items: [
      { id: 'overview', label: '总览驾驶舱', shortLabel: '总览', desc: 'CPU/内存/端口/WAN 实时监测', icon: Activity, mobilePrimary: true },
      { id: 'terminals', label: '终端与用户', shortLabel: '终端', desc: '在线终端、认证会话与账号套餐', icon: Laptop, mobilePrimary: true },
      { id: 'traffic-topology', label: '流量与拓扑', shortLabel: '拓扑', desc: '协议应用画像与动态网络拓扑图', icon: Network },
    ],
  },
  {
    title: '网络与策略',
    items: [
      { id: 'network', label: '网络与接口', shortLabel: '网络', desc: 'WAN/LAN/DHCP/DNS/VLAN/NAT', icon: Globe, mobilePrimary: true },
      { id: 'routing', label: '路由与 QoS', shortLabel: '路由', desc: '静态路由、策略分流与带宽限速', icon: Waypoints },
      { id: 'security', label: '安全防护', shortLabel: '安全', desc: 'ACL、L7、MAC/URL/域名黑名单', icon: ShieldCheck },
    ],
  },
  {
    title: '隧道与审计',
    items: [
      { id: 'vpn-wireless', label: 'VPN · 无线 · 对象', shortLabel: '隧道', desc: 'WireGuard/IPSec/AC/对象组', icon: Lock },
      { id: 'logs', label: '日志审计', shortLabel: '日志', desc: '10 类系统/认证/DHCP/URL 日志', icon: ScrollText },
    ],
  },
  {
    title: '运维与平台',
    items: [
      { id: 'system', label: '系统维护', shortLabel: '维护', desc: 'BBR/SSH/VRRP/备份/固件/服务', icon: Wrench },
      { id: 'fleet', label: '节点与主题', shortLabel: '节点', desc: '多爱快 V4 切换、探测与 6 套主题', icon: ServerCog, mobilePrimary: true },
    ],
  },
]

const allTabs = navGroups.flatMap((g) => g.items)
const mobilePrimaryTabs = allTabs.filter((t) => t.mobilePrimary)
const paletteTabs = allTabs.map(({ id, label, desc }) => ({ id, label, desc }))

const currentTab = ref('overview')
const activeSubResource = ref<string | undefined>(undefined)
const navSeq = ref(0)
const paletteOpen = ref(false)
const mobileMoreOpen = ref(false)

const { authEnabled, logout, readOnly, serverReadOnly, toggleSafetyLock, setServerReadOnly } = useAuth()
const { config, isUnconfigured } = useFleet()
const { currentThemeMeta } = useTheme()

watch(
  () => config.value.server.read_only,
  (val) => {
    if (typeof val === 'boolean') setServerReadOnly(val)
  },
  { immediate: true }
)

watch(
  isUnconfigured,
  (unconfigured) => {
    if (unconfigured) {
      currentTab.value = 'fleet'
    }
  },
  { immediate: true }
)

const isSecondaryActive = computed(() => !mobilePrimaryTabs.some((t) => t.id === currentTab.value))

const viewMap: Record<string, unknown> = {
  overview: OverviewView,
  terminals: TerminalsUsersView,
  'traffic-topology': TrafficTopologyView,
  network: NetworkView,
  routing: RoutingQosView,
  security: SecurityView,
  'vpn-wireless': VpnWirelessObjectsView,
  logs: LogsAuditView,
  system: SystemMaintenanceView,
  fleet: FleetThemeCenterView,
}

const currentComponent = computed(() => viewMap[currentTab.value] ?? OverviewView)

function selectTab(id: string) {
  currentTab.value = id
  activeSubResource.value = undefined
  mobileMoreOpen.value = false
}

function handleNavigate(payload: { tab: string; subResource?: string }) {
  currentTab.value = payload.tab
  activeSubResource.value = payload.subResource
  navSeq.value += 1
  mobileMoreOpen.value = false
}

function handleGlobalKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
    e.preventDefault()
    paletteOpen.value = !paletteOpen.value
  }
}

onMounted(() => window.addEventListener('keydown', handleGlobalKeydown))
onUnmounted(() => window.removeEventListener('keydown', handleGlobalKeydown))
</script>

<style scoped>
.console-shell {
  min-height: 100vh;
  display: flex;
}

.side-rail {
  position: fixed;
  top: 14px;
  bottom: 14px;
  left: 14px;
  width: 248px;
  padding: 16px 14px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  z-index: 40;
}

.brand-row {
  display: flex;
  align-items: center;
  gap: 11px;
  padding: 2px 4px;
}

.brand-icon {
  width: 38px;
  height: 38px;
  flex-shrink: 0;
}

.brand-text {
  min-width: 0;
}

.brand-title-line {
  display: flex;
  align-items: center;
  gap: 6px;
}

.brand-title-line strong {
  font-size: 14px;
  font-weight: 800;
  color: var(--text-strong);
}

.v4-pill {
  padding: 2px 6px;
  border-radius: 6px;
  font-size: 10px;
  font-weight: 800;
  background: var(--accent-primary-soft);
  color: var(--accent-text);
  border: 1px solid color-mix(in srgb, var(--accent-primary) 42%, transparent);
}

.brand-sub {
  display: block;
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 2px;
}

.rail-groups {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding-right: 2px;
}

.rail-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.rail-group-title {
  margin: 0 0 3px 8px;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--text-muted);
}

.rail-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  min-height: 38px;
  padding: 0 12px;
  border-radius: var(--radius-control);
  border: 1px solid transparent;
  background: transparent;
  color: var(--text-normal);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.16s ease;
  text-align: left;
}

.rail-item:hover {
  color: var(--text-strong);
  background: var(--surface-hover);
}

.rail-item.active {
  color: var(--btn-primary-fg);
  border-color: var(--btn-primary-border);
  background: var(--btn-primary-bg);
  box-shadow: var(--btn-primary-shadow);
}

.rail-item.active > * {
  color: var(--btn-primary-fg);
}

.rail-footer {
  padding-top: 12px;
  border-top: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.rail-theme-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-muted);
  padding: 0 4px;
}

.dot-pair {
  display: inline-flex;
  gap: 3px;
}
.dot-pair i {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: block;
}

.logout-btn {
  width: 100%;
  justify-content: center;
}

.workspace-col {
  flex: 1;
  min-width: 0;
  margin-left: 276px;
  padding: 14px 20px 32px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.top-bar {
  padding: 10px 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  position: sticky;
  top: 12px;
  z-index: 95;
  background: var(--surface-chrome);
  -webkit-backdrop-filter: var(--backdrop-chrome);
  backdrop-filter: var(--backdrop-chrome);
}

.top-left {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  min-width: 0;
}

.mobile-logo {
  display: none;
  width: 32px;
  height: 32px;
  flex-shrink: 0;
}

.top-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.lock-label-short {
  display: none;
}

.cmd-trigger {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 38px;
  padding: 0 12px;
  border-radius: var(--radius-control);
  border: 1px solid var(--border-control);
  background: var(--surface-btn);
  color: var(--text-normal);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.16s ease;
}

.cmd-trigger:hover {
  border-color: var(--border-strong);
  color: var(--text-strong);
}

.lock-toggle {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-height: 38px;
  padding: 0 12px;
  border-radius: var(--radius-control);
  border: 1px solid color-mix(in srgb, var(--status-up) 38%, var(--border-subtle));
  background: var(--status-up-soft);
  color: var(--status-up);
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.16s ease;
}

.lock-toggle.locked {
  border-color: color-mix(in srgb, var(--status-warn) 45%, var(--border-subtle));
  background: var(--status-warn-soft);
  color: var(--status-warn);
}

.lock-toggle:disabled {
  cursor: not-allowed;
  opacity: 0.92;
}

.readonly-banner {
  padding: 10px 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-color: color-mix(in srgb, var(--status-warn) 48%, transparent);
  background: var(--status-warn-soft);
  color: var(--status-warn);
  font-size: 12.5px;
  font-weight: 600;
}

.ro-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cmd-trigger kbd {
  padding: 2px 6px;
  border-radius: 6px;
  border: 1px solid var(--border-subtle);
  background: var(--bg-elevated);
  font-size: 11px;
}

.mobile-logout {
  display: none;
}

.main-stage {
  min-width: 0;
}

.mobile-bottom {
  display: none;
}

.mobile-sheet-backdrop {
  position: fixed;
  inset: 0;
  z-index: 240;
  background: rgba(2, 6, 14, 0.56);
  backdrop-filter: blur(6px);
  display: flex;
  align-items: flex-end;
  padding: 12px;
}

.mobile-sheet {
  width: 100%;
  max-height: min(85dvh, 680px);
  overflow-y: auto;
  padding: 16px;
  padding-bottom: max(16px, env(safe-area-inset-bottom));
  background: var(--surface-popover);
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.sheet-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.sheet-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.sheet-themes {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sheet-label {
  font-size: 11px;
  font-weight: 700;
  color: var(--text-muted);
}

@media (max-width: 1120px) {
  .side-rail {
    display: none;
  }

  .workspace-col {
    margin-left: 0;
    padding: 12px 14px calc(94px + env(safe-area-inset-bottom));
  }

  .mobile-logo,
  .mobile-logout {
    display: inline-flex;
  }

  .desktop-themes,
  .cmd-label,
  .cmd-trigger kbd {
    display: none;
  }

  .cmd-trigger,
  .mobile-logout {
    width: 36px;
    min-height: 36px;
    padding: 0;
    justify-content: center;
  }

  .mobile-bottom {
    position: fixed;
    left: 10px;
    right: 10px;
    bottom: max(10px, env(safe-area-inset-bottom));
    z-index: 100;
    display: grid;
    grid-template-columns: repeat(5, minmax(0, 1fr));
    padding: 6px;
    gap: 4px;
    background: var(--surface-chrome);
    -webkit-backdrop-filter: var(--backdrop-chrome);
    backdrop-filter: var(--backdrop-chrome);
    box-shadow: 0 -6px 24px rgba(0, 0, 0, 0.22), var(--shadow-elevated), var(--inner-sheen);
  }

  .m-tab {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 3px;
    min-height: 50px;
    border-radius: 14px;
    border: 1px solid transparent;
    background: transparent;
    color: var(--text-normal);
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.16s ease;
  }

  .m-tab.active {
    color: var(--btn-primary-fg);
    border-color: var(--btn-primary-border);
    background: var(--btn-primary-bg);
    box-shadow: var(--btn-primary-shadow);
  }

  .m-tab.active > * {
    color: var(--btn-primary-fg);
  }
}

@media (max-width: 640px) {
  .workspace-col {
    padding: 8px 10px calc(92px + env(safe-area-inset-bottom));
    gap: 12px;
  }

  .top-bar {
    top: 6px;
    padding: 6px 10px;
    gap: 6px;
    border-radius: var(--radius-lg);
  }

  .mobile-logo {
    width: 26px;
    height: 26px;
  }

  .top-left {
    gap: 6px;
  }

  .top-right {
    gap: 5px;
  }

  .cmd-trigger,
  .mobile-logout {
    width: 34px;
    min-height: 34px;
  }

  .lock-toggle {
    min-height: 34px;
    padding: 0 8px;
    font-size: 11px;
    gap: 4px;
  }

  .lock-label-full {
    display: none;
  }

  .lock-label-short {
    display: inline;
  }

  .readonly-banner {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
    padding: 10px 12px;
    font-size: 12px;
  }
}
</style>

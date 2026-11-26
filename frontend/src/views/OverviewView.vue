<template>
  <section class="overview-page">
    <!-- Router Identity Strip -->
    <header class="identity-banner v4-panel">
      <div class="identity-left">
        <div class="host-row">
          <span class="live-dot" />
          <h1>{{ overview?.router.hostname || activeRouter?.name || 'iKuai V4 路由器' }}</h1>
          <span v-if="overview?.router.version" class="v4-badge v4-badge-accent font-mono">
            v{{ overview.router.version }}
          </span>
          <span v-if="overview?.router.mock" class="v4-badge v4-badge-warn">模拟遥测</span>
        </div>
        <div class="meta-line font-mono">
          <span v-if="overview?.router.verstring">{{ overview.router.verstring }}</span>
          <span v-if="overview?.router.arch">· {{ overview.router.arch }}</span>
          <span v-if="overview?.router.lan_ip">· LAN {{ overview.router.lan_ip }}</span>
          <span v-if="overview?.router.uptime">· 在线 {{ formatUptime(overview.router.uptime) }}</span>
        </div>
      </div>

      <div class="identity-actions">
        <button
          type="button"
          class="v4-btn btn-xs"
          :disabled="actionBusy || readOnly"
          :title="readOnly ? '只读保护中' : '立即同步路由器 NTP 时间'"
          @click="runQuickAction('ntp-sync', 'NTP 对时指令已下发')"
        >
          <Clock :size="13" />
          <span>同步 NTP</span>
        </button>
        <button
          type="button"
          class="v4-btn btn-xs"
          :disabled="actionBusy || readOnly"
          :title="readOnly ? '只读保护中' : '创建路由器配置快照备份'"
          @click="runQuickAction('backup-create', '已创建配置快照备份')"
        >
          <Archive :size="13" />
          <span>立即备份</span>
        </button>
        <button
          type="button"
          class="v4-btn btn-xs"
          :disabled="actionBusy || readOnly"
          :title="readOnly ? '只读保护中' : '在线检查最新系统固件'"
          @click="runQuickAction('upgrade-check', '固件更新检查已触发')"
        >
          <DownloadCloud :size="13" />
          <span>检查更新</span>
        </button>
        <button
          type="button"
          class="v4-btn btn-xs"
          :disabled="actionBusy || readOnly"
          :title="readOnly ? '只读保护中' : '重启局域网 DHCP 服务'"
          @click="runQuickAction('dhcp-restart', 'DHCP 服务已重启')"
        >
          <RotateCw :size="13" />
          <span>重启 DHCP</span>
        </button>
        <button
          type="button"
          class="v4-btn btn-xs"
          :class="{ 'v4-btn-primary': !layoutLocked }"
          :title="layoutLocked ? '解锁首页卡片拖动调整顺序' : '完成调整并固定当前首页布局'"
          @click="layoutLocked ? unlockLayout() : finishLayoutEdit()"
        >
          <Pin v-if="!layoutLocked" :size="13" />
          <LayoutGrid v-else :size="13" />
          <span>{{ layoutLocked ? '调整布局' : '固定布局' }}</span>
        </button>
        <button type="button" class="v4-btn" :disabled="loading" @click="refreshAll">
          <RefreshCw :size="14" :class="{ spin: loading }" />
          <span>立即刷新</span>
        </button>
      </div>
    </header>

    <div v-if="!layoutLocked" class="layout-edit-bar v4-card" role="status">
      <div class="layout-edit-copy">
        <Pin :size="14" />
        <span>布局调整中：按住拖动上方指标卡或下方任意模块即可互换位置（手机端可点 ← / →），完成后点击「固定布局」锁定。</span>
      </div>
      <div class="layout-edit-actions">
        <button type="button" class="v4-btn btn-xs" :disabled="!isCustomLayout" @click="handleResetLayout">
          <RotateCcw :size="13" />
          <span>恢复默认</span>
        </button>
        <button type="button" class="v4-btn v4-btn-primary btn-xs" @click="finishLayoutEdit">
          <Check :size="13" />
          <span>固定布局</span>
        </button>
      </div>
    </div>

    <div
      v-if="actionNotice"
      class="notice-row"
      :class="actionNotice.tone === 'warn' ? 'warn-banner' : actionNotice.tone === 'err' ? 'error-banner' : 'toast-banner'"
    >
      <span>{{ actionNotice.text }}</span>
      <button type="button" class="notice-dismiss" aria-label="关闭提示" @click="actionNotice = null">×</button>
    </div>
    <div v-if="error" class="error-banner">{{ error }}</div>

    <div v-if="loading && !overview" class="loading-state">
      <div class="spinner" />
      <span class="font-mono">正在连接 iKuai V4 遥测总线...</span>
    </div>

    <template v-else-if="overview">
      <!-- 4 Realtime Telemetry KPI Cards -->
      <div class="kpi-grid">
        <article
          v-for="(kpiId, kpiIdx) in kpiOrder"
          :key="kpiId"
          class="kpi-card v4-card"
          :class="{
            clickable: layoutLocked && (kpiId === 'connections' || kpiId === 'users'),
            'layout-editable': !layoutLocked,
            'is-dragging': dragState.group === 'kpi' && dragState.draggingId === kpiId,
            'is-drop-target':
              dragState.group === 'kpi' && dragState.overId === kpiId && dragState.draggingId !== kpiId,
          }"
          :draggable="!layoutLocked"
          :title="kpiCardTitle(kpiId)"
          @click="handleKpiClick(kpiId)"
          @dragstart="onDragStart('kpi', kpiId, $event)"
          @dragover.prevent
          @dragenter.prevent="onDragEnter('kpi', kpiId)"
          @dragleave="onDragLeave('kpi', kpiId, $event)"
          @drop.prevent="onDrop('kpi', kpiId)"
          @dragend="onDragEnd"
        >
          <div v-if="!layoutLocked" class="card-drag-bar">
            <span class="drag-handle-pill">
              <GripVertical :size="12" />
              <span>拖动互换</span>
            </span>
            <div class="step-btns">
              <button
                type="button"
                class="drag-step-btn"
                title="左移"
                :disabled="kpiIdx === 0"
                @click.stop="moveKpi(kpiId, -1)"
              >
                <ArrowLeft :size="11" />
              </button>
              <button
                type="button"
                class="drag-step-btn"
                title="右移"
                :disabled="kpiIdx === kpiOrder.length - 1"
                @click.stop="moveKpi(kpiId, 1)"
              >
                <ArrowRight :size="11" />
              </button>
            </div>
          </div>

          <template v-if="kpiId === 'upload'">
            <div class="kpi-head">
              <span>实时上行速率</span>
              <ArrowUpRight :size="16" class="ic-up" />
            </div>
            <div class="kpi-value font-mono">
              <strong>{{ upParts.value }}</strong>
              <span>{{ upParts.unit }}</span>
            </div>
            <div class="kpi-sub font-mono">外网总计 · 累计上传 {{ formatBytes(overview.stream.total_up) }}</div>
          </template>

          <template v-else-if="kpiId === 'download'">
            <div class="kpi-head">
              <span>实时下行速率</span>
              <ArrowDownLeft :size="16" class="ic-down" />
            </div>
            <div class="kpi-value font-mono">
              <strong>{{ downParts.value }}</strong>
              <span>{{ downParts.unit }}</span>
            </div>
            <div class="kpi-sub font-mono">外网总计 · 累计下载 {{ formatBytes(overview.stream.total_down) }}</div>
          </template>

          <template v-else-if="kpiId === 'connections'">
            <div class="kpi-head">
              <span>活跃 NAT 会话</span>
              <Activity :size="16" class="ic-accent" />
            </div>
            <div class="kpi-value font-mono">
              <strong>{{ overview.connections.total.toLocaleString() }}</strong>
              <span>Conns</span>
            </div>
            <div class="kpi-sub font-mono">
              TCP {{ overview.connections.tcp }} · UDP {{ overview.connections.udp }} · ICMP {{ overview.connections.icmp }}
            </div>
          </template>

          <template v-else-if="kpiId === 'users'">
            <div class="kpi-head">
              <span>在线终端设备</span>
              <Users :size="16" class="ic-up" />
            </div>
            <div class="kpi-value font-mono">
              <strong>{{ overview.online_users.total }}</strong>
              <span>台在线</span>
            </div>
            <div class="kpi-sub font-mono">
              有线 {{ overview.online_users.wired }} · 2.4G {{ overview.online_users.wifi_2g }} · 5G {{ overview.online_users.wifi_5g }}
            </div>
          </template>
        </article>
      </div>

      <!-- Draggable 2-Column Overview Modules Grid -->
      <div class="two-col-grid sections-grid">
        <article
          v-for="(secId, secIdx) in sectionOrder"
          :key="secId"
          class="v4-card section-card"
          :class="{
            'layout-editable': !layoutLocked,
            'is-dragging': dragState.group === 'section' && dragState.draggingId === secId,
            'is-drop-target':
              dragState.group === 'section' && dragState.overId === secId && dragState.draggingId !== secId,
          }"
          :draggable="!layoutLocked"
          @dragstart="onDragStart('section', secId, $event)"
          @dragover.prevent
          @dragenter.prevent="onDragEnter('section', secId)"
          @dragleave="onDragLeave('section', secId, $event)"
          @drop.prevent="onDrop('section', secId)"
          @dragend="onDragEnd"
        >
          <template v-if="secId === 'cpu-mem'">
            <div class="card-title-row">
              <div class="card-title">
                <Cpu :size="16" />
                <h3>计算与内存负载</h3>
              </div>
              <div class="title-right-group">
                <span class="v4-badge font-mono" v-if="overview.cpu?.freq_mhz?.length">
                  {{ overview.cpu.freq_mhz[0] }} MHz
                </span>
                <div v-if="!layoutLocked" class="drag-ctl-group">
                  <span class="drag-handle-pill" title="按住拖动与其他模块互换位置">
                    <GripVertical :size="13" />
                    <span>拖动</span>
                  </span>
                  <button
                    type="button"
                    class="drag-step-btn"
                    title="前移"
                    :disabled="secIdx === 0"
                    @click.stop="moveSection(secId, -1)"
                  >
                    <ArrowLeft :size="12" />
                  </button>
                  <button
                    type="button"
                    class="drag-step-btn"
                    title="后移"
                    :disabled="secIdx === sectionOrder.length - 1"
                    @click.stop="moveSection(secId, 1)"
                  >
                    <ArrowRight :size="12" />
                  </button>
                </div>
              </div>
            </div>

            <div class="hw-split">
              <div class="hw-block">
                <div class="hw-label-row">
                  <span>CPU 综合负载</span>
                  <strong class="font-mono">{{ (overview.cpu?.avg_percent ?? 0).toFixed(2) }}%</strong>
                </div>
                <div class="core-bars">
                  <div v-for="(val, idx) in overview.cpu?.cores || []" :key="idx" class="core-row">
                    <span class="font-mono core-idx">Core {{ idx }}</span>
                    <div class="bar-track">
                      <div class="bar-fill accent" :style="{ width: `${Math.min(100, Math.max(2, val))}%` }" />
                    </div>
                    <span class="font-mono core-val">{{ val.toFixed(1) }}%</span>
                  </div>
                </div>
              </div>

              <div class="hw-block">
                <div class="hw-label-row">
                  <span>物理内存占用</span>
                  <strong class="font-mono">{{ overview.memory.used_percent.toFixed(0) }}%</strong>
                </div>
                <div class="bar-track lg">
                  <div class="bar-fill up" :style="{ width: `${Math.min(100, overview.memory.used_percent)}%` }" />
                </div>
                <div class="mem-details font-mono">
                  <span>总容量 {{ formatBytes(overview.memory.total_kb * 1024) }}</span>
                  <span>可用 {{ formatBytes(overview.memory.available_kb * 1024) }}</span>
                  <span>缓存 {{ formatBytes(overview.memory.cached_kb * 1024) }}</span>
                  <span>缓冲 {{ formatBytes(overview.memory.buffers_kb * 1024) }}</span>
                </div>
              </div>
            </div>
          </template>

          <template v-else-if="secId === 'physical-ports'">
            <div class="card-title-row">
              <div class="card-title">
                <Cable :size="16" />
                <h3>物理网口面板 (ether_info)</h3>
              </div>
              <div class="title-right-group">
                <div class="pill-group">
                  <span class="v4-badge font-mono">{{ overview.physical_ports?.length ?? 0 }} Ports</span>
                  <button type="button" class="mini-pill" @click="emit('navigate', { tab: 'network', subResource: 'physical-nics' })">网口配置</button>
                </div>
                <div v-if="!layoutLocked" class="drag-ctl-group">
                  <span class="drag-handle-pill" title="按住拖动与其他模块互换位置">
                    <GripVertical :size="13" />
                    <span>拖动</span>
                  </span>
                  <button type="button" class="drag-step-btn" title="前移" :disabled="secIdx === 0" @click.stop="moveSection(secId, -1)">
                    <ArrowLeft :size="12" />
                  </button>
                  <button type="button" class="drag-step-btn" title="后移" :disabled="secIdx === sectionOrder.length - 1" @click.stop="moveSection(secId, 1)">
                    <ArrowRight :size="12" />
                  </button>
                </div>
              </div>
            </div>

            <div class="port-matrix">
              <div
                v-for="port in overview.physical_ports || []"
                :key="port.name"
                class="port-slot"
                :class="{ linked: port.link === 1 }"
              >
                <div class="port-top">
                  <span class="port-led" :class="{ on: port.link === 1 }" />
                  <strong class="font-mono">{{ port.name }}</strong>
                  <span class="v4-badge" :class="port.interface && port.interface !== 'Null' ? 'v4-badge-accent' : ''">
                    {{ port.interface && port.interface !== 'Null' ? port.interface : '未分配' }}
                  </span>
                </div>
                <div class="port-mid font-mono">
                  <span>{{ port.speed > 0 ? `${port.speed} Mbps` : port.link === 1 ? '已连接' : '未连接' }}</span>
                  <span>{{ port.type || 'NIC' }}</span>
                </div>
                <div class="port-sub font-mono" :title="port.model">
                  {{ port.mac }} · {{ port.driver }}
                </div>
              </div>
            </div>
          </template>

          <template v-else-if="secId === 'wan-chart'">
            <div class="card-title-row">
              <div class="card-title">
                <Gauge :size="16" />
                <h3>WAN 历史速率与链路延迟</h3>
              </div>
              <div class="title-right-group">
                <div class="v4-seg-bar" role="tablist">
                  <button type="button" class="v4-seg-btn" :class="{ active: wanChartMode === 'speed' }" @click="wanChartMode = 'speed'">速率 KB/s</button>
                  <button type="button" class="v4-seg-btn" :class="{ active: wanChartMode === 'rtt' }" @click="wanChartMode = 'rtt'">延迟 RTT ms</button>
                </div>
                <div v-if="!layoutLocked" class="drag-ctl-group">
                  <span class="drag-handle-pill" title="按住拖动与其他模块互换位置">
                    <GripVertical :size="13" />
                    <span>拖动</span>
                  </span>
                  <button type="button" class="drag-step-btn" title="前移" :disabled="secIdx === 0" @click.stop="moveSection(secId, -1)">
                    <ArrowLeft :size="12" />
                  </button>
                  <button type="button" class="drag-step-btn" title="后移" :disabled="secIdx === sectionOrder.length - 1" @click.stop="moveSection(secId, 1)">
                    <ArrowRight :size="12" />
                  </button>
                </div>
              </div>
            </div>
            <div :ref="setWanChartEl" class="chart-box" />
          </template>

          <template v-else-if="secId === 'load-chart'">
            <div class="card-title-row">
              <div class="card-title">
                <TrendingUp :size="16" />
                <h3>系统资源历史曲线</h3>
              </div>
              <div class="title-right-group">
                <div class="pill-group">
                  <select v-model="loadMetric" class="mini-select" aria-label="监控指标">
                    <option value="cpu">CPU %</option>
                    <option value="memory">内存 %</option>
                    <option value="connections">连接数</option>
                    <option value="terminals">终端数</option>
                    <option value="network">网络负载</option>
                  </select>
                  <select v-model="loadRange" class="mini-select" aria-label="时间范围">
                    <option value="hour">1 小时</option>
                    <option value="day">24 小时</option>
                    <option value="week">7 天</option>
                  </select>
                </div>
                <div v-if="!layoutLocked" class="drag-ctl-group">
                  <span class="drag-handle-pill" title="按住拖动与其他模块互换位置">
                    <GripVertical :size="13" />
                    <span>拖动</span>
                  </span>
                  <button type="button" class="drag-step-btn" title="前移" :disabled="secIdx === 0" @click.stop="moveSection(secId, -1)">
                    <ArrowLeft :size="12" />
                  </button>
                  <button type="button" class="drag-step-btn" title="后移" :disabled="secIdx === sectionOrder.length - 1" @click.stop="moveSection(secId, 1)">
                    <ArrowRight :size="12" />
                  </button>
                </div>
              </div>
            </div>
            <div class="chart-stage">
              <div :ref="setLoadChartEl" class="chart-box" />
              <div v-if="loadLoading && !loadData?.series?.length" class="chart-overlay font-mono">
                正在读取历史曲线...
              </div>
              <div v-else-if="!loadLoading && loadData && loadData.series.length === 0" class="chart-overlay font-mono">
                当前时间范围暂无历史采样数据
              </div>
            </div>
          </template>

          <template v-else-if="secId === 'wan-checks'">
            <div class="card-title-row">
              <div class="card-title">
                <Globe :size="16" />
                <h3>WAN 线路健康检查</h3>
              </div>
              <div class="title-right-group">
                <div class="pill-group">
                  <button type="button" class="mini-pill" @click="emit('navigate', { tab: 'routing', subResource: 'stream-domain' })">分流策略</button>
                  <button type="button" class="mini-pill" @click="emit('navigate', { tab: 'network', subResource: 'dnat-rules' })">DNAT 映射</button>
                </div>
                <div v-if="!layoutLocked" class="drag-ctl-group">
                  <span class="drag-handle-pill" title="按住拖动与其他模块互换位置">
                    <GripVertical :size="13" />
                    <span>拖动</span>
                  </span>
                  <button type="button" class="drag-step-btn" title="前移" :disabled="secIdx === 0" @click.stop="moveSection(secId, -1)">
                    <ArrowLeft :size="12" />
                  </button>
                  <button type="button" class="drag-step-btn" title="后移" :disabled="secIdx === sectionOrder.length - 1" @click.stop="moveSection(secId, 1)">
                    <ArrowRight :size="12" />
                  </button>
                </div>
              </div>
            </div>
            <div class="wan-list">
              <div v-for="w in overview.wan_checks || []" :key="w.interface" class="wan-row">
                <div class="wan-head">
                  <div class="wan-title">
                    <strong class="font-mono">{{ w.interface }}</strong>
                    <span class="v4-badge">{{ w.internet }}</span>
                    <span v-if="w.comment" class="wan-comment">{{ w.comment }}</span>
                  </div>
                  <span class="v4-badge" :class="w.result === 'success' ? 'v4-badge-up' : 'v4-badge-down'">
                    {{ w.errmsg || w.result }}
                  </span>
                </div>
                <div class="wan-meta font-mono">
                  <span>IP {{ w.ip_addr || '—' }}</span>
                  <span>GW {{ w.gateway || '—' }}</span>
                  <span>RTT {{ w.avg_rtt_ms.toFixed(2) }}ms</span>
                  <span>丢包 {{ w.drop_rate.toFixed(2) }}%</span>
                </div>
              </div>
            </div>
          </template>

          <template v-else-if="secId === 'iface-traffic'">
            <div class="card-title-row">
              <div class="card-title">
                <Network :size="16" />
                <h3>接口实时吞吐明细</h3>
              </div>
              <div class="title-right-group">
                <span
                  v-if="wanTrafficSummary.count > 0"
                  class="v4-badge font-mono"
                  title="接口明细来自 /monitoring/interfaces-status；此处汇总全部 WAN 口实时速率（不包含内网 LAN 口），可与顶部外网总速率对照"
                >
                  WAN 合计 ↑{{ formatSpeed(wanTrafficSummary.up) }} · ↓{{ formatSpeed(wanTrafficSummary.down) }}
                </span>
                <div v-if="!layoutLocked" class="drag-ctl-group">
                  <span class="drag-handle-pill" title="按住拖动与其他模块互换位置">
                    <GripVertical :size="13" />
                    <span>拖动</span>
                  </span>
                  <button type="button" class="drag-step-btn" title="前移" :disabled="secIdx === 0" @click.stop="moveSection(secId, -1)">
                    <ArrowLeft :size="12" />
                  </button>
                  <button type="button" class="drag-step-btn" title="后移" :disabled="secIdx === sectionOrder.length - 1" @click.stop="moveSection(secId, 1)">
                    <ArrowRight :size="12" />
                  </button>
                </div>
              </div>
            </div>
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
                        <span
                          class="iface-role-tag"
                          :class="isWanInterface(iface.interface) ? 'wan' : 'lan'"
                        >
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
          </template>
        </article>
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onActivated, onDeactivated, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { echarts } from '@/lib/echarts'
import {
  Activity,
  Archive,
  ArrowDownLeft,
  ArrowLeft,
  ArrowRight,
  ArrowUpRight,
  Cable,
  Check,
  Clock,
  Cpu,
  DownloadCloud,
  Gauge,
  Globe,
  GripVertical,
  LayoutGrid,
  Network,
  Pin,
  RefreshCw,
  RotateCcw,
  RotateCw,
  TrendingUp,
  Users,
} from 'lucide-vue-next'
import { executeV4Action, fetchV4LoadHistory, fetchV4Overview, type LoadHistoryDTO, type OverviewDTO } from '@/api/v4'
import { useAuth } from '@/composables/useAuth'
import { useFleet } from '@/composables/useFleet'
import { useChartTheme } from '@/composables/useChartTheme'
import { formatBytes, formatSpeed, formatSpeedParts, formatUptime } from '@/composables/useFormatters'
import SortIndicator from '@/components/common/SortIndicator.vue'
import { useTableSort, type SortValueHint } from '@/composables/useTableSort'
import { useOverviewLayout, type KpiCardId, type SectionCardId } from '@/composables/useOverviewLayout'

const emit = defineEmits<{
  (e: 'navigate', payload: { tab: string; subResource?: string }): void
}>()

const { readOnly } = useAuth()
const { activeRouter, routerEpoch } = useFleet()
const { palette } = useChartTheme()
const {
  kpiOrder,
  sectionOrder,
  layoutLocked,
  isCustomLayout,
  swapKpi,
  swapSection,
  moveKpi,
  moveSection,
  unlockLayout,
  lockLayout,
  resetLayout,
} = useOverviewLayout()

const loading = ref(true)
const loadLoading = ref(false)
const actionBusy = ref(false)
const actionNotice = ref<{ text: string; tone: 'ok' | 'warn' | 'err' } | null>(null)
const error = ref<string | null>(null)
const overview = ref<OverviewDTO | null>(null)
const loadData = ref<LoadHistoryDTO | null>(null)

const dragState = reactive<{
  group: 'kpi' | 'section' | null
  draggingId: string | null
  overId: string | null
}>({ group: null, draggingId: null, overId: null })

const wanChartMode = ref<'speed' | 'rtt'>('speed')
const loadMetric = ref('cpu')
const loadRange = ref('hour')

const wanChartEl = ref<HTMLElement | null>(null)
const loadChartEl = ref<HTMLElement | null>(null)

let wanChart: echarts.ECharts | null = null
let loadChart: echarts.ECharts | null = null
let resizeObserver: ResizeObserver | null = null
let resizeRaf = 0
let pollTimer: ReturnType<typeof setInterval> | null = null
let overviewReqSeq = 0
let activeOverviewReqs = 0
let loadReqSeq = 0
let activeLoadReqs = 0
let lastWanRenderSig = ''

function setWanChartEl(el: unknown) {
  const dom = el instanceof HTMLElement ? el : null
  if (wanChartEl.value !== dom) {
    if (wanChartEl.value) resizeObserver?.unobserve(wanChartEl.value)
    wanChartEl.value = dom
    if (dom) resizeObserver?.observe(dom)
  }
}

function setLoadChartEl(el: unknown) {
  const dom = el instanceof HTMLElement ? el : null
  if (loadChartEl.value !== dom) {
    if (loadChartEl.value) resizeObserver?.unobserve(loadChartEl.value)
    loadChartEl.value = dom
    if (dom) resizeObserver?.observe(dom)
  }
}

function kpiCardTitle(id: KpiCardId): string {
  if (!layoutLocked.value) return '按住拖动与其他指标卡互换位置'
  if (id === 'upload' || id === 'download') {
    return '直接取自 /api/v4.0/monitoring/system 的 sysinfo.stream（全部 WAN 外网出口总吞吐）'
  }
  return id === 'connections' ? '点击查看流量与协议画像' : '点击查看在线终端列表'
}

function handleKpiClick(id: KpiCardId) {
  if (!layoutLocked.value) return
  if (id === 'connections') emit('navigate', { tab: 'traffic-topology' })
  else if (id === 'users') emit('navigate', { tab: 'terminals', subResource: 'online' })
}

function onDragStart(group: 'kpi' | 'section', id: string, event: DragEvent) {
  if (layoutLocked.value) {
    event.preventDefault()
    return
  }
  const target = event.target as HTMLElement | null
  if (target?.closest('button, select, input, th') && !target.closest('.drag-handle-pill')) {
    event.preventDefault()
    return
  }
  dragState.group = group
  dragState.draggingId = id
  dragState.overId = id
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', `${group}:${id}`)
  }
}

function onDragEnter(group: 'kpi' | 'section', id: string) {
  if (layoutLocked.value || dragState.group !== group || !dragState.draggingId) return
  dragState.overId = id
}

function onDragLeave(group: 'kpi' | 'section', id: string, event: DragEvent) {
  const current = event.currentTarget as HTMLElement | null
  const related = event.relatedTarget as Node | null
  if (current && related && current.contains(related)) return
  if (dragState.group === group && dragState.overId === id && dragState.draggingId !== id) {
    dragState.overId = null
  }
}

function onDrop(group: 'kpi' | 'section', targetId: string) {
  if (layoutLocked.value || dragState.group !== group || !dragState.draggingId) return
  const fromId = dragState.draggingId
  if (fromId !== targetId) {
    if (group === 'kpi') swapKpi(fromId as KpiCardId, targetId as KpiCardId)
    else swapSection(fromId as SectionCardId, targetId as SectionCardId)
  }
  dragState.group = null
  dragState.draggingId = null
  dragState.overId = null
}

function onDragEnd() {
  dragState.group = null
  dragState.draggingId = null
  dragState.overId = null
}

function finishLayoutEdit() {
  onDragEnd()
  lockLayout()
  showToast('首页模块布局已固定保存', 'ok')
}

function handleResetLayout() {
  onDragEnd()
  resetLayout()
  showToast('已恢复默认模块排序', 'ok')
}

function showToast(text: string, tone: 'ok' | 'warn' | 'err') {
  actionNotice.value = { text, tone }
  setTimeout(() => {
    if (actionNotice.value?.text === text) actionNotice.value = null
  }, 4200)
}

const upParts = computed(() => formatSpeedParts(overview.value?.stream.upload_bps))
const downParts = computed(() => formatSpeedParts(overview.value?.stream.download_bps))

function isWanInterface(name: string): boolean {
  const clean = name.trim().toLowerCase()
  if (!clean) return false
  const checks = overview.value?.wan_checks || []
  if (checks.some((w) => w.interface.trim().toLowerCase() === clean)) {
    return true
  }
  return !(clean.startsWith('lan') || clean.startsWith('vlan') || clean.startsWith('br'))
}

const wanTrafficSummary = computed(() => {
  const list = overview.value?.interface_traffic || []
  let up = 0
  let down = 0
  let count = 0
  for (const item of list) {
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
    overview.value?.interface_traffic || [],
    (iface, key) => {
      if (key === 'total_bytes') return (iface.total_up || 0) + (iface.total_down || 0)
      return iface[key]
    },
    (key) => ifaceColumns.find((c) => c.key === key)?.hint ?? 'auto'
  )
)

async function fetchOverviewOnly(clearLoading = true, background = false) {
  if (background && activeOverviewReqs > 0) return
  const seq = ++overviewReqSeq
  activeOverviewReqs += 1
  try {
    const data = await fetchV4Overview()
    if (seq !== overviewReqSeq) return
    overview.value = data
    error.value = null
    await nextTick()
    if (seq !== overviewReqSeq) return
    renderWanChart(false)
  } catch (e: unknown) {
    if (seq !== overviewReqSeq) return
    error.value = extractErr(e)
  } finally {
    activeOverviewReqs = Math.max(0, activeOverviewReqs - 1)
    if (clearLoading && seq === overviewReqSeq) {
      loading.value = false
    }
  }
}

async function fetchLoadOnly() {
  const seq = ++loadReqSeq
  activeLoadReqs += 1
  loadLoading.value = true
  try {
    const data = await fetchV4LoadHistory(loadMetric.value, loadRange.value, 'avg')
    if (seq !== loadReqSeq) return
    loadData.value = data
    await nextTick()
    if (seq !== loadReqSeq) return
    renderLoadChart()
  } catch {
    // Keep previous chart on transient error; background polling retries when empty
  } finally {
    activeLoadReqs = Math.max(0, activeLoadReqs - 1)
    if (seq === loadReqSeq) {
      loadLoading.value = false
    }
  }
}

async function refreshAll() {
  loading.value = true
  try {
    await Promise.all([fetchOverviewOnly(false), fetchLoadOnly()])
  } finally {
    loading.value = false
  }
  await nextTick()
  renderWanChart()
  renderLoadChart()
}

async function runQuickAction(actionKey: string, okText: string) {
  if (readOnly.value) return
  actionBusy.value = true
  try {
    const res = await executeV4Action(actionKey, false)
    const msg = String(res.message || okText)
    const tone: 'ok' | 'warn' = res.warning ? 'warn' : 'ok'
    actionNotice.value = { text: msg, tone }
    setTimeout(() => {
      if (actionNotice.value?.text === msg) actionNotice.value = null
    }, 5500)
    await fetchOverviewOnly()
  } catch (e: unknown) {
    const msg = extractErr(e)
    actionNotice.value = { text: msg, tone: 'err' }
    setTimeout(() => {
      if (actionNotice.value?.text === msg) actionNotice.value = null
    }, 5500)
  } finally {
    actionBusy.value = false
  }
}

function renderWanChart(force = false) {
  if (!wanChartEl.value || !overview.value) return
  if (wanChart && wanChart.getDom() !== wanChartEl.value) {
    wanChart.dispose()
    wanChart = null
  }
  if (!wanChart) {
    wanChart = echarts.init(wanChartEl.value)
    resizeObserver?.observe(wanChartEl.value)
  }
  const pal = palette.value
  const history = overview.value.wan_history || []
  const points = history.filter((p) => p.interface === 'all')
  let list = points
  if (list.length === 0 && history.length > 0) {
    const byTs = new Map<
      number,
      {
        timestamp: number
        interface: string
        avg_upload_kbps: number
        avg_download_kbps: number
        max_upload_kbps: number
        max_download_kbps: number
        avg_rtt_ms: number
        max_rtt_ms: number
        min_rtt_ms: number
        drop_rate: number
        rttSum: number
        rttCount: number
      }
    >()
    for (const p of history) {
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
      }
    }
    list = Array.from(byTs.values())
      .sort((a, b) => a.timestamp - b.timestamp)
      .map((item) => ({
        ...item,
        avg_upload_kbps: Math.round(item.avg_upload_kbps * 100) / 100,
        avg_download_kbps: Math.round(item.avg_download_kbps * 100) / 100,
        avg_rtt_ms: item.rttCount > 0 ? Math.round((item.rttSum / item.rttCount) * 100) / 100 : 0,
      }))
  }
  const lastTs = list[list.length - 1]?.timestamp ?? 0
  const sig = `${wanChartMode.value}|${pal.downColor}|${pal.upColor}|${list.length}|${lastTs}`
  if (!force && sig === lastWanRenderSig) return
  lastWanRenderSig = sig

  const labels = list.map((p) => {
    const d = new Date(p.timestamp * 1000)
    return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  })

  const isSpeed = wanChartMode.value === 'speed'
  wanChart.setOption({
    animationDuration: 260,
    animationDurationUpdate: 180,
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'axis',
      backgroundColor: pal.tooltipBg,
      borderColor: pal.tooltipBorder,
      textStyle: { color: pal.textStrong, fontSize: 12 },
    },
    legend: {
      top: 0,
      textStyle: { color: pal.textMuted, fontSize: 11 },
    },
    grid: { left: 46, right: 16, top: 32, bottom: 26 },
    xAxis: {
      type: 'category',
      data: labels,
      axisLine: { lineStyle: { color: pal.gridLine } },
      axisLabel: { color: pal.textMuted, fontSize: 11 },
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { color: pal.gridLine } },
      axisLabel: { color: pal.textMuted, fontSize: 11 },
    },
    series: isSpeed
      ? [
          {
            name: '平均下行 (KB/s)',
            type: 'line',
            smooth: true,
            showSymbol: false,
            data: list.map((p) => p.avg_download_kbps),
            lineStyle: { width: 2, color: pal.downColor },
            itemStyle: { color: pal.downColor },
            areaStyle: { opacity: 0.16, color: pal.downColor },
          },
          {
            name: '平均上行 (KB/s)',
            type: 'line',
            smooth: true,
            showSymbol: false,
            data: list.map((p) => p.avg_upload_kbps),
            lineStyle: { width: 2, color: pal.upColor },
            itemStyle: { color: pal.upColor },
            areaStyle: { opacity: 0.14, color: pal.upColor },
          },
        ]
      : [
          {
            name: '平均延迟 (ms)',
            type: 'line',
            smooth: true,
            showSymbol: false,
            data: list.map((p) => p.avg_rtt_ms),
            lineStyle: { width: 2, color: pal.accent },
            itemStyle: { color: pal.accent },
            areaStyle: { opacity: 0.16, color: pal.accent },
          },
          {
            name: '最大延迟 (ms)',
            type: 'line',
            smooth: true,
            showSymbol: false,
            data: list.map((p) => p.max_rtt_ms),
            lineStyle: { width: 1.5, type: 'dashed', color: pal.warn },
            itemStyle: { color: pal.warn },
          },
        ],
  }, { replaceMerge: ['series'] })
}

function renderLoadChart() {
  if (!loadChartEl.value || !loadData.value) return
  if (loadChart && loadChart.getDom() !== loadChartEl.value) {
    loadChart.dispose()
    loadChart = null
  }
  if (!loadChart) {
    loadChart = echarts.init(loadChartEl.value)
    resizeObserver?.observe(loadChartEl.value)
  }
  const pal = palette.value
  const rows = loadData.value.series || []
  const labels = rows.map((r) => {
    const ts = Number(r.timestamp ?? r.time ?? 0)
    if (ts > 0) {
      const d = new Date(ts * 1000)
      return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
    }
    return String(r.label ?? '')
  })

  const parsePoint = (raw: unknown): number => parseFloat(String(raw ?? 0).replace('%', '')) || 0
  const hasDualNetwork =
    loadMetric.value === 'network' &&
    rows.some((r) => r.upload !== undefined || r.avg_upload !== undefined || r.download !== undefined || r.avg_download !== undefined)

  const series = hasDualNetwork
    ? [
        {
          name: '下行负载',
          type: 'line',
          smooth: true,
          showSymbol: false,
          data: rows.map((r) => parsePoint(r.download ?? r.avg_download ?? 0)),
          lineStyle: { width: 2.2, color: pal.downColor },
          itemStyle: { color: pal.downColor },
          areaStyle: { opacity: 0.18, color: pal.downColor },
        },
        {
          name: '上行负载',
          type: 'line',
          smooth: true,
          showSymbol: false,
          data: rows.map((r) => parsePoint(r.upload ?? r.avg_upload ?? 0)),
          lineStyle: { width: 2, color: pal.upColor },
          itemStyle: { color: pal.upColor },
          areaStyle: { opacity: 0.14, color: pal.upColor },
        },
      ]
    : [
        {
          name: loadMetric.value.toUpperCase(),
          type: 'line',
          smooth: true,
          showSymbol: false,
          data: rows.map((r) =>
            parsePoint(
              r.value ??
                r.avg ??
                r.cpu ??
                r.memory ??
                r.connect_num ??
                r.terminals ??
                r.user_num ??
                r.online ??
                r.download ??
                r.count ??
                0
            )
          ),
          lineStyle: { width: 2.2, color: pal.downColor },
          itemStyle: { color: pal.downColor },
          areaStyle: { opacity: 0.18, color: pal.downColor },
        },
      ]

  loadChart.setOption({
    animationDuration: 260,
    animationDurationUpdate: 180,
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'axis',
      backgroundColor: pal.tooltipBg,
      borderColor: pal.tooltipBorder,
      textStyle: { color: pal.textStrong, fontSize: 12 },
    },
    grid: { left: 46, right: 16, top: hasDualNetwork ? 30 : 20, bottom: 26 },
    legend: hasDualNetwork ? { top: 0, textStyle: { color: pal.textMuted, fontSize: 11 } } : undefined,
    xAxis: {
      type: 'category',
      data: labels,
      axisLine: { lineStyle: { color: pal.gridLine } },
      axisLabel: { color: pal.textMuted, fontSize: 11 },
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { color: pal.gridLine } },
      axisLabel: { color: pal.textMuted, fontSize: 11 },
    },
    series,
  }, true)
}

function extractErr(e: unknown): string {
  if (e && typeof e === 'object' && 'response' in e) {
    const resp = (e as { response?: { data?: { message?: string } } }).response
    if (resp?.data?.message) return resp.data.message
  }
  return e instanceof Error ? e.message : String(e)
}

watch([wanChartMode, palette], () => renderWanChart(true))
watch([wanChartEl, () => overview.value?.wan_history], () => {
  if (wanChartEl.value && overview.value) renderWanChart(true)
})
watch([loadMetric, loadRange], () => void fetchLoadOnly())
watch([loadChartEl, loadData, palette], () => {
  if (loadChartEl.value && loadData.value) renderLoadChart()
})
watch(
  () => [...sectionOrder.value],
  async () => {
    await nextTick()
    renderWanChart(true)
    renderLoadChart()
    wanChart?.resize()
    loadChart?.resize()
  }
)
watch(routerEpoch, () => {
  lastWanRenderSig = ''
  loadData.value = null
  void refreshAll()
})

let hasMounted = false

function startPolling() {
  if (pollTimer) return
  pollTimer = setInterval(() => {
    if (typeof document !== 'undefined' && document.hidden) return
    void fetchOverviewOnly(true, true)
    if (!loadData.value && activeLoadReqs === 0) {
      void fetchLoadOnly()
    }
  }, 4000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

function handleVisibilityChange() {
  if (typeof document !== 'undefined' && !document.hidden) {
    void fetchOverviewOnly(false)
  }
}

onMounted(async () => {
  hasMounted = true
  resizeObserver = new ResizeObserver(() => {
    if (resizeRaf) cancelAnimationFrame(resizeRaf)
    resizeRaf = requestAnimationFrame(() => {
      wanChart?.resize()
      loadChart?.resize()
    })
  })
  await refreshAll()

  startPolling()
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

onActivated(() => {
  if (hasMounted) {
    void fetchOverviewOnly()
    startPolling()
    document.addEventListener('visibilitychange', handleVisibilityChange)
    void nextTick(() => {
      wanChart?.resize()
      loadChart?.resize()
    })
  }
})

onDeactivated(() => {
  stopPolling()
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})

onUnmounted(() => {
  stopPolling()
  document.removeEventListener('visibilitychange', handleVisibilityChange)
  if (resizeRaf) cancelAnimationFrame(resizeRaf)
  resizeObserver?.disconnect()
  wanChart?.dispose()
  loadChart?.dispose()
})
</script>

<style scoped>
.overview-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.identity-banner {
  padding: 18px 20px;
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.host-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}

.host-row h1 {
  margin: 0;
  font-size: 22px;
  font-weight: 800;
}

.live-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--status-up);
  box-shadow: 0 0 6px color-mix(in srgb, var(--status-up) 55%, transparent);
}

.meta-line {
  margin-top: 6px;
  font-size: 12px;
  color: var(--text-muted);
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.identity-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.kpi-card {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.kpi-card.clickable {
  cursor: pointer;
}

.kpi-card.clickable:hover {
  transform: translateY(-2px);
}

.kpi-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
}

.kpi-value {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.kpi-value strong {
  font-size: 26px;
  font-weight: 800;
  color: var(--text-strong);
}

.kpi-value span {
  font-size: 12px;
  color: var(--text-muted);
}

.kpi-sub {
  font-size: 11px;
  color: var(--text-muted);
}

.ic-up { color: var(--chart-up); }
.ic-down { color: var(--chart-down); }
.ic-accent { color: var(--accent-text); }

.two-col-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.section-card {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-width: 0;
}

.card-title-row {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-muted);
}

.card-title h3 {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
  color: var(--text-strong);
}

.hw-split {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.hw-block {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.hw-label-row {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--text-normal);
}

.core-bars {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.core-row {
  display: grid;
  grid-template-columns: 48px 1fr 44px;
  align-items: center;
  gap: 8px;
  font-size: 11px;
}

.core-idx, .core-val {
  color: var(--text-muted);
}

.bar-track {
  height: 7px;
  border-radius: 999px;
  background: var(--surface-input);
  overflow: hidden;
}

.bar-track.lg {
  height: 11px;
}

.bar-fill {
  height: 100%;
  border-radius: 999px;
  transition: width 0.3s ease;
}

.bar-fill.accent { background: var(--accent-primary); }
.bar-fill.up { background: var(--status-up); }

.mem-details {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
  font-size: 11px;
  color: var(--text-muted);
}

.port-matrix {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(185px, 1fr));
  gap: 10px;
}

.port-slot {
  padding: 10px 12px;
  border-radius: 12px;
  border: 1px solid var(--divider-line);
  background: var(--surface-input);
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.port-slot.linked {
  border-color: color-mix(in srgb, var(--status-up) 30%, var(--divider-line));
}

.port-top {
  display: flex;
  align-items: center;
  gap: 7px;
}

.port-led {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--text-faint);
}

.port-led.on {
  background: var(--status-up);
  box-shadow: 0 0 6px color-mix(in srgb, var(--status-up) 55%, transparent);
}

.port-mid {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: var(--text-normal);
}

.port-sub {
  font-size: 10px;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.pill-group {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.mini-pill,
.mini-select {
  min-height: 28px;
  padding: 0 10px;
  border-radius: 8px;
  border: 1px solid var(--divider-line);
  background: var(--surface-btn);
  color: var(--text-normal);
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
}

.mini-pill.active {
  border-color: var(--accent-primary);
  background: var(--rail-active-bg);
  color: var(--text-strong);
}

.chart-stage {
  position: relative;
  width: 100%;
  height: 240px;
}

.chart-box {
  width: 100%;
  height: 240px;
}

.chart-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: var(--text-muted);
  pointer-events: none;
}

.layout-edit-bar {
  padding: 10px 14px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  border-color: color-mix(in srgb, var(--accent-primary) 38%, transparent);
  background: var(--accent-primary-soft);
}

.layout-edit-copy {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12.5px;
  font-weight: 600;
  color: var(--text-strong);
}

.layout-edit-copy svg {
  color: var(--accent-text);
  flex-shrink: 0;
}

.layout-edit-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.card-drag-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  padding-bottom: 6px;
  border-bottom: 1px dashed var(--divider-line);
}

.title-right-group {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.drag-ctl-group,
.step-btns {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.drag-handle-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 999px;
  border: 1px dashed color-mix(in srgb, var(--accent-primary) 42%, var(--border-control));
  background: var(--accent-primary-soft);
  color: var(--accent-text);
  font-size: 11px;
  font-weight: 600;
  cursor: grab;
  user-select: none;
}

.drag-handle-pill:active {
  cursor: grabbing;
}

.drag-step-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 7px;
  border: 1px solid var(--border-control);
  background: var(--surface-btn);
  color: var(--text-normal);
  cursor: pointer;
}

.drag-step-btn:hover:not(:disabled) {
  border-color: var(--accent-primary);
  color: var(--accent-text);
}

.drag-step-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.layout-editable {
  border-style: dashed;
  border-color: color-mix(in srgb, var(--accent-primary) 34%, var(--border-subtle));
  cursor: grab;
}

.layout-editable.is-dragging {
  opacity: 0.45;
  transform: scale(0.985);
}

.layout-editable.is-drop-target {
  border-style: solid;
  border-color: var(--accent-primary);
  box-shadow:
    0 0 0 2px var(--accent-primary-soft),
    var(--shadow-elevated);
}

.wan-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.wan-row {
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid var(--divider-line);
  background: var(--surface-input);
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.wan-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.wan-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.wan-comment {
  font-size: 12px;
  color: var(--text-muted);
}

.wan-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  font-size: 11px;
  color: var(--text-muted);
}

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

.notice-row {
  justify-content: space-between;
}

.notice-dismiss {
  flex-shrink: 0;
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 999px;
  background: transparent;
  color: inherit;
  font-size: 16px;
  line-height: 1;
  cursor: pointer;
  opacity: 0.78;
}

.notice-dismiss:hover {
  opacity: 1;
  background: rgba(128, 128, 128, 0.14);
}

.spin {
  animation: spin 0.9s linear infinite;
}

@media (max-width: 1180px) {
  .kpi-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .two-col-grid,
  .hw-split {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 600px) {
  .identity-banner {
    padding: 14px;
  }

  .host-row h1 {
    font-size: 18px;
  }

  .identity-actions {
    width: 100%;
    gap: 6px;
  }

  .kpi-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
  }

  .kpi-card {
    padding: 12px;
    gap: 6px;
  }

  .kpi-value strong {
    font-size: 20px;
  }

  .port-matrix {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
  }

  .chart-stage,
  .chart-box {
    height: 200px;
  }
}
</style>

<template>
  <header class="identity-banner v4-panel">
    <div class="identity-left">
      <div class="host-row">
        <span class="live-dot" />
        <h1>{{ overview?.router.hostname || routerName || 'iKuai V4 路由器' }}</h1>
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
        @click="emit('quick-action', { key: 'ntp-sync', okText: 'NTP 对时指令已下发' })"
      >
        <Clock :size="13" />
        <span>同步 NTP</span>
      </button>
      <button
        type="button"
        class="v4-btn btn-xs"
        :disabled="actionBusy || readOnly"
        :title="readOnly ? '只读保护中' : '创建路由器配置快照备份'"
        @click="emit('quick-action', { key: 'backup-create', okText: '已创建配置快照备份' })"
      >
        <Archive :size="13" />
        <span>立即备份</span>
      </button>
      <button
        type="button"
        class="v4-btn btn-xs"
        :disabled="actionBusy || readOnly"
        :title="readOnly ? '只读保护中' : '在线检查最新系统固件'"
        @click="emit('quick-action', { key: 'upgrade-check', okText: '固件更新检查已触发' })"
      >
        <DownloadCloud :size="13" />
        <span>检查更新</span>
      </button>
      <button
        type="button"
        class="v4-btn btn-xs"
        :disabled="actionBusy || readOnly"
        :title="readOnly ? '只读保护中' : '重启局域网 DHCP 服务'"
        @click="emit('quick-action', { key: 'dhcp-restart', okText: 'DHCP 服务已重启' })"
      >
        <RotateCw :size="13" />
        <span>重启 DHCP</span>
      </button>
      <button
        type="button"
        class="v4-btn btn-xs"
        :class="{ 'v4-btn-primary': !layoutLocked }"
        :title="layoutLocked ? '解锁首页卡片拖动调整顺序' : '完成调整并固定当前首页布局'"
        @click="emit('toggle-layout')"
      >
        <Pin v-if="!layoutLocked" :size="13" />
        <LayoutGrid v-else :size="13" />
        <span>{{ layoutLocked ? '调整布局' : '固定布局' }}</span>
      </button>
      <button type="button" class="v4-btn" :disabled="loading" @click="emit('refresh')">
        <RefreshCw :size="14" :class="{ spin: loading }" />
        <span>立即刷新</span>
      </button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { Archive, Clock, DownloadCloud, LayoutGrid, Pin, RefreshCw, RotateCw } from 'lucide-vue-next'
import type { OverviewDTO } from '@/api/v4'
import { formatUptime } from '@/composables/useFormatters'

defineProps<{
  overview: OverviewDTO | null
  routerName?: string
  readOnly: boolean
  actionBusy: boolean
  loading: boolean
  layoutLocked: boolean
}>()

const emit = defineEmits<{
  (e: 'quick-action', payload: { key: string; okText: string }): void
  (e: 'toggle-layout'): void
  (e: 'refresh'): void
}>()
</script>

<style scoped>
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

.identity-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.spin {
  animation: spin 0.9s linear infinite;
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
}
</style>

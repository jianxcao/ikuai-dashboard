<template>
  <div class="client-list-container glass-card">
    <div class="list-scroll">
      <div class="list-header client-grid" role="row">
        <div class="header-col">设备</div>
        <button type="button" class="header-col sortable" @click="handleSort('ip')">
          <span>IP 地址</span>
          <component :is="getSortIcon('ip')" :size="13" class="sort-icon" />
        </button>
        <button type="button" class="header-col sortable numeric" @click="handleSort('upload')">
          <span>上传</span>
          <component :is="getSortIcon('upload')" :size="13" class="sort-icon" />
        </button>
        <button type="button" class="header-col sortable numeric" @click="handleSort('download')">
          <span>下载</span>
          <component :is="getSortIcon('download')" :size="13" class="sort-icon" />
        </button>
        <button type="button" class="header-col sortable numeric" @click="handleSort('connections')">
          <span>连接数</span>
          <component :is="getSortIcon('connections')" :size="13" class="sort-icon" />
        </button>
        <button type="button" class="header-col sortable numeric" @click="handleSort('total')">
          <span>累计流量</span>
          <component :is="getSortIcon('total')" :size="13" class="sort-icon" />
        </button>
      </div>

      <div class="list-body">
        <ClientCard
          v-for="client in clients"
          :key="client.mac"
          :client="client"
        />

        <div v-if="!loading && clients.length === 0" class="empty-state">
          <div class="empty-icon">!</div>
          <p>没有匹配的局域网设备</p>
        </div>

        <div v-if="loading && clients.length === 0" class="loading-state">
          <div class="spinner"></div>
          <p>扫描局域网设备中...</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ArrowDownWideNarrow, ArrowUpNarrowWide, ChevronsUpDown } from 'lucide-vue-next'
import ClientCard from './ClientCard.vue'
import type { ClientDTO } from '@/api/monitor'

const props = withDefaults(defineProps<{
  clients?: ClientDTO[]
  loading?: boolean
  sortBy?: string
}>(), {
  clients: () => [],
  loading: false,
  sortBy: 'default'
})

const emit = defineEmits<{
  (e: 'update:sortBy', val: string): void
}>()

function handleSort(key: string) {
  let next = `${key}_desc`
  if (props.sortBy === `${key}_desc`) {
    next = `${key}_asc`
  } else if (props.sortBy === `${key}_asc`) {
    next = 'default'
  }
  emit('update:sortBy', next)
}

function getSortIcon(key: string) {
  if (props.sortBy === `${key}_desc`) return ArrowDownWideNarrow
  if (props.sortBy === `${key}_asc`) return ArrowUpNarrowWide
  return ChevronsUpDown
}
</script>

<style scoped>
.client-list-container {
  --client-list-columns: minmax(260px, 1.08fr) minmax(340px, 1.32fr) minmax(96px, 0.46fr) minmax(96px, 0.46fr) minmax(80px, 0.34fr) minmax(130px, 0.56fr);

  display: flex;
  flex-direction: column;
  height: calc(100vh - 200px);
  padding: 0;
  overflow: hidden;
}

.list-scroll {
  flex: 1;
  min-height: 0;
  overflow: auto;
  scrollbar-gutter: stable both-edges;
}

.client-grid {
  display: grid;
  grid-template-columns: var(--client-list-columns);
  column-gap: 14px;
  align-items: center;
  min-width: 1072px;
}

.list-header {
  position: sticky;
  top: 0;
  z-index: 2;
  padding: 14px 24px;
  border-bottom: 1px solid var(--glass-border);
  background: var(--table-header-bg);
  backdrop-filter: blur(22px) saturate(160%);
  -webkit-backdrop-filter: blur(22px) saturate(160%);
}

.header-col {
  min-width: 0;
  font-size: 12px;
  font-weight: 700;
  color: var(--text-secondary);
  letter-spacing: 0;
}

button.header-col {
  appearance: none;
  border: 0;
  background: transparent;
  padding: 0;
  font: inherit;
  text-align: left;
}

.sortable {
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: color 0.3s ease;
  user-select: none;
}

.sortable:hover {
  color: var(--text-primary);
}

.numeric {
  justify-content: flex-end;
  text-align: right;
}

.sort-icon {
  flex: 0 0 auto;
  color: var(--system-blue);
  opacity: 0.86;
}

.list-body {
  display: flex;
  flex-direction: column;
  min-width: 1072px;
}

.empty-state, .loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: var(--text-tertiary);
  gap: 16px;
  font-size: 14px;
  border-bottom: 1px solid var(--control-border);
}

.empty-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: var(--control-bg);
  border: 1px solid var(--control-border);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 20px;
}

@media (max-width: 920px) {
  .client-list-container {
    --client-list-columns: 1fr;

    height: auto;
    min-height: 420px;
    overflow: visible;
  }

  .list-scroll {
    overflow: visible;
    scrollbar-gutter: auto;
  }

  .client-grid,
  .list-body {
    min-width: 0;
  }

  .list-header {
    display: none;
  }
}
</style>

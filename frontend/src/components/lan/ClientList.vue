<template>
  <div class="client-list-container glass-card">
    <!-- 列表表头 -->
    <div class="list-header">
      <div class="header-col" style="width: 260px">设备标识 (MAC)</div>
      <div class="header-col sortable" style="width: 280px" @click="handleSort('ip')">
        IP 地址 <span class="sort-icon">{{ getSortIcon('ip') }}</span>
      </div>
      <div class="header-col" style="flex: 1; display: flex; gap: 16px">
        <div class="sortable" style="width: 100px" @click="handleSort('upload')">
          上传 <span class="sort-icon">{{ getSortIcon('upload') }}</span>
        </div>
        <div class="sortable" style="width: 100px" @click="handleSort('download')">
          下载 <span class="sort-icon">{{ getSortIcon('download') }}</span>
        </div>
        <div class="sortable" style="width: 80px" @click="handleSort('connections')">
          连接数 <span class="sort-icon">{{ getSortIcon('connections') }}</span>
        </div>
      </div>
      <div class="header-col sortable" style="width: 120px; justify-content: flex-end" @click="handleSort('total')">
        总计消耗 <span class="sort-icon">{{ getSortIcon('total') }}</span>
      </div>
    </div>

    <!-- 列表滚动区 -->
    <div class="list-body">
      <ClientCard 
        v-for="client in clients" 
        :key="client.mac" 
        :client="client" 
      />
      
      <!-- 空状态 -->
      <div v-if="!loading && clients.length === 0" class="empty-state">
        <div class="empty-icon">!</div>
        <p>没有匹配的局域网设备</p>
      </div>
      
      <!-- 加载中 -->
      <div v-if="loading && clients.length === 0" class="loading-state">
        <div class="spinner"></div>
        <p>扫描局域网设备中...</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
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
  if (props.sortBy === `${key}_desc`) return '↓'
  if (props.sortBy === `${key}_asc`) return '↑'
  return ''
}
</script>

<style scoped>
.client-list-container {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 200px); /* 让列表铺满剩余高度并支持内部滚动 */
  padding: 0;
  overflow: hidden;
}

.list-header {
  display: flex;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid var(--glass-border);
  background: var(--table-header-bg);
}

.header-col {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  letter-spacing: 0.04em;
}

.sortable {
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  transition: color 0.3s ease;
  user-select: none;
}

.sortable:hover {
  color: var(--text-primary);
}

.sort-icon {
  width: 14px;
  text-align: center;
  font-family: ui-monospace, SFMono-Regular, monospace;
  color: var(--system-blue);
  font-weight: 600;
}

.list-body {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

/* 空状态 / 加载状态 */
.empty-state, .loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: var(--text-tertiary);
  gap: 16px;
  font-size: 14px;
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
    height: auto;
    min-height: 420px;
  }

  .list-header {
    display: none;
  }
}
</style>

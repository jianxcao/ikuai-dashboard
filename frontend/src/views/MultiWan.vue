<template>
  <div class="multi-wan-hub">
    <!-- WAN 状态 -->
    <div class="glass-panel section-panel">
      <div class="panel-header">
        <h2 class="panel-title">🌐 物理链路状态</h2>
      </div>
      <div class="panel-content">
        <div v-if="loading" class="loading-text">加载中...</div>
        <div v-else class="grid">
          <div v-for="(wan, i) in data?.wan_status" :key="i" class="wan-card">
            <div class="wan-header">
              <span class="wan-name">{{ wan.name }}</span>
              <span :class="['status-badge', wan.status === 'success' ? 'status-ok' : 'status-err']">
                {{ wan.status }}
              </span>
            </div>
            <div class="wan-details">
              <div class="detail-row">
                <span class="label">IP</span>
                <span class="val">{{ wan.ip }}</span>
              </div>
              <div class="detail-row">
                <span class="label">协议</span>
                <span class="val">{{ wan.proto }}</span>
              </div>
              <div class="detail-row" v-if="wan.comment">
                <span class="label">备注</span>
                <span class="val">{{ wan.comment }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 路由与分流策略 -->
    <div class="glass-panel section-panel mt-6">
      <div class="panel-header">
        <h2 class="panel-title">🔀 分流策略看板</h2>
      </div>
      <div class="panel-content">
        <div v-if="loading" class="loading-text">加载中...</div>
        <div v-else-if="data?.routes.length === 0" class="empty-text">未配置任何分流策略。</div>
        <table v-else class="data-table">
          <thead>
            <tr>
              <th>策略类型</th>
              <th>目标匹配</th>
              <th>出口链路</th>
              <th>状态</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(route, i) in data?.routes" :key="i">
              <td>
                <span class="route-type">{{ route.type }}</span>
              </td>
              <td>{{ route.target }}</td>
              <td><span class="route-iface">{{ route.interface }}</span></td>
              <td>
                <span :class="['status-dot', route.enabled ? 'enabled' : 'disabled']"></span>
                {{ route.enabled ? '启用' : '停用' }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import axios from 'axios'

const data = ref<any>(null)
const loading = ref(false)
const error = ref('')

const fetchData = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await axios.get('http://localhost:8080/api/v1/monitor/multi-wan')
    if (res.data.code === 200) {
      data.value = res.data.data
    } else {
      error.value = res.data.message || '获取数据失败'
    }
  } catch (err: any) {
    error.value = err.message || '请求错误'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.multi-wan-hub {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.panel-header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--glass-border);
}

.panel-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.panel-content {
  padding: 20px;
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(100%, 280px), 1fr));
  gap: 16px;
}

.wan-card {
  background: var(--control-bg);
  border: 1px solid var(--glass-border);
  border-radius: 12px;
  padding: 16px;
}

.wan-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.wan-name {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text-primary);
  text-transform: uppercase;
}

.status-badge {
  padding: 4px 10px;
  border-radius: 100px;
  font-size: 0.8rem;
  font-weight: 500;
}

.status-ok {
  background: var(--system-green-dim);
  color: var(--system-green);
}

.status-err {
  background: var(--system-red-dim);
  color: var(--system-red);
}

.wan-details {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-size: 0.9rem;
}

.detail-row .label {
  color: var(--text-secondary);
}

.detail-row .val {
  font-family: monospace;
  color: var(--text-primary);
  overflow-wrap: anywhere;
  text-align: right;
}

/* Table Styles */
.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th, .data-table td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid var(--control-border);
}

.data-table th {
  color: var(--text-secondary);
  font-weight: 500;
  font-size: 0.9rem;
}

.route-type {
  background: var(--system-blue-dim);
  color: var(--system-blue);
  padding: 4px 8px;
  border-radius: 6px;
  font-size: 0.8rem;
}

.route-iface {
  font-family: monospace;
  background: var(--control-bg);
  border: 1px solid var(--control-border);
  padding: 2px 6px;
  border-radius: 4px;
}

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 6px;
}

.status-dot.enabled {
  background: var(--system-green);
}

.status-dot.disabled {
  background: var(--system-red);
}

.loading-text, .empty-text {
  text-align: center;
  color: var(--text-secondary);
  padding: 30px;
}

@media (max-width: 760px) {
  .panel-content {
    padding: 16px;
    overflow-x: auto;
  }

  .panel-header {
    padding: 16px;
  }

  .data-table {
    min-width: 560px;
  }
}

@media (max-width: 420px) {
  .wan-header,
  .detail-row {
    align-items: flex-start;
    flex-direction: column;
  }

  .detail-row .val {
    text-align: left;
  }
}
</style>

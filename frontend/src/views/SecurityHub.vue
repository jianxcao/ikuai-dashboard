<template>
  <div class="security-hub">
    <!-- 高危端口映射 -->
    <div class="glass-panel section-panel">
      <div class="panel-header">
        <h2 class="panel-title">⚠️ 外网暴露雷达 (高危端口映射)</h2>
      </div>
      <div class="panel-content">
        <div v-if="loading" class="loading-text">加载中...</div>
        <div v-else-if="error" class="error-text">{{ error }}</div>
        <div v-else-if="data?.high_risk_ports.length === 0" class="empty-text">当前无高风险端口暴露。</div>
        <div v-else class="grid">
          <div v-for="(port, i) in data?.high_risk_ports" :key="i" class="port-card danger-card">
            <div class="port-info">
              <h3 class="port-name">{{ port.name }}</h3>
              <p class="port-proto">{{ port.protocol }}</p>
            </div>
            <div class="port-mapping">
              <span class="ext-port">Ext: {{ port.ext_port }}</span>
              <span class="arrow">→</span>
              <span class="int-port">{{ port.int_ip }}:{{ port.int_port }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 异常设备 -->
    <div class="glass-panel section-panel mt-6">
      <div class="panel-header">
        <h2 class="panel-title">🛡️ 并发超限 / 异常设备</h2>
      </div>
      <div class="panel-content">
        <div v-if="loading" class="loading-text">加载中...</div>
        <div v-else-if="data?.abnormal_devices.length === 0" class="empty-text">当前设备运行正常，无高风险行为。</div>
        <div v-else class="list-container">
          <div v-for="(dev, i) in data?.abnormal_devices" :key="i" class="list-item">
            <div class="dev-info">
              <span class="dev-hostname">{{ dev.hostname }}</span>
              <span class="dev-mac">{{ dev.mac }}</span>
            </div>
            <div class="dev-metrics">
              <div class="metric"><span class="label">并发连接</span><span class="val text-red">{{ dev.connections }}</span></div>
              <div class="metric"><span class="label">总上行</span><span class="val">{{ formatBytes(dev.total_up) }}</span></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { fetchSecurityHub, type SecurityHubData } from '@/api/monitor'

const data = ref<SecurityHubData | null>(null)
const loading = ref(false)
const error = ref('')

const fetchData = async () => {
  loading.value = true
  error.value = ''
  try {
    data.value = await fetchSecurityHub()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '请求错误'
  } finally {
    loading.value = false
  }
}

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.security-hub {
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
  grid-template-columns: repeat(auto-fill, minmax(min(100%, 300px), 1fr));
  gap: 16px;
}

.port-card {
  background: var(--system-red-dim);
  border: 1px solid var(--system-red);
  border-radius: 12px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.port-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.port-name {
  font-weight: 600;
  color: var(--system-red);
}

.port-proto {
  background: var(--control-bg);
  border: 1px solid var(--control-border);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.port-mapping {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  font-family: monospace;
  color: var(--text-primary);
  background: var(--control-bg);
  border: 1px solid var(--control-border);
  padding: 8px 12px;
  border-radius: 6px;
  overflow-wrap: anywhere;
}

.arrow {
  color: var(--text-secondary);
}

.list-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.list-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  background: var(--control-bg);
  padding: 16px;
  border-radius: 8px;
  border: 1px solid var(--glass-border);
}

.dev-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.dev-hostname {
  font-weight: 500;
  color: var(--text-primary);
}

.dev-mac {
  font-family: monospace;
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.dev-metrics {
  display: flex;
  gap: 24px;
}

.metric {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.label {
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.val {
  font-weight: 600;
  font-family: monospace;
}

.text-red {
  color: var(--system-red);
}

.loading-text, .error-text, .empty-text {
  text-align: center;
  color: var(--text-secondary);
  padding: 30px;
}

@media (max-width: 640px) {
  .panel-header,
  .panel-content {
    padding: 16px;
  }

  .port-info,
  .port-mapping,
  .list-item,
  .dev-metrics {
    align-items: flex-start;
    flex-direction: column;
  }

  .dev-metrics {
    gap: 10px;
    width: 100%;
  }

  .metric {
    align-items: flex-start;
  }
}
</style>

<template>
  <div class="page-container">
    <div v-if="error" class="error-banner">
      <span class="error-icon">!</span>
      <span>获取数据失败：{{ error }}</span>
    </div>

    <!-- 骨架屏 / Loading -->
    <div v-if="loading && !wanStatus.length" class="loading-state">
      <div class="spinner"></div>
      <p>正在建立加密链路连接...</p>
    </div>

    <template v-else>
      <SummaryCards :summary="summary" />
      <TrafficTables :wanStatus="wanStatus" :trafficDetails="trafficDetails" />
      <TrafficChart 
        :chartLabels="chartLabels" 
        :chartUpload="chartUpload" 
        :chartDownload="chartDownload"
        :loading="loading"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import SummaryCards from '@/components/monitor/SummaryCards.vue'
import TrafficTables from '@/components/monitor/TrafficTables.vue'
import TrafficChart from '@/components/monitor/TrafficChart.vue'
import { useMonitor } from '@/composables/useMonitor'

const {
  loading,
  error,
  summary,
  wanStatus,
  trafficDetails,
  chartLabels,
  chartUpload,
  chartDownload
} = useMonitor()
</script>

<style scoped>
.page-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
  animation: fade-in 0.4s ease-out;
}

@keyframes fade-in {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.loading-state {
  min-height: 400px;
}
</style>

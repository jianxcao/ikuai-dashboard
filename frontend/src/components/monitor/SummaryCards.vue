<template>
  <div class="summary-cards">
    <div class="summary-card glass-card upload">
      <div class="icon-wrapper">
        <div class="card-icon">▲</div>
      </div>
      <div class="card-content">
        <div class="card-label">当前上传</div>
        <div class="card-value font-mono">
          <span class="metric-number">{{ formatSpeedParts(summary.upload_speed).value }}</span>
          <span class="metric-unit">{{ formatSpeedParts(summary.upload_speed).unit }}</span>
        </div>
      </div>
    </div>

    <div class="summary-card glass-card download">
      <div class="icon-wrapper">
        <div class="card-icon">▼</div>
      </div>
      <div class="card-content">
        <div class="card-label">当前下载</div>
        <div class="card-value font-mono">
          <span class="metric-number">{{ formatSpeedParts(summary.download_speed).value }}</span>
          <span class="metric-unit">{{ formatSpeedParts(summary.download_speed).unit }}</span>
        </div>
      </div>
    </div>

    <div class="summary-card glass-card connections">
      <div class="icon-wrapper">
        <div class="card-icon">◎</div>
      </div>
      <div class="card-content">
        <div class="card-label">总连接数</div>
        <div class="card-value font-mono">
          <span class="metric-number">{{ formatCountCompact(summary.total_connections) }}</span>
        </div>
      </div>
    </div>

    <div class="summary-card glass-card online">
      <div class="icon-wrapper">
        <div class="card-icon">⬡</div>
      </div>
      <div class="card-content">
        <div class="card-label">在线终端</div>
        <div class="card-value font-mono">{{ summary.online_users }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatCountCompact, formatSpeedParts } from '@/composables/useFormatters'

defineProps({
  summary: {
    type: Object,
    default: () => ({ upload_speed: 0, download_speed: 0, total_connections: 0, online_users: 0 })
  }
})
</script>

<style scoped>
.summary-cards {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 18px;
}

@media (max-width: 1240px) {
  .summary-cards {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.summary-card {
  min-height: 156px;
  padding: 22px 22px 20px;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
  position: relative;
  overflow: hidden;
  min-width: 0;
}

/* 上传卡片 */
.summary-card.upload .icon-wrapper {
  background: var(--system-green-dim);
  color: var(--system-green);
}

/* 下载卡片 */
.summary-card.download .icon-wrapper {
  background: var(--system-blue-dim);
  color: var(--system-blue);
}

/* 连接数卡片 */
.summary-card.connections .icon-wrapper {
  background: var(--system-orange-dim);
  color: var(--system-orange);
}

/* 在线终端卡片 */
.summary-card.online .icon-wrapper {
  background: var(--system-red-dim);
  color: var(--system-red);
}

.icon-wrapper {
  width: 52px;
  height: 52px;
  border-radius: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
  border: 1px solid var(--control-border);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.18);
}

.card-content {
  width: 100%;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.card-label {
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.card-value {
  display: flex;
  align-items: baseline;
  gap: 8px;
  min-width: 0;
  font-size: clamp(1.9rem, 2vw, 2.45rem);
  font-weight: 640;
  letter-spacing: 0;
  color: var(--text-primary);
  white-space: nowrap;
  line-height: 1;
}

.metric-number {
  min-width: 0;
  flex: 0 1 auto;
}

.metric-unit {
  flex: 0 0 auto;
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 700;
  line-height: 1;
}

@media (max-width: 560px) {
  .summary-cards {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
  }

  .summary-card {
    min-height: 136px;
    padding: 18px 16px;
  }

  .card-value {
    white-space: normal;
    align-items: flex-end;
    gap: 4px;
  }

  .metric-unit {
    font-size: 12px;
  }
}

@media (max-width: 420px) {
  .summary-cards {
    grid-template-columns: 1fr;
  }
}
</style>

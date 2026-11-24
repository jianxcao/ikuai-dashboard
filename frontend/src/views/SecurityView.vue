<template>
  <CatalogWorkspace
    ref="wsRef"
    eyebrow="FIREWALL & SECURITY CENTER"
    title="安全防护与准入中心"
    subtitle="实时汇聚端口暴露面、ACL 防火墙、MAC 过滤模式、L7/URL/域名过滤、二级路由检测与异常终端管控。"
    :tabs="tabs"
    :initialSubResource="initialSubResource"
    :navSeq="navSeq"
  >
    <template #banner>
      <div v-if="hub" class="posture-grid">
        <article class="v4-card posture-card clickable" @click="emit('navigate', { tab: 'network', subResource: 'dnat-rules' })">
          <span>DNAT 外网映射端口</span>
          <strong class="font-mono">{{ hub.high_risk_ports.length }}</strong>
        </article>
        <article class="v4-card posture-card clickable" @click="wsRef?.setTab('acl-rules')">
          <span>ACL 防火墙策略</span>
          <strong class="font-mono">{{ hub.acl_count }}</strong>
        </article>
        <article class="v4-card posture-card clickable" @click="wsRef?.setTab('peerconn-rules')">
          <span>连接数限制规则</span>
          <strong class="font-mono">{{ hub.peerconn_count }}</strong>
        </article>
        <article class="v4-card posture-card clickable" @click="wsRef?.setTab('url-black')">
          <span>URL 黑名单规则</span>
          <strong class="font-mono">{{ hub.url_black_count }}</strong>
        </article>
        <article
          class="v4-card posture-card clickable"
          :class="{ warn: hub.abnormal_devices.length > 0 }"
          @click="emit('navigate', { tab: 'terminals', subResource: 'online' })"
        >
          <span>高连接/高上行终端</span>
          <strong class="font-mono">{{ hub.abnormal_devices.length }}</strong>
        </article>
      </div>
    </template>
  </CatalogWorkspace>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import CatalogWorkspace, { type WorkspaceTab } from '@/components/common/CatalogWorkspace.vue'
import { fetchSecurityHub, type SecurityHubData } from '@/api/monitor'
import { useFleet } from '@/composables/useFleet'

defineProps<{ initialSubResource?: string; navSeq?: number }>()

const emit = defineEmits<{
  (e: 'navigate', payload: { tab: string; subResource?: string }): void
}>()

const { routerEpoch } = useFleet()
const hub = ref<SecurityHubData | null>(null)
const wsRef = ref<InstanceType<typeof CatalogWorkspace> | null>(null)

const tabs: WorkspaceTab[] = [
  { key: 'acl-rules', label: 'ACL 访问控制' },
  { key: 'mac-mode', label: 'MAC 过滤模式', kind: 'singleton' },
  { key: 'mac-rules', label: 'MAC 访问规则' },
  { key: 'l7-rules', label: 'L7 应用管控' },
  { key: 'url-black', label: 'URL 黑名单' },
  { key: 'url-keywords', label: 'URL 关键字' },
  { key: 'url-redirect', label: 'URL 重定向' },
  { key: 'url-replace', label: 'URL 重写' },
  { key: 'domain-blacklist', label: '域名黑名单' },
  { key: 'peerconn-rules', label: '连接数限制' },
  { key: 'terminals', label: '终端标注' },
  { key: 'security-advanced', label: '高级安全防护', kind: 'singleton' },
  { key: 'secondary-route', label: '二级路由检测', kind: 'singleton' },
]

watch(
  routerEpoch,
  async () => {
    try {
      hub.value = await fetchSecurityHub()
    } catch {
      hub.value = null
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.posture-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 12px;
}

.posture-card {
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.posture-card.clickable {
  cursor: pointer;
}

.posture-card.clickable:hover {
  transform: translateY(-2px);
}

.posture-card span {
  font-size: 12px;
  color: var(--text-muted);
}

.posture-card strong {
  font-size: 22px;
  font-weight: 800;
  color: var(--text-strong);
}

.posture-card.warn strong {
  color: var(--status-warn);
}

@media (max-width: 980px) {
  .posture-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .posture-grid {
    gap: 8px;
  }

  .posture-card {
    padding: 11px 12px;
  }

  .posture-card strong {
    font-size: 18px;
  }

  .posture-card:last-child {
    grid-column: 1 / -1;
  }
}
</style>

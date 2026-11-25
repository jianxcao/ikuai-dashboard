<template>
  <CatalogWorkspace
    routeTab="routing"
    eyebrow="ROUTING & QOS SHUNTING"
    title="多线分流路由与 QoS 流控"
    subtitle="全面管理静态路由、域名分流、五元组分流、L7 协议分流、多 WAN 负载均衡、上下行分离及 IP/MAC 限速。"
    :tabs="tabs"
    :initialSubResource="initialSubResource"
    :navSeq="navSeq"
  />
</template>

<script setup lang="ts">
import CatalogWorkspace, { type WorkspaceTab } from '@/components/common/CatalogWorkspace.vue'

defineProps<{ initialSubResource?: string; navSeq?: number }>()

const tabs: WorkspaceTab[] = [
  { key: 'static-routes', label: '静态路由' },
  { key: 'stream-domain', label: '域名分流' },
  { key: 'stream-five-tuple', label: '五元组分流' },
  { key: 'stream-l7', label: 'L7 应用分流' },
  { key: 'stream-load-balance', label: '负载均衡' },
  { key: 'stream-updown', label: '上下行分离' },
  { key: 'qos-ip', label: 'IP 限速 (QoS)', groupLabel: 'QoS' },
  { key: 'qos-mac', label: 'MAC 限速 (QoS)', groupLabel: 'QoS' },
]
</script>

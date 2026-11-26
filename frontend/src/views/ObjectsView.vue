<template>
  <CatalogWorkspace
    routeTab="objects"
    eyebrow="POLICY OBJECT GROUPS"
    title="策略对象分组库"
    subtitle="统一维护 IP、IPv6、MAC、端口、协议、域名与生效时间段分组，供防火墙、分流、流控与无线规则复用，支持一键反查引用。"
    :tabs="tabs"
    :initialSubResource="initialSubResource"
    :navSeq="navSeq"
  />
</template>

<script setup lang="ts">
import CatalogWorkspace, { type WorkspaceTab } from '@/components/common/CatalogWorkspace.vue'

defineProps<{ initialSubResource?: string; navSeq?: number }>()

const tabs: WorkspaceTab[] = [
  { key: 'ip-objects', label: 'IP 对象组' },
  { key: 'ip6-objects', label: 'IPv6 对象组' },
  { key: 'mac-objects', label: 'MAC 对象组' },
  { key: 'port-objects', label: '端口对象组' },
  { key: 'proto-objects', label: '协议对象组' },
  { key: 'domain-objects', label: '域名对象组' },
  { key: 'time-objects', label: '时间段对象组' },
]
</script>

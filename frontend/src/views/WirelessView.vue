<template>
  <CatalogWorkspace
    routeTab="wireless"
    eyebrow="WIRELESS AC & AP CONTROL"
    title="无线 AC 与 AP 管理"
    subtitle="管理内置无线 AC 控制器、受管 AP 设备列表、无线接入黑白名单与无线终端 VLAN 绑定。"
    :tabs="tabs"
    :initialSubResource="initialSubResource"
    :navSeq="navSeq"
  />
</template>

<script setup lang="ts">
import CatalogWorkspace, { type WorkspaceTab } from '@/components/common/CatalogWorkspace.vue'

defineProps<{ initialSubResource?: string; navSeq?: number }>()

const tabs: WorkspaceTab[] = [
  { key: 'wireless-ac', label: 'AC 控制器', kind: 'singleton' },
  { key: 'wireless-aps', label: 'AP 设备列表' },
  { key: 'wireless-blacklist', label: '无线黑白名单' },
  { key: 'wireless-vlan', label: '无线 VLAN 绑定' },
]
</script>

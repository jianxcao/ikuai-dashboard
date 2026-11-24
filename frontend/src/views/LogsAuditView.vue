<template>
  <CatalogWorkspace
    eyebrow="AUDIT & SYSTEM LOGS"
    title="日志审计中心"
    subtitle="支持 10 类爱快 V4 运行与安全日志检索、自动转换可读本地时间并支持确认清空。"
    :tabs="tabs"
    :initialSubResource="initialSubResource"
    :navSeq="navSeq"
  />
</template>

<script setup lang="ts">
import CatalogWorkspace, { type WorkspaceTab } from '@/components/common/CatalogWorkspace.vue'

defineProps<{ initialSubResource?: string; navSeq?: number }>()

const tabs: WorkspaceTab[] = [
  { key: 'log-system', label: '系统日志' },
  { key: 'log-arp', label: 'ARP 日志' },
  { key: 'log-auth', label: '认证日志' },
  { key: 'log-dhcp', label: 'DHCP 日志' },
  { key: 'log-pppoe', label: 'PPPoE 日志' },
  { key: 'log-web', label: 'Web 管理日志' },
  { key: 'log-ddns', label: 'DDNS 日志' },
  { key: 'log-notice', label: '通知日志' },
  { key: 'log-wireless', label: '无线日志' },
  { key: 'log-url-visits', label: 'URL 访问日志' },
]
</script>

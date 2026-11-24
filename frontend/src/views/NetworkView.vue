<template>
  <CatalogWorkspace
    eyebrow="NETWORK & INTERFACES"
    title="网络与接口管理"
    subtitle="覆盖 WAN/LAN/物理网卡、DHCP 地址池/静态绑定/准入控制、DNS 与代理分流、VLAN 及 NAT/DNAT/DMZ。"
    :tabs="tabs"
    :initialSubResource="initialSubResource"
    :navSeq="navSeq"
  />
</template>

<script setup lang="ts">
import CatalogWorkspace, { type WorkspaceTab } from '@/components/common/CatalogWorkspace.vue'

defineProps<{ initialSubResource?: string; navSeq?: number }>()

const tabs: WorkspaceTab[] = [
  { key: 'wan-config', label: 'WAN 配置' },
  { key: 'lan-config', label: 'LAN 配置' },
  { key: 'physical-nics', label: '物理网口' },
  { key: 'wan-vlan', label: 'WAN VLAN' },
  { key: 'dhcp-services', label: 'DHCP 地址池' },
  { key: 'dhcp-static', label: 'DHCP 静态绑定' },
  { key: 'dhcp-leases', label: 'DHCP 租约' },
  { key: 'dhcp-access-mode', label: 'DHCP 准入模式', kind: 'singleton' },
  { key: 'dhcp-access-rules', label: 'DHCP 准入规则' },
  { key: 'dns-config', label: 'DNS 全局设置', kind: 'singleton' },
  { key: 'dns-proxy', label: 'DNS 代理规则' },
  { key: 'vlan', label: 'LAN VLAN' },
  { key: 'dnat-rules', label: '端口映射 (DNAT)' },
  { key: 'nat-rules', label: 'NAT 规则' },
  { key: 'dmz-rules', label: 'DMZ 主机' },
  { key: 'pppoe-config', label: 'PPPoE 服务', kind: 'singleton' },
]
</script>

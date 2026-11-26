<template>
  <CatalogWorkspace
    routeTab="vpn"
    eyebrow="VPN TUNNELS & SERVICES"
    title="VPN 隧道与远程接入"
    subtitle="集中管理 WireGuard (+Peers)、IPSec 站点隧道，以及 OpenVPN / PPTP / L2TP / IKEv2 服务端与客户端。"
    :tabs="tabs"
    :initialSubResource="initialSubResource"
    :navSeq="navSeq"
  />
</template>

<script setup lang="ts">
import CatalogWorkspace, { type WorkspaceTab } from '@/components/common/CatalogWorkspace.vue'

defineProps<{ initialSubResource?: string; navSeq?: number }>()

const tabs: WorkspaceTab[] = [
  { key: 'wireguard', label: 'WireGuard 隧道' },
  { key: 'ipsec-clients', label: 'IPSec 站点隧道' },
  { key: 'openvpn-service', label: 'OpenVPN 服务', kind: 'singleton' },
  { key: 'openvpn-clients', label: 'OpenVPN 客户端' },
  { key: 'pptp-service', label: 'PPTP 服务', kind: 'singleton' },
  { key: 'pptp-clients', label: 'PPTP 客户端' },
  { key: 'l2tp-service', label: 'L2TP 服务', kind: 'singleton' },
  { key: 'l2tp-clients', label: 'L2TP 客户端' },
  { key: 'ikev2-clients', label: 'IKEv2 客户端' },
]
</script>

<template>
  <CatalogWorkspace
    eyebrow="VPN · WIRELESS · OBJECT GROUPS"
    title="VPN 隧道 · 无线 AC · 策略对象库"
    subtitle="集中管理 WireGuard (+Peers)、IPSec、OpenVPN/PPTP/L2TP/IKEv2 服务端与客户端、无线 AC/AP 与 7 类可复用策略对象组。"
    :tabs="tabs"
    :initialSubResource="initialSubResource"
    :navSeq="navSeq"
  />
</template>

<script setup lang="ts">
import CatalogWorkspace, { type WorkspaceTab } from '@/components/common/CatalogWorkspace.vue'

defineProps<{ initialSubResource?: string; navSeq?: number }>()

const tabs: WorkspaceTab[] = [
  { key: 'wireguard', label: 'WireGuard 隧道', groupLabel: 'VPN' },
  { key: 'ipsec-clients', label: 'IPSec 站点隧道', groupLabel: 'VPN' },
  { key: 'openvpn-service', label: 'OpenVPN 服务', kind: 'singleton', groupLabel: 'VPN' },
  { key: 'openvpn-clients', label: 'OpenVPN 客户端', groupLabel: 'VPN' },
  { key: 'pptp-service', label: 'PPTP 服务', kind: 'singleton', groupLabel: 'VPN' },
  { key: 'pptp-clients', label: 'PPTP 客户端', groupLabel: 'VPN' },
  { key: 'l2tp-service', label: 'L2TP 服务', kind: 'singleton', groupLabel: 'VPN' },
  { key: 'l2tp-clients', label: 'L2TP 客户端', groupLabel: 'VPN' },
  { key: 'ikev2-clients', label: 'IKEv2 客户端', groupLabel: 'VPN' },
  { key: 'wireless-ac', label: 'AC 控制器', kind: 'singleton', groupLabel: '无线' },
  { key: 'wireless-aps', label: 'AP 设备列表', groupLabel: '无线' },
  { key: 'wireless-blacklist', label: '无线黑白名单', groupLabel: '无线' },
  { key: 'wireless-vlan', label: '无线 VLAN', groupLabel: '无线' },
  { key: 'ip-objects', label: 'IP 对象组', groupLabel: '对象' },
  { key: 'ip6-objects', label: 'IPv6 对象组', groupLabel: '对象' },
  { key: 'mac-objects', label: 'MAC 对象组', groupLabel: '对象' },
  { key: 'port-objects', label: '端口对象组', groupLabel: '对象' },
  { key: 'proto-objects', label: '协议对象组', groupLabel: '对象' },
  { key: 'domain-objects', label: '域名对象组', groupLabel: '对象' },
  { key: 'time-objects', label: '时间段对象组', groupLabel: '对象' },
]
</script>

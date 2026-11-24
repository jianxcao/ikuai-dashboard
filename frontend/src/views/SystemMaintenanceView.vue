<template>
  <CatalogWorkspace
    eyebrow="SYSTEM & ADVANCED SERVICES"
    title="系统维护与高级服务"
    subtitle="管理主机名与 NTP、SSH/HTTPS 远程访问、BBR 内核调优、CPU 调度、ALG、VRRP、配置备份/自动备份、文件浏览、权限账号及 FTP/HTTP/Samba/SNMPD。"
    :tabs="tabs"
    :initialSubResource="initialSubResource"
    :navSeq="navSeq"
  />
</template>

<script setup lang="ts">
import CatalogWorkspace, { type WorkspaceTab } from '@/components/common/CatalogWorkspace.vue'

defineProps<{ initialSubResource?: string; navSeq?: number }>()

const tabs: WorkspaceTab[] = [
  { key: 'system-basic', label: '基础与 NTP', kind: 'singleton' },
  { key: 'remote-access', label: '远程访问', kind: 'singleton' },
  { key: 'kernel-params', label: '内核调优 (BBR)', kind: 'singleton' },
  { key: 'cpufreq-mode', label: 'CPU 调频', kind: 'singleton' },
  { key: 'alg-config', label: 'ALG 穿透', kind: 'singleton' },
  { key: 'reboot-schedules', label: '定时重启' },
  { key: 'vrrp-config', label: 'VRRP 热备', kind: 'singleton' },
  { key: 'backups', label: '配置快照备份' },
  { key: 'backup-auto', label: '自动定时备份', kind: 'singleton' },
  { key: 'upgrade-info', label: '固件版本检查', kind: 'singleton' },
  { key: 'disks', label: '存储磁盘' },
  { key: 'system-files', label: '文件目录浏览' },
  { key: 'web-admin-groups', label: 'Web 权限组' },
  { key: 'web-admin-accounts', label: 'Web 管理员' },
  { key: 'ftp-config', label: 'FTP 服务', kind: 'singleton', groupLabel: '高级' },
  { key: 'ftp-users', label: 'FTP 账号', groupLabel: '高级' },
  { key: 'http-users', label: 'HTTP 服务', groupLabel: '高级' },
  { key: 'samba-config', label: 'Samba 配置', kind: 'singleton', groupLabel: '高级' },
  { key: 'samba-users', label: 'Samba 共享', groupLabel: '高级' },
  { key: 'snmpd-config', label: 'SNMPD 协议', kind: 'singleton', groupLabel: '高级' },
]
</script>

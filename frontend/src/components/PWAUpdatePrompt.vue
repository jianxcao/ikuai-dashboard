<template></template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { registerSW } from 'virtual:pwa-register'

let refreshing = false

function handleControllerChange() {
  if (refreshing) return
  refreshing = true
  window.location.reload()
}

registerSW({
  immediate: true,
})

onMounted(() => {
  if (typeof navigator !== 'undefined' && 'serviceWorker' in navigator && navigator.serviceWorker.controller) {
    navigator.serviceWorker.addEventListener('controllerchange', handleControllerChange)
  }
})

onUnmounted(() => {
  if (typeof navigator !== 'undefined' && 'serviceWorker' in navigator) {
    navigator.serviceWorker.removeEventListener('controllerchange', handleControllerChange)
  }
})
</script>

<template>
  <div class="app-bg-mesh" aria-hidden="true"></div>
  <AppLayout />
  <PWAUpdatePrompt />
</template>

<script setup lang="ts">
import AppLayout from './components/layout/AppLayout.vue'
import PWAUpdatePrompt from './components/PWAUpdatePrompt.vue'
</script>

<style>
/* Global styles are imported in main.js via src/assets/main.css */
</style>

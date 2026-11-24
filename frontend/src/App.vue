<template>
  <div class="app-bg-mesh" aria-hidden="true" />

  <div v-if="authLoading" class="auth-loading" aria-busy="true">
    <div class="spinner" />
  </div>

  <LoginView v-else-if="authEnabled && !isAuthenticated" />

  <template v-else>
    <AppLayout />
    <PWAUpdatePrompt />
  </template>
</template>

<script setup lang="ts">
import AppLayout from './components/layout/AppLayout.vue'
import PWAUpdatePrompt from './components/PWAUpdatePrompt.vue'
import LoginView from './views/LoginView.vue'
import { useAuth } from './composables/useAuth'
import { useTheme } from './composables/useTheme'

useTheme()
const { authEnabled, isAuthenticated, loading: authLoading } = useAuth()
</script>

<style>
.auth-loading {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 300;
}
</style>

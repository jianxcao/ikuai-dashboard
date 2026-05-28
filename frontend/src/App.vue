<template>
  <div class="app-bg-mesh" aria-hidden="true" />

  <!-- 初始化中：显示加载动画 -->
  <div v-if="authLoading" class="auth-loading" aria-busy="true">
    <div class="spinner" />
  </div>

  <!-- 需要登录 -->
  <LoginView v-else-if="authEnabled && !isAuthenticated" />

  <!-- 已认证：正常 Dashboard -->
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

const { authEnabled, isAuthenticated, loading: authLoading } = useAuth()
</script>

<style>
/* Global styles are imported in main.js via src/assets/main.css */

.auth-loading {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 300;
}
</style>

<template>
  <div class="login-root">
    <div class="login-ambient" aria-hidden="true" />

    <main class="login-center">
      <div class="login-card glass-card">
        <!-- Logo & Brand -->
        <div class="login-brand">
          <img src="/ikuai-icon.svg" alt="iKuai Dashboard" class="login-logo" />
          <div>
            <h1 class="login-title">iKuai Dashboard</h1>
            <p class="login-subtitle">爱快实时监控</p>
          </div>
        </div>

        <div class="login-divider" />

        <!-- Form -->
        <form class="login-form" novalidate @submit.prevent="handleLogin">
          <div v-if="error" class="login-error" role="alert">
            <LockKeyholeOpen :size="15" />
            <span>{{ error }}</span>
          </div>

          <label class="login-field">
            <span class="login-label-text">用户名</span>
            <div class="login-input-wrap">
              <User :size="15" class="login-input-icon" />
              <input
                id="login-username"
                v-model.trim="username"
                type="text"
                class="login-input"
                placeholder="admin"
                autocomplete="username"
                autofocus
                :disabled="loading"
              />
            </div>
          </label>

          <label class="login-field">
            <span class="login-label-text">密码</span>
            <div class="login-input-wrap">
              <KeyRound :size="15" class="login-input-icon" />
              <input
                id="login-password"
                v-model="password"
                :type="showPwd ? 'text' : 'password'"
                class="login-input"
                placeholder="••••••••"
                autocomplete="current-password"
                :disabled="loading"
              />
              <button
                type="button"
                class="login-eye-btn"
                :title="showPwd ? '隐藏密码' : '显示密码'"
                tabindex="-1"
                @click="showPwd = !showPwd"
              >
                <Eye v-if="!showPwd" :size="14" />
                <EyeOff v-else :size="14" />
              </button>
            </div>
          </label>

          <button
            type="submit"
            class="login-submit"
            :disabled="loading || !username || !password"
          >
            <span v-if="loading" class="login-spinner" />
            <LogIn v-else :size="16" />
            <span>{{ loading ? '登录中...' : '登录' }}</span>
          </button>
        </form>

        <p class="login-hint">默认凭据：<code>admin</code> / <code>admin</code></p>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Eye, EyeOff, KeyRound, LogIn, LockKeyholeOpen, User } from 'lucide-vue-next'
import { useAuth } from '@/composables/useAuth'

const { login } = useAuth()

const username = ref('')
const password = ref('')
const showPwd = ref(false)
const loading = ref(false)
const error = ref<string | null>(null)

async function handleLogin() {
  if (!username.value || !password.value) return
  loading.value = true
  error.value = null
  try {
    await login(username.value, password.value)
    // 登录成功，App.vue 会响应 isAuthenticated 变化自动切换视图
  } catch (e: unknown) {
    if (e && typeof e === 'object' && 'response' in e) {
      const resp = (e as { response?: { data?: { message?: string } } }).response
      error.value = resp?.data?.message || '用户名或密码错误'
    } else {
      error.value = '网络错误，请稍后重试'
    }
    password.value = ''
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-root {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
}

.login-ambient {
  position: fixed;
  inset: 0;
  z-index: -1;
  background:
    radial-gradient(circle at 20% 20%, rgba(94, 203, 255, 0.22) 0%, transparent 40%),
    radial-gradient(circle at 80% 80%, rgba(98, 244, 173, 0.16) 0%, transparent 36%),
    radial-gradient(circle at 50% 50%, rgba(209, 154, 255, 0.10) 0%, transparent 48%),
    var(--app-bg);
}

.login-center {
  width: 100%;
  max-width: 400px;
  padding: 16px;
}

.login-card {
  padding: 36px 32px 28px;
  display: flex;
  flex-direction: column;
  gap: 0;
  animation: login-card-in 0.48s cubic-bezier(0.16, 1, 0.3, 1) both;
}

@keyframes login-card-in {
  from {
    opacity: 0;
    transform: translateY(20px) scale(0.97);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.login-brand {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 24px;
}

.login-logo {
  width: 48px;
  height: 48px;
  flex: 0 0 auto;
  filter: drop-shadow(0 0 12px rgba(94, 203, 255, 0.42));
}

.login-title {
  margin: 0;
  font-size: 20px;
  font-weight: 780;
  color: var(--text-primary);
  letter-spacing: -0.3px;
}

.login-subtitle {
  margin: 3px 0 0;
  font-size: 12px;
  color: var(--text-tertiary);
  font-weight: 600;
}

.login-divider {
  height: 1px;
  background: var(--glass-border);
  margin-bottom: 24px;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.login-error {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border: 1px solid var(--system-red-dim);
  border-radius: 10px;
  background: var(--system-red-dim);
  color: var(--system-red);
  font-size: 13px;
  font-weight: 600;
  animation: shake 0.36s ease;
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  20% { transform: translateX(-6px); }
  40% { transform: translateX(6px); }
  60% { transform: translateX(-4px); }
  80% { transform: translateX(4px); }
}

.login-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.login-label-text {
  font-size: 12px;
  font-weight: 760;
  color: var(--text-tertiary);
  letter-spacing: 0.02em;
}

.login-input-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.login-input-icon {
  position: absolute;
  left: 13px;
  color: var(--text-tertiary);
  pointer-events: none;
  flex-shrink: 0;
}

.login-input {
  width: 100%;
  height: 44px;
  padding: 0 40px 0 38px;
  color: var(--text-primary);
  background: var(--control-bg);
  border: 1px solid var(--control-border);
  border-radius: 12px;
  font-size: 14px;
  font-weight: 500;
  transition: border-color 0.18s ease, box-shadow 0.18s ease, background 0.18s ease;
}

.login-input:focus {
  outline: none;
  border-color: var(--system-blue);
  background: color-mix(in srgb, var(--control-bg) 80%, var(--system-blue-dim));
  box-shadow: 0 0 0 3px var(--system-blue-dim);
}

.login-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.login-eye-btn {
  position: absolute;
  right: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--text-tertiary);
  cursor: pointer;
  transition: color 0.16s ease, background 0.16s ease;
}

.login-eye-btn:hover {
  color: var(--text-secondary);
  background: var(--control-bg);
}

.login-submit {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 46px;
  margin-top: 4px;
  padding: 0 24px;
  border: 1px solid color-mix(in srgb, var(--system-blue) 58%, var(--control-border));
  border-radius: 12px;
  background: linear-gradient(135deg, var(--control-active) 0%, rgba(94, 203, 255, 0.16) 100%);
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 760;
  cursor: pointer;
  transition: transform 0.18s ease, border-color 0.18s ease, box-shadow 0.18s ease, opacity 0.18s ease;
}

.login-submit:hover:not(:disabled) {
  transform: translateY(-1px);
  border-color: var(--system-blue);
  box-shadow: 0 4px 24px var(--system-blue-dim);
}

.login-submit:active:not(:disabled) {
  transform: translateY(0);
}

.login-submit:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.login-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid var(--control-border);
  border-top-color: var(--system-blue);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.login-hint {
  margin-top: 20px;
  text-align: center;
  font-size: 12px;
  color: var(--text-muted);
}

.login-hint code {
  padding: 2px 6px;
  border-radius: 6px;
  background: var(--control-bg);
  border: 1px solid var(--control-border);
  font-family: 'SF Mono', ui-monospace, monospace;
  font-size: 11px;
  color: var(--text-tertiary);
}

@media (max-width: 480px) {
  .login-card {
    padding: 28px 22px 22px;
  }
}
</style>

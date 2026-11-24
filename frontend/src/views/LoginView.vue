<template>
  <div class="login-stage">
    <div class="login-orb orb-1" aria-hidden="true" />
    <div class="login-orb orb-2" aria-hidden="true" />

    <main class="login-shell v4-panel">
      <!-- Showcase / Theme Preview Side -->
      <section class="brand-pane">
        <div class="brand-top">
          <div class="logo-ring">
            <img src="/ikuai-icon.svg" alt="iKuai V4" class="logo-img" />
          </div>
          <div>
            <span class="eyebrow font-mono">IKUAI OS V4.0 CONSOLE</span>
            <h1>爱快 V4 多节点控制台</h1>
          </div>
        </div>

        <p class="lead">
          纯原生 `/api/v4.0` + Bearer Token 全能运维面板，支持多路由器秒级切换、全模块增删改查、Dry-Run 变更预览与只读保护锁。
        </p>

        <div class="feature-chips">
          <span class="v4-badge v4-badge-accent">苹果液态玻璃 & 6 套主题</span>
          <span class="v4-badge v4-badge-up">全模块可写管理</span>
          <span class="v4-badge" :class="serverReadOnly ? 'v4-badge-warn' : ''">
            {{ serverReadOnly ? '服务端只读保护已开启' : '支持只读模式 / Dry-Run' }}
          </span>
        </div>

        <div class="theme-showcase">
          <span class="showcase-label">即时切换登录视觉主题</span>
          <ThemePicker />
        </div>
      </section>

      <!-- Credentials Form Side -->
      <section class="form-pane v4-card">
        <div class="form-head">
          <h2>管理员登录</h2>
          <span class="font-mono sub-hint">Session Token · HMAC-SHA256</span>
        </div>

        <div v-if="error" class="error-banner" role="alert">
          <LockKeyholeOpen :size="15" />
          <span>{{ error }}</span>
        </div>

        <form class="login-form" novalidate @submit.prevent="handleLogin">
          <label class="field-group">
            <span>用户名</span>
            <div class="input-icon-wrap">
              <User :size="15" class="lead-ic" />
              <input
                id="login-username"
                v-model.trim="username"
                type="text"
                class="v4-input with-icon"
                placeholder="admin"
                autocomplete="username"
                autofocus
                :disabled="loading"
              />
            </div>
          </label>

          <label class="field-group">
            <span>密码</span>
            <div class="input-icon-wrap">
              <KeyRound :size="15" class="lead-ic" />
              <input
                id="login-password"
                v-model="password"
                :type="showPwd ? 'text' : 'password'"
                class="v4-input with-icon with-eye"
                placeholder="请输入登录密码"
                autocomplete="current-password"
                :disabled="loading"
              />
              <button
                type="button"
                class="eye-btn"
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
            class="v4-btn v4-btn-primary submit-btn"
            :disabled="loading || !username || !password"
          >
            <LogIn :size="16" />
            <span>{{ loading ? '验证中...' : '登录进入 V4 控制台' }}</span>
          </button>
        </form>

        <div class="form-foot">
          <span>默认开发凭据：<code class="font-mono">admin</code> / <code class="font-mono">admin</code></span>
          <button type="button" class="link-fill" @click="fillDefault">一键填入</button>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Eye, EyeOff, KeyRound, LockKeyholeOpen, LogIn, User } from 'lucide-vue-next'
import ThemePicker from '@/components/layout/ThemePicker.vue'
import { useAuth } from '@/composables/useAuth'

const { login, serverReadOnly } = useAuth()

const username = ref('')
const password = ref('')
const showPwd = ref(false)
const loading = ref(false)
const error = ref<string | null>(null)

function fillDefault() {
  username.value = 'admin'
  password.value = 'admin'
  error.value = null
}

async function handleLogin() {
  if (!username.value || !password.value) return
  loading.value = true
  error.value = null
  try {
    await login(username.value, password.value)
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
.login-stage {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  overflow: hidden;
}

.login-orb {
  position: fixed;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.42;
  pointer-events: none;
  z-index: -1;
}

.orb-1 {
  width: 420px;
  height: 420px;
  top: -80px;
  left: -60px;
  background: var(--accent-primary-soft);
}

.orb-2 {
  width: 380px;
  height: 380px;
  bottom: -80px;
  right: -60px;
  background: var(--accent-secondary-soft);
}

.login-shell {
  width: 100%;
  max-width: 960px;
  padding: 28px;
  display: grid;
  grid-template-columns: minmax(0, 1.08fr) minmax(0, 0.92fr);
  gap: 24px;
  align-items: stretch;
}

.brand-pane {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 20px;
  padding: 12px 6px;
}

.brand-top {
  display: flex;
  align-items: center;
  gap: 14px;
}

.logo-ring {
  width: 56px;
  height: 56px;
  border-radius: 18px;
  border: 1px solid var(--border-subtle);
  background: var(--surface-input);
  box-shadow: var(--inner-sheen);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.logo-img {
  width: 36px;
  height: 36px;
}

.eyebrow {
  font-size: 11px;
  font-weight: 700;
  color: var(--accent-text);
  letter-spacing: 0.08em;
}

.brand-top h1 {
  margin: 4px 0 0;
  font-size: 24px;
  font-weight: 800;
  color: var(--text-strong);
}

.lead {
  margin: 0;
  font-size: 13.5px;
  line-height: 1.65;
  color: var(--text-normal);
}

.feature-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.theme-showcase {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding-top: 14px;
  border-top: 1px solid var(--border-subtle);
}

.showcase-label {
  font-size: 11px;
  font-weight: 700;
  color: var(--text-muted);
}

.form-pane {
  padding: 28px 26px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-head h2 {
  margin: 0;
  font-size: 19px;
  font-weight: 800;
  color: var(--text-strong);
}

.sub-hint {
  display: block;
  margin-top: 4px;
  font-size: 11px;
  color: var(--text-muted);
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.field-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-normal);
}

.input-icon-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.lead-ic {
  position: absolute;
  left: 13px;
  color: var(--text-muted);
  pointer-events: none;
}

.with-icon {
  padding-left: 38px;
  min-height: 42px;
}

.with-eye {
  padding-right: 38px;
}

.eye-btn {
  position: absolute;
  right: 10px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;
}

.submit-btn {
  min-height: 44px;
  font-size: 14px;
  margin-top: 4px;
}

.form-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  font-size: 12px;
  color: var(--text-muted);
  padding-top: 10px;
  border-top: 1px solid var(--border-subtle);
}

.form-foot code {
  padding: 2px 6px;
  border-radius: 6px;
  background: var(--surface-input);
  border: 1px solid var(--border-subtle);
  color: var(--text-strong);
}

.link-fill {
  border: none;
  background: transparent;
  color: var(--accent-text);
  font-weight: 700;
  font-size: 12px;
  cursor: pointer;
}

.link-fill:hover {
  text-decoration: underline;
}

@media (max-width: 820px) {
  .login-shell {
    grid-template-columns: 1fr;
    padding: 18px;
  }
}

@media (max-width: 640px) {
  .login-stage {
    padding: 12px;
    align-items: flex-start;
    padding-top: max(18px, env(safe-area-inset-top));
  }

  .login-shell {
    padding: 16px;
    gap: 16px;
  }

  .form-pane {
    padding: 18px 15px;
  }

  .brand-top h1 {
    font-size: 20px;
  }

  .form-foot {
    flex-wrap: wrap;
  }
}
</style>

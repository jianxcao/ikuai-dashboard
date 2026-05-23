<template>
  <section class="settings-view">
    <div class="page-head">
      <div>
        <p class="eyebrow">路由器配置</p>
        <h1>爱快服务器</h1>
      </div>
      <button v-if="!loading && config.routers.length > 0" type="button" class="primary-action" @click="addRouter">
        <Plus :size="17" />
        <span>新增服务器</span>
      </button>
    </div>

    <div v-if="error" class="error-banner">{{ error }}</div>
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <span>正在读取配置...</span>
    </div>

    <template v-else>
      <div v-if="config.routers.length === 0" class="first-run-panel glass-panel">
        <div>
          <span class="status-label">首次配置</span>
          <strong>请先添加一台爱快服务器</strong>
        </div>
        <button type="button" class="primary-action" @click="addRouter">
          <Plus :size="17" />
          <span>添加服务器</span>
        </button>
      </div>

      <div v-else class="status-panel glass-panel">
        <div>
          <span class="status-label">当前服务器</span>
          <strong>{{ activeRouter?.name || '未选择' }}</strong>
        </div>
        <div>
          <span class="status-label">运行模式</span>
          <strong>{{ status?.mode === 'unconfigured' ? '未配置' : '真实连接' }}</strong>
        </div>
        <div v-if="status?.error" class="status-error">
          {{ status.error }}
        </div>
      </div>

      <div class="router-list">
        <article
          v-for="router in config.routers"
          :key="routerKey(router)"
          class="router-card glass-card"
          :class="{ active: router.id === config.active_router_id }"
        >
          <div class="router-card-head">
            <div class="router-title">
              <span class="router-id">{{ router.id }}</span>
              <input v-model.trim="router.name" class="name-input" aria-label="服务器名称" />
            </div>
            <div class="card-head-actions">
              <span class="version-badge">{{ router.version === 'v4' ? 'v4 Token' : 'v3 账号' }}</span>
              <span v-if="router.id === config.active_router_id" class="active-badge">当前</span>
              <button
                type="button"
                class="icon-action danger"
                :disabled="config.routers.length <= 1"
                title="删除服务器"
                aria-label="删除服务器"
                @click="removeRouter(router.id)"
              >
                <Trash2 :size="16" />
              </button>
            </div>
          </div>

          <div class="field-grid">
            <label>
              <span>ID</span>
              <input :value="router.id" @input="updateRouterId(router, $event)" />
            </label>
            <label>
              <span>地址</span>
              <input v-model.trim="router.url" placeholder="https://192.168.1.1" />
            </label>
            <label>
              <span>版本</span>
              <select v-model="router.version">
                <option value="v3">v3 账号密码</option>
                <option value="v4">v4 API Token</option>
              </select>
            </label>
            <label v-if="router.version !== 'v4'">
              <span>账号</span>
              <input v-model.trim="router.username" />
            </label>
            <label v-if="router.version !== 'v4'">
              <span>密码</span>
              <input v-model="router.password" type="password" placeholder="留空表示保留原密码" />
            </label>
            <label v-else class="wide-field">
              <span>API Token</span>
              <input v-model="router.token" type="password" placeholder="留空表示保留原 Token" />
            </label>
          </div>

          <div class="option-row">
            <label class="check-row">
              <input v-model="router.insecure_skip_verify" type="checkbox" />
              <span>允许自签证书</span>
            </label>
          </div>
        </article>
      </div>

      <div v-if="config.routers.length > 0" class="save-bar glass-panel">
        <span>{{ saved ? '配置已保存' : '修改后需要保存才会写入 YAML' }}</span>
        <button type="button" class="save-button" :disabled="saving" @click="save">
          <Save :size="16" />
          <span>{{ saving ? '保存中...' : '保存配置' }}</span>
        </button>
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { Plus, Save, Trash2 } from 'lucide-vue-next'
import { useRouterConfig } from '@/composables/useRouterConfig'
import type { RouterConfig } from '@/api/monitor'

const {
  loading,
  saving,
  error,
  saved,
  config,
  status,
  activeRouter,
  save,
  addRouter,
  removeRouter,
} = useRouterConfig()

let nextRouterKey = 0
const routerKeys = new WeakMap<RouterConfig, string>()

function routerKey(router: RouterConfig) {
  let key = routerKeys.get(router)
  if (!key) {
    key = `router-card-${++nextRouterKey}`
    routerKeys.set(router, key)
  }
  return key
}

function updateRouterId(router: RouterConfig, event: Event) {
  const previousID = router.id
  const nextID = (event.target as HTMLInputElement).value.trim()
  router.id = nextID
  if (config.value.active_router_id === previousID) {
    config.value.active_router_id = nextID
  }
}
</script>

<style scoped>
.settings-view {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.page-head,
.save-bar,
.status-panel,
.first-run-panel,
.router-card-head,
.option-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
}

button,
input,
select {
  font: inherit;
}

.primary-action,
.save-button,
.icon-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--control-border);
  color: var(--text-primary);
  background: var(--control-bg);
  cursor: pointer;
  transition: border-color 0.18s ease, background 0.18s ease, color 0.18s ease, transform 0.18s ease;
}

.primary-action,
.save-button {
  gap: 8px;
  min-height: 42px;
  padding: 0 15px;
  border-radius: 12px;
  border-color: color-mix(in srgb, var(--system-blue) 58%, var(--control-border));
  background: linear-gradient(135deg, var(--control-active), rgba(255, 255, 255, 0.06));
  font-size: 13px;
  font-weight: 760;
}

.primary-action:hover,
.save-button:hover,
.icon-action:hover {
  transform: translateY(-1px);
  border-color: var(--system-blue);
  background: var(--control-active);
}

.icon-action {
  width: 34px;
  height: 34px;
  flex: 0 0 auto;
  border-radius: 10px;
}

.icon-action.danger {
  color: var(--system-red);
  border-color: color-mix(in srgb, var(--system-red) 42%, var(--control-border));
}

.eyebrow,
.status-label,
.router-id {
  color: var(--text-tertiary);
  font-size: 12px;
  font-weight: 800;
}

h1 {
  margin: 4px 0 0;
  color: var(--text-primary);
  font-size: 28px;
}

.status-panel,
.first-run-panel,
.save-bar {
  padding: 16px;
}

.status-panel,
.first-run-panel {
  flex-wrap: wrap;
}

.status-panel strong,
.first-run-panel strong {
  display: block;
  margin-top: 4px;
  color: var(--text-primary);
}

.status-error {
  color: var(--system-orange);
  overflow-wrap: anywhere;
}

.router-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(100%, 430px), 1fr));
  gap: 16px;
}

.router-card {
  padding: 18px;
}

.router-card.active {
  border-color: var(--system-blue);
  box-shadow: var(--glass-shadow), 0 0 28px var(--system-blue-dim);
}

.router-card-head {
  align-items: flex-start;
}

.router-title {
  min-width: 0;
}

.card-head-actions {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex: 0 0 auto;
}

.name-input {
  width: 100%;
  margin-top: 6px;
  min-height: 42px;
  padding: 0 12px;
  color: var(--text-primary);
  font-size: 20px;
  font-weight: 760;
  background: color-mix(in srgb, var(--control-bg) 72%, transparent);
  border: 1px solid transparent;
  border-radius: 10px;
  outline: 0;
  transition: border-color 0.18s ease, background 0.18s ease, box-shadow 0.18s ease;
}

.name-input:hover {
  border-color: var(--control-border);
  background: var(--control-bg);
}

.name-input:focus {
  border-color: var(--system-blue);
  background: var(--control-bg);
  box-shadow: 0 0 0 3px var(--system-blue-dim);
}

.active-badge,
.version-badge {
  flex: 0 0 auto;
  padding: 5px 9px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 800;
}

.active-badge {
  color: var(--system-green);
  border: 1px solid var(--system-green);
  background: var(--system-green-dim);
}

.version-badge {
  color: var(--text-secondary);
  border: 1px solid var(--control-border);
  background: var(--control-bg);
}

.field-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-top: 18px;
}

label span {
  display: block;
  margin-bottom: 7px;
  color: var(--text-tertiary);
  font-size: 12px;
  font-weight: 760;
}

input,
select {
  width: 100%;
  min-height: 40px;
  padding: 0 12px;
  color: var(--text-primary);
  border: 1px solid var(--control-border);
  border-radius: 10px;
  background: var(--control-bg);
}

select {
  appearance: none;
  padding-right: 40px;
  background-image:
    linear-gradient(45deg, transparent 50%, currentColor 50%),
    linear-gradient(135deg, currentColor 50%, transparent 50%);
  background-position:
    calc(100% - 23px) 50%,
    calc(100% - 16px) 50%;
  background-size: 7px 7px, 7px 7px;
  background-repeat: no-repeat;
}

.wide-field {
  grid-column: span 2;
}

.option-row {
  justify-content: flex-start;
  flex-wrap: wrap;
  margin-top: 16px;
}

.check-row {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--text-secondary);
}

.check-row input {
  width: 16px;
  min-height: 16px;
  accent-color: var(--system-blue);
}

.check-row span {
  margin: 0;
}

button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

@media (max-width: 640px) {
  .page-head,
  .save-bar,
  .router-card-head {
    align-items: stretch;
    flex-direction: column;
  }

  .card-head-actions {
    justify-content: space-between;
  }

  .field-grid {
    grid-template-columns: 1fr;
  }

  .wide-field {
    grid-column: span 1;
  }

  .primary-action,
  .save-button {
    width: 100%;
  }
}
</style>

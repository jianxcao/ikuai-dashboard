<template>
  <section class="fleet-studio-page">
    <!-- Section 1: Multi-Router V4 Fleet -->
    <header class="ws-head v4-panel">
      <div>
        <span class="ws-eyebrow font-mono">MULTI-ROUTER V4 FLEET & THEME STUDIO</span>
        <h1>多爱快 V4 节点管理与主题工坊</h1>
        <p>仅支持爱快 OS v4.0 API Token 接入，支持并发健康探测、连接测试及多路由器无缝切换。</p>
      </div>
      <div class="head-actions">
        <button type="button" class="v4-btn" :disabled="probing" @click="refreshProbe">
          <RefreshCw :size="14" :class="{ spin: probing }" />
          <span>探测全部节点</span>
        </button>
        <button type="button" class="v4-btn v4-btn-primary" @click="addBlankRouter">
          <Plus :size="15" />
          <span>新增 V4 路由器</span>
        </button>
      </div>
    </header>

    <div v-if="error" class="error-banner">{{ error }}</div>
    <div v-if="saveToast" class="toast-banner">{{ saveToast }}</div>

    <div v-if="draftRouters.length === 0" class="v4-card empty-fleet">
      <Server :size="32" class="empty-ic" />
      <strong>尚未配置任何爱快 V4 路由器</strong>
      <p>在路由器后台「高级设置 → 系统设置 → API Token」生成令牌后，点击下方按钮添加。</p>
      <button type="button" class="v4-btn v4-btn-primary" @click="addBlankRouter">
        <Plus :size="15" />
        <span>添加第一台 V4 路由器</span>
      </button>
    </div>

    <div v-else class="router-grid">
      <article
        v-for="(router, idx) in draftRouters"
        :key="router._uid"
        class="v4-card router-card"
        :class="{ active: router._uid === activeUid }"
      >
        <div class="card-head">
          <div class="head-left">
            <span class="v4-badge font-mono">{{ router.id || `node-${idx + 1}` }}</span>
            <span v-if="router._uid === activeUid" class="v4-badge v4-badge-up">当前激活</span>
            <span v-if="router.mock" class="v4-badge v4-badge-accent">模拟</span>
            <span v-else-if="router.source === 'env'" class="v4-badge">ENV 注入</span>
          </div>

          <div class="head-right">
            <button
              v-if="router._uid !== activeUid && config.routers.some((r) => r.id === (router.original_id || router.id))"
              type="button"
              class="v4-btn btn-xs"
              :disabled="switchingId !== null"
              @click="switchNow(router)"
            >
              切换至此节点
            </button>
            <button
              type="button"
              class="v4-btn v4-btn-danger btn-xs"
              :disabled="draftRouters.length <= 1 && config.routers.length > 0"
              title="删除节点"
              @click="removeDraft(idx)"
            >
              <Trash2 :size="13" />
            </button>
          </div>
        </div>

        <!-- Live Probe Strip -->
        <div v-if="probeFor(router.original_id || router.id)" class="probe-strip font-mono">
          <span class="v4-badge" :class="probeFor(router.original_id || router.id)?.online ? 'v4-badge-up' : 'v4-badge-down'">
            {{ probeFor(router.original_id || router.id)?.online ? `在线 · ${probeFor(router.original_id || router.id)?.latency_ms}ms` : '离线' }}
          </span>
          <template v-if="probeFor(router.original_id || router.id)?.online">
            <span>{{ probeFor(router.original_id || router.id)?.hostname }}</span>
            <span>v{{ probeFor(router.original_id || router.id)?.version }}</span>
            <span>CPU {{ probeFor(router.original_id || router.id)?.cpu_percent.toFixed(1) }}%</span>
            <span>↓ {{ formatSpeed(probeFor(router.original_id || router.id)?.download_bps) }}</span>
          </template>
          <span v-else-if="probeFor(router.original_id || router.id)?.error" class="probe-err">{{ probeFor(router.original_id || router.id)?.error }}</span>
        </div>

        <div class="field-grid">
          <label class="field">
            <span>节点 ID</span>
            <input v-model.trim="router.id" class="v4-input font-mono" placeholder="hq-v4" />
          </label>
          <label class="field">
            <span>节点名称</span>
            <input v-model.trim="router.name" class="v4-input" placeholder="核心网关 V4" />
          </label>
          <label class="field wide">
            <span>路由器地址 (URL)</span>
            <input v-model.trim="router.url" class="v4-input font-mono" placeholder="https://192.168.9.1" />
          </label>
          <label class="field wide">
            <span>API Token (Bearer)</span>
            <div class="token-row">
              <input
                v-model.trim="router.token"
                :type="tokenVisible[router._uid] ? 'text' : 'password'"
                class="v4-input font-mono"
                :placeholder="router.has_token ? '已保存令牌（留空表示保持原令牌不变）' : '粘贴爱快后台生成的 API Token'"
              />
              <button type="button" class="v4-btn icon-btn" @click="tokenVisible[router._uid] = !tokenVisible[router._uid]">
                <Eye v-if="!tokenVisible[router._uid]" :size="14" />
                <EyeOff v-else :size="14" />
              </button>
            </div>
          </label>
          <label class="field">
            <span>标签 (逗号分隔)</span>
            <input
              :value="router.tags.join(', ')"
              class="v4-input"
              placeholder="主路由, 双线"
              @input="updateTags(router, ($event.target as HTMLInputElement).value)"
            />
          </label>
          <label class="field">
            <span>备注</span>
            <input v-model.trim="router.note" class="v4-input" placeholder="可选节点说明" />
          </label>
        </div>

        <div class="card-opts">
          <label class="check-item">
            <input v-model="router.insecure_skip_verify" type="checkbox" />
            <span>允许自签 HTTPS 证书</span>
          </label>
          <label class="check-item">
            <input v-model="router.mock" type="checkbox" />
            <span>高保真模拟模式 (Mock)</span>
          </label>
        </div>

        <div v-if="testResults[router._uid]" class="test-feedback" :class="testResults[router._uid]!.online ? 'ok' : 'fail'">
          <template v-if="testResults[router._uid]!.online">
            连接测试通过：{{ testResults[router._uid]!.hostname || 'iKuai' }} · {{ testResults[router._uid]!.verstring }} · 延迟 {{ testResults[router._uid]!.latency_ms }}ms
          </template>
          <template v-else>
            连接失败：{{ testResults[router._uid]!.error }}
          </template>
        </div>

        <div class="card-foot">
          <button type="button" class="v4-btn" :disabled="testingUid === router._uid" @click="runTest(router)">
            <Wifi :size="14" />
            <span>{{ testingUid === router._uid ? '测试中...' : '测试连接' }}</span>
          </button>
        </div>
      </article>
    </div>

    <div v-if="draftRouters.length > 0" class="save-bar v4-panel">
      <span>注意：点击保存会将节点列表持久化到后端 YAML 配置（未填写新 Token 的已有节点将保留原令牌）。</span>
      <button type="button" class="v4-btn v4-btn-primary" :disabled="saving" @click="handleSave">
        <Save :size="15" />
        <span>{{ saving ? '保存中...' : '保存节点配置' }}</span>
      </button>
    </div>

    <!-- Section 2: 6-Theme Visual Studio -->
    <section class="v4-card studio-section">
      <div class="section-head">
        <div>
          <h2>多主题视觉工坊 · 苹果液态玻璃领衔</h2>
          <p>6 套独立设计体系（包含真实折射边框的 Apple Liquid Glass 与高对比 Studio 平面风），点击即刻切换并持久化。</p>
        </div>
        <span class="v4-badge v4-badge-accent">{{ themes.length }} 套专属主题</span>
      </div>

      <div class="theme-gallery">
        <button
          v-for="item in themes"
          :key="item.id"
          type="button"
          class="theme-card"
          :class="{ selected: theme === item.id }"
          @click="setTheme(item.id)"
        >
          <div class="theme-preview" :style="{ background: item.previewBg, borderRadius: item.cornerRadius }">
            <div class="preview-glass-card" :class="{ glass: item.isGlass, dark: item.isDark }">
              <div class="preview-bar">
                <span :style="{ background: item.swatchPrimary }" />
                <span :style="{ background: item.swatchSecondary }" />
                <em v-if="item.isGlass" class="glass-chip">Glass</em>
              </div>
              <div class="preview-lines">
                <i :style="{ background: item.swatchPrimary, opacity: 0.85 }" />
                <i :style="{ background: item.swatchSecondary, opacity: 0.6 }" />
              </div>
            </div>
          </div>
          <div class="theme-meta">
            <div class="theme-title">
              <strong>{{ item.label }}</strong>
              <span class="v4-badge" :class="item.isGlass ? 'v4-badge-accent' : ''">{{ item.tag }}</span>
            </div>
            <span class="theme-en font-mono">{{ item.englishName }}</span>
            <p>{{ item.description }}</p>
          </div>
        </button>
      </div>
    </section>

    <!-- Section 3: Write Protection & Dashboard Access Token -->
    <section class="v4-card studio-section">
      <div class="section-head">
        <div>
          <h2>读写安全锁与访问认证保护</h2>
          <p>支持通过环境变量 <code>READ_ONLY=true</code> 启用不可绕过的服务端只读保护，或在此处一键切换前端写保护锁；同时可配置外部 API Token。</p>
        </div>
        <div class="head-actions">
          <button
            type="button"
            class="v4-btn"
            :class="readOnly ? '' : 'v4-btn-primary'"
            :disabled="serverReadOnly"
            @click="toggleSafetyLock"
          >
            {{ serverReadOnly ? '服务端强制只读 (READ_ONLY=true)' : uiSafetyLock ? '写保护已锁定（点击解锁）' : '读写操作已启用（点击加锁）' }}
          </button>
          <span class="v4-badge" :class="tokenEnabled ? 'v4-badge-up' : ''">
            {{ tokenEnabled ? 'API Token 已启用' : 'API Token 未启用' }}
          </span>
        </div>
      </div>

      <div v-if="tokenError" class="error-banner">{{ tokenError }}</div>
      <div v-if="tokenSuccess" class="toast-banner">{{ tokenSuccess }}</div>

      <div v-if="tokenEnabled && currentToken" class="token-row">
        <input
          :type="showDashToken ? 'text' : 'password'"
          :value="currentToken"
          readonly
          class="v4-input font-mono"
        />
        <button type="button" class="v4-btn" @click="showDashToken = !showDashToken">
          <Eye v-if="!showDashToken" :size="14" />
          <EyeOff v-else :size="14" />
        </button>
        <button type="button" class="v4-btn" @click="handleCopyDashToken">
          <Check v-if="copiedDash" :size="14" />
          <Copy v-else :size="14" />
        </button>
      </div>

      <div class="head-actions">
        <button type="button" class="v4-btn v4-btn-primary" :disabled="tokenSaving" @click="generateToken">
          <RefreshCw :size="14" />
          <span>{{ tokenEnabled ? '重新生成 Token' : '生成访问 Token' }}</span>
        </button>
        <button v-if="tokenEnabled" type="button" class="v4-btn v4-btn-danger" :disabled="tokenSaving" @click="clearToken">
          <Trash2 :size="14" />
          <span>清除保护</span>
        </button>
      </div>
    </section>
  </section>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { Check, Copy, Eye, EyeOff, Plus, RefreshCw, Save, Server, Trash2, Wifi } from 'lucide-vue-next'
import type { EditableRouterConfig, RouterHealthDTO } from '@/api/v4'
import { useAuth } from '@/composables/useAuth'
import { useFleet } from '@/composables/useFleet'
import { useTheme } from '@/composables/useTheme'
import { useTokenConfig } from '@/composables/useTokenConfig'
import { formatSpeed } from '@/composables/useFormatters'

const { readOnly, serverReadOnly, uiSafetyLock, toggleSafetyLock } = useAuth()

const {
  config,
  probeItems,
  probing,
  saving,
  switchingId,
  error,
  refreshProbe,
  activateRouter,
  saveRouters,
  testRouter,
} = useFleet()

const { theme, themes, setTheme } = useTheme()
const {
  tokenEnabled,
  currentToken,
  saving: tokenSaving,
  error: tokenError,
  successMsg: tokenSuccess,
  generateToken,
  clearToken,
  copyToken,
} = useTokenConfig()

interface DraftRouterCard extends EditableRouterConfig {
  _uid: string
}

let uidSeq = 0
function nextUid() {
  uidSeq += 1
  return `draft-node-${uidSeq}`
}

const draftRouters = ref<DraftRouterCard[]>([])
const activeUid = ref('')
const tokenVisible = reactive<Record<string, boolean>>({})
const testResults = reactive<Record<string, RouterHealthDTO>>({})
const testingUid = ref<string | null>(null)
const saveToast = ref<string | null>(null)
const showDashToken = ref(false)
const copiedDash = ref(false)

watch(
  config,
  (cfg) => {
    for (const k of Object.keys(testResults)) delete testResults[k]
    for (const k of Object.keys(tokenVisible)) delete tokenVisible[k]
    draftRouters.value = cfg.routers.map((r) => ({
      _uid: nextUid(),
      original_id: r.id,
      id: r.id,
      name: r.name,
      url: r.url,
      token: '',
      has_token: r.has_token,
      tags: [...(r.tags || [])],
      note: r.note || '',
      mock: r.mock,
      insecure_skip_verify: r.insecure_skip_verify,
      source: r.source,
    }))
    const activeCard = draftRouters.value.find((d) => d.id === cfg.active_router_id) ?? draftRouters.value[0]
    activeUid.value = activeCard?._uid ?? ''
  },
  { immediate: true, deep: true }
)

function probeFor(id: string) {
  return probeItems.value.find((p) => p.id === id)
}

function addBlankRouter() {
  const num = draftRouters.value.length + 1
  const id = `router-v4-${num}`
  const card: DraftRouterCard = {
    _uid: nextUid(),
    id,
    name: `爱快 V4 节点 #${num}`,
    url: 'https://192.168.9.1',
    token: '',
    has_token: false,
    tags: ['V4'],
    note: '',
    mock: false,
    insecure_skip_verify: true,
    source: 'yaml',
  }
  draftRouters.value.push(card)
  if (!activeUid.value) activeUid.value = card._uid
}

function removeDraft(idx: number) {
  const removed = draftRouters.value[idx]
  draftRouters.value.splice(idx, 1)
  if (removed && activeUid.value === removed._uid) {
    activeUid.value = draftRouters.value[0]?._uid ?? ''
  }
}

function updateTags(router: EditableRouterConfig, raw: string) {
  router.tags = raw
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)
}

async function runTest(router: DraftRouterCard) {
  testingUid.value = router._uid
  try {
    testResults[router._uid] = await testRouter({
      id: router.id.trim(),
      original_id: router.original_id,
      name: router.name.trim(),
      url: router.url.trim(),
      token: router.token || '',
      insecure_skip_verify: router.insecure_skip_verify,
      mock: router.mock,
    })
  } catch (e: unknown) {
    testResults[router._uid] = {
      id: router.id,
      name: router.name,
      url: router.url,
      tags: [],
      note: '',
      mock: router.mock,
      source: 'test',
      active: false,
      online: false,
      latency_ms: 0,
      hostname: '',
      version: '',
      verstring: '',
      arch: '',
      uptime: 0,
      cpu_percent: 0,
      memory_percent: 0,
      upload_bps: 0,
      download_bps: 0,
      online_users: 0,
      connections: 0,
      checked_at: 0,
      error: e instanceof Error ? e.message : String(e),
    }
  } finally {
    testingUid.value = null
  }
}

async function switchNow(router: DraftRouterCard) {
  activeUid.value = router._uid
  const targetId = config.value.routers.some((r) => r.id === router.id)
    ? router.id
    : router.original_id || router.id
  await activateRouter(targetId)
}

async function handleSave() {
  const activeCard = draftRouters.value.find((r) => r._uid === activeUid.value) ?? draftRouters.value[0]
  const effectiveActiveId = activeCard ? activeCard.id.trim() : ''
  const payloadRouters: EditableRouterConfig[] = draftRouters.value.map(({ _uid, ...rest }) => ({
    ...rest,
    id: rest.id.trim(),
    name: rest.name.trim(),
    url: rest.url.trim(),
  }))
  const ok = await saveRouters(effectiveActiveId, payloadRouters)
  if (ok) {
    saveToast.value = '爱快 V4 节点配置已保存并重载连接池'
    setTimeout(() => {
      saveToast.value = null
    }, 3000)
  }
}

async function handleCopyDashToken() {
  if (await copyToken()) {
    copiedDash.value = true
    setTimeout(() => {
      copiedDash.value = false
    }, 2000)
  }
}
</script>

<style scoped>
.fleet-studio-page { display: flex; flex-direction: column; gap: 16px; }
.ws-head { padding: 18px 20px; display: flex; flex-wrap: wrap; justify-content: space-between; align-items: center; gap: 12px; }
.ws-eyebrow { font-size: 11px; font-weight: 700; color: var(--accent-text); letter-spacing: 0.08em; }
.ws-head h1 { margin: 4px 0 0; font-size: 22px; font-weight: 800; }
.ws-head p { margin: 4px 0 0; font-size: 13px; color: var(--text-muted); }
.head-actions { display: flex; flex-wrap: wrap; gap: 8px; }

.toast-banner { padding: 10px 14px; border-radius: 10px; border: 1px solid color-mix(in srgb, var(--status-up) 45%, transparent); background: var(--status-up-soft); color: var(--status-up); font-size: 13px; font-weight: 600; }
.empty-fleet { padding: 42px 20px; display: flex; flex-direction: column; align-items: center; text-align: center; gap: 10px; }
.empty-ic { color: var(--accent-text); }
.empty-fleet p { margin: 0 0 8px; color: var(--text-muted); font-size: 13px; }

.router-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(min(100%, 460px), 1fr)); gap: 14px; }
.router-card { padding: 16px; display: flex; flex-direction: column; gap: 12px; }
.router-card.active { border-color: var(--accent-primary); }
.card-head { display: flex; justify-content: space-between; align-items: center; gap: 8px; }
.head-left, .head-right { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }

.probe-strip { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; padding: 8px 10px; border-radius: 10px; background: var(--surface-input); font-size: 11px; color: var(--text-normal); }
.probe-err { color: var(--status-down); }

.field-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.field { display: flex; flex-direction: column; gap: 5px; }
.field.wide { grid-column: span 2; }
.field > span { font-size: 12px; font-weight: 600; color: var(--text-normal); }

.token-row { display: flex; gap: 8px; align-items: center; }
.icon-btn { width: 38px; padding: 0; flex-shrink: 0; }

.card-opts { display: flex; flex-wrap: wrap; gap: 16px; font-size: 12px; color: var(--text-normal); }
.check-item { display: inline-flex; align-items: center; gap: 6px; cursor: pointer; }
.check-item input { accent-color: var(--accent-primary); }

.test-feedback { padding: 8px 10px; border-radius: 9px; font-size: 12px; }
.test-feedback.ok { background: var(--status-up-soft); color: var(--status-up); }
.test-feedback.fail { background: var(--status-down-soft); color: var(--status-down); }

.card-foot { display: flex; justify-content: flex-end; }
.btn-xs { min-height: 28px; padding: 0 10px; font-size: 11px; }

.save-bar { padding: 14px 18px; display: flex; flex-wrap: wrap; justify-content: space-between; align-items: center; gap: 12px; font-size: 12px; color: var(--text-muted); }

.studio-section { padding: 18px; display: flex; flex-direction: column; gap: 14px; }
.section-head { display: flex; flex-wrap: wrap; justify-content: space-between; align-items: flex-start; gap: 10px; }
.section-head h2 { margin: 0; font-size: 17px; font-weight: 750; }
.section-head p { margin: 4px 0 0; font-size: 12px; color: var(--text-muted); }

.theme-gallery { display: grid; grid-template-columns: repeat(auto-fill, minmax(230px, 1fr)); gap: 12px; }
.theme-card { padding: 11px; border-radius: var(--radius-lg); border: 1px solid var(--border-subtle); background: var(--surface-input); text-align: left; color: var(--text-strong); cursor: pointer; display: flex; flex-direction: column; gap: 10px; transition: all 0.16s ease; }
.theme-card:hover { border-color: var(--border-strong); transform: translateY(-2px); }
.theme-card.selected { border-color: var(--accent-primary); background: var(--accent-primary-soft); }

.theme-preview { height: 84px; padding: 10px; border: 1px solid rgba(148, 163, 184, 0.3); display: flex; align-items: stretch; }
.preview-glass-card {
  width: 100%; padding: 9px 10px; border-radius: 10px; background: rgba(255, 255, 255, 0.9);
  display: flex; flex-direction: column; justify-content: space-between;
  border: 1px solid rgba(148, 163, 184, 0.25);
}
.preview-glass-card.dark { background: rgba(18, 22, 34, 0.92); color: #f8fafc; }
.preview-glass-card.glass {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.72), rgba(255, 255, 255, 0.36));
  backdrop-filter: blur(12px);
  border-color: rgba(255, 255, 255, 0.78);
  box-shadow: inset 0 1px 1px rgba(255, 255, 255, 0.9);
}
.preview-glass-card.glass.dark {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.16), rgba(255, 255, 255, 0.05));
  border-color: rgba(255, 255, 255, 0.28);
  box-shadow: inset 0 1px 1px rgba(255, 255, 255, 0.32);
}
.preview-bar { display: flex; align-items: center; gap: 6px; }
.preview-bar span { width: 14px; height: 14px; border-radius: 50%; display: block; }
.glass-chip { margin-left: auto; font-style: normal; font-size: 9px; font-weight: 800; padding: 1px 6px; border-radius: 999px; background: rgba(255, 255, 255, 0.45); color: #0b1526; }
.preview-glass-card.glass.dark .glass-chip { background: rgba(255, 255, 255, 0.2); color: #f8fafc; }
.preview-lines { display: flex; flex-direction: column; gap: 5px; }
.preview-lines i { height: 6px; border-radius: 999px; display: block; }

.theme-title { display: flex; justify-content: space-between; align-items: center; }
.theme-en { font-size: 11px; color: var(--accent-text); }
.theme-meta p { margin: 5px 0 0; font-size: 11px; color: var(--text-muted); line-height: 1.45; }

.spin { animation: spin 0.9s linear infinite; }

@media (max-width: 640px) {
  .ws-head { padding: 14px; }
  .ws-head h1 { font-size: 18px; }
  .head-actions { width: 100%; }
  .head-actions .v4-btn { flex: 1; }
  .field-grid { grid-template-columns: 1fr; }
  .field.wide { grid-column: span 1; }
  .save-bar { flex-direction: column; align-items: stretch; padding: 14px; }
  .save-bar .v4-btn { width: 100%; }
  .studio-section { padding: 14px; }
  .theme-gallery { grid-template-columns: 1fr; }
  .theme-title { flex-wrap: wrap; gap: 6px; }
}
</style>

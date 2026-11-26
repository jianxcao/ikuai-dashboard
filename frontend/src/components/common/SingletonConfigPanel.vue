<template>
  <section class="singleton-panel v4-card">
    <header class="panel-head">
      <div>
        <div class="title-row">
          <h3>{{ spec?.label || resourceKey }}</h3>
          <span v-if="spec" class="v4-badge font-mono">{{ spec.v4_path }}</span>
          <span
            v-if="isUpstreamReadOnly"
            class="v4-badge"
            title="爱快 V4.0 OpenAPI 与 ikuai-cli 当前未开放此资源的写接口"
          >
            V4 接口只读
          </span>
          <span
            v-else-if="spec && !canUpdate && hasActions"
            class="v4-badge"
            title="配置字段为只读，可执行顶部服务动作"
          >
            配置只读 · 支持动作
          </span>
          <span v-if="readOnly" class="v4-badge v4-badge-warn">只读锁定</span>
        </div>
        <p v-if="spec?.description" class="desc">{{ spec.description }}</p>
      </div>

      <div class="head-actions">
        <button
          v-if="spec?.cli_command"
          type="button"
          class="v4-btn font-mono cli-copy-btn"
          @click="copyCLI"
        >
          <Terminal :size="13" />
          <span>{{ spec.cli_command }}</span>
          <Check v-if="copied" :size="13" class="ok-icon" />
        </button>

        <button
          v-for="act in spec?.actions || []"
          :key="act.key"
          type="button"
          class="v4-btn"
          :class="act.dangerous ? 'v4-btn-danger' : 'v4-btn-primary'"
          :disabled="saving || readOnly"
          @click="runAction(act)"
        >
          <Zap :size="14" />
          <span>{{ act.label }}</span>
        </button>

        <button type="button" class="v4-btn" :disabled="loading" @click="load">
          <RefreshCw :size="14" :class="{ spin: loading }" />
          <span>刷新</span>
        </button>
      </div>
    </header>

    <div v-if="error" class="error-banner">{{ error }}</div>
    <div v-if="toast" :class="toastWarn ? 'warn-banner' : 'toast-banner'">{{ toast }}</div>

    <div v-if="loading" class="loading-state">
      <div class="spinner" />
      <span class="font-mono">正在读取配置...</span>
    </div>

    <template v-else-if="!loadFailed">
      <div v-if="canUpdate" class="mode-bar">
        <div class="v4-seg-bar" role="tablist">
          <button type="button" class="v4-seg-btn" :class="{ active: !rawJsonMode }" @click="exitRawMode">
            可视化表单
          </button>
          <button type="button" class="v4-seg-btn" :class="{ active: rawJsonMode }" @click="enterRawMode">
            全字段 JSON 视图
          </button>
        </div>
      </div>

      <template v-if="rawJsonMode">
        <textarea v-model="rawJsonText" class="v4-textarea font-mono" rows="13" :disabled="!canUpdate || readOnly" />
      </template>

      <template v-else>
        <div v-if="spec?.fields?.length" class="fields-grid">
          <label v-for="field in spec.fields" :key="field.key" class="field-item">
            <span class="field-label">{{ field.label }}</span>
            <select
              v-if="field.type === 'select' || field.type === 'switch'"
              v-model="form[field.key]"
              class="v4-select"
              :disabled="!canUpdate || readOnly"
            >
              <option v-for="opt in field.options" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
            </select>
            <InterfaceSelect
              v-else-if="field.interface_scope"
              v-model="form[field.key]"
              :scope="field.interface_scope"
              :disabled="!canUpdate || readOnly"
              :placeholder="field.placeholder"
            />
            <input
              v-else-if="field.type === 'number'"
              v-model.number="form[field.key]"
              type="number"
              class="v4-input"
              :disabled="!canUpdate || readOnly"
              :placeholder="field.placeholder"
            />
            <input
              v-else-if="field.type === 'password'"
              v-model="form[field.key]"
              type="password"
              class="v4-input"
              :disabled="!canUpdate || readOnly"
              :placeholder="field.placeholder"
            />
            <input
              v-else
              v-model="form[field.key]"
              type="text"
              class="v4-input"
              :disabled="!canUpdate || readOnly"
              :placeholder="field.placeholder"
            />
          </label>
        </div>

        <div v-else class="kv-grid">
          <div v-for="(val, key) in configData" :key="key" class="kv-item">
            <span class="font-mono kv-k">{{ key }}</span>
            <strong class="font-mono kv-v">{{ val }}</strong>
          </div>
        </div>
      </template>

      <div v-if="dryRunPreview" class="dryrun-box">
        <strong class="dryrun-title">Dry-Run 请求预览</strong>
        <pre class="font-mono">{{ JSON.stringify(dryRunPreview, null, 2) }}</pre>
      </div>

      <footer v-if="canUpdate" class="panel-foot">
        <button type="button" class="v4-btn" :disabled="saving" @click="submit(true)">
          <Eye :size="14" />
          <span>Dry-Run 预览</span>
        </button>
        <button type="button" class="v4-btn v4-btn-primary" :disabled="saving || readOnly" @click="submit(false)">
          <Save :size="14" />
          <span>{{ saving ? '保存中...' : '保存配置' }}</span>
        </button>
      </footer>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Check, Eye, RefreshCw, Save, Terminal, Zap } from 'lucide-vue-next'
import { executeV4Action, fetchV4Resource, updateV4Resource, type ActionSpec, type FieldSpec, type ResourceSpec } from '@/api/v4'
import { useAuth } from '@/composables/useAuth'
import { useCatalog } from '@/composables/useCatalog'
import { useFleet } from '@/composables/useFleet'
import InterfaceSelect from '@/components/common/InterfaceSelect.vue'
import { compoundToCSV } from '@/composables/useFormatters'

const props = defineProps<{
  resourceKey: string
}>()

const { readOnly } = useAuth()
const { getSpec } = useCatalog()
const { routerEpoch } = useFleet()
const loading = ref(false)
const loadFailed = ref(false)
let reqSeq = 0
const saving = ref(false)
const error = ref<string | null>(null)
const toast = ref<string | null>(null)
const toastWarn = ref(false)
const copied = ref(false)
const rawJsonMode = ref(false)
const rawJsonText = ref('{}')
const spec = ref<ResourceSpec | null>(null)
const configData = ref<Record<string, unknown>>({})
const form = ref<Record<string, unknown>>({})
const dryRunPreview = ref<Record<string, unknown> | null>(null)

const canUpdate = computed(() => spec.value?.operations?.includes('update') ?? false)
const hasActions = computed(() => (spec.value?.actions?.length ?? 0) > 0)
const isUpstreamReadOnly = computed(() => Boolean(spec.value) && !canUpdate.value && !hasActions.value)

async function load() {
  const seq = ++reqSeq
  loading.value = true
  error.value = null
  dryRunPreview.value = null
  spec.value = getSpec(props.resourceKey) || spec.value
  try {
    const res = await fetchV4Resource(props.resourceKey)
    if (seq !== reqSeq) return
    loadFailed.value = false
    spec.value = res.resource
    const obj = res.config || res.rows[0] || {}
    configData.value = obj
    populateFormFromObject(obj, res.resource.fields || [])
    rawJsonText.value = JSON.stringify(buildRawSnapshot(obj, form.value), null, 2)
  } catch (e: unknown) {
    if (seq !== reqSeq) return
    loadFailed.value = true
    spec.value = getSpec(props.resourceKey) || null
    configData.value = {}
    form.value = {}
    rawJsonText.value = '{}'
    error.value = extractError(e)
  } finally {
    if (seq === reqSeq) {
      loading.value = false
    }
  }
}

function buildRawSnapshot(base: Record<string, unknown>, edits: Record<string, unknown>): Record<string, unknown> {
  const merged: Record<string, unknown> = { ...base }
  for (const f of spec.value?.fields || []) {
    const val = edits[f.key]
    if (f.type === 'password' && (val === '' || val === undefined)) {
      continue
    }
    const orig = base[f.key]
    if (f.type === 'csv' && orig && typeof orig === 'object' && !Array.isArray(orig)) {
      const existingObj = Array.isArray((orig as { object?: unknown }).object)
        ? (orig as { object: unknown[] }).object
        : []
      const customList = String(val ?? '')
        .split(',')
        .map((s) => s.trim())
        .filter(Boolean)
      merged[f.key] = { custom: customList, object: existingObj }
      continue
    }
    if (val !== undefined) {
      merged[f.key] = val
    }
  }
  return merged
}

function populateFormFromObject(obj: Record<string, unknown>, fields: FieldSpec[]) {
  const nextForm: Record<string, unknown> = {}
  for (const f of fields) {
    const raw = obj[f.key]
    if (f.type === 'csv' && raw && typeof raw === 'object') {
      nextForm[f.key] = compoundToCSV(raw)
    } else if (f.type === 'password') {
      nextForm[f.key] = ''
    } else if (f.type === 'select' || f.type === 'switch') {
      nextForm[f.key] = raw !== undefined && raw !== null ? String(raw) : (f.default ?? '')
    } else {
      nextForm[f.key] = raw !== undefined && raw !== null ? raw : (f.default ?? '')
    }
  }
  form.value = nextForm
}

function enterRawMode() {
  rawJsonText.value = JSON.stringify(buildRawSnapshot(configData.value, form.value), null, 2)
  rawJsonMode.value = true
}

function exitRawMode() {
  if (!rawJsonMode.value) return
  let parsed: unknown
  try {
    parsed = JSON.parse(rawJsonText.value)
  } catch {
    error.value = 'JSON 格式无效，请先修正语法再切回可视化表单'
    return
  }
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
    error.value = 'JSON 根节点必须是对象格式 ({ ... })'
    return
  }
  error.value = null
  const prev = configData.value
  const parsedObj = parsed as Record<string, unknown>
  const mergedObj = { ...prev, ...parsedObj }
  configData.value = mergedObj
  populateFormFromObject(mergedObj, spec.value?.fields || [])
  for (const f of spec.value?.fields || []) {
    if (
      f.type === 'password' &&
      typeof parsedObj[f.key] === 'string' &&
      parsedObj[f.key] !== '' &&
      parsedObj[f.key] !== prev[f.key]
    ) {
      form.value[f.key] = parsedObj[f.key]
    }
  }
  rawJsonMode.value = false
}

async function submit(dryRun: boolean) {
  if (!dryRun && readOnly.value) {
    error.value = '当前处于只读保护模式，请使用 Dry-Run 预览或关闭写保护'
    return
  }
  saving.value = true
  error.value = null
  let payload: Record<string, unknown> = { ...form.value }
  if (rawJsonMode.value) {
    let parsed: unknown
    try {
      parsed = JSON.parse(rawJsonText.value)
    } catch {
      error.value = 'JSON 格式无效，请检查'
      saving.value = false
      return
    }
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
      error.value = 'JSON 根节点必须是对象格式 ({ ... })'
      saving.value = false
      return
    }
    payload = parsed as Record<string, unknown>
  } else {
    for (const f of spec.value?.fields || []) {
      const val = payload[f.key]
      if (f.type === 'password' && val === '') {
        delete payload[f.key]
      } else if (f.type === 'number' && (val === '' || val === null || Number.isNaN(val))) {
        delete payload[f.key]
      }
    }
  }

  try {
    const res = await updateV4Resource(props.resourceKey, null, payload, dryRun)
    if (dryRun) {
      dryRunPreview.value = res
    } else {
      dryRunPreview.value = null
      showToast('配置已应用至爱快 V4 路由器')
      await load()
    }
  } catch (e: unknown) {
    error.value = extractError(e)
  } finally {
    saving.value = false
  }
}

async function runAction(act: ActionSpec) {
  if (readOnly.value) return
  if (act.dangerous && !window.confirm(act.confirm_text || `确认执行 ${act.label} 吗？`)) return
  saving.value = true
  error.value = null
  try {
    const res = await executeV4Action(act.key, false)
    showToast(String(res.message || `动作「${act.label}」执行完成`), Boolean(res.warning))
    await load()
  } catch (e: unknown) {
    error.value = extractError(e)
  } finally {
    saving.value = false
  }
}

async function copyCLI() {
  if (!spec.value?.cli_command) return
  try {
    await navigator.clipboard.writeText(spec.value.cli_command)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 1800)
  } catch {
    // ignore
  }
}

function showToast(msg: string, isWarn = false) {
  toastWarn.value = isWarn
  toast.value = msg
  setTimeout(() => {
    if (toast.value === msg) {
      toast.value = null
      toastWarn.value = false
    }
  }, isWarn ? 5500 : 2800)
}

function extractError(e: unknown): string {
  if (e && typeof e === 'object' && 'response' in e) {
    const resp = (e as { response?: { data?: { message?: string } } }).response
    if (resp?.data?.message) return resp.data.message
  }
  return e instanceof Error ? e.message : String(e)
}

watch(
  [() => props.resourceKey, routerEpoch],
  () => {
    rawJsonMode.value = false
    loadFailed.value = false
    spec.value = getSpec(props.resourceKey) || null
    configData.value = {}
    form.value = {}
    rawJsonText.value = '{}'
    toast.value = null
    error.value = null
    dryRunPreview.value = null
    void load()
  },
  { immediate: true }
)
</script>

<style scoped>
.singleton-panel {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.panel-head {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.title-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.title-row h3 {
  margin: 0;
  font-size: 16px;
}

.desc {
  margin: 4px 0 0;
  font-size: 12px;
  color: var(--text-muted);
}

.head-actions,
.mode-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.fields-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(230px, 1fr));
  gap: 12px;
}

.field-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-normal);
}

.kv-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(210px, 1fr));
  gap: 10px;
}

.kv-item {
  padding: 10px 12px;
  border-radius: var(--radius-md);
  border: 1px solid var(--divider-line);
  background: var(--surface-input);
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.kv-k {
  font-size: 11px;
  color: var(--text-muted);
}

.kv-v {
  font-size: 13px;
  color: var(--text-strong);
  word-break: break-all;
}

.dryrun-box {
  padding: 12px;
  border-radius: var(--radius-md);
  border: 1px solid var(--accent-primary);
  background: var(--surface-input);
}

.dryrun-title {
  display: block;
  margin-bottom: 6px;
  font-size: 12px;
}

.dryrun-box pre {
  margin: 0;
  font-size: 11px;
  overflow-x: auto;
}

.panel-foot {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding-top: 12px;
  border-top: 1px solid var(--divider-line);
}

.ok-icon {
  color: var(--status-up);
}

.spin {
  animation: spin 0.9s linear infinite;
}

.cli-copy-btn {
  max-width: 100%;
  overflow: hidden;
}

.cli-copy-btn span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 640px) {
  .singleton-panel {
    padding: 14px;
    gap: 12px;
  }

  .head-actions {
    width: 100%;
  }

  .fields-grid,
  .kv-grid {
    grid-template-columns: 1fr;
  }

  .panel-foot {
    flex-wrap: wrap;
    justify-content: stretch;
  }

  .panel-foot .v4-btn {
    flex: 1;
  }
}
</style>

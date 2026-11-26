<template>
  <teleport to="body">
    <div v-if="open && groupKey" class="drawer-backdrop center" @click.self="emit('close')">
      <div class="confirm-dialog v4-panel modal-wide obj-quick-modal" role="dialog" :aria-label="modalTitle">
        <div class="modal-top">
          <div class="modal-title-group">
            <h4>{{ modalTitle }}</h4>
            <span v-if="recordId" class="v4-badge font-mono">#{{ recordId }}</span>
            <span v-else-if="!loading && notFound" class="v4-badge v4-badge-warn">未找到 · 可新建</span>
          </div>
          <div class="modal-head-actions">
            <button
              type="button"
              class="v4-btn btn-xs"
              :title="`跳转到「${meta?.label || '对象分组'}」并定位 ${form.group_name || groupName}`"
              @click="handleJump"
            >
              <ExternalLink :size="13" />
              <span>跳转对应分组</span>
            </button>
            <button type="button" class="icon-btn" aria-label="关闭弹窗" @click="emit('close')">
              <X :size="14" />
            </button>
          </div>
        </div>

        <div v-if="error" class="error-banner">{{ error }}</div>
        <div v-if="toast" class="toast-banner">{{ toast }}</div>

        <div v-if="loading" class="loading-state modal-loading">
          <div class="spinner" />
          <span class="font-mono">正在读取 @{{ groupName }}...</span>
        </div>

        <template v-else>
          <div v-if="notFound" class="info-banner">
            当前路由器中尚未找到名为 <strong>{{ groupName }}</strong> 的{{ meta?.label || '对象组' }}，可直接在此快速创建。
          </div>

          <div class="form-grid">
            <label class="form-field">
              <span class="field-label">对象组名 <em>*</em></span>
              <input
                v-model.trim="form.group_name"
                type="text"
                class="v4-input"
                :disabled="saving || Boolean(recordId)"
                :title="recordId ? '为避免破坏当前规则引用，快改模式下锁定分组名；如需重命名请点击右上角跳转分组页' : ''"
              />
            </label>

            <template v-if="groupKey === 'time-objects'">
              <label class="form-field">
                <span class="field-label">时间模式</span>
                <select v-model="form.type" class="v4-select" :disabled="saving">
                  <option value="weekly">按周循环 (weekly)</option>
                  <option value="date">指定日期 (date)</option>
                </select>
              </label>

              <label v-if="form.type === 'weekly'" class="form-field">
                <span class="field-label">星期（如 12345 或 1234567）</span>
                <input v-model.trim="form.weekdays" type="text" class="v4-input font-mono" placeholder="12345" :disabled="saving" />
              </label>

              <div class="form-grid two-cols">
                <label class="form-field">
                  <span class="field-label">开始时间 <em>*</em></span>
                  <input
                    v-model.trim="form.start_time"
                    type="text"
                    class="v4-input font-mono"
                    placeholder="09:00 或 2026-05-01T08:00"
                    :disabled="saving"
                  />
                </label>
                <label class="form-field">
                  <span class="field-label">结束时间 <em>*</em></span>
                  <input
                    v-model.trim="form.end_time"
                    type="text"
                    class="v4-input font-mono"
                    placeholder="18:00 或 2026-05-10T18:00"
                    :disabled="saving"
                  />
                </label>
              </div>

              <label class="form-field">
                <span class="field-label">时段说明</span>
                <input v-model.trim="form.val_comment" type="text" class="v4-input" :disabled="saving" />
              </label>
            </template>

            <label v-else class="form-field">
              <span class="field-label">
                分组成员列表（每行或逗号分隔） <em>*</em>
              </span>
              <textarea
                v-model="form.membersText"
                class="v4-textarea font-mono"
                rows="5"
                :placeholder="meta?.placeholder"
                :disabled="saving"
              />
            </label>

            <label class="form-field">
              <span class="field-label">备注</span>
              <input v-model.trim="form.comment" type="text" class="v4-input" placeholder="可选说明" :disabled="saving" />
            </label>
          </div>

          <div v-if="dryRunPreview" class="dryrun-box">
            <div class="dryrun-head">
              <strong>Dry-Run 请求预览（未写入路由器）</strong>
              <button type="button" class="icon-btn" @click="dryRunPreview = null">
                <X :size="13" />
              </button>
            </div>
            <pre class="font-mono">{{ JSON.stringify(dryRunPreview, null, 2) }}</pre>
          </div>
        </template>

        <div class="confirm-actions">
          <button type="button" class="v4-btn" @click="emit('close')">关闭</button>
          <button type="button" class="v4-btn" :disabled="loading || saving" @click="submit(true)">
            <Eye :size="14" />
            <span>Dry-Run 预览</span>
          </button>
          <button
            type="button"
            class="v4-btn v4-btn-primary"
            :disabled="loading || saving || readOnly"
            :title="readOnly ? '只读保护中（仅支持 Dry-Run 预览）' : '保存到路由器'"
            @click="submit(false)"
          >
            <Save :size="14" />
            <span>{{ saving ? '保存中...' : readOnly ? '只读保护中' : notFound ? '立即创建分组' : '保存分组修改' }}</span>
          </button>
        </div>
      </div>
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ExternalLink, Eye, Save, X } from 'lucide-vue-next'
import { createV4Resource, fetchV4Resource, updateV4Resource } from '@/api/v4'
import { useAuth } from '@/composables/useAuth'
import { formatObjectRef } from '@/composables/useFormatters'
import {
  invalidateObjectGroupCache,
  jumpToObjectGroup,
  OBJECT_GROUP_META,
  type ObjectGroupKey,
} from '@/composables/useObjectGroups'

const props = defineProps<{
  open: boolean
  groupKey: ObjectGroupKey | null
  groupName: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'saved', payload: { groupKey: ObjectGroupKey; groupName: string }): void
}>()

const { readOnly } = useAuth()

const loading = ref(false)
const saving = ref(false)
const notFound = ref(false)
const error = ref<string | null>(null)
const toast = ref<string | null>(null)
const recordId = ref<string | null>(null)
const existingRow = ref<Record<string, unknown> | null>(null)
const dryRunPreview = ref<Record<string, unknown> | null>(null)
let reqSeq = 0

const form = reactive({
  group_name: '',
  membersText: '',
  comment: '',
  type: 'weekly',
  weekdays: '1234567',
  start_time: '09:00',
  end_time: '18:00',
  val_comment: '',
})

const meta = computed(() => (props.groupKey ? OBJECT_GROUP_META[props.groupKey] : null))
const modalTitle = computed(() => {
  const label = meta.value?.label || '策略对象组'
  return `${label} · @${props.groupName}`
})

function resetForm(name: string) {
  form.group_name = name
  form.membersText = ''
  form.comment = ''
  form.type = 'weekly'
  form.weekdays = '1234567'
  form.start_time = '09:00'
  form.end_time = '18:00'
  form.val_comment = ''
}

function extractMembersText(row: Record<string, unknown>, valueKey?: string): string {
  if (typeof row.value === 'string' && row.value.trim() !== '') {
    return row.value
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
      .join('\n')
  }
  const gv = row.group_value
  if (Array.isArray(gv)) {
    const items = gv
      .map((entry) => {
        if (entry && typeof entry === 'object') {
          const rec = entry as Record<string, unknown>
          if (valueKey && rec[valueKey] !== undefined) {
            return String(rec[valueKey] ?? '').trim()
          }
          return formatObjectRef(rec)
        }
        return String(entry ?? '').trim()
      })
      .filter(Boolean)
    return items.join('\n')
  }
  return ''
}

function populateFromRow(row: Record<string, unknown>) {
  form.group_name = String(row.group_name ?? props.groupName)
  form.comment = String(row.comment ?? '')
  if (props.groupKey === 'time-objects') {
    const first =
      Array.isArray(row.group_value) && row.group_value.length > 0 && typeof row.group_value[0] === 'object'
        ? (row.group_value[0] as Record<string, unknown>)
        : {}
    form.type = String(row.type ?? first.type ?? 'weekly')
    form.weekdays = String(row.weekdays ?? first.weekdays ?? '1234567')
    form.start_time = String(row.start_time ?? first.start_time ?? '09:00')
    form.end_time = String(row.end_time ?? first.end_time ?? '18:00')
    form.val_comment = String(row.val_comment ?? first.comment ?? '')
    return
  }
  form.membersText = extractMembersText(row, meta.value?.valueKey)
}

async function loadTarget() {
  if (!props.open || !props.groupKey || !props.groupName) return
  const seq = ++reqSeq
  loading.value = true
  notFound.value = false
  error.value = null
  toast.value = null
  dryRunPreview.value = null
  recordId.value = null
  existingRow.value = null
  resetForm(props.groupName)

  try {
    const res = await fetchV4Resource(props.groupKey, { page: 1, page_size: 200, search: props.groupName })
    if (seq !== reqSeq) return
    const rows = res.rows || []
    const matched = rows.find((r) => String(r.group_name ?? '') === props.groupName)
    if (!matched) {
      notFound.value = true
      return
    }
    existingRow.value = matched
    recordId.value = matched.id !== undefined && matched.id !== null ? String(matched.id) : null
    populateFromRow(matched)
  } catch (e: unknown) {
    if (seq !== reqSeq) return
    error.value = extractError(e)
  } finally {
    if (seq === reqSeq) loading.value = false
  }
}

function buildPayload(): Record<string, unknown> | null {
  const groupName = form.group_name.trim()
  if (!groupName) {
    error.value = '请填写对象组名'
    return null
  }
  if (props.groupKey === 'time-objects') {
    if (!form.start_time.trim() || !form.end_time.trim()) {
      error.value = '请填写开始和结束时间'
      return null
    }
    const firstWindow: Record<string, unknown> = {
      type: form.type || 'weekly',
      start_time: form.start_time.trim(),
      end_time: form.end_time.trim(),
      comment: form.val_comment.trim(),
    }
    if (firstWindow.type === 'weekly') {
      firstWindow.weekdays = form.weekdays.trim() || '1234567'
    }
    const prevGV = Array.isArray(existingRow.value?.group_value) ? existingRow.value.group_value : []
    return {
      group_name: groupName,
      group_value: [firstWindow, ...prevGV.slice(1)],
      comment: form.comment.trim(),
    }
  }

  const members = form.membersText
    .split(/[\n,]+/)
    .map((s) => s.trim())
    .filter(Boolean)
  if (members.length === 0) {
    error.value = '请至少填写一个分组成员值'
    return null
  }
  return {
    group_name: groupName,
    value: members.join(','),
    comment: form.comment.trim(),
  }
}

async function submit(dryRun: boolean) {
  if (!props.groupKey) return
  if (!dryRun && readOnly.value) {
    error.value = '当前处于只读保护模式，请使用 Dry-Run 预览'
    return
  }
  error.value = null
  toast.value = null
  const payload = buildPayload()
  if (!payload) return

  saving.value = true
  try {
    const res = recordId.value
      ? await updateV4Resource(props.groupKey, recordId.value, payload, dryRun)
      : await createV4Resource(props.groupKey, payload, dryRun)
    if (dryRun) {
      dryRunPreview.value = res
    } else {
      dryRunPreview.value = null
      invalidateObjectGroupCache(props.groupKey)
      const savedName = String(payload.group_name ?? props.groupName)
      emit('saved', { groupKey: props.groupKey, groupName: savedName })
      emit('close')
    }
  } catch (e: unknown) {
    error.value = extractError(e)
  } finally {
    saving.value = false
  }
}

function handleJump() {
  if (!props.groupKey) return
  jumpToObjectGroup(props.groupKey, form.group_name.trim() || props.groupName)
  emit('close')
}

function extractError(e: unknown): string {
  if (e && typeof e === 'object' && 'response' in e) {
    const resp = (e as { response?: { data?: { message?: string } } }).response
    if (resp?.data?.message) return resp.data.message
  }
  return e instanceof Error ? e.message : String(e)
}

watch(
  () => [props.open, props.groupKey, props.groupName] as const,
  ([isOpen]) => {
    if (isOpen) {
      void loadTarget()
    } else {
      reqSeq += 1
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.drawer-backdrop {
  position: fixed;
  inset: 0;
  z-index: 260;
  background: rgba(6, 10, 20, 0.44);
  -webkit-backdrop-filter: blur(8px);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.confirm-dialog {
  width: min(560px, calc(100vw - 32px));
  max-height: 88dvh;
  overflow-y: auto;
  padding: 20px;
  background: var(--surface-popover);
  border-color: var(--chrome-border);
  box-shadow: var(--chrome-shadow);
  -webkit-backdrop-filter: var(--backdrop-chrome);
  backdrop-filter: var(--backdrop-chrome);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.modal-top {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.modal-top h4 {
  margin: 0;
  font-size: 15px;
  color: var(--text-strong);
}

.modal-title-group {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.modal-head-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-control);
  background: var(--surface-btn);
  color: var(--text-normal);
  cursor: pointer;
}

.icon-btn:hover:not(:disabled) {
  border-color: var(--accent-primary);
  color: var(--text-strong);
}

.modal-loading {
  min-height: 150px;
}

.info-banner {
  padding: 10px 12px;
  border-radius: var(--radius-md);
  border: 1px solid color-mix(in srgb, var(--accent-primary) 28%, transparent);
  background: var(--accent-primary-soft);
  color: var(--text-normal);
  font-size: 12.5px;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-normal);
}

.field-label em {
  color: var(--status-down);
  font-style: normal;
}

.dryrun-box {
  padding: 12px;
  border-radius: var(--radius-md);
  border: 1px solid var(--accent-primary);
  background: var(--surface-input);
}

.dryrun-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 12px;
}

.dryrun-box pre {
  margin: 0;
  font-size: 11.5px;
  line-height: 1.55;
  overflow-x: auto;
  color: var(--text-normal);
  padding: 12px;
  border-radius: var(--radius-md);
  background: var(--surface-input);
  border: 1px solid var(--divider-line);
}

.confirm-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 4px;
}

.obj-quick-modal .two-cols {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

@media (max-width: 640px) {
  .confirm-dialog {
    width: calc(100vw - 20px);
    padding: 16px;
  }

  .obj-quick-modal .two-cols {
    grid-template-columns: 1fr;
  }
}
</style>

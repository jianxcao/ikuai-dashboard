<template>
  <div class="resources-page">
    <div class="glass-panel page-head">
      <div>
        <h2 class="panel-title">配置管理</h2>
        <p class="panel-subtitle">v3 / v4 共有资源</p>
      </div>
      <div class="head-actions">
        <select class="resource-select" v-model="selectedName" :disabled="loadingResources">
          <option v-for="resource in resources" :key="resource.name" :value="resource.name">
            {{ resource.label }}
          </option>
        </select>
        <button class="liquid-button icon-button" type="button" :disabled="loadingRows" @click="loadSelected">
          <RefreshCw :size="15" :class="{ spinning: loadingRows }" />
        </button>
      </div>
    </div>

    <div v-if="error" class="error-banner">
      <span class="error-icon">!</span>
      <span>{{ error }}</span>
    </div>

    <section class="glass-panel resource-panel">
      <div class="resource-meta">
        <div>
          <h3>{{ activeResource?.label || '资源' }}</h3>
          <p>{{ activeResource?.name }} · {{ activeResource?.group }} · {{ resourceStateLabel }}</p>
        </div>
        <button
          v-if="activeResource?.writable"
          class="liquid-button action-button"
          type="button"
          @click="startCreate"
        >
          <Plus :size="15" />
          <span>新增</span>
        </button>
      </div>

      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th v-for="key in columns" :key="key">{{ key }}</th>
              <th v-if="activeResource?.writable">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, index) in rows" :key="rowKey(row, index)">
              <td v-for="key in columns" :key="key" class="font-mono">{{ formatCell(row[key]) }}</td>
              <td v-if="activeResource?.writable" class="row-actions">
                <button class="liquid-button small-button" type="button" @click="startEdit(row)">
                  <Pencil :size="14" />
                </button>
                <button class="liquid-button small-button danger" type="button" :disabled="!rowId(row)" @click="deleteRow(row)">
                  <Trash2 :size="14" />
                </button>
              </td>
            </tr>
            <tr v-if="!rows.length">
              <td :colspan="columns.length + (activeResource?.writable ? 1 : 0)" class="empty-row">
                {{ loadingRows ? '加载中...' : resourceError || '暂无数据' }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section v-if="editorOpen && activeResource?.writable" class="glass-panel editor-panel">
      <div class="editor-head">
        <div>
          <h3>{{ editorMode === 'create' ? '新增资源' : `编辑 #${editorId}` }}</h3>
          <p>{{ activeResource.label }} · JSON 请求体</p>
        </div>
        <button class="liquid-button small-button" type="button" @click="closeEditor">
          <X :size="15" />
        </button>
      </div>
      <textarea v-model="editorText" class="json-editor" spellcheck="false"></textarea>
      <div class="editor-actions">
        <span class="save-state">{{ saveMessage }}</span>
        <button class="liquid-button action-button" type="button" :disabled="saving" @click="saveEditor">
          <Save :size="15" />
          <span>{{ saving ? '保存中' : '保存' }}</span>
        </button>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Pencil, Plus, RefreshCw, Save, Trash2, X } from 'lucide-vue-next'
import {
  createCommonResource,
  deleteCommonResource,
  fetchCommonResource,
  fetchCommonResources,
  updateCommonResource,
  type CommonResourceDefinition,
} from '@/api/monitor'

const resources = ref<CommonResourceDefinition[]>([])
const selectedName = ref('')
const rows = ref<Record<string, unknown>[]>([])
const loadedResource = ref<CommonResourceDefinition | null>(null)
const resourceError = ref('')
const loadingResources = ref(false)
const loadingRows = ref(false)
const saving = ref(false)
const error = ref('')
const saveMessage = ref('')
const editorOpen = ref(false)
const editorMode = ref<'create' | 'edit'>('create')
const editorId = ref<string | number>('')
const editorText = ref('{}')

const currentResource = computed(() => resources.value.find((item) => item.name === selectedName.value) || null)
const activeResource = computed(() => loadedResource.value || currentResource.value)
const resourceStateLabel = computed(() => {
  if (!activeResource.value) return '未加载'
  if (activeResource.value.available === false) return '当前固件不支持'
  return activeResource.value.writable ? '可读写' : '只读'
})

const columns = computed(() => {
  const preferred = ['id', 'enabled', 'tagname', 'name', 'comment', 'interface', 'domain', 'ip_addr', 'mac', 'wan_port', 'lan_addr', 'lan_port', 'dst_addr', 'gateway']
  const keys = new Set<string>()
  for (const key of preferred) {
    if (rows.value.some((row) => key in row)) keys.add(key)
  }
  for (const row of rows.value) {
    for (const key of Object.keys(row)) {
      if (keys.size >= 10) break
      keys.add(key)
    }
  }
  return Array.from(keys)
})

async function loadResources() {
  loadingResources.value = true
  error.value = ''
  try {
    resources.value = await fetchCommonResources()
    selectedName.value = selectedName.value || resources.value[0]?.name || ''
  } catch (err) {
    error.value = err instanceof Error ? err.message : '资源列表请求失败'
  } finally {
    loadingResources.value = false
  }
}

async function loadSelected() {
  if (!selectedName.value) return
  loadingRows.value = true
  error.value = ''
  saveMessage.value = ''
  loadedResource.value = null
  resourceError.value = ''
  try {
    const data = await fetchCommonResource(selectedName.value)
    loadedResource.value = data.resource
    resourceError.value = data.error || ''
    rows.value = data.rows || []
  } catch (err) {
    error.value = err instanceof Error ? err.message : '资源请求失败'
    loadedResource.value = null
    resourceError.value = ''
    rows.value = []
  } finally {
    loadingRows.value = false
  }
}

function startCreate() {
  if (!activeResource.value) return
  editorMode.value = 'create'
  editorId.value = ''
  editorText.value = JSON.stringify(samplePayload(activeResource.value.name), null, 2)
  editorOpen.value = true
  saveMessage.value = ''
}

function startEdit(row: Record<string, unknown>) {
  const id = rowId(row)
  if (!id) return
  editorMode.value = 'edit'
  editorId.value = id
  editorText.value = JSON.stringify(row, null, 2)
  editorOpen.value = true
  saveMessage.value = ''
}

function closeEditor() {
  editorOpen.value = false
  saveMessage.value = ''
}

async function saveEditor() {
  if (!activeResource.value) return
  let payload: Record<string, unknown>
  try {
    payload = JSON.parse(editorText.value)
  } catch {
    saveMessage.value = 'JSON 格式无效'
    return
  }
  saving.value = true
  saveMessage.value = ''
  try {
    if (editorMode.value === 'create') {
      await createCommonResource(activeResource.value.name, payload)
    } else {
      await updateCommonResource(activeResource.value.name, editorId.value, payload)
    }
    saveMessage.value = '已保存'
    editorOpen.value = false
    await loadSelected()
  } catch (err) {
    saveMessage.value = err instanceof Error ? err.message : '保存失败'
  } finally {
    saving.value = false
  }
}

async function deleteRow(row: Record<string, unknown>) {
  if (!activeResource.value) return
  const id = rowId(row)
  if (!id) return
  if (!window.confirm(`删除 ${activeResource.value.label} #${id}？`)) return
  loadingRows.value = true
  error.value = ''
  try {
    await deleteCommonResource(activeResource.value.name, id)
    await loadSelected()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '删除失败'
  } finally {
    loadingRows.value = false
  }
}

function rowId(row: Record<string, unknown>) {
  const value = row.id ?? row.rowid
  if (typeof value === 'number' || typeof value === 'string') return value
  return ''
}

function rowKey(row: Record<string, unknown>, index: number) {
  return `${rowId(row) || 'row'}-${index}`
}

function formatCell(value: unknown) {
  if (value === null || value === undefined || value === '') return '—'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

function samplePayload(name: string): Record<string, unknown> {
  const base = { enabled: 'yes', comment: 'created by ikuai-dashboard' }
  switch (name) {
    case 'dhcp-static':
      return { ...base, tagname: 'device-name', mac: 'AA:BB:CC:DD:EE:FF', ip_addr: '192.168.50.10' }
    case 'dns-static':
      return { ...base, tagname: 'local-host', domain: 'host.local', ip_addr: '192.168.50.10' }
    case 'dnat-rules':
      return { ...base, tagname: 'ssh', interface: 'wan1', protocol: 'tcp', wan_port: '2222', lan_addr: '192.168.50.10', lan_port: '22' }
    case 'static-routes':
      return { ...base, tagname: 'route-name', dst_addr: '10.10.0.0/16', gateway: '192.168.50.1', interface: 'lan1', metric: 10 }
    default:
      return base
  }
}

onMounted(async () => {
  await loadResources()
  await loadSelected()
})

watch(selectedName, () => {
  closeEditor()
  void loadSelected()
})
</script>

<style scoped>
.resources-page {
  display: flex;
  flex-direction: column;
  gap: 22px;
}

.page-head,
.resource-meta,
.editor-head,
.editor-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.page-head,
.resource-meta,
.editor-head {
  padding: 18px 20px;
}

.panel-title {
  color: var(--text-primary);
  font-size: 1.25rem;
  font-weight: 720;
}

.panel-subtitle,
.resource-meta p,
.editor-head p,
.save-state {
  color: var(--text-secondary);
  font-size: 12px;
}

.head-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.resource-select {
  min-height: 38px;
  min-width: min(280px, 58vw);
  padding: 0 38px 0 12px;
  color: var(--text-primary);
  border: 1px solid var(--control-border);
  border-radius: 10px;
  background: var(--glass-bg-strong);
}

.icon-button,
.small-button {
  width: 38px;
  height: 38px;
}

.small-button {
  width: 32px;
  height: 32px;
}

.action-button {
  min-height: 38px;
  padding: 0 14px;
}

.danger {
  color: var(--system-red);
}

.resource-panel,
.editor-panel {
  overflow: hidden;
}

.resource-meta,
.editor-head {
  border-bottom: 1px solid var(--glass-border);
}

.resource-meta h3,
.editor-head h3 {
  color: var(--text-primary);
  font-size: 15px;
  font-weight: 720;
}

.table-wrap {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  min-width: 760px;
  border-collapse: collapse;
}

.data-table th,
.data-table td {
  max-width: 260px;
  padding: 12px 14px;
  overflow: hidden;
  text-align: left;
  text-overflow: ellipsis;
  white-space: nowrap;
  border-bottom: 1px solid var(--control-border);
}

.data-table th {
  color: var(--text-secondary);
  background: var(--table-header-bg);
  font-size: 12px;
  font-weight: 680;
}

.data-table td {
  color: var(--text-primary);
  font-size: 13px;
}

.row-actions {
  display: flex;
  gap: 8px;
}

.empty-row {
  padding: 34px !important;
  color: var(--text-tertiary) !important;
  text-align: center !important;
}

.json-editor {
  display: block;
  width: calc(100% - 40px);
  min-height: 240px;
  margin: 20px;
  padding: 14px;
  color: var(--text-primary);
  border: 1px solid var(--control-border);
  border-radius: var(--radius-sm);
  background: rgba(0, 0, 0, 0.24);
  font: 13px/1.6 "SF Mono", ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  resize: vertical;
}

.json-editor:focus {
  outline: 2px solid var(--system-blue);
  outline-offset: 2px;
}

.editor-actions {
  padding: 0 20px 20px;
}

.spinning {
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 760px) {
  .page-head,
  .resource-meta,
  .editor-head,
  .editor-actions {
    align-items: stretch;
    flex-direction: column;
  }

  .head-actions,
  .resource-select {
    width: 100%;
  }

  .icon-button,
  .action-button {
    width: 100%;
  }
}
</style>

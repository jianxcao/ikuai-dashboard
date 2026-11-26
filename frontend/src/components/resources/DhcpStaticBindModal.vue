<template>
  <teleport to="body">
    <div v-if="open" class="drawer-backdrop center" @click.self="emit('close')">
      <div class="confirm-dialog v4-panel modal-wide" role="dialog" aria-label="将动态租约转为 DHCP 静态绑定">
        <div class="modal-top">
          <h4>将动态租约转为 DHCP 静态绑定</h4>
          <button type="button" class="icon-btn" aria-label="关闭弹窗" @click="emit('close')">
            <X :size="14" />
          </button>
        </div>

        <div v-if="error" class="error-banner">{{ error }}</div>

        <div class="form-grid">
          <label class="form-field">
            <span class="field-label">设备名称</span>
            <input v-model="form.tagname" class="v4-input" />
          </label>
          <label class="form-field">
            <span class="field-label">固定 IP</span>
            <input v-model="form.ip_addr" class="v4-input font-mono" />
          </label>
          <label class="form-field">
            <span class="field-label">MAC 地址</span>
            <input v-model="form.mac" class="v4-input font-mono" />
          </label>
          <div class="form-field">
            <span class="field-label">所属接口</span>
            <InterfaceSelect v-model="form.interface" scope="lan" />
          </div>
        </div>

        <div class="confirm-actions">
          <button type="button" class="v4-btn" @click="emit('close')">取消</button>
          <button type="button" class="v4-btn v4-btn-primary" :disabled="busy" @click="submit">
            {{ busy ? '绑定中...' : '确认绑定' }}
          </button>
        </div>
      </div>
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { X } from 'lucide-vue-next'
import { createV4Resource } from '@/api/v4'
import InterfaceSelect from '@/components/common/InterfaceSelect.vue'

const props = defineProps<{
  open: boolean
  leaseRow: Record<string, unknown> | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'bound', ip: string): void
}>()

const busy = ref(false)
const error = ref<string | null>(null)
const form = reactive({
  tagname: '',
  ip_addr: '',
  mac: '',
  interface: 'lan1',
})

watch(
  () => [props.open, props.leaseRow] as const,
  ([isOpen, row]) => {
    if (!isOpen || !row) return
    error.value = null
    form.tagname = String(row.hostname || 'StaticDevice')
    form.ip_addr = String(row.ip_addr || '')
    form.mac = String(row.mac || '')
    form.interface = String(row.interface || 'lan1')
  },
  { immediate: true }
)

async function submit() {
  busy.value = true
  error.value = null
  try {
    await createV4Resource('dhcp-static', {
      tagname: form.tagname,
      ip_addr: form.ip_addr,
      mac: form.mac,
      interface: form.interface,
      enabled: 'yes',
    })
    emit('bound', form.ip_addr)
    emit('close')
  } catch (e: unknown) {
    if (e && typeof e === 'object' && 'response' in e) {
      const resp = (e as { response?: { data?: { message?: string } } }).response
      if (resp?.data?.message) {
        error.value = resp.data.message
        return
      }
    }
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    busy.value = false
  }
}
</script>

<style scoped>
.drawer-backdrop {
  position: fixed;
  inset: 0;
  z-index: 255;
  background: rgba(6, 10, 20, 0.42);
  -webkit-backdrop-filter: blur(8px);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.confirm-dialog {
  width: min(520px, calc(100vw - 32px));
  max-height: 86dvh;
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
  justify-content: space-between;
  align-items: center;
}

.modal-top h4 {
  margin: 0;
  font-size: 15px;
  color: var(--text-strong);
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

.icon-btn:hover {
  border-color: var(--accent-primary);
  color: var(--text-strong);
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

.confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 4px;
}

@media (max-width: 640px) {
  .confirm-dialog {
    width: calc(100vw - 20px);
    padding: 16px;
  }
}
</style>

<template>
  <teleport to="body">
    <div v-if="open" class="drawer-backdrop center" @click.self="emit('close')">
      <div class="confirm-dialog v4-panel" role="dialog" :aria-label="modalTitle">
        <div class="modal-top">
          <h4>{{ modalTitle }}</h4>
          <button type="button" class="v4-btn btn-xs" aria-label="关闭弹窗" @click="emit('close')">
            <X :size="14" />
          </button>
        </div>

        <div v-if="error" class="error-banner">{{ error }}</div>

        <div class="modal-form">
          <label class="field-item">
            <span>名称 / 标签</span>
            <input v-model.trim="form.tagname" class="v4-input" />
          </label>
          <label v-if="kind === 'static' || kind === 'qos'" class="field-item">
            <span>IP 地址</span>
            <input v-model.trim="form.ip_addr" class="v4-input font-mono" />
          </label>
          <label v-if="kind !== 'qos'" class="field-item">
            <span>MAC 地址</span>
            <input v-model.trim="form.mac" class="v4-input font-mono" />
          </label>
          <div v-if="kind === 'static' || kind === 'qos'" class="field-item">
            <span>生效接口</span>
            <InterfaceSelect
              v-model="form.iface"
              :scope="kind === 'qos' ? 'wan-all' : 'lan'"
            />
          </div>
          <template v-if="kind === 'qos'">
            <label class="field-item">
              <span>上行限速 (Kbps)</span>
              <input v-model.number="form.upload" type="number" class="v4-input font-mono" />
            </label>
            <label class="field-item">
              <span>下行限速 (Kbps)</span>
              <input v-model.number="form.download" type="number" class="v4-input font-mono" />
            </label>
          </template>
          <label v-if="kind === 'mac-block'" class="field-item">
            <span>规则状态</span>
            <select v-model="form.enabled" class="v4-select">
              <option value="yes">立即生效 (enabled=yes)</option>
              <option value="no">先创建暂不生效 (enabled=no)</option>
            </select>
          </label>
          <label v-if="kind !== 'qos'" class="field-item">
            <span>备注</span>
            <input v-model.trim="form.comment" class="v4-input" />
          </label>
        </div>

        <div class="confirm-actions">
          <button type="button" class="v4-btn" @click="emit('close')">取消</button>
          <button type="button" class="v4-btn v4-btn-primary" :disabled="busy" @click="submit">
            {{ busy ? '执行中...' : '确认应用' }}
          </button>
        </div>
      </div>
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { X } from 'lucide-vue-next'
import { createV4Resource, type ClientRecordDTO } from '@/api/v4'
import InterfaceSelect from '@/components/common/InterfaceSelect.vue'

export type QuickActionKind = 'terminal' | 'static' | 'qos' | 'mac-block'

const props = defineProps<{
  open: boolean
  kind: QuickActionKind
  client: ClientRecordDTO | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'applied', message: string): void
}>()

const busy = ref(false)
const error = ref<string | null>(null)
const form = reactive({
  tagname: '',
  ip_addr: '',
  mac: '',
  iface: 'lan1',
  upload: 1000,
  download: 5000,
  enabled: 'yes',
  comment: '',
})

const modalTitle = computed(() => {
  switch (props.kind) {
    case 'terminal':
      return '快捷添加终端设备标注'
    case 'static':
      return '快捷创建 DHCP 静态地址绑定'
    case 'qos':
      return '快捷创建 IP 限速流控规则'
    case 'mac-block':
      return '快捷添加 MAC 访问控制规则'
  }
})

watch(
  () => [props.open, props.kind, props.client] as const,
  ([isOpen, kind, c]) => {
    if (!isOpen || !c) return
    busy.value = false
    error.value = null
    form.tagname = c.comment || c.hostname || c.ip_addr || 'Device'
    form.ip_addr = c.ip_addr || ''
    form.mac = c.mac || ''
    form.iface = kind === 'qos' ? 'wan1' : 'lan1'
    form.upload = 1000
    form.download = 5000
    form.enabled = 'yes'
    form.comment = c.comment || ''
  },
  { immediate: true }
)

async function submit() {
  busy.value = true
  error.value = null
  try {
    if (props.kind === 'terminal') {
      await createV4Resource('terminals', { tagname: form.tagname, mac: form.mac, comment: form.comment })
      emit('applied', `已标注终端 ${form.tagname}`)
    } else if (props.kind === 'static') {
      await createV4Resource('dhcp-static', {
        tagname: form.tagname,
        ip_addr: form.ip_addr,
        mac: form.mac,
        interface: form.iface,
        enabled: 'yes',
        comment: form.comment,
      })
      emit('applied', `已绑定静态 IP ${form.ip_addr}`)
    } else if (props.kind === 'qos') {
      await createV4Resource('qos-ip', {
        tagname: form.tagname,
        ip_addr: form.ip_addr,
        interface: form.iface,
        upload: form.upload,
        download: form.download,
        protocol: 'any',
        enabled: 'yes',
      })
      emit('applied', `已为 ${form.ip_addr} 应用 QoS 限速`)
    } else if (props.kind === 'mac-block') {
      await createV4Resource('mac-rules', {
        tagname: form.tagname,
        mac: form.mac,
        enabled: form.enabled,
        comment: form.comment,
      })
      emit('applied', `已写入 MAC 控制规则 ${form.mac}`)
    }
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
  z-index: 250;
  background: rgba(6, 10, 20, 0.45);
  -webkit-backdrop-filter: blur(8px);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.confirm-dialog {
  width: min(440px, calc(100vw - 32px));
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
}

.modal-form {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.field-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-normal);
}

.confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>

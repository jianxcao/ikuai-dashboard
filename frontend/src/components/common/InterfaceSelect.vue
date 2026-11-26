<template>
  <div class="iface-select-wrap">
    <template v-if="!customMode">
      <select
        ref="selectEl"
        :value="strValue"
        class="v4-select"
        :disabled="disabled"
        @change="onSelectChange"
      >
        <option v-if="strValue === ''" value="" disabled>
          {{ placeholder || '请选择接口' }}
        </option>

        <option v-if="prefixOption" :value="prefixOption.value">
          {{ prefixOption.label }}
        </option>

        <template v-if="showGrouped">
          <optgroup v-if="wanItems.length > 0" label="外网接口 (WAN)">
            <option v-for="item in wanItems" :key="'wan-' + item.name" :value="item.name">
              {{ formatOptionLabel(item) }}
            </option>
          </optgroup>
          <optgroup v-if="lanItems.length > 0" label="内网接口 (LAN / VLAN)">
            <option v-for="item in lanItems" :key="'lan-' + item.name" :value="item.name">
              {{ formatOptionLabel(item) }}
            </option>
          </optgroup>
        </template>

        <template v-else>
          <option v-for="item in flatItems" :key="item.kind + '-' + item.name" :value="item.name">
            {{ formatOptionLabel(item) }}
          </option>
        </template>

        <option v-if="isUnknownCurrent" :value="strValue">
          {{ strValue }} (当前值)
        </option>

        <option :value="CUSTOM_SENTINEL">✎ 手动输入接口 / IP...</option>
      </select>
    </template>

    <div v-else class="custom-row">
      <input
        ref="inputEl"
        :value="strValue"
        type="text"
        class="v4-input font-mono"
        :disabled="disabled"
        :placeholder="placeholder || '输入接口名或 IP'"
        @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      />
      <button
        type="button"
        class="v4-btn btn-xs list-toggle-btn"
        :disabled="disabled"
        title="切回接口下拉列表"
        @click.prevent="customMode = false"
      >
        列表
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import type { InterfaceOptionItem, InterfaceScope } from '@/api/v4'
import { useInterfaceOptions } from '@/composables/useInterfaceOptions'

const CUSTOM_SENTINEL = '__custom_input__'

const props = defineProps<{
  modelValue: unknown
  scope: InterfaceScope
  disabled?: boolean
  placeholder?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const { options, ensureLoaded } = useInterfaceOptions()
const customMode = ref(false)
const selectEl = ref<HTMLSelectElement | null>(null)
const inputEl = ref<HTMLInputElement | null>(null)

onMounted(() => {
  void ensureLoaded()
})

const strValue = computed(() =>
  props.modelValue === undefined || props.modelValue === null ? '' : String(props.modelValue)
)

const prefixOption = computed(() => {
  switch (props.scope) {
    case 'wan-all':
      return { value: 'all', label: '全部外网线路 (all)' }
    case 'wan-auto':
      return { value: 'auto', label: '自动选择出口 (auto)' }
    case 'any':
      return { value: 'any', label: '任意接口 (any)' }
    case 'physical-all':
      return { value: 'all', label: '全部物理网口 (all)' }
    default:
      return null
  }
})

const wanItems = computed<readonly InterfaceOptionItem[]>(() => {
  if (!['wan', 'wan-all', 'wan-auto', 'wan-lan', 'any'].includes(props.scope)) return []
  if (options.value.wan.length > 0) return options.value.wan
  return [{ name: 'wan1', kind: 'wan' }]
})

const lanItems = computed<readonly InterfaceOptionItem[]>(() => {
  if (!['lan', 'wan-lan', 'any'].includes(props.scope)) return []
  if (options.value.lan.length > 0) return options.value.lan
  return [{ name: 'lan1', kind: 'lan' }]
})

const physicalItems = computed<readonly InterfaceOptionItem[]>(() => {
  if (props.scope !== 'physical-all') return []
  return options.value.physical
})

const showGrouped = computed(() => props.scope === 'wan-lan' || props.scope === 'any')

const flatItems = computed<readonly InterfaceOptionItem[]>(() => {
  if (props.scope === 'physical-all') return physicalItems.value
  if (props.scope === 'lan') return lanItems.value
  return wanItems.value
})

const knownNames = computed(() => {
  const set = new Set<string>()
  if (prefixOption.value) set.add(prefixOption.value.value)
  for (const item of wanItems.value) set.add(item.name)
  for (const item of lanItems.value) set.add(item.name)
  for (const item of physicalItems.value) set.add(item.name)
  return set
})

const isUnknownCurrent = computed(
  () => strValue.value !== '' && !knownNames.value.has(strValue.value)
)

watch(
  [wanItems, lanItems, physicalItems, strValue],
  () => {
    if (selectEl.value && !customMode.value && selectEl.value.value !== strValue.value) {
      selectEl.value.value = strValue.value
    }
  },
  { flush: 'post' }
)

function formatOptionLabel(item: InterfaceOptionItem): string {
  return item.detail ? `${item.name} · ${item.detail}` : item.name
}

async function onSelectChange(e: Event) {
  const val = (e.target as HTMLSelectElement).value
  if (val === CUSTOM_SENTINEL) {
    customMode.value = true
    await nextTick()
    inputEl.value?.focus()
    return
  }
  emit('update:modelValue', val)
}
</script>

<style scoped>
.iface-select-wrap {
  width: 100%;
}

.custom-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.custom-row .v4-input {
  flex: 1;
  min-width: 0;
}

.list-toggle-btn {
  flex-shrink: 0;
  min-height: 36px;
}
</style>

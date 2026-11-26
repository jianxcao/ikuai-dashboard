<template>
  <div class="obj-field-bar">
    <template v-if="mode === 'single'">
      <div class="single-combo">
        <select
          v-model="selectedPreset"
          class="v4-select"
          :disabled="disabled || loadingList"
          @change="handleSinglePresetChange"
        >
          <option value="">
            {{ loadingList ? '加载分组中...' : `选择${meta.label}...` }}
          </option>
          <option v-for="item in groupOptions" :key="item.name" :value="item.name">
            @{{ item.name }}{{ item.summary ? ` — ${item.summary}` : '' }}
          </option>
        </select>

        <input
          :value="singleValue ?? ''"
          type="text"
          class="v4-input font-mono"
          placeholder="或直接输入域名对象组名"
          :disabled="disabled"
          @input="emit('update:singleValue', ($event.target as HTMLInputElement).value)"
        />
      </div>

      <div class="bar-footer">
        <button
          v-if="(singleValue ?? '').trim()"
          type="button"
          class="obj-group-pill"
          @click="emit('inspect', { groupKey, groupName: (singleValue ?? '').trim() })"
        >
          <span>查看/快改 @{{ (singleValue ?? '').trim() }}</span>
        </button>
        <button
          type="button"
          class="manage-link"
          @click="jumpToObjectGroup(groupKey, (singleValue ?? '').trim() || undefined)"
        >
          <ExternalLink :size="12" />
          <span>管理{{ meta.label }}</span>
        </button>
      </div>
    </template>

    <template v-else>
      <div v-if="selectedList.length > 0" class="selected-chips">
        <span v-for="name in selectedList" :key="name" class="obj-group-chip">
          <button
            type="button"
            class="chip-inspect"
            title="点击查看或快速修改该对象组"
            @click="emit('inspect', { groupKey, groupName: name })"
          >
            @{{ name }}
          </button>
          <button
            v-if="!disabled"
            type="button"
            class="chip-remove"
            title="从当前字段移除该分组引用"
            @click="removeGroup(name)"
          >
            <X :size="11" />
          </button>
        </span>
      </div>

      <div class="picker-row">
        <select
          v-model="pendingPick"
          class="v4-select picker-select"
          :disabled="disabled || loadingList"
          @change="handlePickGroup"
        >
          <option value="">
            {{
              loadingList
                ? `正在读取${meta.label}...`
                : availableOptions.length > 0
                  ? `+ 引用${meta.label} (${availableOptions.length})`
                  : groupOptions.length > 0
                    ? `已引用全部可用${meta.label}`
                    : `暂无可用${meta.label}`
            }}
          </option>
          <option v-for="item in availableOptions" :key="item.name" :value="item.name">
            @{{ item.name }}{{ item.summary ? ` (${item.summary})` : '' }}
          </option>
        </select>

        <button
          type="button"
          class="v4-btn btn-xs manage-btn"
          :title="`跳转到${meta.label}管理页面`"
          @click="jumpToObjectGroup(groupKey)"
        >
          <ExternalLink :size="12" />
          <span>管理分组</span>
        </button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onActivated, onMounted, ref, watch } from 'vue'
import { ExternalLink, X } from 'lucide-vue-next'
import {
  OBJECT_GROUP_META,
  type ObjectGroupKey,
  useObjectGroups,
} from '@/composables/useObjectGroups'

const props = withDefaults(
  defineProps<{
    groupKey: ObjectGroupKey
    mode?: 'compound' | 'single'
    selectedGroups?: string[]
    singleValue?: string
    disabled?: boolean
  }>(),
  {
    mode: 'compound',
    selectedGroups: () => [],
    singleValue: '',
    disabled: false,
  }
)

const emit = defineEmits<{
  (e: 'update:selectedGroups', value: string[]): void
  (e: 'update:singleValue', value: string): void
  (e: 'inspect', payload: { groupKey: ObjectGroupKey; groupName: string }): void
}>()

const { groupCache, groupLoading, loadObjectGroupList, jumpToObjectGroup } = useObjectGroups()

const pendingPick = ref('')
const selectedPreset = ref('')

const meta = computed(() => OBJECT_GROUP_META[props.groupKey])
const loadingList = computed(() => Boolean(groupLoading[props.groupKey]))
const selectedList = computed(() => props.selectedGroups || [])

const groupOptions = computed(() => {
  const rows = groupCache[props.groupKey] || []
  return rows
    .map((r) => {
      const name = String(r.group_name ?? '').trim()
      if (!name) return null
      const summary = String(r.value ?? r.comment ?? '').trim()
      return {
        name,
        summary: summary.length > 36 ? `${summary.slice(0, 36)}…` : summary,
      }
    })
    .filter((x): x is { name: string; summary: string } => x !== null)
})

const availableOptions = computed(() => {
  const used = new Set(selectedList.value)
  return groupOptions.value.filter((opt) => !used.has(opt.name))
})

function syncSinglePreset() {
  const val = (props.singleValue ?? '').trim()
  selectedPreset.value = groupOptions.value.some((o) => o.name === val) ? val : ''
}

function handleSinglePresetChange() {
  if (!selectedPreset.value) return
  emit('update:singleValue', selectedPreset.value)
}

function handlePickGroup() {
  const picked = pendingPick.value.trim()
  pendingPick.value = ''
  if (!picked || selectedList.value.includes(picked)) return
  emit('update:selectedGroups', [...selectedList.value, picked])
}

function removeGroup(name: string) {
  emit(
    'update:selectedGroups',
    selectedList.value.filter((item) => item !== name)
  )
}

watch(
  () => [props.singleValue, groupOptions.value.length] as const,
  () => syncSinglePreset(),
  { immediate: true }
)

watch(
  () => props.groupKey,
  (key) => {
    void loadObjectGroupList(key)
  },
  { immediate: true }
)

onMounted(() => {
  void loadObjectGroupList(props.groupKey)
})

onActivated(() => {
  void loadObjectGroupList(props.groupKey)
})
</script>

<style scoped>
.obj-field-bar {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 2px;
}

.selected-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
}

.obj-group-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--accent-primary) 32%, transparent);
  background: var(--accent-primary-soft);
  color: var(--accent-text);
  overflow: hidden;
}

.chip-inspect {
  border: none;
  background: transparent;
  color: inherit;
  font-size: 11px;
  font-weight: 600;
  padding: 2px 7px;
  cursor: pointer;
}

.chip-inspect:hover {
  text-decoration: underline;
}

.chip-remove {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-left: 1px solid color-mix(in srgb, var(--accent-primary) 24%, transparent);
  background: transparent;
  color: inherit;
  padding: 3px 6px 3px 4px;
  cursor: pointer;
}

.chip-remove:hover {
  background: color-mix(in srgb, var(--status-down) 16%, transparent);
  color: var(--status-down);
}

.picker-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.picker-select {
  flex: 1;
  min-width: 0;
  min-height: 32px;
  font-size: 12px;
}

.manage-btn {
  flex-shrink: 0;
  white-space: nowrap;
}

.single-combo {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.bar-footer {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.manage-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  font-size: 11.5px;
  font-weight: 600;
  padding: 2px 0;
  cursor: pointer;
}

.manage-link:hover {
  color: var(--accent-text);
}
</style>

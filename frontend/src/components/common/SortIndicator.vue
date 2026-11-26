<template>
  <span class="sort-indicator" :class="{ active }">
    <ArrowUp v-if="active && dir === 'asc'" :size="12" />
    <ArrowDown v-else-if="active && dir === 'desc'" :size="12" />
    <ArrowUpDown v-else :size="11" />
  </span>
</template>

<script setup lang="ts">
import { ArrowDown, ArrowUp, ArrowUpDown } from 'lucide-vue-next'
import type { SortDirection } from '@/composables/useTableSort'

defineProps<{
  active: boolean
  dir: SortDirection
}>()
</script>

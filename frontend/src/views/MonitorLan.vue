<template>
  <div class="page-container">
    <div v-if="error" class="error-banner">
      <span class="error-icon">!</span>
      <span>获取数据失败：{{ error }}</span>
    </div>

    <div class="controls-row">
      <SearchBar 
        :modelValue="search" 
        @update:modelValue="onSearch" 
        class="flex-1"
      />
    </div>
    
    <ClientList 
      :clients="clients" 
      :loading="loading" 
      v-model:sortBy="sortBy"
    />
  </div>
</template>

<script setup lang="ts">
import SearchBar from '../components/lan/SearchBar.vue'
import ClientList from '../components/lan/ClientList.vue'
import { useLanClients } from '@/composables/useLanClients'

const {
  loading,
  error,
  clients,
  search,
  sortBy,
  onSearch
} = useLanClients()
</script>

<style scoped>
.page-container {
  display: flex;
  flex-direction: column;
  animation: fade-in 0.4s ease-out;
}

.controls-row {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  margin-bottom: 20px;
}

.flex-1 {
  flex: 1;
  margin-bottom: 0 !important; /* override SearchBar margin */
}



@keyframes fade-in {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

</style>

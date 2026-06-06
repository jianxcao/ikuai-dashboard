<template>
  <div class="page-container">
    <div v-if="error" class="error-banner">
      <span class="error-icon">!</span>
      <span>获取数据失败：{{ error }}</span>
    </div>

    <section class="lan-toolbar glass-panel">
      <div class="toolbar-meta">
        <span class="toolbar-label">在线终端</span>
        <strong>{{ clients.length }}</strong>
      </div>
      <SearchBar :modelValue="search" @update:modelValue="onSearch" class="search-shell" />
    </section>

    <ClientList :clients="clients" :loading="loading" v-model:sortBy="sortBy" />
  </div>
</template>

<script setup lang="ts">
import SearchBar from '../components/lan/SearchBar.vue'
import ClientList from '../components/lan/ClientList.vue'
import { useLanClients } from '@/composables/useLanClients'

const { loading, error, clients, search, sortBy, onSearch } = useLanClients()
</script>

<style scoped>
.page-container {
  display: flex;
  flex-direction: column;
  gap: 18px;
  animation: fade-in 0.4s ease-out;
}

.lan-toolbar {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 14px;
  align-items: center;
  padding: 14px 16px;
}

.toolbar-meta strong {
  color: var(--text-primary);
  font-size: 1rem;
  font-weight: 760;
}

.toolbar-label {
  display: block;
  margin-bottom: 4px;
  color: var(--text-tertiary);
  font-size: 11px;
  font-weight: 760;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.search-shell {
  margin-bottom: 0 !important;
}

@keyframes fade-in {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 860px) {
  .lan-toolbar {
    grid-template-columns: 1fr;
  }
}
</style>

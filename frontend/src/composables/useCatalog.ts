import { ref } from 'vue'
import { fetchV4Catalog, type ResourceSpec } from '@/api/v4'

const catalog = ref<ResourceSpec[]>([])
const loading = ref(false)
let loaded = false

async function ensureCatalog() {
  if (loaded || loading.value) return
  loading.value = true
  try {
    catalog.value = await fetchV4Catalog()
    loaded = true
  } finally {
    loading.value = false
  }
}

export function useCatalog() {
  void ensureCatalog()

  function getSpec(key: string): ResourceSpec | undefined {
    return catalog.value.find((item) => item.key === key)
  }

  function getGroupSpecs(group: string): ResourceSpec[] {
    return catalog.value.filter((item) => item.group === group)
  }

  return {
    catalog,
    loading,
    getSpec,
    getGroupSpecs,
    reloadCatalog: ensureCatalog,
  }
}

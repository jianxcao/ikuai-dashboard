import { ref } from 'vue'
import { fetchV4Catalog, type ResourceSpec } from '@/api/v4'

const catalog = ref<ResourceSpec[]>([])
const loading = ref(false)
let loaded = false
let inFlight: Promise<void> | null = null

async function ensureCatalog(): Promise<void> {
  if (loaded) return
  if (inFlight) return inFlight
  loading.value = true
  inFlight = (async () => {
    try {
      catalog.value = await fetchV4Catalog()
      loaded = true
    } catch {
      // Allow retry once authenticated instead of emitting unhandled rejection
    } finally {
      loading.value = false
      inFlight = null
    }
  })()
  return inFlight
}

export function useCatalog(options?: { lazy?: boolean }) {
  if (!options?.lazy) {
    void ensureCatalog()
  }

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

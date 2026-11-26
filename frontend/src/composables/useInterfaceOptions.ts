import { readonly, ref, watch } from 'vue'
import { fetchV4InterfaceOptions, type InterfaceOptionsDTO } from '@/api/v4'
import { useFleet } from './useFleet'

const EMPTY_OPTIONS: InterfaceOptionsDTO = {
  wan: [],
  lan: [],
  physical: [],
}

const options = ref<InterfaceOptionsDTO>(EMPTY_OPTIONS)
const loading = ref(false)
let loadedEpoch = -1
let inflight: Promise<void> | null = null
let inflightEpoch = -1
let watcherAttached = false

export function useInterfaceOptions() {
  const { routerEpoch } = useFleet()

  if (!watcherAttached) {
    watcherAttached = true
    watch(routerEpoch, (nextEpoch) => {
      if (loadedEpoch !== nextEpoch) {
        loadedEpoch = -1
        void refresh()
      }
    })
  }

  async function refresh(): Promise<void> {
    const targetEpoch = routerEpoch.value
    if (inflight && inflightEpoch === targetEpoch) return inflight
    loading.value = true
    inflightEpoch = targetEpoch
    const promise = fetchV4InterfaceOptions()
      .then((res) => {
        if (routerEpoch.value === targetEpoch) {
          options.value = {
            wan: res.wan || [],
            lan: res.lan || [],
            physical: res.physical || [],
          }
          loadedEpoch = targetEpoch
        }
      })
      .catch(() => {
        // Keep existing options or fallback defaults on transient errors
      })
      .finally(() => {
        if (inflightEpoch === targetEpoch) {
          loading.value = false
          inflight = null
        }
      })
    inflight = promise
    return promise
  }

  async function ensureLoaded(): Promise<void> {
    if (loadedEpoch === routerEpoch.value) return
    await refresh()
  }

  return {
    options: readonly(options),
    loading: readonly(loading),
    ensureLoaded,
    refresh,
  }
}

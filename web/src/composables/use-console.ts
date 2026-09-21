import { computed, ref } from 'vue'
import { toast } from 'vue-sonner'

import { api } from '@/lib/api'
import type { EntryGroup, Node, Settings, Status } from '@/types'

const status = ref<Status>()
const nodes = ref<Node[]>([])
const entryGroups = ref<EntryGroup[]>([])
const settings = ref<Settings>({
  listen: '127.0.0.1:9080', mixedPort: 7890, allowLan: true,
  bindAddress: '*', externalAddress: '', externalPort: 27890,
})
const loading = ref(false)
const busy = ref(false)

export function useConsole() {
  const selectedNode = computed(() => nodes.value.find(node => node.name === settings.value.selectedNode))
  const nodeById = computed(() => new Map(nodes.value.map(node => [node.id, node])))

  async function load() {
    loading.value = true
    try {
      ;[status.value, nodes.value, entryGroups.value, settings.value] = await Promise.all([
        api.status(), api.nodes(), api.entryGroups(), api.settings(),
      ])
    }
    catch (error) {
      toast.error((error as Error).message)
    }
    finally {
      loading.value = false
    }
  }

  async function run(action: () => Promise<unknown>, success: string) {
    busy.value = true
    try {
      await action()
      toast.success(success)
      await load()
      return true
    }
    catch (error) {
      toast.error((error as Error).message)
      return false
    }
    finally {
      busy.value = false
    }
  }

  return { status, nodes, entryGroups, settings, selectedNode, nodeById, loading, busy, load, run }
}

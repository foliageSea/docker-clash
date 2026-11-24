<script setup lang="ts">
import { ArrowRightIcon, CopyIcon, PowerIcon } from '@lucide/vue'
import { computed } from 'vue'
import { toast } from 'vue-sonner'

import PageHeader from '@/components/PageHeader.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { useConsole } from '@/composables/use-console'
import { api } from '@/lib/api'

const { status, nodes, entryGroups, settings, selectedNode, busy, run } = useConsole()
const metrics = computed(() => [
  {
    label: '内核状态',
    value: status.value ? (status.value.core.running ? '运行中' : '已停止') : '未知',
    detail: status.value?.core.pid ? `PID ${status.value.core.pid}` : undefined,
  },
  { label: '代理节点', value: status.value ? nodes.value.length : '—' },
  { label: '入口组', value: status.value ? entryGroups.value.length : '—' },
  { label: 'Mixed 端口', value: status.value ? settings.value.mixedPort : '—', detail: 'HTTP + SOCKS5' },
])
const endpoints = computed(() => {
  const address = settings.value.externalAddress.trim()
  if (!address) return []
  const endpoint = `${address}:${settings.value.externalPort || settings.value.mixedPort}`
  return [endpoint, `http://${endpoint}`]
})

async function copy(value: string) {
  try {
    await navigator.clipboard.writeText(value)
    toast.success('代理地址已复制')
  }
  catch {
    window.prompt('请复制代理地址', value)
  }
}
</script>

<template>
  <PageHeader title="运行概览">
    <Button :variant="status?.core.running ? 'destructive' : 'default'" :disabled="busy || !status" @click="run(() => api.coreAction(status?.core.running ? 'stop' : 'start'), status?.core.running ? '内核已停止' : '内核已启动')">
      <PowerIcon data-icon="inline-start" />{{ status?.core.running ? '停止内核' : '启动内核' }}
    </Button>
  </PageHeader>

  <div class="grid grid-cols-2 gap-3 @min-[46rem]:grid-cols-4 sm:gap-4">
    <Card v-for="metric in metrics" :key="metric.label" class="min-w-0 py-4">
      <CardContent class="px-4">
        <dl class="flex flex-col gap-1">
          <dt class="text-sm text-muted-foreground">{{ metric.label }}</dt>
          <dd class="text-xl font-semibold tabular-nums">{{ metric.value }}</dd>
          <dd v-if="metric.detail" class="text-xs text-muted-foreground">{{ metric.detail }}</dd>
        </dl>
      </CardContent>
    </Card>
  </div>

  <div class="mt-4 grid gap-4 @min-[46rem]:grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)] sm:mt-6">
    <Card class="min-w-0 gap-4">
      <CardHeader><CardTitle>默认出口</CardTitle></CardHeader>
      <CardContent class="flex flex-col gap-4">
        <div v-if="selectedNode" class="rounded-md border p-3">
          <strong class="block wrap-anywhere">{{ selectedNode.name }}</strong>
          <div class="flex flex-wrap items-baseline gap-x-3 gap-y-1 text-xs text-muted-foreground">
            <span>{{ selectedNode.type.toUpperCase() }}</span>
            <span class="font-mono wrap-anywhere">{{ selectedNode.server }}:{{ selectedNode.port }}</span>
          </div>
        </div>
        <p v-else class="text-sm text-muted-foreground">{{ status ? '尚未选择出口节点' : '出口状态不可用' }}</p>
        <div v-if="nodes.length" role="group" aria-label="切换默认出口" class="flex flex-wrap gap-2">
          <Button v-for="node in nodes.slice(0, 6)" :key="node.id" size="sm" class="min-w-0 max-w-full" :variant="node.id === selectedNode?.id ? 'default' : 'outline'" :aria-pressed="node.id === selectedNode?.id" :disabled="busy" @click="run(() => api.selectNode(node.id), '默认出口已切换')"><span class="truncate">{{ node.name }}</span></Button>
        </div>
      </CardContent>
      <CardFooter v-if="status && (nodes.length === 0 || nodes.length > 6)" class="mt-auto">
        <Button variant="outline" size="sm" as-child><RouterLink to="/nodes">{{ nodes.length ? '全部节点' : '前往代理节点' }}<ArrowRightIcon data-icon="inline-end" /></RouterLink></Button>
      </CardFooter>
    </Card>

    <Card class="min-w-0 gap-4">
      <CardHeader><CardTitle>外部代理地址</CardTitle></CardHeader>
      <CardContent class="flex flex-col gap-2">
        <template v-if="endpoints.length">
          <Button v-for="endpoint in endpoints" :key="endpoint" variant="outline" class="h-auto min-h-9 min-w-0 justify-between" :aria-label="'复制代理地址 ' + endpoint" @click="copy(endpoint)">
            <span class="whitespace-normal text-left font-mono wrap-anywhere">{{ endpoint }}</span><CopyIcon data-icon="inline-end" />
          </Button>
        </template>
        <p v-else class="text-sm text-muted-foreground">{{ status ? '未配置外部访问地址' : '外部地址不可用' }}</p>
      </CardContent>
      <CardFooter class="mt-auto">
        <Button :variant="endpoints.length ? 'ghost' : 'outline'" size="sm" as-child><RouterLink to="/settings">{{ endpoints.length ? '网络设置' : '配置网络' }}<ArrowRightIcon data-icon="inline-end" /></RouterLink></Button>
      </CardFooter>
    </Card>
  </div>
</template>

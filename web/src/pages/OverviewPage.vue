<script setup lang="ts">
import { ActivityIcon, ArrowRightIcon, CopyIcon, GitBranchIcon, NetworkIcon, PowerIcon, ServerIcon } from '@lucide/vue'
import { computed } from 'vue'
import { toast } from 'vue-sonner'

import PageHeader from '@/components/PageHeader.vue'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { useConsole } from '@/composables/use-console'
import { api } from '@/lib/api'

const { status, nodes, entryGroups, settings, selectedNode, loading, busy, run } = useConsole()
const metrics = computed(() => [
  {
    label: '内核状态', icon: ActivityIcon, path: '/core',
    value: status.value ? (status.value.core.running ? '运行中' : '已停止') : '未知',
    detail: status.value?.core.pid ? `PID ${status.value.core.pid}` : undefined,
  },
  { label: '代理节点', icon: ServerIcon, path: '/nodes', value: status.value ? nodes.value.length : '—' },
  { label: '入口组', icon: GitBranchIcon, path: '/chains', value: status.value ? entryGroups.value.length : '—' },
  { label: 'Mixed 端口', icon: NetworkIcon, path: '/settings', value: status.value ? settings.value.mixedPort : '—', detail: 'HTTP + SOCKS5' },
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

  <div class="grid grid-cols-2 gap-3 @3xl:grid-cols-4 sm:gap-4">
    <RouterLink v-for="metric in metrics" :key="metric.label" :to="metric.path" class="group min-w-0 rounded-lg outline-none focus-visible:ring-2 focus-visible:ring-ring">
      <Card class="h-full min-w-0 gap-0 py-4 transition-colors group-hover:border-primary/40">
        <CardContent class="px-4 sm:px-5">
          <dl class="flex flex-col gap-1.5">
            <dt class="flex items-center gap-2 text-xs text-muted-foreground sm:text-sm">
              <component :is="metric.icon" class="size-4 shrink-0 text-primary" aria-hidden="true" />{{ metric.label }}
            </dt>
            <dd v-if="loading && !status"><Skeleton class="my-1 h-5 w-16" /><span class="sr-only">正在加载</span></dd>
            <dd v-else class="text-xl font-semibold tracking-tight tabular-nums">{{ metric.value }}</dd>
            <dd v-if="metric.detail" class="text-xs text-muted-foreground">{{ metric.detail }}</dd>
          </dl>
        </CardContent>
      </Card>
    </RouterLink>
  </div>

  <div class="mt-4 grid gap-4 @3xl:grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)] sm:mt-6">
    <Card class="min-w-0 gap-4">
      <CardHeader><CardTitle>默认出口</CardTitle></CardHeader>
      <CardContent class="flex flex-col gap-4">
        <div v-if="selectedNode" class="flex flex-col gap-1.5 rounded-md bg-muted/50 px-3 py-2.5">
          <div class="flex flex-wrap items-center gap-2">
            <strong class="wrap-anywhere">{{ selectedNode.name }}</strong>
            <Badge variant="outline">{{ selectedNode.type.toUpperCase() }}</Badge>
          </div>
          <span class="font-mono text-xs wrap-anywhere text-muted-foreground">{{ selectedNode.server }}:{{ selectedNode.port }}</span>
        </div>
        <p v-else class="text-sm text-muted-foreground">{{ status ? '尚未选择出口节点' : '出口状态不可用' }}</p>
        <div v-if="nodes.length" role="group" aria-labelledby="quick-exit-heading" class="flex flex-col items-start gap-2">
          <span id="quick-exit-heading" class="text-xs text-muted-foreground">切换节点</span>
          <div class="flex max-w-full flex-wrap gap-2">
          <Button v-for="node in nodes.slice(0, 6)" :key="node.id" size="sm" class="min-w-0 max-w-full" :variant="node.id === selectedNode?.id ? 'default' : 'outline'" :aria-pressed="node.id === selectedNode?.id" :disabled="busy" @click="run(() => api.selectNode(node.id), '默认出口已切换')"><span class="truncate">{{ node.name }}</span></Button>
          </div>
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

<script setup lang="ts">
import { ActivityIcon, ArrowRightIcon, CopyIcon, NetworkIcon, PowerIcon, ServerIcon } from '@lucide/vue'
import { computed } from 'vue'
import { toast } from 'vue-sonner'

import PageHeader from '@/components/PageHeader.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useConsole } from '@/composables/use-console'
import { api } from '@/lib/api'

const { status, nodes, entryGroups, settings, selectedNode, busy, run } = useConsole()
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
  <PageHeader title="运行概览" description="查看代理入口、默认出口与 mihomo 运行状态">
    <Button :variant="status?.core.running ? 'destructive' : 'default'" :disabled="busy" @click="run(() => api.coreAction(status?.core.running ? 'stop' : 'start'), status?.core.running ? '内核已停止' : '内核已启动')">
      <PowerIcon data-icon="inline-start" />{{ status?.core.running ? '停止内核' : '启动内核' }}
    </Button>
  </PageHeader>

  <section class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
    <Card>
      <CardHeader class="flex-row items-center justify-between gap-3 pb-2"><CardDescription>内核状态</CardDescription><ActivityIcon class="text-muted-foreground" /></CardHeader>
      <CardContent><CardTitle class="text-xl">{{ status?.core.running ? '运行中' : '已停止' }}</CardTitle><p class="mt-1 text-xs text-muted-foreground">{{ status?.core.pid ? `PID ${status.core.pid}` : 'mihomo core' }}</p></CardContent>
    </Card>
    <Card>
      <CardHeader class="flex-row items-center justify-between gap-3 pb-2"><CardDescription>代理节点</CardDescription><ServerIcon class="text-muted-foreground" /></CardHeader>
      <CardContent><CardTitle class="text-xl">{{ nodes.length }}</CardTitle><p class="mt-1 text-xs text-muted-foreground">已导入节点</p></CardContent>
    </Card>
    <Card>
      <CardHeader class="flex-row items-center justify-between gap-3 pb-2"><CardDescription>入口组</CardDescription><NetworkIcon class="text-muted-foreground" /></CardHeader>
      <CardContent><CardTitle class="text-xl">{{ entryGroups.length }}</CardTitle><p class="mt-1 text-xs text-muted-foreground">手动选择与故障转移</p></CardContent>
    </Card>
    <Card>
      <CardHeader class="pb-2"><CardDescription>Mixed 代理端口</CardDescription></CardHeader>
      <CardContent><CardTitle class="font-mono text-xl">{{ settings.mixedPort }}</CardTitle><p class="mt-1 text-xs text-muted-foreground">HTTP + SOCKS5</p></CardContent>
    </Card>
  </section>

  <section class="mt-6 grid gap-6 lg:grid-cols-[minmax(0,1.5fr)_minmax(280px,1fr)]">
    <Card>
      <CardHeader>
        <div class="flex items-center justify-between gap-3"><div><CardTitle>默认出口</CardTitle><CardDescription class="mt-1">DOCKER_CLASH 策略组当前使用的节点</CardDescription></div><Badge :variant="status?.core.running ? 'default' : 'secondary'">{{ status?.core.running ? '在线' : '离线' }}</Badge></div>
      </CardHeader>
      <CardContent class="flex flex-col gap-4">
        <div v-if="selectedNode" class="flex items-center gap-4 rounded-md border p-4">
          <span class="flex size-10 shrink-0 items-center justify-center rounded-md bg-secondary font-mono text-xs font-semibold">{{ selectedNode.type.slice(0, 3).toUpperCase() }}</span>
          <div class="min-w-0"><strong class="block truncate">{{ selectedNode.name }}</strong><span class="block truncate font-mono text-xs text-muted-foreground">{{ selectedNode.server }}:{{ selectedNode.port }}</span></div>
        </div>
        <p v-else class="rounded-md border border-dashed p-8 text-center text-sm text-muted-foreground">尚未选择出口节点</p>
        <div class="flex flex-wrap gap-2">
          <Button v-for="node in nodes.slice(0, 6)" :key="node.id" size="sm" :variant="node.name === settings.selectedNode ? 'default' : 'outline'" :disabled="busy" @click="run(() => api.selectNode(node.id), '默认出口已切换')">{{ node.name }}</Button>
          <Button v-if="nodes.length > 6" variant="ghost" size="sm" as-child><RouterLink to="/nodes">全部节点<ArrowRightIcon data-icon="inline-end" /></RouterLink></Button>
        </div>
      </CardContent>
    </Card>

    <Card>
      <CardHeader><CardTitle>外部代理地址</CardTitle><CardDescription>供其他设备连接的宿主机入口</CardDescription></CardHeader>
      <CardContent class="flex flex-1 flex-col gap-3">
        <template v-if="endpoints.length">
          <Button v-for="endpoint in endpoints" :key="endpoint" variant="outline" class="justify-between font-mono" @click="copy(endpoint)"><span class="truncate">{{ endpoint }}</span><CopyIcon /></Button>
        </template>
        <p v-else class="rounded-md border border-dashed p-6 text-center text-sm text-muted-foreground">请先配置外部访问地址</p>
        <Button variant="ghost" class="mt-auto" as-child><RouterLink to="/settings">配置网络<ArrowRightIcon data-icon="inline-end" /></RouterLink></Button>
      </CardContent>
    </Card>
  </section>
</template>

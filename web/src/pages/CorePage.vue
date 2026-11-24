<script setup lang="ts">
import { PowerIcon, RefreshCwIcon } from '@lucide/vue'
import { onMounted, ref } from 'vue'
import { toast } from 'vue-sonner'

import PageHeader from '@/components/PageHeader.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { useConsole } from '@/composables/use-console'
import { api } from '@/lib/api'

const { status, busy, run } = useConsole()
const log = ref('')
const logLoading = ref(false)

async function loadLog() {
  logLoading.value = true
  try { log.value = await api.coreLog() }
  catch (error) { toast.error((error as Error).message) }
  finally { logLoading.value = false }
}

onMounted(loadLog)
</script>

<template>
  <PageHeader title="内核管理">
    <Button variant="outline" size="sm" :disabled="busy" @click="run(() => api.coreAction('restart'), '内核已重启')">
      <RefreshCwIcon data-icon="inline-start" />重启
    </Button>
    <Button size="sm" :variant="status?.core.running ? 'destructive' : 'default'" :disabled="busy" @click="run(() => api.coreAction(status?.core.running ? 'stop' : 'start'), status?.core.running ? '内核已停止' : '内核已启动')">
      <PowerIcon data-icon="inline-start" />{{ status?.core.running ? '停止' : '启动' }}
    </Button>
  </PageHeader>

  <div class="flex min-w-0 flex-col gap-4">
    <Card class="min-w-0 gap-4 py-4">
      <CardHeader class="px-4 sm:px-5">
        <CardTitle class="flex flex-wrap items-center gap-2">
          mihomo
          <Badge :variant="status?.core.running ? 'default' : 'secondary'">{{ status ? (status.core.running ? '运行中' : '已停止') : '未知' }}</Badge>
        </CardTitle>
      </CardHeader>
      <CardContent class="px-4 sm:px-5">
        <dl class="grid grid-cols-[auto_minmax(0,1fr)] gap-x-6 gap-y-3 text-sm @xl:grid-cols-[minmax(5rem,auto)_minmax(9rem,auto)_minmax(0,1fr)] @xl:gap-x-8">
          <div class="flex flex-col gap-1">
            <dt class="text-muted-foreground">进程 ID</dt>
            <dd class="font-mono font-semibold">{{ status?.core.pid || '-' }}</dd>
          </div>
          <div class="flex min-w-0 flex-col gap-1">
            <dt class="text-muted-foreground">Controller</dt>
            <dd class="wrap-anywhere font-mono font-semibold">127.0.0.1:19090</dd>
          </div>
          <div class="col-span-2 flex min-w-0 flex-col gap-1 @xl:col-span-1">
            <dt class="text-muted-foreground">最近错误</dt>
            <dd class="wrap-anywhere font-semibold">{{ status ? status.core.error || '无' : '—' }}</dd>
          </div>
        </dl>
      </CardContent>
    </Card>

    <Card class="min-w-0 gap-0 pb-0 pt-4">
      <CardHeader class="items-center px-4 pb-4 sm:px-5">
        <CardTitle id="core-log-heading">运行日志</CardTitle>
        <CardDescription id="core-log-scope">最近 64 KiB</CardDescription>
        <CardAction class="self-center">
          <Button variant="outline" size="sm" :disabled="logLoading" @click="loadLog">
            <RefreshCwIcon data-icon="inline-start" />刷新
          </Button>
        </CardAction>
      </CardHeader>
      <Separator />
      <CardContent class="px-0">
        <div
          role="region"
          aria-labelledby="core-log-heading"
          aria-describedby="core-log-scope"
          :tabindex="!logLoading && log ? 0 : undefined"
          class="h-[clamp(18rem,52vh,40rem)] min-w-0 overflow-auto overscroll-contain rounded-b-xl bg-muted/35 focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-ring"
        >
          <p v-if="logLoading || !log" role="status" class="p-4 text-sm text-muted-foreground sm:px-5">{{ logLoading ? '加载中...' : '暂无日志' }}</p>
          <pre v-else class="min-w-0 p-4 font-mono text-xs leading-5 wrap-anywhere whitespace-pre-wrap sm:px-5">{{ log }}</pre>
        </div>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { PowerIcon, RefreshCwIcon } from '@lucide/vue'
import { onMounted, ref } from 'vue'
import { toast } from 'vue-sonner'

import PageHeader from '@/components/PageHeader.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
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
  <PageHeader title="内核管理" />

  <Card class="min-w-0">
    <CardHeader>
      <div class="flex flex-wrap items-center justify-between gap-4">
        <div class="flex flex-wrap items-center gap-2">
          <CardTitle>mihomo</CardTitle>
          <Badge :variant="status?.core.running ? 'default' : 'secondary'">{{ status ? (status.core.running ? '运行中' : '已停止') : '未知' }}</Badge>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <Button variant="outline" :disabled="busy" @click="run(() => api.coreAction('restart'), '内核已重启')"><RefreshCwIcon data-icon="inline-start" />重启</Button>
          <Button :variant="status?.core.running ? 'destructive' : 'default'" :disabled="busy" @click="run(() => api.coreAction(status?.core.running ? 'stop' : 'start'), status?.core.running ? '内核已停止' : '内核已启动')"><PowerIcon data-icon="inline-start" />{{ status?.core.running ? '停止' : '启动' }}</Button>
        </div>
      </div>
    </CardHeader>
    <CardContent>
      <dl class="grid gap-4 text-sm sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_minmax(0,2fr)]">
        <div class="flex flex-col gap-1"><dt class="text-muted-foreground">进程 ID</dt><dd class="font-mono font-semibold">{{ status?.core.pid || '-' }}</dd></div>
        <div class="flex flex-col gap-1"><dt class="text-muted-foreground">Controller</dt><dd class="font-mono font-semibold">127.0.0.1:19090</dd></div>
        <div class="flex min-w-0 flex-col gap-1"><dt class="text-muted-foreground">最近错误</dt><dd class="wrap-anywhere font-semibold">{{ status ? status.core.error || '无' : '—' }}</dd></div>
      </dl>
    </CardContent>
  </Card>

  <section class="mt-6 flex min-w-0 flex-col gap-3" aria-labelledby="core-log-heading">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="flex flex-wrap items-baseline gap-x-2 gap-y-1">
        <h2 id="core-log-heading" class="font-semibold">运行日志</h2>
        <span class="text-sm text-muted-foreground">最近 64 KiB</span>
      </div>
      <Button variant="outline" size="sm" :disabled="logLoading" @click="loadLog"><RefreshCwIcon data-icon="inline-start" />刷新</Button>
    </div>
    <pre class="min-h-72 max-h-[60vh] overflow-auto rounded-md border bg-muted/40 p-4 font-mono text-xs leading-5">{{ logLoading ? '加载中...' : log || '暂无日志' }}</pre>
  </section>
</template>

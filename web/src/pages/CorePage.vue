<script setup lang="ts">
import { ActivityIcon, PowerIcon, RefreshCwIcon, ScrollTextIcon } from '@lucide/vue'
import { onMounted, ref } from 'vue'
import { toast } from 'vue-sonner'

import PageHeader from '@/components/PageHeader.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
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
  <PageHeader title="内核管理" description="控制 mihomo 子进程并查看最近运行日志">
    <Button variant="outline" :disabled="busy" @click="run(() => api.coreAction('restart'), '内核已重启')"><RefreshCwIcon data-icon="inline-start" />重启</Button>
    <Button :variant="status?.core.running ? 'destructive' : 'default'" :disabled="busy" @click="run(() => api.coreAction(status?.core.running ? 'stop' : 'start'), status?.core.running ? '内核已停止' : '内核已启动')"><PowerIcon data-icon="inline-start" />{{ status?.core.running ? '停止' : '启动' }}</Button>
  </PageHeader>

  <Card>
    <CardHeader><div class="flex items-start justify-between gap-4"><div><CardTitle class="flex items-center gap-2"><ActivityIcon />mihomo core</CardTitle><CardDescription class="mt-1">由 Docker Clash 子进程管理器托管</CardDescription></div><Badge :variant="status?.core.running ? 'default' : 'secondary'">{{ status?.core.running ? '运行中' : '已停止' }}</Badge></div></CardHeader>
    <CardContent class="grid gap-4 text-sm sm:grid-cols-3"><div><span class="text-muted-foreground">进程 ID</span><strong class="mt-1 block font-mono">{{ status?.core.pid || '-' }}</strong></div><div><span class="text-muted-foreground">Controller</span><strong class="mt-1 block font-mono">127.0.0.1:19090</strong></div><div><span class="text-muted-foreground">最近错误</span><strong class="mt-1 block break-words">{{ status?.core.error || '无' }}</strong></div></CardContent>
  </Card>

  <section class="mt-6">
    <div class="mb-3 flex items-center justify-between gap-3"><div><h2 class="flex items-center gap-2 font-semibold"><ScrollTextIcon />运行日志</h2><p class="text-sm text-muted-foreground">最近 64 KiB 内核输出</p></div><Button variant="outline" size="sm" :disabled="logLoading" @click="loadLog"><RefreshCwIcon data-icon="inline-start" />刷新</Button></div>
    <pre class="min-h-72 max-h-[60vh] overflow-auto rounded-md border bg-muted/40 p-4 font-mono text-xs leading-5">{{ logLoading ? '加载中...' : log || '暂无日志' }}</pre>
  </section>
</template>

<script setup lang="ts">
import { ActivityIcon, ImportIcon, LocateFixedIcon, Trash2Icon } from '@lucide/vue'
import { computed, nextTick, ref } from 'vue'
import { toast } from 'vue-sonner'

import PageHeader from '@/components/PageHeader.vue'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useConsole } from '@/composables/use-console'
import { api } from '@/lib/api'
import type { Node } from '@/types'

const { nodes, settings, selectedNode, busy, run } = useConsole()
const importOpen = ref(false)
const clearOpen = ref(false)
const uri = ref('')
const delays = ref<Record<string, number>>({})
const testing = ref(new Set<string>())
const query = ref('')
const visibleNodes = computed(() => {
  const term = query.value.trim().toLowerCase()
  if (!term) return nodes.value
  return nodes.value.filter(node =>
    [node.name, node.type, `${node.server}:${node.port}`, node.dialerProxy || 'DIRECT']
      .some(value => value.toLowerCase().includes(term)),
  )
})

async function importNodes() {
  const success = await run(() => api.importNodes(uri.value.trim()), '节点或订阅已导入')
  if (success) { uri.value = ''; importOpen.value = false }
}

async function testNode(node: Node) {
  testing.value = new Set(testing.value).add(node.id)
  try {
    delays.value = { ...delays.value, [node.id]: (await api.delay(node.id)).delay }
  }
  catch (error) { toast.error((error as Error).message) }
  finally { const next = new Set(testing.value); next.delete(node.id); testing.value = next }
}

async function testAll() {
  testing.value = new Set(nodes.value.map(node => node.id))
  const results = await Promise.allSettled(nodes.value.map(async node => ({ id: node.id, delay: (await api.delay(node.id)).delay })))
  const next = { ...delays.value }
  let failed = 0
  for (const result of results) {
    if (result.status === 'fulfilled') next[result.value.id] = result.value.delay
    else failed++
  }
  delays.value = next; testing.value = new Set()
  failed ? toast.error(`延迟测试完成，${failed} 个节点失败`) : toast.success('全部节点测试完成')
}

async function remove(node: Node) {
  if (window.confirm(`确定删除节点“${node.name}”吗？`)) await run(() => api.deleteNode(node.id), '节点已删除')
}

async function clear() {
  await run(api.clearNodes, '全部节点已清空')
}

async function scrollToDefaultNode() {
  const current = selectedNode.value
  if (!current) return
  if (!visibleNodes.value.includes(current)) {
    query.value = ''
    await nextTick()
  }
  document.getElementById(`node-${current.id}`)?.scrollIntoView({ behavior: 'smooth', block: 'center' })
}
</script>

<template>
  <PageHeader title="代理节点">
    <Button @click="importOpen = true"><ImportIcon data-icon="inline-start" />导入节点</Button>
  </PageHeader>

  <section class="flex min-w-0 flex-col gap-3" aria-label="节点列表">
    <div v-if="nodes.length" class="flex flex-wrap items-center justify-between gap-2.5">
      <Input v-model="query" type="search" class="w-full @min-[40rem]:max-w-xs" placeholder="搜索节点" aria-label="搜索节点" />
      <div class="flex flex-wrap gap-2">
        <Button size="sm" variant="outline" :disabled="!selectedNode" @click="scrollToDefaultNode"><LocateFixedIcon data-icon="inline-start" />跳转默认节点</Button>
        <Button size="sm" variant="outline" :disabled="testing.size > 0" @click="testAll"><ActivityIcon data-icon="inline-start" />一键测速</Button>
        <Button size="sm" variant="outline" :disabled="busy" @click="clearOpen = true"><Trash2Icon data-icon="inline-start" />清空全部</Button>
      </div>
    </div>

    <div v-if="visibleNodes.length" class="overflow-hidden rounded-lg border bg-card">
      <Table class="table-fixed @min-[46rem]:table-auto [&_:is(th,td):first-child]:pl-3 [&_:is(th,td):last-child]:pr-3 @min-[46rem]:[&_:is(th,td):first-child]:pl-4 @min-[46rem]:[&_:is(th,td):last-child]:pr-4">
        <caption class="sr-only">代理节点列表</caption>
        <TableHeader class="bg-muted/40">
          <TableRow>
            <TableHead scope="col">节点</TableHead>
            <TableHead scope="col" class="hidden @min-[46rem]:table-cell">协议</TableHead>
            <TableHead scope="col" class="hidden @min-[46rem]:table-cell">服务器</TableHead>
            <TableHead scope="col" class="hidden @min-[46rem]:table-cell">延迟</TableHead>
            <TableHead scope="col" class="hidden @min-[46rem]:table-cell">链路入口</TableHead>
            <TableHead scope="col" class="w-28 text-right @min-[46rem]:w-auto">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="node in visibleNodes" :id="`node-${node.id}`" :key="node.id">
            <TableCell class="align-top @min-[46rem]:align-middle">
              <div class="flex max-w-64 flex-wrap items-center gap-x-2 gap-y-1 whitespace-normal">
                <strong class="font-medium wrap-anywhere">{{ node.name }}</strong>
                <Badge v-if="node.name === settings.selectedNode">默认出口</Badge>
              </div>
              <dl class="mt-1 flex flex-col items-start gap-1 text-xs whitespace-normal @min-[46rem]:hidden">
                <div><dt class="sr-only">协议</dt><dd><Badge variant="secondary">{{ node.type }}</Badge></dd></div>
                <div class="min-w-0 max-w-full"><dt class="sr-only">服务器</dt><dd class="font-mono wrap-anywhere">{{ node.server }}:{{ node.port }}</dd></div>
                <div class="flex min-w-0 max-w-full flex-wrap gap-x-1 text-muted-foreground">
                  <dt>链路入口：</dt><dd class="wrap-anywhere">{{ node.dialerProxy || 'DIRECT' }}</dd>
                </div>
              </dl>
            </TableCell>
            <TableCell class="hidden @min-[46rem]:table-cell"><Badge variant="secondary">{{ node.type }}</Badge></TableCell>
            <TableCell class="hidden max-w-48 font-mono text-xs whitespace-normal wrap-anywhere @min-[46rem]:table-cell">{{ node.server }}:{{ node.port }}</TableCell>
            <TableCell class="hidden @min-[46rem]:table-cell"><Button variant="ghost" size="sm" class="tabular-nums" :disabled="testing.has(node.id)" @click="testNode(node)">{{ testing.has(node.id) ? '测试中' : delays[node.id] === undefined ? '测试' : `${delays[node.id]} ms` }}</Button></TableCell>
            <TableCell class="hidden max-w-48 whitespace-normal wrap-anywhere @min-[46rem]:table-cell">{{ node.dialerProxy || 'DIRECT' }}</TableCell>
            <TableCell>
              <div class="flex flex-wrap justify-end gap-1 @min-[46rem]:flex-nowrap">
                <Button class="tabular-nums @min-[46rem]:hidden" variant="ghost" size="sm" :disabled="testing.has(node.id)" @click="testNode(node)">{{ testing.has(node.id) ? '测试中' : delays[node.id] === undefined ? '测速' : `${delays[node.id]} ms` }}</Button>
                <Button v-if="node.name !== settings.selectedNode" variant="ghost" size="sm" :disabled="busy" @click="run(() => api.selectNode(node.id), '默认出口已切换')">设为默认</Button>
                <Button variant="ghost" size="icon-sm" title="删除节点" :aria-label="`删除节点 ${node.name}`" :disabled="busy" @click="remove(node)"><Trash2Icon /></Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>
    <p v-else class="rounded-lg border border-dashed px-4 py-10 text-center text-sm text-muted-foreground">
      {{ nodes.length ? '未找到匹配的节点。' : '暂无节点，请导入节点或订阅。' }}
    </p>
    <p v-if="nodes.length" class="text-xs text-muted-foreground" aria-live="polite">
      {{ visibleNodes.length === nodes.length ? `共 ${nodes.length} 个节点` : `显示 ${visibleNodes.length} / ${nodes.length} 个节点` }}
    </p>
  </section>

  <Dialog v-model:open="importOpen">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>导入节点或订阅</DialogTitle>
        <DialogDescription>支持标准代理 URI；使用订阅地址可一次导入多个节点。</DialogDescription>
      </DialogHeader>
      <FieldGroup>
        <Field class="min-w-0">
          <FieldLabel for="node-uri">订阅地址或节点 URI</FieldLabel>
          <Textarea id="node-uri" v-model="uri" rows="5" class="max-w-full" />
        </Field>
      </FieldGroup>
      <DialogFooter><Button variant="outline" @click="importOpen = false">取消</Button><Button :disabled="busy || !uri.trim()" @click="importNodes">导入</Button></DialogFooter>
    </DialogContent>
  </Dialog>

  <AlertDialog v-model:open="clearOpen">
    <AlertDialogContent>
      <AlertDialogHeader><AlertDialogTitle>清空全部节点</AlertDialogTitle><AlertDialogDescription>确定清空全部 {{ nodes.length }} 个节点吗？此操作不可撤销。</AlertDialogDescription></AlertDialogHeader>
      <AlertDialogFooter><AlertDialogCancel>取消</AlertDialogCancel><AlertDialogAction variant="destructive" :disabled="busy" @click="clear">清空</AlertDialogAction></AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>

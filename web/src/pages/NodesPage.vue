<script setup lang="ts">
import { ActivityIcon, ImportIcon, LocateFixedIcon, Trash2Icon } from '@lucide/vue'
import { ref } from 'vue'
import { toast } from 'vue-sonner'

import PageHeader from '@/components/PageHeader.vue'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Field, FieldDescription, FieldLabel } from '@/components/ui/field'
import { Textarea } from '@/components/ui/textarea'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useConsole } from '@/composables/use-console'
import { api } from '@/lib/api'
import type { Node } from '@/types'

const { nodes, settings, busy, run } = useConsole()
const importOpen = ref(false)
const clearOpen = ref(false)
const uri = ref('')
const delays = ref<Record<string, number>>({})
const testing = ref(new Set<string>())

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
</script>

<template>
  <PageHeader title="代理节点" description="导入、测速并选择默认代理出口">
    <Button variant="outline" :disabled="!nodes.length || testing.size > 0" @click="testAll"><ActivityIcon data-icon="inline-start" />一键测速</Button>
    <Button variant="destructive" :disabled="!nodes.length || busy" @click="clearOpen = true"><Trash2Icon data-icon="inline-start" />清空</Button>
    <Button @click="importOpen = true"><ImportIcon data-icon="inline-start" />导入节点</Button>
  </PageHeader>

  <div class="overflow-x-auto rounded-md border">
    <Table class="min-w-[900px]">
      <TableHeader><TableRow><TableHead>节点</TableHead><TableHead>协议</TableHead><TableHead>服务器</TableHead><TableHead>延迟</TableHead><TableHead>链路入口</TableHead><TableHead class="text-right">操作</TableHead></TableRow></TableHeader>
      <TableBody>
        <TableRow v-for="node in nodes" :key="node.id">
          <TableCell><div class="flex items-center gap-2"><span :class="['size-2 shrink-0 rounded-full', node.name === settings.selectedNode ? 'bg-emerald-500' : 'bg-muted-foreground/40']" /><strong>{{ node.name }}</strong></div></TableCell>
          <TableCell><Badge variant="secondary">{{ node.type }}</Badge></TableCell>
          <TableCell class="font-mono text-xs">{{ node.server }}:{{ node.port }}</TableCell>
          <TableCell><Button variant="ghost" size="sm" :disabled="testing.has(node.id)" @click="testNode(node)">{{ testing.has(node.id) ? '测试中' : delays[node.id] === undefined ? '测试' : `${delays[node.id]} ms` }}</Button></TableCell>
          <TableCell>{{ node.dialerProxy || 'DIRECT' }}</TableCell>
          <TableCell><div class="flex justify-end gap-1"><Button variant="ghost" size="sm" :disabled="busy || node.name === settings.selectedNode" @click="run(() => api.selectNode(node.id), '默认出口已切换')"><LocateFixedIcon data-icon="inline-start" />设为默认</Button><Button variant="ghost" size="icon" title="删除节点" :disabled="busy" @click="remove(node)"><Trash2Icon /></Button></div></TableCell>
        </TableRow>
        <TableRow v-if="!nodes.length"><TableCell colspan="6" class="h-32 text-center text-muted-foreground">暂无节点，导入节点或订阅后开始使用</TableCell></TableRow>
      </TableBody>
    </Table>
  </div>

  <Dialog v-model:open="importOpen">
    <DialogContent>
      <DialogHeader><DialogTitle>导入节点或订阅</DialogTitle><DialogDescription>支持后端可解析的订阅地址和标准代理 URI</DialogDescription></DialogHeader>
      <Field class="min-w-0"><FieldLabel for="node-uri">订阅或节点 URI</FieldLabel><Textarea id="node-uri" v-model="uri" rows="5" class="max-w-full" placeholder="粘贴订阅地址或代理 URI" /><FieldDescription>多个节点可通过订阅内容一次导入。</FieldDescription></Field>
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

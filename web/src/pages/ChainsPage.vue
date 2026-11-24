<script setup lang="ts">
import { ChevronDownIcon, PencilIcon, PlusIcon, SaveIcon, Trash2Icon } from '@lucide/vue'
import { ref } from 'vue'
import { toast } from 'vue-sonner'

import PageHeader from '@/components/PageHeader.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Field, FieldDescription, FieldGroup, FieldLabel, FieldLegend, FieldSet } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useConsole } from '@/composables/use-console'
import { api } from '@/lib/api'
import type { EntryGroup, EntryGroupInput, Node } from '@/types'

const { nodes, entryGroups, nodeById, busy, run } = useConsole()
const dialogOpen = ref(false)
const editingId = ref<string>()
const draft = ref<EntryGroupInput>({ name: '', type: 'select', nodeIds: [], testUrl: '', interval: 60 })
const dialerOpenId = ref<string>()

function createGroup() {
  editingId.value = undefined
  draft.value = { name: '', type: 'select', nodeIds: [], testUrl: '', interval: 60 }
  dialogOpen.value = true
}

function editGroup(group: EntryGroup) {
  editingId.value = group.id
  draft.value = { name: group.name, type: group.type, nodeIds: [...group.nodeIds], testUrl: group.testUrl || '', interval: group.interval || 60 }
  dialogOpen.value = true
}

function toggleMember(id: string, checked: boolean | 'indeterminate') {
  draft.value.nodeIds = checked
    ? [...new Set([...draft.value.nodeIds, id])]
    : draft.value.nodeIds.filter(nodeId => nodeId !== id)
}

async function saveGroup() {
  const name = draft.value.name.trim()
  if (!name || !draft.value.nodeIds.length) return toast.error('请填写名称并选择至少一个成员节点')
  const payload: EntryGroupInput = {
    name, type: draft.value.type, nodeIds: [...draft.value.nodeIds],
    ...(draft.value.type === 'fallback' ? { testUrl: draft.value.testUrl?.trim() || undefined, interval: Number(draft.value.interval) || 60 } : {}),
  }
  const current = entryGroups.value.find(group => group.id === editingId.value)
  const success = await run(
    () => current ? api.updateEntryGroup({ ...current, ...payload }) : api.createEntryGroup(payload),
    current ? '入口组已更新' : '入口组已创建',
  )
  if (success) dialogOpen.value = false
}

async function removeGroup(group: EntryGroup) {
  if (window.confirm(`确定删除入口组“${group.name}”吗？`)) await run(() => api.deleteEntryGroup(group.id), '入口组已删除')
}

async function setDialer(node: Node, value: string) {
  dialerOpenId.value = undefined
  await run(() => api.updateNode({ ...node, dialerProxy: value === '__direct__' ? undefined : value }), '节点链路已更新')
}
</script>

<template>
  <PageHeader title="入口与链路">
    <Button :disabled="busy || !nodes.length" @click="createGroup"><PlusIcon data-icon="inline-start" />创建入口组</Button>
  </PageHeader>

  <div class="flex min-w-0 flex-col gap-6">
    <section class="flex min-w-0 flex-col gap-3" aria-labelledby="entry-groups-heading">
      <h2 id="entry-groups-heading" class="text-sm font-semibold">入口组</h2>
      <div v-if="entryGroups.length" class="grid gap-3 @min-[39rem]:grid-cols-2 @min-[58rem]:grid-cols-3">
        <Card v-for="group in entryGroups" :key="group.id" class="min-w-0 gap-3 py-4">
          <CardHeader class="grid-cols-[minmax(0,1fr)_auto] grid-rows-1 gap-x-2 px-4">
            <div class="flex min-w-0 flex-col items-start gap-2">
              <CardTitle class="leading-snug wrap-anywhere">{{ group.name }}</CardTitle>
              <div class="flex flex-wrap items-center gap-x-2 gap-y-1">
                <Badge variant="secondary">{{ group.type === 'select' ? '手动选择' : '故障转移' }}</Badge>
                <CardDescription>{{ group.nodeIds.length }} 个成员节点</CardDescription>
              </div>
            </div>
            <div class="flex items-center gap-0.5">
              <Button variant="ghost" size="icon-sm" title="编辑入口组" :aria-label="`编辑入口组 ${group.name}`" @click="editGroup(group)"><PencilIcon /></Button>
              <Button variant="ghost" size="icon-sm" title="删除入口组" :aria-label="`删除入口组 ${group.name}`" @click="removeGroup(group)"><Trash2Icon /></Button>
            </div>
          </CardHeader>
          <CardContent class="mt-auto px-4">
            <Field v-if="group.type === 'select'">
              <FieldLabel :for="`current-member-${group.id}`">当前成员</FieldLabel>
              <Select :model-value="group.selectedNodeId" :disabled="busy" @update:model-value="value => value && run(() => api.selectEntryGroupNode(group.id, String(value)), '入口组成员已切换')">
                <SelectTrigger :id="`current-member-${group.id}`" size="sm" class="w-full min-w-0"><SelectValue class="min-w-0" placeholder="选择成员" /></SelectTrigger>
                <SelectContent><SelectGroup><SelectItem v-for="id in group.nodeIds" :key="id" :value="id">{{ nodeById.get(id)?.name || '节点已移除' }}</SelectItem></SelectGroup></SelectContent>
              </Select>
            </Field>
            <div v-else class="flex flex-col gap-1.5 text-sm">
              <span class="font-medium">优先顺序</span>
              <p class="wrap-anywhere">{{ group.nodeIds.map(id => nodeById.get(id)?.name || '未知节点').join(' → ') }}</p>
              <p class="text-xs text-muted-foreground">每 {{ group.interval || 60 }} 秒检测一次</p>
            </div>
          </CardContent>
        </Card>
      </div>
      <p v-else class="rounded-lg border border-dashed px-4 py-8 text-center text-sm text-muted-foreground">暂无入口组</p>
    </section>

    <section class="flex min-w-0 flex-col gap-3" aria-labelledby="node-chains-heading">
      <h2 id="node-chains-heading" class="text-sm font-semibold">节点拨号入口</h2>
      <div v-if="nodes.length" class="overflow-hidden rounded-lg border bg-card">
        <Table aria-labelledby="node-chains-heading" class="table-fixed">
          <TableHeader class="bg-muted/40">
            <TableRow>
              <TableHead scope="col" class="pl-3 @min-[46rem]:pl-4">节点</TableHead>
              <TableHead scope="col" class="w-2/5 pr-3 @min-[46rem]:w-1/3 @min-[46rem]:pr-4">拨号入口</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="node in nodes" :key="node.id">
              <TableCell class="pl-3 whitespace-normal @min-[46rem]:pl-4">
                <strong class="block font-medium wrap-anywhere">{{ node.name }}</strong>
                <span class="font-mono text-xs text-muted-foreground wrap-anywhere">{{ node.server }}:{{ node.port }}</span>
              </TableCell>
              <TableCell class="pr-3 @min-[46rem]:pr-4">
                <Select
                  v-if="dialerOpenId === node.id"
                  :model-value="node.dialerProxy || '__direct__'"
                  :open="true"
                  :disabled="busy"
                  @update:open="value => { if (!value) dialerOpenId = undefined }"
                  @update:model-value="value => value && setDialer(node, String(value))"
                >
                  <SelectTrigger size="sm" class="w-full min-w-0" :aria-label="`${node.name} 的拨号入口`"><SelectValue class="min-w-0" /></SelectTrigger>
                  <SelectContent>
                    <SelectGroup><SelectLabel>直连</SelectLabel><SelectItem value="__direct__">DIRECT</SelectItem></SelectGroup>
                    <SelectGroup v-if="entryGroups.length"><SelectLabel>入口组</SelectLabel><SelectItem v-for="group in entryGroups" :key="group.id" :value="group.name">{{ group.name }}</SelectItem></SelectGroup>
                    <SelectGroup><SelectLabel>单节点</SelectLabel><SelectItem v-for="candidate in nodes.filter(item => item.id !== node.id)" :key="candidate.id" :value="candidate.name">{{ candidate.name }}</SelectItem></SelectGroup>
                  </SelectContent>
                </Select>
                <Button v-else variant="outline" size="sm" class="w-full min-w-0 justify-between" :disabled="busy" aria-haspopup="listbox" :aria-label="`${node.name} 的拨号入口：${node.dialerProxy || 'DIRECT'}`" @click="dialerOpenId = node.id">
                  <span class="truncate">{{ node.dialerProxy || 'DIRECT' }}</span>
                  <ChevronDownIcon data-icon="inline-end" />
                </Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
      <p v-else class="rounded-lg border border-dashed px-4 py-8 text-center text-sm text-muted-foreground">导入节点后可配置代理链路</p>
    </section>
  </div>

  <Dialog v-model:open="dialogOpen">
    <DialogContent class="sm:max-w-xl">
      <DialogHeader><DialogTitle>{{ editingId ? '编辑入口组' : '创建入口组' }}</DialogTitle><DialogDescription>成员节点可作为其他节点的统一拨号入口。</DialogDescription></DialogHeader>
      <FieldGroup>
        <Field><FieldLabel for="group-name">名称</FieldLabel><Input id="group-name" v-model="draft.name" placeholder="例如：香港入口" /></Field>
        <Field>
          <FieldLabel for="group-mode">模式</FieldLabel>
          <Select v-model="draft.type"><SelectTrigger id="group-mode" aria-describedby="group-mode-help"><SelectValue /></SelectTrigger><SelectContent><SelectGroup><SelectItem value="select">手动选择</SelectItem><SelectItem value="fallback">故障转移</SelectItem></SelectGroup></SelectContent></Select>
          <FieldDescription id="group-mode-help">故障转移按成员顺序自动选择可用节点。</FieldDescription>
        </Field>
        <FieldSet><FieldLegend>成员节点</FieldLegend><FieldGroup class="max-h-48 overflow-y-auto rounded-md border p-3">
          <Field v-for="node in nodes" :key="node.id" orientation="horizontal"><Checkbox :id="`member-${node.id}`" :model-value="draft.nodeIds.includes(node.id)" @update:model-value="checked => toggleMember(node.id, checked)" /><FieldLabel :for="`member-${node.id}`" class="font-normal">{{ node.name }} <span class="text-muted-foreground">({{ node.type }})</span></FieldLabel></Field>
        </FieldGroup></FieldSet>
        <template v-if="draft.type === 'fallback'">
          <Field><FieldLabel for="test-url">检测 URL</FieldLabel><Input id="test-url" v-model="draft.testUrl" placeholder="https://www.gstatic.com/generate_204" /></Field>
          <Field><FieldLabel for="interval">检测间隔（秒）</FieldLabel><Input id="interval" v-model.number="draft.interval" type="number" min="1" /></Field>
        </template>
      </FieldGroup>
      <DialogFooter><Button variant="outline" @click="dialogOpen = false">取消</Button><Button :disabled="busy || !draft.name.trim() || !draft.nodeIds.length" @click="saveGroup"><SaveIcon data-icon="inline-start" />保存</Button></DialogFooter>
    </DialogContent>
  </Dialog>
</template>

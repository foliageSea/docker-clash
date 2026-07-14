<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  Activity,
  Cable,
  ChevronRight,
  CircleGauge,
  GitBranch,
  Link2,
  Pencil,
  Network,
  Plus,
  Power,
  RefreshCw,
  Save,
  Server,
  Settings as SettingsIcon,
  Trash2,
  X,
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Toaster } from '@/components/ui/sonner'
import {
  api,
  type EntryGroup,
  type EntryGroupInput,
  type Node,
  type Settings,
  type Status,
} from '@/lib/api'

type View = 'overview' | 'nodes' | 'chains' | 'settings' | 'core'
const view = ref<View>('overview'),
  status = ref<Status>(),
  nodes = ref<Node[]>([]),
  entryGroups = ref<EntryGroup[]>([]),
  settings = ref<Settings>({
    listen: '127.0.0.1:9080',
    mixedPort: 7890,
    allowLan: true,
    bindAddress: '*',
  })
const busy = ref(false),
  showImport = ref(false),
  showGroupDialog = ref(false),
  editingGroupId = ref<string>(),
  uri = ref(''),
  delays = ref<Record<string, number>>({})
const emptyGroup: EntryGroupInput = {
  name: '',
  type: 'select',
  nodeIds: [],
  testUrl: '',
  interval: 60,
}
const groupDraft = ref<EntryGroupInput>({ ...emptyGroup, nodeIds: [] })
const selected = computed(() => nodes.value.find((n) => n.name === settings.value.selectedNode))
const nodeById = computed(() => new Map(nodes.value.map((node) => [node.id, node])))
const entryGroupNames = computed(() => new Set(entryGroups.value.map((group) => group.name)))
const nav = [
  { id: 'overview', label: '概览', icon: CircleGauge },
  { id: 'nodes', label: '节点', icon: Server },
  { id: 'chains', label: '入口与链路', icon: Link2 },
  { id: 'settings', label: '网络设置', icon: Network },
  { id: 'core', label: '内核', icon: Activity },
] as const
async function load() {
  try {
    ;[status.value, nodes.value, settings.value, entryGroups.value] = await Promise.all([
      api.status(),
      api.nodes(),
      api.settings(),
      api.entryGroups(),
    ])
  } catch (e) {
    toast.error((e as Error).message)
  }
}
async function run(fn: () => Promise<unknown>, message = '操作已完成') {
  busy.value = true
  try {
    await fn()
    toast.success(message)
    await load()
    return true
  } catch (e) {
    toast.error((e as Error).message)
    return false
  } finally {
    busy.value = false
  }
}
async function importURI() {
  await run(() => api.importNode(uri.value), '节点或订阅已导入')
  uri.value = ''
  showImport.value = false
}
async function clearNodes() {
  if (confirm(`确定清空全部 ${nodes.value.length} 个节点吗？此操作不可撤销。`))
    await run(() => api.clearNodes(), '全部节点已清空')
}
async function deleteNode(node: Node) {
  if (confirm(`确定删除节点“${node.name}”吗？此操作不可撤销。`))
    await run(() => api.deleteNode(node.id), '节点已删除')
}
async function testDelay(n: Node) {
  await run(async () => {
    const r = await api.delay(n.id)
    delays.value[n.id] = r.delay
  }, '延迟测试完成')
}
async function setDialer(n: Node, value: string) {
  const copy = { ...n, dialerProxy: value || undefined }
  await run(() => api.updateNode(copy), '代理链已更新')
}
function openCreateGroup() {
  editingGroupId.value = undefined
  groupDraft.value = { ...emptyGroup, nodeIds: [] }
  showGroupDialog.value = true
}
function openEditGroup(group: EntryGroup) {
  editingGroupId.value = group.id
  groupDraft.value = {
    name: group.name,
    type: group.type,
    nodeIds: [...group.nodeIds],
    testUrl: group.testUrl || '',
    interval: group.interval || 60,
  }
  showGroupDialog.value = true
}
async function saveGroup() {
  const name = groupDraft.value.name.trim()
  if (!name) return toast.error('请输入入口组名称')
  if (!groupDraft.value.nodeIds.length) return toast.error('请至少选择一个成员节点')
  const payload: EntryGroupInput = {
    name,
    type: groupDraft.value.type,
    nodeIds: [...groupDraft.value.nodeIds],
    ...(groupDraft.value.type === 'fallback'
      ? {
          testUrl: groupDraft.value.testUrl?.trim() || undefined,
          interval: Number(groupDraft.value.interval) || 60,
        }
      : {}),
  }
  const current = entryGroups.value.find((group) => group.id === editingGroupId.value)
  const saved = await run(
    () =>
      current ? api.updateEntryGroup({ ...current, ...payload }) : api.createEntryGroup(payload),
    current ? '入口组已更新' : '入口组已创建',
  )
  if (saved) showGroupDialog.value = false
}
async function deleteGroup(group: EntryGroup) {
  if (confirm(`确定删除入口组“${group.name}”吗？此操作不可撤销。`))
    await run(() => api.deleteEntryGroup(group.id), '入口组已删除')
}
function nodeName(id?: string) {
  return (id && nodeById.value.get(id)?.name) || '未选择'
}
onMounted(load)
</script>

<template>
  <Toaster position="top-right" rich-colors close-button />
  <div class="shell">
    <aside class="sidebar">
      <div class="brand">
        <div class="mark"><Cable :size="19" /></div>
        <div><strong>Nexus</strong><span>Proxy console</span></div>
      </div>
      <nav>
        <button
          v-for="item in nav"
          :key="item.id"
          :class="{ active: view === item.id }"
          @click="view = item.id"
        >
          <component :is="item.icon" :size="18" /><span>{{ item.label }}</span>
        </button>
      </nav>
      <div class="core-state">
        <span :class="['pulse', status?.core.running ? 'online' : '']"></span>
        <div>
          <strong>{{ status?.core.running ? '内核运行中' : '内核已停止' }}</strong
          ><small>{{ status?.core.pid ? `PID ${status.core.pid}` : 'mihomo core' }}</small>
        </div>
      </div>
    </aside>
    <main>
      <header>
        <div>
          <p class="eyebrow">NEXUS CONTROL PLANE</p>
          <h1>{{ nav.find((x) => x.id === view)?.label }}</h1>
        </div>
        <Button variant="outline" size="icon" title="刷新" @click="load"
          ><RefreshCw :size="17"
        /></Button>
      </header>
      <section v-if="view === 'overview'" class="overview">
        <div class="hero-status">
          <div>
            <span class="status-label"
              ><i :class="status?.core.running ? 'ok' : ''"></i
              >{{ status?.core.running ? '网络就绪' : '等待启动' }}</span
            >
            <h2>{{ selected?.name || '尚未选择出口节点' }}</h2>
            <p>
              {{
                selected
                  ? `${selected.type.toUpperCase()} · ${selected.server}:${selected.port}`
                  : '导入节点后选择默认出口'
              }}
            </p>
          </div>
          <Button
            :variant="status?.core.running ? 'destructive' : 'default'"
            :class="
              status?.core.running
                ? 'border-transparent bg-[#c93838] text-white hover:bg-[#a92f2f]'
                : ''
            "
            @click="run(() => api.core(status?.core.running ? 'stop' : 'start'))"
            ><Power :size="16" />{{ status?.core.running ? '停止' : '启动内核' }}</Button
          >
        </div>
        <div class="metrics">
          <article>
            <span>混合代理端口</span><strong>{{ settings.mixedPort }}</strong
            ><small>HTTP + SOCKS5</small>
          </article>
          <article>
            <span>可用节点</span><strong>{{ nodes.length }}</strong
            ><small>手动管理</small>
          </article>
          <article>
            <span>入口组</span><strong>{{ status?.entryGroupCount ?? entryGroups.length }}</strong
            ><small>手动与故障转移</small>
          </article>
          <article>
            <span>局域网访问</span><strong>{{ settings.allowLan ? '已开放' : '已关闭' }}</strong
            ><small>{{ settings.allowLan ? settings.bindAddress : '仅限本机' }}</small>
          </article>
        </div>
        <div class="section-head">
          <div>
            <h3>快速节点选择</h3>
            <p>当前流量将通过 NEXUS 策略组转发</p>
          </div>
          <Button size="sm" @click="view = 'nodes'">管理节点<ChevronRight :size="15" /></Button>
        </div>
        <div class="node-strip">
          <button
            v-for="n in nodes.slice(0, 6)"
            :key="n.id"
            :class="{ selected: n.name === settings.selectedNode }"
            @click="run(() => api.selectNode(n.id), '出口节点已切换')"
          >
            <span class="protocol">{{ n.type.slice(0, 2).toUpperCase() }}</span
            ><span
              ><strong>{{ n.name }}</strong
              ><small>{{ n.server }}</small></span
            ><i></i>
          </button>
          <div v-if="!nodes.length" class="empty">暂无节点</div>
        </div>
      </section>

      <section v-else-if="view === 'nodes'">
        <div class="section-head">
          <div>
            <h3>代理节点</h3>
            <p>支持 SOCKS5、SS、VMess、VLESS、Trojan、Hysteria2 和 TUIC URI</p>
          </div>
          <div class="button-row">
            <Button variant="destructive" :disabled="!nodes.length || busy" @click="clearNodes"
              ><Trash2 :size="16" />清空全部</Button
            ><Button @click="showImport = true"><Plus :size="16" />导入节点</Button>
          </div>
        </div>
        <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>节点</th>
                <th>协议</th>
                <th>服务器</th>
                <th>延迟</th>
                <th>链路入口</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="n in nodes" :key="n.id">
                <td>
                  <div class="node-name">
                    <i :class="{ active: n.name === settings.selectedNode }"></i
                    ><strong>{{ n.name }}</strong>
                  </div>
                </td>
                <td>
                  <span class="tag">{{ n.type }}</span>
                </td>
                <td class="mono">{{ n.server }}:{{ n.port }}</td>
                <td>
                  <button class="delay" @click="testDelay(n)">
                    {{ delays[n.id] ? `${delays[n.id]} ms` : '测试' }}
                  </button>
                </td>
                <td>
                  <span v-if="n.dialerProxy && entryGroupNames.has(n.dialerProxy)" class="group-tag"
                    ><GitBranch :size="12" />{{ n.dialerProxy }}</span
                  ><span v-else>{{ n.dialerProxy || 'DIRECT' }}</span>
                </td>
                <td class="actions">
                  <Button
                    variant="ghost"
                    size="sm"
                    @click="run(() => api.selectNode(n.id), '默认出口已切换')"
                    >设为默认</Button
                  ><Button
                    variant="ghost"
                    size="icon"
                    title="删除"
                    :disabled="busy"
                    @click="deleteNode(n)"
                    ><Trash2 :size="16"
                  /></Button>
                </td>
              </tr>
              <tr v-if="!nodes.length">
                <td colspan="6" class="empty">导入第一个节点以开始使用</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section v-else-if="view === 'chains'">
        <div class="section-head">
          <div>
            <h3>入口节点组</h3>
            <p>将多个节点组织为手动选择或自动故障转移入口</p>
          </div>
          <Button :disabled="busy" @click="openCreateGroup"><Plus :size="16" />创建组</Button>
        </div>
        <div class="entry-group-grid">
          <article v-for="group in entryGroups" :key="group.id" class="entry-group-card">
            <div class="entry-group-main">
              <span class="group-mode" :class="group.type">
                {{ group.type === 'select' ? '手动' : '故障转移' }}
              </span>
              <div>
                <strong>{{ group.name }}</strong>
                <small>{{ group.nodeIds.length }} 个成员</small>
              </div>
            </div>
            <label v-if="group.type === 'select'" class="group-selection">
              <span>当前成员</span>
              <select
                :value="group.selectedNodeId || ''"
                :disabled="busy || !group.nodeIds.length"
                @change="
                  run(
                    () =>
                      api.selectEntryGroupNode(
                        group.id,
                        ($event.target as HTMLSelectElement).value,
                      ),
                    '入口组成员已切换',
                  )
                "
              >
                <option value="" disabled>选择成员</option>
                <option v-for="id in group.nodeIds" :key="id" :value="id">
                  {{ nodeName(id) }}
                </option>
              </select>
            </label>
            <div v-else class="fallback-summary">
              <span>优先顺序</span
              ><strong>{{ group.nodeIds.map((id) => nodeName(id)).join(' → ') }}</strong>
              <small>{{ group.interval || 60 }} 秒检测</small>
            </div>
            <div class="entry-group-actions">
              <Button
                variant="ghost"
                size="icon"
                title="编辑入口组"
                :disabled="busy"
                @click="openEditGroup(group)"
                ><Pencil :size="15"
              /></Button>
              <Button
                variant="ghost"
                size="icon"
                title="删除入口组"
                :disabled="busy"
                @click="deleteGroup(group)"
                ><Trash2 :size="15"
              /></Button>
            </div>
          </article>
          <div v-if="!entryGroups.length" class="empty group-empty">
            尚无入口组，可将多个节点组成统一入口
          </div>
        </div>
        <div class="section-head chain-section-head">
          <div>
            <h3>节点链路</h3>
            <p>为每个节点指定入口组或单节点；复杂循环由后端校验</p>
          </div>
        </div>
        <div class="chain-list">
          <article v-for="n in nodes" :key="n.id">
            <label
              ><span>入口节点</span
              ><select
                :value="n.dialerProxy || ''"
                @change="setDialer(n, ($event.target as HTMLSelectElement).value)"
              >
                <option value="">DIRECT 直连</option>
                <optgroup v-if="entryGroups.length" label="入口组">
                  <option v-for="group in entryGroups" :key="group.id" :value="group.name">
                    {{ group.name }}
                  </option>
                </optgroup>
                <optgroup label="单节点">
                  <option
                    v-for="hop in nodes.filter((x) => x.id !== n.id)"
                    :key="hop.id"
                    :value="hop.name"
                  >
                    {{ hop.name }}
                  </option>
                </optgroup>
              </select></label
            ><ChevronRight :size="18" />
            <div class="chain-node">
              <span class="protocol">{{ n.type.slice(0, 2).toUpperCase() }}</span>
              <div>
                <strong>{{ n.name }}</strong
                ><small>{{ n.server }}:{{ n.port }}</small>
              </div>
            </div>
          </article>
          <div v-if="!nodes.length" class="empty">添加至少两个节点后配置链式代理</div>
        </div>
      </section>

      <section v-else-if="view === 'settings'" class="settings-form">
        <div class="section-head">
          <div>
            <h3>入站网络</h3>
            <p>修改后内核自动重启并应用新配置</p>
          </div>
          <Button :disabled="busy" @click="run(() => api.saveSettings(settings), '网络设置已保存')"
            ><Save :size="16" />保存</Button
          >
        </div>
        <div class="form-grid">
          <label
            ><span>管理界面监听地址</span><Input v-model="settings.listen" /><small
              >更改此项需要手动重启 Nexus 服务</small
            ></label
          ><label
            ><span>Mixed 代理端口</span
            ><Input v-model="settings.mixedPort" type="number" :min="1" :max="65535" /><small
              >同时接受 HTTP 与 SOCKS5</small
            ></label
          ><label
            ><span>内核绑定地址</span
            ><Input v-model="settings.bindAddress" :disabled="!settings.allowLan" /><small
              >局域网开放时通常使用 *</small
            ></label
          ><label class="switch-row"
            ><div><span>允许局域网连接</span><small>可信网络模式，不启用代理认证</small></div>
            <button
              role="switch"
              :aria-checked="settings.allowLan"
              :class="{ on: settings.allowLan }"
              @click="settings.allowLan = !settings.allowLan"
            >
              <i></i></button
          ></label>
        </div>
        <div v-if="settings.allowLan" class="warning">
          <SettingsIcon :size="18" />
          <div>
            <strong>代理端口对局域网开放</strong>
            <p>当前未配置身份认证。请仅在可信网络中使用，并通过防火墙限制访问范围。</p>
          </div>
        </div>
      </section>

      <section v-else class="core-panel">
        <div class="section-head">
          <div>
            <h3>mihomo 内核</h3>
            <p>由 Nexus 子进程管理器托管</p>
          </div>
        </div>
        <div class="core-card">
          <div>
            <span :class="['core-icon', status?.core.running ? 'running' : '']"
              ><Activity :size="22"
            /></span>
            <div>
              <strong>{{ status?.core.running ? '正在运行' : '已停止' }}</strong
              ><small>{{ status?.core.error || 'Controller 仅监听 127.0.0.1:19090' }}</small>
            </div>
          </div>
          <div class="button-row">
            <Button variant="outline" @click="run(() => api.core('restart'))"
              ><RefreshCw :size="16" />重启</Button
            ><Button
              :variant="status?.core.running ? 'destructive' : 'default'"
              @click="run(() => api.core(status?.core.running ? 'stop' : 'start'))"
              ><Power :size="16" />{{ status?.core.running ? '停止' : '启动' }}</Button
            >
          </div>
        </div>
      </section>
    </main>
    <div v-if="showImport" class="modal-backdrop" @click.self="showImport = false">
      <div class="dialog">
        <div class="dialog-head">
          <div>
            <h3>导入节点或订阅</h3>
            <p>粘贴订阅地址或标准代理 URI</p>
          </div>
          <Button variant="ghost" size="icon" @click="showImport = false"><X :size="17" /></Button>
        </div>
        <textarea
          v-model="uri"
          autofocus
          placeholder="https://...、vless://... 或 ss://..."
        ></textarea>
        <div class="dialog-actions">
          <Button variant="outline" @click="showImport = false">取消</Button
          ><Button :disabled="!uri || busy" @click="importURI">导入</Button>
        </div>
      </div>
    </div>
    <div v-if="showGroupDialog" class="modal-backdrop" @click.self="showGroupDialog = false">
      <div
        class="dialog group-dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="group-title"
      >
        <div class="dialog-head">
          <div>
            <h3 id="group-title">{{ editingGroupId ? '编辑入口组' : '创建入口组' }}</h3>
            <p>成员节点可作为其他节点的统一链路入口</p>
          </div>
          <Button
            variant="ghost"
            size="icon"
            title="关闭"
            :disabled="busy"
            @click="showGroupDialog = false"
            ><X :size="17"
          /></Button>
        </div>
        <div class="group-form">
          <label>
            <span>名称</span>
            <Input v-model="groupDraft.name" autofocus placeholder="例如：香港入口" />
          </label>
          <label>
            <span>模式</span>
            <select v-model="groupDraft.type">
              <option value="select">手动</option>
              <option value="fallback">故障转移</option>
            </select>
          </label>
          <div class="member-field">
            <span>成员节点</span>
            <div class="member-options">
              <label v-for="node in nodes" :key="node.id">
                <input v-model="groupDraft.nodeIds" type="checkbox" :value="node.id" />
                <span
                  ><strong>{{ node.name }}</strong
                  ><small>{{ node.type }}</small></span
                >
              </label>
              <div v-if="!nodes.length" class="empty">暂无可选节点</div>
            </div>
          </div>
          <template v-if="groupDraft.type === 'fallback'">
            <label>
              <span>检测 URL</span>
              <Input
                v-model="groupDraft.testUrl"
                placeholder="https://www.gstatic.com/generate_204"
              />
            </label>
            <label>
              <span>检测间隔（秒）</span>
              <Input v-model="groupDraft.interval" type="number" :min="1" />
            </label>
          </template>
        </div>
        <div class="dialog-actions">
          <Button variant="outline" :disabled="busy" @click="showGroupDialog = false">取消</Button>
          <Button
            :disabled="busy || !groupDraft.name.trim() || !groupDraft.nodeIds.length"
            @click="saveGroup"
          >
            <Save :size="16" />{{ editingGroupId ? '保存' : '创建' }}
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>

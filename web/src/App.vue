<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  Activity,
  Cable,
  ChevronRight,
  CircleGauge,
  Link2,
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
import { api, type Node, type Settings, type Status } from '@/lib/api'

type View = 'overview' | 'nodes' | 'chains' | 'settings' | 'core'
const view = ref<View>('overview'),
  status = ref<Status>(),
  nodes = ref<Node[]>([]),
  settings = ref<Settings>({
    listen: '127.0.0.1:9080',
    mixedPort: 7890,
    allowLan: true,
    bindAddress: '*',
  })
const busy = ref(false),
  showImport = ref(false),
  uri = ref(''),
  delays = ref<Record<string, number>>({})
const selected = computed(() => nodes.value.find((n) => n.name === settings.value.selectedNode))
const nav = [
  { id: 'overview', label: '概览', icon: CircleGauge },
  { id: 'nodes', label: '节点', icon: Server },
  { id: 'chains', label: '代理链', icon: Link2 },
  { id: 'settings', label: '网络设置', icon: Network },
  { id: 'core', label: '内核', icon: Activity },
] as const
async function load() {
  try {
    ;[status.value, nodes.value, settings.value] = await Promise.all([
      api.status(),
      api.nodes(),
      api.settings(),
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
  } catch (e) {
    toast.error((e as Error).message)
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
                <td>{{ n.dialerProxy || 'DIRECT' }}</td>
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
                    @click="run(() => api.deleteNode(n.id), '节点已删除')"
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
            <h3>节点链路入口</h3>
            <p>为每个节点指定一个入口节点，可继续串联；后端会阻止循环链路</p>
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
                <option
                  v-for="hop in nodes.filter((x) => x.id !== n.id)"
                  :key="hop.id"
                  :value="hop.name"
                >
                  {{ hop.name }}
                </option>
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
  </div>
</template>

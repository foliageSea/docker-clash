import { ActivityIcon, GitBranchIcon, LayoutDashboardIcon, NetworkIcon, ServerIcon } from '@lucide/vue'

export const navigation = [
  { name: 'overview', path: '/', label: '运行概览', group: '代理', icon: LayoutDashboardIcon },
  { name: 'nodes', path: '/nodes', label: '代理节点', group: '代理', icon: ServerIcon },
  { name: 'chains', path: '/chains', label: '入口与链路', group: '代理', icon: GitBranchIcon },
  { name: 'settings', path: '/settings', label: '网络设置', group: '系统', icon: NetworkIcon },
  { name: 'core', path: '/core', label: '内核管理', group: '系统', icon: ActivityIcon },
] as const

export type ConsolePage = (typeof navigation)[number]

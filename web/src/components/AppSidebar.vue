<script setup lang="ts">
import { ActivityIcon, CableIcon, GitBranchIcon, LayoutDashboardIcon, NetworkIcon, ServerIcon } from '@lucide/vue'
import { useRoute } from 'vue-router'

import {
  Sidebar, SidebarContent, SidebarFooter, SidebarGroup, SidebarGroupContent, SidebarGroupLabel,
  SidebarHeader, SidebarMenu, SidebarMenuButton, SidebarMenuItem, SidebarRail,
} from '@/components/ui/sidebar'
import { useConsole } from '@/composables/use-console'

const route = useRoute()
const { status } = useConsole()
const groups = [
  { label: '工作区', items: [
    { title: '运行概览', url: '/', icon: LayoutDashboardIcon },
    { title: '代理节点', url: '/nodes', icon: ServerIcon },
    { title: '入口与链路', url: '/chains', icon: GitBranchIcon },
  ] },
  { label: '系统', items: [
    { title: '网络设置', url: '/settings', icon: NetworkIcon },
    { title: '内核管理', url: '/core', icon: ActivityIcon },
  ] },
]
</script>

<template>
  <Sidebar collapsible="icon">
    <SidebarHeader class="border-b py-2">
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton size="lg" tooltip="Docker Clash" as-child>
            <RouterLink to="/">
              <span class="flex size-8 items-center justify-center rounded-md bg-sidebar-primary text-sidebar-primary-foreground">
                <CableIcon />
              </span>
              <span class="flex min-w-0 flex-col leading-tight">
                <strong class="truncate">Docker Clash</strong>
                <small class="truncate text-muted-foreground">Proxy control plane</small>
              </span>
            </RouterLink>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarHeader>
    <SidebarContent>
      <SidebarGroup v-for="group in groups" :key="group.label">
        <SidebarGroupLabel>{{ group.label }}</SidebarGroupLabel>
        <SidebarGroupContent>
          <SidebarMenu>
            <SidebarMenuItem v-for="item in group.items" :key="item.url">
              <SidebarMenuButton as-child :is-active="route.path === item.url" :tooltip="item.title">
                <RouterLink :to="item.url"><component :is="item.icon" /><span>{{ item.title }}</span></RouterLink>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroupContent>
      </SidebarGroup>
    </SidebarContent>
    <SidebarFooter class="border-t py-2">
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton tooltip="内核状态" as-child>
            <RouterLink to="/core">
              <span :class="['size-2 rounded-full', status?.core.running ? 'bg-emerald-500' : 'bg-muted-foreground']" />
              <span>{{ status?.core.running ? '内核运行中' : '内核已停止' }}</span>
            </RouterLink>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarFooter>
    <SidebarRail />
  </Sidebar>
</template>

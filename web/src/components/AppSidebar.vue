<script setup lang="ts">
import { ActivityIcon, CableIcon, GitBranchIcon, LayoutDashboardIcon, NetworkIcon, ServerIcon } from '@lucide/vue'
import { useRoute } from 'vue-router'

import {
  Sidebar, SidebarContent, SidebarFooter, SidebarGroup, SidebarGroupContent,
  SidebarHeader, SidebarMenu, SidebarMenuButton, SidebarMenuItem, SidebarRail, SidebarSeparator, useSidebar,
} from '@/components/ui/sidebar'
import { useConsole } from '@/composables/use-console'
import { cn } from '@/lib/utils'

const route = useRoute()
const { setOpenMobile } = useSidebar()
const { status } = useConsole()
const groups = [
  { id: 'proxy', items: [
    { title: '运行概览', url: '/', icon: LayoutDashboardIcon },
    { title: '代理节点', url: '/nodes', icon: ServerIcon },
    { title: '入口与链路', url: '/chains', icon: GitBranchIcon },
  ] },
  { id: 'system', items: [
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
            <RouterLink to="/" aria-label="Docker Clash" @click="setOpenMobile(false)">
              <span class="flex size-8 shrink-0 items-center justify-center rounded-md bg-sidebar-primary text-sidebar-primary-foreground">
                <CableIcon />
              </span>
              <strong class="truncate group-data-[collapsible=icon]:hidden">Docker Clash</strong>
            </RouterLink>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarHeader>
    <SidebarContent>
      <nav aria-label="主导航" class="flex flex-col gap-1 py-1">
        <template v-for="(group, index) in groups" :key="group.id">
          <SidebarSeparator v-if="index > 0" />
          <SidebarGroup>
            <SidebarGroupContent>
              <SidebarMenu>
                <SidebarMenuItem v-for="item in group.items" :key="item.url">
                  <SidebarMenuButton as-child :is-active="route.path === item.url" :tooltip="item.title">
                    <RouterLink :to="item.url" :aria-label="item.title" @click="setOpenMobile(false)">
                      <component :is="item.icon" /><span class="group-data-[collapsible=icon]:hidden">{{ item.title }}</span>
                    </RouterLink>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </template>
      </nav>
    </SidebarContent>
    <SidebarFooter v-if="status" class="border-t py-2">
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton :tooltip="status.core.running ? '内核运行中' : '内核已停止'" as-child>
            <RouterLink to="/core" :aria-label="status.core.running ? '内核运行中' : '内核已停止'" @click="setOpenMobile(false)">
              <span aria-hidden="true" :class="cn('size-2 shrink-0 rounded-full group-data-[collapsible=icon]:mx-auto', status.core.running ? 'bg-primary' : 'bg-muted-foreground')" />
              <span class="group-data-[collapsible=icon]:hidden">{{ status.core.running ? '内核运行中' : '内核已停止' }}</span>
            </RouterLink>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarFooter>
    <SidebarRail />
  </Sidebar>
</template>

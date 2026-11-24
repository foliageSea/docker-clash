<script setup lang="ts">
import { CableIcon } from '@lucide/vue'
import { watch } from 'vue'
import { useRoute } from 'vue-router'

import {
  Sidebar, SidebarContent, SidebarFooter, SidebarGroup, SidebarGroupContent,
  SidebarGroupLabel, SidebarHeader, SidebarMenu, SidebarMenuButton, SidebarMenuItem, useSidebar,
} from '@/components/ui/sidebar'
import { useConsole } from '@/composables/use-console'
import { navigation } from '@/lib/navigation'
import { cn } from '@/lib/utils'

const route = useRoute()
const { isMobile, setOpenMobile } = useSidebar()
const { status } = useConsole()
const groups = [...new Set(navigation.map(page => page.group))]
watch([isMobile, () => route.path], () => setOpenMobile(false))
</script>

<template>
  <Sidebar collapsible="icon" class="h-full border-r-0">
    <SidebarHeader class="h-20 justify-center px-3 group-data-[collapsible=icon]:items-center">
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton size="lg" tooltip="Docker Clash" as-child>
            <RouterLink to="/" aria-label="Docker Clash" @click="setOpenMobile(false)">
              <span class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground"><CableIcon class="size-4" aria-hidden="true" /></span>
              <strong class="truncate group-data-[collapsible=icon]:hidden">Docker Clash</strong>
            </RouterLink>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarHeader>
    <SidebarContent>
      <nav aria-label="主导航" class="flex flex-col gap-4 px-2.5">
        <SidebarGroup v-for="group in groups" :key="group" class="gap-1 p-0">
          <SidebarGroupLabel class="px-2 group-data-[collapsible=icon]:hidden">{{ group }}</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu class="group-data-[collapsible=icon]:items-center">
              <SidebarMenuItem v-for="page in navigation.filter(item => item.group === group)" :key="page.path">
                <SidebarMenuButton class="h-11 md:h-9 md:text-xs" as-child :is-active="route.path === page.path" :tooltip="page.label">
                  <RouterLink :to="page.path" :aria-label="page.label" @click="setOpenMobile(false)">
                    <component :is="page.icon" aria-hidden="true" /><span class="group-data-[collapsible=icon]:hidden">{{ page.label }}</span>
                  </RouterLink>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </nav>
    </SidebarContent>
    <SidebarFooter v-if="status" class="border-t p-3 group-data-[collapsible=icon]:items-center">
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
  </Sidebar>
</template>

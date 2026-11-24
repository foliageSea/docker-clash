<script setup lang="ts">
import { MoonIcon, SunIcon } from '@lucide/vue'
import { useColorMode } from '@vueuse/core'
import { computed } from 'vue'

import AppSidebar from '@/components/AppSidebar.vue'
import { Button } from '@/components/ui/button'
import { SidebarInset, SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'

const mode = useColorMode()
const dark = computed(() => mode.value === 'dark')
</script>

<template>
  <SidebarProvider>
    <AppSidebar />
    <SidebarInset class="min-w-0">
      <header class="sticky top-0 z-40 flex h-14 items-center gap-3 border-b bg-background/95 px-4 backdrop-blur">
        <SidebarTrigger aria-label="切换导航栏" />
        <span class="truncate text-sm font-semibold md:hidden">Docker Clash</span>
        <Button class="ml-auto" variant="ghost" size="icon" :aria-label="dark ? '切换到浅色模式' : '切换到深色模式'" @click="mode = dark ? 'light' : 'dark'">
          <SunIcon v-if="dark" /><MoonIcon v-else />
        </Button>
      </header>
      <div class="@container mx-auto w-full max-w-[1500px] p-4 sm:p-6">
        <RouterView />
      </div>
    </SidebarInset>
  </SidebarProvider>
</template>

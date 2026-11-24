<script setup lang="ts">
import { ChevronRightIcon } from '@lucide/vue'
import { computed, nextTick, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import AppSidebar from '@/components/AppSidebar.vue'
import WorkspaceTitleBar from '@/components/WorkspaceTitleBar.vue'
import { SidebarProvider } from '@/components/ui/sidebar'
import { navigation } from '@/lib/navigation'

const route = useRoute()
const currentPage = computed(() => navigation.find(page => page.path === route.path))
const viewport = ref<HTMLElement>()
watch(() => route.path, async () => {
  await nextTick()
  viewport.value?.scrollTo({ top: 0, left: 0 })
})
</script>

<template>
  <SidebarProvider class="h-dvh min-h-0 flex-col overflow-hidden bg-sidebar">
    <WorkspaceTitleBar />
    <div class="flex min-h-0 min-w-0 flex-1">
      <AppSidebar />
      <div class="flex min-w-0 flex-1 flex-col">
        <nav v-if="currentPage" aria-label="当前位置" class="flex h-10 shrink-0 items-center gap-2 px-4 text-xs text-muted-foreground md:h-12 md:px-7">
          <span>{{ currentPage.group }}</span><ChevronRightIcon class="size-3" aria-hidden="true" /><span aria-current="page" class="text-foreground">{{ currentPage.label }}</span>
        </nav>
        <main ref="viewport" tabindex="-1" class="min-h-0 min-w-0 flex-1 overflow-auto overscroll-contain border-t bg-background md:rounded-tl-xl md:border-l">
          <div class="@container mx-auto w-full min-w-0 max-w-7xl p-4 sm:p-5 md:p-7">
            <RouterView />
          </div>
        </main>
      </div>
    </div>
  </SidebarProvider>
</template>

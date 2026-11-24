<script setup lang="ts">
import { CableIcon, MoonIcon, SearchIcon, SunIcon, XIcon } from '@lucide/vue'
import { useColorMode, useEventListener } from '@vueuse/core'
import { computed, nextTick, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { SidebarTrigger, useSidebar } from '@/components/ui/sidebar'
import { cn } from '@/lib/utils'
import { navigation, type ConsolePage } from '@/lib/navigation'

const route = useRoute()
const router = useRouter()
const { isMobile, open, openMobile } = useSidebar()
const mode = useColorMode()
const dark = computed(() => mode.value === 'dark')
const opened = ref<ConsolePage[]>([navigation[0]])
const tabs = ref<HTMLElement>()
const searching = ref(false)
const query = ref('')
const results = computed(() => {
  const value = query.value.trim().toLowerCase()
  return navigation.filter(page => (page.label + ' ' + page.group + ' ' + page.name).toLowerCase().includes(value))
})

function tabLink(path: string) {
  return [...(tabs.value?.querySelectorAll<HTMLAnchorElement>('[data-route-tab]') ?? [])]
    .find(link => link.dataset.routeTab === path)
}

watch(() => route.path, async (path) => {
  const page = navigation.find(item => item.path === path)
  if (page && !opened.value.some(item => item.path === path)) opened.value.push(page)
  await nextTick()
  tabLink(path)?.scrollIntoView({ block: 'nearest', inline: 'nearest' })
}, { immediate: true })

async function closeTab(page: ConsolePage) {
  const index = opened.value.findIndex(item => item.path === page.path)
  if (index < 1) return
  const restoreFocus = tabLink(page.path)?.parentElement?.contains(document.activeElement)
  if (route.path === page.path) {
    await router.push(opened.value[index - 1]!.path)
    if (route.path === page.path) return
  }
  opened.value = opened.value.filter(item => item.path !== page.path)
  await nextTick()
  if (restoreFocus) tabLink(route.path)?.focus()
}

function navigate(page?: ConsolePage) {
  if (!page) return
  searching.value = false
  void router.push(page.path)
}

function submitSearch(event: KeyboardEvent) {
  if (event.isComposing || event.keyCode === 229) return
  event.preventDefault()
  navigate(results.value[0])
}

function restoreSearchFocus(event: Event) {
  const trigger = document.getElementById('page-search-trigger')
  if (!trigger) return
  event.preventDefault()
  trigger.focus({ preventScroll: true })
}

watch(searching, (value) => { if (!value) query.value = '' })
const shortcut = /Mac|iPhone|iPad/.test(navigator.userAgent) ? '⌘ K' : 'Ctrl K'
useEventListener(window, 'keydown', (event) => {
  if (!event.repeat && (event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'k') {
    event.preventDefault()
    searching.value = !searching.value
  }
})
</script>

<template>
  <header class="flex shrink-0 flex-wrap items-center border-b bg-sidebar md:h-11 md:flex-nowrap">
    <div class="flex h-11 shrink-0 items-center gap-2 px-3">
      <SidebarTrigger id="sidebar-toggle" :aria-label="isMobile ? '打开主导航' : '切换导航栏'" :aria-expanded="isMobile ? openMobile : open" />
      <CableIcon class="size-4 text-primary" aria-hidden="true" />
      <span class="text-xs font-semibold md:hidden xl:inline">Docker Clash</span>
    </div>
    <nav aria-label="打开的页面" class="order-last h-10 w-full min-w-0 flex-none border-t px-3 md:order-none md:h-full md:w-auto md:flex-1 md:border-t-0">
      <div ref="tabs" class="route-tab-list flex h-full min-w-0 items-center gap-1 overflow-x-auto">
        <div v-for="page in opened" :key="page.path" :class="cn('flex h-8 shrink-0 items-center gap-1 rounded-md border p-0.5', page.path === route.path ? 'border-border bg-background text-foreground shadow-xs' : 'border-transparent text-muted-foreground hover:bg-accent hover:text-foreground')">
          <RouterLink :to="page.path" :data-route-tab="page.path" class="flex h-full items-center gap-2 rounded-sm px-2 text-xs outline-none focus-visible:ring-2 focus-visible:ring-ring">
            <component :is="page.icon" class="size-3.5" aria-hidden="true" />{{ page.label }}
          </RouterLink>
          <Button v-if="page.path !== '/'" variant="ghost" size="icon-xs" :aria-label="'关闭' + page.label" @click="closeTab(page)"><XIcon /></Button>
        </div>
      </div>
    </nav>
    <div class="ml-auto flex h-11 shrink-0 items-center gap-1 pr-3">
      <Dialog v-model:open="searching">
        <DialogTrigger as-child>
          <Button id="page-search-trigger" variant="ghost" size="sm" class="px-2" aria-label="搜索页面" :aria-keyshortcuts="shortcut === '⌘ K' ? 'Meta+K' : 'Control+K'">
            <SearchIcon data-icon="inline-start" /><span class="hidden lg:inline">搜索页面</span><kbd class="hidden rounded-sm border bg-background px-1 font-mono text-[10px] text-muted-foreground lg:inline">{{ shortcut }}</kbd>
          </Button>
        </DialogTrigger>
        <DialogContent class="gap-3" @close-auto-focus="restoreSearchFocus">
          <DialogTitle>搜索页面</DialogTitle>
          <DialogDescription class="sr-only">按页面名称筛选，按回车打开首个结果。</DialogDescription>
          <Input v-model="query" aria-label="页面名称" placeholder="输入页面名称…" @keydown.enter="submitSearch" />
          <ul v-if="results.length" class="flex flex-col gap-1">
            <li v-for="page in results" :key="page.path">
              <Button variant="ghost" class="w-full justify-start" @click="navigate(page)"><component :is="page.icon" data-icon="inline-start" />{{ page.label }}</Button>
            </li>
          </ul>
          <p v-else role="status" class="text-sm text-muted-foreground">没有匹配的页面</p>
        </DialogContent>
      </Dialog>
      <Button variant="ghost" size="icon-sm" :aria-label="dark ? '切换到浅色模式' : '切换到深色模式'" @click="mode = dark ? 'light' : 'dark'"><SunIcon v-if="dark" /><MoonIcon v-else /></Button>
    </div>
  </header>
</template>

<style scoped>
.route-tab-list { scrollbar-width: none; }
</style>

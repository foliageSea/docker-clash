import { createRouter, createWebHistory } from 'vue-router'

import { navigation } from '@/lib/navigation'
import ChainsPage from '@/pages/ChainsPage.vue'
import CorePage from '@/pages/CorePage.vue'
import NodesPage from '@/pages/NodesPage.vue'
import OverviewPage from '@/pages/OverviewPage.vue'
import SettingsPage from '@/pages/SettingsPage.vue'

const pages = { overview: OverviewPage, nodes: NodesPage, chains: ChainsPage, settings: SettingsPage, core: CorePage }

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    ...navigation.map(({ path, name, label }) => ({ path, name, component: pages[name], meta: { title: label } })),
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

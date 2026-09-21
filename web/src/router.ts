import { createRouter, createWebHistory } from 'vue-router'

import ChainsPage from '@/pages/ChainsPage.vue'
import CorePage from '@/pages/CorePage.vue'
import NodesPage from '@/pages/NodesPage.vue'
import OverviewPage from '@/pages/OverviewPage.vue'
import SettingsPage from '@/pages/SettingsPage.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'overview', component: OverviewPage, meta: { title: '运行概览' } },
    { path: '/nodes', name: 'nodes', component: NodesPage, meta: { title: '代理节点' } },
    { path: '/chains', name: 'chains', component: ChainsPage, meta: { title: '入口与链路' } },
    { path: '/settings', name: 'settings', component: SettingsPage, meta: { title: '网络设置' } },
    { path: '/core', name: 'core', component: CorePage, meta: { title: '内核管理' } },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

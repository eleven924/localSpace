import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/files'
  },
  {
    path: '/files',
    name: 'Files',
    component: () => import('@/views/FilesView.vue')
  },
  {
    path: '/import',
    name: 'Import',
    component: () => import('@/views/ImportView.vue')
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('@/views/SettingsView.vue')
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router
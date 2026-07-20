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
    path: '/collections',
    name: 'Collections',
    component: () => import('@/views/CollectionsView.vue')
  },
  {
    path: '/import',
    name: 'Import',
    component: () => import('@/views/ImportView.vue')
  },
  {
    path: '/tasks',
    name: 'Tasks',
    component: () => import('@/views/TasksView.vue')
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('@/views/SettingsView.vue')
  },
  {
    path: '/documentation',
    name: 'Documentation',
    component: () => import('@/views/DocumentationView.vue')
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router

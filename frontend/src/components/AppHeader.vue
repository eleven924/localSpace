<template>
  <div class="app-chrome">
    <aside class="app-sidebar" :class="{ 'is-collapsed': uiStore.sidebarCollapsed }" aria-label="主导航">
      <div class="sidebar-brand">
        <router-link to="/files" class="brand-mark" aria-label="返回资料库">LS</router-link>
        <div v-if="!uiStore.sidebarCollapsed" class="brand-copy">
          <strong>LocalSpace</strong>
          <span>静谧资料工作台</span>
        </div>
      </div>

      <nav class="sidebar-nav" aria-label="应用导航">
        <span v-if="!uiStore.sidebarCollapsed" class="sidebar-section-label">工作区</span>
        <router-link
          v-for="item in primaryLinks"
          :key="item.path"
          :to="item.path"
          class="sidebar-link"
          :title="item.label"
          :aria-label="item.label"
        >
          <svg class="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path v-for="path in item.iconPaths" :key="path" :d="path" />
          </svg>
          <span v-if="!uiStore.sidebarCollapsed" class="sidebar-link-copy">
            <span>{{ item.label }}</span>
            <small v-if="item.path === '/tasks' && jobsStore.totalRunningCount > 0">{{ jobsStore.totalRunningCount }} 项进行中</small>
          </span>
        </router-link>
      </nav>

      <div class="sidebar-footer">
        <router-link to="/settings" class="sidebar-link" title="设置" aria-label="设置">
          <svg class="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <circle cx="12" cy="12" r="3" />
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 1 1-4 0v-.09a1.65 1.65 0 0 0-1-1.51 1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.6 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 1 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.6a1.65 1.65 0 0 0 1-1.51V3a2 2 0 1 1 4 0v.09A1.65 1.65 0 0 0 15 4.6a1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9c.17.61.72 1 1.51 1H21a2 2 0 1 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
          </svg>
          <span v-if="!uiStore.sidebarCollapsed" class="sidebar-link-copy"><span>设置</span></span>
        </router-link>
        <router-link to="/documentation" class="sidebar-link" title="使用说明" aria-label="使用说明">
          <svg class="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M4 5.5A2.5 2.5 0 0 1 6.5 3H20v16H6.5A2.5 2.5 0 0 0 4 21.5z" />
            <path d="M4 5.5v16" />
            <path d="M8 7h8M8 11h8" />
          </svg>
          <span v-if="!uiStore.sidebarCollapsed" class="sidebar-link-copy"><span>使用说明</span></span>
        </router-link>
        <button
          type="button"
          class="sidebar-collapse-row"
          :title="uiStore.sidebarCollapsed ? '展开侧栏' : '收起侧栏'"
          :aria-label="uiStore.sidebarCollapsed ? '展开侧栏' : '收起侧栏'"
          :aria-expanded="String(!uiStore.sidebarCollapsed)"
          @click="uiStore.toggleSidebar"
        >
          <svg class="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path v-if="uiStore.sidebarCollapsed" d="m9 18 6-6-6-6" />
            <path v-else d="m15 18-6-6 6-6" />
          </svg>
          <span v-if="!uiStore.sidebarCollapsed">收起侧栏</span>
        </button>
      </div>
    </aside>

    <header class="app-header">
      <div class="header-actions" aria-label="系统状态">
        <TaskStatusIndicator />
        <NotificationBell />
      </div>
    </header>
  </div>
</template>

<script setup lang="ts">
import NotificationBell from '@/components/NotificationBell.vue'
import TaskStatusIndicator from '@/components/TaskStatusIndicator.vue'
import { useJobsStore } from '@/store/modules/jobs'
import { useUiStore } from '@/store/modules/ui'

const jobsStore = useJobsStore()
const uiStore = useUiStore()

const primaryLinks = [
  { path: '/files', label: '资料库', iconPaths: ['M3 7.5A2.5 2.5 0 0 1 5.5 5h4l2 2h7A2.5 2.5 0 0 1 21 9.5v7a2.5 2.5 0 0 1-2.5 2.5h-13A2.5 2.5 0 0 1 3 16.5z'] },
  { path: '/collections', label: '合集', iconPaths: ['M4 6h6l2 2h8v10H4z', 'M4 10h16'] },
  { path: '/import', label: '导入', iconPaths: ['M12 3v12', 'm7 10 5 5 5-5', 'M4 19h16'] },
  { path: '/tasks', label: '任务', iconPaths: ['M5 5h14v14H5z', 'm8 12 2.5 2.5L16 9'] },
]

</script>

<style scoped>
.app-chrome { display: contents; }
.app-sidebar { position: fixed; z-index: 80; inset: 12px auto 12px 12px; display: flex; flex-direction: column; width: 236px; padding: 18px 12px 12px; border: 1px solid rgba(146,165,192,.22); border-radius: 22px; background: rgba(249,251,254,.9); box-shadow: 0 18px 44px rgba(55,75,110,.12); backdrop-filter: blur(18px); transition: width .2s ease; }
.app-sidebar.is-collapsed { width: 72px; }
.sidebar-brand { display: flex; align-items: center; gap: 10px; min-height: 42px; padding: 0 8px 18px; }
.brand-mark { display: inline-flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: 0 0 34px; border-radius: 11px; background: #2d405e; color: #fff; font-size: 12px; font-weight: 800; letter-spacing: .08em; }
.brand-copy, .sidebar-link-copy { display: grid; gap: 2px; min-width: 0; }
.brand-copy strong { color: var(--text-color); font-size: 14px; }
.brand-copy span, .sidebar-section-label { color: var(--text-faint); font-size: 10px; letter-spacing: .12em; text-transform: uppercase; }
.sidebar-nav, .sidebar-footer { display: grid; gap: 5px; }
.sidebar-nav { padding-top: 12px; }
.sidebar-section-label { padding: 0 10px 7px; }
.sidebar-link, .sidebar-collapse-row { position: relative; display: flex; align-items: center; gap: 11px; width: 100%; min-height: 40px; padding: 8px 10px; border-radius: 12px; color: var(--text-soft); text-align: left; }
.sidebar-link:hover, .sidebar-collapse-row:hover { background: rgba(231,237,247,.78); color: var(--text-color); }
.sidebar-link.router-link-active { background: rgba(221,230,246,.92); color: #3f61a3; box-shadow: inset 3px 0 0 var(--primary-color); }
.sidebar-icon { width: 18px; height: 18px; flex: 0 0 18px; }
.sidebar-link-copy { font-size: 13px; font-weight: 600; }
.sidebar-link-copy small { color: var(--text-faint); font-size: 10px; font-weight: 500; }
.sidebar-footer { margin-top: auto; padding-top: 12px; border-top: 1px solid rgba(146,165,192,.18); }
.sidebar-collapse-row { color: var(--text-faint); font-size: 12px; }
.app-sidebar.is-collapsed .sidebar-brand, .app-sidebar.is-collapsed .sidebar-link, .app-sidebar.is-collapsed .sidebar-collapse-row { justify-content: center; padding-inline: 0; }
.app-header { position: static; z-index: 70; display: flex; align-items: center; justify-content: flex-end; width: auto; height: 42px; min-height: 42px; margin: 0; padding: 0 22px; border-bottom: 1px solid rgba(199,211,226,.72); background: rgba(244,247,251,.34); box-shadow: none; pointer-events: auto; }
.header-context { display: flex; align-items: baseline; gap: 9px; }
.header-kicker { color: var(--text-faint); font-size: 10px; font-weight: 700; letter-spacing: .16em; }
.header-separator { color: var(--border-color); }
.header-context strong { color: var(--text-color); font-size: 14px; }
.header-actions { display: flex; align-items: center; gap: 8px; margin-left: 0; pointer-events: auto; }
:global(.page-shell:has(.app-sidebar)) { position: relative; display: grid; grid-template-columns: var(--sidebar-width) minmax(0, 1fr); grid-template-rows: 42px minmax(0, 1fr); column-gap: 0; }
:global(.page-shell:has(.app-sidebar) .app-sidebar) { position: static; grid-column: 1; grid-row: 1 / 3; width: auto; min-width: 0; margin: 0; padding: 24px 14px 16px; border-width: 0 1px 0 0; border-radius: 0; box-shadow: none; }
:global(.page-shell:has(.app-sidebar) .app-sidebar.is-collapsed) { width: auto; }
:global(.page-shell:has(.app-sidebar) .app-header) { grid-column: 2; grid-row: 1; pointer-events: auto; }
:global(.page-shell:has(.app-sidebar) .page-content) { grid-column: 2; grid-row: 2; padding: 18px 44px 50px; transition: none; }
:global(.page-shell:has(.app-sidebar)) { --sidebar-width: 224px; }
:global(.page-shell:has(.app-sidebar.is-collapsed)) { --sidebar-width: 76px; }
:global(.page-shell:has(.app-sidebar) .page-topbar-title) { font-size: 18px; }
:global(.page-shell:has(.app-sidebar) .page-topbar-note) { font-size: 12px; }
:global(.app-header .task-button),
:global(.app-header .notification-bell) { width: 32px; height: 32px; border: 0; border-radius: 8px; background: transparent; box-shadow: none; }
@media (max-width: 900px) {
  :global(.page-shell:has(.app-sidebar) .app-sidebar) { border-radius: 0; }
  :global(.page-shell:has(.app-sidebar) .app-header) { height: 38px; min-height: 38px; padding-inline: 16px; }
  :global(.page-shell:has(.app-sidebar) .page-content) { padding: 16px 24px 44px; }
}
@media (max-width: 640px) {
  :global(.page-shell:has(.app-sidebar)) { display: flex; }
  .app-sidebar { display: none; }
  .app-header { height: 38px; min-height: 38px; padding-inline: 12px; }
}
</style>

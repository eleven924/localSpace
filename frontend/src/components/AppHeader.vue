<template>
  <header class="app-header">
    <div class="header-left">
      <router-link to="/files" class="brand-title">LocalSpace</router-link>

      <nav class="header-nav" aria-label="主导航">
        <router-link to="/files" class="nav-pill">文件</router-link>
        <router-link to="/collections" class="nav-pill">合集</router-link>
        <router-link to="/tasks" class="nav-pill">任务</router-link>
      </nav>
    </div>

    <div class="header-actions">
      <TaskStatusIndicator />
      <NotificationBell />
      <div class="actions-divider" aria-hidden="true"></div>
      <router-link
        to="/import"
        class="icon-action"
        :class="{ active: isImportActive }"
        title="导入"
        aria-label="导入"
      >
        <svg
          class="action-icon"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
          <polyline points="7 10 12 15 17 10" />
          <line x1="12" y1="15" x2="12" y2="3" />
        </svg>
      </router-link>
      <router-link
        to="/settings"
        class="icon-action"
        :class="{ active: isSettingsActive }"
        title="设置"
        aria-label="设置"
      >
        <svg
          class="action-icon"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <circle cx="12" cy="12" r="3" />
          <path
            d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"
          />
        </svg>
      </router-link>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import NotificationBell from '@/components/NotificationBell.vue'
import TaskStatusIndicator from '@/components/TaskStatusIndicator.vue'

const route = useRoute()

const isImportActive = computed(() => route.name === 'Import')
const isSettingsActive = computed(() => ['Settings', 'Documentation'].includes(String(route.name ?? '')))
</script>

<style scoped>
.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  padding: 12px 20px;
  background-color: color-mix(in srgb, var(--surface-color) 92%, transparent);
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
  min-width: 0;
}

.brand-title {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-color);
}

.header-nav {
  display: inline-flex;
  gap: 6px;
  padding: 4px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--surface-color) 82%, transparent);
  border: 1px solid rgba(148, 163, 184, 0.16);
}

.nav-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 64px;
  padding: 7px 12px;
  border-radius: 999px;
  color: var(--text-color);
  font-size: 12px;
  font-weight: 600;
}

.nav-pill.router-link-active {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--primary-hover) 100%);
  color: #fff;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.actions-divider {
  width: 1px;
  height: 20px;
  background-color: var(--border-color);
  margin: 0 4px;
}

.icon-action {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 1px solid transparent;
  background: transparent;
  color: var(--text-color);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s ease, color 0.15s ease, border-color 0.15s ease;
}

.icon-action:hover {
  background: var(--hover-bg, rgba(148, 163, 184, 0.24));
  color: var(--primary-color);
}

.icon-action.active {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--primary-hover) 100%);
  color: #fff;
  border-color: var(--primary-color);
}

.action-icon {
  width: 18px;
  height: 18px;
  pointer-events: none;
}

@media (max-width: 900px) {
  .app-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .header-left {
    width: 100%;
    flex-direction: column;
    align-items: flex-start;
  }

  .header-actions {
    width: 100%;
    justify-content: flex-end;
  }
}

@media (max-width: 640px) {
  .app-header {
    padding: 12px 14px;
  }
}
</style>

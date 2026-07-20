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
      <router-link to="/import" :class="['btn', isImportActive ? 'primary' : 'secondary']">
        导入
      </router-link>
      <router-link to="/settings" :class="['btn', isSettingsActive ? 'primary' : 'secondary']">
        设置
      </router-link>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
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

  .header-actions {
    width: 100%;
    flex-direction: column;
  }
}
</style>

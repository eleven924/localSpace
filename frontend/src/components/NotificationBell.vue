<template>
  <div class="notification-bell-wrapper" ref="wrapperRef">
    <button
      type="button"
      class="notification-bell"
      :class="{ active: panelVisible }"
      aria-label="消息中心"
      @click="togglePanel"
    >
      <span class="bell-icon">🔔</span>
      <span v-if="unreadCount > 0" class="unread-badge">{{ unreadCount > 99 ? '99+' : unreadCount }}</span>
    </button>
    <NotificationPanel
      :visible="panelVisible"
      @navigate="handleNavigate"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useNotificationsStore } from '@/store/modules/notifications'
import NotificationPanel from './NotificationPanel.vue'

const { unreadCount } = useNotificationsStore()
const router = useRouter()

const panelVisible = ref(false)
const wrapperRef = ref<HTMLElement | null>(null)

const togglePanel = () => {
  panelVisible.value = !panelVisible.value
}

const handleNavigate = (path: string) => {
  panelVisible.value = false
  router.push(path)
}

const handleClickOutside = (event: MouseEvent) => {
  if (wrapperRef.value && !wrapperRef.value.contains(event.target as Node)) {
    panelVisible.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.notification-bell-wrapper {
  position: relative;
}

.notification-bell {
  position: relative;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: transparent;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s ease;
}

.notification-bell:hover,
.notification-bell.active {
  background: var(--hover-bg, rgba(148, 163, 184, 0.12));
}

.bell-icon {
  font-size: 16px;
}

.unread-badge {
  position: absolute;
  top: 2px;
  right: 2px;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  border-radius: 8px;
  background: #ef4444;
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
</style>

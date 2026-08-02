<template>
  <div
    class="notification-bell-wrapper"
    ref="wrapperRef"
    @mouseenter="openPanel"
    @mouseleave="closePanel"
    @focusin="openPanel"
    @focusout="handleFocusOut"
  >
    <button
      type="button"
      class="notification-bell"
      :class="{ active: panelVisible }"
      aria-label="消息中心"
      title="消息中心"
      aria-haspopup="true"
      :aria-expanded="panelVisible"
      @click.stop="onBellClick"
    >
      <svg
        class="bell-icon"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9" />
        <path d="M10.3 21a1.94 1.94 0 0 0 3.4 0" />
      </svg>
      <span v-if="unreadCount > 0" class="unread-badge">{{ unreadCount > 99 ? '99+' : unreadCount }}</span>
    </button>
    <NotificationPanel
      :visible="panelVisible"
      @navigate="handleNavigate"
      @click.stop
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { useNotificationsStore } from '@/store/modules/notifications'
import NotificationPanel from './NotificationPanel.vue'

const store = useNotificationsStore()
const { unreadCount } = storeToRefs(store)
const router = useRouter()

const panelVisible = ref(false)
const wrapperRef = ref<HTMLElement | null>(null)

const openPanel = () => {
  panelVisible.value = true
}

const closePanel = () => {
  panelVisible.value = false
}

const onBellClick = () => {
  panelVisible.value = !panelVisible.value
}

const handleNavigate = (path: string) => {
  panelVisible.value = false
  router.push(path)
}

const handleFocusOut = (event: FocusEvent) => {
  const nextTarget = event.relatedTarget
  if (nextTarget instanceof Node && event.currentTarget instanceof Node && event.currentTarget.contains(nextTarget)) {
    return
  }
  panelVisible.value = false
}
</script>

<style scoped>
.notification-bell-wrapper {
  position: relative;
  z-index: 100;
}

.notification-bell {
  position: relative;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: transparent;
  color: var(--text-color, #1e293b);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s ease, color 0.15s ease;
}

.notification-bell:hover {
  background: var(--hover-bg, rgba(148, 163, 184, 0.24));
  color: var(--primary-color, #2196f3);
}

.notification-bell.active {
  background: rgba(33, 150, 243, 0.16);
  color: var(--primary-color, #2196f3);
}

.bell-icon {
  width: 18px;
  height: 18px;
  pointer-events: none;
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
  box-shadow: 0 0 0 2px var(--surface-color, #fff);
  pointer-events: none;
}
</style>

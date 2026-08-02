<template>
  <div v-if="visible" class="notification-panel">
    <div class="notification-panel-header">
      <strong>消息中心</strong>
      <div class="notification-actions">
        <button type="button" class="text-btn" @click="markAllAsRead">全部已读</button>
        <button type="button" class="text-btn" @click="clearAll">清空</button>
      </div>
    </div>

    <div v-if="items.length === 0" class="notification-empty">暂无消息</div>

    <ul v-else class="notification-list">
      <li
        v-for="item in items"
        :key="item.id"
        :class="['notification-item', { unread: !item.read }]"
        @click="handleClick(item)"
      >
        <div v-if="!item.read" class="notification-dot"></div>
        <div class="notification-content">
          <div class="notification-title">{{ item.title }}</div>
          <div class="notification-message">{{ item.message }}</div>
          <div class="notification-time">{{ formatTime(item.createdAt) }}</div>
        </div>
        <button
          type="button"
          class="notification-close"
          @click.stop="removeNotification(item.id)"
        >
          ×
        </button>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { useNotificationsStore } from '@/store/modules/notifications'
import type { NotificationEvent } from '@/types/notifications'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'navigate', path: string): void
}>()

const { items, markAllAsRead, markAsRead, removeNotification, clearAll } = useNotificationsStore()

const handleClick = (item: NotificationEvent) => {
  markAsRead(item.id)
  if (item.type === 'job_completed' && item.payload?.jobId) {
    emit('navigate', '/tasks')
  }
}

const formatTime = (iso: string) => {
  const date = new Date(iso)
  return date.toLocaleString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    month: 'short',
    day: 'numeric',
  })
}
</script>

<style scoped>
.notification-panel {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  width: 360px;
  max-height: 420px;
  background: var(--surface-color, #fff);
  border: 1px solid var(--border-color, rgba(148, 163, 184, 0.16));
  border-radius: 18px;
  box-shadow: 0 24px 48px rgba(44, 62, 94, 0.12);
  display: flex;
  flex-direction: column;
  z-index: 1000;
  overflow: hidden;
}

.notification-panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  border-bottom: 1px solid var(--border-color, rgba(148, 163, 184, 0.16));
  font-size: 14px;
}

.notification-actions {
  display: flex;
  gap: 12px;
}

.text-btn {
  background: none;
  border: none;
  color: var(--primary-color, #2196f3);
  font-size: 12px;
  cursor: pointer;
}

.notification-empty {
  padding: 32px 16px;
  text-align: center;
  color: var(--text-faint, #94a3b8);
  font-size: 13px;
}

.notification-list {
  list-style: none;
  margin: 0;
  padding: 8px;
  overflow-y: auto;
}

.notification-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px;
  border-radius: 14px;
  cursor: pointer;
  transition: background 0.15s ease;
}

.notification-item:hover {
  background: var(--hover-bg, rgba(148, 163, 184, 0.08));
}

.notification-item.unread {
  background: var(--unread-bg, rgba(33, 150, 243, 0.04));
}

.notification-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--primary-color, #2196f3);
  margin-top: 6px;
  flex-shrink: 0;
}

.notification-content {
  flex: 1;
  min-width: 0;
}

.notification-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-color, #1e293b);
}

.notification-message {
  margin-top: 4px;
  font-size: 12px;
  color: var(--text-soft, #64748b);
  line-height: 1.5;
  word-break: break-word;
}

.notification-time {
  margin-top: 6px;
  font-size: 11px;
  color: var(--text-faint, #94a3b8);
}

.notification-close {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  border: none;
  background: none;
  color: var(--text-faint, #94a3b8);
  font-size: 18px;
  line-height: 1;
  cursor: pointer;
  flex-shrink: 0;
}

.notification-close:hover {
  background: var(--hover-bg, rgba(148, 163, 184, 0.12));
}
</style>

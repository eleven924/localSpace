import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { EventsOn } from '../../../wailsjs/runtime/runtime'
import { isWailsAvailable } from '@/api'
import type { NotificationEvent } from '@/types/notifications'

const MAX_NOTIFICATIONS = 50

export const useNotificationsStore = defineStore('notifications', () => {
  const items = ref<NotificationEvent[]>([])

  const unreadCount = computed(() => items.value.filter((item) => !item.read).length)
  const hasUnread = computed(() => unreadCount.value > 0)

  const addNotification = (notification: NotificationEvent) => {
    items.value.unshift(notification)
    if (items.value.length > MAX_NOTIFICATIONS) {
      items.value.length = MAX_NOTIFICATIONS
    }
  }

  const markAsRead = (id: string) => {
    const item = items.value.find((n) => n.id === id)
    if (item) {
      item.read = true
    }
  }

  const markAllAsRead = () => {
    items.value.forEach((item) => {
      item.read = true
    })
  }

  const removeNotification = (id: string) => {
    items.value = items.value.filter((item) => item.id !== id)
  }

  const clearAll = () => {
    items.value = []
  }

  const initialize = () => {
    if (!isWailsAvailable()) {
      return
    }

    EventsOn('notification:new', (notification: NotificationEvent) => {
      addNotification(notification)
    })
  }

  return {
    items,
    unreadCount,
    hasUnread,
    addNotification,
    markAsRead,
    markAllAsRead,
    removeNotification,
    clearAll,
    initialize,
  }
})

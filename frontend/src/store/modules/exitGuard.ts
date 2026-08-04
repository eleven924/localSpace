import { defineStore } from 'pinia'
import { ref } from 'vue'
import { EventsOn } from '../../../wailsjs/runtime/runtime'
import { api, isWailsAvailable } from '@/api'
import type { ExitGuardSnapshot } from '@/types/jobs'

export const useExitGuardStore = defineStore('exitGuard', () => {
  const visible = ref(false)
  const snapshot = ref<ExitGuardSnapshot | null>(null)
  const loading = ref(false)
  const errorMessage = ref('')
  const initialized = ref(false)

  const open = (guardSnapshot: ExitGuardSnapshot) => {
    snapshot.value = guardSnapshot
    visible.value = guardSnapshot.hasProtectedJobs
    errorMessage.value = ''
  }

  const close = () => {
    visible.value = false
    errorMessage.value = ''
  }

  const initialize = () => {
    if (initialized.value) {
      return
    }
    initialized.value = true

    if (!isWailsAvailable()) {
      return
    }

    EventsOn('exit:guard', (guardSnapshot: ExitGuardSnapshot) => {
      open(guardSnapshot)
    })
  }

  const confirmQuit = async () => {
    loading.value = true
    errorMessage.value = ''
    try {
      await api.system.confirmQuit()
      // 确认退出后窗口会关闭；这里先收起弹层，避免在关闭动画期间残留遮罩。
      close()
    } catch (error) {
      errorMessage.value = error instanceof Error ? error.message : '退出失败'
      throw error
    } finally {
      loading.value = false
    }
  }

  return {
    visible,
    snapshot,
    loading,
    errorMessage,
    initialize,
    open,
    close,
    confirmQuit,
  }
})

<template>
  <div class="task-retention-settings">
    <div class="set-rows" aria-label="任务历史保留策略">
      <div class="set-row">
        <div class="set-row-copy">
          <label for="retention-enabled">自动清理</label>
          <p>应用启动后自动提交一次清理，只处理已完成、失败或取消的历史记录。</p>
        </div>
        <div class="set-control">
          <label class="set-switch">
            <input id="retention-enabled" v-model="config.enabled" type="checkbox" />
            <span class="set-switch-track"></span>
            <span class="set-switch-label">{{ config.enabled ? '已启用' : '未启用' }}</span>
          </label>
        </div>
      </div>

      <div class="set-row">
        <div class="set-row-copy">
          <label for="retention-count">最多保留条数</label>
          <p>超过数量后优先清理更早的终态任务记录。</p>
        </div>
        <div class="set-control inline">
          <label class="set-check">
            <input v-model="limitCountEnabled" type="checkbox" />
            <span>启用</span>
          </label>
          <input
            id="retention-count"
            v-model.number="config.maxCount"
            class="set-field num"
            type="number"
            min="1"
            :disabled="!limitCountEnabled"
            aria-label="最多保留条数"
          />
          <span class="set-unit">条</span>
        </div>
      </div>

      <div class="set-row">
        <div class="set-row-copy">
          <label for="retention-days">最多保留天数</label>
          <p>超过天数后可清理旧的终态任务记录，保留近期操作痕迹。</p>
        </div>
        <div class="set-control inline">
          <label class="set-check">
            <input v-model="limitDaysEnabled" type="checkbox" />
            <span>启用</span>
          </label>
          <input
            id="retention-days"
            v-model.number="config.maxDays"
            class="set-field num"
            type="number"
            min="1"
            :disabled="!limitDaysEnabled"
            aria-label="最多保留天数"
          />
          <span class="set-unit">天</span>
        </div>
      </div>
    </div>

    <p class="set-hint retention-note">至少启用“条数”或“天数”之一，保存后清理策略才会生效。</p>

    <div class="set-commit">
      <p
        v-if="actionMessage"
        class="set-feedback"
        :class="actionMessage.success ? 'success' : 'error'"
        :role="actionMessage.success ? 'status' : 'alert'"
      >
        {{ actionMessage.text }}
      </p>
      <button type="button" class="btn secondary" :disabled="cleaning" @click="manualCleanup">
        {{ cleaning ? '正在提交...' : '立即手动清理' }}
      </button>
      <button type="button" class="btn primary" :disabled="saving" @click="save">
        {{ saving ? '保存中...' : '保存设置' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onUnmounted, reactive, ref } from 'vue'
import { api } from '@/api'
import type { JobRetentionConfig } from '@/types'

const config = reactive<JobRetentionConfig>({ id: 1, enabled: false, maxCount: 0, maxDays: 0, updatedAt: '' })
const saving = ref(false)
const cleaning = ref(false)
const actionMessage = ref<{ success: boolean; text: string } | null>(null)
let actionMessageTimer: ReturnType<typeof window.setTimeout> | undefined

// 勾选限制项时给出一个可理解的默认值，取消勾选则用 0 表示该限制不参与清理策略。
const limitCountEnabled = computed({
  get: () => config.maxCount > 0,
  set: (v) => { config.maxCount = v ? 100 : 0 }
})
const limitDaysEnabled = computed({
  get: () => config.maxDays > 0,
  set: (v) => { config.maxDays = v ? 30 : 0 }
})

const showActionMessage = (success: boolean, text: string) => {
  // 保存反馈短暂停留在底部动作行，和导入页的提交结果是同一个位置。
  actionMessage.value = { success, text }
  if (actionMessageTimer) {
    window.clearTimeout(actionMessageTimer)
  }
  actionMessageTimer = window.setTimeout(() => {
    actionMessage.value = null
    actionMessageTimer = undefined
  }, 2600)
}

const load = async () => {
  // 页面进入时读取后端配置，避免设置页展示本地默认值造成误判。
  const loaded = await api.jobs.getRetentionConfig()
  Object.assign(config, loaded)
}

const save = async () => {
  saving.value = true
  try {
    await api.jobs.updateRetentionConfig(config)
    await load()
    showActionMessage(true, '已保存')
  } catch {
    showActionMessage(false, '保存失败，请重试')
  } finally {
    saving.value = false
  }
}

const manualCleanup = async () => {
  cleaning.value = true
  try {
    // 手动清理只提交后台任务，不在设置页里阻塞等待清理完成。
    await api.jobs.submitCleanup()
    showActionMessage(true, '已提交清理任务')
  } catch {
    showActionMessage(false, '提交清理失败')
  } finally {
    cleaning.value = false
  }
}

load()

onUnmounted(() => {
  if (actionMessageTimer) {
    window.clearTimeout(actionMessageTimer)
  }
})
</script>

<style scoped>
.task-retention-settings {
  width: 100%;
}

.retention-note {
  margin-top: 14px;
}
</style>

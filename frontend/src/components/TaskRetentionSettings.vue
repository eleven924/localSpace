<template>
  <div class="task-retention-settings">
    <div class="preference-list" aria-label="任务历史保留策略">
      <div class="preference-row">
        <div class="preference-copy">
          <strong>自动清理</strong>
          <p>应用启动后自动提交一次清理，只处理已完成、失败或取消的历史记录。</p>
        </div>
        <label class="switch-control">
          <input v-model="config.enabled" type="checkbox" />
          <span class="switch-track" aria-hidden="true">
            <span class="switch-thumb"></span>
          </span>
          <span class="control-status">{{ config.enabled ? '已启用' : '未启用' }}</span>
        </label>
      </div>

      <div class="preference-row">
        <div class="preference-copy">
          <strong>最多保留条数</strong>
          <p>超过数量后优先清理更早的终态任务记录。</p>
        </div>
        <div class="row-controls">
          <label class="check-control">
            <input v-model="limitCountEnabled" type="checkbox" />
            <span>启用</span>
          </label>
          <label class="number-control" :class="{ disabled: !limitCountEnabled }">
            <input
              v-model.number="config.maxCount"
              type="number"
              min="1"
              :disabled="!limitCountEnabled"
              aria-label="最多保留条数"
            />
            <span>条</span>
          </label>
        </div>
      </div>

      <div class="preference-row">
        <div class="preference-copy">
          <strong>最多保留天数</strong>
          <p>超过天数后可清理旧的终态任务记录，保留近期操作痕迹。</p>
        </div>
        <div class="row-controls">
          <label class="check-control">
            <input v-model="limitDaysEnabled" type="checkbox" />
            <span>启用</span>
          </label>
          <label class="number-control" :class="{ disabled: !limitDaysEnabled }">
            <input
              v-model.number="config.maxDays"
              type="number"
              min="1"
              :disabled="!limitDaysEnabled"
              aria-label="最多保留天数"
            />
            <span>天</span>
          </label>
        </div>
      </div>
    </div>

    <div class="retention-footer">
      <p class="retention-note">至少启用“条数”或“天数”之一，保存后清理策略才会生效。</p>

      <div class="config-actions">
        <button class="btn primary" :disabled="saving" @click="save">
          {{ saving ? '保存中...' : '保存设置' }}
        </button>
        <button class="btn secondary" :disabled="cleaning" @click="manualCleanup">立即手动清理</button>
        <span
          v-if="actionMessage"
          class="action-message"
          :class="actionMessage.success ? 'success' : 'error'"
          :role="actionMessage.success ? 'status' : 'alert'"
        >
          {{ actionMessage.text }}
        </span>
      </div>
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
  // 保存反馈短暂停留在操作区，既让用户确认结果，也不破坏设置页的安静布局。
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
  display: grid;
  gap: 0;
}

.preference-list {
  display: grid;
  gap: 0;
}

.preference-row {
  display: grid;
  grid-template-columns: minmax(320px, 1fr) minmax(220px, auto);
  gap: 24px;
  align-items: center;
  min-height: 74px;
  padding: 17px 0;
  border-bottom: 1px solid color-mix(in srgb, var(--border-color) 66%, transparent);
}

.preference-row:first-child {
  padding-top: 0;
}

.preference-copy {
  display: grid;
  gap: 6px;
  min-width: 0;
}

.preference-copy strong {
  display: block;
  color: var(--text-color);
  font-size: 14px;
  font-weight: 700;
}

.preference-copy p {
  max-width: 520px;
  color: color-mix(in srgb, var(--text-soft) 92%, var(--text-color));
  font-size: 13px;
  line-height: 1.55;
  margin: 0;
}

.row-controls {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  min-width: 0;
}

.check-control,
.switch-control,
.number-control {
  display: inline-flex;
  align-items: center;
  flex-shrink: 0;
}

.check-control {
  gap: 8px;
  min-width: 62px;
  color: color-mix(in srgb, var(--text-color) 88%, var(--text-soft));
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}

.task-retention-settings .check-control input {
  width: 16px;
  height: 16px;
  accent-color: var(--primary-color);
  cursor: pointer;
}

.switch-control {
  justify-content: flex-end;
  gap: 10px;
  color: color-mix(in srgb, var(--text-color) 88%, var(--text-soft));
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}

.switch-control input {
  position: absolute;
  width: 1px;
  height: 1px;
  opacity: 0;
  pointer-events: none;
}

.switch-track {
  position: relative;
  width: 38px;
  height: 21px;
  border: 1px solid color-mix(in srgb, var(--border-color) 90%, transparent);
  border-radius: 999px;
  background: color-mix(in srgb, var(--surface-muted) 86%, transparent);
  transition: background 0.16s ease, border-color 0.16s ease;
}

.switch-thumb {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 13px;
  height: 13px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--text-soft) 72%, var(--surface-color));
  box-shadow: 0 1px 2px color-mix(in srgb, var(--shadow-color) 26%, transparent);
  transition: transform 0.16s ease, background 0.16s ease;
}

.switch-control input:checked + .switch-track {
  border-color: color-mix(in srgb, var(--primary-color) 54%, transparent);
  background: color-mix(in srgb, var(--primary-color) 18%, transparent);
}

.switch-control input:checked + .switch-track .switch-thumb {
  transform: translateX(17px);
  background: var(--primary-color);
}

.switch-control input:focus-visible + .switch-track,
.task-retention-settings .check-control input:focus-visible,
.task-retention-settings .number-control input:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--primary-color) 40%, transparent);
  outline-offset: 2px;
}

.control-status {
  min-width: 42px;
  text-align: right;
}

.number-control {
  gap: 8px;
  min-height: 34px;
  padding: 0 9px;
  border: 1px solid color-mix(in srgb, var(--border-color) 88%, transparent);
  border-radius: 6px;
  background: color-mix(in srgb, var(--surface-color) 94%, transparent);
  color: var(--text-color);
  font-size: 13px;
}

.number-control.disabled {
  color: var(--text-faint);
  background: color-mix(in srgb, var(--surface-muted) 72%, transparent);
}

.task-retention-settings .number-control input {
  width: 68px;
  min-height: 28px;
  padding: 4px 2px;
  border: none;
  background: transparent;
  color: inherit;
  text-align: right;
  font-size: 14px;
}

.task-retention-settings .number-control input:focus {
  outline: none;
  box-shadow: none;
}

.task-retention-settings .number-control input:disabled {
  cursor: not-allowed;
}

.retention-footer {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
  padding-top: 15px;
}

.retention-note {
  max-width: 520px;
  margin: 0;
  color: color-mix(in srgb, var(--text-soft) 92%, var(--text-color));
  font-size: 13px;
  line-height: 1.6;
}

.config-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  justify-content: flex-end;
  margin-top: 0;
  padding-top: 0;
  border-top: none;
}

.action-message {
  min-height: 20px;
  font-size: 13px;
  line-height: 20px;
  white-space: nowrap;
}

.action-message.success {
  color: color-mix(in srgb, var(--success-color) 92%, var(--text-color));
}

.action-message.error {
  color: var(--error-color);
}

@media (max-width: 760px) {
  .preference-row {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .row-controls {
    justify-content: flex-start;
    flex-wrap: wrap;
  }

  .switch-control {
    justify-content: flex-start;
  }

  .retention-footer {
    display: grid;
  }

  .config-actions {
    justify-content: flex-start;
  }
}
</style>

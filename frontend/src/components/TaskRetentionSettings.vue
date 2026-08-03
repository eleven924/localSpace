<template>
  <div class="task-retention-settings">
    <div class="section-block">
      <label class="toggle-card">
        <input v-model="config.enabled" type="checkbox" />
        <div>
          <strong>启动时自动清理历史任务</strong>
          <p>仅删除已终态任务记录。</p>
        </div>
      </label>
    </div>

    <div class="section-block">
      <label class="toggle-card">
        <input v-model="limitCountEnabled" type="checkbox" />
        <div>
          <strong>最多保留条数</strong>
          <p v-if="limitCountEnabled">
            保留最近的 <input v-model.number="config.maxCount" type="number" min="1" /> 条
          </p>
        </div>
      </label>
    </div>

    <div class="section-block">
      <label class="toggle-card">
        <input v-model="limitDaysEnabled" type="checkbox" />
        <div>
          <strong>最多保留天数</strong>
          <p v-if="limitDaysEnabled">
            删除超过 <input v-model.number="config.maxDays" type="number" min="1" /> 天的记录
          </p>
        </div>
      </label>
    </div>

    <p class="hint">至少启用“条数”或“天数”之一，保存才会生效。</p>

    <div class="config-actions">
      <button class="btn primary" :disabled="saving" @click="save">保存设置</button>
      <button class="btn secondary" :disabled="cleaning" @click="manualCleanup">立即手动清理</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { api } from '@/api'
import type { JobRetentionConfig } from '@/types'

const config = reactive<JobRetentionConfig>({ id: 1, enabled: false, maxCount: 0, maxDays: 0, updatedAt: '' })
const saving = ref(false)
const cleaning = ref(false)

const limitCountEnabled = computed({
  get: () => config.maxCount > 0,
  set: (v) => { config.maxCount = v ? 100 : 0 }
})
const limitDaysEnabled = computed({
  get: () => config.maxDays > 0,
  set: (v) => { config.maxDays = v ? 30 : 0 }
})

const load = async () => {
  const loaded = await api.jobs.getRetentionConfig()
  Object.assign(config, loaded)
}

const save = async () => {
  saving.value = true
  try {
    await api.jobs.updateRetentionConfig(config)
    await load()
  } finally {
    saving.value = false
  }
}

const manualCleanup = async () => {
  cleaning.value = true
  try {
    await api.jobs.submitCleanup()
  } finally {
    cleaning.value = false
  }
}

load()
</script>

<style scoped>
.task-retention-settings {
  display: grid;
  gap: 24px;
}

.section-block {
  padding: 0;
  background: transparent;
  border: none;
}

.toggle-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.58);
  border: 1px solid rgba(146, 165, 192, 0.16);
  cursor: pointer;
}

.toggle-card input {
  margin-top: 4px;
}

.toggle-card strong {
  display: block;
  color: var(--text-color);
  font-size: 14px;
}

.toggle-card p {
  margin-top: 6px;
  color: var(--text-soft);
  font-size: 13px;
}

.toggle-card input[type="number"] {
  width: 80px;
  margin: 0 6px;
}

.hint {
  color: var(--text-soft);
  font-size: 12px;
}

.config-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}
</style>

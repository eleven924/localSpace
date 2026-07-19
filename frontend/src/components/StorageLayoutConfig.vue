<template>
  <div class="storage-layout-config">
    <div class="config-header">
      <h4>存储规则</h4>
      <p class="subtitle">控制新导入文件在主目录中的落盘层级结构。</p>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else class="config-content">
      <div class="form-group">
        <label for="storage-strategy">存储结构</label>
        <select id="storage-strategy" v-model="config.strategy" @change="handleChange">
          <option value="type_only">按类型</option>
          <option value="type_collection">按类型 + 合集</option>
        </select>
        <p class="form-hint">“按类型 + 合集” 会把文件放到 `类型/合集/文件名` 这一层。</p>
      </div>

      <div class="form-group">
        <label for="unsorted-folder">未填写合集时的目录名</label>
        <input
          id="unsorted-folder"
          v-model="config.unsortedFolderName"
          type="text"
          placeholder="_unsorted"
          @input="handleChange"
        />
        <p class="form-hint">当开启“按类型 + 合集”且导入时未填写合集，会使用这个目录名。</p>
      </div>

      <div class="form-group checkbox-group">
        <label class="checkbox-label" for="sanitize-folder-name">
          <span>自动清理非法目录字符</span>
        </label>
        <input
          id="sanitize-folder-name"
          v-model="config.sanitizeFolderName"
          type="checkbox"
          @change="handleChange"
        />
        <p class="form-hint">建议开启，避免 Windows 路径中出现非法字符。</p>
      </div>

      <div class="preview-card">
        <h5>路径预览</h5>
        <p v-if="config.strategy === 'type_collection'">
          `视频 / 甄嬛传 / 第01集.mkv`
        </p>
        <p v-else>
          `视频 / 第01集.mkv`
        </p>
      </div>
    </div>

    <div class="config-actions">
      <button class="btn primary" @click="handleSave" :disabled="saving">
        <span v-if="saving" class="spinner small"></span>
        <span v-else>💾</span>
        {{ saving ? '保存中...' : '保存规则' }}
      </button>
      <button v-if="hasChanges" class="btn secondary" @click="handleReset">
        重置
      </button>
    </div>

    <p v-if="message" class="message" :class="message.success ? 'success' : 'error'">
      {{ message.text }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '@/api'
import type { StorageLayoutConfig as StorageLayoutConfigModel } from '@/types'

const createDefaultConfig = (): StorageLayoutConfigModel => ({
  strategy: 'type_collection',
  unsortedFolderName: '_unsorted',
  sanitizeFolderName: true,
})

const config = ref<StorageLayoutConfigModel>(createDefaultConfig())
const originalConfig = ref<StorageLayoutConfigModel>(createDefaultConfig())
const loading = ref(false)
const saving = ref(false)
const message = ref<{ success: boolean; text: string } | null>(null)

const hasChanges = computed(() => JSON.stringify(config.value) !== JSON.stringify(originalConfig.value))

onMounted(() => {
  void loadConfig()
})

const normalizeConfig = (raw: any): StorageLayoutConfigModel => ({
  strategy: raw?.strategy === 'type_only' ? 'type_only' : 'type_collection',
  unsortedFolderName: (raw?.unsortedFolderName || '_unsorted').trim() || '_unsorted',
  sanitizeFolderName: raw?.sanitizeFolderName !== false,
})

const loadConfig = async () => {
  loading.value = true
  message.value = null

  try {
    const loaded = await api.storageLayout.getConfig()
    const normalized = normalizeConfig(loaded)
    config.value = { ...normalized }
    originalConfig.value = { ...normalized }
  } catch (error) {
    console.error('Failed to load storage layout config:', error)
    message.value = { success: false, text: '加载存储规则失败' }
  } finally {
    loading.value = false
  }
}

const handleChange = () => {
  message.value = null
}

const handleSave = async () => {
  saving.value = true
  message.value = null

  try {
    const payload = normalizeConfig(config.value)
    await api.storageLayout.updateConfig(payload)
    config.value = { ...payload }
    originalConfig.value = { ...payload }
    message.value = { success: true, text: '存储规则已保存' }
  } catch (error) {
    console.error('Failed to save storage layout config:', error)
    message.value = { success: false, text: '保存存储规则失败' }
  } finally {
    saving.value = false
  }
}

const handleReset = () => {
  config.value = { ...originalConfig.value }
  message.value = null
}
</script>

<style scoped>
.storage-layout-config {
  width: 100%;
}

.config-header {
  margin-bottom: 14px;
}

.config-header h4 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-color);
  margin: 0 0 8px 0;
}

.subtitle {
  font-size: 14px;
  color: var(--text-color);
  opacity: 0.7;
  margin: 0;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  gap: 16px;
  color: var(--text-color);
}

.spinner {
  border: 2px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  width: 20px;
  height: 20px;
  animation: spin 0.8s linear infinite;
}

.spinner.small {
  width: 16px;
  height: 16px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.config-content {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color);
}

.form-group input,
.form-group select {
  min-height: 38px;
  padding: 8px 11px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background-color: var(--bg-color);
  color: var(--text-color);
  font-size: 14px;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: var(--primary-color);
}

.checkbox-group {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 8px 12px;
  align-items: center;
}

.checkbox-label {
  margin: 0;
}

.checkbox-group input[type="checkbox"] {
  width: 18px;
  height: 18px;
  accent-color: var(--primary-color);
}

.form-hint {
  margin: 0;
  font-size: 12px;
  color: var(--text-color);
  opacity: 0.65;
  line-height: 1.5;
}

.preview-card {
  padding: 14px;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  background-color: var(--bg-color);
}

.preview-card h5 {
  margin: 0 0 8px 0;
  font-size: 14px;
  color: var(--text-color);
}

.preview-card p {
  margin: 0;
  font-size: 13px;
  color: var(--text-color);
  opacity: 0.75;
}

.config-actions {
  display: flex;
  gap: 12px;
  margin-top: 14px;
}

.btn {
  min-height: 40px;
  padding: 10px 16px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.btn.primary {
  background-color: var(--primary-color);
  color: white;
  border: none;
}

.btn.secondary {
  background-color: var(--surface-color);
  color: var(--text-color);
  border: 1px solid var(--border-color);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.message {
  margin: 14px 0 0 0;
  padding: 10px 12px;
  border-radius: 8px;
  font-size: 13px;
}

.message.success {
  background-color: rgba(76, 175, 80, 0.1);
  color: var(--success-color);
}

.message.error {
  background-color: rgba(244, 67, 54, 0.1);
  color: var(--error-color);
}

@media (max-width: 768px) {
  .config-actions {
    flex-direction: column;
  }

  .btn {
    width: 100%;
  }
}
</style>

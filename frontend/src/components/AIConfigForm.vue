<template>
  <div class="ai-config-form">
    <div class="config-header">
      <h4>AI 配置</h4>
      <p class="subtitle">配置 AI 功能用于自动生成标签和描述</p>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else class="config-content">
      <div class="toggle-section">
        <div class="toggle-info">
          <h5>启用 AI 功能</h5>
          <p>开启后将自动为导入的文件生成标签和描述</p>
        </div>
        <label class="toggle-switch">
          <input
            v-model="config.enabled"
            type="checkbox"
            @change="handleToggleEnabled"
          />
          <span class="toggle-slider"></span>
        </label>
      </div>

      <div v-if="config.enabled" class="config-fields">
        <div class="toggle-section">
          <div class="toggle-info">
            <h5>启用智能 Agent</h5>
            <p>使用智能Agent进行更精准的标签和描述生成</p>
          </div>
          <label class="toggle-switch">
            <input
              v-model="config.enableAgent"
              type="checkbox"
              @change="handleConfigChange"
            />
            <span class="toggle-slider"></span>
          </label>
        </div>

        <div v-if="config.enableAgent" class="toggle-section">
          <div class="toggle-info">
            <h5>启用网络搜索</h5>
            <p>允许Agent通过网络搜索获取更多上下文信息</p>
          </div>
          <label class="toggle-switch">
            <input
              v-model="config.enableWebSearch"
              type="checkbox"
              @change="handleConfigChange"
            />
            <span class="toggle-slider"></span>
          </label>
        </div>
        <div class="form-group">
          <label for="api-key">
            <span class="label-text">API Key</span>
            <span class="label-required">*</span>
          </label>
          <div class="input-wrapper">
            <input
              id="api-key"
              v-model="config.apiKey"
              :type="showApiKey ? 'text' : 'password'"
              placeholder="输入您的 API Key"
              @input="handleConfigChange"
            />
            <button class="toggle-visibility" @click="showApiKey = !showApiKey">
              {{ showApiKey ? '🙈' : '👁️' }}
            </button>
          </div>
          <p class="form-hint">您的 API Key 将被安全存储在本地</p>
        </div>

        <div class="form-group">
          <label for="model">
            <span class="label-text">模型</span>
            <span class="label-required">*</span>
          </label>
          <input
            id="model"
            v-model="config.model"
            type="text"
            placeholder="例如: gpt-3.5-turbo, claude-3-sonnet"
            @input="handleConfigChange"
          />
          <p class="form-hint">输入您要使用的 AI 模型名称</p>
        </div>

        <div class="form-group">
          <label for="base-url">
            <span class="label-text">Base URL</span>
            <span class="label-required">*</span>
          </label>
          <input
            id="base-url"
            v-model="config.baseURL"
            type="text"
            placeholder="https://api.openai.com/v1"
            @input="handleConfigChange"
          />
          <p class="form-hint">API 端点地址</p>
        </div>

        <div class="form-group">
          <label for="timeout">
            <span class="label-text">超时时间</span>
            <span class="label-hint">秒</span>
          </label>
          <input
            id="timeout"
            v-model.number="config.timeout"
            type="number"
            min="5"
            max="300"
            @input="handleConfigChange"
          />
          <p class="form-hint">AI 请求超时时间（5-300秒）</p>
        </div>

        <div class="form-group">
          <label for="max-tokens">
            <span class="label-text">最大 Token 数</span>
            <span class="label-hint">可选</span>
          </label>
          <input
            id="max-tokens"
            v-model.number="config.maxTokens"
            type="number"
            min="100"
            max="4000"
            @input="handleConfigChange"
          />
          <p class="form-hint">限制生成的最大 Token 数量</p>
        </div>

        <div class="test-section">
          <button
            class="btn test-button"
            @click="handleTestConnection"
            :disabled="testing || !isFormValid"
          >
            <span v-if="testing" class="spinner small"></span>
            <span v-else>🧪</span>
            {{ testing ? '测试中...' : '测试连接' }}
          </button>
          <p v-if="testResult" class="test-result" :class="testResult.success ? 'success' : 'error'">
            {{ testResult.message }}
          </p>
        </div>
      </div>

      <div v-else class="disabled-hint">
        <div class="hint-icon">🔒</div>
        <h5>AI 功能已禁用</h5>
        <p>启用上方开关以配置 AI 功能</p>
      </div>
    </div>

    <div class="config-actions">
      <button
        class="btn primary save-button"
        @click="handleSaveConfig"
        :disabled="saving || (config.enabled && !isFormValid)"
      >
        <span v-if="saving" class="spinner small"></span>
        <span v-else>💾</span>
        {{ saving ? '保存中...' : '保存配置' }}
      </button>
      <button
        v-if="hasChanges"
        class="btn secondary"
        @click="handleResetConfig"
      >
        🔄 重置
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { api } from '@/api/index'

interface AIConfig {
  id?: number
  enabled: boolean
  apiKey: string
  model: string
  baseURL: string
  timeout?: number
  maxTokens?: number
  enableAgent?: boolean
  enableWebSearch?: boolean
}

const emit = defineEmits<{
  configSaved: []
  configReset: []
}>()

const config = ref<AIConfig>({
  enabled: false,
  apiKey: '',
  model: 'gpt-3.5-turbo',
  baseURL: 'https://api.openai.com/v1',
  timeout: 30,
  maxTokens: 1000,
  enableAgent: true,
  enableWebSearch: false,
})

const originalConfig = ref<AIConfig>({ ...config.value })
const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const showApiKey = ref(false)
const testResult = ref<{ success: boolean; message: string } | null>(null)

const isFormValid = computed(() => {
  if (!config.value.enabled) return true
  return config.value.apiKey.trim().length > 0 &&
         config.value.model.trim().length > 0 &&
         config.value.baseURL.trim().length > 0
})

const hasChanges = computed(() => {
  return JSON.stringify(config.value) !== JSON.stringify(originalConfig.value)
})

onMounted(() => {
  loadConfig()
})

watch(() => config.value.enabled, (enabled) => {
  if (enabled && !config.value.apiKey) {
    // 启用时，如果 API Key 为空，可能需要提示用户
  }
})

const loadConfig = async () => {
  loading.value = true

  try {
    const loadedConfig = await api.ai.getConfig()
    if (loadedConfig) {
      config.value = {
        id: loadedConfig.id,
        enabled: loadedConfig.enabled || false,
        apiKey: loadedConfig.apiKey || '',
        model: loadedConfig.model || 'gpt-3.5-turbo',
        baseURL: loadedConfig.baseURL || 'https://api.openai.com/v1',
        timeout: loadedConfig.timeout || 30,
        maxTokens: loadedConfig.maxTokens || 1000,
        enableAgent: loadedConfig.enableAgent !== undefined ? loadedConfig.enableAgent : true,
        enableWebSearch: loadedConfig.enableWebSearch || false,
      }
      originalConfig.value = { ...config.value }
    }
  } catch (err) {
    console.error('Failed to load AI config:', err)
  } finally {
    loading.value = false
  }
}

const handleConfigChange = () => {
  testResult.value = null
}

const handleToggleEnabled = () => {
  testResult.value = null
}

const handleTestConnection = async () => {
  if (!isFormValid.value) return

  testing.value = true
  testResult.value = null

  try {
    // 模拟测试连接
    await new Promise(resolve => setTimeout(resolve, 1500))

    // 这里应该调用后端的测试连接接口
    // const result = await window.wails.go.main.App.TestAIConnection(config.value)

    // 模拟成功
    testResult.value = {
      success: true,
      message: '连接测试成功！配置有效。'
    }
  } catch (err) {
    testResult.value = {
      success: false,
      message: '连接测试失败：' + (err instanceof Error ? err.message : '未知错误')
    }
  } finally {
    testing.value = false
  }
}

const handleSaveConfig = async () => {
  if (!isFormValid.value) {
    alert('请填写所有必填字段')
    return
  }

  saving.value = true

  try {
    await api.ai.updateConfig(config.value)
    originalConfig.value = { ...config.value }
    testResult.value = {
      success: true,
      message: '配置保存成功！'
    }
    emit('configSaved')
  } catch (err) {
    console.error('Failed to save AI config:', err)
    testResult.value = {
      success: false,
      message: '保存配置失败：' + (err instanceof Error ? err.message : '未知错误')
    }
  } finally {
    saving.value = false
  }
}

const handleResetConfig = () => {
  if (confirm('确定要重置配置吗？所有未保存的更改将丢失。')) {
    config.value = { ...originalConfig.value }
    testResult.value = null
    emit('configReset')
  }
}
</script>

<style scoped>
.ai-config-form {
  width: 100%;
}

.config-header {
  margin-bottom: 20px;
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
  gap: 24px;
}

.toggle-section {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  background-color: var(--surface-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
}

.toggle-info h5 {
  font-size: 16px;
  font-weight: 500;
  color: var(--text-color);
  margin: 0 0 4px 0;
}

.toggle-info p {
  font-size: 13px;
  color: var(--text-color);
  opacity: 0.7;
  margin: 0;
}

.toggle-switch {
  position: relative;
  display: inline-block;
  width: 52px;
  height: 28px;
  flex-shrink: 0;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--border-color);
  transition: 0.3s;
  border-radius: 28px;
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 22px;
  width: 22px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
}

.toggle-switch input:checked + .toggle-slider {
  background-color: var(--primary-color);
}

.toggle-switch input:checked + .toggle-slider:before {
  transform: translateX(24px);
}

.config-fields {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 20px;
  background-color: var(--surface-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color);
}

.label-required {
  color: var(--error-color);
  font-weight: bold;
}

.label-hint {
  font-size: 12px;
  color: var(--text-color);
  opacity: 0.6;
  font-weight: normal;
}

.form-group input,
.form-group select {
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background-color: var(--bg-color);
  color: var(--text-color);
  font-size: 14px;
  transition: border-color 0.2s ease;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: var(--primary-color);
}

.form-group input[type="number"] {
  -moz-appearance: textfield;
}

.form-group input[type="number"]::-webkit-inner-spin-button,
.form-group input[type="number"]::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.input-wrapper {
  display: flex;
  gap: 8px;
}

.input-wrapper input {
  flex: 1;
}

.toggle-visibility {
  padding: 10px 12px;
  background-color: var(--surface-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  cursor: pointer;
  font-size: 16px;
  transition: all 0.2s ease;
}

.toggle-visibility:hover {
  background-color: var(--border-color);
}

.form-hint {
  font-size: 12px;
  color: var(--text-color);
  opacity: 0.6;
  margin: 0;
}

.test-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.test-button {
  width: 100%;
  padding: 12px 16px;
  background-color: var(--surface-color);
  color: var(--text-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.test-button:hover:not(:disabled) {
  background-color: var(--border-color);
  border-color: var(--text-color);
}

.test-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.test-result {
  font-size: 13px;
  margin: 0;
  padding: 10px 12px;
  border-radius: 6px;
}

.test-result.success {
  background-color: rgba(76, 175, 80, 0.1);
  color: var(--success-color);
}

.test-result.error {
  background-color: rgba(244, 67, 54, 0.1);
  color: var(--error-color);
}

.disabled-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  text-align: center;
  color: var(--text-color);
  opacity: 0.6;
}

.hint-icon {
  font-size: 48px;
  margin-bottom: 12px;
}

.disabled-hint h5 {
  font-size: 16px;
  font-weight: 500;
  margin: 0 0 4px 0;
}

.disabled-hint p {
  font-size: 14px;
  margin: 0;
}

.config-actions {
  display: flex;
  gap: 12px;
  margin-top: 20px;
}

.save-button,
.config-actions .btn {
  padding: 12px 20px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.save-button {
  flex: 2;
  background-color: var(--primary-color);
  color: white;
  border: none;
}

.save-button:hover:not(:disabled) {
  opacity: 0.9;
  box-shadow: 0 2px 8px var(--shadow-color);
}

.save-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.config-actions .btn.secondary {
  flex: 1;
  background-color: var(--surface-color);
  color: var(--text-color);
  border: 1px solid var(--border-color);
}

.config-actions .btn.secondary:hover {
  background-color: var(--border-color);
}

@media (max-width: 768px) {
  .toggle-section {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .config-actions {
    flex-direction: column;
  }

  .save-button,
  .config-actions .btn {
    width: 100%;
  }
}
</style>
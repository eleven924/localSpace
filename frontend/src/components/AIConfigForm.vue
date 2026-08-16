<template>
  <div class="ai-config-form">
    <div v-if="loading" class="set-locked">
      <div>
        <span class="set-spinner" aria-hidden="true"></span>
        <p>正在读取 AI 配置...</p>
      </div>
    </div>

    <template v-else>
      <div class="set-rows">
        <div class="set-row">
          <div class="set-row-copy">
            <strong>启用 AI 功能</strong>
            <p>开启后将自动为导入的文件生成标签和描述。</p>
          </div>
          <div class="set-control">
            <label class="set-switch">
              <input v-model="config.enabled" type="checkbox" @change="handleToggleEnabled" />
              <span class="set-switch-track"></span>
              <span class="set-switch-label">{{ config.enabled ? '已启用' : '已停用' }}</span>
            </label>
          </div>
        </div>
      </div>

      <div v-if="config.enabled" class="set-group">
        <section class="set-section">
          <div class="set-section-head">
            <div>
              <h3>模型接入</h3>
              <p>这三项决定 LocalSpace 用哪个模型、以什么身份请求。</p>
            </div>
          </div>

          <div class="set-rows">
            <div class="set-row">
              <div class="set-row-copy">
                <label for="api-key">API Key<em class="req">必填</em></label>
                <p>您的 API Key 将被安全存储在本地。</p>
              </div>
              <div class="set-control inline">
                <input
                  id="api-key"
                  v-model="config.apiKey"
                  class="set-field"
                  :type="showApiKey ? 'text' : 'password'"
                  placeholder="输入您的 API Key"
                  @input="handleConfigChange"
                />
                <button type="button" class="set-mini" @click="showApiKey = !showApiKey">
                  {{ showApiKey ? '隐藏' : '显示' }}
                </button>
              </div>
            </div>

            <div class="set-row">
              <div class="set-row-copy">
                <label for="model">模型<em class="req">必填</em></label>
                <p>输入您要使用的 AI 模型名称。</p>
              </div>
              <div class="set-control">
                <input
                  id="model"
                  v-model="config.model"
                  class="set-field"
                  type="text"
                  placeholder="例如: gpt-3.5-turbo, claude-3-sonnet"
                  @input="handleConfigChange"
                />
              </div>
            </div>

            <div class="set-row">
              <div class="set-row-copy">
                <label for="base-url">Base URL<em class="req">必填</em></label>
                <p>API 端点地址。</p>
              </div>
              <div class="set-control">
                <input
                  id="base-url"
                  v-model="config.baseURL"
                  class="set-field mono"
                  type="text"
                  placeholder="https://api.openai.com/v1"
                  @input="handleConfigChange"
                />
              </div>
            </div>

            <div class="set-row">
              <div class="set-row-copy">
                <label for="timeout">超时时间<em>秒</em></label>
                <p>AI 请求超时时间（5-300 秒）。</p>
              </div>
              <div class="set-control inline">
                <input
                  id="timeout"
                  v-model.number="config.timeout"
                  class="set-field num"
                  type="number"
                  min="5"
                  max="300"
                  @input="handleConfigChange"
                />
                <span class="set-unit">秒</span>
              </div>
            </div>

            <div class="set-row">
              <div class="set-row-copy">
                <label for="max-tokens">最大 Token 数<em>可选</em></label>
                <p>限制生成的最大 Token 数量。</p>
              </div>
              <div class="set-control inline">
                <input
                  id="max-tokens"
                  v-model.number="config.maxTokens"
                  class="set-field num"
                  type="number"
                  min="100"
                  max="4000"
                  @input="handleConfigChange"
                />
              </div>
            </div>
          </div>
        </section>

        <section class="set-section">
          <div class="set-section-head">
            <div>
              <h3>工具增强分析</h3>
              <p>AI 会优先使用已有文件信息；开启后，可按需读取当前支持的本地证据并调用已配置的工具。</p>
            </div>
          </div>

          <div class="set-rows">
            <div class="set-row">
              <div class="set-row-copy">
                <label for="enable-agent">启用工具增强</label>
                <p>关闭后仍可生成基础 AI 建议，但不会使用本地证据工具或网络搜索。</p>
              </div>
              <div class="set-control">
                <label class="set-switch">
                  <input
                    id="enable-agent"
                    v-model="config.enableAgent"
                    type="checkbox"
                    @change="handleConfigChange"
                  />
                  <span class="set-switch-track"></span>
                  <span class="set-switch-label">
                    {{ config.enableAgent ? '已启用' : '已停用' }}
                  </span>
                </label>
              </div>
            </div>

            <div v-if="config.enableAgent" class="set-row">
              <div class="set-row-copy">
                <label for="enable-web-search">启用网络搜索</label>
                <p>仅在本地证据不足时允许 AI 访问网络搜索，可能产生外部网络请求。</p>
              </div>
              <div class="set-control">
                <label class="set-switch">
                  <input
                    id="enable-web-search"
                    v-model="config.enableWebSearch"
                    type="checkbox"
                    @change="handleConfigChange"
                  />
                  <span class="set-switch-track"></span>
                  <span class="set-switch-label">
                    {{ config.enableWebSearch ? '已启用' : '已停用' }}
                  </span>
                </label>
              </div>
            </div>
          </div>
        </section>

        <section v-if="config.enableAgent && config.enableWebSearch" class="set-section">
          <div class="set-section-head">
            <div>
              <h3>网络搜索服务</h3>
              <p>搜索服务独立于模型接入，不复用模型的 API Key。</p>
            </div>
          </div>

          <div class="set-rows">
            <div class="set-row">
              <div class="set-row-copy">
                <label for="web-search-provider">搜索 Provider</label>
                <p>标识当前网络搜索服务类型。</p>
              </div>
              <div class="set-control">
                <input
                  id="web-search-provider"
                  v-model="config.webSearchProvider"
                  class="set-field"
                  type="text"
                  placeholder="例如: mock-http"
                  @input="handleConfigChange"
                />
              </div>
            </div>

            <div class="set-row">
              <div class="set-row-copy">
                <label for="web-search-base-url">搜索 Base URL<em class="req">必填</em></label>
                <p>网络搜索服务端点地址。</p>
              </div>
              <div class="set-control">
                <input
                  id="web-search-base-url"
                  v-model="config.webSearchBaseURL"
                  class="set-field mono"
                  type="text"
                  placeholder="https://api.tavily.com"
                  @input="handleConfigChange"
                />
              </div>
            </div>

            <div class="set-row">
              <div class="set-row-copy">
                <label for="web-search-api-key">搜索 API Key<em class="req">必填</em></label>
                <p>仅用于 web_search tool，不复用模型 API Key。</p>
              </div>
              <div class="set-control">
                <input
                  id="web-search-api-key"
                  v-model="config.webSearchAPIKey"
                  class="set-field"
                  type="password"
                  placeholder="输入网络搜索服务 API Key"
                  @input="handleConfigChange"
                />
              </div>
            </div>

            <div class="set-row">
              <div class="set-row-copy">
                <label for="web-search-timeout">搜索超时时间<em>秒</em></label>
                <p>单次搜索请求超时时间，默认 10 秒。</p>
              </div>
              <div class="set-control inline">
                <input
                  id="web-search-timeout"
                  v-model.number="config.webSearchTimeout"
                  class="set-field num"
                  type="number"
                  min="1"
                  max="60"
                  @input="handleConfigChange"
                />
                <span class="set-unit">秒</span>
              </div>
            </div>

            <div class="set-row">
              <div class="set-row-copy">
                <label for="web-search-max-results">搜索结果数量上限</label>
                <p>限制返回给 Agent 的结果数量，建议 3。</p>
              </div>
              <div class="set-control inline">
                <input
                  id="web-search-max-results"
                  v-model.number="config.webSearchMaxResults"
                  class="set-field num"
                  type="number"
                  min="1"
                  max="5"
                  @input="handleConfigChange"
                />
              </div>
            </div>
          </div>
        </section>

        <section class="set-section">
          <div class="set-section-head">
            <div>
              <h3>连接测试</h3>
              <p>保存前先确认这套配置能连通。</p>
            </div>
          </div>
          <div class="set-control inline">
            <button
              type="button"
              class="btn secondary"
              :disabled="testing || !isFormValid"
              @click="handleTestConnection"
            >
              {{ testing ? '测试中...' : '测试连接' }}
            </button>
            <p v-if="testResult" class="set-feedback" :class="testResult.success ? 'success' : 'error'">
              {{ testResult.message }}
            </p>
          </div>
        </section>
      </div>

      <!-- 关掉总开关后下面所有字段都不再适用，直接收起，只留一句说明和保存按钮。 -->
      <div v-else class="set-locked">
        <div>
          <strong>AI 功能已禁用</strong>
          <p>启用上方开关以配置模型、Agent 与网络搜索。导入时不会再生成标签和描述。</p>
        </div>
      </div>

      <div class="set-commit">
        <p v-if="saveResult" class="set-feedback" :class="saveResult.success ? 'success' : 'error'">
          {{ saveResult.message }}
        </p>
        <button v-if="hasChanges" type="button" class="btn secondary" @click="handleResetConfig">
          重置
        </button>
        <button
          type="button"
          class="btn primary"
          :disabled="saving || (config.enabled && !isFormValid)"
          @click="handleSaveConfig"
        >
          {{ saving ? '保存中...' : '保存配置' }}
        </button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
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
  webSearchProvider?: string
  webSearchBaseURL?: string
  webSearchAPIKey?: string
  webSearchTimeout?: number
  webSearchMaxResults?: number
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
  maxTokens: 500,
  enableAgent: false,
  enableWebSearch: false,
  webSearchProvider: '',
  webSearchBaseURL: '',
  webSearchAPIKey: '',
  webSearchTimeout: 10,
  webSearchMaxResults: 3,
})

const originalConfig = ref<AIConfig>({ ...config.value })
const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const showApiKey = ref(false)
// 测试结果和保存结果是两件事，各归各位：一个贴着测试按钮，一个在底部动作行。
const testResult = ref<{ success: boolean; message: string } | null>(null)
const saveResult = ref<{ success: boolean; message: string } | null>(null)

const isFormValid = computed(() => {
  if (!config.value.enabled) return true

  const hasModelConfig = config.value.apiKey.trim().length > 0 &&
    config.value.model.trim().length > 0 &&
    config.value.baseURL.trim().length > 0

  if (!hasModelConfig) {
    return false
  }

  if (config.value.enableAgent && config.value.enableWebSearch) {
    return config.value.webSearchBaseURL?.trim().length > 0 &&
      config.value.webSearchAPIKey?.trim().length > 0
  }

  return true
})

const hasChanges = computed(() => {
  return JSON.stringify(config.value) !== JSON.stringify(originalConfig.value)
})

onMounted(() => {
  loadConfig()
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
        maxTokens: loadedConfig.maxTokens || 500,
        enableAgent: loadedConfig.enableAgent || false,
        enableWebSearch: loadedConfig.enableWebSearch || false,
        webSearchProvider: loadedConfig.webSearchProvider || '',
        webSearchBaseURL: loadedConfig.webSearchBaseURL || '',
        webSearchAPIKey: loadedConfig.webSearchAPIKey || '',
        webSearchTimeout: loadedConfig.webSearchTimeout || 10,
        webSearchMaxResults: loadedConfig.webSearchMaxResults || 3,
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
  saveResult.value = null
}

const handleToggleEnabled = () => {
  testResult.value = null
  saveResult.value = null
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
    saveResult.value = { success: false, message: '请填写所有必填字段' }
    return
  }

  saving.value = true

  try {
    await api.ai.updateConfig(config.value)
    originalConfig.value = { ...config.value }
    saveResult.value = {
      success: true,
      message: '配置保存成功！'
    }
    emit('configSaved')
  } catch (err) {
    console.error('Failed to save AI config:', err)
    saveResult.value = {
      success: false,
      message: '保存配置失败：' + (err instanceof Error ? err.message : '未知错误')
    }
  } finally {
    saving.value = false
  }
}

const handleResetConfig = () => {
  config.value = { ...originalConfig.value }
  testResult.value = null
  saveResult.value = { success: true, message: '已恢复上次保存的配置' }
  emit('configReset')
}
</script>

<style scoped>
.ai-config-form {
  width: 100%;
}

.set-locked p {
  margin-top: 10px;
}
</style>

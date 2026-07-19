<template>
  <div class="open-with-config">
    <div class="config-header">
      <h4>打开方式</h4>
      <p class="subtitle">优先按扩展名匹配，其次按文件类型匹配；未配置时使用系统默认打开方式。</p>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else class="config-content">
      <section class="section-block">
        <div class="section-title">
          <h5>文件类型默认软件</h5>
          <p>为常见文件类型指定 LocalSpace 内优先使用的软件。</p>
        </div>

        <div
          v-for="option in fileTypeOptions"
          :key="option.value"
          class="form-group"
        >
          <label :for="`file-type-${option.value}`">
            <span class="label-main">{{ option.label }}</span>
            <span class="label-sub">{{ option.value }}</span>
          </label>

          <div class="path-editor">
            <input
              :id="`file-type-${option.value}`"
              v-model="config.byFileType[option.value]"
              type="text"
              placeholder="未配置时使用系统默认"
              @input="handleChange"
            />
            <button class="mini-btn" @click="browseForFileType(option.value)">
              浏览
            </button>
            <button
              class="mini-btn ghost"
              @click="clearFileType(option.value)"
              :disabled="!config.byFileType[option.value]"
            >
              清空
            </button>
          </div>
        </div>
      </section>

      <section class="section-block">
        <div class="section-title section-title-row">
          <div>
            <h5>扩展名覆盖规则</h5>
            <p>例如让 `.mkv` 使用专门播放器，而其他视频仍按类型规则打开。</p>
          </div>
          <button class="add-rule-btn" @click="addExtensionRule">
            添加规则
          </button>
        </div>

        <div v-if="extensionRules.length === 0" class="empty-state">
          <span class="empty-icon">🧩</span>
          <p>暂无扩展名覆盖规则</p>
        </div>

        <div
          v-for="(rule, index) in extensionRules"
          :key="rule.id"
          class="rule-card"
        >
          <div class="rule-grid">
            <div class="rule-field extension-field">
              <label :for="`extension-${rule.id}`">扩展名</label>
              <input
                :id="`extension-${rule.id}`"
                v-model="rule.extension"
                type="text"
                placeholder=".mkv"
                @input="handleExtensionChange(index)"
              />
            </div>

            <div class="rule-field path-field">
              <label :for="`app-path-${rule.id}`">软件路径</label>
              <div class="path-editor">
                <input
                  :id="`app-path-${rule.id}`"
                  v-model="rule.appPath"
                  type="text"
                  placeholder="选择可执行文件"
                  @input="handleChange"
                />
                <button class="mini-btn" @click="browseForExtension(index)">
                  浏览
                </button>
              </div>
            </div>
          </div>

          <div class="rule-actions">
            <button class="mini-btn ghost" @click="removeExtensionRule(index)">
              删除规则
            </button>
          </div>
        </div>
      </section>
    </div>

    <div class="config-actions">
      <button
        class="btn primary"
        @click="handleSave"
        :disabled="saving"
      >
        <span v-if="saving" class="spinner small"></span>
        <span v-else>💾</span>
        {{ saving ? '保存中...' : '保存配置' }}
      </button>
      <button
        v-if="hasChanges"
        class="btn secondary"
        @click="handleReset"
      >
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
import type { OpenWithConfig as OpenWithConfigModel } from '@/types'

type ExtensionRule = {
  id: number
  extension: string
  appPath: string
}

const fileTypeOptions = [
  { value: 'video', label: '视频' },
  { value: 'music', label: '音频' },
  { value: 'image', label: '图片' },
  { value: 'document', label: '文档' },
  { value: 'archive', label: '压缩包' },
  { value: 'installer', label: '安装包' },
  { value: 'other', label: '其他' },
]

const createEmptyConfig = (): OpenWithConfigModel => ({
  byFileType: {},
  byExtension: {},
})

const cloneConfig = (config: OpenWithConfigModel): OpenWithConfigModel => ({
  byFileType: { ...config.byFileType },
  byExtension: { ...config.byExtension },
})

const nextRuleId = ref(1)
const config = ref<OpenWithConfigModel>(createEmptyConfig())
const originalConfig = ref<OpenWithConfigModel>(createEmptyConfig())
const extensionRules = ref<ExtensionRule[]>([])
const loading = ref(false)
const saving = ref(false)
const message = ref<{ success: boolean; text: string } | null>(null)

const hasChanges = computed(() => {
  return JSON.stringify(buildPayload()) !== JSON.stringify(originalConfig.value)
})

onMounted(() => {
  void loadConfig()
})

const hydrateExtensionRules = (source: Record<string, string>) => {
  extensionRules.value = Object.entries(source).map(([extension, appPath]) => ({
    id: nextRuleId.value++,
    extension,
    appPath,
  }))
}

const loadConfig = async () => {
  loading.value = true
  message.value = null

  try {
    const loaded = await api.openWith.getConfig()
    const normalized = normalizeConfig(loaded)
    config.value = cloneConfig(normalized)
    originalConfig.value = cloneConfig(normalized)
    hydrateExtensionRules(normalized.byExtension)
  } catch (error) {
    console.error('Failed to load open-with config:', error)
    message.value = { success: false, text: '加载打开方式配置失败' }
  } finally {
    loading.value = false
  }
}

const normalizeExtension = (value: string): string => {
  const trimmed = value.trim().toLowerCase()
  if (!trimmed) {
    return ''
  }
  return trimmed.startsWith('.') ? trimmed : `.${trimmed}`
}

const normalizeConfig = (raw: any): OpenWithConfigModel => {
  const normalized = createEmptyConfig()

  for (const [key, value] of Object.entries(raw?.byFileType || {})) {
    const trimmed = String(value).trim()
    if (trimmed) {
      normalized.byFileType[key] = trimmed
    }
  }

  for (const [key, value] of Object.entries(raw?.byExtension || {})) {
    const extension = normalizeExtension(String(key))
    const trimmed = String(value).trim()
    if (extension && trimmed) {
      normalized.byExtension[extension] = trimmed
    }
  }

  return normalized
}

const buildPayload = (): OpenWithConfigModel => {
  const byFileType: Record<string, string> = {}
  const byExtension: Record<string, string> = {}

  for (const [key, value] of Object.entries(config.value.byFileType)) {
    const trimmed = value.trim()
    if (trimmed) {
      byFileType[key] = trimmed
    }
  }

  for (const rule of extensionRules.value) {
    const extension = normalizeExtension(rule.extension)
    const appPath = rule.appPath.trim()
    if (extension && appPath) {
      byExtension[extension] = appPath
    }
  }

  return { byFileType, byExtension }
}

const handleChange = () => {
  message.value = null
}

const handleExtensionChange = (index: number) => {
  const rule = extensionRules.value[index]
  if (!rule) {
    return
  }
  rule.extension = normalizeExtension(rule.extension)
  handleChange()
}

const browseForFileType = async (fileType: string) => {
  try {
    const selected = await api.system.selectExecutable()
    if (!selected) {
      return
    }
    config.value.byFileType[fileType] = selected
    handleChange()
  } catch (error) {
    console.error('Failed to select executable:', error)
    message.value = { success: false, text: '选择打开软件失败' }
  }
}

const browseForExtension = async (index: number) => {
  try {
    const selected = await api.system.selectExecutable()
    if (!selected) {
      return
    }
    extensionRules.value[index].appPath = selected
    handleChange()
  } catch (error) {
    console.error('Failed to select executable:', error)
    message.value = { success: false, text: '选择打开软件失败' }
  }
}

const clearFileType = (fileType: string) => {
  delete config.value.byFileType[fileType]
  config.value.byFileType = { ...config.value.byFileType }
  handleChange()
}

const addExtensionRule = () => {
  extensionRules.value.push({
    id: nextRuleId.value++,
    extension: '',
    appPath: '',
  })
  handleChange()
}

const removeExtensionRule = (index: number) => {
  extensionRules.value.splice(index, 1)
  handleChange()
}

const handleSave = async () => {
  saving.value = true
  message.value = null

  try {
    const payload = buildPayload()
    await api.openWith.updateConfig(payload)
    config.value = cloneConfig(payload)
    originalConfig.value = cloneConfig(payload)
    hydrateExtensionRules(payload.byExtension)
    message.value = { success: true, text: '打开方式配置已保存' }
  } catch (error) {
    console.error('Failed to save open-with config:', error)
    message.value = {
      success: false,
      text: '保存打开方式配置失败',
    }
  } finally {
    saving.value = false
  }
}

const handleReset = () => {
  config.value = cloneConfig(originalConfig.value)
  hydrateExtensionRules(originalConfig.value.byExtension)
  message.value = null
}
</script>

<style scoped>
.open-with-config {
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
  line-height: 1.6;
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
  gap: 20px;
}

.section-block {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.section-title h5 {
  margin: 0 0 6px 0;
  font-size: 16px;
  color: var(--text-color);
}

.section-title p {
  margin: 0;
  font-size: 13px;
  color: var(--text-color);
  opacity: 0.7;
  line-height: 1.6;
}

.section-title-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.form-group {
  display: grid;
  grid-template-columns: minmax(130px, 0.35fr) minmax(260px, 1fr);
  gap: 12px;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid var(--border-color);
}

.form-group label {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.label-main {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color);
}

.label-sub {
  font-size: 12px;
  color: var(--text-color);
  opacity: 0.6;
}

.path-editor {
  display: flex;
  gap: 8px;
}

.path-editor input,
.rule-field input {
  flex: 1;
  min-height: 38px;
  padding: 8px 11px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background-color: var(--bg-color);
  color: var(--text-color);
  font-size: 14px;
}

.path-editor input:focus,
.rule-field input:focus {
  outline: none;
  border-color: var(--primary-color);
}

.mini-btn,
.add-rule-btn,
.btn {
  min-height: 38px;
  padding: 8px 14px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.mini-btn,
.add-rule-btn,
.btn.secondary {
  background-color: var(--surface-color);
  color: var(--text-color);
  border: 1px solid var(--border-color);
}

.mini-btn:hover:not(:disabled),
.add-rule-btn:hover,
.btn.secondary:hover {
  background-color: var(--border-color);
}

.mini-btn.ghost {
  background-color: transparent;
}

.mini-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.add-rule-btn {
  white-space: nowrap;
}

.empty-state {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px;
  border: 1px dashed var(--border-color);
  border-radius: 8px;
  color: var(--text-color);
  opacity: 0.7;
}

.rule-card {
  padding: 14px;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  background-color: var(--bg-color);
}

.rule-grid {
  display: grid;
  grid-template-columns: minmax(120px, 0.32fr) minmax(240px, 0.68fr);
  gap: 12px;
}

.rule-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.rule-field label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-color);
}

.rule-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 10px;
}

.config-actions {
  display: flex;
  gap: 12px;
  margin-top: 18px;
}

.btn {
  min-width: 120px;
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

.btn.primary:hover:not(:disabled) {
  opacity: 0.92;
  box-shadow: 0 2px 8px var(--shadow-color);
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
  .form-group,
  .rule-grid {
    grid-template-columns: 1fr;
  }

  .section-title-row,
  .config-actions,
  .path-editor {
    flex-direction: column;
  }

  .add-rule-btn,
  .btn,
  .mini-btn {
    width: 100%;
  }
}
</style>

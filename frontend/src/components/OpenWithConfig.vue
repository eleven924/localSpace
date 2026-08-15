<template>
  <div class="open-with-config">
    <div v-if="loading" class="set-locked">
      <div>
        <span class="set-spinner" aria-hidden="true"></span>
        <p>正在读取打开方式配置...</p>
      </div>
    </div>

    <template v-else>
      <section class="set-section">
        <div class="set-section-head">
          <div>
            <h3>文件类型默认软件</h3>
            <p>为常见文件类型指定 LocalSpace 内优先使用的软件。未配置时使用系统默认打开方式。</p>
          </div>
        </div>

        <div class="set-rows paths">
          <div v-for="option in fileTypeOptions" :key="option.value" class="set-row">
            <div class="set-type-row">
              <span class="set-type" :class="option.value" aria-hidden="true">
                {{ shortLabel(option.value) }}
              </span>
              <div class="set-row-copy">
                <label :for="`file-type-${option.value}`">{{ option.label }}</label>
                <small>{{ option.value }}</small>
              </div>
            </div>

            <div class="set-control inline">
              <input
                :id="`file-type-${option.value}`"
                v-model="config.byFileType[option.value]"
                class="set-field mono"
                type="text"
                placeholder="未配置时使用系统默认"
                @input="handleChange"
              />
              <button type="button" class="set-mini" @click="browseForFileType(option.value)">
                浏览
              </button>
              <button
                type="button"
                class="set-mini"
                :disabled="!config.byFileType[option.value]"
                @click="clearFileType(option.value)"
              >
                清空
              </button>
            </div>
          </div>
        </div>
      </section>

      <section class="set-section">
        <div class="set-section-head">
          <div>
            <h3>扩展名覆盖规则</h3>
            <p>
              优先按扩展名匹配，其次按文件类型匹配。例如让 <code>.mkv</code>
              使用专门播放器，而其他视频仍按类型规则打开。
            </p>
          </div>
          <button type="button" class="set-mini" @click="addExtensionRule">添加规则</button>
        </div>

        <div v-if="extensionRules.length > 0" class="set-list">
          <div v-for="(rule, index) in extensionRules" :key="rule.id" class="set-rule">
            <input
              v-model="rule.extension"
              class="set-field mono"
              type="text"
              placeholder=".mkv"
              aria-label="扩展名"
              @input="handleExtensionChange(index)"
            />
            <input
              v-model="rule.appPath"
              class="set-field mono"
              type="text"
              placeholder="选择可执行文件"
              aria-label="软件路径"
              @input="handleChange"
            />
            <span class="set-control inline">
              <button type="button" class="set-mini" @click="browseForExtension(index)">浏览</button>
              <button type="button" class="set-mini warn" @click="removeExtensionRule(index)">
                删除规则
              </button>
            </span>
          </div>
        </div>

        <div v-else class="set-empty">
          <strong>暂无扩展名覆盖规则</strong>
          <p>只有某个扩展名需要区别对待时才添加，其余文件按上面的类型规则打开。</p>
        </div>
      </section>

      <div class="set-commit">
        <p v-if="message" class="set-feedback" :class="message.success ? 'success' : 'error'">
          {{ message.text }}
        </p>
        <button v-if="hasChanges" type="button" class="btn secondary" @click="handleReset">
          重置
        </button>
        <button type="button" class="btn primary" :disabled="saving" @click="handleSave">
          {{ saving ? '保存中...' : '保存配置' }}
        </button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '@/api'
import type { OpenWithConfig as OpenWithConfigModel } from '@/types'
import { FILE_TYPE_META } from '@/utils/constants'

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

const shortLabel = (fileType: string): string => FILE_TYPE_META[fileType]?.shortLabel || '其'

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

.set-locked p {
  margin-top: 10px;
}
</style>

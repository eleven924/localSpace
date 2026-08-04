<template>
  <div class="metadata-fields">
    <div class="field-block">
      <label for="file-tags">
        <span class="label-text">标签</span>
        <span class="label-hint">用逗号分隔</span>
      </label>
      <input
        id="file-tags"
        v-model="tagsStr"
        type="text"
        placeholder="工作, 重要, 报告"
      />
      <div v-if="parsedTags.length > 0" class="tags-preview">
        <span v-for="tag in parsedTags" :key="tag" class="tag-preview">
          {{ tag }}
          <button class="tag-remove" type="button" @click="removeTag(tag)">×</button>
        </span>
      </div>
    </div>

    <div class="field-block">
      <label for="file-description">
        <span class="label-text">描述</span>
        <span class="label-hint">可选</span>
      </label>
      <textarea
        id="file-description"
        v-model="description"
        rows="4"
        placeholder="输入文件的描述信息..."
      />
      <div class="char-count">
        {{ description.length }}/500
      </div>
    </div>

    <div v-if="aiEnabled" class="ai-section">
      <div class="ai-header">
        <div class="ai-title-copy">
          <span class="ai-icon">AI</span>
          <div>
            <h4>AI 分析</h4>
            <p>生成可选用的标签和描述建议。</p>
          </div>
        </div>

        <div class="ai-actions">
          <button v-if="!aiLoading" class="ai-trigger-button" type="button" @click="handleAIAnalyze">
            {{ aiAnalysis ? '重新生成' : '生成' }}
          </button>
          <button v-if="aiLoading" class="ai-trigger-button" type="button" disabled>
            分析中...
          </button>
        </div>
      </div>

      <div v-if="aiLoading" class="ai-loading">
        <div class="spinner"></div>
        <p>AI 正在分析文件...</p>
      </div>

      <div v-if="aiAnalysis" class="ai-content">
        <div class="ai-section-block">
          <h5>建议标签</h5>
          <div class="ai-tags">
            <button v-for="tag in aiAnalysis.tags" :key="tag" class="ai-tag" type="button" @click="addTag(tag)">
              {{ tag }}
            </button>
          </div>
        </div>

        <div class="ai-section-block">
          <h5>建议描述</h5>
          <div class="ai-description">{{ aiAnalysis.description }}</div>
          <div class="ai-buttons">
            <button class="apply-button" type="button" @click="applyAIDescription">
              应用描述
            </button>
            <button class="apply-all-button" type="button" @click="applyAllAI">
              应用全部
            </button>
          </div>
        </div>
      </div>

      <div v-if="!aiLoading && !aiAnalysis" class="ai-hint">
        <p>生成后会在这里显示建议内容，不会自动覆盖已填写的信息。</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted } from 'vue'
import { api } from '@/api'
import type { AIAnalysis } from '@/types'

const props = defineProps<{
  fileName: string
  fileType: string
  userKeywords?: string
  modelValueTags: string[]
  modelValueDescription: string
}>()

const emit = defineEmits<{
  'update:modelValueTags': [value: string[]]
  'update:modelValueDescription': [value: string]
}>()

const aiAnalysis = ref<AIAnalysis | null>(null)
const aiLoading = ref(false)
const aiEnabled = ref(false)

const tagsStr = computed({
  get: () => props.modelValueTags.join(', '),
  set: (value: string) => {
    const next = Array.from(
      new Set(
        value
          .split(',')
          .map(tag => tag.trim())
          .filter(Boolean)
      )
    )
    emit('update:modelValueTags', next)
  }
})

const description = computed({
  get: () => props.modelValueDescription,
  set: (value: string) => emit('update:modelValueDescription', value)
})

const parsedTags = computed(() => props.modelValueTags)

const checkAIEnabled = async () => {
  try {
    const config = await api.ai.getConfig()
    aiEnabled.value = config.enabled || false
  } catch (error) {
    console.error('Failed to check AI config:', error)
    aiEnabled.value = false
  }
}

onMounted(checkAIEnabled)
watch(() => [props.fileName, props.fileType], checkAIEnabled)

const removeTag = (tag: string) => {
  emit('update:modelValueTags', props.modelValueTags.filter(item => item !== tag))
}

const addTag = (tag: string) => {
  const normalized = tag.trim()
  if (!normalized || props.modelValueTags.some(item => item.trim() === normalized)) {
    return
  }
  emit('update:modelValueTags', [...props.modelValueTags, normalized])
}

const handleAIAnalyze = async () => {
  aiLoading.value = true
  try {
    // 使用当前表单内容作为上下文，让 AI 建议贴近用户已经填写的信息。
    aiAnalysis.value = await api.ai.analyze(
      props.fileName,
      props.fileType,
      props.userKeywords || '',
      props.modelValueTags,
      props.modelValueDescription
    )
  } catch (error) {
    console.error('AI analysis failed:', error)
  } finally {
    aiLoading.value = false
  }
}

const applyAIDescription = () => {
  if (aiAnalysis.value) {
    emit('update:modelValueDescription', aiAnalysis.value.description)
  }
}

const applyAllAI = () => {
  if (!aiAnalysis.value) return

  const merged = [...props.modelValueTags]
  aiAnalysis.value.tags.forEach(tag => {
    const normalized = tag.trim()
    if (normalized && !merged.some(item => item.trim() === normalized)) {
      merged.push(normalized)
    }
  })

  emit('update:modelValueTags', merged)
  emit('update:modelValueDescription', aiAnalysis.value.description)
}
</script>

<style scoped>
.metadata-fields {
  display: grid;
  gap: 14px;
}

.field-block {
  display: grid;
  gap: 8px;
}

.field-block label {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
}

.label-text {
  font-size: 12px;
  font-weight: 600;
  color: var(--import-text-soft, var(--text-color));
}

.label-hint {
  font-size: 12px;
  color: var(--import-text-faint, var(--text-faint));
}

.field-block input,
.field-block textarea {
  width: 100%;
  padding: 11px 12px;
  border: 1px solid color-mix(in srgb, var(--border-color) 84%, transparent);
  border-radius: 10px;
  background: color-mix(in srgb, var(--surface-color) 92%, transparent);
  color: var(--text-color);
  font-size: 14px;
  font-family: inherit;
  transition: border-color 0.2s ease, box-shadow 0.2s ease, background-color 0.2s ease;
}

.field-block input:focus,
.field-block textarea:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--primary-color) 16%, transparent);
}

.field-block textarea {
  resize: vertical;
  min-height: 96px;
}

.char-count {
  text-align: right;
  font-size: 12px;
  color: var(--import-text-faint, var(--text-faint));
}

.tags-preview {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-preview {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 10px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--primary-color) 12%, transparent);
  color: var(--primary-color);
  border: 1px solid color-mix(in srgb, var(--primary-color) 18%, transparent);
  font-size: 12px;
  font-weight: 600;
}

.tag-remove {
  background: none;
  border: none;
  color: inherit;
  cursor: pointer;
  padding: 0;
  line-height: 1;
}

.tag-remove:hover {
  opacity: 0.82;
}

.ai-section {
  display: grid;
  gap: 10px;
  padding: 12px;
  border: 1px solid color-mix(in srgb, var(--border-color) 76%, transparent);
  border-radius: 12px;
  background:
    linear-gradient(180deg, rgba(126, 176, 255, 0.05), transparent),
    color-mix(in srgb, var(--surface-color) 68%, transparent);
}

.ai-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.ai-title-copy {
  display: flex;
  gap: 10px;
  align-items: center;
  min-width: 0;
}

.ai-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 8px;
  background: rgba(126, 176, 255, 0.12);
  color: var(--import-accent-strong, var(--primary-color));
  font-size: 11px;
  font-weight: 700;
  flex-shrink: 0;
}

.ai-title-copy h4 {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-color);
  margin: 0;
  line-height: 1.3;
}

.ai-title-copy p {
  margin-top: 2px;
  color: var(--import-text-faint, var(--text-faint));
  font-size: 12px;
  line-height: 1.45;
}

.ai-actions,
.ai-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.ai-trigger-button,
.apply-button,
.apply-all-button {
  min-height: 30px;
  padding: 5px 10px;
  border-radius: 8px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.ai-trigger-button {
  background: color-mix(in srgb, var(--primary-color) 10%, transparent);
  color: var(--primary-color);
  border: 1px solid color-mix(in srgb, var(--primary-color) 22%, transparent);
}

.ai-trigger-button:hover:not(:disabled) {
  background: color-mix(in srgb, var(--primary-color) 16%, transparent);
}

.ai-trigger-button:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.ai-loading {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 0 2px 38px;
}

.ai-loading p {
  font-size: 13px;
  color: var(--import-text-soft, var(--text-color));
}

.ai-content {
  display: grid;
  gap: 14px;
  animation: fadeIn 0.24s ease;
}

.ai-section-block {
  display: grid;
  gap: 10px;
  padding-bottom: 14px;
  border-bottom: 1px solid color-mix(in srgb, var(--border-color) 76%, transparent);
}

.ai-section-block:last-child {
  padding-bottom: 0;
  border-bottom: none;
}

.ai-section-block h5 {
  font-size: 12px;
  font-weight: 700;
  color: var(--text-color);
}

.ai-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.ai-tag {
  padding: 5px 10px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--surface-muted) 78%, transparent);
  color: var(--text-color);
  border: 1px solid color-mix(in srgb, var(--border-color) 86%, transparent);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.ai-tag:hover {
  background: color-mix(in srgb, var(--primary-color) 10%, transparent);
  color: var(--primary-color);
  border-color: color-mix(in srgb, var(--primary-color) 24%, transparent);
}

.ai-description {
  padding: 12px;
  border-radius: 10px;
  background: color-mix(in srgb, var(--surface-color) 94%, transparent);
  border: 1px solid color-mix(in srgb, var(--border-color) 84%, transparent);
  color: var(--text-color);
  font-size: 13px;
  line-height: 1.7;
}

.apply-button {
  background: color-mix(in srgb, var(--surface-color) 88%, transparent);
  color: var(--text-color);
  border: 1px solid color-mix(in srgb, var(--border-color) 82%, transparent);
}

.apply-button:hover {
  background: var(--surface-muted);
}

.apply-all-button {
  background: color-mix(in srgb, var(--primary-color) 10%, transparent);
  color: var(--primary-color);
  border: 1px solid color-mix(in srgb, var(--primary-color) 22%, transparent);
  font-weight: 600;
}

.apply-all-button:hover {
  background: color-mix(in srgb, var(--primary-color) 16%, transparent);
}

.ai-hint {
  padding-left: 38px;
}

.ai-hint p {
  margin: 0;
  color: var(--import-text-faint, var(--text-faint));
  font-size: 12px;
  line-height: 1.6;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}
</style>

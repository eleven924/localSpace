<template>
  <div class="metadata-fields">
    <div class="form-group">
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
          <button class="tag-remove" @click="removeTag(tag)">✕</button>
        </span>
      </div>
    </div>

    <div class="form-group">
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
        <span class="ai-icon">🤖</span>
        <h4>AI 分析</h4>
        <div class="ai-actions">
          <button v-if="!aiLoading" class="ai-trigger-button" @click="handleAIAnalyze">
            {{ aiAnalysis ? '重新生成' : '生成' }}
          </button>
          <button v-if="aiLoading" class="ai-trigger-button" disabled>
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
            <span v-for="tag in aiAnalysis.tags" :key="tag" class="ai-tag" @click="addTag(tag)">
              {{ tag }}
            </span>
          </div>
        </div>

        <div class="ai-section-block">
          <h5>建议描述</h5>
          <div class="ai-description">{{ aiAnalysis.description }}</div>
          <div class="ai-buttons">
            <button class="apply-button" @click="applyAIDescription">
              应用此描述
            </button>
            <button class="apply-all-button" @click="applyAllAI">
              一键应用全部
            </button>
          </div>
        </div>
      </div>

      <div v-if="!aiLoading && !aiAnalysis" class="ai-hint">
        <p>点击"生成"按钮，AI 将为您分析文件并生成建议的标签和描述</p>
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
  display: contents;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.label-text {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color);
}

.label-hint {
  font-size: 12px;
  color: var(--text-color);
  opacity: 0.6;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background-color: var(--bg-color);
  color: var(--text-color);
  font-size: 14px;
  font-family: inherit;
  transition: border-color 0.2s ease;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: var(--primary-color);
}

.form-group textarea {
  resize: vertical;
  min-height: 80px;
}

.char-count {
  text-align: right;
  font-size: 12px;
  color: var(--text-color);
  opacity: 0.6;
  margin-top: 4px;
}

.tags-preview {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}

.tag-preview {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background-color: var(--primary-color);
  color: white;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.tag-remove {
  background: none;
  border: none;
  color: white;
  cursor: pointer;
  padding: 0;
  font-size: 12px;
  line-height: 1;
}

.tag-remove:hover {
  opacity: 0.8;
}

.ai-section {
  margin-top: 24px;
  padding: 20px;
  background-color: var(--surface-color);
  border: 1px solid var(--border-color);
  border-radius: 12px;
}

.ai-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}

.ai-icon {
  font-size: 20px;
}

.ai-header h4 {
  font-size: 16px;
  font-weight: 500;
  color: var(--text-color);
  margin: 0;
  flex: 1;
}

.ai-trigger-button {
  padding: 6px 12px;
  background-color: var(--primary-color);
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.ai-trigger-button:hover:not(:disabled) {
  opacity: 0.9;
}

.ai-trigger-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.ai-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 20px;
  gap: 12px;
}

.ai-loading .spinner {
  border: 2px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  width: 24px;
  height: 24px;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.ai-loading p {
  font-size: 14px;
  color: var(--text-color);
  margin: 0;
  opacity: 0.8;
}

.ai-content {
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.ai-section-block {
  margin-bottom: 16px;
}

.ai-section-block:last-child {
  margin-bottom: 0;
}

.ai-section-block h5 {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color);
  margin: 0 0 8px 0;
}

.ai-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.ai-tag {
  padding: 4px 10px;
  background-color: var(--surface-color);
  color: var(--text-color);
  border: 1px solid var(--primary-color);
  border-radius: 12px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.ai-tag:hover {
  background-color: var(--primary-color);
  color: white;
}

.ai-description {
  padding: 12px;
  background-color: var(--bg-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 13px;
  color: var(--text-color);
  line-height: 1.6;
  margin-bottom: 12px;
}

.ai-buttons {
  display: flex;
  gap: 8px;
}

.apply-button {
  padding: 6px 12px;
  background-color: var(--surface-color);
  color: var(--text-color);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.apply-button:hover {
  background-color: var(--border-color);
}

.apply-all-button {
  padding: 6px 12px;
  background-color: var(--primary-color);
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-weight: 500;
}

.apply-all-button:hover {
  opacity: 0.9;
}

.ai-hint {
  padding: 12px;
  background-color: var(--bg-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 13px;
  color: var(--text-color);
  opacity: 0.7;
}

.ai-hint p {
  margin: 0;
  text-align: center;
}
</style>

<template>
  <div class="file-meta-form">
    <div class="form-header">
      <h3>编辑文件信息</h3>
      <p class="subtitle">您可以编辑文件的名称、标签和描述</p>
    </div>

    <div class="form-content">
      <div class="form-group">
        <label for="file-name">
          <span class="label-text">文件名</span>
          <span class="label-hint">必填</span>
        </label>
        <input
          id="file-name"
          v-model="fileName"
          type="text"
          placeholder="输入文件名"
          @input="handleFileNameChange"
        />
      </div>

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
          @input="handleTagsChange"
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

      <!-- AI Analysis Section -->
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
              <span
                v-for="tag in aiAnalysis.tags"
                :key="tag"
                class="ai-tag"
                @click="addTag(tag)"
              >
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

    <div class="form-actions">
      <button class="btn secondary" @click="handleCancel">
        取消
      </button>
      <button class="btn primary" @click="handleImport" :disabled="!isValid">
        确认导入
      </button>
    </div>

    <!-- Progress Modal -->
    <div v-if="importing" class="progress-overlay">
      <div class="progress-modal">
        <div class="spinner large"></div>
        <h3>正在导入文件...</h3>
        <p>{{ progressText }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { api } from '@/api'

interface FileInfo {
  name: string
  path: string
  size: number
  type?: string
}

interface AIAnalysis {
  tags: string[]
  description: string
}

const props = defineProps<{
  file: FileInfo
  fileType: string
}>()

const emit = defineEmits<{
  imported: []
  cancel: []
}>()

const fileName = ref(props.file.name)
const tagsStr = ref('')
const description = ref('')
const parsedTags = ref<string[]>([])
const aiAnalysis = ref<AIAnalysis | null>(null)
const aiLoading = ref(false)
const importing = ref(false)
const progressText = ref('准备中...')
const aiEnabled = ref(false)
const configLoading = ref(true)

// 表单验证
const isValid = computed(() => {
  return fileName.value.trim().length > 0
})

// 检查 AI 是否启用
const checkAIEnabled = async () => {
  try {
    const config = await api.ai.getConfig()
    aiEnabled.value = config.enabled || false
  } catch (err) {
    console.error('Failed to check AI config:', err)
    aiEnabled.value = false
  } finally {
    configLoading.value = false
  }
}

onMounted(() => {
  checkAIEnabled()
})

// 监听文件变化，重置表单
watch(() => props.file, (newFile) => {
  fileName.value = newFile.name
  tagsStr.value = ''
  description.value = ''
  parsedTags.value = []
  aiAnalysis.value = null
  // 重新检查 AI 配置
  checkAIEnabled()
})

const handleFileNameChange = () => {
  // 可以添加文件名验证
}

const handleTagsChange = () => {
  const tags = tagsStr.value
    .split(',')
    .map(t => t.trim())
    .filter(t => t.length > 0)

  // 去重
  const uniqueTags = new Set(tags)
  parsedTags.value = Array.from(uniqueTags)
}

const removeTag = (tag: string) => {
  parsedTags.value = parsedTags.value.filter(t => t !== tag)
  tagsStr.value = parsedTags.value.join(', ')
}

const addTag = (tag: string) => {
  // 标准化标签：去除前后空格
  const normalizedTag = tag.trim()

  // 检查是否已存在（严格比较）
  if (normalizedTag && !parsedTags.value.some(t => t.trim() === normalizedTag)) {
    parsedTags.value.push(normalizedTag)
    tagsStr.value = parsedTags.value.join(', ')
  }
}

const handleAIAnalyze = async () => {
  aiLoading.value = true

  try {
    const analysis = await api.ai.analyze(props.file.name, props.fileType)
    aiAnalysis.value = analysis
  } catch (err) {
    console.error('AI analysis failed:', err)
    // 静默失败，不影响导入流程
  } finally {
    aiLoading.value = false
  }
}

const applyAIDescription = () => {
  if (aiAnalysis.value) {
    description.value = aiAnalysis.value.description
  }
}

const applyAllAI = () => {
  if (aiAnalysis.value) {
    // 应用描述
    description.value = aiAnalysis.value.description

    // 标准化并追加标签
    aiAnalysis.value.tags.forEach(tag => {
      // 标准化标签：去除前后空格
      const normalizedTag = tag.trim()

      // 检查是否已存在（严格比较）
      if (normalizedTag && !parsedTags.value.some(t => t.trim() === normalizedTag)) {
        parsedTags.value.push(normalizedTag)
      }
    })
    tagsStr.value = parsedTags.value.join(', ')
  }
}

const handleImport = async () => {
  if (!isValid.value) return

  importing.value = true
  progressText.value = '正在导入文件...'

  try {
    const tags = parsedTags.value.length > 0 ? parsedTags.value : undefined

    await api.file.import(
      props.file.path,
      fileName.value,
      description.value,
      tags || []
    )

    progressText.value = '导入成功！'

    // 延迟后跳转，让用户看到成功消息
    setTimeout(() => {
      emit('imported')
    }, 500)
  } catch (err) {
    console.error('Import failed:', err)
    alert('导入文件失败: ' + (err instanceof Error ? err.message : '未知错误'))
    importing.value = false
  }
}

const handleCancel = () => {
  emit('cancel')
}
</script>

<style scoped>
.file-meta-form {
  width: 100%;
  max-width: 600px;
  margin: 0 auto;
}

.form-header {
  text-align: center;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-color);
}

.form-header h3 {
  font-size: 20px;
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

.form-content {
  margin-bottom: 24px;
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

.form-actions {
  display: flex;
  gap: 12px;
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
}

.btn {
  flex: 1;
  padding: 12px 20px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn.primary {
  background-color: var(--primary-color);
  color: white;
  border: none;
}

.btn.primary:hover:not(:disabled) {
  opacity: 0.9;
  box-shadow: 0 2px 8px var(--shadow-color);
}

.btn.primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn.secondary {
  background-color: var(--surface-color);
  color: var(--text-color);
  border: 1px solid var(--border-color);
}

.btn.secondary:hover {
  background-color: var(--border-color);
}

.progress-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.progress-modal {
  background-color: var(--surface-color);
  border-radius: 12px;
  padding: 32px;
  text-align: center;
  max-width: 400px;
}

.progress-modal .spinner.large {
  border: 3px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  width: 48px;
  height: 48px;
  margin: 0 auto 16px;
  animation: spin 0.8s linear infinite;
}

.progress-modal h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-color);
  margin: 0 0 8px 0;
}

.progress-modal p {
  font-size: 14px;
  color: var(--text-color);
  opacity: 0.8;
  margin: 0;
}

@media (max-width: 768px) {
  .form-header h3 {
    font-size: 18px;
  }

  .form-actions {
    flex-direction: column-reverse;
  }

  .btn {
    width: 100%;
  }
}

@media (max-width: 480px) {
  .ai-section {
    padding: 16px;
  }

  .form-group input,
  .form-group textarea {
    padding: 8px 10px;
    font-size: 13px;
  }
}
</style>
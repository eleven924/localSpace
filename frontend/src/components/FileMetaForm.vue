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
        <label for="file-keywords">
          <span class="label-text">关键词</span>
          <span class="label-hint">可选</span>
        </label>
        <input
          id="file-keywords"
          v-model="keywords"
          type="text"
          placeholder="输入关键词，多个关键词可用逗号分隔"
        />
      </div>

      <FileMetadataFields
        :file-name="fileName"
        :file-type="fileType"
        v-model:modelValueTags="parsedTags"
        v-model:modelValueDescription="description"
      />
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
import { ref, computed, watch } from 'vue'
import { api } from '@/api'
import FileMetadataFields from './FileMetadataFields.vue'

interface FileInfo {
  name: string
  path: string
  size: number
  type?: string
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
const parsedTags = ref<string[]>([])
const description = ref('')
const keywords = ref('')
const importing = ref(false)
const progressText = ref('准备中...')

const isValid = computed(() => {
  return fileName.value.trim().length > 0
})

watch(() => props.file, (newFile) => {
  fileName.value = newFile.name
  parsedTags.value = []
  description.value = ''
  keywords.value = ''
})

const handleFileNameChange = () => {
  // 可以添加文件名验证
}

const handleImport = async () => {
  if (!isValid.value) return

  importing.value = true
  progressText.value = '正在导入文件...'

  try {
    const tags = parsedTags.value.length > 0 ? parsedTags.value : undefined

    await api.file.importWithKeywords(
      props.file.path,
      fileName.value,
      description.value,
      tags || [],
      keywords.value.trim()
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

.form-group input {
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

.form-group input:focus {
  outline: none;
  border-color: var(--primary-color);
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

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
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
</style>
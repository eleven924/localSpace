<template>
  <div class="file-selector">
    <div class="selector-header">
      <h3>选择要导入的文件</h3>
      <p class="subtitle">支持视频、文档、音乐、压缩包、安装包、图片等格式</p>
    </div>

    <button class="select-button" @click="handleSelectFile">
      <span class="button-icon">📁</span>
      <span class="button-text">选择文件</span>
    </button>

    <div v-if="selectedFile" class="selected-file-card">
      <div class="file-info">
        <div class="file-icon">{{ getFileIcon(fileType) }}</div>
        <div class="file-details">
          <h4 class="file-name">{{ selectedFile.name }}</h4>
          <div class="file-meta">
            <span class="file-type">{{ fileTypeLabel }}</span>
            <span class="separator">|</span>
            <span class="file-size">{{ formatFileSize(selectedFile.size) }}</span>
          </div>
        </div>
      </div>
      <button class="remove-button" @click="handleRemoveFile" title="移除文件">
        ✕
      </button>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-if="error" class="error-state">
      <span class="error-icon">⚠️</span>
      <p>{{ error }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { FILE_TYPES, formatFileSize, EXTENSION_TO_TYPE } from '@/utils/constants'
import { api, isWailsAvailable } from '@/api/index'

interface FileInfo {
  name: string
  path: string
  size: number
  type?: string
}

const emit = defineEmits<{
  fileSelected: [file: FileInfo]
  fileRemoved: []
}>()

const selectedFile = ref<FileInfo | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)

const fileType = ref<string>('document')
const fileTypeLabel = ref<string>('文档')

const handleSelectFile = async () => {
  error.value = null
  loading.value = true

  // Check if Wails is available
  if (!isWailsAvailable()) {
    error.value = '文件选择功能仅在桌面应用中可用'
    console.warn('Wails environment not available for file selection')
    loading.value = false
    return
  }

  try {
    const result = await api.system.selectFile()

    if (result && result !== 'success') {
      const file: FileInfo = {
        name: result.name || result.split(/[/\\]/).pop() || '',
        path: result,
        size: 0,
      }

      // 尝试获取文件大小
      try {
        console.log('Getting metadata for file:', result)
        const metadata = await api.system.getMetadata(result, 'file')
        console.log('File metadata received:', metadata)
        console.log('Metadata size field:', metadata?.size)
        console.log('Metadata size type:', typeof metadata?.size)

        if (metadata) {
          if (metadata.size !== undefined && metadata.size !== null && metadata.size > 0) {
            file.size = metadata.size
            console.log('File size set successfully:', metadata.size)
          } else {
            console.warn('File size is missing or zero in metadata:', metadata.size)
          }
        } else {
          console.warn('No metadata received from API')
        }
      } catch (err) {
        console.error('Failed to get file metadata:', err)
        console.error('Error details:', err.message, err.stack)
      }

      // 获取文件类型
      const extension = result.split('.').pop()?.toLowerCase() || ''
      console.log('File extension:', extension)

      const parsedType = await api.fileType.parse(`.${extension}`)
      console.log('Parsed file type:', parsedType)

      if (parsedType && parsedType.name) {
        fileType.value = parsedType.name
        fileTypeLabel.value = parsedType.displayName || parsedType.name
        console.log('File type recognized:', fileType.value, fileTypeLabel.value)
      } else {
        // Fallback to extension-based type mapping
        const extKey = `.${extension}`
        const fallbackType = EXTENSION_TO_TYPE[extKey] || 'document'
        const typeInfo = FILE_TYPES.find(t => t.value === fallbackType)

        fileType.value = fallbackType
        fileTypeLabel.value = typeInfo?.displayName || typeInfo?.label || fallbackType
        console.log('Failed to recognize file type, using fallback:', extKey, '->', fallbackType)
      }

      file.type = fileType.value
      selectedFile.value = file
      emit('fileSelected', file)
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '选择文件失败'
    console.error('Failed to select file:', err)
  } finally {
    loading.value = false
  }
}

const handleRemoveFile = () => {
  selectedFile.value = null
  error.value = null
  emit('fileRemoved')
}

const getFileIcon = (type: string): string => {
  const fileTypeInfo = FILE_TYPES.find(t => t.value === type)
  return fileTypeInfo?.icon || '📄'
}

// 暴露方法供父组件调用
defineExpose({
  clearFile: handleRemoveFile,
  selectFile: handleSelectFile,
})
</script>

<style scoped>
.file-selector {
  width: 100%;
  max-width: 600px;
  margin: 0 auto;
}

.selector-header {
  text-align: center;
  margin-bottom: 24px;
}

.selector-header h3 {
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

.select-button {
  width: 100%;
  padding: 40px 20px;
  border: 2px dashed var(--border-color);
  border-radius: 12px;
  background-color: var(--surface-color);
  color: var(--text-color);
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.select-button:hover {
  border-color: var(--primary-color);
  background-color: var(--bg-color);
  transform: translateY(-2px);
}

.select-button:active {
  transform: translateY(0);
}

.button-icon {
  font-size: 48px;
  opacity: 0.8;
}

.button-text {
  font-size: 14px;
}

.selected-file-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background-color: var(--surface-color);
  border: 1px solid var(--primary-color);
  border-radius: 12px;
  margin-top: 16px;
  animation: slideIn 0.3s ease;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.file-info {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.file-icon {
  font-size: 40px;
  flex-shrink: 0;
}

.file-details {
  flex: 1;
  min-width: 0;
}

.file-name {
  font-size: 16px;
  font-weight: 500;
  color: var(--text-color);
  margin: 0 0 4px 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-color);
  opacity: 0.7;
}

.separator {
  opacity: 0.5;
}

.remove-button {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background-color: var(--error-color);
  color: white;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.remove-button:hover {
  opacity: 0.9;
  transform: scale(1.1);
}

.remove-button:active {
  transform: scale(0.95);
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

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.loading-state p {
  font-size: 14px;
  margin: 0;
  opacity: 0.8;
}

.error-state {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background-color: rgba(244, 67, 54, 0.1);
  border: 1px solid var(--error-color);
  border-radius: 8px;
  margin-top: 16px;
}

.error-icon {
  font-size: 16px;
}

.error-state p {
  font-size: 14px;
  color: var(--error-color);
  margin: 0;
}

@media (max-width: 768px) {
  .selector-header h3 {
    font-size: 18px;
  }

  .subtitle {
    font-size: 13px;
  }

  .select-button {
    padding: 32px 16px;
  }

  .button-icon {
    font-size: 40px;
  }

  .selected-file-card {
    padding: 16px;
  }

  .file-icon {
    font-size: 32px;
  }

  .file-name {
    font-size: 14px;
  }

  .file-meta {
    font-size: 11px;
  }
}

@media (max-width: 480px) {
  .selected-file-card {
    flex-direction: column;
    align-items: flex-start;
  }

  .file-info {
    width: 100%;
  }

  .remove-button {
    width: 100%;
    border-radius: 8px;
  }
}
</style>

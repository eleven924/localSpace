<template>
  <div class="import-view">
    <header class="header">
      <div class="header-left">
        <router-link to="/files" class="back-link">
          <span class="back-icon">←</span>
          返回
        </router-link>
        <div class="header-info">
          <h1>导入文件</h1>
          <p class="subtitle">选择文件并编辑信息后导入到 LocalSpace</p>
        </div>
      </div>
    </header>

    <div class="content">
      <div v-if="!selectedFile" class="initial-state">
        <FileSelector
          ref="fileSelectorRef"
          @file-selected="handleFileSelected"
        />
      </div>

      <div v-else class="form-state">
        <div class="file-summary">
          <div class="file-summary-card">
            <div class="file-icon-large">{{ getFileIcon(fileType) }}</div>
            <div class="file-summary-info">
              <h3>{{ selectedFile.name }}</h3>
              <div class="file-meta">
                <span>{{ fileTypeLabel }}</span>
                <span class="separator">|</span>
                <span>{{ formatFileSize(selectedFile.size) }}</span>
              </div>
            </div>
            <button class="change-file-button" @click="handleChangeFile">
              更换文件
            </button>
          </div>
        </div>

        <FileMetaForm
          ref="fileMetaFormRef"
          :file="selectedFile"
          :file-type="fileType"
          @imported="handleImported"
          @cancel="handleCancel"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/index'
import FileSelector from '@/components/FileSelector.vue'
import FileMetaForm from '@/components/FileMetaForm.vue'
import { FILE_TYPES, formatFileSize, EXTENSION_TO_TYPE } from '@/utils/constants'

interface FileInfo {
  name: string
  path: string
  size: number
  type?: string
}

const router = useRouter()

const selectedFile = ref<FileInfo | null>(null)
const fileType = ref('document')
const fileTypeLabel = ref('文档')
const fileSelectorRef = ref<InstanceType<typeof FileSelector> | null>(null)
const fileMetaFormRef = ref<InstanceType<typeof FileMetaForm> | null>(null)

onMounted(() => {
  // 可以在这里添加初始化逻辑
})

const handleFileSelected = async (file: FileInfo) => {
  selectedFile.value = file

  // 使用从 FileSelector 传递过来的文件类型（已包含 fallback 逻辑）
  if (file.type) {
    fileType.value = file.type
    const fileTypeInfo = FILE_TYPES.find(t => t.value === file.type)
    fileTypeLabel.value = fileTypeInfo?.displayName || fileTypeInfo?.label || file.type
  } else {
    // 如果没有传递文件类型，则使用 fallback 映射
    const extension = file.path.split('.').pop()?.toLowerCase() || ''
    const extKey = `.${extension}`
    const fallbackType = EXTENSION_TO_TYPE[extKey] || 'document'
    const typeInfo = FILE_TYPES.find(t => t.value === fallbackType)

    fileType.value = fallbackType
    fileTypeLabel.value = typeInfo?.displayName || typeInfo?.label || fallbackType
  }
}

const handleChangeFile = () => {
  // 清除当前选择的文件
  selectedFile.value = null
  fileType.value = 'document'
  fileTypeLabel.value = '文档'

  // 重新聚焦文件选择器
  if (fileSelectorRef.value) {
    fileSelectorRef.value.selectFile()
  }
}

const handleImported = () => {
  router.push('/files')
}

const handleCancel = () => {
  selectedFile.value = null
  fileType.value = 'document'
  fileTypeLabel.value = '文档'
}

const getFileIcon = (type: string): string => {
  const fileTypeInfo = FILE_TYPES.find(t => t.value === type)
  return fileTypeInfo?.icon || '📄'
}
</script>

<style scoped>
.import-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: var(--bg-color);
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background-color: var(--surface-color);
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.back-link {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  color: var(--text-color);
  text-decoration: none;
  border-radius: 8px;
  transition: all 0.2s ease;
}

.back-link:hover {
  background-color: var(--border-color);
}

.back-icon {
  font-size: 16px;
}

.header-info h1 {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color);
  margin: 0 0 4px 0;
}

.subtitle {
  font-size: 14px;
  color: var(--text-color);
  opacity: 0.7;
  margin: 0;
}

.content {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.initial-state {
  display: flex;
  justify-content: center;
  padding: 24px 0;
}

.form-state {
  display: flex;
  flex-direction: column;
  gap: 24px;
  max-width: 800px;
  margin: 0 auto;
  width: 100%;
}

.file-summary {
  animation: slideDown 0.3s ease;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.file-summary-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background-color: var(--surface-color);
  border: 1px solid var(--primary-color);
  border-radius: 12px;
}

.file-icon-large {
  font-size: 48px;
  flex-shrink: 0;
}

.file-summary-info {
  flex: 1;
  min-width: 0;
}

.file-summary-info h3 {
  font-size: 18px;
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
  font-size: 13px;
  color: var(--text-color);
  opacity: 0.7;
}

.separator {
  opacity: 0.5;
}

.change-file-button {
  padding: 8px 16px;
  background-color: var(--surface-color);
  color: var(--text-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.change-file-button:hover {
  background-color: var(--border-color);
  border-color: var(--text-color);
}

@media (max-width: 768px) {
  .header {
    padding: 12px 16px;
  }

  .header-left {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .header-info h1 {
    font-size: 20px;
  }

  .content {
    padding: 16px;
  }

  .file-summary-card {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .file-summary-info {
    width: 100%;
  }

  .file-icon-large {
    font-size: 40px;
  }

  .change-file-button {
    width: 100%;
  }
}

@media (max-width: 480px) {
  .header {
    padding: 10px 12px;
  }

  .header-info h1 {
    font-size: 18px;
  }

  .content {
    padding: 12px;
  }

  .file-summary-card {
    padding: 16px;
  }

  .file-icon-large {
    font-size: 36px;
  }
}
</style>
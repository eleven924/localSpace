<template>
  <div class="import-view">
    <AppHeader />

    <div class="content">
      <section class="page-intro">
        <div class="page-intro-copy">
          <p class="eyebrow">导入中心</p>
          <h2>导入文件</h2>
          <p class="subtitle">选择文件并补充标签、简介、合集等信息后，再保存到 LocalSpace。</p>
        </div>

        <p class="intro-meta">
          {{ selectedFile ? '已选择文件，继续完善信息后即可导入。' : '支持先选文件，再整理元数据。' }}
        </p>
      </section>

      <div v-if="!selectedFile" class="initial-state">
        <FileSelector ref="fileSelectorRef" @file-selected="handleFileSelected" />
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

            <button class="change-file-button" @click="handleChangeFile">更换文件</button>
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
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import AppHeader from '@/components/AppHeader.vue'
import FileMetaForm from '@/components/FileMetaForm.vue'
import FileSelector from '@/components/FileSelector.vue'
import { EXTENSION_TO_TYPE, FILE_TYPES, formatFileSize } from '@/utils/constants'

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
  // 保留挂载点，后续如果要做最近导入记录或拖拽提示可以直接扩展。
})

const handleFileSelected = async (file: FileInfo) => {
  selectedFile.value = file

  if (file.type) {
    fileType.value = file.type
    const fileTypeInfo = FILE_TYPES.find((item) => item.value === file.type)
    fileTypeLabel.value = fileTypeInfo?.displayName || fileTypeInfo?.label || file.type
    return
  }

  const extension = file.path.split('.').pop()?.toLowerCase() || ''
  const extKey = `.${extension}`
  const fallbackType = EXTENSION_TO_TYPE[extKey] || 'document'
  const typeInfo = FILE_TYPES.find((item) => item.value === fallbackType)

  fileType.value = fallbackType
  fileTypeLabel.value = typeInfo?.displayName || typeInfo?.label || fallbackType
}

const handleChangeFile = () => {
  selectedFile.value = null
  fileType.value = 'document'
  fileTypeLabel.value = '文档'

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
  const fileTypeInfo = FILE_TYPES.find((item) => item.value === type)
  return fileTypeInfo?.icon || '📄'
}
</script>

<style scoped>
.import-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background:
    radial-gradient(circle at top left, rgba(33, 150, 243, 0.08), transparent 22%),
    var(--app-bg-color, var(--bg-color));
}

.content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px 18px 18px;
  overflow-y: auto;
}

.page-intro {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 16px;
  border-radius: 18px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background-color: color-mix(in srgb, var(--surface-color) 90%, transparent);
}

.page-intro-copy {
  min-width: 0;
}

.eyebrow {
  margin: 0 0 6px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--primary-color);
}

.page-intro h2 {
  margin: 0;
  font-size: 24px;
  line-height: 1.15;
  color: var(--text-color);
}

.subtitle {
  margin: 6px 0 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-color);
  opacity: 0.74;
}

.intro-meta {
  margin: 0;
  max-width: 260px;
  font-size: 12px;
  line-height: 1.6;
  color: var(--text-color);
  opacity: 0.66;
  text-align: right;
}

.initial-state {
  display: flex;
  justify-content: center;
  padding: 8px 0 20px;
}

.form-state {
  display: flex;
  flex-direction: column;
  gap: 20px;
  width: 100%;
  max-width: 860px;
  margin: 0 auto;
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
  padding: 18px 20px;
  background-color: color-mix(in srgb, var(--surface-color) 92%, transparent);
  border: 1px solid color-mix(in srgb, var(--primary-color) 62%, var(--border-color));
  border-radius: 16px;
}

.file-icon-large {
  flex-shrink: 0;
  font-size: 44px;
}

.file-summary-info {
  flex: 1;
  min-width: 0;
}

.file-summary-info h3 {
  margin: 0 0 4px;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-color);
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
  border-radius: 10px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.change-file-button:hover {
  background-color: var(--border-color);
  border-color: color-mix(in srgb, var(--text-color) 32%, var(--border-color));
}

@media (max-width: 768px) {
  .content {
    padding: 12px 14px 14px;
  }

  .page-intro {
    flex-direction: column;
    align-items: flex-start;
    padding: 14px;
  }

  .page-intro h2 {
    font-size: 22px;
  }

  .intro-meta {
    max-width: none;
    text-align: left;
  }

  .file-summary-card {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .file-summary-info {
    width: 100%;
  }

  .change-file-button {
    width: 100%;
  }
}

@media (max-width: 640px) {
  .content {
    padding: 12px;
  }

  .page-intro {
    border-radius: 16px;
  }

  .page-intro h2 {
    font-size: 20px;
  }

  .subtitle {
    font-size: 12px;
  }

  .file-summary-card {
    padding: 16px;
  }

  .file-icon-large {
    font-size: 40px;
  }
}
</style>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="show" class="modal-overlay" @click="handleOverlayClick">
        <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3 class="modal-title">文件详情</h3>
          <button @click="close" class="modal-close" aria-label="Close">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>

        <div class="modal-body">
          <div class="detail-item">
            <label>文件名</label>
            <span class="detail-value file-name">{{ file.fileName }}</span>
          </div>

          <div class="detail-item">
            <label>文件路径</label>
            <span class="detail-value file-path" :title="file.filePath">{{ file.filePath }}</span>
          </div>

          <div class="detail-item">
            <label>文件类型</label>
            <span class="detail-value">
              {{ file.fileType }}{{ file.fileSubType ? ` / ${file.fileSubType}` : '' }}
            </span>
          </div>

          <div class="detail-item">
            <label>文件大小</label>
            <span class="detail-value">{{ formatFileSize(file.fileSize) }}</span>
          </div>

          <div class="detail-item">
            <label>创建时间</label>
            <span class="detail-value">{{ formatDate(file.createdAt) }}</span>
          </div>

          <div class="detail-item">
            <label>修改时间</label>
            <span class="detail-value">{{ formatDate(file.modifiedAt) }}</span>
          </div>

          <div v-if="file.tags && file.tags.length > 0" class="detail-item">
            <label>标签</label>
            <div class="tags-container">
              <span
                v-for="tag in file.tags"
                :key="tag"
                class="detail-tag"
              >
                {{ tag }}
              </span>
            </div>
          </div>

          <div v-if="file.description" class="detail-item">
            <label>描述</label>
            <p class="detail-value description">{{ file.description }}</p>
          </div>

          <div class="detail-item">
            <label>文件 ID</label>
            <span class="detail-value file-id">{{ file.id }}</span>
          </div>
        </div>

        <div class="modal-footer">
          <button type="button" @click="close" class="btn btn-primary">关闭</button>
        </div>
      </div>
    </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { formatFileSize, formatDate } from '@/utils/constants'

interface File {
  id: number
  fileName: string
  filePath: string
  fileType: string
  fileSubType?: string
  fileSize: number
  tags: string[]
  description?: string
  thumbnail?: string
  createdAt: string
  modifiedAt: string
}

interface Props {
  show: boolean
  file: File
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:show': [value: boolean]
}>()

const close = () => {
  emit('update:show', false)
}

const handleOverlayClick = () => {
  close()
}

// Close dialog on Escape key
watch(() => props.show, (show) => {
  if (show) {
    document.addEventListener('keydown', handleEscape)
  } else {
    document.removeEventListener('keydown', handleEscape)
  }
})

const handleEscape = (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    close()
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 16px;
}

.modal-content {
  background-color: var(--surface-color);
  border-radius: 12px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.2);
  max-width: 500px;
  width: 100%;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-color);
}

.modal-title {
  font-size: 18px;
  font-weight: 600;
  margin: 0;
  color: var(--text-color);
}

.modal-close {
  background: none;
  border: none;
  color: var(--text-color);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s;
}

.modal-close:hover {
  background-color: var(--border-color);
}

.modal-body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}

.detail-item {
  margin-bottom: 20px;
}

.detail-item:last-child {
  margin-bottom: 0;
}

.detail-item label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-color);
  opacity: 0.7;
  margin-bottom: 6px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.detail-value {
  display: block;
  font-size: 14px;
  color: var(--text-color);
  line-height: 1.5;
  word-break: break-all;
}

.file-name {
  font-weight: 600;
  font-size: 16px;
}

.file-path {
  font-family: 'Courier New', monospace;
  font-size: 12px;
  background-color: var(--bg-color);
  padding: 8px 12px;
  border-radius: 4px;
  border: 1px solid var(--border-color);
  overflow-x: auto;
  white-space: pre-wrap;
}

.file-id {
  font-family: 'Courier New', monospace;
  font-size: 12px;
  opacity: 0.6;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.detail-tag {
  display: inline-block;
  padding: 4px 10px;
  background-color: var(--primary-color);
  color: white;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.description {
  white-space: pre-wrap;
  background-color: var(--bg-color);
  padding: 12px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid var(--border-color);
}

.btn {
  padding: 10px 20px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
}

.btn-primary {
  background-color: var(--primary-color);
  color: white;
}

.btn-primary:hover {
  opacity: 0.9;
}

/* Modal transitions */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-active .modal-content,
.modal-leave-active .modal-content {
  transition: transform 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal-content,
.modal-leave-to .modal-content {
  transform: scale(0.95) translateY(-10px);
}

/* Responsive adjustments */
@media (max-width: 480px) {
  .modal-content {
    border-radius: 8px;
    max-height: 95vh;
  }

  .modal-header,
  .modal-body,
  .modal-footer {
    padding: 16px;
  }

  .modal-title {
    font-size: 16px;
  }

  .btn {
    width: 100%;
  }
}

/* Reduced motion preference */
@media (prefers-reduced-motion: reduce) {
  .modal-enter-active,
  .modal-leave-active,
  .modal-enter-active .modal-content,
  .modal-leave-active .modal-content {
    transition: none;
  }
}
</style>
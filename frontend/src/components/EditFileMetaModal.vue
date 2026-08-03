<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="show" class="modal-overlay" @click="close">
        <div class="modal-content" @click.stop>
          <div class="modal-header">
            <h3 class="modal-title">编辑文件信息</h3>
            <button @click="close" class="modal-close" aria-label="Close">×</button>
          </div>
          <div class="modal-body">
            <div class="file-context">
              <label>当前文件</label>
              <div class="file-name">{{ file.fileName }}</div>
            </div>

            <div class="form-group">
              <label>合集</label>
              <CollectionSelector v-model="collectionId" :collections="collections" />
            </div>

            <FileMetadataFields
              :file-name="file.fileName"
              :file-type="file.fileType"
              :model-value-tags="tags"
              :model-value-description="description"
              @update:model-value-tags="tags = $event"
              @update:model-value-description="description = $event"
            />
          </div>
          <div class="modal-footer">
            <button type="button" @click="close" class="btn btn-secondary">取消</button>
            <button type="button" @click="handleSave" class="btn btn-primary" :disabled="saving">
              {{ saving ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { api } from '@/api'
import type { Collection } from '@/types'
import CollectionSelector from './CollectionSelector.vue'
import FileMetadataFields from './FileMetadataFields.vue'

const props = defineProps<{
  show: boolean
  file: {
    id: number
    fileName: string
    fileType: string
    tags: string[]
    description?: string
    collectionId?: number
  }
  collections: Collection[]
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  updated: []
}>()

const tags = ref<string[]>([])
const description = ref('')
const collectionId = ref<number | undefined>(undefined)
const saving = ref(false)

watch(
  () => [props.show, props.file.id, props.file.tags, props.file.description, props.file.collectionId] as const,
  () => {
    tags.value = [...props.file.tags]
    description.value = props.file.description || ''
    collectionId.value = props.file.collectionId
  },
  { immediate: true }
)

const close = () => emit('update:show', false)

const handleSave = async () => {
  saving.value = true
  try {
    await api.file.updateMetadata(props.file.id, tags.value, description.value, collectionId.value)
    emit('updated')
    close()
  } catch (error) {
    console.error('Failed to update file metadata:', error)
    alert(error instanceof Error ? error.message : '保存失败，请重试')
  } finally {
    saving.value = false
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
  background-color: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  padding: 24px;
}

.modal-content {
  max-width: 680px;
  width: 100%;
  background-color: var(--surface-color);
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.4), 0 8px 20px rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.1);
  animation: modalSlideIn 0.3s ease-out;
}

@keyframes modalSlideIn {
  from {
    opacity: 0;
    transform: translateY(-20px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px 28px;
  border-bottom: 1px solid var(--border-color);
  background: linear-gradient(135deg, rgba(33, 150, 243, 0.05) 0%, rgba(33, 150, 243, 0.02) 100%);
  border-radius: 16px 16px 0 0;
}

.modal-title {
  font-size: 20px;
  font-weight: 700;
  margin: 0;
  color: var(--text-color);
}

.modal-close {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: var(--text-color);
  cursor: pointer;
  padding: 8px 10px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  font-size: 18px;
  line-height: 1;
}

.modal-close:hover {
  background: rgba(255, 255, 255, 0.2);
  transform: scale(1.1);
}

.modal-close:active {
  transform: scale(0.95);
}

.modal-body {
  padding: 28px;
}

.file-context {
  margin-bottom: 20px;
}

.file-context label {
  display: block;
  font-size: 13px;
  color: var(--text-color);
  opacity: 0.7;
  margin-bottom: 8px;
}

.file-name {
  padding: 12px 14px;
  border-radius: 10px;
  background-color: var(--bg-color);
  border: 1px solid var(--border-color);
  color: var(--text-color);
  font-size: 14px;
  word-break: break-all;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 14px;
  padding: 20px 28px 28px;
  background: rgba(0, 0, 0, 0.02);
  border-top: 1px solid var(--border-color);
  border-radius: 0 0 16px 16px;
}

.btn {
  padding: 12px 24px;
  border-radius: 10px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  border: none;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background: linear-gradient(135deg, #2196F3 0%, #1976D2 100%);
  color: white;
  box-shadow: 0 4px 15px rgba(33, 150, 243, 0.4);
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(33, 150, 243, 0.5);
}

.btn-secondary {
  background: var(--bg-color);
  color: var(--text-color);
  border: 2px solid var(--border-color);
}

.btn-secondary:hover {
  background: var(--border-color);
  border-color: var(--text-color);
  transform: translateY(-2px);
}

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
</style>

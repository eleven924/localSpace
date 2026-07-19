<template>
  <div class="file-card" :class="{ 'menu-open': showMenu }" @click="handleClick">
    <!-- Menu button -->
    <div class="menu-button" @click.stop="toggleMenu">
      <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
        <circle cx="12" cy="5" r="2" />
        <circle cx="12" cy="12" r="2" />
        <circle cx="12" cy="19" r="2" />
      </svg>
    </div>

    <!-- Dropdown menu -->
    <div v-if="showMenu" class="dropdown-menu" @click.stop>
      <div class="menu-item" @click="handleShowDetails">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
          <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/>
        </svg>
        <span>文件详情</span>
      </div>
      <div class="menu-item" @click="handleOpenPreferred">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
          <path d="M8 5v14l11-7z"/>
        </svg>
        <span>用默认软件打开</span>
      </div>
      <div class="menu-item" @click="handleOpenSystemDefault">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
          <path d="M19 19H5V5h7V3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z"/>
        </svg>
        <span>用系统默认打开</span>
      </div>
      <div class="menu-item" @click="handleOpenLocation">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
          <path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/>
        </svg>
        <span>打开文件夹</span>
      </div>
      <div class="menu-item" @click="handleRename">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
          <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/>
        </svg>
        <span>重命名</span>
      </div>
      <div class="menu-item" @click="handleEditMetadata">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
          <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25z"/>
        </svg>
        <span>编辑标签和描述</span>
      </div>
      <div class="menu-item delete" @click="handleDelete">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
          <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/>
        </svg>
        <span>删除文件</span>
      </div>
    </div>

    <div class="file-thumbnail">
      <img
        v-if="showThumbnail"
        :src="thumbnailSrc"
        :alt="file.fileName"
        class="thumbnail-image"
        @error="handleThumbnailError"
      />
      <div v-else class="file-icon">
        {{ getFileIcon(file.fileType) }}
      </div>
    </div>
    <div class="file-info">
      <div
        ref="hoverZoneRef"
        class="file-info-hover-zone"
        @mouseenter="updateHoverPreviewPlacement"
      >
        <h3 class="file-name" :title="file.fileName">{{ file.fileName }}</h3>
        <div v-if="file.tags && file.tags.length > 0" class="file-tags">
          <span v-for="tag in visibleTags" :key="tag" class="tag" :title="tag">
            {{ tag }}
          </span>
          <span v-if="hiddenTags.length > 0" class="tag more more-tags-trigger">
            +{{ hiddenTags.length }}
            <span class="hidden-tags-tooltip">
              <span v-for="tag in hiddenTags" :key="`hidden-${tag}`" class="tag tooltip-tag">
                {{ tag }}
              </span>
            </span>
          </span>
        </div>
        <div
          v-if="hasHoverPreview"
          class="file-hover-preview"
          :class="previewPlacementClass"
          @click.stop
        >
          <div class="hover-preview-header">{{ file.fileName }}</div>
          <div v-if="file.collectionName" class="hover-preview-row">
            <span class="hover-preview-label">合集</span>
            <span class="hover-preview-text">{{ file.collectionName }}</span>
          </div>
          <div v-if="file.tags && file.tags.length > 0" class="hover-preview-row preview-tags-row">
            <span class="hover-preview-label">标签</span>
            <div class="hover-preview-tags">
              <span v-for="tag in file.tags" :key="`preview-${tag}`" class="hover-preview-tag">
                {{ tag }}
              </span>
            </div>
          </div>
          <div v-if="file.description" class="hover-preview-row preview-description-row">
            <span class="hover-preview-label">简介</span>
            <p class="hover-preview-description">{{ file.description }}</p>
          </div>
        </div>
      </div>
      <div class="file-meta">
        <span class="file-size">{{ formatFileSize(file.fileSize) }}</span>
        <span class="file-date">{{ formatDate(file.createdAt) }}</span>
      </div>
    </div>

    <!-- Rename dialog -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showRenameDialog" class="modal-overlay" @click="handleRenameDialogOverlayClick">
          <div class="modal-content rename-dialog" @click.stop>
            <div class="modal-header">
              <h3 class="modal-title">重命名文件</h3>
              <button @click="closeRenameDialog" class="modal-close" aria-label="Close">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="18" y1="6" x2="6" y2="18"></line>
                  <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
              </button>
            </div>
            <div class="modal-body">
              <div class="form-group">
                <label for="file-name">文件名</label>
                <input
                  id="file-name"
                  v-model="newFileName"
                  type="text"
                  class="form-input"
                  placeholder="输入新的文件名"
                  @keyup.enter="confirmRename"
                  ref="fileNameInput"
                />
                <p class="help-text">不包含扩展名，扩展名将自动保留</p>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" @click="closeRenameDialog" class="btn btn-secondary">取消</button>
              <button type="button" @click="confirmRename" class="btn btn-primary" :disabled="!newFileName || isRenaming">
                {{ isRenaming ? '重命名中...' : '重命名' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- File detail modal -->
    <FileDetailModal
      v-model:show="showDetailModal"
      :file="file"
    />

    <EditFileMetaModal
      v-model:show="showEditMetaDialog"
      :file="file"
      @updated="handleMetadataUpdated"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, nextTick, watch, onMounted, onUnmounted } from 'vue'
import { FILE_TYPES, formatFileSize, formatDate } from '@/utils/constants'
import { api } from '@/api/index'
import FileDetailModal from './FileDetailModal.vue'
import EditFileMetaModal from './EditFileMetaModal.vue'

interface File {
  id: number
  fileName: string
  collectionName?: string
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

const props = defineProps<{
  file: File
}>()

const emit = defineEmits<{
  open: [id: number]
  click: [file: File]
  delete: [id: number]
  updated: [id: number]
}>()

const thumbnailError = ref(false)
const showMenu = ref(false)
const showDetailModal = ref(false)
const showRenameDialog = ref(false)
const showEditMetaDialog = ref(false)
const newFileName = ref('')
const isRenaming = ref(false)
const fileNameInput = ref<HTMLInputElement | null>(null)
const hoverZoneRef = ref<HTMLElement | null>(null)
const previewPlacement = ref<{ horizontal: 'right' | 'left'; vertical: 'down' | 'up' }>({
  horizontal: 'right',
  vertical: 'down',
})
const visibleTags = computed(() => props.file.tags.slice(0, 3))
const hiddenTags = computed(() => props.file.tags.slice(3))
const showThumbnail = computed(() => Boolean(props.file.thumbnail) && !thumbnailError.value)
const hasHoverPreview = computed(() => {
  return Boolean(props.file.collectionName || props.file.description || props.file.tags.length > 0)
})
const previewPlacementClass = computed(() => ({
  'preview-left': previewPlacement.value.horizontal === 'left',
  'preview-right': previewPlacement.value.horizontal === 'right',
  'preview-up': previewPlacement.value.vertical === 'up',
  'preview-down': previewPlacement.value.vertical === 'down',
}))
const thumbnailSrc = computed(() => {
  return props.file.thumbnail || ''
})

const getPreviewMetrics = () => {
  const isCompactViewport = window.innerWidth <= 768
  return {
    width: Math.min(isCompactViewport ? 320 : 380, window.innerWidth - (isCompactViewport ? 32 : 56)),
    height: isCompactViewport ? 280 : 320,
    viewportPadding: isCompactViewport ? 16 : 24,
  }
}

const updateHoverPreviewPlacement = () => {
  if (!hoverZoneRef.value) {
    return
  }

  const rect = hoverZoneRef.value.getBoundingClientRect()
  const { width, height, viewportPadding } = getPreviewMetrics()

  const canOpenRight = rect.left + width <= window.innerWidth - viewportPadding
  const canOpenLeft = rect.right - width >= viewportPadding
  const canOpenDown = rect.top + height <= window.innerHeight - viewportPadding
  const canOpenUp = rect.bottom - height >= viewportPadding

  previewPlacement.value = {
    horizontal: !canOpenRight && canOpenLeft ? 'left' : 'right',
    vertical: !canOpenDown && canOpenUp ? 'up' : 'down',
  }
}

const handleClick = () => {
  emit('click', props.file)
  emit('open', props.file.id)
}

const handleThumbnailError = () => {
  thumbnailError.value = true
}

watch(
  () => props.file.thumbnail,
  () => {
    thumbnailError.value = false
  }
)

const getFileIcon = (fileType: string): string => {
  const type = FILE_TYPES.find(t => t.value === fileType)
  return type?.icon || '📄'
}

const toggleMenu = () => {
  showMenu.value = !showMenu.value
}

const handleShowDetails = () => {
  showDetailModal.value = true
  showMenu.value = false
}

const handleOpenLocation = async () => {
  try {
    await api.file.openLocation(props.file.id)
    showMenu.value = false
  } catch (error) {
    console.error('Failed to open file location:', error)
    alert('打开文件夹失败')
  }
}

const handleOpenPreferred = async () => {
  try {
    await api.file.openPreferred(props.file.id)
    showMenu.value = false
  } catch (error: any) {
    console.error('Failed to open file with preferred app:', error)
    alert(error?.message || '指定打开软件不可用，请到设置中检查路径，或使用系统默认打开。')
  }
}

const handleOpenSystemDefault = async () => {
  try {
    await api.file.openSystemDefault(props.file.id)
    showMenu.value = false
  } catch (error) {
    console.error('Failed to open file with system default:', error)
    alert('使用系统默认打开失败')
  }
}

const handleDelete = async () => {
  if (!confirm(`确定要删除文件 "${props.file.fileName}" 吗?`)) {
    return
  }

  try {
    await api.file.delete(props.file.id)
    showMenu.value = false
    emit('delete', props.file.id)
  } catch (error) {
    console.error('Failed to delete file:', error)
    alert('删除文件失败')
  }
}

const handleRename = () => {
  // Get file name without extension
  const fileName = props.file.fileName
  const lastDotIndex = fileName.lastIndexOf('.')
  const nameWithoutExt = lastDotIndex > 0 ? fileName.substring(0, lastDotIndex) : fileName

  newFileName.value = nameWithoutExt
  showRenameDialog.value = true
  showMenu.value = false

  // Focus input after dialog opens
  nextTick(() => {
    fileNameInput.value?.focus()
    fileNameInput.value?.select()
  })
}

const handleEditMetadata = () => {
  showEditMetaDialog.value = true
  showMenu.value = false
}

const handleMetadataUpdated = () => {
  emit('updated', props.file.id)
}

const closeRenameDialog = () => {
  showRenameDialog.value = false
  newFileName.value = ''
  isRenaming.value = false
}

const handleRenameDialogOverlayClick = () => {
  closeRenameDialog()
}

const confirmRename = async () => {
  if (!newFileName.value.trim()) {
    alert('请输入文件名')
    return
  }

  isRenaming.value = true

  try {
    // Get file extension
    const fileName = props.file.fileName
    const lastDotIndex = fileName.lastIndexOf('.')
    const extension = lastDotIndex > 0 ? fileName.substring(lastDotIndex) : ''

    // Construct new full file name with extension
    const newFullName = newFileName.value.trim() + extension

    await api.file.rename(props.file.id, newFullName)

    // Emit event to refresh file list
    emit('delete', props.file.id) // Reuse delete event to trigger refresh

    closeRenameDialog()
  } catch (error: any) {
    console.error('Failed to rename file:', error)
    alert(error.message || '重命名文件失败')
  } finally {
    isRenaming.value = false
  }
}

// Close menu when clicking outside
const handleClickOutside = (event: MouseEvent) => {
  const card = (event.target as HTMLElement).closest('.file-card')
  if (!card) {
    showMenu.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  window.addEventListener('resize', updateHoverPreviewPlacement)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  window.removeEventListener('resize', updateHoverPreviewPlacement)
})
</script>

<style scoped>
.file-card {
  background-color: var(--surface-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  min-width: 0;
  min-height: 0;
  position: relative;
  overflow: visible;
  z-index: 0;
}

.file-card:hover {
  border-color: var(--primary-color);
  box-shadow: 0 4px 12px var(--shadow-color);
  transform: translateY(-2px);
  z-index: 80;
}

.file-card.menu-open {
  z-index: 120;
}

.menu-button {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 32px;
  height: 32px;
  border-radius: 6px;
  background-color: var(--surface-color);
  border: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  z-index: 10;
  opacity: 0;
  transition: all 0.2s ease;
}

.file-card:hover .menu-button,
.menu-button:hover {
  opacity: 1;
}

.menu-button:hover {
  background-color: var(--bg-color);
  border-color: var(--primary-color);
}

.dropdown-menu {
  position: absolute;
  top: 48px;
  right: 8px;
  background-color: var(--surface-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  box-shadow: 0 8px 24px var(--shadow-color);
  min-width: 160px;
  z-index: 100;
  overflow: hidden;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  font-size: 14px;
  color: var(--text-color);
  cursor: pointer;
  transition: background-color 0.15s ease;
}

.menu-item:hover {
  background-color: var(--bg-color);
}

.menu-item.delete {
  color: #ef4444;
}

.menu-item.delete:hover {
  background-color: rgba(239, 68, 68, 0.1);
}

.file-thumbnail {
  width: 100%;
  min-width: 0;
  aspect-ratio: 16 / 9;
  height: auto;
  max-height: 132px;
  background-color: var(--bg-color);
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12px;
  overflow: hidden;
  flex-shrink: 0;
}

.thumbnail-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.file-icon {
  font-size: 64px;
  opacity: 0.6;
}

.file-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: visible;
}

.file-info-hover-zone {
  position: relative;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  overflow: visible;
}

.file-name {
  font-size: 14px;
  font-weight: 500;
  margin: 0;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.4;
}

.file-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  min-width: 0;
  max-height: 48px;
  overflow: hidden;
  flex-shrink: 0;
}

.tag {
  font-size: 12px;
  padding: 2px 8px;
  background-color: var(--primary-color);
  color: white;
  border-radius: 12px;
  white-space: nowrap;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tag.more {
  background-color: var(--border-color);
  color: var(--text-color);
}

.more-tags-trigger {
  position: relative;
}

.hidden-tags-tooltip {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  display: none;
  min-width: 160px;
  max-width: 240px;
  padding: 8px;
  border-radius: 8px;
  background-color: var(--surface-color);
  border: 1px solid var(--border-color);
  box-shadow: 0 8px 24px var(--shadow-color);
  gap: 6px;
  flex-wrap: wrap;
  z-index: 30;
}

.more-tags-trigger:hover .hidden-tags-tooltip {
  display: flex;
}

.tooltip-tag {
  background-color: var(--primary-color);
  color: white;
}

.file-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  min-width: 0;
  font-size: 12px;
  color: var(--text-color);
  opacity: 0.7;
  margin-top: 10px;
  flex-shrink: 0;
  padding-top: 8px;
  border-top: 1px solid rgba(148, 163, 184, 0.18);
}

.file-size,
.file-date {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-hover-preview {
  position: absolute;
  top: -8px;
  left: -6px;
  width: min(380px, calc(100vw - 56px));
  min-height: 220px;
  max-height: 320px;
  padding: 18px 18px 20px;
  border-radius: 18px;
  background:
    linear-gradient(145deg, rgba(15, 23, 42, 0.76) 0%, rgba(30, 41, 59, 0.68) 100%);
  border: 1px solid rgba(255, 255, 255, 0.14);
  backdrop-filter: blur(18px) saturate(165%);
  -webkit-backdrop-filter: blur(18px) saturate(165%);
  color: #f8fafc;
  box-shadow:
    0 30px 60px rgba(15, 23, 42, 0.28),
    0 10px 24px rgba(15, 23, 42, 0.16);
  opacity: 0;
  visibility: hidden;
  transform: translateY(8px);
  transform-origin: top left;
  transition: opacity 0.18s ease, transform 0.18s ease, visibility 0s linear 0.18s;
  pointer-events: none;
  overflow-y: auto;
  overflow-x: hidden;
  z-index: 220;
  scrollbar-width: thin;
  scrollbar-color: rgba(148, 163, 184, 0.5) transparent;
}

.file-hover-preview.preview-left {
  left: auto;
  right: -6px;
  transform-origin: top right;
}

.file-hover-preview.preview-up {
  top: auto;
  bottom: -8px;
  transform: translateY(-8px);
}

.file-hover-preview::-webkit-scrollbar {
  width: 6px;
}

.file-hover-preview::-webkit-scrollbar-thumb {
  background: rgba(148, 163, 184, 0.5);
  border-radius: 999px;
}

.file-info-hover-zone:hover .file-hover-preview {
  opacity: 1;
  visibility: visible;
  transform: translateY(0);
  transition: opacity 0.18s ease, transform 0.18s ease, visibility 0s linear 0s;
  pointer-events: auto;
}

.file-card.menu-open .file-hover-preview {
  opacity: 0;
  visibility: hidden;
  transform: translateY(8px);
  transition: opacity 0.18s ease, transform 0.18s ease, visibility 0s linear 0.18s;
  pointer-events: none;
}

.file-card.menu-open .file-hover-preview.preview-up {
  transform: translateY(-8px);
}

.hover-preview-header {
  font-size: 15px;
  font-weight: 700;
  line-height: 1.5;
  margin-bottom: 12px;
  word-break: break-word;
  color: rgba(255, 255, 255, 0.98);
}

.hover-preview-row {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 10px;
}

.hover-preview-row:last-child {
  margin-bottom: 0;
}

.hover-preview-label {
  flex-shrink: 0;
  min-width: 34px;
  font-size: 11px;
  font-weight: 600;
  line-height: 1.5;
  color: rgba(191, 219, 254, 0.88);
}

.hover-preview-text {
  font-size: 13px;
  line-height: 1.6;
  color: rgba(241, 245, 249, 0.96);
  word-break: break-word;
}

.hover-preview-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.hover-preview-tag {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 999px;
  background: rgba(59, 130, 246, 0.2);
  border: 1px solid rgba(147, 197, 253, 0.16);
  color: #e0f2fe;
  font-size: 11px;
  line-height: 1.4;
}

.preview-tags-row,
.preview-description-row {
  align-items: flex-start;
}

.hover-preview-description {
  margin: 0;
  font-size: 13px;
  line-height: 1.7;
  color: rgba(241, 245, 249, 0.92);
  white-space: pre-wrap;
  word-break: break-word;
}

@media (max-width: 768px) {
  .file-thumbnail {
    aspect-ratio: 16 / 9;
  }

  .file-icon {
    font-size: 48px;
  }

  .file-hover-preview {
    left: -4px;
    width: min(320px, calc(100vw - 32px));
    min-height: 200px;
    max-height: 280px;
    padding: 14px 14px 16px;
    border-radius: 16px;
  }

  .file-hover-preview.preview-left {
    right: -4px;
  }
}

/* Rename dialog styles */
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

.rename-dialog {
  max-width: 500px;
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
  display: flex;
  align-items: center;
  gap: 10px;
}

.modal-title::before {
  content: '✏️';
  font-size: 20px;
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
  font-size: 14px;
}

.modal-close:hover {
  background: rgba(255, 255, 255, 0.2);
  transform: scale(1.1);
}

.modal-close:active {
  transform: scale(0.95);
}

.modal-body {
  padding: 32px 28px;
}

.form-group {
  margin-bottom: 24px;
}

.form-group label {
  display: block;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-color);
  margin-bottom: 12px;
  letter-spacing: 0.3px;
}

.form-input {
  width: 100%;
  padding: 14px 16px;
  border: 2px solid var(--border-color);
  border-radius: 10px;
  font-size: 16px;
  font-weight: 500;
  background-color: var(--bg-color);
  color: var(--text-color);
  transition: all 0.3s ease;
  letter-spacing: 0.5px;
}

.form-input:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 4px rgba(33, 150, 243, 0.15);
  transform: translateY(-1px);
}

.form-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  background-color: var(--border-color);
}

.form-input::placeholder {
  color: var(--text-color);
  opacity: 0.5;
}

.help-text {
  font-size: 13px;
  color: var(--text-color);
  opacity: 0.6;
  margin: 10px 0 0 0;
  font-weight: 400;
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
  letter-spacing: 0.3px;
  position: relative;
  overflow: hidden;
}

.btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
  transition: left 0.5s ease;
}

.btn:hover::before {
  left: 100%;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none !important;
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

.btn-primary:active:not(:disabled) {
  transform: translateY(0);
  box-shadow: 0 2px 10px rgba(33, 150, 243, 0.3);
}

.btn-secondary {
  background: var(--bg-color);
  color: var(--text-color);
  border: 2px solid var(--border-color);
  font-weight: 600;
}

.btn-secondary:hover {
  background: var(--border-color);
  border-color: var(--text-color);
  transform: translateY(-2px);
}

.btn-secondary:active {
  transform: translateY(0);
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
</style>

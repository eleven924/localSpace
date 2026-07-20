<template>
  <div class="file-card" :class="{ 'menu-open': showMenu }" @click="handleClick">
    <button class="menu-button" type="button" @click.stop="toggleMenu">
      <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
        <circle cx="12" cy="5" r="1.8" />
        <circle cx="12" cy="12" r="1.8" />
        <circle cx="12" cy="19" r="1.8" />
      </svg>
    </button>

    <div v-if="showMenu" class="dropdown-menu" @click.stop>
      <button class="menu-item" type="button" @click="handleShowDetails">
        <span>查看详情</span>
      </button>
      <button class="menu-item" type="button" @click="handleOpenPreferred">
        <span>用设置的软件打开</span>
      </button>
      <button class="menu-item" type="button" @click="handleOpenSystemDefault">
        <span>用系统默认打开</span>
      </button>
      <button class="menu-item" type="button" @click="handleOpenLocation">
        <span>打开所在位置</span>
      </button>
      <button class="menu-item" type="button" @click="handleRename">
        <span>重命名</span>
      </button>
      <button class="menu-item" type="button" @click="handleEditMetadata">
        <span>编辑标签和描述</span>
      </button>
      <button class="menu-item delete" type="button" @click="handleDelete">
        <span>删除文件</span>
      </button>
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
        <div class="file-topline">
          <span v-if="file.collectionName" class="collection-pill">{{ file.collectionName }}</span>
        </div>

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
            <span class="hover-preview-label">描述</span>
            <p class="hover-preview-description">{{ file.description }}</p>
          </div>
        </div>
      </div>

      <div class="file-meta">
        <span class="file-size">{{ formatFileSize(file.fileSize) }}</span>
        <span class="file-date">{{ formatDate(file.createdAt) }}</span>
      </div>
    </div>

    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showRenameDialog" class="modal-overlay" @click="handleRenameDialogOverlayClick">
          <div class="modal-content rename-dialog" @click.stop>
            <div class="modal-header">
              <h3 class="modal-title">重命名文件</h3>
              <button type="button" class="modal-close" aria-label="Close" @click="closeRenameDialog">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
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
                  ref="fileNameInput"
                  v-model="newFileName"
                  type="text"
                  class="form-input"
                  placeholder="输入新的文件名"
                  @keyup.enter="confirmRename"
                />
                <p class="help-text">不需要带扩展名，系统会自动保留原扩展名。</p>
              </div>
            </div>

            <div class="modal-footer">
              <button type="button" class="btn secondary" @click="closeRenameDialog">取消</button>
              <button type="button" class="btn primary" :disabled="!newFileName || isRenaming" @click="confirmRename">
                {{ isRenaming ? '正在重命名...' : '确认重命名' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <FileDetailModal v-model:show="showDetailModal" :file="file" />

    <EditFileMetaModal
      v-model:show="showEditMetaDialog"
      :file="file"
      @updated="handleMetadataUpdated"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { api } from '@/api/index'
import { FILE_TYPES, formatDate, formatFileSize } from '@/utils/constants'
import EditFileMetaModal from './EditFileMetaModal.vue'
import FileDetailModal from './FileDetailModal.vue'

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
const thumbnailSrc = computed(() => props.file.thumbnail || '')

const getPreviewMetrics = () => {
  const isCompactViewport = window.innerWidth <= 768
  return {
    width: Math.min(isCompactViewport ? 320 : 380, window.innerWidth - (isCompactViewport ? 32 : 56)),
    height: isCompactViewport ? 280 : 320,
    viewportPadding: isCompactViewport ? 16 : 24,
  }
}

const updateHoverPreviewPlacement = () => {
  if (!hoverZoneRef.value) return

  const rect = hoverZoneRef.value.getBoundingClientRect()
  const { width, height, viewportPadding } = getPreviewMetrics()
  const floatingGap = 10

  const canOpenRight = rect.left + width <= window.innerWidth - viewportPadding
  const canOpenLeft = rect.right - width >= viewportPadding
  const availableBelow = window.innerHeight - rect.bottom - viewportPadding - floatingGap
  const availableAbove = rect.top - viewportPadding - floatingGap
  const canOpenDown = availableBelow >= Math.min(height, 180)
  const canOpenUp = availableAbove >= Math.min(height, 180)

  previewPlacement.value = {
    horizontal: !canOpenRight && canOpenLeft ? 'left' : 'right',
    vertical: canOpenDown || availableBelow >= availableAbove || !canOpenUp ? 'down' : 'up',
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
  const type = FILE_TYPES.find((item) => item.value === fileType)
  return type?.icon || '◌'
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
    window.alert('打开文件所在位置失败')
  }
}

const handleOpenPreferred = async () => {
  try {
    await api.file.openPreferred(props.file.id)
    showMenu.value = false
  } catch (error: any) {
    console.error('Failed to open file with preferred app:', error)
    window.alert(error?.message || '指定打开软件不可用，请到设置中检查。')
  }
}

const handleOpenSystemDefault = async () => {
  try {
    await api.file.openSystemDefault(props.file.id)
    showMenu.value = false
  } catch (error) {
    console.error('Failed to open file with system default:', error)
    window.alert('使用系统默认打开失败')
  }
}

const handleDelete = async () => {
  if (!window.confirm(`确定要删除文件“${props.file.fileName}”吗？`)) {
    return
  }

  try {
    await api.file.delete(props.file.id)
    showMenu.value = false
    emit('delete', props.file.id)
  } catch (error) {
    console.error('Failed to delete file:', error)
    window.alert('删除文件失败')
  }
}

const handleRename = () => {
  const fileName = props.file.fileName
  const lastDotIndex = fileName.lastIndexOf('.')
  const nameWithoutExt = lastDotIndex > 0 ? fileName.substring(0, lastDotIndex) : fileName

  newFileName.value = nameWithoutExt
  showRenameDialog.value = true
  showMenu.value = false

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
    window.alert('请输入文件名')
    return
  }

  isRenaming.value = true

  try {
    const fileName = props.file.fileName
    const lastDotIndex = fileName.lastIndexOf('.')
    const extension = lastDotIndex > 0 ? fileName.substring(lastDotIndex) : ''
    const newFullName = newFileName.value.trim() + extension

    await api.file.rename(props.file.id, newFullName)
    emit('delete', props.file.id)
    closeRenameDialog()
  } catch (error: any) {
    console.error('Failed to rename file:', error)
    window.alert(error.message || '重命名文件失败')
  } finally {
    isRenaming.value = false
  }
}

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
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 272px;
  padding: 12px;
  border-radius: 10px;
  border: 1px solid var(--border-color);
  background-color: var(--surface-color);
  box-shadow: 0 4px 14px var(--shadow-color);
  cursor: pointer;
  transition: all 0.2s ease;
  overflow: visible;
}

.file-card:hover {
  transform: translateY(-2px);
  border-color: var(--primary-color);
  box-shadow: 0 8px 22px var(--shadow-strong);
  z-index: 60;
}

.file-card.menu-open {
  z-index: 90;
}

.menu-button {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
  background: var(--surface-color);
  color: var(--text-faint);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.2s ease, background-color 0.2s ease, color 0.2s ease;
}

.file-card:hover .menu-button,
.file-card.menu-open .menu-button {
  opacity: 1;
}

.menu-button:hover {
  background: var(--surface-muted);
  color: var(--text-color);
}

.dropdown-menu {
  position: absolute;
  top: 40px;
  right: 8px;
  min-width: 172px;
  padding: 6px;
  border-radius: 10px;
  border: 1px solid var(--border-color);
  background: var(--surface-color);
  box-shadow: 0 10px 24px var(--shadow-strong);
  z-index: 100;
}

.menu-item {
  width: 100%;
  padding: 9px 10px;
  border-radius: 8px;
  text-align: left;
  font-size: 13px;
  color: var(--text-soft);
}

.menu-item:hover {
  background: var(--surface-muted);
  color: var(--text-color);
}

.menu-item.delete {
  color: #a85d5d;
}

.file-thumbnail {
  width: 100%;
  aspect-ratio: 16 / 10;
  border-radius: 8px;
  background: var(--surface-muted);
  border: 1px solid var(--border-color);
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.thumbnail-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.file-icon {
  font-size: 48px;
  color: var(--text-faint);
}

.file-info {
  display: flex;
  flex: 1;
  flex-direction: column;
  justify-content: space-between;
  gap: 8px;
  min-height: 0;
}

.file-info-hover-zone {
  position: relative;
  display: flex;
  flex: 0 1 auto;
  flex-direction: column;
  gap: 8px;
  min-height: 0;
  overflow: visible;
}

.file-topline {
  display: flex;
  justify-content: flex-start;
}

.collection-pill {
  display: inline-flex;
  align-items: center;
  padding: 3px 7px;
  border-radius: 999px;
  background: rgba(45, 140, 240, 0.08);
  color: var(--primary-color);
  font-size: 11px;
  font-weight: 600;
}

.file-name {
  font-size: 14px;
  line-height: 1.4;
  color: var(--text-color);
  display: -webkit-box;
  min-height: calc(1.4em * 2);
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.file-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  min-width: 0;
  max-height: 44px;
  overflow: hidden;
  align-content: flex-start;
}

.tag {
  display: inline-flex;
  align-items: center;
  min-width: 0;
  max-width: 100%;
  padding: 3px 7px;
  border-radius: 999px;
  background: var(--surface-muted);
  color: var(--text-soft);
  font-size: 11px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tag.more {
  background: #edf2f7;
}

.more-tags-trigger {
  position: relative;
}

.hidden-tags-tooltip {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  display: none;
  flex-wrap: wrap;
  gap: 6px;
  min-width: 160px;
  max-width: 240px;
  padding: 8px;
  border-radius: 10px;
  border: 1px solid var(--border-color);
  background: var(--surface-color);
  box-shadow: 0 10px 22px var(--shadow-strong);
  z-index: 40;
}

.more-tags-trigger:hover .hidden-tags-tooltip {
  display: flex;
}

.tooltip-tag {
  background: var(--surface-muted);
}

.file-meta {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  padding-top: 8px;
  border-top: 1px solid var(--border-color);
  font-size: 12px;
  color: var(--text-faint);
  flex-shrink: 0;
}

.file-hover-preview {
  position: absolute;
  top: calc(100% + 10px);
  left: 0;
  width: min(380px, calc(100vw - 56px));
  min-height: 220px;
  max-height: 320px;
  padding: 16px;
  border-radius: 12px;
  background: rgba(24, 32, 43, 0.96);
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 0 16px 32px rgba(15, 23, 42, 0.28);
  color: #eef4ff;
  opacity: 0;
  visibility: hidden;
  transform: translateY(8px);
  transition: opacity 0.18s ease, transform 0.18s ease, visibility 0s linear 0.18s;
  pointer-events: none;
  z-index: 180;
  overflow-y: auto;
}

.file-hover-preview.preview-left {
  left: auto;
  right: 0;
}

.file-hover-preview.preview-up {
  top: auto;
  bottom: calc(100% + 10px);
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
}

.hover-preview-header {
  font-size: 15px;
  font-weight: 700;
  line-height: 1.6;
  margin-bottom: 12px;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.hover-preview-row {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 10px;
}

.hover-preview-label {
  min-width: 34px;
  font-size: 11px;
  color: rgba(212, 224, 244, 0.72);
}

.hover-preview-text,
.hover-preview-description {
  font-size: 13px;
  line-height: 1.7;
  color: rgba(238, 244, 255, 0.92);
  word-break: break-word;
  overflow-wrap: anywhere;
}

.hover-preview-description {
  margin: 0;
  white-space: pre-wrap;
}

.hover-preview-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.hover-preview-tag {
  display: inline-flex;
  align-items: center;
  padding: 4px 8px;
  border-radius: 999px;
  background: rgba(77, 163, 255, 0.16);
  color: rgba(238, 244, 255, 0.92);
  font-size: 11px;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: rgba(15, 23, 42, 0.42);
  backdrop-filter: blur(10px);
}

.rename-dialog {
  width: min(520px, 100%);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  background: var(--surface-color);
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.18);
  overflow: hidden;
}

.modal-header,
.modal-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 18px 22px;
}

.modal-header {
  border-bottom: 1px solid var(--border-color);
}

.modal-title {
  font-size: 20px;
  color: var(--text-color);
}

.modal-close {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  color: var(--text-faint);
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.modal-close:hover {
  background: var(--surface-muted);
  color: var(--text-color);
}

.modal-body {
  padding: 22px;
}

.form-group label {
  display: block;
  margin-bottom: 10px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-color);
}

.help-text {
  margin-top: 10px;
  font-size: 12px;
  line-height: 1.6;
  color: var(--text-faint);
}

.modal-footer {
  justify-content: flex-end;
  border-top: 1px solid var(--border-color);
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
  transform: scale(0.96) translateY(-8px);
}

@media (max-width: 768px) {
  .file-card {
    min-height: 248px;
    padding: 10px;
  }

  .file-thumbnail {
    border-radius: 8px;
  }

  .file-hover-preview {
    width: min(320px, calc(100vw - 32px));
    left: 0;
  }

  .file-hover-preview.preview-left {
    right: 0;
  }
}
</style>

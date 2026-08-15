<template>
  <div class="file-card" :class="{ 'menu-open': showMenu, selected }" @click="handleClick">
    <label v-if="selectable" class="select-box" @click.stop>
      <input
        type="checkbox"
        :checked="selected"
        @change="emit('toggleSelect', props.file.id)"
      />
    </label>

    <button ref="menuButtonRef" class="menu-button" type="button" @click.stop="toggleMenu">
      <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
        <circle cx="12" cy="5" r="1.8" />
        <circle cx="12" cy="12" r="1.8" />
        <circle cx="12" cy="19" r="1.8" />
      </svg>
    </button>

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
        @mouseenter="handleHoverPreviewEnter"
        @mouseleave="handleHoverPreviewLeave"
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

      </div>

      <div class="file-meta">
        <span class="file-size">{{ formatFileSize(file.fileSize) }}</span>
        <span class="file-date">{{ formatDate(file.createdAt) }}</span>
      </div>
    </div>

    <Teleport to="body">
      <Transition name="hover-preview">
        <div
          v-if="isHoverPreviewVisible && hasHoverPreview && !showMenu"
          class="file-hover-preview"
          :class="previewPlacementClass"
          :style="hoverPreviewStyle"
          @mouseenter="handleHoverPreviewEnter"
          @mouseleave="handleHoverPreviewLeave"
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
      </Transition>
    </Teleport>

    <Teleport to="body">
      <Transition name="menu">
        <div
          v-if="showMenu"
          ref="menuRef"
          class="dropdown-menu"
          :style="menuStyle"
          @click.stop
        >
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
      </Transition>
    </Teleport>

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
import type { CSSProperties } from 'vue'
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
  selectable?: boolean
  selected?: boolean
}>()

const emit = defineEmits<{
  open: [id: number]
  click: [file: File]
  delete: [id: number]
  updated: [id: number]
  toggleSelect: [id: number]
}>()

const thumbnailError = ref(false)
const showMenu = ref(false)
const showDetailModal = ref(false)
const showRenameDialog = ref(false)
const showEditMetaDialog = ref(false)
const newFileName = ref('')
const isRenaming = ref(false)
const fileNameInput = ref<HTMLInputElement | null>(null)
const menuButtonRef = ref<HTMLButtonElement | null>(null)
const menuRef = ref<HTMLElement | null>(null)
const menuStyle = ref<CSSProperties>({})
const hoverZoneRef = ref<HTMLElement | null>(null)
const isHoverPreviewVisible = ref(false)
const previewPlacement = ref<{ horizontal: 'right' | 'left'; vertical: 'down' | 'up' }>({
  horizontal: 'right',
  vertical: 'down',
})
const hoverPreviewStyle = ref<CSSProperties>({})
let hoverPreviewCloseTimer: ReturnType<typeof window.setTimeout> | null = null

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

const updateMenuPlacement = () => {
  if (!showMenu.value || !menuButtonRef.value || !menuRef.value) return

  const buttonRect = menuButtonRef.value.getBoundingClientRect()
  const menuRect = menuRef.value.getBoundingClientRect()
  const viewportPadding = 12
  const menuWidth = Math.min(menuRect.width, window.innerWidth - viewportPadding * 2)
  const menuHeight = menuRect.height
  const gap = 8
  const canOpenBelow = buttonRect.bottom + gap + menuHeight <= window.innerHeight - viewportPadding
  const rawTop = canOpenBelow
    ? buttonRect.bottom + gap
    : buttonRect.top - menuHeight - gap
  const rawLeft = buttonRect.right - menuWidth
  const clamp = (value: number, min: number, max: number) => Math.min(Math.max(value, min), max)

  // 菜单挂到 body 后使用视口坐标，避免被文件列表的滚动容器裁剪。
  menuStyle.value = {
    top: `${clamp(rawTop, viewportPadding, Math.max(viewportPadding, window.innerHeight - menuHeight - viewportPadding))}px`,
    left: `${clamp(rawLeft, viewportPadding, Math.max(viewportPadding, window.innerWidth - menuWidth - viewportPadding))}px`,
    width: `${menuWidth}px`,
    visibility: 'visible',
  }
}

const getPreviewMetrics = () => {
  const isCompactViewport = window.innerWidth <= 768
  return {
    width: Math.min(isCompactViewport ? 320 : 380, window.innerWidth - (isCompactViewport ? 32 : 56)),
    preferredHeight: isCompactViewport ? 280 : 320,
    minHeight: isCompactViewport ? 160 : 180,
    viewportPadding: isCompactViewport ? 16 : 24,
  }
}

const updateHoverPreviewPlacement = () => {
  if (!isHoverPreviewVisible.value) return
  if (!hoverZoneRef.value) return

  const rect = hoverZoneRef.value.getBoundingClientRect()
  const { width, preferredHeight, minHeight, viewportPadding } = getPreviewMetrics()
  const floatingGap = 10

  const viewportWidth = window.innerWidth
  const viewportHeight = window.innerHeight
  const canOpenRight = rect.left + width <= viewportWidth - viewportPadding
  const canOpenLeft = rect.right - width >= viewportPadding
  const availableBelow = window.innerHeight - rect.bottom - viewportPadding - floatingGap
  const availableAbove = rect.top - viewportPadding - floatingGap
  const canOpenDown = availableBelow >= minHeight
  const canOpenUp = availableAbove >= minHeight
  const horizontal = !canOpenRight && canOpenLeft ? 'left' : 'right'
  const vertical = canOpenDown || availableBelow >= availableAbove || !canOpenUp ? 'down' : 'up'
  const horizontalStart = horizontal === 'left' ? rect.right - width : rect.left
  const availableHeight = vertical === 'down' ? availableBelow : availableAbove
  const maxHeight = Math.min(preferredHeight, Math.max(minHeight, availableHeight))
  const rawTop = vertical === 'down' ? rect.bottom + floatingGap : rect.top - floatingGap - maxHeight

  const clamp = (value: number, min: number, max: number) => Math.min(Math.max(value, min), max)
  const left = clamp(horizontalStart, viewportPadding, viewportWidth - viewportPadding - width)
  const top = clamp(rawTop, viewportPadding, viewportHeight - viewportPadding - maxHeight)

  previewPlacement.value = {
    horizontal,
    vertical,
  }

  // 预览层已经 Teleport 到 body，需要用视口坐标手动定位，避免被列表容器裁剪。
  hoverPreviewStyle.value = {
    width: `${width}px`,
    minHeight: `${Math.min(220, maxHeight)}px`,
    maxHeight: `${maxHeight}px`,
    left: `${left}px`,
    top: `${top}px`,
  }
}

const clearHoverPreviewCloseTimer = () => {
  if (hoverPreviewCloseTimer !== null) {
    window.clearTimeout(hoverPreviewCloseTimer)
    hoverPreviewCloseTimer = null
  }
}

const handleHoverPreviewEnter = () => {
  if (!hasHoverPreview.value || showMenu.value) return

  clearHoverPreviewCloseTimer()
  isHoverPreviewVisible.value = true
  updateHoverPreviewPlacement()
}

const handleHoverPreviewLeave = () => {
  clearHoverPreviewCloseTimer()

  // 鼠标从卡片移动到 Teleport 出来的浮层时会短暂离开原区域，延迟关闭可以保留可滚动预览。
  hoverPreviewCloseTimer = window.setTimeout(() => {
    isHoverPreviewVisible.value = false
  }, 120)
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
  isHoverPreviewVisible.value = false

  if (showMenu.value) {
    // 首次渲染先隐藏菜单，待拿到真实尺寸后再定位，避免用固定高度误判上下空间。
    menuStyle.value = { visibility: 'hidden' }
    nextTick(updateMenuPlacement)
  } else {
    menuStyle.value = {}
  }
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
  window.addEventListener('scroll', updateHoverPreviewPlacement, true)
  window.addEventListener('resize', updateMenuPlacement)
  window.addEventListener('scroll', updateMenuPlacement, true)
})

onUnmounted(() => {
  clearHoverPreviewCloseTimer()
  document.removeEventListener('click', handleClickOutside)
  window.removeEventListener('resize', updateHoverPreviewPlacement)
  window.removeEventListener('scroll', updateHoverPreviewPlacement, true)
  window.removeEventListener('resize', updateMenuPlacement)
  window.removeEventListener('scroll', updateMenuPlacement, true)
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

.select-box {
  position: absolute;
  top: 8px;
  left: 8px;
  z-index: 100;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  opacity: 0;
  transition: opacity 0.2s ease;
}

/* 选择控件默认退到卡片视觉层后方，悬停、已选中或键盘操作时再显示。 */
.file-card:hover .select-box,
.file-card:focus-within .select-box,
.file-card.selected .select-box,
.file-card.menu-open .select-box {
  opacity: 1;
}

.select-box input[type="checkbox"] {
  width: 18px;
  height: 18px;
  margin: 0;
  cursor: pointer;
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
  position: fixed;
  width: min(212px, calc(100vw - 24px));
  min-width: 0;
  padding: 6px;
  border-radius: 10px;
  border: 1px solid var(--border-color);
  background: var(--surface-color);
  box-shadow: 0 10px 24px var(--shadow-strong);
  z-index: 1900;
  transform-origin: top right;
}

.menu-item {
  width: 100%;
  padding: 9px 10px;
  border-radius: 8px;
  text-align: left;
  font-size: 13px;
  color: var(--text-soft);
  white-space: nowrap;
}

.menu-enter-active,
.menu-leave-active {
  transition: opacity 0.14s ease, transform 0.14s ease;
}

.menu-enter-from,
.menu-leave-to {
  opacity: 0;
  transform: translateY(-4px) scale(0.98);
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
  position: fixed;
  width: min(380px, calc(100vw - 56px));
  padding: 16px 12px 16px 16px;
  border-radius: 12px;
  background: rgba(24, 32, 43, 0.96);
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 0 16px 32px rgba(15, 23, 42, 0.28);
  color: #eef4ff;
  pointer-events: auto;
  z-index: 1800;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: rgba(226, 236, 255, 0.38) transparent;
}

.file-hover-preview::-webkit-scrollbar {
  width: 8px;
}

.file-hover-preview::-webkit-scrollbar-track {
  margin: 10px 0;
  background: transparent;
}

.file-hover-preview::-webkit-scrollbar-thumb {
  min-height: 42px;
  border: 2px solid rgba(24, 32, 43, 0.96);
  border-radius: 999px;
  background: rgba(226, 236, 255, 0.34);
}

.file-hover-preview::-webkit-scrollbar-thumb:hover {
  background: rgba(226, 236, 255, 0.56);
}

.hover-preview-enter-active,
.hover-preview-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.hover-preview-enter-from,
.hover-preview-leave-to {
  opacity: 0;
  transform: translateY(8px);
}

.file-hover-preview.preview-up.hover-preview-enter-from,
.file-hover-preview.preview-up.hover-preview-leave-to {
  transform: translateY(-8px);
}

.hover-preview-enter-to,
.hover-preview-leave-from {
  opacity: 1;
  transform: translateY(0);
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
  }
}
</style>

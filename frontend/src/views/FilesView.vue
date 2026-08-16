<template>
  <div class="page-shell files-view" @click="showFilterPopover = false">
    <AppHeader />

    <div class="page-content">
      <div class="page-stack">
        <section class="library-toolbar" @click.stop>
          <div class="library-tools-row">
            <SearchBar
              v-model="searchQuery"
              class="toolbar-search"
              :debounce="300"
              @search="handleSearch"
              @clear="handleClearSearch"
            />

            <div class="library-tool-actions">
              <div class="filter-control">
                <button
                  type="button"
                  class="filter-trigger"
                  :class="{ active: showFilterPopover || activeFilterCount > 0 }"
                  @click="showFilterPopover = !showFilterPopover"
                >
                  <span class="filter-trigger-dot"></span>
                  <span>筛选</span>
                  <span v-if="activeFilterCount" class="filter-trigger-count">{{ activeFilterCount }}</span>
                </button>

                <div v-if="showFilterPopover" class="filter-popover" @click.stop>
                  <div class="popover-head">
                    <strong>缩小范围</strong>
                    <button type="button" class="text-button" @click="clearAllFilters">清除全部</button>
                  </div>

                  <div class="filter-group">
                    <span class="filter-group-label">文件类型</span>
                    <FileTypeFilter
                      v-model="filesStore.currentFileType"
                      :counts="filesStore.fileTypeCounts"
                      @filter="handleFilter"
                    />
                  </div>

                  <div class="filter-group">
                    <span class="filter-group-label">合集</span>
                    <CollectionFilter
                      v-model="filesStore.currentCollectionId"
                      :collections="collections"
                      :files="filesStore.files"
                      :counts="filesStore.collectionFilterCounts"
                      :total-count="filesStore.totalFiles"
                      @filter="handleCollectionFilter"
                    />
                  </div>
                </div>
              </div>

              <ListDisplayModeToggle v-model="listMode" />
            </div>
          </div>
        </section>

        <FilesBatchToolbar
          v-if="filesStore.selectedFileIds.length > 0"
          :selected-ids="filesStore.selectedFileIds"
          @move="openMoveDialog"
          @delete="confirmBatchDelete"
          @clear="filesStore.clearSelection()"
        />

        <div v-if="filesStore.loading" class="state-panel loading-state">
          <div class="spinner"></div>
          <p>正在加载资料库...</p>
        </div>

        <div v-else-if="filesStore.error" class="state-panel error-state">
          <h3>加载失败</h3>
          <p>{{ filesStore.error }}</p>
          <button class="btn primary" @click="handleRetry">重试</button>
        </div>

        <section v-else class="file-list-panel">
          <FileList
            :files="filesStore.files"
            :group-by-collection="listMode === 'grouped'"
            :selectable="true"
            :selected-ids="filesStore.selectedFileIds"
            :empty-title="emptyStateTitle"
            :empty-description="emptyStateDescription"
            :empty-state-mode="emptyStateMode"
            @toggle-select="filesStore.toggleSelection"
            @open="handleOpenFile"
            @click="handleClickFile"
            @delete="handleDeleteFile"
            @updated="handleUpdateFileMetadata"
          />

          <PaginationControls
            v-if="filesStore.totalFiles > 0"
            variant="overlay"
            :page="filesStore.currentPage"
            :page-size="filesStore.pageSize"
            :total="filesStore.totalFiles"
            :page-sizes="[20, 50, 100]"
            @change="filesStore.goToPage"
            @change-size="filesStore.setPageSize"
          />
        </section>
      </div>
    </div>

    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showMoveDialog" class="modal-overlay" @click="closeMoveDialog">
          <div class="modal-content move-dialog" @click.stop>
            <div class="modal-header">
              <h3 class="modal-title">移动合集</h3>
              <button type="button" class="modal-close" aria-label="Close" @click="closeMoveDialog">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="18" y1="6" x2="6" y2="18"></line>
                  <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
              </button>
            </div>

            <div class="modal-body">
              <div class="form-group">
                <label>目标合集</label>
                <CollectionSelector v-model="targetCollectionId" :collections="collections" />
              </div>
            </div>

            <div class="modal-footer">
              <button type="button" class="btn secondary" @click="closeMoveDialog">取消</button>
              <button type="button" class="btn primary" :disabled="moving" @click="confirmMove">
                {{ moving ? '正在移动...' : '确认移动' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <BatchMoveProgressOverlay
      v-if="moving"
      :total="moveTotal"
      :completed="moveCompleted"
      :message="moveMessage"
      :collection-name="moveCollectionName"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, isWailsAvailable } from '@/api/index'
import AppHeader from '@/components/AppHeader.vue'
import BatchMoveProgressOverlay from '@/components/BatchMoveProgressOverlay.vue'
import CollectionFilter from '@/components/CollectionFilter.vue'
import CollectionSelector from '@/components/CollectionSelector.vue'
import FileList from '@/components/FileList.vue'
import FileTypeFilter from '@/components/FileTypeFilter.vue'
import FilesBatchToolbar from '@/components/FilesBatchToolbar.vue'
import ListDisplayModeToggle from '@/components/ListDisplayModeToggle.vue'
import PaginationControls from '@/components/PaginationControls.vue'
import SearchBar from '@/components/SearchBar.vue'
import { useFilesStore } from '@/store/modules/files'
import type { Collection } from '@/types'
import { isUnsortedCollectionName, UNSORTED_COLLECTION_KEY } from '@/utils/constants'

const filesStore = useFilesStore()
const route = useRoute()
const router = useRouter()
const searchQuery = ref('')
const listMode = ref<'flat' | 'grouped'>('flat')
const collections = ref<Collection[]>([])
const checkIntervals = new Map<number, NodeJS.Timeout>()
let isUnmounted = false
const showFilterPopover = ref(false)

const showMoveDialog = ref(false)
const targetCollectionId = ref<number | undefined>(undefined)
const moving = ref(false)
const moveTotal = ref(0)
const moveCompleted = ref(0)
const moveMessage = ref('')
const moveCollectionName = ref('')
const routeCollectionKey = ref('all')

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

const getQueryValue = (value: unknown) => {
  return typeof value === 'string' ? value : ''
}

const applyRouteState = () => {
  const routeType = getQueryValue(route.query.type) || 'all'
  const routeCollection = getQueryValue(route.query.collection) || 'all'
  const routeView = getQueryValue(route.query.view) === 'grouped' ? 'grouped' : 'flat'
  const routeSearch = getQueryValue(route.query.q)

  filesStore.currentFileType = routeType
  // 先保留路由里的原始集合键，等集合列表加载完成后再解析成 id，避免名称和 id 混用.
  routeCollectionKey.value = routeCollection
  filesStore.currentCollectionId = 'all'
  searchQuery.value = routeSearch
  listMode.value = routeView
}

const resolveRouteCollectionId = () => {
  const normalized = routeCollectionKey.value.trim()
  if (!normalized || normalized === 'all') {
    return 'all' as const
  }
  if (
    normalized === 'unsorted' ||
    normalized === UNSORTED_COLLECTION_KEY ||
    isUnsortedCollectionName(normalized)
  ) {
    return 'unsorted' as const
  }

  const numericId = Number(normalized)
  if (Number.isInteger(numericId) && numericId > 0) {
    return numericId
  }

  const matchedCollection = collections.value.find((collection) => collection.name === normalized)
  return matchedCollection?.id ?? 'all'
}

const syncRouteCollectionSelection = () => {
  filesStore.currentCollectionId = resolveRouteCollectionId()
}

const persistRouteState = async () => {
  const query: Record<string, string> = {}

  if (filesStore.currentFileType !== 'all') query.type = filesStore.currentFileType
  if (filesStore.currentCollectionId !== 'all') {
    query.collection = String(filesStore.currentCollectionId)
  }
  if (listMode.value !== 'flat') query.view = listMode.value
  if (searchQuery.value.trim()) query.q = searchQuery.value.trim()

  await router.replace({ name: 'Files', query })
}

const refreshCurrentResults = async () => {
  if (searchQuery.value.trim()) {
    await filesStore.searchFiles(searchQuery.value.trim(), filesStore.currentPage, filesStore.pageSize)
    return
  }

  await filesStore.loadFiles(filesStore.currentFileType, filesStore.currentPage, filesStore.pageSize)
}

const waitForWails = async (timeoutMs = 5000, intervalMs = 100) => {
  const deadline = Date.now() + timeoutMs

  while (!isWailsAvailable() && Date.now() < deadline) {
    if (isUnmounted) return false
    await sleep(intervalMs)
  }

  return isWailsAvailable()
}

const loadInitialFiles = async () => {
  const wailsReady = await waitForWails()
  if (!wailsReady) {
    console.warn('Wails not available after timeout, loading with fallback behavior...')
  }

  if (!isUnmounted) {
    await loadCollections()
    syncRouteCollectionSelection()
    await refreshCurrentResults()
  }
}

const loadCollections = async () => {
  try {
    collections.value = await api.collection.getAll()
  } catch (err) {
    console.error('Failed to load collections:', err)
  }
}

const activeFilterCount = computed(() => {
  let count = 0
  if (filesStore.currentFileType !== 'all') count += 1
  if (filesStore.currentCollectionId !== 'all') count += 1
  return count
})

const emptyStateTitle = computed(() => {
  return searchQuery.value.trim() ? '没有匹配的文件' : '资料库还是空的'
})

const emptyStateDescription = computed(() => {
  return searchQuery.value.trim()
    ? '试试换个关键词，或者清空搜索后查看全部资料。'
    : '从导入页添加文件后，会在这里形成你的资料库。'
})

const emptyStateMode = computed<'library' | 'search'>(() => {
  return searchQuery.value.trim() ? 'search' : 'library'
})

onMounted(() => {
  applyRouteState()
  void loadInitialFiles()
})

onUnmounted(() => {
  isUnmounted = true
  checkIntervals.forEach((interval) => clearInterval(interval))
  checkIntervals.clear()
})

watch([() => filesStore.currentFileType, () => filesStore.currentCollectionId, listMode], () => {
  void persistRouteState()
})

const handleFilter = async (fileType: string) => {
  searchQuery.value = ''
  filesStore.selectedFileIds = []
  await filesStore.setCurrentFileType(fileType)
}

const handleCollectionFilter = async (collectionId: 'all' | 'unsorted' | number) => {
  filesStore.selectedFileIds = []
  await filesStore.setCurrentCollectionId(collectionId)
}

const clearCollectionFilter = () => {
  handleCollectionFilter('all')
}

const clearAllFilters = async () => {
  // 原型里的“清除全部”只重置当前已有筛选，不改变搜索文本之外的产品状态。
  showFilterPopover.value = false
  if (filesStore.currentFileType !== 'all') {
    await handleFilter('all')
  }
  if (filesStore.currentCollectionId !== 'all') {
    await handleCollectionFilter('all')
  }
}

const handleSearch = async (query: string) => {
  const trimmedQuery = query.trim()
  searchQuery.value = trimmedQuery
  filesStore.selectedFileIds = []

  if (!trimmedQuery) {
    await handleClearSearch()
    return
  }

  await filesStore.searchFiles(trimmedQuery, 1, filesStore.pageSize)
  await persistRouteState()
}

const handleClearSearch = async () => {
  searchQuery.value = ''
  filesStore.selectedFileIds = []
  await filesStore.clearSearch()
  await persistRouteState()
}

const handleOpenFile = async (id: number) => {
  try {
    await api.file.open(id)
    startFileUpdateCheck(id)
  } catch (error) {
    console.error('Failed to open file:', error)
    window.alert('打开文件失败')
  }
}

const startFileUpdateCheck = (fileId: number) => {
  if (checkIntervals.has(fileId)) {
    clearInterval(checkIntervals.get(fileId))
  }

  let originalFile: any = null

  checkFileUpdate(fileId, originalFile).then((updatedFile) => {
    if (updatedFile) originalFile = updatedFile
  })

  const interval = setInterval(async () => {
    const updatedFile = await checkFileUpdate(fileId, originalFile)
    if (updatedFile) originalFile = updatedFile
  }, 3000)

  checkIntervals.set(fileId, interval)

  setTimeout(() => {
    stopFileUpdateCheck(fileId)
  }, 5 * 60 * 1000)
}

const checkFileUpdate = async (fileId: number, originalFile: any | null): Promise<any> => {
  try {
    const updatedFile = await api.file.refresh(fileId)
    if (!updatedFile) return null
    if (!originalFile) return updatedFile

    const hasChanges =
      updatedFile.fileName !== originalFile.fileName ||
      updatedFile.fileSize !== originalFile.fileSize

    if (hasChanges) {
      await refreshCurrentResults()
      return updatedFile
    }

    return null
  } catch (error: any) {
    if (error.message && error.message.includes('file no longer exists')) {
      stopFileUpdateCheck(fileId)
      await refreshCurrentResults()
    } else {
      console.error('Failed to check file update:', error)
    }

    return null
  }
}

const stopFileUpdateCheck = (fileId: number) => {
  if (checkIntervals.has(fileId)) {
    clearInterval(checkIntervals.get(fileId))
    checkIntervals.delete(fileId)
  }
}

const handleClickFile = (file: any) => {
  console.log('File clicked:', file.fileName)
}

const handleUpdateFileMetadata = async () => {
  await refreshCurrentResults()
}

const handleDeleteFile = async () => {
  filesStore.selectedFileIds = filesStore.selectedFileIds.filter(
    (id) => !filesStore.files.some((file) => file.id === id)
  )
  await refreshCurrentResults()
}

const handleRetry = async () => {
  await refreshCurrentResults()
}

const openMoveDialog = () => {
  targetCollectionId.value = undefined
  showMoveDialog.value = true
}

const closeMoveDialog = () => {
  showMoveDialog.value = false
  targetCollectionId.value = undefined
}

const confirmMove = async () => {
  const ids = filesStore.selectedFileIds
  if (ids.length === 0) {
    closeMoveDialog()
    return
  }

  const selectedTargetCollectionId = targetCollectionId.value
  const targetName = collections.value.find((c) => c.id === selectedTargetCollectionId)?.name || '未分配合集'
  moveCollectionName.value = targetName
  moveTotal.value = ids.length
  moveCompleted.value = 0
  moveMessage.value = '准备移动...'
  moving.value = true
  closeMoveDialog()

  try {
    const failedMessages: string[] = []
    for (let i = 0; i < ids.length; i++) {
      moveCompleted.value = i
      moveMessage.value = `正在处理第 ${i + 1} / ${ids.length} 个文件`
      const res = await api.file.batchUpdateCollection([ids[i]], selectedTargetCollectionId)
      if (res.failedCount > 0) {
        console.error('Failed to move file:', res.failedItems)
        // 后端批量接口会返回单文件失败项，这里集中提示，避免用户误以为没有执行。
        ;(res.failedItems || []).forEach((item: { fileName?: string; fileID?: number; error?: string }) => {
          const label = item.fileName || `ID ${item.fileID ?? ids[i]}`
          failedMessages.push(`${label}: ${item.error || '移动失败'}`)
        })
      }
    }
    moveCompleted.value = ids.length
    moveMessage.value = '移动完成'
    filesStore.clearSelection()
    await refreshCurrentResults()
    if (failedMessages.length > 0) {
      window.alert(`移动完成，但有 ${failedMessages.length} 个文件失败：\n${failedMessages.slice(0, 5).join('\n')}`)
    }
  } catch (error) {
    console.error('Failed to batch move files:', error)
    window.alert('移动文件失败')
  } finally {
    setTimeout(() => {
      moving.value = false
    }, 600)
  }
}

const confirmBatchDelete = async () => {
  const ids = filesStore.selectedFileIds
  if (ids.length === 0) return

  if (!window.confirm(`确定要删除选中的 ${ids.length} 个文件吗？`)) {
    return
  }

  try {
    const res = await api.file.batchDelete(ids)
    if (res.failedCount > 0) {
      console.error('Failed to delete files:', res.failedItems)
      window.alert(`删除完成：成功 ${res.successCount} 个，失败 ${res.failedCount} 个`)
    }
    filesStore.clearSelection()
    await refreshCurrentResults()
  } catch (error) {
    console.error('Failed to batch delete files:', error)
    window.alert('删除文件失败')
  }
}
</script>

<style scoped>
.files-view {
  background: transparent;
}

.files-view .page-content {
  overflow: hidden;
}

.files-view .page-stack {
  height: 100%;
  min-height: 0;
}

.library-toolbar {
  position: relative;
  display: flex;
  align-items: center;
  flex-shrink: 0;
  padding: 0 0 18px;
}

.library-tools-row {
  display: flex;
  align-items: center;
  width: 100%;
  gap: 10px;
}

.toolbar-search {
  flex: 1;
  min-width: 200px;
}

.library-tool-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.filter-control {
  position: static;
}

.filter-trigger {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  height: 43px;
  padding: 0 13px;
  border: 1px solid var(--border-color);
  border-radius: 12px;
  background: var(--content-control-bg);
  color: var(--text-soft);
  font-size: 12px;
  font-weight: 650;
}

.filter-trigger:hover,
.filter-trigger.active {
  border-color: rgba(111, 143, 216, 0.42);
  color: var(--primary-hover);
  background: var(--content-control-hover);
}

.filter-trigger-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--primary-color);
}

.filter-trigger-count {
  min-width: 17px;
  padding: 2px 5px;
  border-radius: 6px;
  background: rgba(111, 143, 216, 0.14);
  font-size: 10px;
  text-align: center;
}

.filter-popover {
  position: absolute;
  z-index: 20;
  top: calc(100% + 9px);
  right: 72px;
  width: min(430px, calc(100vw - 50px));
  padding: 16px;
  border: 1px solid rgba(146, 165, 192, 0.36);
  border-radius: 15px;
  background: var(--content-popover-bg);
  box-shadow: 0 18px 38px rgba(29, 48, 78, 0.14);
}

.filter-popover::before {
  content: '';
  position: absolute;
  top: -6px;
  right: 119px;
  width: 11px;
  height: 11px;
  border-top: 1px solid rgba(146, 165, 192, 0.36);
  border-left: 1px solid rgba(146, 165, 192, 0.36);
  background: var(--content-popover-bg);
  transform: rotate(45deg);
}

.popover-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 13px;
  border-bottom: 1px solid rgba(146, 165, 192, 0.24);
}

.popover-head strong {
  color: var(--text-color);
  font-size: 13px;
}

.text-button {
  padding: 2px 0;
  color: var(--primary-color);
  background: transparent;
  font-size: 11px;
}

.filter-group {
  padding-top: 14px;
}

.filter-group-label {
  display: block;
  margin-bottom: 9px;
  color: var(--text-faint);
  font-size: 10px;
  font-weight: 750;
}

.filter-popover :deep(.file-type-filter),
.filter-popover :deep(.collection-filter) {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
  overflow: visible;
  padding-bottom: 0;
}

/* 筛选条件按文字长度自适应，避免短标签被拉成大块，长标签也能自然换行。 */
.filter-popover :deep(.file-type-filter) {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.filter-popover :deep(.file-type-filter button),
.filter-popover :deep(.collection-filter button) {
  justify-content: center;
  width: auto;
  flex: 0 0 auto;
  min-height: 30px;
  padding: 6px 9px;
  border-radius: 8px;
  background: var(--content-option-bg);
  font-size: 11px;
  white-space: nowrap;
}

.filter-popover :deep(.file-type-filter button) {
  justify-content: center;
  width: auto;
  flex: 0 0 auto;
  white-space: nowrap;
}

.filter-popover :deep(.filter-icon) {
  display: none;
}

.filter-popover :deep(.filter-label),
.filter-popover :deep(.collection-filter-label) {
  font-size: 11px;
}

.filter-popover :deep(.filter-count),
.filter-popover :deep(.collection-filter-count) {
  display: none;
}

/* 资料库搜索字段对齐原型的工具条样式，搜索仍由现有 debounce 与接口逻辑驱动。 */
.toolbar-search :deep(.search-bar) {
  gap: 0;
}

.toolbar-search :deep(.search-input-wrapper) {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 43px;
  padding: 0 13px;
  border: 1px solid var(--border-color);
  border-radius: 12px;
  background: var(--content-control-bg);
  box-shadow: 0 5px 13px rgba(40, 60, 90, 0.03);
}

.toolbar-search :deep(.search-icon),
.toolbar-search :deep(.clear-button) {
  position: static;
  flex-shrink: 0;
  transform: none;
}

.toolbar-search :deep(.search-input) {
  min-height: 0;
  padding: 0;
  border: 0;
  outline: 0;
  background: transparent;
  box-shadow: none;
}

.toolbar-search :deep(.search-button) {
  display: none;
}

.file-list-panel {
  display: flex;
  position: relative;
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.state-panel {
  display: flex;
  flex: 1;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  min-height: 0;
  border-radius: 28px;
  border: 1px solid rgba(146, 165, 192, 0.18);
  background: var(--content-state-bg);
  color: var(--text-soft);
}

/* 加载只需要反馈进度，不应在页面切换期间制造一块抢眼的大面板。 */
.state-panel.loading-state {
  border-color: transparent;
  border-radius: 0;
  background: transparent;
}

.error-state h3 {
  font-size: 22px;
  color: var(--text-color);
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

.move-dialog {
  width: min(420px, 100%);
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

@media (max-width: 980px) {
  .library-tools-row {
    flex-wrap: wrap;
  }

  .toolbar-search {
    flex-basis: 100%;
  }
}

@media (max-width: 640px) {
  .library-toolbar {
    padding-bottom: 0;
  }

  .library-tool-actions {
    width: 100%;
    justify-content: space-between;
  }

  .filter-popover {
    right: 0;
    width: min(430px, calc(100vw - 24px));
  }
}
</style>



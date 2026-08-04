<template>
  <div class="page-shell files-view">
    <AppHeader />

    <div class="page-content">
      <div class="page-stack">
        <section class="toolbar-panel library-toolbar">
          <div class="toolbar-top">
            <FileTypeFilter
              v-model="filesStore.currentFileType"
              @filter="handleFilter"
            />
          </div>

          <div class="toolbar-middle">
            <CollectionFilter
              v-model="filesStore.currentCollectionId"
              :collections="collections"
              :files="filesStore.files"
              :counts="filesStore.collectionFilterCounts"
              :total-count="filesStore.totalFiles"
              @filter="handleCollectionFilter"
            />
          </div>

          <FilesBatchToolbar
            v-if="filesStore.selectedFileIds.length > 0"
            :selected-ids="filesStore.selectedFileIds"
            @move="openMoveDialog"
            @delete="confirmBatchDelete"
            @clear="filesStore.clearSelection()"
          />

          <div class="toolbar-bottom">
            <SearchBar
              v-model="searchQuery"
              class="toolbar-search"
              :debounce="300"
              @search="handleSearch"
              @clear="handleClearSearch"
            />

            <div class="toolbar-side">
              <p class="results-summary">{{ resultsSummary }}</p>

              <div class="search-row-actions">
                <button
                  v-if="filesStore.currentCollectionId !== 'all'"
                  type="button"
                  class="btn secondary compact-button"
                  @click="clearCollectionFilter"
                >
                  清除合集筛选
                </button>

                <ListDisplayModeToggle v-model="listMode" />
              </div>
            </div>
          </div>
        </section>

        <div v-if="filesStore.loading" class="state-panel">
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
  if (normalized === 'unsorted' || normalized === UNSORTED_COLLECTION_KEY || isUnsortedCollectionName(normalized)) {
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

const resultsSummary = computed(() => {
  if (searchQuery.value.trim()) {
    return `搜索到 ${filesStore.totalFiles} 个文件`
  }

  if (filesStore.currentCollectionId !== 'all') {
    return `当前筛选共 ${filesStore.totalFiles} 个文件`
  }

  return `共 ${filesStore.totalFiles} 个文件`
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
  display: grid;
  gap: 10px;
  flex-shrink: 0;
  padding: 14px 16px;
}

.toolbar-bottom {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 12px;
  align-items: end;
}

.toolbar-side {
  display: grid;
  gap: 8px;
  justify-items: end;
  min-width: 250px;
}

.results-summary {
  font-size: 13px;
  color: var(--text-faint);
}

.search-row-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.compact-button {
  min-width: 124px;
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
  background: rgba(255, 255, 255, 0.64);
  color: var(--text-soft);
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
  .toolbar-bottom {
    grid-template-columns: 1fr;
  }

  .toolbar-side {
    justify-items: stretch;
  }
}

@media (max-width: 640px) {
  .library-toolbar {
    padding: 12px;
  }

  .search-row-actions {
    width: 100%;
    flex-direction: column;
    align-items: stretch;
  }
}
</style>



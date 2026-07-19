<template>
  <div class="files-view">
    <AppHeader />

    <div class="content">
      <section class="toolbar-panel">
        <div class="filters-section">
          <FileTypeFilter
            v-model="filesStore.currentFileType"
            :counts="filesStore.fileTypeCounts"
            @filter="handleFilter"
          />

          <CollectionFilter
            v-if="filesStore.collectionSummaries.length > 0"
            v-model="filesStore.currentCollectionName"
            :options="filesStore.collectionSummaries"
            :total-count="filesStore.collectionCounts.all"
            @filter="handleCollectionFilter"
          />
        </div>

        <div class="search-row">
          <SearchBar
            v-model="searchQuery"
            class="toolbar-search"
            :debounce="300"
            @search="handleSearch"
            @clear="handleClearSearch"
          />

          <p class="results-summary">{{ resultsSummary }}</p>

          <div class="search-row-actions">
            <button
              v-if="filesStore.currentCollectionName !== 'all'"
              type="button"
              class="link-button"
              @click="clearCollectionFilter"
            >
              清除合集筛选
            </button>

            <ListDisplayModeToggle v-model="listMode" />
          </div>
        </div>
      </section>

      <div v-if="filesStore.loading" class="loading-state">
        <div class="spinner"></div>
        <p>加载中...</p>
      </div>

      <div v-else-if="filesStore.error" class="error-state">
        <div class="error-icon">⚠️</div>
        <h3>加载失败</h3>
        <p>{{ filesStore.error }}</p>
        <button class="btn primary" @click="handleRetry">重试</button>
      </div>

      <FileList
        v-else
        :files="filesStore.files"
        :group-by-collection="listMode === 'grouped'"
        :empty-title="emptyStateTitle"
        :empty-description="emptyStateDescription"
        :empty-state-mode="emptyStateMode"
        @open="handleOpenFile"
        @click="handleClickFile"
        @delete="handleDeleteFile"
        @updated="handleUpdateFileMetadata"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, isWailsAvailable } from '@/api/index'
import AppHeader from '@/components/AppHeader.vue'
import CollectionFilter from '@/components/CollectionFilter.vue'
import FileList from '@/components/FileList.vue'
import FileTypeFilter from '@/components/FileTypeFilter.vue'
import ListDisplayModeToggle from '@/components/ListDisplayModeToggle.vue'
import SearchBar from '@/components/SearchBar.vue'
import { useFilesStore } from '@/store/modules/files'

const filesStore = useFilesStore()
const route = useRoute()
const router = useRouter()
const searchQuery = ref('')
const listMode = ref<'flat' | 'grouped'>('flat')
const checkIntervals = new Map<number, NodeJS.Timeout>()
let isUnmounted = false

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
  filesStore.currentCollectionName = routeCollection
  searchQuery.value = routeSearch
  listMode.value = routeView
}

const persistRouteState = async () => {
  const query: Record<string, string> = {}

  if (filesStore.currentFileType !== 'all') {
    query.type = filesStore.currentFileType
  }

  if (filesStore.currentCollectionName !== 'all') {
    query.collection = filesStore.currentCollectionName
  }

  if (listMode.value !== 'flat') {
    query.view = listMode.value
  }

  if (searchQuery.value.trim()) {
    query.q = searchQuery.value.trim()
  }

  await router.replace({ name: 'Files', query })
}

const refreshCurrentResults = async () => {
  if (searchQuery.value.trim()) {
    await filesStore.searchFiles(searchQuery.value.trim())
    return
  }

  await filesStore.loadFiles(filesStore.currentFileType)
}

const waitForWails = async (timeoutMs = 5000, intervalMs = 100) => {
  const deadline = Date.now() + timeoutMs

  while (!isWailsAvailable() && Date.now() < deadline) {
    if (isUnmounted) {
      return false
    }

    await sleep(intervalMs)
  }

  return isWailsAvailable()
}

const loadInitialFiles = async () => {
  const wailsReady = await waitForWails()

  if (!wailsReady) {
    console.warn('Wails not available after timeout, loading with current API fallback behavior...')
  } else {
    console.log('Wails is ready, loading files...')
  }

  if (isUnmounted) {
    return
  }

  await refreshCurrentResults()
}

const resultsSummary = computed(() => {
  const fileCount = filesStore.files.length
  const collectionCount = filesStore.collectionSummaries.length

  if (searchQuery.value.trim()) {
    return `搜索结果 ${fileCount} 个文件`
  }

  if (listMode.value === 'grouped') {
    return `当前共 ${fileCount} 个文件，按 ${collectionCount} 个合集分组展示`
  }

  if (filesStore.currentCollectionName !== 'all') {
    return `当前合集下共 ${fileCount} 个文件`
  }

  return `当前显示 ${fileCount} 个文件`
})

const emptyStateTitle = computed(() => {
  if (searchQuery.value.trim()) {
    return '没有找到匹配文件'
  }

  return '暂无文件'
})

const emptyStateDescription = computed(() => {
  if (searchQuery.value.trim()) {
    return '试试更换关键词，或者清空搜索后查看全部文件。'
  }

  return '点击上方“导入文件”开始添加内容。'
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

  checkIntervals.forEach((interval) => {
    clearInterval(interval)
  })
  checkIntervals.clear()
})

watch([() => filesStore.currentFileType, () => filesStore.currentCollectionName, listMode], () => {
  void persistRouteState()
})

const handleFilter = async (fileType: string) => {
  searchQuery.value = ''
  filesStore.searchQuery = ''
  await filesStore.loadFiles(fileType)
}

const handleCollectionFilter = (collectionName: string) => {
  filesStore.setCurrentCollectionName(collectionName)
}

const clearCollectionFilter = () => {
  filesStore.setCurrentCollectionName('all')
}

const handleSearch = async (query: string) => {
  const trimmedQuery = query.trim()
  searchQuery.value = trimmedQuery

  if (!trimmedQuery) {
    await handleClearSearch()
    return
  }

  await filesStore.searchFiles(trimmedQuery)
  await persistRouteState()
}

const handleClearSearch = async () => {
  searchQuery.value = ''
  await filesStore.clearSearch()
  await persistRouteState()
}

const handleOpenFile = async (id: number) => {
  try {
    await api.file.open(id)
    startFileUpdateCheck(id)
  } catch (error) {
    console.error('Failed to open file:', error)
    alert('打开文件失败')
  }
}

const startFileUpdateCheck = (fileId: number) => {
  if (checkIntervals.has(fileId)) {
    clearInterval(checkIntervals.get(fileId))
  }

  let originalFile: any = null

  checkFileUpdate(fileId, originalFile).then((updatedFile) => {
    if (updatedFile) {
      originalFile = updatedFile
    }
  })

  const interval = setInterval(async () => {
    const updatedFile = await checkFileUpdate(fileId, originalFile)
    if (updatedFile) {
      originalFile = updatedFile
    }
  }, 3000)

  checkIntervals.set(fileId, interval)

  setTimeout(() => {
    stopFileUpdateCheck(fileId)
  }, 5 * 60 * 1000)
}

const checkFileUpdate = async (fileId: number, originalFile: any | null): Promise<any> => {
  try {
    const updatedFile = await api.file.refresh(fileId)
    if (!updatedFile) {
      return null
    }

    if (!originalFile) {
      return updatedFile
    }

    const hasChanges =
      updatedFile.fileName !== originalFile.fileName ||
      updatedFile.fileSize !== originalFile.fileSize

    if (hasChanges) {
      console.log('File updated:', updatedFile.fileName, updatedFile.fileSize)
      await refreshCurrentResults()
      return updatedFile
    }

    return null
  } catch (error: any) {
    if (error.message && error.message.includes('file no longer exists')) {
      console.log('File no longer exists, stopping check and refreshing list')
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
  await refreshCurrentResults()
}

const handleRetry = async () => {
  await refreshCurrentResults()
}
</script>

<style scoped>
.files-view {
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
  padding: 14px 18px 18px;
  overflow: hidden;
  min-height: 0;
  gap: 12px;
}

.toolbar-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px 16px;
  border-radius: 18px;
  background-color: color-mix(in srgb, var(--surface-color) 90%, transparent);
  border: 1px solid rgba(148, 163, 184, 0.16);
}

.filters-section {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.search-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-width: 0;
}

.results-summary {
  margin: 0;
  flex-shrink: 0;
  font-size: 12px;
  color: var(--text-color);
  opacity: 0.68;
}

.link-button {
  padding: 0;
  border: none;
  background: transparent;
  color: var(--primary-color);
  font-size: 12px;
  font-weight: 600;
}

.search-row-actions {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.toolbar-search {
  flex: 1;
  min-width: 0;
}

.toolbar-panel :deep(.file-type-filter) {
  gap: 6px;
  padding: 0;
}

.toolbar-panel :deep(.file-type-filter button) {
  gap: 5px;
  padding: 6px 12px;
  border-radius: 16px;
  font-size: 13px;
}

.toolbar-panel :deep(.filter-icon) {
  font-size: 14px;
}

.toolbar-panel :deep(.filter-count) {
  min-width: 18px;
  padding: 1px 5px;
  font-size: 11px;
}

.toolbar-panel :deep(.collection-filter) {
  gap: 8px;
  padding: 0 0 4px;
}

.toolbar-panel :deep(.collection-filter button) {
  padding: 7px 11px;
}

.toolbar-panel :deep(.collection-filter-label) {
  font-size: 12px;
}

.toolbar-panel :deep(.collection-filter-count) {
  min-width: 20px;
  padding: 1px 6px;
  font-size: 11px;
}

.toolbar-panel :deep(.search-bar) {
  max-width: none;
  margin: 0;
}

.toolbar-panel :deep(.search-input) {
  padding-top: 9px;
  padding-bottom: 9px;
  border-radius: 16px;
}

.toolbar-panel :deep(.search-button) {
  padding: 9px 16px;
  border-radius: 16px;
}

.toolbar-panel :deep(.display-mode-toggle) {
  padding: 3px;
}

.toolbar-panel :deep(.display-mode-toggle button) {
  min-width: 84px;
  padding: 7px 12px;
  font-size: 12px;
}

.loading-state,
.error-state {
  display: flex;
  flex: 1;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: var(--text-color);
  gap: 16px;
  text-align: center;
}

.error-icon {
  font-size: 48px;
}

@media (max-width: 900px) {
  .search-row {
    flex-direction: column;
    align-items: stretch;
  }

  .results-summary {
    flex-shrink: 1;
  }

  .search-row-actions {
    justify-content: space-between;
  }
}

@media (max-width: 640px) {
  .content {
    padding: 12px 14px 14px;
  }

  .toolbar-panel {
    padding: 10px 12px;
    border-radius: 16px;
  }

  .search-row-actions {
    width: 100%;
  }
}
</style>

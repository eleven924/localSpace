<template>
  <div class="page-shell files-view">
    <AppHeader />

    <div class="page-content">
      <div class="page-stack">
        <section class="toolbar-panel library-toolbar">
          <div class="toolbar-top">
            <FileTypeFilter
              v-model="filesStore.currentFileType"
              :counts="filesStore.fileTypeCounts"
              @filter="handleFilter"
            />
          </div>

          <div v-if="filesStore.collectionSummaries.length > 0" class="toolbar-middle">
            <CollectionFilter
              v-model="filesStore.currentCollectionName"
              :options="filesStore.collectionSummaries"
              :total-count="filesStore.collectionCounts.all"
              @filter="handleCollectionFilter"
            />
          </div>

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
                  v-if="filesStore.currentCollectionName !== 'all'"
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
            :empty-title="emptyStateTitle"
            :empty-description="emptyStateDescription"
            :empty-state-mode="emptyStateMode"
            @open="handleOpenFile"
            @click="handleClickFile"
            @delete="handleDeleteFile"
            @updated="handleUpdateFileMetadata"
          />
        </section>
      </div>
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

  if (filesStore.currentFileType !== 'all') query.type = filesStore.currentFileType
  if (filesStore.currentCollectionName !== 'all') query.collection = filesStore.currentCollectionName
  if (listMode.value !== 'flat') query.view = listMode.value
  if (searchQuery.value.trim()) query.q = searchQuery.value.trim()

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
    await refreshCurrentResults()
  }
}

const resultsSummary = computed(() => {
  const fileCount = filesStore.files.length
  const collectionCount = filesStore.collectionSummaries.length

  if (searchQuery.value.trim()) {
    return `搜索到 ${fileCount} 个文件`
  }

  if (listMode.value === 'grouped') {
    return `当前共 ${fileCount} 个文件，按 ${collectionCount} 个合集分组`
  }

  if (filesStore.currentCollectionName !== 'all') {
    return `当前合集下共 ${fileCount} 个文件`
  }

  return `当前显示 ${fileCount} 个文件`
})

const emptyStateTitle = computed(() => {
  return searchQuery.value.trim() ? '没有匹配的文件' : '资料库还是空的'
})

const emptyStateDescription = computed(() => {
  return searchQuery.value.trim()
    ? '试试更换关键词，或者清空搜索后查看全部资料。'
    : '从导入页面添加文件后，会在这里形成你的资料库。'
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
  await refreshCurrentResults()
}

const handleRetry = async () => {
  await refreshCurrentResults()
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

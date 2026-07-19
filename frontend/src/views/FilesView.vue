<template>
  <div class="files-view">
    <header class="header">
      <div class="header-left">
        <div class="brand-block">
          <h1>LocalSpace</h1>
          <p class="subtitle">按文件类型管理，也能从合集维度连续浏览</p>
        </div>

        <nav class="header-nav">
          <router-link to="/files" class="nav-pill active">文件</router-link>
          <router-link to="/collections" class="nav-pill">合集</router-link>
        </nav>
      </div>

      <div class="header-actions">
        <router-link to="/import" class="btn primary">导入文件</router-link>
        <router-link to="/settings" class="btn secondary">设置</router-link>
      </div>
    </header>

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
            :debounce="300"
            @search="handleSearch"
            @clear="handleClearSearch"
          />

          <ListDisplayModeToggle v-model="listMode" />
        </div>

        <div class="results-bar">
          <p class="results-summary">{{ resultsSummary }}</p>
          <button
            v-if="filesStore.currentCollectionName !== 'all'"
            type="button"
            class="link-button"
            @click="clearCollectionFilter"
          >
            清除合集筛选
          </button>
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

watch(
  [() => filesStore.currentFileType, () => filesStore.currentCollectionName, listMode],
  () => {
    void persistRouteState()
  }
)

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
  searchQuery.value = query
  await filesStore.searchFiles(query.trim())
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

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
  padding: 16px 24px;
  background-color: color-mix(in srgb, var(--surface-color) 92%, transparent);
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
  min-width: 0;
}

.brand-block h1 {
  margin: 0 0 4px;
  font-size: 24px;
  font-weight: 700;
  color: var(--text-color);
}

.subtitle {
  margin: 0;
  color: var(--text-color);
  opacity: 0.7;
  font-size: 13px;
}

.header-nav {
  display: inline-flex;
  gap: 8px;
  padding: 6px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--surface-color) 82%, transparent);
  border: 1px solid rgba(148, 163, 184, 0.16);
}

.nav-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 72px;
  padding: 8px 14px;
  border-radius: 999px;
  color: var(--text-color);
  font-size: 13px;
  font-weight: 600;
}

.nav-pill.active,
.nav-pill.router-link-active {
  background: linear-gradient(135deg, var(--primary-color) 0%, color-mix(in srgb, var(--primary-color) 78%, #0f172a) 100%);
  color: #fff;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.content {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 20px 24px 24px;
  overflow: hidden;
  min-height: 0;
  gap: 16px;
}

.toolbar-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 18px 20px;
  border-radius: 22px;
  background-color: color-mix(in srgb, var(--surface-color) 90%, transparent);
  border: 1px solid rgba(148, 163, 184, 0.16);
}

.filters-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.search-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.results-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.results-summary {
  margin: 0;
  font-size: 13px;
  color: var(--text-color);
  opacity: 0.72;
}

.link-button {
  padding: 0;
  border: none;
  background: transparent;
  color: var(--primary-color);
  font-size: 13px;
  font-weight: 600;
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
  .header {
    flex-direction: column;
    align-items: flex-start;
  }

  .header-left {
    width: 100%;
    flex-direction: column;
    align-items: flex-start;
  }

  .header-actions {
    width: 100%;
    justify-content: flex-end;
  }

  .search-row {
    flex-direction: column;
    align-items: stretch;
  }
}

@media (max-width: 640px) {
  .header {
    padding: 14px 16px;
  }

  .content {
    padding: 16px;
  }

  .toolbar-panel {
    padding: 16px;
    border-radius: 18px;
  }

  .results-bar {
    flex-direction: column;
    align-items: flex-start;
  }

  .header-actions {
    width: 100%;
    flex-direction: column;
  }

  .header-actions .btn {
    width: 100%;
  }
}
</style>

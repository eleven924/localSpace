<template>
  <div class="files-view">
    <header class="header">
      <div class="header-left">
        <h1>LocalSpace</h1>
      </div>
      <div class="header-actions">
        <router-link to="/import" class="btn primary">
          导入文件
        </router-link>
        <router-link to="/settings" class="btn secondary">
          设置
        </router-link>
      </div>
    </header>

    <div class="content">
      <div class="filters-section">
        <FileTypeFilter
          v-model="filesStore.currentFileType"
          :counts="filesStore.fileTypeCounts"
          @filter="handleFilter"
        />
      </div>

      <div class="search-section">
        <SearchBar
          v-model="searchQuery"
          :debounce="300"
          @search="handleSearch"
          @clear="handleClearSearch"
        />
      </div>

      <div v-if="filesStore.loading" class="loading-state">
        <div class="spinner"></div>
        <p>加载中...</p>
      </div>

      <div v-else-if="filesStore.error" class="error-state">
        <div class="error-icon">⚠️</div>
        <h3>加载失败</h3>
        <p>{{ filesStore.error }}</p>
        <button class="btn primary" @click="handleRetry">
          重试
        </button>
      </div>

      <FileList
        v-else
        :files="filesStore.files"
        @open="handleOpenFile"
        @click="handleClickFile"
        @delete="handleDeleteFile"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useFilesStore } from '@/store/modules/files'
import { api, isWailsAvailable } from '@/api/index'
import FileList from '@/components/FileList.vue'
import FileTypeFilter from '@/components/FileTypeFilter.vue'
import SearchBar from '@/components/SearchBar.vue'

const filesStore = useFilesStore()
const searchQuery = ref('')
const checkIntervals = new Map<number, NodeJS.Timeout>()

// Wait for Wails to be ready before loading files
const waitForWailsAndLoad = async () => {
  let attempts = 0
  const maxAttempts = 20 // Wait up to 10 seconds (20 * 500ms)

  const checkInterval = setInterval(() => {
    attempts++

    if (isWailsAvailable()) {
      clearInterval(checkInterval)
      console.log('Wails is ready, loading files...')
      filesStore.loadFiles('all')
    } else if (attempts >= maxAttempts) {
      clearInterval(checkInterval)
      console.warn('Wails not available after timeout, trying to load anyway...')
      filesStore.loadFiles('all')
    }
  }, 500)
}

onMounted(() => {
  // Wait for Wails backend to be ready before loading files
  waitForWailsAndLoad()
})

onUnmounted(() => {
  // Clean up all file update check intervals
  checkIntervals.forEach((interval) => {
    clearInterval(interval)
  })
  checkIntervals.clear()
})

const handleFilter = (fileType: string) => {
  filesStore.loadFiles(fileType)
  searchQuery.value = ''
}

const handleSearch = (query: string) => {
  if (query.trim()) {
    filesStore.searchFiles(query)
  } else {
    filesStore.loadFiles('all')
  }
}

const handleClearSearch = () => {
  filesStore.clearSearch()
  searchQuery.value = ''
}

const handleOpenFile = async (id: number) => {
  try {
    await api.file.open(id)

    // Start periodic checking for file updates
    startFileUpdateCheck(id)
  } catch (error) {
    console.error('Failed to open file:', error)
    alert('打开文件失败')
  }
}

// Start periodic checking for file updates
const startFileUpdateCheck = (fileId: number) => {
  // Clear existing interval for this file if any
  if (checkIntervals.has(fileId)) {
    clearInterval(checkIntervals.get(fileId))
  }

  // Store original file info for comparison
  let originalFile: any = null

  // Check immediately first
  checkFileUpdate(fileId, originalFile).then(updatedFile => {
    if (updatedFile) {
      originalFile = updatedFile
    }
  })

  // Then check every 3 seconds
  const interval = setInterval(async () => {
    const updatedFile = await checkFileUpdate(fileId, originalFile)
    if (updatedFile) {
      originalFile = updatedFile
    }
  }, 3000)

  checkIntervals.set(fileId, interval)

  // Auto-stop checking after 5 minutes (file likely closed by then)
  setTimeout(() => {
    stopFileUpdateCheck(fileId)
  }, 5 * 60 * 1000)
}

// Check if file has been updated
const checkFileUpdate = async (fileId: number, originalFile: any | null): Promise<any> => {
  try {
    const updatedFile = await api.file.refresh(fileId)
    if (!updatedFile) {
      return null
    }

    // If this is the first check, store the file info
    if (!originalFile) {
      return updatedFile
    }

    // Compare with original file to detect changes
    const hasChanges =
      updatedFile.fileName !== originalFile.fileName ||
      updatedFile.fileSize !== originalFile.fileSize

    if (hasChanges) {
      console.log('File updated:', updatedFile.fileName, updatedFile.fileSize)
      // Refresh file list
      if (searchQuery.value) {
        filesStore.searchFiles(searchQuery.value)
      } else {
        filesStore.loadFiles(filesStore.currentFileType)
      }
      return updatedFile
    }

    return null
  } catch (error: any) {
    // If file no longer exists (likely renamed), stop checking and refresh list
    if (error.message && error.message.includes('file no longer exists')) {
      console.log('File no longer exists, stopping check and refreshing list')
      stopFileUpdateCheck(fileId)
      // Refresh file list to show updated info
      if (searchQuery.value) {
        filesStore.searchFiles(searchQuery.value)
      } else {
        filesStore.loadFiles(filesStore.currentFileType)
      }
    } else {
      console.error('Failed to check file update:', error)
    }
    return null
  }
}

// Stop checking for file updates
const stopFileUpdateCheck = (fileId: number) => {
  if (checkIntervals.has(fileId)) {
    clearInterval(checkIntervals.get(fileId))
    checkIntervals.delete(fileId)
  }
}

const handleClickFile = (file: any) => {
  // Handle file click (can be used for future features like file details view)
  console.log('File clicked:', file.fileName)
}

const handleDeleteFile = async (id: number) => {
  // Refresh the file list after deletion
  if (searchQuery.value) {
    filesStore.searchFiles(searchQuery.value)
  } else {
    filesStore.loadFiles(filesStore.currentFileType)
  }
}

const handleRetry = () => {
  if (searchQuery.value) {
    filesStore.searchFiles(searchQuery.value)
  } else {
    filesStore.loadFiles(filesStore.currentFileType)
  }
}
</script>

<style scoped>
.files-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: var(--bg-color);
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background-color: var(--surface-color);
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 24px;
}

.header h1 {
  font-size: 24px;
  font-weight: 600;
  margin: 0;
  color: var(--text-color);
}

.header-actions {
  display: flex;
  gap: 8px;
}

.content {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 16px 24px;
  overflow: hidden;
  min-height: 0;
}

.filters-section {
  margin-bottom: 12px;
  flex-shrink: 0;
}

.search-section {
  margin-bottom: 16px;
  flex-shrink: 0;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: var(--text-color);
  gap: 16px;
}

.error-state {
  display: flex;
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

.error-state h3 {
  font-size: 20px;
  margin: 0;
}

.error-state p {
  font-size: 14px;
  margin: 0;
  opacity: 0.8;
  max-width: 400px;
}

@media (max-width: 768px) {
  .header {
    flex-direction: column;
    gap: 12px;
    padding: 12px 16px;
  }

  .header-left {
    width: 100%;
    justify-content: space-between;
    gap: 12px;
  }

  .header h1 {
    font-size: 20px;
  }

  .header-actions {
    width: 100%;
    justify-content: flex-end;
  }

  .content {
    padding: 12px 16px;
  }
}

@media (max-width: 480px) {
  .header {
    padding: 10px 12px;
  }

  .header-left {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .header-actions {
    width: 100%;
    flex-direction: column;
  }

  .header-actions .btn {
    width: 100%;
  }

  .content {
    padding: 10px 12px;
  }

  .filters-section {
    margin-bottom: 8px;
  }

  .search-section {
    margin-bottom: 12px;
  }
}
</style>
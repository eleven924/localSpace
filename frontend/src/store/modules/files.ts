import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { api } from '@/api/index'
import type { File as LibraryFile, FileListResponse } from '@/types'

export const useFilesStore = defineStore('files', () => {
  const files = ref<LibraryFile[]>([])
  const currentPage = ref(1)
  const pageSize = ref(50)
  const totalFiles = ref(0)
  const currentFileType = ref('all')
  const currentCollectionId = ref<number | 'all' | 'unsorted'>('all')
  const searchQuery = ref('')
  const selectedFileIds = ref<number[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const resolveCollectionId = (): number => {
    if (currentCollectionId.value === 'all') {
      return 0
    }
    if (currentCollectionId.value === 'unsorted') {
      return -1
    }
    return currentCollectionId.value
  }

  const loadFiles = async (
    fileType: string = currentFileType.value,
    page: number = currentPage.value,
    pageSizeValue: number = pageSize.value
  ) => {
    loading.value = true
    error.value = null
    currentFileType.value = fileType
    currentPage.value = page
    pageSize.value = pageSizeValue
    searchQuery.value = ''

    try {
      const resp: FileListResponse = await api.file.list(
        page,
        pageSizeValue,
        fileType,
        resolveCollectionId()
      )
      files.value = resp.items
      totalFiles.value = resp.total
    } catch (err) {
      error.value = err instanceof Error ? err.message : '加载文件失败'
      files.value = []
      totalFiles.value = 0
    } finally {
      loading.value = false
    }
  }

  const searchFiles = async (
    query: string,
    page: number = 1,
    pageSizeValue: number = pageSize.value
  ) => {
    loading.value = true
    error.value = null
    searchQuery.value = query
    currentPage.value = page
    pageSize.value = pageSizeValue

    try {
      const resp: FileListResponse = await api.file.search(query, page, pageSizeValue)
      files.value = resp.items
      totalFiles.value = resp.total
    } catch (err) {
      error.value = err instanceof Error ? err.message : '搜索失败'
      files.value = []
      totalFiles.value = 0
    } finally {
      loading.value = false
    }
  }

  const clearSearch = async () => {
    searchQuery.value = ''
    await loadFiles(currentFileType.value, 1, pageSize.value)
  }

  const setCurrentFileType = (fileType: string) => {
    currentFileType.value = fileType
    currentCollectionId.value = 'all'
    currentPage.value = 1
    return loadFiles(fileType, 1, pageSize.value)
  }

  const setCurrentCollectionId = (collectionId: number | 'all' | 'unsorted') => {
    currentCollectionId.value = collectionId
    currentPage.value = 1
    return loadFiles(currentFileType.value, 1, pageSize.value)
  }

  const setPageSize = (size: number) => {
    pageSize.value = size
    currentPage.value = 1
    if (searchQuery.value.trim()) {
      return searchFiles(searchQuery.value, 1, size)
    }
    return loadFiles(currentFileType.value, 1, size)
  }

  const goToPage = (page: number) => {
    currentPage.value = page
    if (searchQuery.value.trim()) {
      return searchFiles(searchQuery.value, page, pageSize.value)
    }
    return loadFiles(currentFileType.value, page, pageSize.value)
  }

  const refreshCurrent = async () => {
    if (searchQuery.value.trim()) {
      await searchFiles(searchQuery.value, currentPage.value, pageSize.value)
    } else {
      await loadFiles(currentFileType.value, currentPage.value, pageSize.value)
    }
  }

  const toggleSelection = (id: number) => {
    const index = selectedFileIds.value.indexOf(id)
    if (index >= 0) {
      selectedFileIds.value.splice(index, 1)
    } else {
      selectedFileIds.value.push(id)
    }
  }

  const clearSelection = () => {
    selectedFileIds.value = []
  }

  const selectAll = () => {
    selectedFileIds.value = files.value.map((file) => file.id)
  }

  const totalPages = computed(() => Math.max(1, Math.ceil(totalFiles.value / pageSize.value)))

  return {
    files,
    currentPage,
    pageSize,
    totalFiles,
    totalPages,
    currentFileType,
    currentCollectionId,
    searchQuery,
    selectedFileIds,
    loading,
    error,
    loadFiles,
    searchFiles,
    clearSearch,
    setCurrentFileType,
    setCurrentCollectionId,
    setPageSize,
    goToPage,
    refreshCurrent,
    toggleSelection,
    clearSelection,
    selectAll,
  }
})

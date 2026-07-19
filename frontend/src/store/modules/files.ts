import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { api } from '@/api/index'
import type { CollectionSummary, File as LibraryFile } from '@/types'
import { UNSORTED_COLLECTION_KEY, UNSORTED_COLLECTION_LABEL } from '@/utils/constants'

export const useFilesStore = defineStore('files', () => {
  const files = ref<LibraryFile[]>([])
  const allFiles = ref<LibraryFile[]>([])
  const currentFileType = ref('all')
  const currentCollectionName = ref('all')
  const searchQuery = ref('')
  const loading = ref(false)
  const error = ref<string | null>(null)

  const normalizeCollectionName = (collectionName?: string) => {
    return collectionName?.trim() || UNSORTED_COLLECTION_KEY
  }

  const getFilesForCurrentType = () => {
    if (currentFileType.value === 'all') {
      return allFiles.value
    }

    return allFiles.value.filter((file) => file.fileType === currentFileType.value)
  }

  const applyFilters = () => {
    let nextFiles = [...getFilesForCurrentType()]

    if (currentCollectionName.value !== 'all') {
      nextFiles = nextFiles.filter((file) => {
        return normalizeCollectionName(file.collectionName) === currentCollectionName.value
      })
    }

    files.value = nextFiles
  }

  const ensureCollectionSelectionIsValid = () => {
    if (currentCollectionName.value === 'all') {
      return
    }

    const hasCollection = getFilesForCurrentType().some((file) => {
      return normalizeCollectionName(file.collectionName) === currentCollectionName.value
    })

    if (!hasCollection) {
      currentCollectionName.value = 'all'
    }
  }

  const loadFiles = async (fileType: string = currentFileType.value, page: number = 1) => {
    loading.value = true
    error.value = null
    currentFileType.value = fileType

    try {
      console.log(`Loading files: fileType=${fileType}, page=${page}`)

      const timeoutPromise = new Promise((_, reject) => {
        setTimeout(() => {
          reject(new Error('API call timeout'))
        }, 15000)
      })

      const result = await Promise.race([
        api.file.list(page, 50, 'all'),
        timeoutPromise,
      ]) as LibraryFile[]

      console.log(`Loaded ${result.length} files`)

      allFiles.value = result
      ensureCollectionSelectionIsValid()
      applyFilters()
    } catch (err) {
      error.value = err instanceof Error ? err.message : '加载文件失败'
      console.error('Failed to load files:', err)
      files.value = []
      allFiles.value = []
    } finally {
      loading.value = false
    }
  }

  const searchFiles = async (query: string) => {
    loading.value = true
    error.value = null
    searchQuery.value = query

    try {
      console.log(`Searching files: query=${query}`)

      const timeoutPromise = new Promise((_, reject) => {
        setTimeout(() => {
          reject(new Error('API call timeout'))
        }, 15000)
      })

      const result = await Promise.race([
        api.file.search(query),
        timeoutPromise,
      ]) as LibraryFile[]

      console.log(`Found ${result.length} files`)

      allFiles.value = result
      ensureCollectionSelectionIsValid()
      applyFilters()
    } catch (err) {
      error.value = err instanceof Error ? err.message : '搜索失败'
      console.error('Failed to search files:', err)
      files.value = []
      allFiles.value = []
    } finally {
      loading.value = false
    }
  }

  const clearSearch = async () => {
    searchQuery.value = ''
    await loadFiles(currentFileType.value)
  }

  const setCurrentFileType = (fileType: string) => {
    currentFileType.value = fileType
    ensureCollectionSelectionIsValid()
    applyFilters()
  }

  const setCurrentCollectionName = (collectionName: string) => {
    currentCollectionName.value = collectionName
    applyFilters()
  }

  const fileTypeCounts = computed(() => {
    const counts: Record<string, number> = {}

    allFiles.value.forEach((file) => {
      counts[file.fileType] = (counts[file.fileType] || 0) + 1
    })

    counts.all = allFiles.value.length

    return counts
  })

  const collectionCounts = computed(() => {
    const counts: Record<string, number> = {
      all: getFilesForCurrentType().length,
    }

    getFilesForCurrentType().forEach((file) => {
      const collectionKey = normalizeCollectionName(file.collectionName)
      counts[collectionKey] = (counts[collectionKey] || 0) + 1
    })

    return counts
  })

  const collectionSummaries = computed<CollectionSummary[]>(() => {
    const summaryMap = new Map<string, CollectionSummary>()

    getFilesForCurrentType().forEach((file) => {
      const collectionKey = normalizeCollectionName(file.collectionName)
      const existing = summaryMap.get(collectionKey) || {
        value: collectionKey,
        label: collectionKey === UNSORTED_COLLECTION_KEY ? UNSORTED_COLLECTION_LABEL : collectionKey,
        count: 0,
        fileTypes: {},
        latestModifiedAt: file.modifiedAt || file.createdAt || '',
      }

      existing.count += 1
      existing.fileTypes[file.fileType] = (existing.fileTypes[file.fileType] || 0) + 1

      const candidateDate = file.modifiedAt || file.createdAt || ''
      if (candidateDate && (!existing.latestModifiedAt || candidateDate > existing.latestModifiedAt)) {
        existing.latestModifiedAt = candidateDate
      }

      summaryMap.set(collectionKey, existing)
    })

    return Array.from(summaryMap.values()).sort((left, right) => {
      if (left.value === UNSORTED_COLLECTION_KEY) {
        return 1
      }

      if (right.value === UNSORTED_COLLECTION_KEY) {
        return -1
      }

      if (right.count !== left.count) {
        return right.count - left.count
      }

      return left.label.localeCompare(right.label, 'zh-CN')
    })
  })

  return {
    files,
    allFiles,
    currentFileType,
    currentCollectionName,
    searchQuery,
    loading,
    error,
    loadFiles,
    searchFiles,
    clearSearch,
    setCurrentFileType,
    setCurrentCollectionName,
    fileTypeCounts,
    collectionCounts,
    collectionSummaries,
  }
})

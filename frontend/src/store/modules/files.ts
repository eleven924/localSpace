import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/api/index'

export const useFilesStore = defineStore('files', () => {
  const files = ref<File[]>([])
  const allFiles = ref<File[]>([])
  const currentFileType = ref('all')
  const searchQuery = ref('')
  const loading = ref(false)
  const error = ref<string | null>(null)

  const loadFiles = async (fileType: string = 'all', page: number = 1) => {
    loading.value = true
    error.value = null

    try {
      console.log(`Loading files: fileType=${fileType}, page=${page}`)

      // Add timeout protection
      const timeoutPromise = new Promise((_, reject) => {
        setTimeout(() => {
          reject(new Error('API call timeout'))
        }, 15000) // 15 second timeout to give backend more time
      })

      const result = await Promise.race([
        api.file.list(page, 50, 'all'),
        timeoutPromise
      ]) as File[]

      console.log(`Loaded ${result.length} files`)

      // Save all files for counting
      allFiles.value = result

      // Filter files based on fileType
      if (fileType === 'all') {
        files.value = result
      } else {
        files.value = result.filter(file => file.fileType === fileType)
      }

      currentFileType.value = fileType
    } catch (err) {
      error.value = err instanceof Error ? err.message : '加载文件失败'
      console.error('Failed to load files:', err)

      // Set empty array on error to prevent hanging
      files.value = []
      allFiles.value = []
    } finally {
      loading.value = false
    }
  }

  const searchFiles = async (query: string) => {
    loading.value = true
    error.value = null

    try {
      console.log(`Searching files: query=${query}`)

      // Add timeout protection
      const timeoutPromise = new Promise((_, reject) => {
        setTimeout(() => {
          reject(new Error('API call timeout'))
        }, 15000) // 15 second timeout to give backend more time
      })

      const result = await Promise.race([
        api.file.search(query),
        timeoutPromise
      ]) as File[]

      console.log(`Found ${result.length} files`)

      files.value = result
      searchQuery.value = query
    } catch (err) {
      error.value = err instanceof Error ? err.message : '搜索失败'
      console.error('Failed to search files:', err)

      // Set empty array on error to prevent hanging
      files.value = []
    } finally {
      loading.value = false
    }
  }

  const clearSearch = async () => {
    searchQuery.value = ''
    await loadFiles(currentFileType.value)
  }

  // Calculate file type counts from all files
  const fileTypeCounts = computed(() => {
    const counts: Record<string, number> = {}

    allFiles.value.forEach((file) => {
      const type = file.fileType
      counts[type] = (counts[type] || 0) + 1
    })

    // Add 'all' count
    counts.all = allFiles.value.length

    return counts
  })

  return {
    files,
    allFiles,
    currentFileType,
    searchQuery,
    loading,
    error,
    loadFiles,
    searchFiles,
    clearSearch,
    fileTypeCounts
  }
})
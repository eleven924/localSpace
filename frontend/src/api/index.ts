// Wails runtime API
declare global {
  interface Window {
    go?: {
      app?: {
        App?: {
          // Configuration
          InitConfig: (storagePath: string) => Promise<string>
          GetConfig: (key: string) => Promise<string>
          UpdateConfig: (key: string, value: string) => Promise<string>

          // File types
          GetFileTypes: () => Promise<any[]>
          ParseFileType: (extension: string) => Promise<any>

          // Collection operations
          GetCollections: () => Promise<any[]>
          GetCollectionFilterCounts: (fileType: string) => Promise<any>
          AddCollection: (name: string) => Promise<number>
          RemoveCollection: (id: number) => Promise<string>

          // File operations
          ImportFile: (filePath: string, fileName: string, description: string, tags: string[]) => Promise<string>
          ImportFileWithKeywords: (filePath: string, fileName: string, description: string, tags: string[], keywords: string) => Promise<string>
          ImportFileWithMetadata: (filePath: string, fileName: string, description: string, tags: string[], keywords: string, collectionId: number) => Promise<string>
          GetFiles: (page: number, pageSize: number, fileType: string, collectionId: number) => Promise<any>

          SearchFiles: (query: string, page: number, pageSize: number) => Promise<any>
          GetFile: (id: number) => Promise<any>
          DeleteFile: (id: number) => Promise<string>
          OpenFile: (id: number) => Promise<string>
          OpenFileWithPreferredApp: (id: number) => Promise<string>
          OpenFileWithSystemDefault: (id: number) => Promise<string>
          OpenFileLocation: (id: number) => Promise<string>
          RefreshFile: (id: number) => Promise<any>
          RenameFile: (id: number, newName: string) => Promise<string>
          UpdateFileMetadata: (id: number, tags: string[], description: string, collectionId: number) => Promise<void>
          BatchUpdateFilesCollection: (ids: number[], collectionId: number) => Promise<any>
          BatchDeleteFiles: (ids: number[]) => Promise<any>

          // AI operations
          GetAIAnalysis: (
            fileName: string,
            fileType: string,
            userKeywords: string,
            userTags: string[],
            userDescription: string
          ) => Promise<any>

          // Storage
          GetStorageDirectories: () => Promise<any[]>
          GetMasterDirectories: () => Promise<any[]>
          AddStorageDirectory: (path: string, fileType: string) => Promise<string>
          AddMasterDirectory: (path: string, maxSizeGB: number) => Promise<string>
          RemoveStorageDirectory: (id: number) => Promise<string>
          ToggleStorageDirectory: (id: number, isActive: boolean) => Promise<string>
          SetDefaultMasterDirectory: (id: number) => Promise<string>
          CheckStorageSpace: (fileSize: number) => Promise<boolean>
          CheckPathConflict: (path: string, excludeID: number) => Promise<boolean>
          CheckMasterStorageSpace: (masterID: number, fileSize: number) => Promise<boolean>
          GetMasterStoragePathForFile: (masterID: number, fileType: string, collectionName: string, fileName: string) => Promise<string>

          // Theme
          GetThemeConfig: () => Promise<any>
          UpdateThemeConfig: (config: any) => Promise<string>
          GetOpenWithConfig: () => Promise<any>
          UpdateOpenWithConfig: (config: any) => Promise<string>
          GetStorageLayoutConfig: () => Promise<any>
          UpdateStorageLayoutConfig: (config: any) => Promise<string>

          // AI Config
          GetAIConfig: () => Promise<any>
          UpdateAIConfig: (config: any) => Promise<string>

          // File selection
          SelectFile: () => Promise<string>
          SelectFiles: () => Promise<any[]>
          SelectDirectory: () => Promise<string>
          SelectExecutable: () => Promise<string>
          ConfirmQuit: () => Promise<void>
          GetExitGuardSnapshot: () => Promise<any>

          // Jobs
          SubmitBatchImportJob: (payload: any) => Promise<any>
          SubmitSingleImportJob: (
            filePath: string,
            fileName: string,
            description: string,
            tags: string[],
            keywords: string,
            collectionId: number
          ) => Promise<any>
          GetActiveJobs: () => Promise<any[]>
          GetResumableJobs: () => Promise<any[]>
          ListJobs: (page: number, pageSize: number, jobType: string) => Promise<any>
          GetJob: (jobID: number) => Promise<any>
          ResumeJob: (jobID: number) => Promise<string>
          CancelJob: (jobID: number) => Promise<string>
          GetJobRetentionConfig: () => Promise<any>
          UpdateJobRetentionConfig: (config: any) => Promise<string>
          SubmitJobCleanup: () => Promise<any>
          DeleteJobRecord: (id: number) => Promise<string>

          // Metadata
          GetFileMetadata: (filePath: string, fileType: string) => Promise<any>

          // Thumbnail
          GenerateThumbnail: (fileID: number) => Promise<string>
          GetThumbnailCacheInfo: () => Promise<[number, number, number]>
          ClearThumbnailCache: () => Promise<string>
          GetThumbnailCacheStats: () => Promise<any>
        }
      }
    }
  }
}

interface SingleImportJobRequest {
  filePath: string
  fileName: string
  description: string
  tags: string[]
  keywords: string
  collectionId?: number
}

// Helper function to safely call Wails API
function safeWailsCall<T>(
  fn: () => Promise<T>,
  fallbackValue: T,
  errorMessage: string
): Promise<T> {
  return new Promise((resolve, reject) => {
    try {
      // Check if Wails is available
      if (!window.go || !window.go.app || !window.go.app.App) {
        console.warn(`Wails API not available: ${errorMessage}`)
        console.warn('This is expected in browser development. Using fallback values.')
        resolve(fallbackValue)
        return
      }

      // Call the Wails function
      fn()
        .then(resolve)
        .catch((error) => {
          console.error(`Wails API error: ${errorMessage}`, error)
          reject(error)
        })
    } catch (error) {
      console.error(`Unexpected error calling Wails API: ${errorMessage}`, error)
      reject(error)
    }
  })
}

// Check if running in Wails environment
export function isWailsAvailable(): boolean {
  return !!(window.go && window.go.app && window.go.app.App)
}

// API wrapper with safe calls
export const api = {
  config: {
    init: (storagePath: string) =>
      safeWailsCall(
        () => window.go!.app!.App.InitConfig(storagePath),
        'success',
        'InitConfig'
      ),
    get: (key: string) =>
      safeWailsCall(
        () => window.go!.app!.App.GetConfig(key),
        '',
        `GetConfig(${key})`
      ),
    update: (key: string, value: string) =>
      safeWailsCall(
        () => window.go!.app!.App.UpdateConfig(key, value),
        'success',
        `UpdateConfig(${key})`
      ),
  },

  fileType: {
    getAll: () =>
      safeWailsCall(
        () => window.go!.app!.App.GetFileTypes(),
        [],
        'GetFileTypes'
      ),
    parse: (extension: string) =>
      safeWailsCall(
        () => window.go!.app!.App.ParseFileType(extension),
        null,
        `ParseFileType(${extension})`
      ),
  },

  collection: {
    getAll: () =>
      safeWailsCall(
        () => window.go!.app!.App.GetCollections(),
        [],
        'GetCollections'
      ),
    getFilterCounts: (fileType: string) =>
      safeWailsCall(
        () => window.go!.app!.App.GetCollectionFilterCounts(fileType),
        { total: 0, unsorted: 0, collections: {} },
        `GetCollectionFilterCounts(${fileType})`
      ),
    add: (name: string) =>
      safeWailsCall(
        () => window.go!.app!.App.AddCollection(name),
        0,
        'AddCollection'
      ),
    remove: (id: number) =>
      safeWailsCall(
        () => window.go!.app!.App.RemoveCollection(id),
        'success',
        'RemoveCollection'
      ),
  },

  file: {
    import: (filePath: string, fileName: string, description: string, tags: string[]) =>
      // Import should throw errors on failure - use direct call
      new Promise((resolve, reject) => {
        try {
          if (!window.go || !window.go.app || !window.go.app.App) {
            reject(new Error('Wails API not available'))
            return
          }
          window.go!.app!.App.ImportFile(filePath, fileName, description, tags)
            .then(() => resolve('success'))
            .catch(reject)
        } catch (error) {
          reject(error)
        }
      }),
    importWithKeywords: (filePath: string, fileName: string, description: string, tags: string[], keywords: string) =>
      new Promise((resolve, reject) => {
        try {
          if (!window.go || !window.go.app || !window.go.app.App) {
            reject(new Error('Wails API not available'))
            return
          }
          window.go!.app!.App.ImportFileWithKeywords(filePath, fileName, description, tags, keywords)
            .then(() => resolve('success'))
            .catch(reject)
        } catch (error) {
          reject(error)
        }
      }),
    list: (page: number, pageSize: number, fileType: string, collectionId?: number) =>
      safeWailsCall(
        () => window.go!.app!.App.GetFiles(page, pageSize, fileType, collectionId ?? 0),
        { items: [], page, pageSize, total: 0 },
        `GetFiles(${fileType})`
      ),
    search: (query: string, page: number = 1, pageSize: number = 50) =>
      safeWailsCall(
        () => window.go!.app!.App.SearchFiles(query, page, pageSize),
        { items: [], page, pageSize, total: 0 },
        `SearchFiles(${query})`
      ),
    get: (id: number) =>
      safeWailsCall(
        () => window.go!.app!.App.GetFile(id),
        null,
        `GetFile(${id})`
      ),
    delete: (id: number) =>
      safeWailsCall(
        () => window.go!.app!.App.DeleteFile(id),
        'success',
        `DeleteFile(${id})`
      ),
    open: (id: number) =>
      new Promise((resolve, reject) => {
        try {
          if (!window.go || !window.go.app || !window.go.app.App) {
            reject(new Error('Wails API not available'))
            return
          }
          window.go!.app!.App.OpenFile(id)
            .then(() => resolve('success'))
            .catch(reject)
        } catch (error) {
          reject(error)
        }
      }),
    importWithMetadata: (filePath: string, fileName: string, description: string, tags: string[], keywords: string, collectionId?: number) =>
      new Promise((resolve, reject) => {
        try {
          if (!window.go || !window.go.app || !window.go.app.App) {
            reject(new Error('Wails API not available'))
            return
          }
          window.go!.app!.App.ImportFileWithMetadata(filePath, fileName, description, tags, keywords, collectionId || 0)
            .then(() => resolve('success'))
            .catch(reject)
        } catch (error) {
          reject(error)
        }
      }),
    openPreferred: (id: number) =>
      new Promise((resolve, reject) => {
        try {
          if (!window.go || !window.go.app || !window.go.app.App) {
            reject(new Error('Wails API not available'))
            return
          }
          window.go!.app!.App.OpenFileWithPreferredApp(id)
            .then(() => resolve('success'))
            .catch(reject)
        } catch (error) {
          reject(error)
        }
      }),
    openSystemDefault: (id: number) =>
      new Promise((resolve, reject) => {
        try {
          if (!window.go || !window.go.app || !window.go.app.App) {
            reject(new Error('Wails API not available'))
            return
          }
          window.go!.app!.App.OpenFileWithSystemDefault(id)
            .then(() => resolve('success'))
            .catch(reject)
        } catch (error) {
          reject(error)
        }
      }),
    openLocation: (id: number) =>
      safeWailsCall(
        () => window.go!.app!.App.OpenFileLocation(id),
        'success',
        `OpenFileLocation(${id})`
      ),
    refresh: (id: number) =>
      safeWailsCall(
        () => window.go!.app!.App.RefreshFile(id),
        null,
        `RefreshFile(${id})`
      ),
    rename: (id: number, newName: string) =>
      new Promise((resolve, reject) => {
        try {
          if (!window.go || !window.go.app || !window.go.app.App) {
            reject(new Error('Wails API not available'))
            return
          }
          window.go!.app!.App.RenameFile(id, newName)
            .then(() => resolve('success'))
            .catch(reject)
        } catch (error) {
          reject(error)
        }
      }),
    updateMetadata: (id: number, tags: string[], description: string, collectionId?: number) =>
      new Promise<void>((resolve, reject) => {
        try {
          if (!window.go || !window.go.app || !window.go.app.App) {
            reject(new Error('Wails API not available'))
            return
          }
          window.go!.app!.App.UpdateFileMetadata(id, tags, description, collectionId || 0)
            .then(() => resolve())
            .catch(reject)
        } catch (error) {
          reject(error)
        }
      }),
    batchUpdateCollection: (ids: number[], collectionId?: number) =>
      safeWailsCall(
        () => window.go!.app!.App.BatchUpdateFilesCollection(ids, collectionId || 0),
        { successCount: 0, failedCount: 0, failedItems: [] },
        'BatchUpdateFilesCollection'
      ),
    batchDelete: (ids: number[]) =>
      safeWailsCall(
        () => window.go!.app!.App.BatchDeleteFiles(ids),
        { successCount: 0, failedCount: 0, failedItems: [] },
        'BatchDeleteFiles'
      ),
  },

  ai: {
    analyze: (fileName: string, fileType: string, userKeywords = '', userTags: string[] = [], userDescription = '') =>
      safeWailsCall(
        () => window.go!.app!.App.GetAIAnalysis(fileName, fileType, userKeywords, userTags, userDescription),
        null,
        `GetAIAnalysis(${fileName})`
      ),
    getConfig: () =>
      safeWailsCall(
        () => window.go!.app!.App.GetAIConfig(),
        { enabled: false },
        'GetAIConfig'
      ),
    updateConfig: (config: any) =>
      safeWailsCall(
        () => window.go!.app!.App.UpdateAIConfig(config),
        'success',
        'UpdateAIConfig'
      ),
  },

  storage: {
    getDirectories: () =>
      safeWailsCall(
        () => window.go!.app!.App.GetStorageDirectories(),
        [],
        'GetStorageDirectories'
      ),
    addDirectory: (path: string, fileType: string) =>
      safeWailsCall(
        () => window.go!.app!.App.AddStorageDirectory(path, fileType),
        'success',
        'AddStorageDirectory'
      ),
    removeDirectory: (id: number) =>
      safeWailsCall(
        () => window.go!.app!.App.RemoveStorageDirectory(id),
        'success',
        `RemoveStorageDirectory(${id})`
      ),
    toggleDirectory: (id: number, isActive: boolean) =>
      safeWailsCall(
        () => window.go!.app!.App.ToggleStorageDirectory(id, isActive),
        'success',
        `ToggleStorageDirectory(${id}, ${isActive})`
      ),
    checkSpace: (fileSize: number) =>
      safeWailsCall(
        () => window.go!.app!.App.CheckStorageSpace(fileSize),
        true,
        `CheckStorageSpace(${fileSize})`
      ),
    getMasterDirectories: () =>
      safeWailsCall(
        () => window.go!.app!.App.GetMasterDirectories(),
        [],
        'GetMasterDirectories'
      ),
    addMasterDirectory: (path: string, maxSizeGB: number) =>
      safeWailsCall(
        () => window.go!.app!.App.AddMasterDirectory(path, maxSizeGB),
        'success',
        'AddMasterDirectory'
      ),
    setDefaultMasterDirectory: (id: number) =>
      safeWailsCall(
        () => window.go!.app!.App.SetDefaultMasterDirectory(id),
        'success',
        `SetDefaultMasterDirectory(${id})`
      ),
    checkPathConflict: (path: string, excludeID: number) =>
      safeWailsCall(
        () => window.go!.app!.App.CheckPathConflict(path, excludeID),
        false,
        `CheckPathConflict(${path})`
      ),
    checkMasterSpace: (masterID: number, fileSize: number) =>
      safeWailsCall(
        () => window.go!.app!.App.CheckMasterStorageSpace(masterID, fileSize),
        true,
        `CheckMasterStorageSpace(${masterID}, ${fileSize})`
      ),
    getMasterPathForFile: (masterID: number, fileType: string, collectionName: string, fileName: string) =>
      safeWailsCall(
        () => window.go!.app!.App.GetMasterStoragePathForFile(masterID, fileType, collectionName, fileName),
        '',
        `GetMasterStoragePathForFile(${masterID}, ${fileType}, ${collectionName}, ${fileName})`
      ),
  },

  theme: {
    getConfig: () =>
      safeWailsCall(
        () => window.go!.app!.App.GetThemeConfig(),
        { themeMode: 'light', primaryColor: '#2196F3', backgroundImage: '' },
        'GetThemeConfig'
      ),
    updateConfig: (config: any) =>
      safeWailsCall(
        () => window.go!.app!.App.UpdateThemeConfig(config),
        'success',
        'UpdateThemeConfig'
      ),
  },

  openWith: {
    getConfig: () =>
      safeWailsCall(
        () => window.go!.app!.App.GetOpenWithConfig(),
        { byFileType: {}, byExtension: {} },
        'GetOpenWithConfig'
      ),
    updateConfig: (config: any) =>
      new Promise((resolve, reject) => {
        try {
          if (!window.go || !window.go.app || !window.go.app.App) {
            reject(new Error('Wails API not available'))
            return
          }
          window.go!.app!.App.UpdateOpenWithConfig(config)
            .then(() => resolve('success'))
            .catch(reject)
        } catch (error) {
          reject(error)
        }
      }),
  },

  storageLayout: {
    getConfig: () =>
      safeWailsCall(
        () => window.go!.app!.App.GetStorageLayoutConfig(),
        { strategy: 'type_collection', unsortedFolderName: '_unsorted', sanitizeFolderName: true },
        'GetStorageLayoutConfig'
      ),
    updateConfig: (config: any) =>
      new Promise((resolve, reject) => {
        try {
          if (!window.go || !window.go.app || !window.go.app.App) {
            reject(new Error('Wails API not available'))
            return
          }
          window.go!.app!.App.UpdateStorageLayoutConfig(config)
            .then(() => resolve('success'))
            .catch(reject)
        } catch (error) {
          reject(error)
        }
      }),
  },

  system: {
    selectFile: () =>
      safeWailsCall(
        () => window.go!.app!.App.SelectFile(),
        '',
        'SelectFile'
      ),
    selectFiles: () =>
      safeWailsCall(
        () => window.go!.app!.App.SelectFiles(),
        [],
        'SelectFiles'
      ),
    selectDirectory: () =>
      safeWailsCall(
        () => window.go!.app!.App.SelectDirectory(),
        '',
        'SelectDirectory'
      ),
    selectExecutable: () =>
      safeWailsCall(
        () => window.go!.app!.App.SelectExecutable(),
        '',
        'SelectExecutable'
      ),
    confirmQuit: () =>
      safeWailsCall(
        () => window.go!.app!.App.ConfirmQuit(),
        undefined,
        'ConfirmQuit'
      ),
    getExitGuardSnapshot: () =>
      safeWailsCall(
        () => window.go!.app!.App.GetExitGuardSnapshot(),
        { hasProtectedJobs: false, total: 0, statusCounts: {}, jobs: [] },
        'GetExitGuardSnapshot'
      ),
    getMetadata: (filePath: string, fileType: string) =>
      safeWailsCall(
        () => window.go!.app!.App.GetFileMetadata(filePath, fileType),
        {},
        `GetFileMetadata(${filePath})`
      ),
  },

  jobs: {
    submitBatchImportJob: (payload: any) =>
      new Promise((resolve, reject) => {
        try {
          if (!window.go || !window.go.app || !window.go.app.App) {
            reject(new Error('Wails API not available'))
            return
          }
          window.go!.app!.App.SubmitBatchImportJob(payload)
            .then(resolve)
            .catch(reject)
        } catch (error) {
          reject(error)
        }
      }),
    submitSingleImportJob: (payload: SingleImportJobRequest) =>
      new Promise((resolve, reject) => {
        try {
          if (!window.go || !window.go.app || !window.go.app.App) {
            reject(new Error('Wails API not available'))
            return
          }
          window.go!.app!.App.SubmitSingleImportJob(
            payload.filePath,
            payload.fileName,
            payload.description,
            payload.tags,
            payload.keywords,
            payload.collectionId || 0
          )
            .then(resolve)
            .catch(reject)
        } catch (error) {
          reject(error)
        }
      }),
    getActiveJobs: () =>
      safeWailsCall(
        () => window.go!.app!.App.GetActiveJobs(),
        [],
        'GetActiveJobs'
      ),
    getResumableJobs: () =>
      safeWailsCall(
        () => window.go!.app!.App.GetResumableJobs(),
        [],
        'GetResumableJobs'
      ),
    listJobs: (page: number, pageSize: number, jobType = '') =>
      safeWailsCall(
        () => window.go!.app!.App.ListJobs(page, pageSize, jobType),
        { items: [], page, pageSize, total: 0 },
        `ListJobs(${page}, ${pageSize}, ${jobType})`
      ),
    getJob: (jobID: number) =>
      safeWailsCall(
        () => window.go!.app!.App.GetJob(jobID),
        null,
        `GetJob(${jobID})`
      ),
    resume: (jobID: number) =>
      new Promise((resolve, reject) => {
        try {
          if (!window.go || !window.go.app || !window.go.app.App) {
            reject(new Error('Wails API not available'))
            return
          }
          window.go!.app!.App.ResumeJob(jobID)
            .then(() => resolve('success'))
            .catch(reject)
        } catch (error) {
          reject(error)
        }
      }),
    cancel: (jobID: number) =>
      new Promise((resolve, reject) => {
        try {
          if (!window.go || !window.go.app || !window.go.app.App) {
            reject(new Error('Wails API not available'))
            return
          }
          window.go!.app!.App.CancelJob(jobID)
            .then(() => resolve('success'))
            .catch(reject)
        } catch (error) {
          reject(error)
        }
      }),
    getRetentionConfig: () =>
      safeWailsCall(
        () => window.go!.app!.App.GetJobRetentionConfig(),
        { id: 1, enabled: false, maxCount: 0, maxDays: 0, updatedAt: '' },
        'GetJobRetentionConfig'
      ),
    updateRetentionConfig: (config: import('./types').JobRetentionConfig) =>
      safeWailsCall(
        () => window.go!.app!.App.UpdateJobRetentionConfig(config),
        'success',
        'UpdateJobRetentionConfig'
      ),
    submitCleanup: () =>
      safeWailsCall(
        () => window.go!.app!.App.SubmitJobCleanup(),
        null,
        'SubmitJobCleanup'
      ),
    deleteRecord: (id: number) =>
      safeWailsCall(
        () => window.go!.app!.App.DeleteJobRecord(id),
        'success',
        'DeleteJobRecord'
      ),
  },

  thumbnail: {
    generate: (fileID: number) =>
      safeWailsCall(
        () => window.go!.app!.App.GenerateThumbnail(fileID),
        '',
        `GenerateThumbnail(${fileID})`
      ),
    getCacheInfo: () =>
      safeWailsCall(
        () => window.go!.app!.App.GetThumbnailCacheInfo(),
        [0, 0, 0],
        'GetThumbnailCacheInfo'
      ),
    clearCache: () =>
      safeWailsCall(
        () => window.go!.app!.App.ClearThumbnailCache(),
        'success',
        'ClearThumbnailCache'
      ),
    getCacheStats: () =>
      safeWailsCall(
        () => window.go!.app!.App.GetThumbnailCacheStats(),
        {},
        'GetThumbnailCacheStats'
      ),
  },
}

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

          // File operations
          ImportFile: (filePath: string, fileName: string, description: string, tags: string[]) => Promise<string>
          GetFiles: (page: number, pageSize: number, fileType: string) => Promise<any[]>
          SearchFiles: (query: string) => Promise<any[]>
          GetFile: (id: number) => Promise<any>
          DeleteFile: (id: number) => Promise<string>
          OpenFile: (id: number) => Promise<string>
          OpenFileLocation: (id: number) => Promise<string>
          RefreshFile: (id: number) => Promise<any>
          RenameFile: (id: number, newName: string) => Promise<string>

          // AI operations
          GetAIAnalysis: (fileName: string, fileType: string) => Promise<any>

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
          GetMasterStoragePathForFile: (masterID: number, fileType: string, fileName: string) => Promise<string>

          // Theme
          GetThemeConfig: () => Promise<any>
          UpdateThemeConfig: (config: any) => Promise<string>

          // AI Config
          GetAIConfig: () => Promise<any>
          UpdateAIConfig: (config: any) => Promise<string>

          // File selection
          SelectFile: () => Promise<string>
          SelectDirectory: () => Promise<string>

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

// Helper function to safely call Wails API
function safeWailsCall<T>(fn: () => Promise<T>, fallbackValue: T, errorMessage: string): Promise<T> {
  return new Promise((resolve) => {
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
          resolve(fallbackValue)
        })
    } catch (error) {
      console.error(`Unexpected error calling Wails API: ${errorMessage}`, error)
      resolve(fallbackValue)
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
    list: (page: number, pageSize: number, fileType: string) =>
      safeWailsCall(
        () => window.go!.app!.App.GetFiles(page, pageSize, fileType),
        [],
        `GetFiles(${fileType})`
      ),
    search: (query: string) =>
      safeWailsCall(
        () => window.go!.app!.App.SearchFiles(query),
        [],
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
      safeWailsCall(
        () => window.go!.app!.App.OpenFile(id),
        'success',
        `OpenFile(${id})`
      ),
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
    getMasterPathForFile: (masterID: number, fileType: string, fileName: string) =>
      safeWailsCall(
        () => window.go!.app!.App.GetMasterStoragePathForFile(masterID, fileType, fileName),
        '',
        `GetMasterStoragePathForFile(${masterID}, ${fileType}, ${fileName})`
      ),
  },

  theme: {
    getConfig: () =>
      safeWailsCall(
        () => window.go!.app!.App.GetThemeConfig(),
        { mode: 'light', primaryColor: '#2196F3' },
        'GetThemeConfig'
      ),
    updateConfig: (config: any) =>
      safeWailsCall(
        () => window.go!.app!.App.UpdateThemeConfig(config),
        'success',
        'UpdateThemeConfig'
      ),
  },

  system: {
    selectFile: () =>
      safeWailsCall(
        () => window.go!.app!.App.SelectFile(),
        '',
        'SelectFile'
      ),
    selectDirectory: () =>
      safeWailsCall(
        () => window.go!.app!.App.SelectDirectory(),
        '',
        'SelectDirectory'
      ),
    getMetadata: (filePath: string, fileType: string) =>
      safeWailsCall(
        () => window.go!.app!.App.GetFileMetadata(filePath, fileType),
        {},
        `GetFileMetadata(${filePath})`
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
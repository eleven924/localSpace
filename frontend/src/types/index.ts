// File types
export interface File {
  id: number
  fileName: string
  originalName: string
  collectionName: string
  filePath: string
  fileType: string
  fileSubType: string
  fileSize: number
  tags: string[]
  description: string
  metadata: Metadata
  thumbnail: string
  checksum: string
  isDeleted: boolean
  deletedAt: string
  createdAt: string
  modifiedAt: string
}

export interface Metadata {
  width?: number
  height?: number
  duration?: number
  pageCount?: number
  author?: string
  title?: string
}

// Configuration types
export interface Config {
  key: string
  value: string
  description?: string
}

export interface AIConfig {
  id: number
  apiKey: string
  model: string
  baseURL: string
  enabled: boolean
  enableAgent?: boolean
  enableWebSearch?: boolean
  maxTokens?: number
  timeout?: number
  webSearchProvider?: string
  webSearchBaseURL?: string
  webSearchAPIKey?: string
  webSearchTimeout?: number
  webSearchMaxResults?: number
}

export interface ThemeConfig {
  id: number
  themeMode: 'light' | 'dark'
  primaryColor: string
  backgroundColor?: string
  backgroundImage: string
}

export interface OpenWithConfig {
  byFileType: Record<string, string>
  byExtension: Record<string, string>
}

export interface StorageLayoutConfig {
  strategy: 'type_only' | 'type_collection'
  unsortedFolderName: string
  sanitizeFolderName: boolean
}

export interface StorageDir {
  id: number
  path: string
  fileType: string
  currentSize: number
  maxSize: number
  isActive: boolean
  createdAt: string
}

// File type
export interface FileType {
  id: number
  name: string
  displayName: string
  extensions: string
  subTypes: string[]
  createdAt: string
}

// AI Analysis
export interface AIAnalysis {
  tags: string[]
  description: string
}

// API Response types
export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: string
}

export interface CollectionSummary {
  value: string
  label: string
  count: number
  fileTypes: Record<string, number>
  latestModifiedAt: string
}

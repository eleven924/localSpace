// File types
export interface File {
  id: number
  fileName: string
  originalName: string
  collectionName: string
  collectionId?: number
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

export interface Collection {
  id: number
  name: string
  createdAt: string
  updatedAt: string
}

export interface FileListResponse {
  items: File[]
  page: number
  pageSize: number
  total: number
}

export interface CollectionFilterCounts {
  total: number
  unsorted: number
  collections: Record<number, number>
}

export interface BatchMoveFailedItem {
  fileId: number
  fileName: string
  error: string
}

export interface BatchMoveResult {
  successCount: number
  failedCount: number
  failedItems: BatchMoveFailedItem[]
}

export interface BatchDeleteFailedItem {
  fileId: number
  fileName: string
  error: string
}

export interface BatchDeleteResult {
  successCount: number
  failedCount: number
  failedItems: BatchDeleteFailedItem[]
}

export interface JobRetentionConfig {
  id: number
  enabled: boolean
  maxCount: number
  maxDays: number
  updatedAt: string
}

export * from './jobs'

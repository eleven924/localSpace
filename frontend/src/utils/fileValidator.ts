// 文件验证工具函数

export interface FileValidationResult {
  valid: boolean
  errors: string[]
  warnings: string[]
}

export interface FileValidationConfig {
  maxSize?: number // bytes
  allowedExtensions?: string[]
  minNameLength?: number
  maxNameLength?: number
  minTags?: number
  maxTags?: number
}

const DEFAULT_CONFIG: FileValidationConfig = {
  maxSize: 10 * 1024 * 1024 * 1024, // 10GB
  allowedExtensions: [], // Empty means all extensions allowed
  minNameLength: 1,
  maxNameLength: 255,
  minTags: 0,
  maxTags: 10,
}

// 支持的文件类型映射
const FILE_TYPE_EXTENSIONS: Record<string, string[]> = {
  video: ['mp4', 'avi', 'mkv', 'mov', 'wmv', 'flv', 'webm', 'm4v'],
  document: ['pdf', 'doc', 'docx', 'txt', 'md', 'rtf', 'odt', 'xls', 'xlsx', 'ppt', 'pptx'],
  music: ['mp3', 'wav', 'flac', 'aac', 'ogg', 'wma', 'm4a', 'opus'],
  game: ['iso', 'zip', 'rar', '7z', 'exe', 'msi', 'dmg', 'app'],
  image: ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg', 'ico'],
  installer: ['exe', 'msi', 'dmg', 'pkg', 'deb', 'rpm', 'apk'],
}

/**
 * 验证文件基本信息
 */
export function validateFile(
  file: {
    name: string
    size: number
    path: string
  },
  config: FileValidationConfig = {}
): FileValidationResult {
  const mergedConfig = { ...DEFAULT_CONFIG, ...config }
  const errors: string[] = []
  const warnings: string[] = []

  // 验证文件大小
  if (mergedConfig.maxSize && file.size > mergedConfig.maxSize) {
    errors.push(`文件大小超出限制（最大 ${formatFileSize(mergedConfig.maxSize)}）`)
  }

  if (file.size === 0) {
    errors.push('文件大小不能为 0')
  }

  // 验证文件名
  const fileName = file.name.trim()

  if (!fileName) {
    errors.push('文件名不能为空')
  } else {
    if (mergedConfig.minNameLength && fileName.length < mergedConfig.minNameLength) {
      errors.push(`文件名长度不能少于 ${mergedConfig.minNameLength} 个字符`)
    }

    if (mergedConfig.maxNameLength && fileName.length > mergedConfig.maxNameLength) {
      warnings.push(`文件名过长（最大 ${mergedConfig.maxNameLength} 个字符），可能会被截断`)
    }

    // 检查文件名中是否包含非法字符
    const invalidChars = /[<>:"/\\|?*\x00-\x1F]/
    if (invalidChars.test(fileName)) {
      errors.push('文件名包含非法字符')
    }

    // 检查是否为系统保留名称
    const reservedNames = ['CON', 'PRN', 'AUX', 'NUL', 'COM1', 'COM2', 'COM3', 'COM4', 'COM5', 'COM6', 'COM7', 'COM8', 'COM9', 'LPT1', 'LPT2', 'LPT3', 'LPT4', 'LPT5', 'LPT6', 'LPT7', 'LPT8', 'LPT9']
    const nameWithoutExt = fileName.split('.')[0].toUpperCase()
    if (reservedNames.includes(nameWithoutExt)) {
      errors.push('文件名不能使用系统保留名称')
    }
  }

  // 验证文件扩展名
  const extension = file.path.split('.').pop()?.toLowerCase()

  if (extension && mergedConfig.allowedExtensions.length > 0) {
    if (!mergedConfig.allowedExtensions.includes(extension)) {
      errors.push(`不支持的文件类型（.${extension}）`)
    }
  }

  return {
    valid: errors.length === 0,
    errors,
    warnings,
  }
}

/**
 * 验证文件元数据
 */
export function validateFileMetadata(
  metadata: {
    fileName: string
    tags: string[]
    description: string
  },
  config: FileValidationConfig = {}
): FileValidationResult {
  const mergedConfig = { ...DEFAULT_CONFIG, ...config }
  const errors: string[] = []
  const warnings: string[] = []

  // 验证文件名
  if (!metadata.fileName.trim()) {
    errors.push('文件名不能为空')
  } else if (mergedConfig.minNameLength && metadata.fileName.trim().length < mergedConfig.minNameLength) {
    errors.push(`文件名长度不能少于 ${mergedConfig.minNameLength} 个字符`)
  }

  // 验证标签
  const validTags = metadata.tags.filter(tag => tag.trim().length > 0)

  if (mergedConfig.minTags && validTags.length < mergedConfig.minTags) {
    warnings.push(`建议添加至少 ${mergedConfig.minTags} 个标签`)
  }

  if (mergedConfig.maxTags && validTags.length > mergedConfig.maxTags) {
    errors.push(`标签数量不能超过 ${mergedConfig.maxTags} 个`)
  }

  // 验证标签内容
  validTags.forEach((tag, index) => {
    if (tag.length > 50) {
      errors.push(`第 ${index + 1} 个标签过长（最大 50 个字符）`)
    }

    if (!/^[一-龥a-zA-Z0-9_\- ]+$/.test(tag)) {
      warnings.push(`第 ${index + 1} 个标签包含特殊字符，可能影响搜索`)
    }
  })

  // 验证描述
  if (metadata.description.length > 500) {
    errors.push('描述长度不能超过 500 个字符')
  }

  if (metadata.description.length > 0 && metadata.description.trim().length === 0) {
    warnings.push('描述只包含空格字符')
  }

  return {
    valid: errors.length === 0,
    errors,
    warnings,
  }
}

/**
 * 根据文件扩展名获取文件类型
 */
export function getFileTypeByExtension(extension: string): string {
  const ext = extension.toLowerCase()

  for (const [type, extensions] of Object.entries(FILE_TYPE_EXTENSIONS)) {
    if (extensions.includes(ext)) {
      return type
    }
  }

  return 'document' // 默认类型
}

/**
 * 检查文件扩展名是否支持
 */
export function isExtensionSupported(extension: string): boolean {
  const ext = extension.toLowerCase()

  for (const extensions of Object.values(FILE_TYPE_EXTENSIONS)) {
    if (extensions.includes(ext)) {
      return true
    }
  }

  return false
}

/**
 * 获取所有支持的扩展名
 */
export function getAllSupportedExtensions(): string[] {
  const extensions: string[] = []

  for (const exts of Object.values(FILE_TYPE_EXTENSIONS)) {
    extensions.push(...exts)
  }

  return [...new Set(extensions)].sort()
}

/**
 * 格式化文件大小
 */
function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'

  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))

  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}

/**
 * 验证导入前的完整信息
 */
export function validateImport(
  file: {
    name: string
    size: number
    path: string
  },
  metadata: {
    fileName: string
    tags: string[]
    description: string
  },
  config: FileValidationConfig = {}
): FileValidationResult {
  const fileValidation = validateFile(file, config)
  const metadataValidation = validateFileMetadata(metadata, config)

  return {
    valid: fileValidation.valid && metadataValidation.valid,
    errors: [...fileValidation.errors, ...metadataValidation.errors],
    warnings: [...fileValidation.warnings, ...metadataValidation.warnings],
  }
}
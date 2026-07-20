export const FILE_TYPES = [
  { value: 'all', label: '全部', icon: '◌' },
  { value: 'video', label: '视频', icon: '▶' },
  { value: 'document', label: '文档', icon: '▤' },
  { value: 'music', label: '音频', icon: '♪' },
  { value: 'archive', label: '压缩包', icon: '⬚' },
  { value: 'installer', label: '安装包', icon: '⌘' },
  { value: 'image', label: '图片', icon: '◧' },
] as const

export const EXTENSION_TO_TYPE: Record<string, string> = {
  '.mp4': 'video',
  '.avi': 'video',
  '.mkv': 'video',
  '.mov': 'video',
  '.wmv': 'video',
  '.flv': 'video',
  '.webm': 'video',
  '.m4v': 'video',
  '.pdf': 'document',
  '.doc': 'document',
  '.docx': 'document',
  '.xls': 'document',
  '.xlsx': 'document',
  '.ppt': 'document',
  '.pptx': 'document',
  '.txt': 'document',
  '.rtf': 'document',
  '.odt': 'document',
  '.ods': 'document',
  '.odp': 'document',
  '.md': 'document',
  '.mp3': 'music',
  '.wav': 'music',
  '.flac': 'music',
  '.aac': 'music',
  '.ogg': 'music',
  '.m4a': 'music',
  '.wma': 'music',
  '.zip': 'archive',
  '.rar': 'archive',
  '.7z': 'archive',
  '.tar': 'archive',
  '.gz': 'archive',
  '.exe': 'installer',
  '.app': 'installer',
  '.msi': 'installer',
  '.apk': 'installer',
  '.ipa': 'installer',
  '.deb': 'installer',
  '.rpm': 'installer',
  '.pkg': 'installer',
  '.dmg': 'installer',
  '.iso': 'installer',
  '.img': 'installer',
  '.vdi': 'installer',
  '.vmdk': 'installer',
  '.jpg': 'image',
  '.jpeg': 'image',
  '.png': 'image',
  '.gif': 'image',
  '.bmp': 'image',
  '.webp': 'image',
  '.svg': 'image',
  '.ico': 'image',
}

export const getFileTypeLabel = (fileType: string): string => {
  const fileTypeInfo = FILE_TYPES.find((type) => type.value === fileType)
  return fileTypeInfo?.label || fileType
}

export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'

  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))

  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}

export const formatDate = (dateString: string): string => {
  const date = new Date(dateString)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export const THEME_MODES = [
  { value: 'light', label: '浅色', icon: '◐' },
  { value: 'dark', label: '深色', icon: '◑' },
] as const

export const DEFAULT_SETTINGS = {
  AI_MODEL: 'gpt-3.5-turbo',
  AI_BASE_URL: 'https://api.openai.com/v1',
  PRIMARY_COLOR: '#6F8FD8',
}

export const UNSORTED_COLLECTION_KEY = '__unsorted__'
export const UNSORTED_COLLECTION_LABEL = '未分配合集'

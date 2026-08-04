export interface BatchImportFileInput {
  sourcePath: string
  displayName: string
}

export interface BatchImportJobRequest {
  files: BatchImportFileInput[]
  sharedTags: string[]
  sharedDescription: string
  collectionId?: number
  enableAIGeneratedTags: boolean
  enableAIGeneratedDescription: boolean
}

export interface SingleImportJobRequest {
  filePath: string
  fileName: string
  description: string
  tags: string[]
  keywords: string
  collectionId?: number
}

export interface BatchImportFailedItem {
  sourcePath: string
  displayName: string
  error: string
  fileName?: string
  fileId?: number
}

export interface BatchImportResult {
  successCount: number
  failedCount: number
  failedItems: BatchImportFailedItem[]
}

export interface CleanupJobResult {
  deleted: number
}

export type JobResult = BatchImportResult | CleanupJobResult

export interface Job {
  id: number
  jobType: string
  status: string
  title: string
  payload: BatchImportJobRequest | null
  result: JobResult | null
  progressTotal: number
  progressCompleted: number
  progressMessage: string
  exclusiveKey: string
  canResume: boolean
  startedAt: string
  heartbeatAt: string
  finishedAt: string
  timeoutAt: string
  errorMessage: string
}

export interface JobListResponse {
  items: Job[]
  page: number
  pageSize: number
  total: number
}

export interface ExitGuardJobSnapshot {
  id: number
  jobType: string
  status: string
  title: string
  progressTotal: number
  progressCompleted: number
  progressMessage: string
  canResume: boolean
}

export interface ExitGuardSnapshot {
  hasProtectedJobs: boolean
  total: number
  statusCounts: Record<string, number>
  jobs: ExitGuardJobSnapshot[]
}

export interface SelectedFile {
  name: string
  path: string
  size: number
}

export const ACTIVE_JOB_STATUSES = ['pending', 'running', 'recovering'] as const
export const RESUMABLE_JOB_STATUSES = ['awaiting_resume', 'timed_out'] as const
export const CLOSE_PROTECTED_JOB_STATUSES = ['pending', 'running', 'recovering'] as const

export function jobProgressPercent(job: Job | null | undefined): number {
  if (!job || !job.progressTotal) {
    return 0
  }

  return Math.max(0, Math.min(100, Math.round((job.progressCompleted / job.progressTotal) * 100)))
}

export function jobStatusLabel(status: string): string {
  const labels: Record<string, string> = {
    pending: '排队中',
    running: '执行中',
    recovering: '恢复中',
    awaiting_resume: '等待继续',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消',
    timed_out: '超时',
    cleanup_failed: '清理失败',
  }

  return labels[status] || status
}

export interface NotificationEvent {
  id: string
  type: 'job_completed' | 'job_failed' | 'job_cancelled'
  title: string
  message: string
  payload: {
    jobId?: number
    jobType?: string
    [key: string]: any
  }
  createdAt: string
  read: boolean
}

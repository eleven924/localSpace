import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { EventsOn } from '../../../wailsjs/runtime/runtime'
import { api, isWailsAvailable } from '@/api'
import type { BatchImportJobRequest, Job, JobListResponse } from '@/types/jobs'

const DEFAULT_PAGE_SIZE = 10

const TERMINAL_STATUSES = ['completed', 'failed', 'cancelled', 'timed_out', 'cleanup_failed']
const ACTIVE_STATUSES = ['pending', 'running', 'recovering']
const RESUMABLE_STATUSES = ['awaiting_resume', 'timed_out']

export const useJobsStore = defineStore('jobs', () => {
  const jobs = ref<Job[]>([])
  const page = ref(1)
  const pageSize = ref(DEFAULT_PAGE_SIZE)
  const total = ref(0)
  const activeJobs = ref<Job[]>([])
  const loading = ref(false)
  const initialized = ref(false)

  const totalRunningCount = computed(() => activeJobs.value.length)

  const summaryJobs = computed(() => {
    const map = new Map<number, Job>()
    for (const job of activeJobs.value) {
      map.set(job.id, job)
    }
    for (const job of jobs.value) {
      if (!map.has(job.id)) {
        map.set(job.id, job)
      }
    }
    return Array.from(map.values()).slice(0, 3)
  })

  const upsertJob = (target: Job[], incoming: Job) => {
    const index = target.findIndex((item) => item.id === incoming.id)
    if (index >= 0) {
      target.splice(index, 1, incoming)
      return
    }
    target.unshift(incoming)
  }

  const removeJob = (target: Job[], jobId: number) => {
    const index = target.findIndex((item) => item.id === jobId)
    if (index >= 0) {
      target.splice(index, 1)
    }
  }

  const syncJob = (job: Job | null | undefined) => {
    if (!job) {
      return
    }

    upsertJob(jobs.value, job)

    if (ACTIVE_STATUSES.includes(job.status) || RESUMABLE_STATUSES.includes(job.status)) {
      upsertJob(activeJobs.value, job)
    } else {
      removeJob(activeJobs.value, job.id)
    }
  }

  const initialize = async () => {
    if (initialized.value) {
      return
    }
    initialized.value = true

    if (isWailsAvailable()) {
      EventsOn('job:created', (job: Job) => syncJob(job))
      EventsOn('job:updated', (job: Job) => syncJob(job))
      EventsOn('job:completed', (job: Job) => syncJob(job))
      EventsOn('job:failed', (job: Job) => syncJob(job))
      EventsOn('job:needs-resume', (job: Job) => syncJob(job))
    }

    await Promise.all([loadActiveJobs(), loadJobs(1, pageSize.value)])
  }

  const loadJobs = async (p = page.value, ps = pageSize.value) => {
    loading.value = true
    try {
      page.value = p
      pageSize.value = ps
      const resp: JobListResponse = await api.jobs.listJobs(p, ps, '')
      jobs.value = resp.items
      total.value = resp.total
    } finally {
      loading.value = false
    }
  }

  const loadActiveJobs = async () => {
    loading.value = true
    try {
      activeJobs.value = await api.jobs.getActiveJobs()
    } finally {
      loading.value = false
    }
  }

  const submitBatchImportJob = async (payload: BatchImportJobRequest) => {
    const job = await api.jobs.submitBatchImportJob(payload)
    syncJob(job)
    return job
  }

  const resumeJob = async (jobId: number) => {
    await api.jobs.resume(jobId)
    await Promise.all([loadActiveJobs(), loadJobs(page.value, pageSize.value)])
  }

  const cancelJob = async (jobId: number) => {
    await api.jobs.cancel(jobId)
    await Promise.all([loadActiveJobs(), loadJobs(page.value, pageSize.value)])
  }

  const deleteJobRecord = async (jobId: number) => {
    await api.jobs.deleteRecord(jobId)
    await Promise.all([loadActiveJobs(), loadJobs(page.value, pageSize.value)])
  }

  const refreshJob = async (jobId: number) => {
    const job = await api.jobs.getJob(jobId)
    syncJob(job)
    return job
  }

  return {
    jobs,
    page,
    pageSize,
    total,
    activeJobs,
    totalRunningCount,
    summaryJobs,
    loading,
    initialize,
    loadJobs,
    loadActiveJobs,
    submitBatchImportJob,
    resumeJob,
    cancelJob,
    deleteJobRecord,
    refreshJob,
  }
})

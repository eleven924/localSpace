import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { EventsOn } from '../../../wailsjs/runtime/runtime'
import { api, isWailsAvailable } from '@/api'
import type { BatchImportJobRequest, Job, JobListResponse } from '@/types/jobs'

const DEFAULT_HISTORY_PAGE_SIZE = 10

export const useJobsStore = defineStore('jobs', () => {
  const activeJobs = ref<Job[]>([])
  const resumableJobs = ref<Job[]>([])
  const history = ref<JobListResponse>({
    items: [],
    page: 1,
    pageSize: DEFAULT_HISTORY_PAGE_SIZE,
    total: 0,
  })
  const loading = ref(false)
  const initialized = ref(false)

  const activeJob = computed(() => activeJobs.value[0] ?? null)
  const resumableJob = computed(() => resumableJobs.value[0] ?? null)
  const totalRunningCount = computed(() => activeJobs.value.length + resumableJobs.value.length)

  const summaryJobs = computed(() => {
    const priorityJobs = [...resumableJobs.value, ...activeJobs.value]
    const map = new Map<number, Job>()
    for (const job of priorityJobs) {
      map.set(job.id, job)
    }
    for (const job of history.value.items) {
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

  const upsertHistoryJob = (incoming: Job) => {
    const items = [...history.value.items]
    const index = items.findIndex((item) => item.id === incoming.id)
    if (index >= 0) {
      items.splice(index, 1, incoming)
    } else {
      items.unshift(incoming)
      if (items.length > history.value.pageSize) {
        items.length = history.value.pageSize
      }
    }
    history.value = {
      ...history.value,
      items,
    }
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

    upsertHistoryJob(job)

    if (['pending', 'running', 'recovering'].includes(job.status)) {
      upsertJob(activeJobs.value, job)
      removeJob(resumableJobs.value, job.id)
      return
    }

    if (['awaiting_resume', 'timed_out'].includes(job.status)) {
      upsertJob(resumableJobs.value, job)
      removeJob(activeJobs.value, job.id)
      return
    }

    removeJob(activeJobs.value, job.id)
    removeJob(resumableJobs.value, job.id)
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

    await Promise.all([
      loadActiveJobs(),
      loadResumableJobs(),
      loadJobHistory(1, history.value.pageSize),
    ])
  }

  const loadActiveJobs = async () => {
    loading.value = true
    try {
      activeJobs.value = await api.jobs.getActiveJobs()
      activeJobs.value.forEach(upsertHistoryJob)
    } finally {
      loading.value = false
    }
  }

  const loadResumableJobs = async () => {
    loading.value = true
    try {
      resumableJobs.value = await api.jobs.getResumableJobs()
      resumableJobs.value.forEach(upsertHistoryJob)
    } finally {
      loading.value = false
    }
  }

  const loadJobHistory = async (page = history.value.page, pageSize = history.value.pageSize, jobType = '') => {
    loading.value = true
    try {
      history.value = await api.jobs.listJobs(page, pageSize, jobType)
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
    await Promise.all([
      loadActiveJobs(),
      loadResumableJobs(),
      loadJobHistory(history.value.page, history.value.pageSize),
    ])
  }

  const cancelJob = async (jobId: number) => {
    await api.jobs.cancel(jobId)
    await Promise.all([
      loadActiveJobs(),
      loadResumableJobs(),
      loadJobHistory(history.value.page, history.value.pageSize),
    ])
  }

  const refreshJob = async (jobId: number) => {
    const job = await api.jobs.getJob(jobId)
    syncJob(job)
    return job
  }

  return {
    activeJobs,
    resumableJobs,
    history,
    activeJob,
    resumableJob,
    totalRunningCount,
    summaryJobs,
    loading,
    initialize,
    loadActiveJobs,
    loadResumableJobs,
    loadJobHistory,
    submitBatchImportJob,
    resumeJob,
    cancelJob,
    refreshJob,
  }
})

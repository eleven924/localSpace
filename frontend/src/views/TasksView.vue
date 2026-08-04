<template>
  <div class="page-shell tasks-view">
    <AppHeader />

    <div class="page-content">
      <div class="page-stack">
        <section class="page-topbar tasks-topbar">
          <div class="page-topbar-copy">
            <h1 class="page-topbar-title">任务中心</h1>
            <p class="page-topbar-note">查看任务进度、恢复中断任务，并分页管理历史记录。</p>
          </div>

          <div class="page-topbar-meta">
            <div class="metric-badge">
              <strong>{{ jobsStore.activeJobs.length }}</strong>
              <span>活动任务</span>
            </div>
            <div class="metric-badge">
              <strong>{{ jobsStore.total }}</strong>
              <span>总记录</span>
            </div>
          </div>
        </section>

        <section class="section-panel section-shell">
          <div class="section-head">
            <div>
              <p class="eyebrow">All Tasks</p>
              <h2 class="section-title">任务列表</h2>
              <p class="section-copy">支持分页浏览，失败任务可展开查看详情。</p>
            </div>

            <div class="section-actions">
              <button class="btn secondary" type="button" @click="refreshAll">刷新状态</button>
              <button class="btn primary" type="button" @click="router.push('/import')">返回导入</button>
            </div>
          </div>

          <div v-if="jobsStore.loading && jobsStore.jobs.length === 0" class="empty-surface">
            正在加载任务列表...
          </div>

          <div v-else-if="jobsStore.jobs.length === 0" class="empty-surface">
            暂时还没有任务记录。
          </div>

          <div v-else class="task-table-wrap">
            <table class="task-table">
              <thead>
                <tr>
                  <th class="col-title">任务</th>
                  <th class="col-time">开始时间</th>
                  <th class="col-time">结束时间</th>
                  <th class="col-progress">进度</th>
                  <th class="col-status">状态</th>
                  <th class="col-actions">操作</th>
                </tr>
              </thead>
              <tbody>
                <template v-for="job in jobsStore.jobs" :key="job.id">
                  <tr class="task-row" :class="{ 'is-expanded': expandedJobs.has(job.id) }">
                    <td class="col-title">
                      <strong>{{ job.title }}</strong>
                      <p class="task-message">{{ job.progressMessage || jobStatusLabel(job.status) }}</p>
                    </td>
                    <td class="col-time">{{ formatDate(job.startedAt) || '—' }}</td>
                    <td class="col-time">{{ formatDate(job.finishedAt) || '—' }}</td>
                    <td class="col-progress">
                      <ProgressBar :percentage="jobProgressPercent(job)" :show-label="false" small />
                      <span class="progress-count">{{ job.progressCompleted }} / {{ job.progressTotal }}</span>
                    </td>
                    <td class="col-status">
                      <span class="status-pill" :class="`status-${job.status}`">{{ jobStatusLabel(job.status) }}</span>
                    </td>
                    <td class="col-actions">
                      <div class="action-cell">
                        <button
                          v-if="canResume(job)"
                          class="btn primary small"
                          type="button"
                          @click="jobsStore.resumeJob(job.id)"
                        >
                          继续
                        </button>
                        <button
                          v-if="canDelete(job)"
                          class="btn ghost small"
                          type="button"
                          @click="confirmDelete(job.id)"
                        >
                          删除
                        </button>
                        <button
                          v-if="hasDetails(job)"
                          class="btn ghost small"
                          type="button"
                          @click="toggleExpanded(job.id)"
                        >
                          {{ expandedJobs.has(job.id) ? '收起' : '详情' }}
                        </button>
                      </div>
                    </td>
                  </tr>
                  <tr v-if="expandedJobs.has(job.id)" class="task-detail-row">
                    <td colspan="6">
                      <div class="task-detail-card">
                        <div v-if="job.errorMessage" class="detail-block">
                          <strong>错误信息</strong>
                          <p class="error-text">{{ job.errorMessage }}</p>
                        </div>
                        <div v-if="job.result" class="detail-block">
                          <strong>执行结果</strong>
                          <template v-if="cleanupDeleted(job) !== null">
                            <p>已清理 {{ cleanupDeleted(job) }} 条任务记录</p>
                          </template>
                          <template v-else-if="batchResult(job)">
                            <p>成功 {{ batchResult(job)?.successCount }} 个 · 失败 {{ batchResult(job)?.failedCount }} 个</p>
                            <ul v-if="batchResult(job)?.failedItems?.length" class="failed-items">
                              <li v-for="(item, index) in batchResult(job)?.failedItems" :key="index">
                                {{ item.displayName || item.fileName || item.sourcePath || `文件 ${item.fileId}` }}：{{ item.error }}
                              </li>
                            </ul>
                          </template>
                        </div>
                        <div class="detail-block">
                          <strong>任务 ID</strong>
                          <p class="meta-text">{{ job.id }}</p>
                        </div>
                      </div>
                    </td>
                  </tr>
                </template>
              </tbody>
            </table>
          </div>

          <PaginationControls
            v-if="!jobsStore.loading || jobsStore.jobs.length > 0"
            :page="jobsStore.page"
            :page-size="jobsStore.pageSize"
            :total="jobsStore.total"
            :page-sizes="[10, 20, 50]"
            @change="jobsStore.loadJobs"
            @change-size="handleSizeChange"
          />
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import AppHeader from '@/components/AppHeader.vue'
import PaginationControls from '@/components/PaginationControls.vue'
import ProgressBar from '@/components/ProgressBar.vue'
import { useJobsStore } from '@/store/modules/jobs'
import { jobProgressPercent, jobStatusLabel, type BatchImportResult, type CleanupJobResult, type Job } from '@/types/jobs'
import { formatDate } from '@/utils/constants'

const jobsStore = useJobsStore()
const router = useRouter()
const expandedJobs = ref<Set<number>>(new Set())

const TERMINAL_STATUSES = ['completed', 'failed', 'cancelled', 'timed_out', 'cleanup_failed']
const RESUMABLE_STATUSES = ['awaiting_resume', 'timed_out']

const canResume = (job: Job) => RESUMABLE_STATUSES.includes(job.status)
const canDelete = (job: Job) => TERMINAL_STATUSES.includes(job.status)
const hasDetails = (job: Job) => Boolean(job.errorMessage || job.result)
const cleanupDeleted = (job: Job) => {
  // 清理任务的 result 与导入任务不同，单独读取 deleted 字段用于详情展示。
  const result = job.result as CleanupJobResult | null
  return job.jobType === 'job_cleanup' && typeof result?.deleted === 'number' ? result.deleted : null
}
const batchResult = (job: Job): BatchImportResult | null => {
  if (!job.result || job.jobType === 'job_cleanup') {
    return null
  }
  return job.result as BatchImportResult
}

const toggleExpanded = (jobId: number) => {
  if (expandedJobs.value.has(jobId)) {
    expandedJobs.value.delete(jobId)
  } else {
    expandedJobs.value.add(jobId)
  }
}

const confirmDelete = async (jobId: number) => {
  if (!window.confirm('确定删除该任务记录？')) return
  await jobsStore.deleteJobRecord(jobId)
}

const handleSizeChange = async (size: number) => {
  await jobsStore.loadJobs(1, size)
}

const refreshAll = async () => {
  await Promise.all([jobsStore.loadActiveJobs(), jobsStore.loadJobs(jobsStore.page, jobsStore.pageSize)])
}

onMounted(() => {
  void refreshAll()
})
</script>

<style scoped>
.tasks-topbar {
  align-items: center;
}

.section-shell {
  padding: 22px;
}

.section-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 14px;
  flex-wrap: wrap;
}

.section-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.empty-surface {
  margin-top: 18px;
  padding: 32px;
  text-align: center;
  border-radius: 24px;
  border: 1px solid rgba(146, 165, 192, 0.18);
  background: rgba(255, 255, 255, 0.58);
  color: var(--text-soft);
}

.task-table-wrap {
  margin-top: 18px;
  overflow-x: auto;
  border-radius: 16px;
}

.task-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.task-table th {
  padding: 12px 14px;
  text-align: left;
  font-weight: 600;
  color: var(--text-soft);
  border-bottom: 1px solid rgba(146, 165, 192, 0.18);
  background: rgba(255, 255, 255, 0.42);
  backdrop-filter: blur(8px);
  white-space: nowrap;
}

.task-row {
  background: rgba(255, 255, 255, 0.42);
  backdrop-filter: blur(8px);
  border-bottom: 1px solid rgba(146, 165, 192, 0.18);
  transition: background-color 0.2s ease;
}

[data-theme='dark'] .task-row {
  background: rgba(15, 23, 42, 0.42);
}

.task-row:hover {
  background: rgba(255, 255, 255, 0.64);
}

[data-theme='dark'] .task-row:hover {
  background: rgba(30, 41, 59, 0.58);
}

.task-row td {
  padding: 14px;
  vertical-align: middle;
}

.col-title {
  min-width: 200px;
}

.col-title strong {
  display: block;
  color: var(--text-color);
  font-size: 14px;
}

.task-message {
  margin-top: 4px;
  color: var(--text-soft);
  font-size: 12px;
}

.col-time {
  white-space: nowrap;
  color: var(--text-soft);
}

.col-progress {
  min-width: 140px;
}

.progress-count {
  display: block;
  margin-top: 4px;
  font-size: 11px;
  color: var(--text-faint);
}

.status-pill {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  background: var(--surface-muted);
  color: var(--text-soft);
}

.status-pill.status-running,
.status-pill.status-recovering,
.status-pill.status-pending {
  background: rgba(45, 140, 240, 0.12);
  color: #2d8cf0;
}

.status-pill.status-completed {
  background: rgba(82, 196, 26, 0.12);
  color: #52c41a;
}

.status-pill.status-failed,
.status-pill.status-timed_out,
.status-pill.status-cleanup_failed {
  background: rgba(204, 108, 108, 0.12);
  color: #a85d5d;
}

.status-pill.status-cancelled,
.status-pill.status-awaiting_resume {
  background: rgba(250, 173, 20, 0.12);
  color: #d48806;
}

.col-actions {
  white-space: nowrap;
}

.action-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn.small {
  min-height: 28px;
  padding: 4px 10px;
  font-size: 12px;
  border-radius: 8px;
}

.task-detail-row td {
  padding: 0;
  border-bottom: 1px solid rgba(146, 165, 192, 0.18);
}

.task-detail-card {
  padding: 16px 20px;
  background: rgba(248, 250, 252, 0.72);
  backdrop-filter: blur(8px);
}

[data-theme='dark'] .task-detail-card {
  background: rgba(15, 23, 42, 0.58);
}

.detail-block {
  margin-bottom: 14px;
}

.detail-block:last-child {
  margin-bottom: 0;
}

.detail-block strong {
  display: block;
  font-size: 12px;
  color: var(--text-soft);
  margin-bottom: 6px;
}

.detail-block p {
  margin: 0;
  color: var(--text-color);
  line-height: 1.6;
}

.detail-block .error-text {
  color: #a85d5d;
}

.detail-block .meta-text {
  color: var(--text-faint);
  font-size: 12px;
}

.failed-items {
  margin: 8px 0 0;
  padding-left: 18px;
  color: var(--text-soft);
  font-size: 12px;
  line-height: 1.7;
}

@media (max-width: 980px) {
  .section-head {
    flex-direction: column;
    align-items: flex-start;
  }

  .task-table th,
  .task-row td {
    padding: 10px 12px;
  }
}

@media (max-width: 760px) {
  .task-table {
    display: block;
  }

  .task-table thead {
    display: none;
  }

  .task-row,
  .task-detail-row {
    display: block;
    margin-bottom: 10px;
    border-radius: 14px;
    border: 1px solid rgba(146, 165, 192, 0.18);
  }

  .task-row td,
  .task-detail-row td {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px 14px;
    border: none;
  }

  .task-row td::before {
    content: attr(data-label);
    font-weight: 600;
    color: var(--text-soft);
  }

  .col-title,
  .col-progress,
  .col-actions {
    min-width: 0;
  }

  .action-cell {
    justify-content: flex-end;
  }
}
</style>

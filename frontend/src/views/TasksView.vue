<template>
  <div class="page-shell tasks-view">
    <AppHeader />

    <div class="page-content">
      <div class="page-stack">
        <section class="page-topbar tasks-topbar">
          <div class="page-topbar-copy">
            <h1 class="page-topbar-title">任务中心</h1>
            <p class="page-topbar-note">查看导入任务进度、恢复中断任务，并分页回看历史记录。</p>
          </div>

          <div class="page-topbar-meta">
            <div class="metric-badge">
              <strong>{{ jobsStore.totalRunningCount }}</strong>
              <span>活动任务</span>
            </div>
            <div class="metric-badge">
              <strong>{{ resumableCount }}</strong>
              <span>等待继续</span>
            </div>
            <div class="metric-badge">
              <strong>{{ jobsStore.history.total }}</strong>
              <span>历史记录</span>
            </div>
          </div>
        </section>

        <section class="section-panel section-shell">
          <div class="section-head">
            <div>
              <p class="eyebrow">Running Now</p>
              <h2 class="section-title">当前任务</h2>
              <p class="section-copy">右上角入口用于快速查看，这里保留完整操作。</p>
            </div>
            <div class="section-actions">
              <button class="btn secondary" type="button" @click="refreshAll">刷新状态</button>
              <button class="btn primary" type="button" @click="router.push('/import')">返回导入</button>
            </div>
          </div>

          <div v-if="activeAndResumable.length === 0" class="empty-surface">
            当前没有需要关注的任务。
          </div>

          <div v-else class="active-grid">
            <article v-for="job in activeAndResumable" :key="job.id" class="job-card">
              <div class="job-card-head">
                <div>
                  <strong>{{ job.title }}</strong>
                  <p>{{ job.progressMessage || jobStatusLabel(job.status) }}</p>
                </div>
                <span class="status-pill">{{ jobStatusLabel(job.status) }}</span>
              </div>

              <ProgressBar :percentage="jobProgressPercent(job)" :show-label="true" striped animated />

              <div class="job-card-foot">
                <span class="job-count">{{ job.progressCompleted }} / {{ job.progressTotal }}</span>
                <div class="job-actions">
                  <button
                    v-if="job.status === 'awaiting_resume' || job.status === 'timed_out'"
                    class="btn primary"
                    type="button"
                    @click="jobsStore.resumeJob(job.id)"
                  >
                    继续任务
                  </button>
                  <button class="btn secondary" type="button" @click="jobsStore.cancelJob(job.id)">
                    {{ job.status === 'awaiting_resume' || job.status === 'timed_out' ? '放弃并清理' : '取消任务' }}
                  </button>
                </div>
              </div>
            </article>
          </div>
        </section>

        <section class="section-panel section-shell">
          <div class="section-head history-head">
            <div>
              <p class="eyebrow">History</p>
              <h2 class="section-title">导入历史</h2>
              <p class="section-copy">支持分页浏览，并可指定每页展示数量。</p>
            </div>

            <div class="toolbar">
              <div class="range-pill">{{ showingRange }}</div>
              <label class="size-picker">
                <span>每页</span>
                <select v-model.number="pageSize" @change="changePageSize">
                  <option :value="10">10</option>
                  <option :value="20">20</option>
                  <option :value="50">50</option>
                </select>
              </label>
            </div>
          </div>

          <div v-if="jobsStore.history.items.length === 0" class="empty-surface">
            暂时还没有导入历史。
          </div>

          <div v-else class="history-list">
            <article v-for="job in jobsStore.history.items" :key="job.id" class="history-card">
              <div class="history-main">
                <div class="history-copy">
                  <div class="history-title-line">
                    <strong>{{ job.title }}</strong>
                    <span class="status-pill">{{ jobStatusLabel(job.status) }}</span>
                  </div>
                  <p>{{ job.progressMessage || '没有额外进度说明' }}</p>
                  <div class="history-meta">
                    <span>创建于 {{ formatDate(job.createdAt) }}</span>
                    <span v-if="job.finishedAt">完成于 {{ formatDate(job.finishedAt) }}</span>
                  </div>
                </div>

                <div class="history-side">
                  <span class="percent">{{ jobProgressPercent(job) }}%</span>
                  <span class="count">{{ job.progressCompleted }} / {{ job.progressTotal }}</span>
                </div>
              </div>

              <ProgressBar :percentage="jobProgressPercent(job)" :show-label="false" small />

              <div v-if="job.result" class="result-summary">
                <span>成功 {{ job.result.successCount }}</span>
                <span>失败 {{ job.result.failedCount }}</span>
              </div>

              <div v-if="job.errorMessage" class="history-error">
                {{ job.errorMessage }}
              </div>
            </article>
          </div>

          <div class="pagination">
            <button class="btn secondary" type="button" :disabled="page <= 1" @click="goToPage(page - 1)">
              上一页
            </button>
            <span class="page-indicator">第 {{ page }} / {{ totalPages }} 页</span>
            <button class="btn secondary" type="button" :disabled="page >= totalPages" @click="goToPage(page + 1)">
              下一页
            </button>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import AppHeader from '@/components/AppHeader.vue'
import ProgressBar from '@/components/ProgressBar.vue'
import { useJobsStore } from '@/store/modules/jobs'
import { jobProgressPercent, jobStatusLabel, type Job } from '@/types/jobs'
import { formatDate } from '@/utils/constants'

const jobsStore = useJobsStore()
const router = useRouter()

const page = ref(jobsStore.history.page || 1)
const pageSize = ref(jobsStore.history.pageSize || 10)

const activeAndResumable = computed(() => {
  const deduped = new Map<number, Job>()
  for (const job of jobsStore.resumableJobs) deduped.set(job.id, job)
  for (const job of jobsStore.activeJobs) {
    if (!deduped.has(job.id)) deduped.set(job.id, job)
  }
  return Array.from(deduped.values())
})

const resumableCount = computed(() => jobsStore.resumableJobs.length)
const totalPages = computed(() => Math.max(1, Math.ceil(jobsStore.history.total / pageSize.value)))
const showingRange = computed(() => {
  if (jobsStore.history.total === 0) return '0 条记录'
  const start = (page.value - 1) * pageSize.value + 1
  const end = Math.min(page.value * pageSize.value, jobsStore.history.total)
  return `${start}-${end} / ${jobsStore.history.total}`
})

const refreshAll = async () => {
  await Promise.all([
    jobsStore.loadActiveJobs(),
    jobsStore.loadResumableJobs(),
    jobsStore.loadJobHistory(page.value, pageSize.value),
  ])
}

const goToPage = async (nextPage: number) => {
  page.value = nextPage
  await jobsStore.loadJobHistory(page.value, pageSize.value)
}

const changePageSize = async () => {
  page.value = 1
  await jobsStore.loadJobHistory(page.value, pageSize.value)
}

watch(
  () => jobsStore.history.page,
  (value) => {
    if (value) page.value = value
  }
)

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

.section-head,
.job-card-head,
.job-card-foot,
.history-main,
.pagination,
.toolbar,
.section-actions {
  display: flex;
  justify-content: space-between;
  gap: 14px;
}

.section-head,
.pagination,
.job-card-foot,
.toolbar,
.section-actions {
  align-items: center;
}

.active-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 14px;
  margin-top: 18px;
}

.job-card,
.history-card,
.range-pill {
  border: 1px solid rgba(146, 165, 192, 0.18);
  background: rgba(255, 255, 255, 0.58);
}

.job-card,
.history-card {
  display: grid;
  gap: 14px;
  padding: 18px;
  border-radius: 24px;
}

.job-card-head strong,
.history-title-line strong {
  font-size: 15px;
  color: var(--text-color);
}

.job-card-head p,
.history-copy p,
.job-count,
.count,
.history-meta {
  margin-top: 6px;
  color: var(--text-soft);
  font-size: 12px;
  line-height: 1.7;
}

.job-actions,
.history-title-line,
.history-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 18px;
}

.history-side {
  min-width: 90px;
  text-align: right;
}

.percent {
  display: block;
  font-size: 22px;
  font-weight: 700;
  color: var(--text-color);
}

.toolbar {
  flex-wrap: wrap;
}

.range-pill {
  padding: 10px 12px;
  border-radius: 999px;
  font-size: 12px;
  color: var(--text-soft);
}

.size-picker {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-soft);
}

.size-picker select {
  width: auto;
  min-width: 84px;
  padding: 9px 12px;
  border-radius: 999px;
}

.result-summary {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: var(--text-faint);
}

.history-error {
  padding: 10px 12px;
  border-radius: 16px;
  background: rgba(204, 108, 108, 0.08);
  border: 1px solid rgba(204, 108, 108, 0.18);
  color: #a85d5d;
  font-size: 12px;
}

.page-indicator {
  font-size: 13px;
  color: var(--text-soft);
}

@media (max-width: 1040px) {
  .section-head,
  .pagination {
    grid-template-columns: 1fr;
    display: grid;
  }

  .section-actions {
    flex-wrap: wrap;
  }
}

@media (max-width: 760px) {
  .history-main,
  .job-card-head,
  .job-card-foot,
  .toolbar,
  .section-actions {
    flex-direction: column;
    align-items: flex-start;
  }

  .history-side {
    min-width: 0;
    text-align: left;
  }
}

@media (max-width: 640px) {
  .section-shell {
    padding: 16px;
  }

  .section-actions,
  .job-actions,
  .pagination {
    width: 100%;
    flex-direction: column;
  }
}
</style>

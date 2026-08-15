<template>
  <div class="page-shell tasks-view">
    <AppHeader />

    <div class="page-content">
      <div class="page-stack">
        <!-- 侧栏和面包屑已经写明当前是任务页，标题不再重复。 -->
        <header class="tk-header">
          <h1 class="visually-hidden">任务</h1>

          <div class="tk-tools">
            <button
              type="button"
              class="tk-filter-btn"
              :class="{ active: filterOpen }"
              :aria-expanded="filterOpen"
              @click.stop="filterOpen = !filterOpen"
            >
              <span class="tk-filter-dot" aria-hidden="true"></span>
              筛选<span v-if="jobsStore.jobType"> · 1</span>
            </button>

            <div class="tk-tools-meta">
              <span>共 <b>{{ jobsStore.total }}</b> 条记录</span>
              <span v-if="activeCount > 0" class="tk-live">
                <span class="tk-live-dot" aria-hidden="true"></span>
                <b>{{ activeCount }}</b> 个进行中
              </span>
              <button type="button" class="tk-refresh" :disabled="jobsStore.loading" @click="loadTaskCenter">
                {{ jobsStore.loading ? '正在刷新...' : '刷新状态' }}
              </button>
            </div>

            <div v-if="filterOpen" class="tk-popover" @click.stop>
              <div class="tk-popover-head">
                <strong>缩小范围</strong>
                <button type="button" class="tk-text-button" @click="setType('')">清除全部</button>
              </div>
              <span class="tk-popover-label">任务类型</span>
              <div class="tk-options">
                <button
                  v-for="option in TYPE_OPTIONS"
                  :key="option.value"
                  type="button"
                  class="tk-option"
                  :class="{ selected: jobsStore.jobType === option.value }"
                  @click="setType(option.value)"
                >
                  {{ option.label }}
                </button>
              </div>
            </div>
          </div>

          <div v-if="jobsStore.jobType" class="tk-chips">
            <span class="tk-chip">
              {{ currentTypeLabel }}
              <button type="button" :aria-label="`移除筛选 ${currentTypeLabel}`" @click="setType('')">×</button>
            </span>
          </div>
        </header>

        <div v-if="jobsStore.loading && jobsStore.jobs.length === 0" class="tk-empty">
          <div><strong>正在加载任务列表...</strong></div>
        </div>

        <!-- 空列表的出路取决于它为什么空：被筛没的应该清筛选，真没有记录的才去导入。 -->
        <div v-else-if="jobsStore.jobs.length === 0" class="tk-empty">
          <div>
            <strong>{{ jobsStore.jobType ? `没有${currentTypeLabel}任务` : '还没有任务记录' }}</strong>
            <p v-if="jobsStore.jobType">这个类型下还没有记录，换一个类型或清除筛选看看全部。</p>
            <p v-else>导入开始后，进行中的进度和历史结果都会集中显示在这里。</p>
            <button v-if="jobsStore.jobType" type="button" class="btn secondary" @click="setType('')">清除筛选</button>
            <router-link v-else class="btn primary" to="/import">前往导入</router-link>
          </div>
        </div>

        <template v-else>
          <div class="tk-list">
            <article v-for="row in viewJobs" :key="row.job.id" class="tk-item" :class="row.rail">
              <div class="tk-row">
                <span class="tk-rail" aria-hidden="true"></span>
                <span class="tk-type" :class="row.typeKey" aria-hidden="true">{{ row.typeChip }}</span>

                <div class="tk-main">
                  <strong>{{ row.job.title }}</strong>
                  <p :class="{ err: !!row.job.errorMessage }">{{ row.message }}</p>
                </div>

                <div class="tk-time">
                  <span>开始<i>{{ formatTime(row.job.startedAt) }}</i></span>
                  <span>结束<i>{{ formatTime(row.job.finishedAt) }}</i></span>
                </div>

                <!-- 进度条只在还在跑的时候出现；终态显示结果本身更有信息量。 -->
                <div class="tk-progress">
                  <template v-if="row.showBar">
                    <ProgressBar :percentage="jobProgressPercent(row.job)" :show-label="false" small />
                    <span class="tk-count">{{ row.job.progressCompleted }} / {{ row.job.progressTotal }}</span>
                  </template>
                  <p v-else class="tk-outcome">
                    {{ row.outcome.text }}
                    <span v-if="row.outcome.failed > 0"> · <em>失败 {{ row.outcome.failed }} 个</em></span>
                  </p>
                </div>

                <span class="tk-status" :class="row.statusTone">{{ jobStatusLabel(row.job.status) }}</span>

                <div class="tk-actions">
                  <button
                    v-if="row.canResume"
                    type="button"
                    class="tk-btn go"
                    @click="jobsStore.resumeJob(row.job.id)"
                  >
                    继续
                  </button>
                  <button
                    v-if="row.canCancel"
                    type="button"
                    class="tk-btn warn"
                    :disabled="cancelling.has(row.job.id)"
                    @click="confirmCancel(row.job)"
                  >
                    {{ cancelling.has(row.job.id) ? '正在停止...' : '取消' }}
                  </button>
                  <button v-if="row.hasDetails" type="button" class="tk-btn" @click="toggleExpanded(row.job.id)">
                    {{ expandedJobs.has(row.job.id) ? '收起' : '详情' }}
                  </button>
                  <button
                    v-if="row.canDelete"
                    type="button"
                    class="tk-btn warn"
                    @click="confirmDelete(row.job.id)"
                  >
                    删除
                  </button>
                </div>
              </div>

              <!-- 详情是行的延伸，不是浮层 -->
              <div v-if="expandedJobs.has(row.job.id)" class="tk-detail">
                <div class="tk-detail-grid">
                  <div v-if="row.job.errorMessage" class="tk-detail-block">
                    <strong>错误信息</strong>
                    <p class="err">{{ row.job.errorMessage }}</p>
                  </div>
                  <div v-if="row.job.result" class="tk-detail-block">
                    <strong>执行结果</strong>
                    <template v-if="cleanupDeleted(row.job) !== null">
                      <p>已清理 {{ cleanupDeleted(row.job) }} 条任务记录</p>
                    </template>
                    <template v-else-if="batchResult(row.job)">
                      <p>成功 {{ batchResult(row.job)?.successCount }} 个 · 失败 {{ batchResult(row.job)?.failedCount }} 个</p>
                      <ul v-if="batchResult(row.job)?.failedItems?.length" class="tk-failed">
                        <li v-for="(item, index) in batchResult(row.job)?.failedItems" :key="index">
                          {{ item.displayName || item.fileName || item.sourcePath || `文件 ${item.fileId}` }}：{{ item.error }}
                        </li>
                      </ul>
                    </template>
                  </div>
                  <div v-if="row.job.jobType === 'batch_import'" class="tk-detail-block tk-detail-files">
                    <div class="tk-detail-heading">
                      <strong>文件明细</strong>
                      <span v-if="batchImportItems[row.job.id]" class="tk-detail-summary">
                        {{ batchItemSummary(batchImportItems[row.job.id]) }}
                      </span>
                    </div>
                    <p v-if="detailLoading.has(row.job.id)" class="tk-detail-state">正在加载文件明细...</p>
                    <p v-else-if="detailErrors.has(row.job.id)" class="tk-detail-state err">
                      {{ detailErrors.get(row.job.id) }}
                    </p>
                    <div v-else-if="batchImportItems[row.job.id]?.length" class="tk-file-results">
                      <div
                        v-for="item in batchImportItems[row.job.id]"
                        :key="item.id"
                        class="tk-file-result"
                      >
                        <span class="tk-file-result-status" :class="batchItemStatusTone(item.status)">
                          {{ batchItemStatusLabel(item.status) }}
                        </span>
                        <div class="tk-file-result-copy">
                          <strong>{{ item.displayName || item.sourcePath }}</strong>
                          <p v-if="item.errorMessage" class="err">{{ item.errorMessage }}</p>
                        </div>
                      </div>
                    </div>
                    <p v-else class="tk-detail-state">暂无文件明细</p>
                  </div>
                  <div class="tk-detail-block">
                    <strong>任务 ID</strong>
                    <p class="meta">{{ row.job.id }}</p>
                  </div>
                </div>
              </div>
            </article>
          </div>

          <PaginationControls
            v-if="jobsStore.total > 0"
            :page="jobsStore.page"
            :page-size="jobsStore.pageSize"
            :total="jobsStore.total"
            :page-sizes="[10, 20, 50]"
            @change="jobsStore.loadJobs"
            @change-size="handleSizeChange"
          />
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { api } from '@/api'
import AppHeader from '@/components/AppHeader.vue'
import PaginationControls from '@/components/PaginationControls.vue'
import ProgressBar from '@/components/ProgressBar.vue'
import { useJobsStore } from '@/store/modules/jobs'
import {
  jobProgressPercent,
  jobStatusLabel,
  type BatchImportItem,
  type BatchImportResult,
  type CleanupJobResult,
  type Job,
} from '@/types/jobs'
import { formatDate } from '@/utils/constants'

const jobsStore = useJobsStore()
const expandedJobs = ref<Set<number>>(new Set())
const cancelling = ref<Set<number>>(new Set())
const detailLoading = ref<Set<number>>(new Set())
const detailErrors = ref<Map<number, string>>(new Map())
const batchImportItems = ref<Record<number, BatchImportItem[]>>({})
const filterOpen = ref(false)

const TERMINAL_STATUSES = ['completed', 'failed', 'cancelled', 'timed_out', 'cleanup_failed']
const RESUMABLE_STATUSES = ['awaiting_resume', 'timed_out']
const CANCELLABLE_STATUSES = ['pending', 'running', 'recovering']
// 需要注意的状态才上色：完成是常态，不着色，否则历史列表会变成一片绿。
const ATTENTION_STATUSES = ['awaiting_resume', 'timed_out']
const ERROR_STATUSES = ['failed', 'cleanup_failed']

const TYPE_OPTIONS = [
  { value: '', label: '全部' },
  { value: 'batch_import', label: '批量导入' },
  { value: 'single_import', label: '单文件导入' },
  { value: 'job_cleanup', label: '清理' },
]

const TYPE_META: Record<string, { chip: string; key: string; label: string }> = {
  batch_import: { chip: '批', key: 'batch', label: '批量导入' },
  single_import: { chip: '单', key: 'single', label: '单文件导入' },
  job_cleanup: { chip: '清', key: 'cleanup', label: '清理' },
}

const currentTypeLabel = computed(
  () => TYPE_OPTIONS.find(option => option.value === jobsStore.jobType)?.label || '全部'
)

// GetActiveJobs 返回的是完整的活动任务集合（不分页），所以按类型过滤后仍然准确。
// 不这样做的话，筛到某个类型时会出现「共 1 条记录 · 2 个进行中」这种自相矛盾的话。
const activeCount = computed(() =>
  jobsStore.jobType
    ? jobsStore.activeJobs.filter(job => job.jobType === jobsStore.jobType).length
    : jobsStore.activeJobs.length
)

// 排队中的任务没有 startedAt，formatDate 会返回字符串 "Invalid Date"——
// 它是真值，所以 `|| '—'` 兜不住，必须先判空再判非法。
const formatTime = (value: string) => {
  if (!value) return '—'
  return Number.isNaN(new Date(value).getTime()) ? '—' : formatDate(value)
}

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

const batchItemStatusLabel = (status: string) => {
  const labels: Record<string, string> = {
    pending: '未开始',
    processing: '导入中',
    completed: '已导入',
    failed: '未导入',
    recovering: '待恢复',
  }
  return labels[status] || status
}

const batchItemStatusTone = (status: string) => {
  if (status === 'completed') return 'is-completed'
  if (status === 'failed') return 'is-failed'
  if (status === 'processing' || status === 'recovering') return 'is-processing'
  return 'is-pending'
}

const batchItemSummary = (items: BatchImportItem[]) => {
  const imported = items.filter(item => item.status === 'completed').length
  const notImported = items.filter(item => item.status === 'failed').length
  const pending = items.length - imported - notImported
  const parts = [`已导入 ${imported} 个`, `未导入 ${notImported} 个`]
  if (pending > 0) parts.push(`待处理 ${pending} 个`)
  return parts.join(' · ')
}

const loadBatchImportItems = async (jobId: number) => {
  if (batchImportItems.value[jobId] || detailLoading.value.has(jobId)) return

  detailLoading.value.add(jobId)
  detailErrors.value.delete(jobId)
  try {
    batchImportItems.value[jobId] = await api.jobs.getBatchImportItems(jobId)
  } catch (error: any) {
    detailErrors.value.set(jobId, error?.message || '文件明细加载失败')
  } finally {
    detailLoading.value.delete(jobId)
  }
}

const outcomeOf = (job: Job) => {
  const deleted = cleanupDeleted(job)
  if (deleted !== null) {
    return { text: `已清理 ${deleted} 条记录`, failed: 0 }
  }
  const result = batchResult(job)
  if (result) {
    return { text: `成功 ${result.successCount} 个`, failed: result.failedCount }
  }
  return { text: `${job.progressCompleted} / ${job.progressTotal}`, failed: 0 }
}

const viewJobs = computed(() =>
  jobsStore.jobs.map(job => {
    const meta = TYPE_META[job.jobType] || { chip: '任', key: 'other', label: job.jobType }
    const isAttention = ATTENTION_STATUSES.includes(job.status)
    const isError = ERROR_STATUSES.includes(job.status)
    const isActive = CANCELLABLE_STATUSES.includes(job.status)

    return {
      job,
      typeChip: meta.chip,
      typeKey: meta.key,
      message: job.errorMessage || job.progressMessage || jobStatusLabel(job.status),
      rail: isActive ? 'is-active' : isAttention ? 'is-attention' : isError ? 'is-error' : '',
      statusTone: isActive ? 'run' : isAttention ? 'wait' : isError ? 'err' : '',
      showBar: isActive || isAttention,
      outcome: outcomeOf(job),
      canResume: RESUMABLE_STATUSES.includes(job.status),
      canCancel: isActive,
      canDelete: TERMINAL_STATUSES.includes(job.status),
      // 批量导入即使没有 result，也要允许查看逐文件状态，尤其是取消任务。
      hasDetails: job.jobType === 'batch_import' || Boolean(job.errorMessage || job.result),
    }
  })
)

const setType = async (value: string) => {
  filterOpen.value = false
  if (jobsStore.jobType === value) return
  await jobsStore.loadJobs(1, jobsStore.pageSize, value)
}

const toggleExpanded = (jobId: number) => {
  if (expandedJobs.value.has(jobId)) {
    expandedJobs.value.delete(jobId)
  } else {
    expandedJobs.value.add(jobId)
    const job = jobsStore.jobs.find(item => item.id === jobId)
    if (job?.jobType === 'batch_import') {
      void loadBatchImportItems(jobId)
    }
  }
}

const confirmCancel = async (job: Job) => {
  // 批量导入取消后会清理暂存文件，说清楚代价再让用户确认。
  const message =
    job.jobType === 'batch_import'
      ? '取消后会停止导入，并清理尚未归档的暂存文件。已经归档的文件会保留。确定取消？'
      : '确定取消这个任务？'
  if (!window.confirm(message)) return

  cancelling.value.add(job.id)
  try {
    await jobsStore.cancelJob(job.id)
  } finally {
    cancelling.value.delete(job.id)
  }
}

const confirmDelete = async (jobId: number) => {
  if (!window.confirm('确定删除该任务记录？')) return
  await jobsStore.deleteJobRecord(jobId)
}

const handleSizeChange = async (size: number) => {
  await jobsStore.loadJobs(1, size)
}

const loadTaskCenter = async () => {
  // 页面进入时同步活动任务和历史列表，顶部导航已经提供导入入口，这里保持列表本身轻量。
  await Promise.all([jobsStore.loadActiveJobs(), jobsStore.loadJobs(jobsStore.page, jobsStore.pageSize)])
}

const closeFilter = () => {
  filterOpen.value = false
}

onMounted(() => {
  void loadTaskCenter()
  document.addEventListener('click', closeFilter)
})

onUnmounted(() => {
  document.removeEventListener('click', closeFilter)
})
</script>

<style scoped>
.tasks-view {
  --tk-surface: rgba(255, 255, 255, 0.86);
  --tk-surface-muted: rgba(243, 246, 250, 0.82);
  --tk-control: rgba(248, 250, 253, 0.92);
  --tk-border: rgba(148, 163, 184, 0.28);
  --tk-border-strong: rgba(148, 163, 184, 0.5);
  --tk-accent: #2d8cf0;
  --tk-accent-strong: #1565c0;
  --tk-accent-soft: rgba(45, 140, 240, 0.1);
  --tk-text: #223042;
  --tk-text-soft: #5f6f82;
  --tk-text-faint: #7f8ea3;
  --tk-wait: #a06c14;
  --tk-wait-bg: rgba(226, 173, 73, 0.16);
  --tk-wait-rail: #e2ad49;
  --tk-err: #a1414e;
  --tk-err-bg: rgba(201, 115, 127, 0.14);
  --tk-err-rail: #c9737f;
  --tk-detail-bg: rgba(248, 250, 252, 0.9);
}

[data-theme='dark'] .tasks-view {
  --tk-surface: rgba(35, 42, 52, 0.9);
  --tk-surface-muted: rgba(42, 50, 61, 0.7);
  --tk-control: rgba(30, 37, 47, 0.7);
  --tk-border: rgba(173, 191, 214, 0.16);
  --tk-border-strong: rgba(173, 191, 214, 0.3);
  --tk-accent: #4da3ff;
  --tk-accent-strong: #8cc2ff;
  --tk-accent-soft: rgba(77, 163, 255, 0.14);
  --tk-text: #f2f6fb;
  --tk-text-soft: #c2cfdd;
  --tk-text-faint: #92a0b1;
  --tk-wait: #e2bd77;
  --tk-wait-bg: rgba(190, 140, 40, 0.2);
  --tk-wait-rail: #d9a441;
  --tk-err: #eda3ad;
  --tk-err-bg: rgba(210, 90, 110, 0.2);
  --tk-err-rail: #d2637a;
  --tk-detail-bg: rgba(18, 24, 33, 0.55);
}

.tasks-view .page-content {
  padding: 20px 44px 50px;
}

.tasks-view .page-stack {
  gap: 16px;
}

.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  margin: -1px;
  padding: 0;
  overflow: hidden;
  clip: rect(0 0 0 0);
  white-space: nowrap;
  border: 0;
}

/* 筛选沿用资料库的工具带 + 浮层 + chip，只是这里没有搜索：
   ListJobs 只支持 jobType，不支持按标题检索，也不支持按状态过滤。 */
.tk-tools {
  position: relative;
  display: flex;
  align-items: center;
  gap: 12px;
  padding-bottom: 14px;
  border-bottom: 1px solid var(--tk-border);
}

.tk-filter-btn {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  min-height: 36px;
  padding: 0 13px;
  border: 1px solid var(--tk-border);
  border-radius: 10px;
  background: var(--tk-control);
  color: var(--tk-text-soft);
  font-size: 13px;
  font-weight: 600;
}

.tk-filter-btn:hover,
.tk-filter-btn.active {
  border-color: color-mix(in srgb, var(--tk-accent) 40%, transparent);
  background: var(--tk-accent-soft);
  color: var(--tk-accent-strong);
}

.tk-filter-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--tk-accent);
}

.tk-tools-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-left: auto;
  color: var(--tk-text-faint);
  font-size: 12px;
}

.tk-tools-meta b {
  color: var(--tk-text-soft);
  font-weight: 700;
}

.tk-live {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--tk-accent-strong);
}

.tk-live b {
  color: inherit;
}

.tk-live-dot {
  position: relative;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--tk-accent);
}

.tk-live-dot::after {
  content: '';
  position: absolute;
  inset: -4px;
  border: 1px solid var(--tk-accent);
  border-radius: 50%;
  opacity: 0.35;
  animation: tk-pulse 1.8s ease-out infinite;
}

@keyframes tk-pulse {
  0% {
    transform: scale(0.6);
    opacity: 0.5;
  }
  100% {
    transform: scale(1.25);
    opacity: 0;
  }
}

.tk-refresh {
  background: transparent;
  color: var(--tk-accent-strong);
  font-size: 12px;
  font-weight: 700;
}

.tk-refresh:disabled {
  color: var(--tk-text-faint);
  cursor: not-allowed;
}

.tk-popover {
  position: absolute;
  z-index: 5;
  top: 46px;
  left: 0;
  width: max-content;
  max-width: min(360px, calc(100vw - 48px));
  padding: 14px;
  border: 1px solid var(--tk-border-strong);
  border-radius: 14px;
  background: var(--tk-surface);
  box-shadow: 0 18px 38px var(--shadow-strong);
  backdrop-filter: blur(12px);
}

.tk-popover-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--tk-border);
}

.tk-popover-head strong {
  color: var(--tk-text);
  font-size: 13px;
}

.tk-text-button {
  background: transparent;
  color: var(--tk-accent-strong);
  font-size: 12px;
  font-weight: 600;
}

.tk-popover-label {
  display: block;
  margin: 13px 0 9px;
  color: var(--tk-text-faint);
  font-size: 12px;
  font-weight: 700;
}

.tk-options {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
}

.tk-option {
  padding: 7px 10px;
  border: 1px solid var(--tk-border);
  border-radius: 8px;
  background: var(--tk-control);
  color: var(--tk-text-soft);
  font-size: 12px;
}

.tk-option:hover {
  border-color: var(--tk-border-strong);
}

.tk-option.selected {
  border-color: color-mix(in srgb, var(--tk-accent) 40%, transparent);
  background: var(--tk-accent-soft);
  color: var(--tk-accent-strong);
  font-weight: 600;
}

.tk-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
  margin-top: 13px;
}

.tk-chip {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 6px 10px;
  border: 1px solid color-mix(in srgb, var(--tk-accent) 32%, transparent);
  border-radius: 8px;
  background: var(--tk-accent-soft);
  color: var(--tk-accent-strong);
  font-size: 12px;
  font-weight: 600;
}

.tk-chip button {
  background: transparent;
  color: inherit;
  font-size: 14px;
  line-height: 1;
}

.tk-list {
  display: grid;
  gap: 8px;
}

.tk-item {
  overflow: hidden;
  border: 1px solid var(--tk-border);
  border-radius: var(--radius-md);
  background: var(--tk-surface);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.tk-item:hover {
  border-color: var(--tk-border-strong);
  box-shadow: 0 8px 22px var(--shadow-color);
}

/* 每一行都是独立 grid，所以操作列必须定宽，否则按钮数量不同的行列位会错开，纵向就扫不成列了。 */
.tk-row {
  display: grid;
  grid-template-columns: 3px 34px minmax(0, 1fr) 132px 132px 74px 152px;
  align-items: center;
  gap: 13px;
  padding: 12px 14px 12px 0;
}

/* 行首色条：只有需要注意的状态才上色。 */
.tk-rail {
  align-self: stretch;
  border-radius: 0 3px 3px 0;
  background: transparent;
}

.tk-item.is-active .tk-rail {
  background: var(--tk-accent);
}

.tk-item.is-attention .tk-rail {
  background: var(--tk-wait-rail);
}

.tk-item.is-error .tk-rail {
  background: var(--tk-err-rail);
}

.tk-type {
  display: grid;
  place-items: center;
  width: 34px;
  height: 34px;
  border-radius: 9px;
  background: var(--tk-surface-muted);
  color: var(--tk-text-faint);
  font-size: 13px;
  font-weight: 700;
}

.tk-type.batch {
  color: var(--tk-accent-strong);
  background: var(--tk-accent-soft);
}

.tk-type.single {
  color: #46618a;
  background: rgba(120, 150, 190, 0.16);
}

.tk-type.cleanup {
  color: var(--tk-wait);
  background: var(--tk-wait-bg);
}

[data-theme='dark'] .tk-type.single {
  color: #a8bcd4;
}

.tk-main {
  min-width: 0;
}

.tk-main strong {
  display: block;
  overflow: hidden;
  color: var(--tk-text);
  font-size: 13px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tk-main p {
  margin-top: 4px;
  overflow: hidden;
  color: var(--tk-text-faint);
  font-size: 12px;
  line-height: 1.5;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tk-main p.err {
  color: var(--tk-err);
}

.tk-time {
  display: grid;
  gap: 3px;
  color: var(--tk-text-faint);
  font-size: 11px;
}

.tk-time span {
  display: flex;
  gap: 7px;
}

.tk-time i {
  font-style: normal;
  opacity: 0.85;
}

.tk-progress {
  display: grid;
  gap: 5px;
}

.tk-count {
  color: var(--tk-text-faint);
  font-size: 11px;
}

.tk-outcome {
  color: var(--tk-text-soft);
  font-size: 12px;
  line-height: 1.5;
}

.tk-outcome em {
  color: var(--tk-err);
  font-style: normal;
}

.tk-status {
  justify-self: start;
  padding: 4px 9px;
  border-radius: 7px;
  background: var(--tk-surface-muted);
  color: var(--tk-text-faint);
  font-size: 11px;
  font-weight: 700;
  white-space: nowrap;
}

.tk-status.run {
  color: var(--tk-accent-strong);
  background: var(--tk-accent-soft);
}

.tk-status.wait {
  color: var(--tk-wait);
  background: var(--tk-wait-bg);
}

.tk-status.err {
  color: var(--tk-err);
  background: var(--tk-err-bg);
}

.tk-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
}

.tk-btn {
  padding: 6px 10px;
  border: 1px solid var(--tk-border);
  border-radius: 8px;
  background: var(--tk-control);
  color: var(--tk-text-soft);
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
}

.tk-btn:hover:not(:disabled) {
  border-color: var(--tk-border-strong);
  color: var(--tk-text);
}

.tk-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.tk-btn.go {
  border-color: var(--tk-accent);
  background: var(--tk-accent);
  color: #fff;
}

.tk-btn.go:hover:not(:disabled) {
  background: var(--primary-hover);
  border-color: var(--primary-hover);
  color: #fff;
}

.tk-btn.warn:hover:not(:disabled) {
  border-color: color-mix(in srgb, var(--tk-err) 45%, transparent);
  background: var(--tk-err-bg);
  color: var(--tk-err);
}

/* 详情是行的延伸，不是浮层，所以贴着行的底边内嵌。 */
.tk-detail {
  padding: 14px 16px 15px 20px;
  border-top: 1px solid var(--tk-border);
  background: var(--tk-detail-bg);
}

.tk-detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.tk-detail-block strong {
  display: block;
  margin-bottom: 6px;
  color: var(--tk-text-faint);
  font-size: 11px;
  font-weight: 700;
}

.tk-detail-block p {
  color: var(--tk-text-soft);
  font-size: 12px;
  line-height: 1.65;
}

.tk-detail-block p.err {
  color: var(--tk-err);
}

.tk-detail-block p.meta {
  color: var(--tk-text-faint);
}

.tk-detail-files {
  grid-column: 1 / -1;
}

.tk-detail-heading {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
}

.tk-detail-summary {
  color: var(--tk-text-faint);
  font-size: 11px;
  font-weight: 500;
}

.tk-detail-state {
  margin-top: 4px;
  color: var(--tk-text-faint);
  font-size: 12px;
}

.tk-detail-state.err {
  color: var(--tk-err);
}

.tk-file-results {
  display: grid;
  gap: 6px;
  max-height: 320px;
  margin-top: 8px;
  overflow-y: auto;
  padding-right: 3px;
}

.tk-file-result {
  display: grid;
  grid-template-columns: 56px minmax(0, 1fr);
  align-items: start;
  gap: 10px;
  padding: 8px 10px;
  border: 1px solid var(--tk-border);
  border-radius: 8px;
  background: var(--tk-control);
}

.tk-file-result-status {
  justify-self: start;
  padding: 3px 5px;
  border-radius: 5px;
  font-size: 11px;
  line-height: 1.2;
  white-space: nowrap;
}

.tk-file-result-status.is-completed {
  background: color-mix(in srgb, #3c9367 14%, transparent);
  color: #347a57;
}

.tk-file-result-status.is-failed {
  background: var(--tk-err-bg);
  color: var(--tk-err);
}

.tk-file-result-status.is-processing,
.tk-file-result-status.is-pending {
  background: var(--tk-wait-bg);
  color: var(--tk-wait);
}

.tk-file-result-copy {
  min-width: 0;
}

.tk-file-result-copy strong {
  display: block;
  overflow: hidden;
  color: var(--tk-text-soft);
  font-size: 12px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tk-file-result-copy p {
  margin-top: 3px;
  overflow-wrap: anywhere;
}

.tk-failed {
  margin: 7px 0 0;
  padding-left: 16px;
  color: var(--tk-text-faint);
  font-size: 12px;
  line-height: 1.75;
}

.tk-empty {
  display: grid;
  place-items: center;
  min-height: 260px;
  padding: 32px;
  border: 1px solid var(--tk-border);
  border-radius: var(--radius-md);
  background: var(--tk-surface-muted);
  text-align: center;
}

.tk-empty strong {
  display: block;
  color: var(--tk-text);
  font-size: 14px;
}

.tk-empty p {
  max-width: 340px;
  margin: 8px auto 16px;
  color: var(--tk-text-faint);
  font-size: 12px;
  line-height: 1.7;
}

@media (max-width: 1080px) {
  /* 先收掉绝对时间：进度和状态比“几点开始”更值得占位。 */
  .tk-row {
    grid-template-columns: 3px 34px minmax(0, 1fr) 132px 74px 152px;
  }

  .tk-time {
    display: none;
  }
}

@media (max-width: 760px) {
  .tasks-view .page-content {
    padding: 16px 12px 40px;
  }

  .tk-tools {
    flex-wrap: wrap;
  }

  .tk-tools-meta {
    width: 100%;
    margin-left: 0;
  }

  /* 窄屏改成三层：标题、进度、状态与操作同处一行，不让每段各占一整行。 */
  .tk-row {
    grid-template-columns: 3px 34px minmax(0, 1fr) auto;
    grid-template-areas:
      'rail type main main'
      'rail prog prog prog'
      'rail stat stat acts';
    row-gap: 10px;
    padding-bottom: 13px;
  }

  .tk-rail {
    grid-area: rail;
  }

  .tk-type {
    grid-area: type;
    align-self: start;
  }

  .tk-main {
    grid-area: main;
  }

  .tk-progress {
    grid-area: prog;
  }

  .tk-status {
    grid-area: stat;
    align-self: center;
  }

  .tk-actions {
    grid-area: acts;
    flex-wrap: wrap;
  }
}

@media (prefers-reduced-motion: reduce) {
  .tk-live-dot::after {
    animation: none;
  }
}
</style>

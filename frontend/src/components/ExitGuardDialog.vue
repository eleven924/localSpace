<template>
  <Teleport to="body">
    <Transition name="exit-guard">
      <div
        v-if="visible && snapshot"
        class="exit-guard-overlay"
        role="dialog"
        aria-modal="true"
        aria-labelledby="exit-guard-title"
        @click="handleBackdropClick"
      >
        <section class="exit-guard-panel" @click.stop>
          <header class="exit-guard-header">
            <div class="exit-guard-icon" aria-hidden="true">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M12 3v10" />
                <path d="M7.5 6.5a7 7 0 1 0 9 0" />
              </svg>
            </div>
            <div class="exit-guard-copy">
              <p class="exit-guard-eyebrow">退出前检查</p>
              <h2 id="exit-guard-title">当前有任务正在处理</h2>
              <p class="exit-guard-summary">
                退出会中断这些任务。继续前，先确认是否要保留当前进度。
              </p>
            </div>
          </header>

          <div class="exit-guard-stats">
            <div class="stat-card">
              <strong>{{ snapshot.total }}</strong>
              <span>受保护任务</span>
            </div>
            <div class="stat-card">
              <strong>{{ runningCount }}</strong>
              <span>运行中</span>
            </div>
            <div class="stat-card">
              <strong>{{ recoveringCount }}</strong>
              <span>恢复中</span>
            </div>
            <div class="stat-card">
              <strong>{{ pendingCount }}</strong>
              <span>排队中</span>
            </div>
          </div>

          <div class="exit-guard-body">
            <div class="exit-guard-list-head">
              <h3>受影响任务</h3>
              <span v-if="hiddenCount > 0" class="exit-guard-hint">还有 {{ hiddenCount }} 条未展示</span>
            </div>

            <div v-if="displayJobs.length === 0" class="empty-line">
              没有可展示的任务摘要。
            </div>

            <div v-else class="exit-guard-list">
              <article
                v-for="job in displayJobs"
                :key="job.id"
                class="exit-guard-row"
              >
                <div class="row-copy">
                  <strong>{{ job.title }}</strong>
                  <p>{{ job.progressMessage || jobStatusLabel(job.status) }}</p>
                </div>
                <div class="row-meta">
                  <span class="status-pill">{{ jobStatusLabel(job.status) }}</span>
                  <span class="resume-pill" :class="{ safe: job.canResume }">
                    {{ job.canResume ? '可恢复' : '不可恢复' }}
                  </span>
                  <span class="percent">{{ jobProgressPercent(job) }}%</span>
                </div>
              </article>
            </div>
          </div>

          <footer class="exit-guard-footer">
            <p v-if="errorMessage" class="exit-guard-error">{{ errorMessage }}</p>
            <div class="exit-guard-actions">
              <button type="button" class="btn secondary" :disabled="loading" @click="closeDialog">
                返回应用
              </button>
              <button type="button" class="btn secondary ghost-action" :disabled="loading" @click="goToTasks">
                查看任务
              </button>
              <button type="button" class="btn danger" :disabled="loading" @click="confirmExit">
                <LoadingSpinner v-if="loading" :size="'small'" :inline="true" />
                <span v-else>仍然退出</span>
              </button>
            </div>
          </footer>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, watch } from 'vue'
import { useRouter } from 'vue-router'
import LoadingSpinner from './LoadingSpinner.vue'
import { useExitGuardStore } from '@/store/modules/exitGuard'
import { jobProgressPercent, jobStatusLabel } from '@/types/jobs'

const router = useRouter()
const exitGuardStore = useExitGuardStore()

const visible = computed(() => exitGuardStore.visible)
const snapshot = computed(() => exitGuardStore.snapshot)
const loading = computed(() => exitGuardStore.loading)
const errorMessage = computed(() => exitGuardStore.errorMessage)

const statusCounts = computed(() => snapshot.value?.statusCounts ?? {})
const runningCount = computed(() => statusCounts.value.running ?? 0)
const recoveringCount = computed(() => statusCounts.value.recovering ?? 0)
const pendingCount = computed(() => statusCounts.value.pending ?? 0)

const displayJobs = computed(() => {
  const jobs = snapshot.value?.jobs ?? []
  const priority: Record<string, number> = {
    running: 0,
    recovering: 1,
    pending: 2,
  }

  return [...jobs]
    .sort((left, right) => (priority[left.status] ?? 99) - (priority[right.status] ?? 99))
    .slice(0, 5)
})

const hiddenCount = computed(() => Math.max((snapshot.value?.total ?? 0) - displayJobs.value.length, 0))

const closeDialog = () => {
  exitGuardStore.close()
}

const goToTasks = () => {
  exitGuardStore.close()
  void router.push('/tasks')
}

const confirmExit = async () => {
  try {
    await exitGuardStore.confirmQuit()
  } catch {
    // 错误由 store 持有，这里只保留弹层以便用户继续处理。
  }
}

const handleBackdropClick = () => {
  closeDialog()
}

const handleEscape = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && visible.value) {
    closeDialog()
  }
}

watch(visible, (isVisible) => {
  if (isVisible) {
    document.addEventListener('keydown', handleEscape)
  } else {
    document.removeEventListener('keydown', handleEscape)
  }
}, { immediate: true })

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleEscape)
})
</script>

<style scoped>
.exit-guard-overlay {
  position: fixed;
  inset: 0;
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background: rgba(10, 15, 23, 0.58);
  backdrop-filter: blur(14px);
}

.exit-guard-panel {
  width: min(860px, 100%);
  max-height: min(88vh, 860px);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  border-radius: 16px;
  border: 1px solid rgba(251, 191, 36, 0.24);
  background:
    linear-gradient(180deg, rgba(255, 248, 235, 0.98), rgba(255, 255, 255, 0.96));
  color: #1f2937;
  box-shadow: 0 28px 80px rgba(15, 23, 42, 0.34);
}

[data-theme='dark'] .exit-guard-panel {
  border-color: rgba(251, 191, 36, 0.22);
  background:
    linear-gradient(180deg, rgba(28, 32, 40, 0.98), rgba(22, 27, 34, 0.96));
  color: var(--text-color);
}

.exit-guard-header {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 16px;
  padding: 22px 24px 18px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.18);
  border-top: 4px solid #d97706;
}

.exit-guard-icon {
  width: 56px;
  height: 56px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 14px;
  background: linear-gradient(135deg, rgba(217, 119, 6, 0.18), rgba(239, 68, 68, 0.14));
  color: #c2410c;
}

[data-theme='dark'] .exit-guard-icon {
  color: #f59e0b;
}

.exit-guard-icon svg {
  width: 26px;
  height: 26px;
}

.exit-guard-copy {
  min-width: 0;
}

.exit-guard-eyebrow {
  display: inline-flex;
  padding: 3px 8px;
  border-radius: 999px;
  background: rgba(251, 191, 36, 0.14);
  color: #92400e;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.02em;
}

[data-theme='dark'] .exit-guard-eyebrow {
  color: #fbbf24;
}

.exit-guard-copy h2 {
  margin-top: 8px;
  font-size: 22px;
  line-height: 1.3;
  font-weight: 700;
}

.exit-guard-summary {
  margin-top: 6px;
  font-size: 13px;
  line-height: 1.7;
  color: color-mix(in srgb, var(--text-color) 72%, transparent);
}

.exit-guard-stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
  padding: 16px 24px 0;
}

.stat-card {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px 14px;
  border-radius: 12px;
  background: color-mix(in srgb, var(--surface-muted) 88%, transparent);
  border: 1px solid rgba(148, 163, 184, 0.16);
}

.stat-card strong {
  font-size: 22px;
  line-height: 1;
  font-weight: 700;
}

.stat-card span {
  font-size: 11px;
  color: var(--text-soft);
}

.exit-guard-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 0;
  padding: 18px 24px 20px;
}

.exit-guard-list-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.exit-guard-list-head h3 {
  font-size: 15px;
  font-weight: 700;
}

.exit-guard-hint {
  font-size: 12px;
  color: var(--text-soft);
}

.exit-guard-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-height: 360px;
  overflow: auto;
  padding-right: 4px;
}

.exit-guard-row {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 14px;
  border-radius: 12px;
  background: color-mix(in srgb, var(--surface-muted) 96%, transparent);
  border: 1px solid rgba(148, 163, 184, 0.16);
}

.row-copy {
  min-width: 0;
}

.row-copy strong {
  display: block;
  font-size: 14px;
  line-height: 1.45;
  font-weight: 600;
  color: var(--text-color);
}

.row-copy p {
  margin-top: 5px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--text-soft);
}

.row-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: flex-end;
  align-content: flex-start;
}

.resume-pill {
  display: inline-flex;
  align-items: center;
  padding: 4px 8px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 600;
  color: #9a3412;
  background: rgba(251, 191, 36, 0.16);
}

.resume-pill.safe {
  color: #166534;
  background: rgba(74, 222, 128, 0.12);
}

[data-theme='dark'] .resume-pill {
  color: #fdba74;
}

[data-theme='dark'] .resume-pill.safe {
  color: #86efac;
}

.percent {
  min-width: 44px;
  text-align: right;
  font-size: 12px;
  font-weight: 700;
  color: var(--text-faint);
}

.empty-line {
  padding: 14px;
  border-radius: 12px;
  border: 1px dashed rgba(148, 163, 184, 0.24);
  color: var(--text-soft);
  font-size: 13px;
}

.exit-guard-footer {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 0 24px 24px;
}

.exit-guard-error {
  color: var(--error-color);
  font-size: 13px;
  line-height: 1.6;
}

.exit-guard-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.ghost-action {
  border-color: rgba(148, 163, 184, 0.22);
}

.exit-guard-enter-active,
.exit-guard-leave-active {
  transition: opacity 0.18s ease;
}

.exit-guard-enter-active .exit-guard-panel,
.exit-guard-leave-active .exit-guard-panel {
  transition: transform 0.18s ease, opacity 0.18s ease;
}

.exit-guard-enter-from,
.exit-guard-leave-to {
  opacity: 0;
}

.exit-guard-enter-from .exit-guard-panel,
.exit-guard-leave-to .exit-guard-panel {
  transform: translateY(10px) scale(0.98);
  opacity: 0.98;
}

@media (max-width: 760px) {
  .exit-guard-overlay {
    padding: 12px;
  }

  .exit-guard-panel {
    max-height: calc(100vh - 24px);
  }

  .exit-guard-header,
  .exit-guard-body,
  .exit-guard-footer {
    padding-left: 16px;
    padding-right: 16px;
  }

  .exit-guard-header {
    grid-template-columns: 1fr;
  }

  .exit-guard-stats {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    padding-left: 16px;
    padding-right: 16px;
  }

  .exit-guard-actions {
    flex-direction: column-reverse;
  }

  .exit-guard-actions .btn {
    width: 100%;
  }

  .exit-guard-row {
    flex-direction: column;
  }

  .row-meta {
    justify-content: flex-start;
  }
}

@media (prefers-reduced-motion: reduce) {
  .exit-guard-enter-active,
  .exit-guard-leave-active,
  .exit-guard-enter-active .exit-guard-panel,
  .exit-guard-leave-active .exit-guard-panel {
    transition: none;
  }
}
</style>

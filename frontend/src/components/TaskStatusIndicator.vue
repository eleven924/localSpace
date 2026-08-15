<template>
  <div
    class="task-indicator"
    @mouseenter="open = true"
    @mouseleave="open = false"
    @focusin="open = true"
    @focusout="handleFocusOut"
  >
    <button
      class="task-button"
      :class="{ running: hasRunningJobs }"
      type="button"
      :title="indicatorLabel"
      :aria-label="indicatorLabel"
      @click="togglePanel"
    >
      <svg
        class="task-icon"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="1.8"
        stroke-linecap="round"
        aria-hidden="true"
      >
        <g class="task-icon-orbit">
          <path d="M12.8 3A9 9 0 1 0 21 11.2" />
        </g>
        <rect x="8.6" y="8.6" width="6.8" height="6.8" rx="2.1" fill="currentColor" stroke="none" />
      </svg>
      <span v-if="hasRunningJobs" class="running-badge">{{ runningCount > 99 ? '99+' : runningCount }}</span>
    </button>

    <transition name="popover-fade">
      <div v-if="open" class="task-popover">
        <div class="popover-head">
          <h4>{{ hasRunningJobs ? '运行中的任务' : '最近任务' }}</h4>
          <button class="link-button" type="button" @click="goToTasks">任务中心</button>
        </div>

        <div v-if="summaryJobs.length === 0" class="empty-state">
          还没有任务记录，创建批量导入后会在这里显示。
        </div>

        <div v-else class="task-list">
          <button
            v-for="job in summaryJobs"
            :key="job.id"
            class="task-row"
            type="button"
            @click="goToTasks"
          >
            <div class="task-copy">
              <strong>{{ job.title }}</strong>
              <p>{{ job.progressMessage || jobStatusLabel(job.status) }}</p>
            </div>
            <div class="task-meta">
              <span class="status-pill">{{ jobStatusLabel(job.status) }}</span>
              <span class="percent">{{ jobProgressPercent(job) }}%</span>
            </div>
          </button>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useJobsStore } from '@/store/modules/jobs'
import { jobProgressPercent, jobStatusLabel } from '@/types/jobs'

const jobsStore = useJobsStore()
const router = useRouter()
const open = ref(false)

const summaryJobs = computed(() => jobsStore.summaryJobs)
const runningCount = computed(() => jobsStore.totalRunningCount)
const hasRunningJobs = computed(() => runningCount.value > 0)
const indicatorLabel = computed(() =>
  hasRunningJobs.value ? `${runningCount.value} 项任务进行中` : '任务中心'
)

const togglePanel = () => {
  open.value = !open.value
}

const goToTasks = () => {
  open.value = false
  router.push('/tasks')
}

const handleFocusOut = (event: FocusEvent) => {
  const nextTarget = event.relatedTarget
  if (nextTarget instanceof Node && event.currentTarget instanceof Node && event.currentTarget.contains(nextTarget)) {
    return
  }

  open.value = false
}
</script>

<style scoped>
.task-indicator {
  position: relative;
}

.task-button {
  position: relative;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 1px solid transparent;
  background: transparent;
  color: var(--text-soft);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s ease, color 0.15s ease;
}

.task-button:hover {
  background: var(--hover-bg, rgba(148, 163, 184, 0.24));
  color: var(--text-color);
}

.task-button.running {
  color: var(--primary-color, #2196f3);
}

.task-icon {
  width: 18px;
  height: 18px;
  pointer-events: none;
}

/* 只让外圈弧线转动，中心方块保持静止：旋转挂在 <g> 上而不是整个 <svg>。 */
.task-icon-orbit {
  transform-box: view-box;
  transform-origin: 12px 12px;
  opacity: 0.55;
  transition: opacity 0.15s ease;
}

.task-button.running .task-icon-orbit {
  opacity: 1;
  animation: task-orbit 1.8s linear infinite;
}

.running-badge {
  position: absolute;
  top: -1px;
  right: -1px;
  min-width: 13px;
  height: 13px;
  padding: 0 3px;
  border-radius: 7px;
  background: var(--primary-color, #2196f3);
  color: #fff;
  font-size: 9px;
  font-weight: 700;
  line-height: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 0 0 2px var(--bg-color, #f4f7fb);
  pointer-events: none;
}

.task-popover {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  width: 340px;
  padding: 12px;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  background-color: var(--surface-color);
  box-shadow: 0 12px 28px var(--shadow-strong);
  z-index: 50;
}

.popover-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.popover-head h4 {
  font-size: 15px;
  color: var(--text-color);
}

.link-button {
  padding: 0;
  font-size: 12px;
  font-weight: 600;
  color: var(--primary-color);
}

.empty-state {
  margin-top: 10px;
  padding: 12px;
  border-radius: 10px;
  background-color: var(--surface-muted);
  color: var(--text-soft);
  line-height: 1.6;
  font-size: 13px;
}

.task-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 10px;
}

.task-row {
  width: 100%;
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  background-color: var(--surface-muted);
  text-align: left;
}

.task-copy strong {
  display: block;
  font-size: 13px;
  color: var(--text-color);
}

.task-copy p {
  margin-top: 4px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--text-soft);
}

.task-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 6px;
}

.percent {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-faint);
}

.popover-fade-enter-active,
.popover-fade-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.popover-fade-enter-from,
.popover-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

@keyframes task-orbit {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}

@media (prefers-reduced-motion: reduce) {
  .task-button.running .task-icon-orbit {
    animation: none;
  }
}

@media (max-width: 640px) {
  .task-popover {
    width: min(340px, calc(100vw - 24px));
  }
}
</style>

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
      :title="hasRunningJobs ? '有任务正在运行' : '任务中心'"
      :aria-label="hasRunningJobs ? '有任务正在运行' : '任务中心'"
      @click="togglePanel"
    >
      <svg
        class="task-icon"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2.5"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <line x1="12" y1="12" x2="20" y2="12" />
      </svg>
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
const hasRunningJobs = computed(() => jobsStore.totalRunningCount > 0)

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
  --task-gradient-dark: var(--text-soft);
  --task-gradient-light: var(--surface-color);

  position: relative;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 1px solid transparent;
  background: transparent;
  color: var(--text-color);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s ease;
}

.task-button:hover {
  background: var(--hover-bg, rgba(148, 163, 184, 0.24));
}

.task-button.running {
  --task-gradient-dark: var(--primary-hover);
  --task-gradient-light: var(--surface-color);
}

.task-icon {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  color: var(--task-gradient-dark);
  background: conic-gradient(from 90deg, var(--task-gradient-light) 0deg, var(--task-gradient-light) 90deg, var(--task-gradient-dark) 360deg);
  pointer-events: none;
}

.task-button.running .task-icon {
  animation: task-spin 1s linear infinite;
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

@keyframes task-spin {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 640px) {
  .task-popover {
    width: min(340px, calc(100vw - 24px));
  }
}
</style>

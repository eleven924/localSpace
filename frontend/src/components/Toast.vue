<template>
  <Transition name="toast">
    <div v-if="show" class="toast" :class="[`toast-${type}`, position]">
      <div class="toast-icon">
        <svg v-if="type === 'success'" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
          <polyline points="22 4 12 14.01 9 11.01"></polyline>
        </svg>
        <svg v-else-if="type === 'error'" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="15" y1="9" x2="9" y2="15"></line>
          <line x1="9" y1="9" x2="15" y2="15"></line>
        </svg>
        <svg v-else-if="type === 'warning'" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
          <line x1="12" y1="9" x2="12" y2="13"></line>
          <line x1="12" y1="17" x2="12.01" y2="17"></line>
        </svg>
        <svg v-else width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="12" y1="16" x2="12" y2="12"></line>
          <line x1="12" y1="8" x2="12.01" y2="8"></line>
        </svg>
      </div>
      <div class="toast-content">
        <h4 v-if="title" class="toast-title">{{ title }}</h4>
        <p class="toast-message">{{ message }}</p>
      </div>
      <button v-if="dismissible" @click="dismiss" class="toast-close" aria-label="Close">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>
      <div v-if="duration > 0" class="toast-progress" :style="{ animationDuration: `${duration}ms` }"></div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'

interface Props {
  show: boolean
  message: string
  title?: string
  type?: 'success' | 'error' | 'warning' | 'info'
  position?: 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left' | 'top' | 'bottom'
  duration?: number
  dismissible?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  type: 'info',
  position: 'top-right',
  duration: 3000,
  dismissible: true
})

const emit = defineEmits<{
  'update:show': [value: boolean]
  dismiss: []
}>()

let timer: number | null = null

const dismiss = () => {
  emit('dismiss')
  emit('update:show', false)
}

const startTimer = () => {
  if (props.duration > 0) {
    timer = window.setTimeout(() => {
      dismiss()
    }, props.duration)
  }
}

const clearTimer = () => {
  if (timer) {
    clearTimeout(timer)
    timer = null
  }
}

watch(() => props.show, (show) => {
  if (show) {
    startTimer()
  } else {
    clearTimer()
  }
})

onMounted(() => {
  if (props.show) {
    startTimer()
  }
})

onUnmounted(() => {
  clearTimer()
})
</script>

<style scoped>
.toast {
  position: fixed;
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px;
  background-color: var(--surface-color);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 9999;
  max-width: 400px;
  min-width: 300px;
}

/* Position variants */
.toast.top-right {
  top: 20px;
  right: 20px;
}

.toast.top-left {
  top: 20px;
  left: 20px;
}

.toast.bottom-right {
  bottom: 20px;
  right: 20px;
}

.toast.bottom-left {
  bottom: 20px;
  left: 20px;
}

.toast.top {
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
}

.toast.bottom {
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
}

/* Type variants */
.toast-success {
  border-left: 4px solid var(--success-color);
}

.toast-success .toast-icon {
  color: var(--success-color);
}

.toast-error {
  border-left: 4px solid var(--error-color);
}

.toast-error .toast-icon {
  color: var(--error-color);
}

.toast-warning {
  border-left: 4px solid var(--warning-color);
}

.toast-warning .toast-icon {
  color: var(--warning-color);
}

.toast-info {
  border-left: 4px solid var(--primary-color);
}

.toast-info .toast-icon {
  color: var(--primary-color);
}

.toast-icon {
  flex-shrink: 0;
  margin-top: 2px;
}

.toast-content {
  flex: 1;
  min-width: 0;
}

.toast-title {
  font-size: 14px;
  font-weight: 600;
  margin: 0 0 4px 0;
  color: var(--text-color);
}

.toast-message {
  font-size: 13px;
  line-height: 1.4;
  margin: 0;
  color: var(--text-color);
}

.toast-close {
  flex-shrink: 0;
  background: none;
  border: none;
  color: var(--text-color);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s;
}

.toast-close:hover {
  background-color: var(--border-color);
}

.toast-progress {
  position: absolute;
  bottom: 0;
  left: 0;
  height: 2px;
  background-color: var(--border-color);
  animation: toast-progress linear forwards;
}

@keyframes toast-progress {
  from {
    width: 100%;
  }
  to {
    width: 0%;
  }
}

/* Toast transitions */
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from {
  opacity: 0;
  transform: translateY(-20px);
}

.toast-leave-to {
  opacity: 0;
  transform: translateY(20px);
}

/* Position-specific transitions */
.toast.bottom-right .toast-enter-from,
.toast.bottom-left .toast-enter-from {
  transform: translateY(20px);
}

.toast.bottom-right .toast-leave-to,
.toast.bottom-left .toast-leave-to {
  transform: translateY(-20px);
}

/* Responsive adjustments */
@media (max-width: 480px) {
  .toast {
    left: 10px;
    right: 10px;
    max-width: calc(100vw - 20px);
    min-width: auto;
  }

  .toast.top,
  .toast.bottom {
    transform: none;
  }

  .toast.top {
    top: 10px;
  }

  .toast.bottom {
    bottom: 10px;
  }

  .toast.top-right,
  .toast.top-left,
  .toast.bottom-right,
  .toast.bottom-left {
    top: auto;
    bottom: auto;
    right: 10px;
    left: 10px;
  }
}

/* Reduced motion preference */
@media (prefers-reduced-motion: reduce) {
  .toast-enter-active,
  .toast-leave-active {
    transition: opacity 0.3s ease;
  }

  .toast-enter-from,
  .toast-leave-to {
    transform: none;
  }

  .toast-progress {
    animation: none;
  }
}
</style>
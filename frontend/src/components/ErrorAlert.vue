<template>
  <div v-if="show" class="error-alert" :class="`error-${severity}`">
    <div class="error-content">
      <div class="error-icon">
        <svg v-if="severity === 'error'" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="12" y1="8" x2="12" y2="12"></line>
          <line x1="12" y1="16" x2="12.01" y2="16"></line>
        </svg>
        <svg v-else-if="severity === 'warning'" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
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
      <div class="error-text">
        <h4 v-if="title" class="error-title">{{ title }}</h4>
        <p class="error-message">{{ message }}</p>
      </div>
      <button v-if="dismissible" @click="dismiss" class="error-close" aria-label="Close">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>
    </div>
    <div v-if="actions && actions.length > 0" class="error-actions">
      <button
        v-for="(action, index) in actions"
        :key="index"
        @click="action.handler"
        class="btn btn-sm"
        :class="action.variant || 'primary'"
      >
        {{ action.text }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface ErrorAction {
  text: string
  handler: () => void
  variant?: string
}

interface Props {
  message: string
  title?: string
  severity?: 'error' | 'warning' | 'info'
  dismissible?: boolean
  actions?: ErrorAction[]
}

const props = withDefaults(defineProps<Props>(), {
  severity: 'error',
  dismissible: true,
  actions: () => []
})

const emit = defineEmits<{
  dismiss: []
}>()

const show = ref(true)

const dismiss = () => {
  show.value = false
  emit('dismiss')
}
</script>

<style scoped>
.error-alert {
  background-color: var(--surface-color);
  border-left: 4px solid var(--error-color);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px var(--shadow-color);
}

.error-alert.error-warning {
  border-left-color: var(--warning-color);
}

.error-alert.error-info {
  border-left-color: var(--primary-color);
}

.error-content {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.error-icon {
  flex-shrink: 0;
  color: currentColor;
  margin-top: 2px;
}

.error-alert.error-warning .error-icon {
  color: var(--warning-color);
}

.error-alert.error-info .error-icon {
  color: var(--primary-color);
}

.error-text {
  flex: 1;
  min-width: 0;
}

.error-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 4px 0;
  color: var(--text-color);
}

.error-message {
  font-size: 14px;
  line-height: 1.5;
  margin: 0;
  color: var(--text-color);
}

.error-close {
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
}

.error-close:hover {
  background-color: var(--border-color);
}

.error-actions {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}

.btn-sm {
  font-size: 13px;
  padding: 6px 12px;
}

@media (max-width: 480px) {
  .error-content {
    flex-direction: column;
  }

  .error-close {
    position: absolute;
    top: 8px;
    right: 8px;
  }

  .error-actions {
    flex-direction: column;
  }

  .error-actions .btn {
    width: 100%;
  }
}
</style>
<template>
  <div class="loading-spinner" :class="{ [`size-${size}`]: size, inline }">
    <div class="spinner"></div>
    <p v-if="text" class="loading-text">{{ text }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  size?: 'small' | 'medium' | 'large'
  text?: string
  inline?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  size: 'medium',
  text: '',
  inline: false
})
</script>

<style scoped>
.loading-spinner {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.loading-spinner.inline {
  flex-direction: row;
  gap: 8px;
  padding: 0;
}

.loading-spinner.inline .spinner {
  width: 16px;
  height: 16px;
  border-width: 2px;
}

.loading-spinner.inline .loading-text {
  font-size: 14px;
}

.loading-spinner.size-small .spinner {
  width: 20px;
  height: 20px;
  border-width: 2px;
}

.loading-spinner.size-small .loading-text {
  font-size: 12px;
}

.loading-spinner.size-medium .spinner {
  width: 40px;
  height: 40px;
  border-width: 3px;
}

.loading-spinner.size-medium .loading-text {
  font-size: 14px;
}

.loading-spinner.size-large .spinner {
  width: 60px;
  height: 60px;
  border-width: 4px;
}

.loading-spinner.size-large .loading-text {
  font-size: 16px;
}

.spinner {
  border: 3px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

.loading-text {
  margin-top: 12px;
  color: var(--text-color);
  font-weight: 500;
  text-align: center;
}

.loading-spinner.inline .loading-text {
  margin-top: 0;
  margin-left: 8px;
}

/* Dark mode adjustments */
[data-theme='dark'] .spinner {
  border-color: var(--surface-color);
  border-top-color: var(--primary-color);
}

/* Reduced motion preference */
@media (prefers-reduced-motion: reduce) {
  .spinner {
    animation: none;
    border-top-color: var(--border-color);
  }
}
</style>
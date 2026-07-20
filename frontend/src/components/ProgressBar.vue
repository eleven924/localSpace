<template>
  <div class="progress-bar" :class="{ striped, animated, small }">
    <div
      class="progress-fill"
      :class="variant"
      :style="{ width: `${normalizedPercentage}%` }"
      role="progressbar"
      :aria-valuenow="normalizedPercentage"
      :aria-valuemin="0"
      :aria-valuemax="100"
    >
      <span v-if="showLabel && normalizedPercentage >= 10" class="progress-text">
        {{ normalizedPercentage }}%
      </span>
    </div>
    <div v-if="showLabel && normalizedPercentage < 10" class="progress-label">
      {{ normalizedPercentage }}%
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  percentage: number
  variant?: 'primary' | 'success' | 'warning' | 'error' | 'info'
  striped?: boolean
  animated?: boolean
  showLabel?: boolean
  small?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  striped: false,
  animated: false,
  showLabel: false,
  small: false,
})

const normalizedPercentage = computed(() => Math.max(0, Math.min(100, props.percentage)))
</script>

<style scoped>
.progress-bar {
  position: relative;
  width: 100%;
  height: 10px;
  border-radius: 999px;
  background-color: var(--surface-muted);
  overflow: hidden;
}

.progress-bar.small {
  height: 6px;
}

.progress-fill {
  height: 100%;
  border-radius: inherit;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding-right: 8px;
  transition: width 0.24s ease;
}

.progress-fill.primary,
.progress-fill.info {
  background-color: var(--primary-color);
}

.progress-fill.success {
  background-color: var(--success-color);
}

.progress-fill.warning {
  background-color: var(--warning-color);
}

.progress-fill.error {
  background-color: var(--error-color);
}

.progress-bar.striped .progress-fill {
  background-image:
    linear-gradient(
      45deg,
      rgba(255, 255, 255, 0.18) 25%,
      transparent 25%,
      transparent 50%,
      rgba(255, 255, 255, 0.18) 50%,
      rgba(255, 255, 255, 0.18) 75%,
      transparent 75%,
      transparent
    );
  background-size: 1rem 1rem;
}

.progress-bar.animated.striped .progress-fill {
  animation: progress-bar-stripes 1s linear infinite;
}

.progress-text,
.progress-label {
  font-size: 11px;
  font-weight: 700;
}

.progress-text {
  color: #fff;
}

.progress-label {
  position: absolute;
  left: 10px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-soft);
}

@keyframes progress-bar-stripes {
  from {
    background-position: 1rem 0;
  }

  to {
    background-position: 0 0;
  }
}
</style>

<template>
  <div class="progress-bar" :class="{ striped, animated, small }">
    <div
      class="progress-fill"
      :class="variant"
      :style="{ width: `${percentage}%` }"
      role="progressbar"
      :aria-valuenow="percentage"
      :aria-valuemin="0"
      :aria-valuemax="100"
    >
      <span v-if="showLabel && percentage >= 10" class="progress-text">
        {{ percentage }}%
      </span>
    </div>
    <div v-if="showLabel && percentage < 10" class="progress-label">
      {{ percentage }}%
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
  small: false
})

// Ensure percentage is between 0 and 100
const normalizedPercentage = computed(() => {
  return Math.max(0, Math.min(100, props.percentage))
})
</script>

<style scoped>
.progress-bar {
  position: relative;
  width: 100%;
  height: 24px;
  background-color: var(--surface-color);
  border-radius: 12px;
  overflow: hidden;
  box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.1);
}

.progress-bar.small {
  height: 8px;
  border-radius: 4px;
}

.progress-bar.small .progress-text,
.progress-bar.small .progress-label {
  display: none;
}

.progress-fill {
  height: 100%;
  border-radius: 12px;
  transition: width 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding-right: 8px;
}

.progress-bar.small .progress-fill {
  border-radius: 4px;
}

/* Variants */
.progress-fill.primary {
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

.progress-fill.info {
  background-color: #2196F3;
}

/* Striped effect */
.progress-bar.striped .progress-fill {
  background-image: linear-gradient(
    45deg,
    rgba(255, 255, 255, 0.15) 25%,
    transparent 25%,
    transparent 50%,
    rgba(255, 255, 255, 0.15) 50%,
    rgba(255, 255, 255, 0.15) 75%,
    transparent 75%,
    transparent
  );
  background-size: 1rem 1rem;
}

/* Animated striped effect */
.progress-bar.animated.striped .progress-fill {
  animation: progress-bar-stripes 1s linear infinite;
}

@keyframes progress-bar-stripes {
  0% {
    background-position: 1rem 0;
  }
  100% {
    background-position: 0 0;
  }
}

.progress-text {
  color: white;
  font-size: 12px;
  font-weight: 600;
  line-height: 1;
}

.progress-label {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  color: var(--text-color);
  font-size: 12px;
  font-weight: 600;
}

/* Dark mode adjustments */
[data-theme='dark'] .progress-bar {
  background-color: var(--surface-color);
  box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.3);
}

[data-theme='dark'] .progress-bar.striped .progress-fill {
  background-image: linear-gradient(
    45deg,
    rgba(0, 0, 0, 0.15) 25%,
    transparent 25%,
    transparent 50%,
    rgba(0, 0, 0, 0.15) 50%,
    rgba(0, 0, 0, 0.15) 75%,
    transparent 75%,
    transparent
  );
}

/* Reduced motion preference */
@media (prefers-reduced-motion: reduce) {
  .progress-fill {
    transition: width 0.5s ease;
  }

  .progress-bar.animated.striped .progress-fill {
    animation: none;
  }
}

/* Responsive adjustments */
@media (max-width: 480px) {
  .progress-bar {
    height: 20px;
  }

  .progress-text {
    font-size: 11px;
    padding-right: 6px;
  }

  .progress-label {
    font-size: 11px;
  }
}
</style>
<template>
  <div class="empty-state">
    <div class="empty-icon">
      <slot name="icon">
        <!-- Default empty folder icon -->
        <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
        </svg>
      </slot>
    </div>
    <h3 class="empty-title">{{ title }}</h3>
    <p class="empty-description">{{ description }}</p>
    <div v-if="$slots.actions" class="empty-actions">
      <slot name="actions"></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  title?: string
  description?: string
}

const props = withDefaults(defineProps<Props>(), {
  title: 'No items found',
  description: 'There are no items to display at the moment.'
})
</script>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  text-align: center;
  min-height: 300px;
}

.empty-icon {
  margin-bottom: 16px;
  color: var(--text-color);
  opacity: 0.5;
}

.empty-icon svg {
  width: 64px;
  height: 64px;
}

.empty-title {
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 8px 0;
  color: var(--text-color);
}

.empty-description {
  font-size: 14px;
  line-height: 1.5;
  margin: 0 0 24px 0;
  color: var(--text-color);
  opacity: 0.8;
  max-width: 400px;
}

.empty-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  justify-content: center;
}

/* Responsive adjustments */
@media (max-width: 480px) {
  .empty-state {
    padding: 32px 16px;
    min-height: 200px;
  }

  .empty-icon svg {
    width: 48px;
    height: 48px;
  }

  .empty-title {
    font-size: 16px;
  }

  .empty-description {
    font-size: 13px;
  }

  .empty-actions {
    flex-direction: column;
    width: 100%;
    max-width: 200px;
  }

  .empty-actions .btn {
    width: 100%;
  }
}

/* Dark mode adjustments */
[data-theme='dark'] .empty-icon {
  opacity: 0.6;
}
</style>
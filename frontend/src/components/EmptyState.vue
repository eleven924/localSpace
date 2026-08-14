<template>
  <div class="empty-state">
    <div class="empty-icon">
      <slot name="icon">
        <!-- 与资料库空状态复用同一套文件夹轮廓，避免页面切换时图标发生跳变。 -->
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.6"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path d="M3 7.5A2.5 2.5 0 0 1 5.5 5h4l2 2h7A2.5 2.5 0 0 1 21 9.5v8A2.5 2.5 0 0 1 18.5 20h-13A2.5 2.5 0 0 1 3 17.5z"></path>
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
  flex: 1;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  min-height: 0;
  padding: 56px 24px;
  text-align: center;
  color: var(--text-soft);
}

.empty-icon {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 86px;
  height: 86px;
  border: 1px solid var(--border-color);
  border-radius: 18px;
  background: var(--surface-muted);
  color: var(--text-faint);
}

.empty-icon svg {
  width: 40px;
  height: 40px;
}

.empty-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color);
}

.empty-description {
  /* 合集说明略长，保持单行后才能与资料库空状态获得相同的垂直几何中心。 */
  max-width: 460px;
  margin: 0;
  color: var(--text-soft);
  line-height: 1.7;
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
    padding: 36px 18px;
  }

  .empty-title {
    font-size: 24px;
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

</style>

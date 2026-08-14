<template>
  <div class="display-mode-toggle" role="tablist" aria-label="列表展示模式">
    <button type="button" :class="{ active: modelValue === 'flat' }" @click="emit('update:modelValue', 'flat')">
      <!-- 固定 viewBox 的 SVG 不受字体字面框影响，图标始终位于按钮几何中心。 -->
      <svg class="display-mode-icon display-mode-icon-grid" viewBox="0 0 18 18" aria-hidden="true">
        <rect x="2.5" y="2.5" width="4" height="4" rx="1" />
        <rect x="11.5" y="2.5" width="4" height="4" rx="1" />
        <rect x="2.5" y="11.5" width="4" height="4" rx="1" />
        <rect x="11.5" y="11.5" width="4" height="4" rx="1" />
      </svg>
      平铺
    </button>
    <button
      type="button"
      :class="{ active: modelValue === 'grouped' }"
      @click="emit('update:modelValue', 'grouped')"
    >
      <!-- 三条等距短线表达按合集分组，和网格图标保持同一视觉重量。 -->
      <svg class="display-mode-icon display-mode-icon-grouped" viewBox="0 0 18 18" aria-hidden="true">
        <path d="M3 4.25h12M3 9h12M3 13.75h12" />
      </svg>
      按合集分组
    </button>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  modelValue: 'flat' | 'grouped'
}>()

const emit = defineEmits<{
  'update:modelValue': [value: 'flat' | 'grouped']
}>()
</script>

<style scoped>
.display-mode-toggle {
  display: inline-flex;
  gap: 4px;
  padding: 3px;
  border-radius: 10px;
  border: 1px solid var(--border-color);
  background-color: var(--surface-color);
}

.display-mode-toggle button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  min-width: 34px;
  min-height: 34px;
  padding: 0;
  border-radius: 8px;
  font-size: 0;
  color: var(--text-soft);
}

.display-mode-icon {
  display: block;
  width: 18px;
  height: 18px;
  overflow: visible;
  fill: currentColor;
  color: currentColor;
}

.display-mode-icon-grouped {
  fill: none;
  stroke: currentColor;
  stroke-width: 1.7;
  stroke-linecap: round;
}

.display-mode-toggle button.active {
  background-color: rgba(45, 140, 240, 0.08);
  color: var(--primary-color);
}

@media (max-width: 640px) {
  .display-mode-toggle {
    width: 100%;
  }

  .display-mode-toggle button {
    flex: 1;
    min-width: 34px;
  }
}
</style>

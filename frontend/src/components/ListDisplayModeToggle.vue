<template>
  <div class="display-mode-toggle" role="tablist" aria-label="列表展示模式">
    <button type="button" :class="{ active: modelValue === 'flat' }" @click="emit('update:modelValue', 'flat')">
      平铺
    </button>
    <button
      type="button"
      :class="{ active: modelValue === 'grouped' }"
      @click="emit('update:modelValue', 'grouped')"
    >
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
  min-width: 84px;
  min-height: 34px;
  padding: 7px 12px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-soft);
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
    min-width: 0;
  }
}
</style>

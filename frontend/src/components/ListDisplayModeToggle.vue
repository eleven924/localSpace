<template>
  <div class="display-mode-toggle" role="tablist" aria-label="List display mode">
    <button
      type="button"
      :class="{ active: modelValue === 'flat' }"
      @click="emit('update:modelValue', 'flat')"
    >
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
  align-items: center;
  gap: 4px;
  padding: 4px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--surface-color) 88%, transparent);
  border: 1px solid rgba(148, 163, 184, 0.18);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.16);
}

.display-mode-toggle button {
  min-width: 96px;
  padding: 8px 14px;
  border-radius: 999px;
  background: transparent;
  color: var(--text-color);
  font-size: 13px;
  font-weight: 600;
}

.display-mode-toggle button.active {
  background: linear-gradient(135deg, var(--primary-color) 0%, color-mix(in srgb, var(--primary-color) 78%, #0f172a) 100%);
  color: #fff;
  box-shadow: 0 10px 20px rgba(33, 150, 243, 0.2);
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

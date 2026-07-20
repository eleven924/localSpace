<template>
  <div class="file-type-filter">
    <button
      v-for="type in fileTypes"
      :key="type.value"
      type="button"
      :class="{ active: currentType === type.value }"
      :title="type.label"
      @click="handleFilter(type.value)"
    >
      <span class="filter-icon">{{ type.icon }}</span>
      <span class="filter-label">{{ type.label }}</span>
      <span v-if="counts?.[type.value]" class="filter-count">{{ counts[type.value] }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { FILE_TYPES } from '@/utils/constants'

const props = defineProps<{
  modelValue: string
  counts?: Record<string, number>
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  filter: [value: string]
}>()

const currentType = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

const fileTypes = FILE_TYPES

const handleFilter = (type: string) => {
  currentType.value = type
  emit('filter', type)
}
</script>

<style scoped>
.file-type-filter {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.file-type-filter button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 36px;
  padding: 7px 12px;
  border-radius: 16px;
  border: 1px solid var(--border-color);
  background-color: var(--surface-color);
  color: var(--text-soft);
}

.file-type-filter button:hover {
  background-color: var(--surface-muted);
  color: var(--text-color);
}

.file-type-filter button.active {
  background-color: rgba(45, 140, 240, 0.08);
  border-color: rgba(45, 140, 240, 0.28);
  color: var(--primary-color);
}

.filter-icon {
  font-size: 14px;
}

.filter-label {
  font-size: 13px;
  font-weight: 600;
}

.filter-count {
  min-width: 20px;
  padding: 1px 6px;
  border-radius: 10px;
  background-color: var(--surface-muted);
  color: var(--text-faint);
  font-size: 11px;
  font-weight: 700;
}

@media (max-width: 480px) {
  .filter-label {
    display: none;
  }
}
</style>

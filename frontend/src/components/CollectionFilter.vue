<template>
  <div v-if="options.length > 0" class="collection-filter">
    <button type="button" :class="{ active: currentValue === 'all' }" @click="handleFilter('all')">
      <span class="collection-filter-label">全部合集</span>
      <span class="collection-filter-count">{{ totalCount }}</span>
    </button>

    <button
      v-for="option in options"
      :key="option.value"
      type="button"
      :class="{ active: currentValue === option.value }"
      :title="option.label"
      @click="handleFilter(option.value)"
    >
      <span class="collection-filter-label">{{ option.label }}</span>
      <span class="collection-filter-count">{{ option.count }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { CollectionSummary } from '@/types'

const props = defineProps<{
  modelValue: string
  options: CollectionSummary[]
  totalCount?: number
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  filter: [value: string]
}>()

const currentValue = computed({
  get: () => props.modelValue,
  set: (value: string) => emit('update:modelValue', value),
})

const totalCount = computed(() => {
  return props.totalCount ?? props.options.reduce((sum, option) => sum + option.count, 0)
})

const handleFilter = (value: string) => {
  currentValue.value = value
  emit('filter', value)
}
</script>

<style scoped>
.collection-filter {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 2px;
}

.collection-filter button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  min-height: 36px;
  padding: 7px 12px;
  border-radius: 16px;
  border: 1px solid var(--border-color);
  background-color: var(--surface-color);
  color: var(--text-soft);
}

.collection-filter button:hover {
  background-color: var(--surface-muted);
  color: var(--text-color);
}

.collection-filter button.active {
  background-color: rgba(45, 140, 240, 0.08);
  border-color: rgba(45, 140, 240, 0.28);
  color: var(--primary-color);
}

.collection-filter-label {
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
  font-weight: 600;
}

.collection-filter-count {
  min-width: 20px;
  padding: 1px 6px;
  border-radius: 10px;
  background-color: var(--surface-muted);
  color: var(--text-faint);
  font-size: 11px;
  font-weight: 700;
}

@media (max-width: 768px) {
  .collection-filter-label {
    max-width: 120px;
  }
}
</style>

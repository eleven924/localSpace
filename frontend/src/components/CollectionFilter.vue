<template>
  <div v-if="options.length > 0" class="collection-filter">
    <button
      type="button"
      :class="{ active: currentValue === 'all' }"
      @click="handleFilter('all')"
    >
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
  set: (value: string) => {
    emit('update:modelValue', value)
  },
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
  gap: 10px;
  overflow-x: auto;
  overflow-y: hidden;
  padding: 4px 0 10px;
  scrollbar-width: thin;
}

.collection-filter button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  min-width: 0;
  padding: 9px 14px;
  border-radius: 999px;
  border: 1px solid rgba(148, 163, 184, 0.22);
  background: color-mix(in srgb, var(--surface-color) 84%, transparent);
  color: var(--text-color);
  transition: all 0.2s ease;
}

.collection-filter button:hover {
  transform: translateY(-1px);
  border-color: color-mix(in srgb, var(--primary-color) 35%, var(--border-color));
  box-shadow: 0 10px 22px rgba(15, 23, 42, 0.08);
}

.collection-filter button.active {
  background: linear-gradient(135deg, var(--primary-color) 0%, color-mix(in srgb, var(--primary-color) 78%, #0f172a) 100%);
  border-color: transparent;
  color: #fff;
  box-shadow: 0 12px 26px rgba(33, 150, 243, 0.24);
}

.collection-filter-label {
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
  font-weight: 600;
}

.collection-filter-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 24px;
  padding: 2px 8px;
  border-radius: 999px;
  background: rgba(15, 23, 42, 0.08);
  font-size: 12px;
  font-weight: 700;
}

.collection-filter button.active .collection-filter-count {
  background: rgba(255, 255, 255, 0.18);
}

@media (max-width: 768px) {
  .collection-filter {
    gap: 8px;
    padding-bottom: 8px;
  }

  .collection-filter button {
    padding: 8px 12px;
  }

  .collection-filter-label {
    max-width: 120px;
  }
}
</style>

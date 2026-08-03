<template>
  <div class="collection-filter">
    <button type="button" :class="{ active: currentValue === 'all' }" @click="handleFilter('all')">
      <span class="collection-filter-label">全部</span>
      <span class="collection-filter-count">{{ totalCount }}</span>
    </button>

    <button type="button" :class="{ active: currentValue === 'unsorted' }" @click="handleFilter('unsorted')">
      <span class="collection-filter-label">未分配</span>
      <span class="collection-filter-count">{{ unsortedCount }}</span>
    </button>

    <button
      v-for="collection in collections"
      :key="collection.id"
      type="button"
      :class="{ active: currentValue === collection.id }"
      :title="collection.name"
      @click="handleFilter(collection.id)"
    >
      <span class="collection-filter-label">{{ collection.name }}</span>
      <span class="collection-filter-count">{{ countMap[collection.id] ?? 0 }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Collection, File } from '@/types'

const props = defineProps<{
  modelValue: 'all' | 'unsorted' | number
  collections: Collection[]
  files: File[]
  totalCount?: number
}>()

const emit = defineEmits<{
  'update:modelValue': [value: 'all' | 'unsorted' | number]
  filter: [value: 'all' | 'unsorted' | number]
}>()

const currentValue = computed({
  get: () => props.modelValue,
  set: (value: 'all' | 'unsorted' | number) => emit('update:modelValue', value),
})

const countMap = computed(() => {
  const map: Record<number, number> = {}
  props.files.forEach((file) => {
    if (file.collectionId) {
      map[file.collectionId] = (map[file.collectionId] || 0) + 1
    }
  })
  return map
})

const unsortedCount = computed(() => {
  return props.files.filter((file) => !file.collectionId).length
})

const totalCount = computed(() => {
  return props.totalCount ?? props.files.length
})

const handleFilter = (value: 'all' | 'unsorted' | number) => {
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

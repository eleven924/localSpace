<template>
  <div class="file-type-filter">
    <button
      v-for="type in fileTypes"
      :key="type.value"
      :class="{ active: currentType === type.value }"
      @click="handleFilter(type.value)"
      :title="type.label"
    >
      <span class="filter-icon">{{ type.icon }}</span>
      <span class="filter-label">{{ type.label }}</span>
      <span v-if="counts[type.value]" class="filter-count">{{ counts[type.value] }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { FILE_TYPES } from '@/utils/constants'

interface FileType {
  value: string
  label: string
  icon: string
}

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
  set: (value) => {
    emit('update:modelValue', value)
  }
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
  gap: 8px;
  flex-wrap: wrap;
  padding: 8px 0;
}

.file-type-filter button {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background-color: var(--surface-color);
  color: var(--text-color);
  border: 1px solid var(--border-color);
  border-radius: 20px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 14px;
  white-space: nowrap;
}

.file-type-filter button:hover {
  background-color: var(--border-color);
  transform: translateY(-1px);
}

.file-type-filter button:active {
  transform: translateY(0);
}

.file-type-filter button.active {
  background-color: var(--primary-color);
  color: white;
  border-color: var(--primary-color);
  box-shadow: 0 2px 8px var(--shadow-color);
}

.filter-icon {
  font-size: 16px;
}

.filter-label {
  font-weight: 500;
}

.filter-count {
  background-color: rgba(0, 0, 0, 0.2);
  padding: 2px 6px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 600;
  min-width: 20px;
  text-align: center;
}

.file-type-filter button.active .filter-count {
  background-color: rgba(255, 255, 255, 0.2);
}

@media (max-width: 768px) {
  .file-type-filter {
    gap: 6px;
  }

  .file-type-filter button {
    padding: 6px 12px;
    font-size: 13px;
  }

  .filter-icon {
    font-size: 14px;
  }

  .filter-count {
    font-size: 11px;
    padding: 1px 5px;
    min-width: 18px;
  }
}

@media (max-width: 480px) {
  .file-type-filter {
    gap: 4px;
  }

  .file-type-filter button {
    padding: 5px 10px;
    font-size: 12px;
  }

  .filter-label {
    display: none;
  }

  .filter-icon {
    font-size: 16px;
  }

  .filter-count {
    margin-left: -2px;
  }
}
</style>
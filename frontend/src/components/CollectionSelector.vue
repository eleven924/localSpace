<template>
  <select :value="normalizedValue" @change="handleChange">
    <option :value="0">未分配合集</option>
    <option v-for="collection in collections" :key="collection.id" :value="collection.id">
      {{ collection.name }}
    </option>
  </select>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Collection } from '@/types'

const props = defineProps<{
  modelValue?: number | null
  collections: Collection[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: number | undefined]
}>()

const normalizedValue = computed(() => props.modelValue || 0)

const handleChange = (event: Event) => {
  const target = event.target as HTMLSelectElement
  const value = parseInt(target.value, 10)
  emit('update:modelValue', value > 0 ? value : undefined)
}
</script>

<style scoped>
select {
  width: 100%;
  min-height: 42px;
  padding: 10px 12px;
  border: 1px solid color-mix(in srgb, var(--border-color) 84%, transparent);
  border-radius: 10px;
  background: color-mix(in srgb, var(--surface-color) 92%, transparent);
  color: var(--text-color);
  font-size: 14px;
  font-family: inherit;
  transition: border-color 0.2s ease, box-shadow 0.2s ease, background-color 0.2s ease;
}

select:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--primary-color) 16%, transparent);
}

option {
  background: var(--surface-color);
  color: var(--text-color);
}
</style>

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

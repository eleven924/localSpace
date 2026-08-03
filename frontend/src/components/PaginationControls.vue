<template>
  <div class="pagination-controls">
    <button class="btn secondary" :disabled="page <= 1" @click="emit('change', page - 1)">上一页</button>
    <span>第 {{ page }} / {{ totalPages }} 页</span>
    <button class="btn secondary" :disabled="page >= totalPages" @click="emit('change', page + 1)">下一页</button>
    <select :value="pageSize" @change="handleSizeChange">
      <option v-for="size in pageSizes" :key="size" :value="size">{{ size }} / 页</option>
    </select>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  page: number
  pageSize: number
  total: number
  pageSizes?: number[]
}>()

const emit = defineEmits(['change', 'changeSize'])

const pageSizes = computed(() => props.pageSizes || [20, 50, 100])
const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))

const handleSizeChange = (event: Event) => {
  const target = event.target as HTMLSelectElement
  emit('changeSize', Number(target.value))
}
</script>

<style scoped>
.pagination-controls {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  font-size: 13px;
  color: var(--text-soft);
}

.pagination-controls .btn {
  width: auto;
}

.pagination-controls select {
  min-height: 36px;
  padding: 0 10px;
  border-radius: 10px;
}
</style>

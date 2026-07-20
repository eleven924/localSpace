<template>
  <div class="search-bar">
    <div class="search-input-wrapper">
      <span class="search-icon">⌕</span>
      <input
        v-model="query"
        type="text"
        class="search-input"
        placeholder="搜索文件名、标签、描述或合集"
        @keyup.enter="handleSearch"
        @input="handleInput"
      />
      <button
        v-if="query"
        class="clear-button"
        type="button"
        title="清空搜索"
        @click="handleClear"
      >
        ×
      </button>
    </div>
    <button class="btn primary search-button" type="button" @click="handleSearch">搜索</button>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  modelValue?: string
  debounce?: number
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  search: [query: string]
  clear: []
}>()

const query = ref(props.modelValue || '')
let debounceTimer: number | null = null

watch(
  () => props.modelValue,
  (newValue) => {
    query.value = newValue || ''
  }
)

const clearDebounceTimer = () => {
  if (debounceTimer) {
    clearTimeout(debounceTimer)
    debounceTimer = null
  }
}

const handleSearch = () => {
  const trimmedQuery = query.value.trim()
  emit('update:modelValue', query.value)

  if (!trimmedQuery) {
    emit('clear')
    return
  }

  emit('search', trimmedQuery)
}

const handleInput = () => {
  emit('update:modelValue', query.value)

  if (!query.value.trim()) {
    clearDebounceTimer()
    emit('clear')
    return
  }

  if (props.debounce && props.debounce > 0) {
    clearDebounceTimer()
    debounceTimer = window.setTimeout(() => {
      const trimmedQuery = query.value.trim()
      if (trimmedQuery) emit('search', trimmedQuery)
    }, props.debounce)
  }
}

const handleClear = () => {
  query.value = ''
  emit('update:modelValue', '')
  emit('clear')
  clearDebounceTimer()
}

const focus = () => {
  const input = document.querySelector('.search-input') as HTMLInputElement | null
  input?.focus()
}

defineExpose({
  focus,
})
</script>

<style scoped>
.search-bar {
  display: flex;
  gap: 10px;
  width: 100%;
}

.search-input-wrapper {
  position: relative;
  flex: 1;
}

.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-faint);
  font-size: 15px;
  pointer-events: none;
}

.search-input {
  padding-left: 34px;
  padding-right: 36px;
  min-height: 40px;
  border-radius: 10px;
}

.clear-button {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  width: 24px;
  height: 24px;
  border-radius: 50%;
  color: var(--text-faint);
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.clear-button:hover {
  background-color: var(--surface-muted);
  color: var(--text-color);
}

.search-button {
  min-width: 84px;
}

@media (max-width: 768px) {
  .search-bar {
    flex-direction: column;
  }
}
</style>

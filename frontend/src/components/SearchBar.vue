<template>
  <div class="search-bar">
    <div class="search-input-wrapper">
      <span class="search-icon">🔍</span>
      <input
        v-model="query"
        type="text"
        class="search-input"
        placeholder="搜索文件名、标签、描述..."
        @keyup.enter="handleSearch"
        @input="handleInput"
      />
      <button
        v-if="query"
        class="clear-button"
        @click="handleClear"
        title="清除搜索"
      >
        ✕
      </button>
    </div>
    <button class="search-button" @click="handleSearch">
      搜索
    </button>
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

watch(() => props.modelValue, (newValue) => {
  query.value = newValue || ''
})

const handleSearch = () => {
  if (query.value.trim()) {
    emit('search', query.value)
  }
}

const handleInput = () => {
  emit('update:modelValue', query.value)

  if (props.debounce && props.debounce > 0) {
    if (debounceTimer) {
      clearTimeout(debounceTimer)
    }

    debounceTimer = window.setTimeout(() => {
      if (query.value.trim()) {
        emit('search', query.value)
      }
    }, props.debounce)
  }
}

const handleClear = () => {
  query.value = ''
  emit('update:modelValue', '')
  emit('clear')

  if (debounceTimer) {
    clearTimeout(debounceTimer)
    debounceTimer = null
  }
}

const focus = () => {
  const input = document.querySelector('.search-input') as HTMLInputElement
  if (input) {
    input.focus()
  }
}

defineExpose({
  focus,
})
</script>

<style scoped>
.search-bar {
  display: flex;
  gap: 8px;
  width: 100%;
  max-width: 800px;
  margin: 0 auto;
}

.search-input-wrapper {
  flex: 1;
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 12px;
  font-size: 16px;
  opacity: 0.5;
  pointer-events: none;
  z-index: 1;
}

.search-input {
  width: 100%;
  padding: 10px 40px 10px 36px;
  border: 1px solid var(--border-color);
  border-radius: 20px;
  background-color: var(--bg-color);
  color: var(--text-color);
  font-size: 14px;
  transition: all 0.2s ease;
}

.search-input:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px rgba(33, 150, 243, 0.1);
}

.search-input::placeholder {
  color: var(--text-color);
  opacity: 0.5;
}

.clear-button {
  position: absolute;
  right: 12px;
  background: none;
  border: none;
  color: var(--text-color);
  opacity: 0.5;
  cursor: pointer;
  padding: 4px;
  font-size: 16px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  transition: all 0.2s ease;
}

.clear-button:hover {
  opacity: 1;
  background-color: var(--border-color);
}

.search-button {
  padding: 10px 20px;
  background-color: var(--primary-color);
  color: white;
  border: none;
  border-radius: 20px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.search-button:hover {
  opacity: 0.9;
  box-shadow: 0 2px 8px var(--shadow-color);
}

.search-button:active {
  transform: scale(0.98);
}

@media (max-width: 768px) {
  .search-bar {
    flex-direction: column;
    gap: 8px;
  }

  .search-input-wrapper {
    width: 100%;
  }

  .search-button {
    width: 100%;
  }
}

@media (max-width: 480px) {
  .search-input {
    padding: 8px 36px 8px 32px;
    font-size: 13px;
  }

  .search-button {
    padding: 8px 16px;
    font-size: 13px;
  }
}
</style>
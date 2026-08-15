<template>
  <div class="pagination-controls" :class="`is-${variant}`" role="navigation" aria-label="分页导航">
    <p class="pagination-summary">
      <span>{{ totalText }}</span>
      <span class="summary-dot" aria-hidden="true"></span>
      <span>{{ rangeText }}</span>
    </p>

    <div class="pagination-actions">
      <button
        type="button"
        class="pager-button"
        :disabled="page <= 1"
        aria-label="上一页"
        title="上一页"
        @click="emit('change', page - 1)"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="M15 18l-6-6 6-6" />
        </svg>
      </button>

      <div class="page-badge" aria-live="polite">
        <strong>{{ page }}</strong>
        <span>/ {{ totalPages }}</span>
      </div>

      <button
        type="button"
        class="pager-button"
        :disabled="page >= totalPages"
        aria-label="下一页"
        title="下一页"
        @click="emit('change', page + 1)"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="M9 18l6-6-6-6" />
        </svg>
      </button>

      <div class="page-size-field" @focusout="handlePageSizeFocusOut">
        <span>每页</span>
        <button
          type="button"
          class="page-size-button"
          aria-haspopup="listbox"
          :aria-expanded="String(pageSizeMenuOpen)"
          @click="togglePageSizeMenu"
        >
          <span>{{ pageSize }} 条</span>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M6 9l6 6 6-6" />
          </svg>
        </button>

        <div v-if="pageSizeMenuOpen" class="page-size-menu" role="listbox">
          <button
            v-for="size in pageSizes"
            :key="size"
            type="button"
            class="page-size-option"
            :class="{ active: size === pageSize }"
            role="option"
            :aria-selected="String(size === pageSize)"
            @click="selectPageSize(size)"
          >
            {{ size }} 条
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

const props = defineProps<{
  page: number
  pageSize: number
  total: number
  pageSizes?: number[]
  variant?: 'default' | 'overlay'
}>()

const emit = defineEmits<{
  (event: 'change', page: number): void
  (event: 'changeSize', size: number): void
}>()

const pageSizes = computed(() => props.pageSizes || [20, 50, 100])
const variant = computed(() => props.variant || 'default')
const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const pageSizeMenuOpen = ref(false)

// 按当前页码和每页数量计算结果区间，给用户一个明确的“看到了多少”的反馈。
const startItem = computed(() => {
  if (props.total <= 0) return 0
  return (props.page - 1) * props.pageSize + 1
})

const endItem = computed(() => {
  if (props.total <= 0) return 0
  return Math.min(props.total, props.page * props.pageSize)
})

const totalText = computed(() => {
  return props.total > 0 ? `${props.total} 条` : '暂无结果'
})

const rangeText = computed(() => {
  return props.total > 0 ? `${startItem.value}-${endItem.value}` : '当前没有可翻页内容'
})

const togglePageSizeMenu = () => {
  pageSizeMenuOpen.value = !pageSizeMenuOpen.value
}

const selectPageSize = (size: number) => {
  // 使用自定义小菜单替代原生 select，避免不同系统主题下选项文字和背景对比度失控。
  pageSizeMenuOpen.value = false
  emit('changeSize', size)
}

const handlePageSizeFocusOut = (event: FocusEvent) => {
  const wrapper = event.currentTarget as HTMLElement
  requestAnimationFrame(() => {
    if (!wrapper.contains(document.activeElement)) {
      pageSizeMenuOpen.value = false
    }
  })
}
</script>

<style scoped>
.pagination-controls {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 8px 14px;
  align-items: center;
  padding: 8px 4px 0;
  color: var(--text-faint);
}

.pagination-summary {
  display: flex;
  align-items: center;
  gap: 7px;
  min-width: 0;
  flex-wrap: wrap;
  font-size: 12px;
  line-height: 1.5;
  color: var(--text-faint);
}

.summary-dot {
  width: 3px;
  height: 3px;
  border-radius: 999px;
  background: currentColor;
  opacity: 0.58;
}

.pagination-actions {
  display: inline-flex;
  align-items: center;
  justify-self: end;
  gap: 6px;
}

.pager-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border: 1px solid rgba(146, 165, 192, 0.22);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-soft);
}

.pager-button:hover:not(:disabled) {
  background: color-mix(in srgb, var(--surface-muted) 74%, transparent);
  color: var(--text-color);
}

.pager-button svg {
  flex: none;
}

.page-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  min-width: 58px;
  height: 30px;
  padding: 0 8px;
  color: var(--text-faint);
  font-size: 12px;
}

.page-badge strong {
  font-size: 13px;
  color: var(--text-color);
  min-width: 1.25ch;
  text-align: center;
}

.page-size-field {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding-left: 4px;
  color: var(--text-faint);
  font-size: 12px;
}

.page-size-button {
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  width: auto;
  min-width: 86px;
  max-width: 96px;
  height: 30px;
  min-height: 30px;
  padding: 0 8px;
  border-radius: 8px;
  border-color: rgba(146, 165, 192, 0.22);
  background-color: rgba(255, 255, 255, 0.1);
  color: var(--text-soft);
  font-size: 12px;
}

.page-size-button:hover {
  background: color-mix(in srgb, var(--surface-muted) 74%, transparent);
  color: var(--text-color);
}

.page-size-button svg {
  flex: none;
}

.page-size-menu {
  position: absolute;
  right: 0;
  bottom: calc(100% + 6px);
  z-index: 20;
  display: grid;
  min-width: 96px;
  padding: 4px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--surface-color);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.18);
}

.page-size-option {
  min-height: 28px;
  padding: 0 10px;
  border-radius: 6px;
  color: var(--text-color);
  font-size: 12px;
  text-align: left;
}

.page-size-option:hover,
.page-size-option.active {
  background: color-mix(in srgb, var(--primary-color) 12%, var(--surface-muted));
  color: var(--text-color);
}

.pagination-controls.is-overlay {
  position: absolute;
  right: 0;
  bottom: 0;
  z-index: 5;
  display: inline-flex;
  width: auto;
  max-width: calc(100% - 16px);
  padding: 4px;
  border: 1px solid rgba(146, 165, 192, 0.2);
  border-radius: 10px;
  background: rgba(17, 24, 39, 0.48);
  backdrop-filter: blur(12px);
  box-shadow: 0 8px 22px rgba(15, 23, 42, 0.18);
}

.pagination-controls.is-overlay .pagination-summary {
  display: none;
}

.pagination-controls.is-overlay .pagination-actions {
  justify-self: auto;
  gap: 3px;
}

.pagination-controls.is-overlay .pager-button,
.pagination-controls.is-overlay .page-badge,
.pagination-controls.is-overlay .page-size-button {
  height: 26px;
  min-height: 26px;
}

.pagination-controls.is-overlay .pager-button {
  width: 26px;
  border-color: rgba(203, 213, 225, 0.18);
  background: rgba(255, 255, 255, 0.07);
  color: rgba(226, 232, 240, 0.82);
}

.pagination-controls.is-overlay .pager-button:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.14);
  color: #fff;
}

.pagination-controls.is-overlay .page-badge {
  min-width: 44px;
  padding: 0 6px;
  color: rgba(226, 232, 240, 0.72);
}

.pagination-controls.is-overlay .page-badge strong {
  color: #fff;
}

.pagination-controls.is-overlay .page-size-field {
  padding-left: 3px;
  color: rgba(226, 232, 240, 0.72);
}

.pagination-controls.is-overlay .page-size-button {
  min-width: 76px;
  max-width: 84px;
  border-color: rgba(203, 213, 225, 0.18);
  background-color: rgba(255, 255, 255, 0.07);
  color: #fff;
}

.pagination-controls.is-overlay .page-size-button:hover {
  background: rgba(255, 255, 255, 0.14);
  color: #fff;
}

.pagination-controls.is-overlay .page-size-menu {
  min-width: 84px;
  border-color: rgba(203, 213, 225, 0.2);
  background: rgba(17, 24, 39, 0.96);
  box-shadow: 0 12px 28px rgba(0, 0, 0, 0.32);
}

.pagination-controls.is-overlay .page-size-option {
  color: rgba(226, 232, 240, 0.9);
}

.pagination-controls.is-overlay .page-size-option:hover,
.pagination-controls.is-overlay .page-size-option.active {
  background: rgba(59, 130, 246, 0.24);
  color: #fff;
}

@media (max-width: 860px) {
  .pagination-controls {
    grid-template-columns: 1fr;
    justify-items: stretch;
  }

  .pagination-actions {
    justify-self: stretch;
    justify-content: flex-end;
  }
}

@media (max-width: 540px) {
  .pagination-controls {
    padding-inline: 12px;
  }

  .pagination-actions {
    flex-wrap: wrap;
    justify-content: space-between;
  }

  .page-badge {
    flex: 1;
  }

  .page-size-field {
    width: 100%;
  }

  .page-size-button {
    flex: 1;
    width: 100%;
    max-width: none;
  }

  .pagination-controls.is-overlay {
    left: 8px;
    justify-content: flex-end;
  }

  .pagination-controls.is-overlay .pagination-actions {
    width: 100%;
  }

  .pagination-controls.is-overlay .page-size-field {
    width: auto;
  }

  .pagination-controls.is-overlay .page-size-button {
    width: auto;
    flex: none;
  }
}
</style>

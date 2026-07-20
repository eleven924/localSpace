<template>
  <article class="collection-card" @click="emit('open', summary.value)">
    <div class="collection-card-top">
      <div class="collection-card-title-wrap">
        <h3 class="collection-card-title" :title="summary.label">{{ summary.label }}</h3>
      </div>
      <div class="collection-card-count">{{ summary.count }}</div>
    </div>

    <div class="collection-card-body">
      <div class="collection-card-types">
        <span v-for="type in topTypes" :key="type.value" class="type-chip">
          <span class="type-chip-label">{{ getFileTypeLabel(type.value) }}</span>
          <span class="type-chip-count">{{ type.count }}</span>
        </span>
      </div>

      <p class="collection-card-meta">最近更新：{{ latestModifiedLabel }}</p>
    </div>

    <div class="collection-card-footer">
      <button type="button" class="card-action" @click.stop="emit('open', summary.value)">查看合集</button>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { CollectionSummary } from '@/types'
import { formatDate, getFileTypeLabel } from '@/utils/constants'

const props = defineProps<{
  summary: CollectionSummary
}>()

const emit = defineEmits<{
  open: [value: string]
}>()

const topTypes = computed(() => {
  return Object.entries(props.summary.fileTypes)
    .map(([value, count]) => ({ value, count }))
    .sort((left, right) => right.count - left.count)
    .slice(0, 3)
})

const latestModifiedLabel = computed(() => {
  if (!props.summary.latestModifiedAt) {
    return '暂无记录'
  }

  return formatDate(props.summary.latestModifiedAt)
})
</script>

<style scoped>
.collection-card {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 210px;
  padding: 18px;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  background-color: var(--surface-color);
  box-shadow: 0 6px 18px var(--shadow-color);
  cursor: pointer;
  transition: all 0.2s ease;
}

.collection-card:hover {
  transform: translateY(-2px);
  border-color: var(--primary-color);
  box-shadow: 0 10px 24px var(--shadow-strong);
}

.collection-card-top {
  display: flex;
  justify-content: space-between;
  gap: 16px;
}

.collection-card-title {
  font-size: 18px;
  line-height: 1.25;
  color: var(--text-color);
  word-break: break-word;
}

.collection-card-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 44px;
  height: 44px;
  padding: 0 12px;
  border-radius: 10px;
  background-color: var(--surface-muted);
  border: 1px solid var(--border-color);
  color: var(--primary-color);
  font-size: 18px;
  font-weight: 700;
}

.collection-card-body {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 14px;
}

.collection-card-types {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.type-chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 5px 9px;
  border-radius: 999px;
  background-color: var(--surface-muted);
  color: var(--text-soft);
  font-size: 12px;
  font-weight: 600;
}

.type-chip-count {
  min-width: 18px;
  padding: 1px 5px;
  border-radius: 10px;
  background-color: rgba(45, 140, 240, 0.08);
}

.collection-card-meta {
  color: var(--text-faint);
  font-size: 13px;
  line-height: 1.6;
}

.collection-card-footer {
  display: flex;
}

.card-action {
  color: var(--primary-color);
  font-size: 13px;
  font-weight: 600;
}

@media (max-width: 640px) {
  .collection-card {
    min-height: 190px;
    padding: 16px;
  }

  .collection-card-title {
    font-size: 17px;
  }
}
</style>

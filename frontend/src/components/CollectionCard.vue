<template>
  <article class="collection-card" @click="emit('open', summary.value)">
    <div class="collection-card-top">
      <div class="collection-card-title-wrap">
        <p class="collection-card-eyebrow">合集</p>
        <h3 class="collection-card-title" :title="summary.label">{{ summary.label }}</h3>
      </div>
      <div class="collection-card-count">{{ summary.count }}</div>
    </div>

    <div class="collection-card-body">
      <div class="collection-card-types">
        <span
          v-for="type in topTypes"
          :key="type.value"
          class="type-chip"
        >
          <span class="type-chip-label">{{ getFileTypeLabel(type.value) }}</span>
          <span class="type-chip-count">{{ type.count }}</span>
        </span>
      </div>

      <p class="collection-card-meta">最近更新：{{ latestModifiedLabel }}</p>
    </div>

    <div class="collection-card-footer">
      <button type="button" class="card-action" @click.stop="emit('open', summary.value)">
        查看合集
      </button>
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
  gap: 18px;
  min-height: 220px;
  padding: 22px;
  border-radius: 22px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background:
    radial-gradient(circle at top right, rgba(33, 150, 243, 0.12), transparent 32%),
    linear-gradient(180deg, color-mix(in srgb, var(--surface-color) 92%, white 8%) 0%, var(--surface-color) 100%);
  box-shadow:
    0 18px 44px rgba(15, 23, 42, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.16);
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}

.collection-card:hover {
  transform: translateY(-3px);
  border-color: rgba(33, 150, 243, 0.32);
  box-shadow:
    0 24px 56px rgba(15, 23, 42, 0.12),
    0 8px 20px rgba(33, 150, 243, 0.12);
}

.collection-card-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.collection-card-title-wrap {
  min-width: 0;
}

.collection-card-eyebrow {
  margin: 0 0 8px;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--primary-color) 72%, var(--text-color));
}

.collection-card-title {
  margin: 0;
  font-size: 20px;
  line-height: 1.3;
  color: var(--text-color);
  word-break: break-word;
}

.collection-card-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 52px;
  height: 52px;
  padding: 0 14px;
  border-radius: 16px;
  background: linear-gradient(135deg, var(--primary-color) 0%, color-mix(in srgb, var(--primary-color) 78%, #0f172a) 100%);
  color: #fff;
  font-size: 18px;
  font-weight: 800;
  box-shadow: 0 12px 24px rgba(33, 150, 243, 0.22);
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
  gap: 6px;
  padding: 6px 10px;
  border-radius: 999px;
  background: rgba(15, 23, 42, 0.06);
  color: var(--text-color);
  font-size: 12px;
  font-weight: 600;
}

.type-chip-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  padding: 2px 6px;
  border-radius: 999px;
  background: rgba(33, 150, 243, 0.16);
  color: color-mix(in srgb, var(--primary-color) 72%, #0f172a);
}

.collection-card-meta {
  margin: 0;
  color: var(--text-color);
  opacity: 0.72;
  font-size: 13px;
  line-height: 1.6;
}

.collection-card-footer {
  display: flex;
  justify-content: flex-start;
}

.card-action {
  padding: 0;
  border: none;
  background: transparent;
  color: var(--primary-color);
  font-size: 14px;
  font-weight: 700;
}

@media (max-width: 640px) {
  .collection-card {
    min-height: 200px;
    padding: 18px;
    border-radius: 18px;
  }

  .collection-card-title {
    font-size: 18px;
  }
}
</style>

<template>
  <div class="file-list-container">
    <div v-if="files.length === 0" class="empty-state">
      <div class="empty-icon" :class="`is-${emptyStateMode}`">
        <svg
          v-if="emptyStateMode === 'search'"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.8"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <circle cx="11" cy="11" r="6"></circle>
          <line x1="20" y1="20" x2="15.8" y2="15.8"></line>
          <line x1="8.5" y1="11" x2="13.5" y2="11"></line>
        </svg>
        <svg
          v-else
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.6"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path d="M3 7.5A2.5 2.5 0 0 1 5.5 5h4l2 2h7A2.5 2.5 0 0 1 21 9.5v8A2.5 2.5 0 0 1 18.5 20h-13A2.5 2.5 0 0 1 3 17.5z"></path>
        </svg>
      </div>
      <h3>{{ emptyTitle }}</h3>
       param($m) if($m.Value -like '*filesStore.loading*'){ $m.Value } else { $m.Value } 
    </div>

    <div v-else-if="groupByCollection" class="grouped-file-list scroll-soft">
      <section
        v-for="group in groupedFiles"
        :key="group.key"
        class="file-group"
      >
        <div class="file-group-header">
          <div class="file-group-title-wrap">
            <h3 class="file-group-title">{{ group.label }}</h3>
            <p class="file-group-subtitle">{{ group.files.length }} 个文件</p>
          </div>
        </div>

        <div
          class="file-list"
          :class="`columns-${columns}`"
          :style="{ gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))` }"
        >
          <FileCard
            v-for="file in group.files"
            :key="file.id"
            :file="file"
            :selectable="selectable"
            :selected="selectedIds.includes(file.id)"
            @open="handleOpen"
            @click="handleClick"
            @delete="handleDelete"
            @updated="handleUpdated"
            @toggle-select="handleToggleSelect"
          />
        </div>
      </section>
    </div>

    <div
      v-else
      class="file-list scroll-soft"
      :class="`columns-${columns}`"
      :style="{ gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))` }"
    >
      <FileCard
        v-for="file in files"
        :key="file.id"
        :file="file"
        :selectable="selectable"
        :selected="selectedIds.includes(file.id)"
        @open="handleOpen"
        @click="handleClick"
        @delete="handleDelete"
        @updated="handleUpdated"
        @toggle-select="handleToggleSelect"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import type { File as LibraryFile } from '@/types'
import {
  normalizeCollectionGroupKey,
  normalizeCollectionGroupLabel,
  UNSORTED_COLLECTION_KEY,
} from '@/utils/constants'
import FileCard from './FileCard.vue'

interface FileGroup {
  key: string
  label: string
  files: LibraryFile[]
}

const props = withDefaults(defineProps<{
  files: LibraryFile[]
  groupByCollection?: boolean
  selectable?: boolean
  selectedIds?: number[]
  emptyTitle?: string
  emptyDescription?: string
  emptyStateMode?: 'library' | 'search'
}>(), {
  groupByCollection: false,
  selectable: false,
  selectedIds: () => [],
  emptyTitle: '暂无文件',
  emptyDescription: '从导入页面添加文件后，会在这里展示。',
  emptyStateMode: 'library',
})

const emit = defineEmits<{
  open: [id: number]
  click: [file: LibraryFile]
  delete: [id: number]
  updated: [id: number]
  toggleSelect: [id: number]
}>()

const columns = ref(3)

const groupedFiles = computed<FileGroup[]>(() => {
  const groups = new Map<string, FileGroup>()

  props.files.forEach((file) => {
    const key = normalizeCollectionGroupKey(file.collectionName)
    const existing = groups.get(key) || {
      key,
      label: normalizeCollectionGroupLabel(file.collectionName),
      files: [],
    }

    existing.files.push(file)
    groups.set(key, existing)
  })

  return Array.from(groups.values()).sort((left, right) => {
    if (left.key === UNSORTED_COLLECTION_KEY) return 1
    if (right.key === UNSORTED_COLLECTION_KEY) return -1
    if (right.files.length !== left.files.length) return right.files.length - left.files.length
    return left.label.localeCompare(right.label, 'zh-CN')
  })
})

const updateColumns = () => {
  const width = window.innerWidth
  if (width < 620) {
    columns.value = 2
  } else if (width < 920) {
    columns.value = 3
  } else if (width < 1240) {
    columns.value = 4
  } else {
    columns.value = 5
  }
}

onMounted(() => {
  updateColumns()
  window.addEventListener('resize', updateColumns)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateColumns)
})

const handleOpen = (id: number) => emit('open', id)
const handleClick = (file: LibraryFile) => emit('click', file)
const handleDelete = (id: number) => emit('delete', id)
const handleUpdated = (id: number) => emit('updated', id)
const handleToggleSelect = (id: number) => emit('toggleSelect', id)
</script>

<style scoped>
.file-list-container {
  width: 100%;
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.empty-state {
  display: flex;
  flex: 1;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 56px 24px;
  text-align: center;
  color: var(--text-soft);
}

.empty-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 86px;
  height: 86px;
  border-radius: 18px;
  background: var(--surface-muted);
  border: 1px solid var(--border-color);
  color: var(--text-faint);
}

.empty-icon.is-search {
  background: rgba(235, 241, 250, 0.88);
}

.empty-icon svg {
  width: 40px;
  height: 40px;
}

.empty-state h3 {
  font-size: 24px;
  color: var(--text-color);
}

.empty-state p {
  max-width: 420px;
  line-height: 1.7;
}

.grouped-file-list {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 16px;
  overflow-y: auto;
  overflow-x: clip;
  padding-right: 4px;
}

.file-group {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.file-group-header {
  padding: 12px 16px;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  background: var(--surface-color);
  box-shadow: 0 4px 14px var(--shadow-color);
}

.file-group-title {
  font-size: 18px;
  color: var(--text-color);
}

.file-group-subtitle {
  margin-top: 4px;
  color: var(--text-faint);
  font-size: 12px;
}

.file-list {
  display: grid;
  gap: 12px;
  align-content: start;
  align-items: stretch;
  overflow-y: auto;
  overflow-x: clip;
  padding-right: 4px;
}

.grouped-file-list .file-list {
  overflow: visible;
}

.file-list.columns-5 {
  gap: 12px;
}

@media (max-width: 640px) {
  .empty-state {
    padding: 36px 18px;
  }

  .empty-state h3 {
    font-size: 24px;
  }

  .file-group-header {
    padding: 12px 14px;
  }

  .file-list {
    gap: 10px;
  }
}
</style>


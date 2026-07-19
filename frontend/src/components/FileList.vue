<template>
  <div class="file-list-container">
    <div v-if="files.length === 0" class="empty-state">
      <div class="empty-icon">馃搧</div>
      <h3>暂无文件</h3>
      <p>点击上方“导入文件”开始添加内容</p>
    </div>

    <div
      v-else-if="groupByCollection"
      class="grouped-file-list"
    >
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
            @open="handleOpen"
            @click="handleClick"
            @delete="handleDelete"
            @updated="handleUpdated"
          />
        </div>
      </section>
    </div>

    <div
      v-else
      class="file-list"
      :class="`columns-${columns}`"
      :style="{ gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))` }"
    >
      <FileCard
        v-for="file in files"
        :key="file.id"
        :file="file"
        @open="handleOpen"
        @click="handleClick"
        @delete="handleDelete"
        @updated="handleUpdated"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import type { File as LibraryFile } from '@/types'
import { UNSORTED_COLLECTION_KEY, UNSORTED_COLLECTION_LABEL } from '@/utils/constants'
import FileCard from './FileCard.vue'

interface FileGroup {
  key: string
  label: string
  files: LibraryFile[]
}

const props = withDefaults(defineProps<{
  files: LibraryFile[]
  groupByCollection?: boolean
}>(), {
  groupByCollection: false,
})

const emit = defineEmits<{
  open: [id: number]
  click: [file: LibraryFile]
  delete: [id: number]
  updated: [id: number]
}>()

const columns = ref(3)

const normalizeCollectionName = (collectionName?: string) => {
  return collectionName?.trim() || UNSORTED_COLLECTION_KEY
}

const groupedFiles = computed<FileGroup[]>(() => {
  const groups = new Map<string, FileGroup>()

  props.files.forEach((file) => {
    const key = normalizeCollectionName(file.collectionName)
    const existing = groups.get(key) || {
      key,
      label: key === UNSORTED_COLLECTION_KEY ? UNSORTED_COLLECTION_LABEL : key,
      files: [],
    }

    existing.files.push(file)
    groups.set(key, existing)
  })

  return Array.from(groups.values()).sort((left, right) => {
    if (left.key === UNSORTED_COLLECTION_KEY) {
      return 1
    }

    if (right.key === UNSORTED_COLLECTION_KEY) {
      return -1
    }

    if (right.files.length !== left.files.length) {
      return right.files.length - left.files.length
    }

    return left.label.localeCompare(right.label, 'zh-CN')
  })
})

const updateColumns = () => {
  const width = window.innerWidth

  if (width < 600) {
    columns.value = 2
  } else if (width < 900) {
    columns.value = 3
  } else if (width < 1200) {
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

const handleOpen = (id: number) => {
  emit('open', id)
}

const handleClick = (file: LibraryFile) => {
  emit('click', file)
}

const handleDelete = (id: number) => {
  emit('delete', id)
}

const handleUpdated = (id: number) => {
  emit('updated', id)
}
</script>

<style scoped>
.file-list-container {
  width: 100%;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: var(--text-color);
  opacity: 0.6;
  height: 100%;
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 16px;
}

.empty-state h3 {
  font-size: 20px;
  margin: 0 0 8px 0;
}

.empty-state p {
  font-size: 14px;
  margin: 0;
}

.grouped-file-list {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 20px;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 0 8px 8px;
}

.file-group {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.file-group-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 18px;
  border-radius: 18px;
  background:
    radial-gradient(circle at top right, rgba(33, 150, 243, 0.12), transparent 36%),
    color-mix(in srgb, var(--surface-color) 90%, transparent);
  border: 1px solid rgba(148, 163, 184, 0.18);
}

.file-group-title-wrap {
  min-width: 0;
}

.file-group-title {
  margin: 0 0 4px;
  font-size: 18px;
  color: var(--text-color);
  word-break: break-word;
}

.file-group-subtitle {
  margin: 0;
  font-size: 13px;
  color: var(--text-color);
  opacity: 0.68;
}

.file-list {
  display: grid;
  gap: 16px;
  padding: 0 8px;
  overflow-y: auto;
  overflow-x: hidden;
  height: 100%;
  width: 100%;
  min-width: 0;
  align-content: start;
  align-items: stretch;
  grid-auto-rows: var(--file-card-height, 292px);
}

.grouped-file-list .file-list {
  height: auto;
  overflow: visible;
  padding: 0;
}

.file-list > * {
  min-width: 0;
}

.file-list.columns-2 {
  gap: 12px;
}

.file-list.columns-5 {
  gap: 20px;
}

.grouped-file-list::-webkit-scrollbar,
.file-list::-webkit-scrollbar {
  width: 8px;
}

.grouped-file-list::-webkit-scrollbar-track,
.file-list::-webkit-scrollbar-track {
  background: var(--surface-color);
}

.grouped-file-list::-webkit-scrollbar-thumb,
.file-list::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 4px;
}

.grouped-file-list::-webkit-scrollbar-thumb:hover,
.file-list::-webkit-scrollbar-thumb:hover {
  background: var(--text-color);
}

@media (max-width: 600px) {
  .file-list {
    padding: 0 4px;
    gap: 12px;
    --file-card-height: 260px;
  }

  .grouped-file-list {
    gap: 16px;
    padding: 0 4px 8px;
  }

  .grouped-file-list .file-list {
    padding: 0;
  }

  .file-group-header {
    padding: 12px 14px;
    border-radius: 16px;
  }

  .empty-state {
    padding: 40px 20px;
  }
}

@media (max-width: 480px) {
  .file-list {
    padding: 0 2px;
    gap: 8px;
    --file-card-height: 248px;
  }

  .grouped-file-list {
    padding: 0 2px 8px;
  }

  .file-group-title {
    font-size: 16px;
  }
}
</style>

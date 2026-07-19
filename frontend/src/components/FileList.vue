<template>
  <div class="file-list-container">
    <div v-if="files.length === 0" class="empty-state">
      <div class="empty-icon">📁</div>
      <h3>暂无文件</h3>
      <p>点击上方"导入文件"按钮开始添加文件</p>
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
import { ref, onMounted, onUnmounted } from 'vue'
import FileCard from './FileCard.vue'

interface File {
  id: number
  fileName: string
  collectionName?: string
  filePath: string
  fileType: string
  fileSubType?: string
  fileSize: number
  tags: string[]
  description?: string
  thumbnail?: string
  createdAt: string
  modifiedAt: string
}

const props = defineProps<{
  files: File[]
}>()

const emit = defineEmits<{
  open: [id: number]
  click: [file: File]
  delete: [id: number]
  updated: [id: number]
}>()

const columns = ref(3)

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

const handleClick = (file: File) => {
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

.file-list > * {
  min-width: 0;
}

.file-list.columns-2 {
  gap: 12px;
}

.file-list.columns-5 {
  gap: 20px;
}

/* Scrollbar styling for the file list */
.file-list::-webkit-scrollbar {
  width: 8px;
}

.file-list::-webkit-scrollbar-track {
  background: var(--surface-color);
}

.file-list::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 4px;
}

.file-list::-webkit-scrollbar-thumb:hover {
  background: var(--text-color);
}

@media (max-width: 600px) {
  .file-list {
    padding: 0 4px;
    gap: 12px;
    --file-card-height: 260px;
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
}
</style>

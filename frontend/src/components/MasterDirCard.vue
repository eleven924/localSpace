<template>
  <div class="master-dir-card" :class="{ inactive: !master.isActive }">
    <div class="card-header" @click="toggleExpanded">
      <div class="header-left">
        <div class="expand-icon" :class="{ expanded: master.expanded }">
          <span v-if="master.expanded">▼</span>
          <span v-else>▶</span>
        </div>
        <div class="master-icon">📁</div>
        <div class="master-info">
          <h5 class="master-path">{{ master.path }}</h5>
          <div class="master-meta">
            <span v-if="master.isDefault" class="default-badge">默认</span>
            <span class="size-info">
              {{ formatSize(master.totalSize) }} / {{ formatSize(master.maxSize || 0) }}
            </span>
            <span class="subdir-count">{{ master.subDirs.length }} 个子目录</span>
          </div>
        </div>
      </div>
      <div class="header-actions">
        <button
          v-if="!master.isDefault"
          class="action-button"
          title="设为默认"
          @click.stop="handleSetDefault"
        >
          ⭐
        </button>
        <button
          class="action-button delete"
          title="删除"
          @click.stop="handleDelete"
        >
          🗑️
        </button>
      </div>
    </div>

    <div v-if="master.expanded" class="subdirs-panel">
      <div v-if="master.subDirs.length === 0" class="empty-subdirs">
        <span class="empty-icon">📂</span>
        <span class="empty-text">暂无子目录</span>
      </div>
      <div v-else class="subdirs-list">
        <div
          v-for="sub in master.subDirs"
          :key="sub.id"
          class="subdir-item"
        >
          <div class="subdir-icon">{{ getSubDirIcon(sub.fileType) }}</div>
          <div class="subdir-info">
            <span class="subdir-type">{{ getSubDirLabel(sub.fileType) }}</span>
            <span class="subdir-size">{{ formatSize(sub.currentSize) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { FILE_TYPES } from '@/utils/constants'
import { formatFileSize } from '@/utils/constants'

interface StorageDir {
  id: number
  path: string
  fileType: string
  currentSize: number
  maxSize?: number
  isActive: boolean
  isDefault: boolean
  subDirs?: StorageDir[]
  totalSize?: number
  expanded: boolean
}

const props = defineProps<{
  master: StorageDir
}>()

const emit = defineEmits<{
  'set-default': [id: number]
  'delete': [id: number]
  'toggle-expanded': [id: number]
}>()

const formatSize = (bytes: number): string => {
  if (!bytes || bytes === 0) return '未限制'
  return formatFileSize(bytes)
}

const getSubDirIcon = (fileType: string): string => {
  const type = FILE_TYPES.find(t => t.value === fileType)
  return type?.icon || '📁'
}

const getSubDirLabel = (fileType: string): string => {
  const type = FILE_TYPES.find(t => t.value === fileType)
  return type?.label || fileType
}

const toggleExpanded = () => {
  emit('toggle-expanded', props.master.id)
}

const handleSetDefault = () => {
  emit('set-default', props.master.id)
}

const handleDelete = () => {
  emit('delete', props.master.id)
}
</script>

<style scoped>
.master-dir-card {
  background-color: var(--surface-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
  transition: all 0.2s ease;
}

.master-dir-card:hover {
  border-color: var(--primary-color);
  box-shadow: 0 2px 8px var(--shadow-color);
}

.master-dir-card.inactive {
  opacity: 0.6;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  cursor: pointer;
  user-select: none;
}

.card-header:hover {
  background-color: var(--bg-color);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  min-width: 0;
}

.expand-icon {
  width: 20px;
  text-align: center;
  font-size: 12px;
  color: var(--text-color);
  opacity: 0.7;
  transition: transform 0.2s ease;
}

.expand-icon.expanded {
  transform: rotate(90deg);
}

.master-icon {
  font-size: 20px;
  flex-shrink: 0;
}

.master-info {
  flex: 1;
  min-width: 0;
}

.master-path {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color);
  margin: 0 0 4px 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.master-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-color);
  opacity: 0.7;
}

.default-badge {
  background-color: var(--primary-color);
  color: white;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.size-info {
  font-weight: 500;
}

.subdir-count {
  opacity: 0.8;
}

.header-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.action-button {
  width: 30px;
  height: 30px;
  padding: 0;
  border-radius: 50%;
  background-color: var(--bg-color);
  border: 1px solid var(--border-color);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  transition: all 0.2s ease;
}

.action-button:hover {
  background-color: var(--border-color);
  transform: scale(1.1);
}

.action-button.delete:hover {
  background-color: var(--error-color);
  border-color: var(--error-color);
}

.subdirs-panel {
  border-top: 1px solid var(--border-color);
  padding: 10px 14px;
  background-color: var(--bg-color);
}

.empty-subdirs {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px;
  color: var(--text-color);
  opacity: 0.6;
}

.empty-icon {
  font-size: 20px;
}

.empty-text {
  font-size: 14px;
}

.subdirs-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.subdir-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 8px;
  background-color: var(--surface-color);
  border-radius: 6px;
}

.subdir-icon {
  font-size: 16px;
}

.subdir-info {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 12px;
}

.subdir-type {
  color: var(--text-color);
  font-weight: 500;
}

.subdir-size {
  color: var(--text-color);
  opacity: 0.7;
}
</style>

<template>
  <div class="master-dir-card" :class="{ inactive: !master.isActive }">
    <div class="card-header" @click="toggleExpanded">
      <div class="header-left">
        <div class="expand-icon" :class="{ expanded: master.expanded }" aria-hidden="true">
          <span>›</span>
        </div>
        <div class="master-icon" aria-hidden="true">目录</div>
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
          设为默认
        </button>
        <button
          class="action-button delete"
          title="删除"
          @click.stop="handleDelete"
        >
          删除
        </button>
      </div>
    </div>

    <div v-if="master.expanded" class="subdirs-panel">
      <div v-if="master.subDirs.length === 0" class="empty-subdirs">
        <span class="empty-text">暂无子目录</span>
      </div>
      <div v-else class="subdirs-list">
        <div
          v-for="sub in master.subDirs"
          :key="sub.id"
          class="subdir-item"
        >
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
  background-color: transparent;
  border: none;
  border-bottom: 1px solid var(--border-color);
  border-radius: 0;
  overflow: hidden;
  transition: all 0.2s ease;
}

.master-dir-card:hover {
  border-color: var(--border-color);
  box-shadow: none;
}

.master-dir-card.inactive {
  opacity: 0.6;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 0;
  cursor: pointer;
  user-select: none;
}

.card-header:hover {
  background-color: transparent;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  min-width: 0;
}

.expand-icon {
  width: 18px;
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
  padding: 2px 6px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  color: var(--text-faint);
  font-size: 11px;
  font-weight: 700;
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
  color: var(--text-soft);
  opacity: 1;
}

.default-badge {
  background-color: color-mix(in srgb, var(--primary-color) 10%, transparent);
  color: var(--primary-color);
  padding: 2px 7px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 700;
}

.size-info {
  font-weight: 500;
}

.subdir-count {
  opacity: 0.8;
}

.header-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.action-button {
  width: auto;
  min-height: 30px;
  padding: 5px 9px;
  border-radius: 6px;
  background-color: var(--surface-color);
  border: 1px solid var(--border-color);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-soft);
  font-size: 12px;
  font-weight: 600;
  transition: all 0.2s ease;
}

.action-button:hover {
  background-color: var(--surface-muted);
  color: var(--text-color);
}

.action-button.delete:hover {
  background-color: color-mix(in srgb, var(--error-color) 8%, transparent);
  border-color: var(--error-color);
  color: var(--error-color);
}

.subdirs-panel {
  border-top: 1px solid var(--border-color);
  padding: 8px 0 12px 36px;
  background-color: transparent;
}

.empty-subdirs {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 8px;
  padding: 10px;
  color: var(--text-color);
  opacity: 0.6;
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
  padding: 7px 0;
  background-color: transparent;
  border-radius: 6px;
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

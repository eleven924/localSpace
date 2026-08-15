<template>
  <article class="set-dir" :class="{ open: master.expanded, inactive: !master.isActive }">
    <div class="set-dir-head">
      <button
        type="button"
        class="set-dir-toggle"
        :aria-expanded="String(master.expanded)"
        @click="toggleExpanded"
      >
        <i class="set-dir-chevron" aria-hidden="true">›</i>
        <span class="set-dir-copy">
          <span class="set-dir-path">
            <strong>{{ master.path }}</strong>
            <span v-if="master.isDefault" class="set-badge">默认</span>
          </span>
          <span class="set-dir-meta">
            <span><b>{{ usedSize }}</b> / {{ quota }}</span>
            <span><b>{{ subDirs.length }}</b> 个子目录</span>
          </span>
        </span>
      </button>

      <div class="set-dir-actions">
        <button
          v-if="!master.isDefault"
          type="button"
          class="set-mini"
          title="设为默认"
          @click="handleSetDefault"
        >
          设为默认
        </button>
        <button type="button" class="set-mini warn" title="删除主目录" @click="handleDelete">
          删除
        </button>
      </div>
    </div>

    <div v-if="master.expanded" class="set-subdirs">
      <p v-if="subDirs.length === 0" class="set-subdir-empty">
        暂无子目录，导入文件后会按类型自动创建。
      </p>
      <div v-for="sub in subDirs" v-else :key="sub.id" class="set-subdir">
        <span class="set-type sm" :class="sub.fileType" aria-hidden="true">
          {{ shortLabel(sub.fileType) }}
        </span>
        <span class="set-subdir-name">{{ typeLabel(sub.fileType) }}</span>
        <b>{{ formatFileSize(sub.currentSize) }}</b>
      </div>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { FILE_TYPE_META, formatFileSize } from '@/utils/constants'

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

const subDirs = computed(() => props.master.subDirs ?? [])

const usedSize = computed(() => formatFileSize(props.master.totalSize || 0))

// 已用和上限是两件事：上限为 0 是「未限制」，已用为 0 只是还没放东西。
const quota = computed(() =>
  props.master.maxSize ? formatFileSize(props.master.maxSize) : '未限制',
)

const typeLabel = (fileType: string): string => FILE_TYPE_META[fileType]?.label || fileType

const shortLabel = (fileType: string): string => FILE_TYPE_META[fileType]?.shortLabel || '其'

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
.set-dir.inactive {
  opacity: 0.58;
}
</style>

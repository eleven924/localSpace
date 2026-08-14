<template>
  <div class="page-shell collections-view" @click="showFilterPopover = false">
    <AppHeader />

    <div class="page-content">
      <div class="page-stack">
        <section class="collections-toolbar" @click.stop>
          <div class="collection-tools-row">
            <div class="search-bar collections-search">
              <div class="search-input-wrapper">
                <span class="search-icon">⌕</span>
                <input
                  v-model="collectionQuery"
                  class="search-input"
                  type="text"
                  placeholder="搜索合集名称"
                />
              </div>
            </div>

            <div class="collection-filter-control">
              <button
                type="button"
                class="filter-trigger"
                :class="{ active: showFilterPopover || filesStore.currentFileType !== 'all' }"
                @click="showFilterPopover = !showFilterPopover"
              >
                <span class="filter-trigger-dot"></span>
                <span>筛选</span>
              </button>

              <div v-if="showFilterPopover" class="filter-popover" @click.stop>
                <div class="popover-head">
                  <strong>缩小范围</strong>
                  <button type="button" class="text-button" @click="clearTypeFilter">清除全部</button>
                </div>

                <div class="filter-group">
                  <span class="filter-group-label">文件类型</span>
                  <FileTypeFilter
                    v-model="filesStore.currentFileType"
                    :counts="filesStore.fileTypeCounts"
                    @filter="handleTypeFilter"
                  />
                </div>
              </div>
            </div>

          </div>
        </section>

        <div v-if="filesStore.loading" class="state-panel">
          <div class="spinner"></div>
          <p>正在加载合集...</p>
        </div>

        <div v-else-if="filesStore.error" class="state-panel error-state">
          <h3>加载失败</h3>
          <p>{{ filesStore.error }}</p>
          <button class="btn primary" @click="handleRetry">重试</button>
        </div>

        <div v-else-if="filteredCollections.length === 0" class="collection-empty-state">
          <EmptyState
            title="还没有可见合集"
            description="你可以先导入文件并填写合集名称，或者调整当前筛选条件。"
          />
        </div>

        <section v-else class="collections-grid">
          <CollectionCard
            v-for="summary in filteredCollections"
            :key="summary.value"
            :summary="summary"
            @open="openCollection"
          />
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppHeader from '@/components/AppHeader.vue'
import CollectionCard from '@/components/CollectionCard.vue'
import EmptyState from '@/components/EmptyState.vue'
import FileTypeFilter from '@/components/FileTypeFilter.vue'
import { useFilesStore } from '@/store/modules/files'
import { api, isWailsAvailable } from '@/api'
import type { Collection } from '@/types'
import { UNSORTED_COLLECTION_KEY } from '@/utils/constants'

const filesStore = useFilesStore()
const router = useRouter()
const route = useRoute()
const collectionQuery = ref('')
const collections = ref<Collection[]>([])
const showFilterPopover = ref(false)
let isUnmounted = false

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

const getRouteType = () => {
  return typeof route.query.type === 'string' ? route.query.type : 'all'
}

const waitForWails = async (timeoutMs = 5000, intervalMs = 100) => {
  const deadline = Date.now() + timeoutMs
  while (!isWailsAvailable() && Date.now() < deadline) {
    if (isUnmounted) return false
    await sleep(intervalMs)
  }
  return isWailsAvailable()
}

const ensureLibraryLoaded = async () => {
  const wailsReady = await waitForWails()
  if (!wailsReady || isUnmounted) return
  filesStore.searchQuery = ''
  await filesStore.loadFiles(filesStore.currentFileType)
}

const loadCollections = async () => {
  try {
    collections.value = await api.collection.getAll()
  } catch (err) {
    console.error('Failed to load collections:', err)
  }
}

const filteredCollections = computed(() => {
  const keyword = collectionQuery.value.trim().toLowerCase()
  if (!keyword) {
    return filesStore.collectionSummaries
  }

  return filesStore.collectionSummaries.filter((summary) => {
    return summary.label.toLowerCase().includes(keyword)
  })
})

onMounted(() => {
  filesStore.setCurrentFileType(getRouteType())
  void loadCollections()
  void ensureLibraryLoaded()
})

onUnmounted(() => {
  isUnmounted = true
})

const handleTypeFilter = async (fileType: string) => {
  showFilterPopover.value = false
  filesStore.setCurrentFileType(fileType)
  await filesStore.loadFiles(fileType)

  const query = fileType === 'all' ? {} : { type: fileType }
  await router.replace({ name: 'Collections', query })
}

const clearTypeFilter = async () => {
  // 合集页只清除它实际支持的类型筛选，合集搜索仍由输入值直接控制。
  await handleTypeFilter('all')
}

const openCollection = (collectionName: string) => {
  const normalizedName = collectionName.trim()
  const query: Record<string, string> = {
    collection: normalizedName,
    view: 'grouped',
  }

  // 文件页筛选以 collectionId 为准；合集页入口优先传 id，旧链接仍由文件页按名称兜底。
  if (normalizedName === UNSORTED_COLLECTION_KEY) {
    query.collection = 'unsorted'
  } else {
    const matchedCollection = collections.value.find((collection) => collection.name === normalizedName)
    if (matchedCollection) {
      query.collection = String(matchedCollection.id)
    }
  }

  if (filesStore.currentFileType !== 'all') {
    query.type = filesStore.currentFileType
  }

  void router.push({ name: 'Files', query })
}

const handleRetry = async () => {
  const wailsReady = await waitForWails()
  if (!wailsReady || isUnmounted) return
  await filesStore.loadFiles(filesStore.currentFileType)
}
</script>

<style scoped>
.collections-toolbar {
  position: relative;
  display: flex;
  align-items: center;
  flex-shrink: 0;
  /* 与资料库工具条使用相同的底部节奏，确保下方空状态拥有一致的可用高度。 */
  padding: 0 0 18px;
}

.collections-view .page-stack {
  /* 与资料库保持相同的剩余空间计算，让空状态以内容区为基准居中。 */
  height: 100%;
  min-height: 0;
}

.collection-tools-row {
  display: flex;
  align-items: center;
  width: 100%;
  gap: 10px;
}

.collections-search {
  flex: 1;
  min-width: 200px;
}

.collection-filter-control {
  position: static;
}

.filter-trigger {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  height: 43px;
  padding: 0 13px;
  border: 1px solid var(--border-color);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.65);
  color: var(--text-soft);
  font-size: 12px;
  font-weight: 650;
}

.filter-trigger:hover,
.filter-trigger.active {
  border-color: rgba(111, 143, 216, 0.42);
  color: var(--primary-hover);
  background: rgba(235, 241, 250, 0.9);
}

.filter-trigger-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--primary-color);
}

/* 合集页沿用原型里的轻量搜索字段，输入值仍直接驱动已有本地筛选逻辑。 */
.collections-search .search-input-wrapper {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 43px;
  padding: 0 13px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.74);
  box-shadow: 0 5px 13px rgba(40, 60, 90, 0.03);
}

.collections-search .search-icon {
  position: static;
  flex-shrink: 0;
  transform: none;
}

.collections-search .search-input {
  flex: 1;
  min-width: 0;
  min-height: 0;
  padding: 0;
  border: 0;
  outline: 0;
  background: transparent;
  box-shadow: none;
}

.collections-view .filter-popover {
  position: absolute;
  z-index: 20;
  top: calc(100% + 9px);
  right: 72px;
  width: min(430px, calc(100vw - 50px));
  padding: 16px;
  border: 1px solid rgba(146, 165, 192, 0.36);
  border-radius: 15px;
  background: rgba(255, 255, 255, 0.97);
  box-shadow: 0 18px 38px rgba(29, 48, 78, 0.14);
}

.collections-view .filter-popover::before {
  content: '';
  position: absolute;
  top: -6px;
  right: 119px;
  width: 11px;
  height: 11px;
  border-top: 1px solid rgba(146, 165, 192, 0.36);
  border-left: 1px solid rgba(146, 165, 192, 0.36);
  background: #fff;
  transform: rotate(45deg);
}

.collections-view .popover-head,
.collections-view .filter-group {
  position: relative;
}

.collections-view .popover-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 13px;
  border-bottom: 1px solid rgba(146, 165, 192, 0.24);
}

.collections-view .popover-head strong {
  color: var(--text-color);
  font-size: 13px;
}

.collections-view .text-button {
  padding: 2px 0;
  color: var(--primary-color);
  background: transparent;
  font-size: 11px;
}

.collections-view .filter-group {
  padding-top: 14px;
}

.collections-view .filter-group-label {
  display: block;
  margin-bottom: 9px;
  color: var(--text-faint);
  font-size: 10px;
  font-weight: 750;
}

.collections-view .filter-popover :deep(.file-type-filter) {
  /* 与资料库保持同一排列节奏，筛选内容在两个页面之间不发生视觉跳变。 */
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.collections-view .filter-popover :deep(.file-type-filter button) {
  justify-content: center;
  width: auto;
  flex: 0 0 auto;
  min-height: 30px;
  padding: 6px 9px;
  border-radius: 8px;
  background: #fbfcfe;
  font-size: 11px;
  white-space: nowrap;
}

.collections-view .filter-popover :deep(.filter-icon),
.collections-view .filter-popover :deep(.filter-count) {
  display: none;
}

.collections-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(270px, 1fr));
  gap: 18px;
}

.collection-empty-state {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
}

.collection-empty-state :deep(.empty-state) {
  /* EmptyState 是共享组件，这里只覆盖合集页的布局，不改变其他页面事实。 */
  flex: 1;
  min-height: 0;
}

.state-panel {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  min-height: 220px;
  border-radius: 28px;
  border: 1px solid rgba(146, 165, 192, 0.18);
  background: rgba(255, 255, 255, 0.64);
  color: var(--text-soft);
}

@media (max-width: 640px) {
  .collections-toolbar {
    padding-bottom: 0;
  }

  .collection-tools-row {
    flex-wrap: wrap;
  }

  .collections-search {
    flex-basis: 100%;
  }

  .collections-view .filter-popover {
    right: 0;
    width: min(430px, calc(100vw - 24px));
  }
}
</style>


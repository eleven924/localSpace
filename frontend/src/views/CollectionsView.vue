<template>
  <div class="page-shell collections-view">
    <AppHeader />

    <div class="page-content">
      <div class="page-stack">
        <section class="toolbar-panel collections-toolbar">
          <FileTypeFilter
            v-model="filesStore.currentFileType"
            :counts="filesStore.fileTypeCounts"
            @filter="handleTypeFilter"
          />

          <div class="collection-search">
            <span class="collection-search-icon">⌕</span>
            <input v-model="collectionQuery" type="text" placeholder="搜索合集名称" />
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

        <div v-else-if="filteredCollections.length === 0" class="section-panel empty-panel">
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
import { isWailsAvailable } from '@/api'

const filesStore = useFilesStore()
const router = useRouter()
const route = useRoute()
const collectionQuery = ref('')
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
  void ensureLibraryLoaded()
})

onUnmounted(() => {
  isUnmounted = true
})

const handleTypeFilter = async (fileType: string) => {
  filesStore.setCurrentFileType(fileType)
  await filesStore.loadFiles(fileType)

  const query = fileType === 'all' ? {} : { type: fileType }
  await router.replace({ name: 'Collections', query })
}

const openCollection = (collectionName: string) => {
  const query: Record<string, string> = {
    collection: collectionName,
    view: 'grouped',
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
  display: grid;
  gap: 14px;
  padding: 18px;
}

.collection-search {
  position: relative;
  width: min(420px, 100%);
}

.collection-search-icon {
  position: absolute;
  left: 14px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-faint);
}

.collection-search input {
  min-height: 44px;
  padding-left: 38px;
  border-radius: 999px;
}

.collections-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(270px, 1fr));
  gap: 18px;
}

.empty-panel {
  padding: 10px;
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
    padding: 14px;
  }
}
</style>

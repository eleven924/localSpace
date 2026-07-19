<template>
  <div class="collections-view">
    <header class="header">
      <div class="header-left">
        <div class="brand-block">
          <h1>LocalSpace</h1>
          <p class="subtitle">从“系列 / 剧集 / 课程 / 项目”角度浏览你的内容库</p>
        </div>
        <nav class="header-nav">
          <router-link to="/files" class="nav-pill">文件</router-link>
          <router-link to="/collections" class="nav-pill active">合集</router-link>
        </nav>
      </div>

      <div class="header-actions">
        <router-link to="/import" class="btn primary">导入文件</router-link>
        <router-link to="/settings" class="btn secondary">设置</router-link>
      </div>
    </header>

    <div class="content">
      <section class="hero-panel">
        <div class="hero-copy">
          <h2>合集视图</h2>
          <p>适合视频剧集、课程资料、系列资源这类“需要连续浏览”的内容管理方式。</p>
        </div>

        <div class="hero-stats">
          <div class="stat-card">
            <span class="stat-value">{{ visibleCollectionCount }}</span>
            <span class="stat-label">可见合集</span>
          </div>
          <div class="stat-card">
            <span class="stat-value">{{ visibleFileCount }}</span>
            <span class="stat-label">覆盖文件</span>
          </div>
        </div>
      </section>

      <section class="toolbar-panel">
        <div class="toolbar-top">
          <FileTypeFilter
            v-model="filesStore.currentFileType"
            :counts="filesStore.fileTypeCounts"
            @filter="handleTypeFilter"
          />
        </div>

        <div class="toolbar-bottom">
          <div class="collection-search">
            <span class="collection-search-icon">⌕</span>
            <input
              v-model="collectionQuery"
              type="text"
              placeholder="搜索合集名称"
            />
          </div>
        </div>
      </section>

      <div v-if="filesStore.loading" class="loading-state">
        <div class="spinner"></div>
        <p>加载合集中...</p>
      </div>

      <div v-else-if="filesStore.error" class="error-state">
        <div class="error-icon">⚠️</div>
        <h3>加载失败</h3>
        <p>{{ filesStore.error }}</p>
        <button class="btn primary" @click="handleRetry">重试</button>
      </div>

      <div v-else-if="filteredCollections.length === 0" class="empty-wrap">
        <EmptyState
          title="暂无可见合集"
          description="可以先导入文件并填写合集名，或者调整当前类型筛选与搜索条件。"
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
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import CollectionCard from '@/components/CollectionCard.vue'
import EmptyState from '@/components/EmptyState.vue'
import FileTypeFilter from '@/components/FileTypeFilter.vue'
import { useFilesStore } from '@/store/modules/files'

const filesStore = useFilesStore()
const router = useRouter()
const route = useRoute()
const collectionQuery = ref('')

const getRouteType = () => {
  return typeof route.query.type === 'string' ? route.query.type : 'all'
}

const ensureLibraryLoaded = async () => {
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

const visibleCollectionCount = computed(() => filteredCollections.value.length)
const visibleFileCount = computed(() => {
  return filteredCollections.value.reduce((sum, summary) => sum + summary.count, 0)
})

onMounted(() => {
  filesStore.setCurrentFileType(getRouteType())
  void ensureLibraryLoaded()
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
  await filesStore.loadFiles(filesStore.currentFileType)
}
</script>

<style scoped>
.collections-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background:
    radial-gradient(circle at top left, rgba(33, 150, 243, 0.08), transparent 22%),
    var(--app-bg-color, var(--bg-color));
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
  padding: 16px 24px;
  background-color: color-mix(in srgb, var(--surface-color) 92%, transparent);
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
  min-width: 0;
}

.brand-block h1 {
  margin: 0 0 4px;
  font-size: 24px;
  font-weight: 700;
  color: var(--text-color);
}

.subtitle {
  margin: 0;
  color: var(--text-color);
  opacity: 0.7;
  font-size: 13px;
}

.header-nav {
  display: inline-flex;
  gap: 8px;
  padding: 6px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--surface-color) 82%, transparent);
  border: 1px solid rgba(148, 163, 184, 0.16);
}

.nav-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 72px;
  padding: 8px 14px;
  border-radius: 999px;
  color: var(--text-color);
  font-size: 13px;
  font-weight: 600;
}

.nav-pill.active,
.nav-pill.router-link-active {
  background: linear-gradient(135deg, var(--primary-color) 0%, color-mix(in srgb, var(--primary-color) 78%, #0f172a) 100%);
  color: #fff;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 0;
  padding: 20px 24px 24px;
}

.hero-panel {
  display: flex;
  align-items: stretch;
  justify-content: space-between;
  gap: 20px;
  padding: 22px 24px;
  border-radius: 24px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background:
    radial-gradient(circle at top right, rgba(33, 150, 243, 0.14), transparent 28%),
    linear-gradient(180deg, color-mix(in srgb, var(--surface-color) 90%, white 10%) 0%, var(--surface-color) 100%);
}

.hero-copy h2 {
  margin: 0 0 8px;
  font-size: 24px;
  color: var(--text-color);
}

.hero-copy p {
  margin: 0;
  max-width: 680px;
  color: var(--text-color);
  opacity: 0.74;
  line-height: 1.7;
}

.hero-stats {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  min-width: 240px;
}

.stat-card {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 6px;
  padding: 16px 18px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.16);
  border: 1px solid rgba(148, 163, 184, 0.16);
}

.stat-value {
  font-size: 28px;
  font-weight: 800;
  color: var(--text-color);
}

.stat-label {
  font-size: 13px;
  color: var(--text-color);
  opacity: 0.7;
}

.toolbar-panel {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 18px 20px;
  border-radius: 22px;
  background-color: color-mix(in srgb, var(--surface-color) 90%, transparent);
  border: 1px solid rgba(148, 163, 184, 0.16);
}

.toolbar-bottom {
  display: flex;
  justify-content: flex-start;
}

.collection-search {
  position: relative;
  width: min(420px, 100%);
}

.collection-search-icon {
  position: absolute;
  top: 50%;
  left: 14px;
  transform: translateY(-50%);
  color: var(--text-color);
  opacity: 0.45;
}

.collection-search input {
  width: 100%;
  padding: 12px 16px 12px 38px;
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.18);
}

.loading-state,
.error-state,
.empty-wrap {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.error-state {
  flex-direction: column;
  gap: 14px;
  text-align: center;
}

.error-icon {
  font-size: 48px;
}

.collections-grid {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 18px;
  padding: 4px;
}

@media (max-width: 900px) {
  .header {
    flex-direction: column;
    align-items: flex-start;
  }

  .header-left {
    width: 100%;
    flex-direction: column;
    align-items: flex-start;
  }

  .header-actions {
    width: 100%;
    justify-content: flex-end;
  }

  .hero-panel {
    flex-direction: column;
  }

  .hero-stats {
    min-width: 0;
  }
}

@media (max-width: 640px) {
  .content {
    padding: 16px;
  }

  .header {
    padding: 14px 16px;
  }

  .header-actions {
    width: 100%;
    flex-direction: column;
  }

  .header-actions .btn {
    width: 100%;
  }

  .hero-panel,
  .toolbar-panel {
    padding: 18px;
    border-radius: 18px;
  }

  .hero-stats {
    grid-template-columns: 1fr;
  }
}
</style>

<template>
  <div class="thumbnail-cache-settings">
    <h3>缩略图缓存管理</h3>

    <div class="cache-stats">
      <div class="stat-item">
        <div class="stat-label">缓存数量</div>
        <div class="stat-value">{{ stats.count || 0 }}</div>
      </div>
      <div class="stat-item">
        <div class="stat-label">当前大小</div>
        <div class="stat-value">{{ formatSize(stats.currentSize || 0) }}</div>
      </div>
      <div class="stat-item">
        <div class="stat-label">最大大小</div>
        <div class="stat-value">{{ formatSize(stats.maxSize || 0) }}</div>
      </div>
      <div class="stat-item">
        <div class="stat-label">使用率</div>
        <div class="stat-value">{{ stats.usagePercent?.toFixed(1) || 0 }}%</div>
      </div>
    </div>

    <div class="cache-info">
      <div class="info-item">
        <span class="info-label">默认尺寸:</span>
        <span class="info-value">{{ stats.defaultWidth }}×{{ stats.defaultHeight }}</span>
      </div>
      <div class="info-item">
        <span class="info-label">最大保留时间:</span>
        <span class="info-value">{{ stats.maxAge || '7天' }}</span>
      </div>
    </div>

    <div class="cache-actions">
      <button
        class="btn btn-danger"
        @click="handleClearCache"
        :disabled="stats.count === 0"
      >
        清空缓存
      </button>
      <button class="btn btn-secondary" @click="handleRefresh">
        刷新信息
      </button>
    </div>

    <div class="cache-usage">
      <div class="usage-bar">
        <div
          class="usage-fill"
          :style="{ width: `${stats.usagePercent || 0}%` }"
          :class="{ 'usage-high': stats.usagePercent > 80 }"
        ></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'

interface CacheStats {
  count: number
  currentSize: number
  maxSize: number
  usagePercent: number
  defaultWidth: number
  defaultHeight: number
  maxAge: string
}

const stats = ref<CacheStats>({
  count: 0,
  currentSize: 0,
  maxSize: 0,
  usagePercent: 0,
  defaultWidth: 300,
  defaultHeight: 300,
  maxAge: '7天',
})

const formatSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
}

const loadCacheStats = async () => {
  try {
    const cacheStats = await api.thumbnail.getCacheStats()
    stats.value = {
      count: cacheStats.count || 0,
      currentSize: cacheStats.currentSize || 0,
      maxSize: cacheStats.maxSize || 0,
      usagePercent: cacheStats.usagePercent || 0,
      defaultWidth: cacheStats.defaultWidth || 300,
      defaultHeight: cacheStats.defaultHeight || 300,
      maxAge: cacheStats.maxAge || '7天',
    }
  } catch (err) {
    console.error('Failed to load cache stats:', err)
  }
}

const handleClearCache = async () => {
  if (confirm('确定要清空所有缩略图缓存吗？此操作不可撤销。')) {
    try {
      await api.thumbnail.clearCache()
      await loadCacheStats()
      alert('缓存已清空')
    } catch (err) {
      console.error('Failed to clear cache:', err)
      alert('清空缓存失败')
    }
  }
}

const handleRefresh = async () => {
  await loadCacheStats()
}

onMounted(() => {
  loadCacheStats()
})
</script>

<style scoped>
.thumbnail-cache-settings {
  background-color: var(--surface-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 20px;
}

.thumbnail-cache-settings h3 {
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 500;
  color: var(--text-color);
}

.cache-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.stat-item {
  text-align: center;
  padding: 12px;
  background-color: var(--bg-color);
  border-radius: 6px;
}

.stat-label {
  font-size: 12px;
  color: var(--text-color);
  opacity: 0.7;
  margin-bottom: 4px;
}

.stat-value {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-color);
}

.cache-info {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 20px;
}

.info-item {
  display: flex;
  gap: 8px;
  font-size: 14px;
}

.info-label {
  color: var(--text-color);
  opacity: 0.7;
}

.info-value {
  color: var(--text-color);
  font-weight: 500;
}

.cache-actions {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.btn {
  padding: 8px 16px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background-color: var(--bg-color);
  color: var(--text-color);
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.btn:hover:not(:disabled) {
  background-color: var(--primary-color);
  color: white;
  border-color: var(--primary-color);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-danger:hover:not(:disabled) {
  background-color: #dc3545;
  border-color: #dc3545;
}

.btn-secondary:hover:not(:disabled) {
  background-color: var(--border-color);
}

.cache-usage {
  margin-top: 16px;
}

.usage-bar {
  height: 8px;
  background-color: var(--bg-color);
  border-radius: 4px;
  overflow: hidden;
}

.usage-fill {
  height: 100%;
  background-color: var(--primary-color);
  transition: width 0.3s ease;
}

.usage-high {
  background-color: #dc3545;
}

@media (max-width: 768px) {
  .cache-stats {
    grid-template-columns: repeat(2, 1fr);
  }

  .cache-actions {
    flex-direction: column;
  }

  .btn {
    width: 100%;
  }
}
</style>
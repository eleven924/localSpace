<template>
  <div class="thumbnail-preview" v-if="visible" @click="close">
    <div class="preview-container" @click.stop>
      <button class="close-button" @click="close">×</button>
      <div class="preview-content">
        <img
          v-if="thumbnailPath && !loading && !error"
          :src="`file://${thumbnailPath}`"
          :alt="fileName"
          class="preview-image"
          @error="handleError"
        />
        <div v-if="loading" class="preview-loading">
          <div class="spinner"></div>
          <p>加载缩略图...</p>
        </div>
        <div v-if="error" class="preview-error">
          <div class="error-icon">⚠️</div>
          <p>加载失败</p>
        </div>
      </div>
      <div class="preview-info" v-if="thumbnailPath && !loading && !error">
        <h3>{{ fileName }}</h3>
        <p>尺寸: {{ metadata.width || 'N/A' }} × {{ metadata.height || 'N/A' }}</p>
        <p>文件类型: {{ fileType }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { api } from '@/api'

interface Props {
  visible: boolean
  fileId?: number
  fileName?: string
  fileType?: string
  metadata?: {
    width?: number
    height?: number
    duration?: number
  }
}

const props = defineProps<Props>()

const emit = defineEmits<{
  close: []
}>()

const thumbnailPath = ref('')
const loading = ref(false)
const error = ref(false)

const close = () => {
  emit('close')
}

const handleError = () => {
  error.value = true
}

watch(() => props.visible, async (newVal) => {
  if (newVal && props.fileId) {
    await loadThumbnail()
  } else {
    thumbnailPath.value = ''
    loading.value = false
    error.value = false
  }
})

const loadThumbnail = async () => {
  if (!props.fileId) return

  loading.value = true
  error.value = false

  try {
    thumbnailPath.value = await api.thumbnail.generate(props.fileId)
  } catch (err) {
    console.error('Failed to load thumbnail:', err)
    error.value = true
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.thumbnail-preview {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.preview-container {
  background-color: var(--surface-color);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
  max-width: 90%;
  max-height: 90%;
  position: relative;
  overflow: hidden;
}

.close-button {
  position: absolute;
  top: 10px;
  right: 10px;
  width: 32px;
  height: 32px;
  border: none;
  background-color: rgba(0, 0, 0, 0.5);
  color: white;
  border-radius: 50%;
  cursor: pointer;
  font-size: 24px;
  line-height: 1;
  z-index: 10;
  transition: background-color 0.2s;
}

.close-button:hover {
  background-color: rgba(0, 0, 0, 0.7);
}

.preview-content {
  position: relative;
  background-color: var(--bg-color);
  min-width: 400px;
  min-height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-image {
  max-width: 100%;
  max-height: 60vh;
  object-fit: contain;
}

.preview-loading,
.preview-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  color: var(--text-color);
  padding: 40px;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.error-icon {
  font-size: 48px;
}

.preview-info {
  padding: 20px;
  border-top: 1px solid var(--border-color);
}

.preview-info h3 {
  margin: 0 0 8px 0;
  font-size: 16px;
  color: var(--text-color);
}

.preview-info p {
  margin: 4px 0;
  font-size: 14px;
  color: var(--text-color);
  opacity: 0.7;
}

@media (max-width: 768px) {
  .preview-container {
    max-width: 100%;
    max-height: 100%;
    border-radius: 0;
  }

  .preview-content {
    min-width: 300px;
    min-height: 300px;
  }
}
</style>
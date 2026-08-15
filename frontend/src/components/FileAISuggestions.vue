<template>
  <div v-if="aiEnabled" class="ai-block">
    <div class="ai-block-head">
      <div class="ai-block-title">
        <h4>AI 分析</h4>
        <p>生成标签和描述建议，应用前不会覆盖已填写的内容。</p>
      </div>
      <button class="btn secondary ai-trigger" type="button" :disabled="aiLoading" @click="handleAIAnalyze">
        {{ aiLoading ? '分析中...' : aiAnalysis ? '重新生成' : '生成' }}
      </button>
    </div>

    <div v-if="aiLoading" class="ai-block-loading">
      <span class="ai-spinner"></span>
      <p>AI 正在分析文件...</p>
    </div>

    <template v-else-if="aiAnalysis">
      <span class="ai-block-label">建议标签</span>
      <div class="ai-block-tags">
        <button
          v-for="tag in aiAnalysis.tags"
          :key="tag"
          class="ai-block-tag"
          type="button"
          @click="addTag(tag)"
        >
          {{ tag }}
        </button>
      </div>

      <span class="ai-block-label">建议描述</span>
      <p class="ai-block-desc">{{ aiAnalysis.description }}</p>

      <div class="ai-block-actions">
        <button class="btn secondary" type="button" @click="applyDescription">应用描述</button>
        <button class="btn secondary" type="button" @click="applyAll">应用全部</button>
      </div>
    </template>

    <p v-else class="ai-block-hint">生成后会在这里显示建议内容。</p>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { api } from '@/api'
import type { AIAnalysis } from '@/types'

const props = defineProps<{
  fileName: string
  fileType: string
  userKeywords?: string
  tags: string[]
  description: string
}>()

const emit = defineEmits<{
  'update:tags': [value: string[]]
  'update:description': [value: string]
  // 让父级知道这一块是否会渲染，导入页要据此决定单栏还是双栏。
  availability: [value: boolean]
}>()

const aiAnalysis = ref<AIAnalysis | null>(null)
const aiLoading = ref(false)
const aiEnabled = ref(false)

const checkAIEnabled = async () => {
  try {
    const config = await api.ai.getConfig()
    aiEnabled.value = config.enabled || false
  } catch (error) {
    console.error('Failed to check AI config:', error)
    aiEnabled.value = false
  }
  emit('availability', aiEnabled.value)
}

onMounted(checkAIEnabled)
watch(() => [props.fileName, props.fileType], checkAIEnabled)

const addTag = (tag: string) => {
  const normalized = tag.trim()
  if (!normalized || props.tags.some(item => item.trim() === normalized)) {
    return
  }
  emit('update:tags', [...props.tags, normalized])
}

const handleAIAnalyze = async () => {
  aiLoading.value = true
  try {
    // 使用当前表单内容作为上下文，让 AI 建议贴近用户已经填写的信息。
    aiAnalysis.value = await api.ai.analyze(
      props.fileName,
      props.fileType,
      props.userKeywords || '',
      props.tags,
      props.description
    )
  } catch (error) {
    console.error('AI analysis failed:', error)
  } finally {
    aiLoading.value = false
  }
}

const applyDescription = () => {
  if (aiAnalysis.value) {
    emit('update:description', aiAnalysis.value.description)
  }
}

const applyAll = () => {
  if (!aiAnalysis.value) return

  const merged = [...props.tags]
  aiAnalysis.value.tags.forEach(tag => {
    const normalized = tag.trim()
    if (normalized && !merged.some(item => item.trim() === normalized)) {
      merged.push(normalized)
    }
  })

  emit('update:tags', merged)
  emit('update:description', aiAnalysis.value.description)
}
</script>

<style scoped>
/* 这个组件只负责内容，外框由使用它的面板提供。 */
.ai-block-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.ai-block-title h4 {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-color);
  line-height: 1.3;
}

.ai-block-title p {
  margin-top: 5px;
  color: var(--import-text-faint, var(--text-faint));
  font-size: 12px;
  line-height: 1.55;
}

.ai-trigger {
  flex: none;
  min-height: 32px;
  padding: 6px 12px;
  font-size: 12px;
}

.ai-block-label {
  display: block;
  margin: 16px 0 9px;
  color: var(--import-text-faint, var(--text-faint));
  font-size: 12px;
  font-weight: 700;
}

.ai-block-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
}

.ai-block-tag {
  padding: 6px 10px;
  border-radius: 8px;
  border: 1px solid color-mix(in srgb, var(--border-color) 86%, transparent);
  background: color-mix(in srgb, var(--surface-color) 88%, transparent);
  color: var(--text-color);
  font-size: 12px;
  transition: all 0.2s ease;
}

.ai-block-tag:hover {
  border-color: color-mix(in srgb, var(--primary-color) 30%, transparent);
  background: color-mix(in srgb, var(--primary-color) 10%, transparent);
  color: var(--primary-color);
}

.ai-block-desc {
  padding: 11px 12px;
  border-radius: 10px;
  border: 1px solid color-mix(in srgb, var(--border-color) 84%, transparent);
  background: color-mix(in srgb, var(--surface-color) 92%, transparent);
  color: var(--import-text-soft, var(--text-soft));
  font-size: 13px;
  line-height: 1.7;
}

.ai-block-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 10px;
}

.ai-block-actions .btn {
  min-height: 32px;
  padding: 6px 12px;
  font-size: 12px;
}

.ai-block-loading {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 14px;
}

.ai-block-loading p {
  color: var(--import-text-soft, var(--text-soft));
  font-size: 13px;
}

.ai-spinner {
  width: 15px;
  height: 15px;
  flex: none;
  border: 2px solid color-mix(in srgb, var(--primary-color) 26%, transparent);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: ai-spin 0.9s linear infinite;
}

.ai-block-hint {
  margin-top: 14px;
  color: var(--import-text-faint, var(--text-faint));
  font-size: 12px;
  line-height: 1.7;
}

@keyframes ai-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (prefers-reduced-motion: reduce) {
  .ai-spinner {
    animation-duration: 2.4s;
  }
}
</style>

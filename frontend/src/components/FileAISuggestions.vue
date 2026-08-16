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

      <template v-if="aiAnalysis.relatedTags?.length">
        <span class="ai-block-label">可复用标签</span>
        <div class="ai-block-tags">
          <button
            v-for="tag in aiAnalysis.relatedTags"
            :key="`related-${tag}`"
            class="ai-block-tag ai-block-tag-related"
            type="button"
            @click="addTag(tag)"
          >
            {{ tag }}
          </button>
        </div>
      </template>

      <span class="ai-block-label">建议描述</span>
      <p class="ai-block-desc">{{ aiAnalysis.description }}</p>

      <p v-if="aiAnalysis.needsReview" class="ai-block-review">
        当前建议置信度较低（{{ Math.round((aiAnalysis.confidence || 0) * 100) }}%），请确认后再应用。
      </p>

      <details v-if="aiAnalysis.fallbackReason" class="ai-diagnostic-details">
        <summary>分析诊断</summary>
        <p class="ai-diagnostic-error">{{ aiAnalysis.fallbackReason }}</p>
        <p v-if="aiAnalysis.outputDiagnostic" class="ai-diagnostic-meta">{{ aiAnalysis.outputDiagnostic }}</p>
      </details>

      <div class="ai-block-actions">
        <button class="btn secondary" type="button" @click="applyDescription">应用描述</button>
        <button class="btn secondary" type="button" @click="applyAll">应用全部</button>
      </div>
      <details v-if="aiAnalysis.toolCalls?.length" class="ai-tool-details">
        <summary class="ai-tool-summary">
          <span class="ai-tool-summary-title">调用详情</span>
          <span class="ai-tool-summary-meta">{{ aiAnalysis.toolCalls.length }} 次 · {{ searchCallCount }} 次网络搜索</span>
        </summary>
        <div class="ai-tool-call-list">
          <article v-for="(call, index) in aiAnalysis.toolCalls" :key="`${call.name}-${index}`" class="ai-tool-call">
            <div class="ai-tool-call-head">
              <div>
                <span class="ai-tool-call-index">{{ String(index + 1).padStart(2, '0') }}</span>
                <strong>{{ toolLabel(call.name) }}</strong>
              </div>
              <span :class="['ai-tool-status', call.status === 'success' ? 'is-success' : 'is-failed']">
                {{ call.status === 'success' ? '成功' : '失败' }}
              </span>
            </div>
            <div class="ai-tool-call-meta">耗时 {{ call.durationMs }} ms</div>
            <div class="ai-tool-payload">
              <span>输入</span>
              <pre>{{ formatToolPayload(call.input) }}</pre>
            </div>
            <div v-if="call.output" class="ai-tool-payload">
              <span>返回结果</span>
              <pre>{{ formatToolPayload(call.output) }}</pre>
            </div>
            <div v-if="call.error" class="ai-tool-error">
              <span>错误</span>
              <pre>{{ call.error }}</pre>
            </div>
          </article>
        </div>
      </details>
      <div v-else class="ai-tool-empty">本次分析未调用工具</div>
    </template>

    <p v-else class="ai-block-hint">生成后会在这里显示建议内容。</p>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
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

const searchCallCount = computed(() =>
  aiAnalysis.value?.toolCalls?.filter(call => call.name === 'web_search').length || 0
)

const toolLabel = (name: string) => {
  const labels: Record<string, string> = {
    web_search: '网络搜索',
    get_local_file_metadata: '本地文件元数据',
    extract_document_text: '文档文本提取',
    list_archive_entries: '压缩包目录读取',
  }
  return labels[name] || name
}

const formatToolPayload = (value?: string) => {
  if (!value) return '—'
  try {
    return JSON.stringify(JSON.parse(value), null, 2)
  } catch {
    return value
  }
}

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

.ai-block-review {
  margin-top: 10px;
  padding: 9px 11px;
  border-radius: 9px;
  background: color-mix(in srgb, #d97706 12%, transparent);
  color: #b45309;
  font-size: 12px;
  line-height: 1.55;
}

.ai-block-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 10px;
}

.ai-diagnostic-details {
  margin-top: 10px;
  color: var(--import-text-faint, var(--text-faint));
  font-size: 11px;
}

.ai-diagnostic-details summary {
  cursor: pointer;
  font-weight: 700;
}

.ai-diagnostic-error {
  margin-top: 7px;
  color: #b45309;
  line-height: 1.55;
  word-break: break-word;
}

.ai-diagnostic-meta {
  margin-top: 5px;
  color: var(--import-text-faint, var(--text-faint));
  line-height: 1.55;
}

.ai-block-actions .btn {
  min-height: 32px;
  padding: 6px 12px;
  font-size: 12px;
}

.ai-tool-details {
  margin-top: 16px;
  border-top: 1px solid color-mix(in srgb, var(--border-color) 78%, transparent);
  border-bottom: 1px solid color-mix(in srgb, var(--border-color) 78%, transparent);
}

.ai-tool-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 42px;
  cursor: pointer;
  list-style: none;
  color: var(--text-color);
  font-size: 12px;
}

.ai-tool-summary::-webkit-details-marker {
  display: none;
}

.ai-tool-summary::before {
  content: '›';
  order: 2;
  color: var(--primary-color);
  font-size: 19px;
  line-height: 1;
  transition: transform 0.2s ease;
}

.ai-tool-details[open] .ai-tool-summary::before {
  transform: rotate(90deg);
}

.ai-tool-summary-title {
  font-weight: 700;
}

.ai-tool-summary-meta {
  margin-left: auto;
  color: var(--import-text-faint, var(--text-faint));
}

.ai-tool-call-list {
  display: grid;
  gap: 9px;
  padding: 0 0 12px;
}

.ai-tool-call {
  padding: 11px 12px;
  border: 1px solid color-mix(in srgb, var(--border-color) 78%, transparent);
  border-radius: 10px;
  background: color-mix(in srgb, var(--surface-color) 88%, transparent);
}

.ai-tool-call-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  color: var(--text-color);
  font-size: 12px;
}

.ai-tool-call-index {
  display: inline-block;
  width: 28px;
  margin-right: 5px;
  color: var(--primary-color);
  font-family: ui-monospace, SFMono-Regular, Consolas, monospace;
  font-size: 11px;
}

.ai-tool-status {
  padding: 3px 7px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 700;
}

.ai-tool-status.is-success {
  background: color-mix(in srgb, #16a34a 13%, transparent);
  color: #15803d;
}

.ai-tool-status.is-failed {
  background: color-mix(in srgb, #dc2626 13%, transparent);
  color: #b91c1c;
}

.ai-tool-call-meta {
  margin-top: 5px;
  color: var(--import-text-faint, var(--text-faint));
  font-size: 11px;
}

.ai-tool-payload,
.ai-tool-error {
  margin-top: 10px;
}

.ai-tool-payload > span,
.ai-tool-error > span {
  display: block;
  margin-bottom: 5px;
  color: var(--import-text-faint, var(--text-faint));
  font-size: 11px;
  font-weight: 700;
}

.ai-tool-payload pre,
.ai-tool-error pre {
  max-height: 150px;
  margin: 0;
  overflow: auto;
  padding: 8px 9px;
  border-radius: 7px;
  background: color-mix(in srgb, var(--background-color, #000) 11%, transparent);
  color: var(--import-text-soft, var(--text-soft));
  font: 11px/1.55 ui-monospace, SFMono-Regular, Consolas, monospace;
  white-space: pre-wrap;
  word-break: break-word;
}

.ai-tool-error pre {
  color: #b91c1c;
}

.ai-tool-empty {
  margin-top: 15px;
  color: var(--import-text-faint, var(--text-faint));
  font-size: 11px;
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

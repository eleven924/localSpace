<template>
  <div class="metadata-fields">
    <div class="field-block">
      <label for="file-tags">
        <span class="label-text">标签</span>
        <span class="label-hint">用逗号分隔</span>
      </label>
      <input
        id="file-tags"
        v-model="tagsStr"
        type="text"
        placeholder="工作, 重要, 报告"
      />
      <div v-if="parsedTags.length > 0" class="tags-preview">
        <span v-for="tag in parsedTags" :key="tag" class="tag-preview">
          {{ tag }}
          <button class="tag-remove" type="button" @click="removeTag(tag)">×</button>
        </span>
      </div>
    </div>

    <div class="field-block">
      <label for="file-description">
        <span class="label-text">描述</span>
        <span class="label-hint">可选</span>
      </label>
      <textarea
        id="file-description"
        v-model="description"
        rows="4"
        placeholder="输入文件的描述信息..."
      />
      <div class="char-count">
        {{ description.length }}/500
      </div>
    </div>

    <div v-if="embedAi" class="ai-section" :class="{ 'is-active': aiEnabled }">
      <FileAISuggestions
        :file-name="fileName"
        :file-type="fileType"
        :user-keywords="userKeywords"
        :tags="modelValueTags"
        :description="modelValueDescription"
        @update:tags="emit('update:modelValueTags', $event)"
        @update:description="emit('update:modelValueDescription', $event)"
        @availability="aiEnabled = $event"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import FileAISuggestions from './FileAISuggestions.vue'

const props = withDefaults(
  defineProps<{
    fileName: string
    fileType: string
    userKeywords?: string
    modelValueTags: string[]
    modelValueDescription: string
    // 导入页把 AI 建议放在独立面板里，所以需要关掉内嵌的那一份。
    embedAi?: boolean
  }>(),
  { embedAi: true }
)

const emit = defineEmits<{
  'update:modelValueTags': [value: string[]]
  'update:modelValueDescription': [value: string]
}>()

const aiEnabled = ref(false)

const tagsStr = computed({
  get: () => props.modelValueTags.join(', '),
  set: (value: string) => {
    const next = Array.from(
      new Set(
        value
          .split(',')
          .map(tag => tag.trim())
          .filter(Boolean)
      )
    )
    emit('update:modelValueTags', next)
  }
})

const description = computed({
  get: () => props.modelValueDescription,
  set: (value: string) => emit('update:modelValueDescription', value)
})

const parsedTags = computed(() => props.modelValueTags)

const removeTag = (tag: string) => {
  emit('update:modelValueTags', props.modelValueTags.filter(item => item !== tag))
}
</script>

<style scoped>
.metadata-fields {
  display: grid;
  gap: 14px;
}

.field-block {
  display: grid;
  gap: 8px;
}

.field-block label {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
}

.label-text {
  font-size: 12px;
  font-weight: 600;
  color: var(--import-text-soft, var(--text-color));
}

.label-hint {
  font-size: 12px;
  color: var(--import-text-faint, var(--text-faint));
}

.field-block input,
.field-block textarea {
  width: 100%;
  padding: 11px 12px;
  border: 1px solid color-mix(in srgb, var(--border-color) 84%, transparent);
  border-radius: 10px;
  background: color-mix(in srgb, var(--surface-color) 92%, transparent);
  color: var(--text-color);
  font-size: 14px;
  font-family: inherit;
  transition: border-color 0.2s ease, box-shadow 0.2s ease, background-color 0.2s ease;
}

.field-block input:focus,
.field-block textarea:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--primary-color) 16%, transparent);
}

.field-block textarea {
  resize: vertical;
  min-height: 96px;
}

.char-count {
  text-align: right;
  font-size: 12px;
  color: var(--import-text-faint, var(--text-faint));
}

.tags-preview {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-preview {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 10px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--primary-color) 12%, transparent);
  color: var(--primary-color);
  border: 1px solid color-mix(in srgb, var(--primary-color) 18%, transparent);
  font-size: 12px;
  font-weight: 600;
}

.tag-remove {
  background: none;
  border: none;
  color: inherit;
  cursor: pointer;
  padding: 0;
  line-height: 1;
}

.tag-remove:hover {
  opacity: 0.82;
}

/* AI 建议本身由 FileAISuggestions 渲染，这里只提供外框；
   未配置 AI 时子组件不输出内容，但必须保持挂载才能上报可用状态。 */
.ai-section {
  padding: 12px;
  border: 1px solid color-mix(in srgb, var(--border-color) 76%, transparent);
  border-radius: 12px;
  background: color-mix(in srgb, var(--surface-color) 68%, transparent);
}

.ai-section:not(.is-active) {
  display: none;
}
</style>

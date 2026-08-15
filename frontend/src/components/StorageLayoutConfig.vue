<template>
  <div class="storage-layout-config">
    <div v-if="loading" class="set-locked">
      <div>
        <span class="set-spinner" aria-hidden="true"></span>
        <p>正在读取存储规则...</p>
      </div>
    </div>

    <template v-else>
      <div class="set-rows">
        <div class="set-row">
          <div class="set-row-copy">
            <label for="storage-strategy">存储结构</label>
            <p>“按类型 + 合集” 会把文件放到 <code>类型/合集/文件名</code> 这一层。</p>
          </div>
          <div class="set-control">
            <select
              id="storage-strategy"
              v-model="config.strategy"
              class="set-field"
              @change="handleChange"
            >
              <option value="type_collection">按类型 + 合集</option>
              <option value="type_only">按类型</option>
            </select>
          </div>
        </div>

        <div class="set-row">
          <div class="set-row-copy">
            <label>未分配目录</label>
            <p>未填写合集时使用内置目录名；多个主目录会各自创建这一目录。</p>
          </div>
          <div class="set-control">
            <code class="set-fixed-value">_unsorted</code>
          </div>
        </div>

        <div class="set-row">
          <div class="set-row-copy">
            <label for="sanitize-folder-name">自动清理非法目录字符</label>
            <p>建议开启，避免 Windows 路径中出现非法字符。</p>
          </div>
          <div class="set-control">
            <label class="set-switch">
              <input
                id="sanitize-folder-name"
                v-model="config.sanitizeFolderName"
                type="checkbox"
                @change="handleChange"
              />
              <span class="set-switch-track"></span>
              <span class="set-switch-label">
                {{ config.sanitizeFolderName ? '已开启' : '已关闭' }}
              </span>
            </label>
          </div>
        </div>
      </div>

      <div class="set-preview">
        <b>路径预览</b>
        <code>{{ pathPreview }}</code>
      </div>

      <div class="set-commit">
        <p v-if="message" class="set-feedback" :class="message.success ? 'success' : 'error'">
          {{ message.text }}
        </p>
        <button v-if="hasChanges" type="button" class="btn secondary" @click="handleReset">
          重置
        </button>
        <button type="button" class="btn primary" :disabled="saving" @click="handleSave">
          {{ saving ? '保存中...' : '保存规则' }}
        </button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '@/api'
import type { StorageLayoutConfig as StorageLayoutConfigModel } from '@/types'

const createDefaultConfig = (): StorageLayoutConfigModel => ({
  strategy: 'type_collection',
  sanitizeFolderName: true,
})

const config = ref<StorageLayoutConfigModel>(createDefaultConfig())
const originalConfig = ref<StorageLayoutConfigModel>(createDefaultConfig())
const loading = ref(false)
const saving = ref(false)
const message = ref<{ success: boolean; text: string } | null>(null)

const hasChanges = computed(() => JSON.stringify(config.value) !== JSON.stringify(originalConfig.value))

// 预览跟着选择走：规则改了但看不到结果，用户只能靠猜。
const pathPreview = computed(() =>
  config.value.strategy === 'type_collection'
    ? `视频 / 甄嬛传 / 第01集.mkv`
    : '视频 / 第01集.mkv',
)

onMounted(() => {
  void loadConfig()
})

const normalizeConfig = (raw: any): StorageLayoutConfigModel => ({
  strategy: raw?.strategy === 'type_only' ? 'type_only' : 'type_collection',
  sanitizeFolderName: raw?.sanitizeFolderName !== false,
})

const loadConfig = async () => {
  loading.value = true
  message.value = null

  try {
    const loaded = await api.storageLayout.getConfig()
    const normalized = normalizeConfig(loaded)
    config.value = { ...normalized }
    originalConfig.value = { ...normalized }
  } catch (error) {
    console.error('Failed to load storage layout config:', error)
    message.value = { success: false, text: '加载存储规则失败' }
  } finally {
    loading.value = false
  }
}

const handleChange = () => {
  message.value = null
}

const handleSave = async () => {
  saving.value = true
  message.value = null

  try {
    const payload = normalizeConfig(config.value)
    await api.storageLayout.updateConfig(payload)
    config.value = { ...payload }
    originalConfig.value = { ...payload }
    message.value = { success: true, text: '存储规则已保存' }
  } catch (error) {
    console.error('Failed to save storage layout config:', error)
    message.value = { success: false, text: '保存存储规则失败' }
  } finally {
    saving.value = false
  }
}

const handleReset = () => {
  config.value = { ...originalConfig.value }
  message.value = null
}
</script>

<style scoped>
.storage-layout-config {
  width: 100%;
}

.set-fixed-value {
  display: inline-flex;
  align-items: center;
  min-height: 38px;
  padding: 0 12px;
  border: 1px solid var(--set-border, var(--border-color));
  border-radius: 10px;
  background: var(--set-control, var(--surface-muted));
  color: var(--set-text-soft, var(--text-soft));
  font-size: 13px;
}

.set-locked p {
  margin-top: 10px;
}
</style>

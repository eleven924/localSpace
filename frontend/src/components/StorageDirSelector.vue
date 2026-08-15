<template>
  <div class="storage-dir-selector">
    <div v-if="loading" class="set-locked">
      <div>
        <span class="set-spinner" aria-hidden="true"></span>
        <p>正在读取存储目录...</p>
      </div>
    </div>

    <div v-else-if="error" class="set-empty">
      <strong>无法读取存储目录</strong>
      <p>{{ error }}</p>
      <button type="button" class="btn secondary retry-button" @click="loadMasterDirectories">
        重试
      </button>
    </div>

    <template v-else>
      <div v-if="masterDirectories.length > 0" class="set-list">
        <MasterDirCard
          v-for="master in masterDirectories"
          :key="master.id"
          :master="master"
          @set-default="handleSetDefault"
          @delete="requestDelete"
          @toggle-expanded="handleToggleExpanded"
        />
      </div>

      <div v-else class="set-empty">
        <strong>暂无主目录</strong>
        <p>点击下方按钮添加主目录，导入的文件才有落脚的地方。</p>
      </div>

      <!-- 添加主目录不是「提交」，所以它紧跟在清单下面，不进底部动作行。 -->
      <div class="set-inline-action">
        <button type="button" class="btn primary" @click="openAddDialog">添加主目录</button>
        <p v-if="feedback" class="set-feedback" :class="feedback.ok ? 'success' : 'error'">
          {{ feedback.text }}
        </p>
        <p v-else class="set-hint">主目录是资料落盘的根位置，子目录按文件类型自动创建。</p>
      </div>
    </template>

    <!-- 添加主目录 -->
    <div v-if="showAddDialog" class="set-modal" @click.self="closeAddDialog">
      <div class="set-modal-panel" role="dialog" aria-modal="true" aria-labelledby="add-dir-title">
        <header class="set-modal-head">
          <span class="set-modal-mark" aria-hidden="true">+</span>
          <div>
            <h3 id="add-dir-title">添加主目录</h3>
            <p>主目录是资料的落盘根位置，下面会按文件类型自动分子目录。</p>
          </div>
        </header>

        <div class="set-modal-body">
          <div class="set-modal-field">
            <label for="master-path">主目录路径</label>
            <div class="set-control inline">
              <input
                id="master-path"
                ref="pathInput"
                v-model="newMaster.path"
                class="set-field mono"
                type="text"
                placeholder="选择或输入主目录路径"
                @input="handlePathInput"
              />
              <button type="button" class="set-mini" @click="handleBrowseDir">浏览</button>
            </div>
            <p v-if="pathConflict" class="set-feedback error">
              路径冲突：该路径已存在或包含在其他目录中
            </p>
          </div>

          <div class="set-modal-field">
            <label for="max-size">最大容量 (GB)</label>
            <input
              id="max-size"
              v-model.number="newMaster.maxSize"
              class="set-field"
              type="number"
              placeholder="0"
              min="0"
              step="1"
            />
            <p class="set-hint">填 0 表示不限制容量。</p>
          </div>

          <p class="set-modal-target">
            {{
              masterCount === 0
                ? '这将是您的第一个主目录，将自动设为默认'
                : `当前有 ${masterCount} 个主目录，1 个为默认`
            }}
          </p>
        </div>

        <footer class="set-modal-foot">
          <button type="button" class="btn secondary" @click="closeAddDialog">取消</button>
          <button type="button" class="btn primary" :disabled="!isFormValid" @click="handleAddMaster">
            确定
          </button>
        </footer>
      </div>
    </div>

    <!-- 删除主目录会清空目录内容，是整页风险最高的动作，所以它有自己的确认对话框。 -->
    <div v-if="pendingDelete" class="set-modal" @click.self="cancelDelete">
      <div
        class="set-modal-panel danger"
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="delete-dir-title"
      >
        <header class="set-modal-head">
          <span class="set-modal-mark" aria-hidden="true">!</span>
          <div>
            <h3 id="delete-dir-title">删除主目录及其所有子目录</h3>
            <p>此操作不可恢复，请先确认重要数据已备份。</p>
          </div>
        </header>

        <div class="set-modal-body">
          <p class="set-modal-target path">{{ pendingDelete.path }}</p>
          <p class="set-modal-warning">
            <b>警告：此操作将清空该目录下的所有文件内容。</b>
            目录中已导入的资料会一并被删除，资料库里对应的记录也会失效。
          </p>
        </div>

        <footer class="set-modal-foot">
          <button type="button" class="btn secondary" @click="cancelDelete">取消</button>
          <button
            type="button"
            class="btn secondary set-danger-action"
            :disabled="deleting"
            @click="confirmDelete"
          >
            {{ deleting ? '正在删除...' : '删除目录' }}
          </button>
        </footer>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import { api } from '@/api/index'
import MasterDirCard from './MasterDirCard.vue'

interface StorageDir {
  id: number
  path: string
  fileType: string
  currentSize: number
  maxSize?: number
  isActive: boolean
  isDefault: boolean
  parentId?: number
  createdAt: string
  totalSize?: number
  subDirs?: StorageDir[]
  expanded?: boolean
}

const emit = defineEmits<{
  masterDirAdded: []
  masterDirRemoved: []
  masterDirDefaultChanged: []
}>()

const masterDirectories = ref<StorageDir[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const showAddDialog = ref(false)
const pathConflict = ref(false)
const pathInput = ref<HTMLInputElement | null>(null)
const pendingDelete = ref<StorageDir | null>(null)
const deleting = ref(false)
const feedback = ref<{ text: string; ok: boolean } | null>(null)
let feedbackTimer: number | undefined

const newMaster = ref({
  path: '',
  maxSize: 0,
})

const masterCount = computed(() => masterDirectories.value.length)

// 表单验证
const isFormValid = computed(() => {
  return newMaster.value.path.trim().length > 0 && !pathConflict.value
})

// 这一分区的动作即时生效，没有提交按钮，所以结果直接写在清单下方那一行。
const setFeedback = (text: string, ok: boolean) => {
  feedback.value = { text, ok }
  window.clearTimeout(feedbackTimer)
  feedbackTimer = window.setTimeout(() => {
    feedback.value = null
  }, ok ? 2600 : 5000)
}

const handleEscape = (event: KeyboardEvent) => {
  if (event.key !== 'Escape') return
  if (deleting.value) return
  showAddDialog.value = false
  pendingDelete.value = null
}

onMounted(() => {
  loadMasterDirectories()
  window.addEventListener('keydown', handleEscape)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleEscape)
  window.clearTimeout(feedbackTimer)
})

const loadMasterDirectories = async () => {
  loading.value = true
  error.value = null

  try {
    const dirs = await api.storage.getMasterDirectories()
    masterDirectories.value = dirs.map((dir: any) => ({
      ...dir,
      expanded: false,
      subDirs: dir.subDirs || [],
      totalSize: dir.totalSize || 0
    }))
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载存储目录失败'
    console.error('Failed to load master directories:', err)
  } finally {
    loading.value = false
  }
}

const openAddDialog = async () => {
  showAddDialog.value = true
  await nextTick()
  pathInput.value?.focus()
}

const closeAddDialog = () => {
  showAddDialog.value = false
}

const handlePathInput = async () => {
  if (newMaster.value.path.trim().length > 0) {
    const hasConflict = await api.storage.checkPathConflict(newMaster.value.path, 0)
    pathConflict.value = hasConflict
  } else {
    pathConflict.value = false
  }
}

const handleBrowseDir = async () => {
  try {
    const dirPath = await api.system.selectDirectory()
    if (dirPath) {
      newMaster.value.path = dirPath
      await handlePathInput()
    }
  } catch (err) {
    console.error('Failed to browse directory:', err)
    setFeedback('选择目录失败', false)
  }
}

const handleAddMaster = async () => {
  if (!isFormValid.value) return

  try {
    await api.storage.addMasterDirectory(
      newMaster.value.path,
      newMaster.value.maxSize
    )

    const addedPath = newMaster.value.path.trim()
    showAddDialog.value = false
    newMaster.value = { path: '', maxSize: 0 }
    pathConflict.value = false
    await loadMasterDirectories()
    setFeedback(`已添加主目录 ${addedPath}`, true)
    emit('masterDirAdded')
  } catch (err) {
    console.error('Failed to add master directory:', err)
    setFeedback('添加主目录失败：' + (err instanceof Error ? err.message : '未知错误'), false)
  }
}

const handleSetDefault = async (id: number) => {
  try {
    await api.storage.setDefaultMasterDirectory(id)
    await loadMasterDirectories()
    setFeedback('已设为默认主目录', true)
    emit('masterDirDefaultChanged')
  } catch (err) {
    console.error('Failed to set default master directory:', err)
    setFeedback('设置默认主目录失败', false)
  }
}

const requestDelete = (id: number) => {
  pendingDelete.value = masterDirectories.value.find((master) => master.id === id) || null
}

const cancelDelete = () => {
  if (deleting.value) return
  pendingDelete.value = null
}

const confirmDelete = async () => {
  const target = pendingDelete.value
  if (!target) return

  deleting.value = true
  try {
    await api.storage.removeDirectory(target.id)
    pendingDelete.value = null
    await loadMasterDirectories()
    setFeedback('主目录已删除', true)
    emit('masterDirRemoved')
  } catch (err) {
    console.error('Failed to remove master directory:', err)
    pendingDelete.value = null
    setFeedback('删除主目录失败：' + (err instanceof Error ? err.message : '未知错误'), false)
  } finally {
    deleting.value = false
  }
}

const handleToggleExpanded = (id: number) => {
  const dir = masterDirectories.value.find(m => m.id === id)
  if (dir) {
    dir.expanded = !dir.expanded
  }
}
</script>

<style scoped>
.storage-dir-selector {
  width: 100%;
}

.set-locked p {
  margin-top: 10px;
}

.retry-button {
  margin-top: 14px;
}
</style>

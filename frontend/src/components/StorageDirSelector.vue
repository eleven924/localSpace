<template>
  <div class="storage-dir-selector">
    <div class="selector-header">
      <h4>存储目录管理</h4>
      <p class="subtitle">管理主目录及其子目录</p>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else-if="error" class="error-state">
      <span class="error-icon">⚠️</span>
      <p>{{ error }}</p>
      <button class="btn secondary" @click="loadMasterDirectories">重试</button>
    </div>

    <div v-else class="storage-content">
      <div v-if="masterDirectories.length === 0" class="empty-state">
        <div class="empty-icon">📁</div>
        <h4>暂无主目录</h4>
        <p>点击下方按钮添加主目录</p>
      </div>

      <div v-else class="master-list">
        <MasterDirCard
          v-for="master in masterDirectories"
          :key="master.id"
          :master="master"
          @set-default="handleSetDefault"
          @delete="handleDelete"
          @toggle-expanded="handleToggleExpanded"
        />
      </div>

      <button class="add-button" @click="showAddDialog = true">
        <span class="add-icon">➕</span>
        <span class="add-text">添加主目录</span>
      </button>
    </div>

    <!-- 添加主目录对话框 -->
    <div v-if="showAddDialog" class="dialog-overlay">
      <div class="dialog-content">
        <h4>添加主目录</h4>

        <div class="form-group">
          <label for="master-path">主目录路径</label>
          <div class="path-input-wrapper">
            <input
              id="master-path"
              v-model="newMaster.path"
              type="text"
              placeholder="选择或输入主目录路径"
              @input="handlePathInput"
            />
            <button class="browse-button" @click="handleBrowseDir">
              浏览
            </button>
          </div>
          <div v-if="pathConflict" class="error-text">
            ⚠️ 路径冲突：该路径已存在或包含在其他目录中
          </div>
        </div>

        <div class="form-group">
          <label for="max-size">最大容量 (GB)</label>
          <input
            id="max-size"
            v-model.number="newMaster.maxSize"
            type="number"
            placeholder="0 表示无限制"
            min="0"
            step="1"
          />
        </div>

        <div class="form-group">
          <label>预计使用情况</label>
          <div class="usage-preview">
            <span v-if="masterCount === 0" class="preview-text">
              这将是您的第一个主目录，将自动设为默认
            </span>
            <span v-else class="preview-text">
              当前有 {{ masterCount }} 个主目录，1个为默认
            </span>
          </div>
        </div>

        <div class="dialog-actions">
          <button class="btn secondary" @click="showAddDialog = false">
            取消
          </button>
          <button
            class="btn primary"
            @click="handleAddMaster"
            :disabled="!isFormValid"
          >
            确定
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '@/api/index'
import MasterDirCard from './MasterDirCard.vue'
import { formatFileSize } from '@/utils/constants'

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

const newMaster = ref({
  path: '',
  maxSize: 0,
})

const masterCount = computed(() => masterDirectories.value.length)

// 表单验证
const isFormValid = computed(() => {
  return newMaster.value.path.trim().length > 0 && !pathConflict.value
})

onMounted(() => {
  loadMasterDirectories()
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
    alert('选择目录失败')
  }
}

const handleAddMaster = async () => {
  if (!isFormValid.value) return

  try {
    await api.storage.addMasterDirectory(
      newMaster.value.path,
      newMaster.value.maxSize
    )

    showAddDialog.value = false
    newMaster.value = { path: '', maxSize: 0 }
    pathConflict.value = false
    await loadMasterDirectories()
    emit('masterDirAdded')
  } catch (err) {
    console.error('Failed to add master directory:', err)
    alert('添加主目录失败: ' + (err instanceof Error ? err.message : '未知错误'))
  }
}

const handleSetDefault = async (id: number) => {
  try {
    await api.storage.setDefaultMasterDirectory(id)
    await loadMasterDirectories()
    emit('masterDirDefaultChanged')
  } catch (err) {
    console.error('Failed to set default master directory:', err)
    alert('设置默认主目录失败')
  }
}

const handleDelete = async (id: number) => {
  if (!confirm('确定要删除这个主目录及其所有子目录吗？\n\n⚠️ 警告：此操作将清空该目录下的所有文件内容！\n\n请确认您已备份重要数据，此操作不可恢复。')) {
    return
  }

  try {
    await api.storage.removeDirectory(id)
    await loadMasterDirectories()
    emit('masterDirRemoved')
  } catch (err) {
    console.error('Failed to remove master directory:', err)
    alert('删除主目录失败: ' + (err instanceof Error ? err.message : '未知错误'))
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

.selector-header {
  margin-bottom: 20px;
}

.selector-header h4 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-color);
  margin: 0 0 8px 0;
}

.subtitle {
  font-size: 14px;
  color: var(--text-color);
  opacity: 0.7;
  margin: 0;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  gap: 16px;
  color: var(--text-color);
}

.spinner {
  border: 2px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  width: 20px;
  height: 20px;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.loading-state p,
.error-state p {
  font-size: 14px;
  margin: 0;
  opacity: 0.8;
}

.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 32px 20px;
  text-align: center;
}

.error-icon {
  font-size: 32px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: var(--text-color);
  opacity: 0.6;
  text-align: center;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 12px;
}

.empty-state h4 {
  font-size: 16px;
  font-weight: 500;
  margin: 0 0 8px 0;
}

.empty-state p {
  font-size: 14px;
  margin: 0;
}

.storage-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.master-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.add-button {
  width: 100%;
  padding: 12px 20px;
  background-color: var(--primary-color);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.add-button:hover {
  opacity: 0.9;
  box-shadow: 0 2px 8px var(--shadow-color);
}

.add-icon {
  font-size: 16px;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog-content {
  background-color: var(--surface-color);
  border-radius: 12px;
  padding: 24px;
  max-width: 500px;
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;
}

.dialog-content h4 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-color);
  margin: 0 0 20px 0;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color);
}

.path-input-wrapper {
  display: flex;
  gap: 8px;
}

.path-input-wrapper input {
  flex: 1;
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background-color: var(--bg-color);
  color: var(--text-color);
  font-size: 14px;
}

.path-input-wrapper input:focus {
  outline: none;
  border-color: var(--primary-color);
}

.browse-button {
  padding: 10px 16px;
  background-color: var(--surface-color);
  color: var(--text-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.browse-button:hover {
  background-color: var(--border-color);
}

.form-group input[type="number"] {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background-color: var(--bg-color);
  color: var(--text-color);
  font-size: 14px;
}

.form-group input[type="number"]:focus {
  outline: none;
  border-color: var(--primary-color);
}

.error-text {
  margin-top: 6px;
  font-size: 12px;
  color: var(--error-color);
}

.usage-preview {
  padding: 12px;
  background-color: var(--bg-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
}

.preview-text {
  font-size: 13px;
  color: var(--text-color);
  opacity: 0.7;
}

.dialog-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
}

.dialog-actions .btn {
  flex: 1;
  padding: 10px 16px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}

.dialog-actions .btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .master-list {
    flex-direction: column;
  }
}
</style>
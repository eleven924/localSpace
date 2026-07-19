<template>
  <div class="theme-config">
    <div class="config-header">
      <h4>主题设置</h4>
      <p class="subtitle">自定义应用的外观和风格</p>
    </div>

    <div class="config-content">
      <!-- 主题模式选择 -->
      <div class="config-section">
        <h5>主题模式</h5>
        <div class="theme-modes">
          <button
            v-for="mode in themeModes"
            :key="mode.value"
            :class="{ active: themeStore.themeMode === mode.value }"
            @click="handleSetThemeMode(mode.value)"
          >
            <span class="mode-icon">{{ mode.icon }}</span>
            <span class="mode-label">{{ mode.label }}</span>
          </button>
        </div>
      </div>

      <!-- 主题颜色选择 -->
      <div class="config-section">
        <h5>主题颜色</h5>
        <div class="color-options">
          <div
            v-for="color in presetColors"
            :key="color"
            class="color-option"
            :class="{ active: themeStore.primaryColor === color }"
            :style="{ backgroundColor: color }"
            @click="handleSetColor(color)"
            :title="color"
          >
            <span v-if="themeStore.primaryColor === color" class="checkmark">✓</span>
          </div>
        </div>
        <div class="custom-color">
          <label for="custom-color" class="custom-color-label">
            <span class="custom-color-icon">🎨</span>
            <span>自定义颜色</span>
          </label>
          <div class="color-input-wrapper">
            <input
              id="custom-color"
              type="color"
              :value="themeStore.primaryColor"
              @input="handleCustomColorChange"
              class="color-input"
            />
            <input
              type="text"
              :value="themeStore.primaryColor"
              @input="handleColorTextChange"
              class="color-text-input"
              placeholder="#2196F3"
            />
          </div>
        </div>
      </div>

      <!-- 背景设置 -->
      <div class="config-section">
        <h5>背景图片</h5>
        <div class="background-options">
          <div class="current-background" v-if="themeStore.backgroundImage">
            <div class="background-preview" :style="{ backgroundImage: `url(${themeStore.backgroundImage})` }"></div>
            <button class="remove-bg-button" @click="handleRemoveBackground">
              移除背景
            </button>
          </div>
          <div v-else class="no-background">
            <span class="no-bg-icon">🖼️</span>
            <p>未设置背景图片</p>
          </div>
          <div class="upload-section">
            <label for="background-upload" class="upload-button">
              <span class="upload-icon">📤</span>
              <span>上传背景图片</span>
            </label>
            <input
              id="background-upload"
              type="file"
              @change="handleBackgroundUpload"
              accept="image/*"
              class="hidden-input"
            />
            <p class="upload-hint">支持 JPG、PNG、GIF 格式</p>
          </div>
        </div>
      </div>

      <!-- 重置按钮 -->
      <div class="config-section">
        <button class="btn reset-button" @click="handleResetTheme">
          <span class="reset-icon">🔄</span>
          <span>重置为默认主题</span>
        </button>
      </div>
    </div>

    <!-- 预览卡片 -->
    <div class="theme-preview">
      <h5>预览</h5>
      <div class="preview-card">
        <div class="preview-header">
          <div class="preview-title">示例标题</div>
          <div class="preview-badge">标签</div>
        </div>
        <div class="preview-body">
          <p>这是一段示例文本，用于展示当前主题的样式效果。</p>
        </div>
        <div class="preview-footer">
          <button class="preview-button">主要按钮</button>
          <button class="preview-button secondary">次要按钮</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useThemeStore } from '@/store/modules/theme'

const emit = defineEmits<{
  themeChanged: []
  themeReset: []
}>()

const themeStore = useThemeStore()

const themeModes = [
  { value: 'light', label: '浅色', icon: '☀️' },
  { value: 'dark', label: '深色', icon: '🌙' }
]

const presetColors = [
  '#2196F3', // Blue
  '#4CAF50', // Green
  '#FF9800', // Orange
  '#E91E63', // Pink
  '#9C27B0', // Purple
  '#00BCD4', // Cyan
  '#F44336', // Red
  '#607D8B', // Blue Gray
  '#3F51B5', // Indigo
  '#FF5722', // Deep Orange
]

onMounted(() => {
  // 主题已通过 main.ts 初始化
})

const handleSetThemeMode = (mode: 'light' | 'dark') => {
  themeStore.setThemeMode(mode)
  emit('themeChanged')
}

const handleSetColor = (color: string) => {
  themeStore.setPrimaryColor(color)
  emit('themeChanged')
}

const handleCustomColorChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  themeStore.setPrimaryColor(target.value)
  emit('themeChanged')
}

const handleColorTextChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  const color = target.value.trim()

  // 验证颜色格式
  if (/^#[0-9A-Fa-f]{6}$/.test(color)) {
    themeStore.setPrimaryColor(color)
    emit('themeChanged')
  }
}

const handleBackgroundUpload = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]

  if (!file) return

  // 验证文件类型
  if (!file.type.match(/^image\/(jpeg|png|gif|webp)$/)) {
    alert('请选择有效的图片文件（JPG、PNG、GIF、WebP）')
    return
  }

  // 验证文件大小（最大 5MB）
  if (file.size > 5 * 1024 * 1024) {
    alert('图片文件不能超过 5MB')
    return
  }

  try {
    // 读取文件为 Data URL
    const reader = new FileReader()
    reader.onload = (e) => {
      const dataUrl = e.target?.result as string
      themeStore.setBackgroundImage(dataUrl)
      emit('themeChanged')
    }
    reader.onerror = () => {
      alert('读取图片文件失败')
    }
    reader.readAsDataURL(file)
  } catch (err) {
    console.error('Failed to upload background:', err)
    alert('上传背景图片失败')
  }

  // 清空 input 以允许重复选择同一文件
  target.value = ''
}

const handleRemoveBackground = () => {
  if (confirm('确定要移除背景图片吗？')) {
    themeStore.setBackgroundImage('')
    emit('themeChanged')
  }
}

const handleResetTheme = () => {
  if (confirm('确定要重置为默认主题吗？所有自定义设置将丢失。')) {
    themeStore.resetTheme()
    emit('themeReset')
  }
}
</script>

<style scoped>
.theme-config {
  width: 100%;
}

.config-header {
  margin-bottom: 14px;
}

.config-header h4 {
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

.config-content {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.config-section {
  padding: 14px 0;
  background-color: transparent;
  border: none;
  border-bottom: 1px solid var(--border-color);
  border-radius: 0;
}

.config-section h5 {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-color);
  margin: 0 0 10px 0;
}

/* 主题模式 */
.theme-modes {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}

.theme-modes button {
  min-height: 38px;
  padding: 8px 12px;
  background-color: var(--surface-color);
  color: var(--text-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  flex-direction: row;
  justify-content: center;
  align-items: center;
  gap: 6px;
}

.theme-modes button:hover {
  border-color: var(--primary-color);
  background-color: var(--surface-color);
}

.theme-modes button.active {
  border-color: var(--primary-color);
  background-color: var(--primary-color);
  color: white;
}

.mode-icon {
  font-size: 18px;
}

.mode-label {
  font-size: 14px;
  font-weight: 500;
}

/* 主题颜色 */
.color-options {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 14px;
}

.color-option {
  width: 34px;
  height: 34px;
  border-radius: 8px;
  cursor: pointer;
  border: 2px solid transparent;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.color-option:hover {
  transform: scale(1.1);
  box-shadow: 0 2px 8px var(--shadow-color);
}

.color-option.active {
  border-color: var(--text-color);
  box-shadow: 0 0 0 2px var(--bg-color), 0 0 0 4px var(--text-color);
}

.checkmark {
  color: white;
  font-size: 16px;
  font-weight: bold;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
}

.custom-color {
  display: grid;
  grid-template-columns: minmax(120px, 0.35fr) minmax(180px, 0.65fr);
  gap: 8px 14px;
  align-items: center;
}

.custom-color-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color);
  cursor: pointer;
}

.custom-color-icon {
  font-size: 18px;
}

.color-input-wrapper {
  display: flex;
  gap: 8px;
}

.color-input {
  width: 44px;
  height: 38px;
  padding: 2px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  cursor: pointer;
  background-color: var(--bg-color);
}

.color-input::-webkit-color-swatch-wrapper {
  padding: 0;
}

.color-input::-webkit-color-swatch {
  border: none;
  border-radius: 6px;
}

.color-text-input {
  flex: 1;
  min-height: 38px;
  padding: 8px 11px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background-color: var(--bg-color);
  color: var(--text-color);
  font-size: 14px;
  font-family: monospace;
}

.color-text-input:focus {
  outline: none;
  border-color: var(--primary-color);
}

/* 背景设置 */
.background-options {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.current-background {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.background-preview {
  width: 100%;
  height: 96px;
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  border-radius: 8px;
  border: 1px solid var(--border-color);
}

.remove-bg-button {
  padding: 8px 16px;
  background-color: var(--error-color);
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  align-self: flex-start;
}

.remove-bg-button:hover {
  opacity: 0.9;
}

.no-background {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: flex-start;
  gap: 10px;
  padding: 12px;
  border: 1px dashed var(--border-color);
  border-radius: 8px;
  color: var(--text-color);
  opacity: 0.6;
}

.no-bg-icon {
  font-size: 26px;
  margin-bottom: 0;
}

.no-background p {
  font-size: 14px;
  margin: 0;
}

.upload-section {
  display: grid;
  grid-template-columns: minmax(180px, 0.55fr) minmax(140px, 0.45fr);
  gap: 8px 12px;
  align-items: center;
}

.upload-button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 38px;
  padding: 9px 14px;
  background-color: var(--surface-color);
  color: var(--text-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.upload-button:hover {
  background-color: var(--border-color);
  border-color: var(--text-color);
}

.upload-icon {
  font-size: 18px;
}

.hidden-input {
  display: none;
}

.upload-hint {
  font-size: 12px;
  color: var(--text-color);
  opacity: 0.6;
  margin: 0;
}

/* 重置按钮 */
.reset-button {
  width: 100%;
  min-height: 38px;
  padding: 9px 14px;
  background-color: var(--surface-color);
  color: var(--text-color);
  border: 1px solid var(--border-color);
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

.reset-button:hover {
  background-color: var(--border-color);
  border-color: var(--text-color);
}

.reset-icon {
  font-size: 16px;
}

/* 预览 */
.theme-preview {
  margin-top: 0;
  padding: 14px 0 0;
  background-color: transparent;
  border: none;
  border-radius: 0;
}

.theme-preview h5 {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-color);
  margin: 0 0 12px 0;
}

.preview-card {
  padding: 14px;
  background-color: var(--surface-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.preview-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--text-color);
}

.preview-badge {
  padding: 4px 12px;
  background-color: var(--primary-color);
  color: white;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.preview-body {
  margin-bottom: 12px;
}

.preview-body p {
  font-size: 14px;
  color: var(--text-color);
  line-height: 1.6;
  margin: 0;
}

.preview-footer {
  display: flex;
  gap: 12px;
}

.preview-button {
  flex: 1;
  min-height: 36px;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.preview-button {
  background-color: var(--primary-color);
  color: white;
  border: none;
}

.preview-button:hover {
  opacity: 0.9;
}

.preview-button.secondary {
  background-color: var(--surface-color);
  color: var(--text-color);
  border: 1px solid var(--border-color);
}

.preview-button.secondary:hover {
  background-color: var(--border-color);
}

@media (max-width: 768px) {
  .theme-modes {
    grid-template-columns: 1fr;
  }

  .custom-color,
  .upload-section {
    grid-template-columns: 1fr;
  }
}
</style>

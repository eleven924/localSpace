<template>
  <div class="theme-config">
    <div class="config-content">
      <section class="config-section preference-row">
        <div class="section-copy">
          <h5>主题模式</h5>
          <p>选择适合当前环境的界面明暗模式。</p>
        </div>
        <div class="theme-modes" role="group" aria-label="主题模式">
          <button
            v-for="mode in themeModes"
            :key="mode.value"
            type="button"
            :class="{ active: themeStore.themeMode === mode.value }"
            @click="handleSetThemeMode(mode.value)"
          >
            {{ mode.label }}
          </button>
        </div>
      </section>

      <section class="config-section color-section">
        <div class="section-copy">
          <h5>强调色</h5>
          <p>用于导航选中态、主要按钮和关键操作提示。</p>
        </div>

        <div class="color-controls">
          <div class="color-options" role="listbox" aria-label="预设强调色">
            <button
              v-for="color in presetColors"
              :key="color"
              type="button"
              class="color-option"
              :class="{ active: themeStore.primaryColor === color }"
              :style="{ backgroundColor: color }"
              :aria-label="`选择颜色 ${color}`"
              :aria-selected="themeStore.primaryColor === color"
              @click="handleSetColor(color)"
            >
              <span v-if="themeStore.primaryColor === color" class="checkmark">✓</span>
            </button>
          </div>

          <div class="color-input-wrapper">
            <label for="custom-color" class="custom-color-label">自定义</label>
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
      </section>

      <section class="config-section background-section">
        <div class="section-copy">
          <h5>背景图片</h5>
          <p>背景会在应用外层透出，设置页内容区会保留稳定浅色遮罩以保证可读。</p>
        </div>

        <div class="background-options">
          <div
            class="background-preview"
            :class="{ empty: !themeStore.backgroundImage }"
            :style="themeStore.backgroundImage ? { backgroundImage: `url(${themeStore.backgroundImage})` } : undefined"
            role="img"
            :aria-label="themeStore.backgroundImage ? '当前背景图片预览' : '未设置背景图片'"
          >
            <span v-if="!themeStore.backgroundImage">未设置背景图片</span>
          </div>

          <div class="upload-section">
            <label for="background-upload" class="upload-button">
              <span>上传背景图片</span>
            </label>
            <input
              id="background-upload"
              type="file"
              @change="handleBackgroundUpload"
              accept="image/*"
              class="hidden-input"
            />
            <button
              v-if="themeStore.backgroundImage"
              type="button"
              class="remove-bg-button"
              @click="handleRemoveBackground"
            >
              移除背景
            </button>
            <p class="upload-hint">支持 JPG、PNG、GIF、WebP，最大 5MB</p>
          </div>
        </div>
      </section>

      <section class="config-section theme-actions">
        <div class="section-copy">
          <h5>恢复默认</h5>
          <p>重置主题模式、强调色和背景图片。</p>
        </div>
        <button type="button" class="btn reset-button" @click="handleResetTheme">
          <span>重置为默认主题</span>
        </button>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useThemeStore } from '@/store/modules/theme'

const emit = defineEmits<{
  themeChanged: []
  themeReset: []
}>()

const themeStore = useThemeStore()

const themeModes = [
  { value: 'light', label: '浅色' },
  { value: 'dark', label: '深色' }
]

const presetColors = [
  '#2196F3',
  '#4CAF50',
  '#FF9800',
  '#E91E63',
  '#9C27B0',
  '#00BCD4',
  '#F44336',
  '#607D8B',
  '#3F51B5',
  '#FF5722',
]

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

  // 输入框只在完整 Hex 颜色时提交，避免用户输入一半时频繁写入无效主题。
  if (/^#[0-9A-Fa-f]{6}$/.test(color)) {
    themeStore.setPrimaryColor(color)
    emit('themeChanged')
  }
}

const handleBackgroundUpload = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]

  if (!file) return

  // 上传前校验图片类型，避免把非图片内容写入主题配置。
  if (!file.type.match(/^image\/(jpeg|png|gif|webp)$/)) {
    alert('请选择有效的图片文件（JPG、PNG、GIF、WebP）')
    return
  }

  // 背景图存为 Data URL，限制体积可以避免配置过大影响启动速度。
  if (file.size > 5 * 1024 * 1024) {
    alert('图片文件不能超过 5MB')
    return
  }

  try {
    // 读取为 Data URL 后交给主题 Store 持久化，页面会立即应用新背景。
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

.config-content {
  display: grid;
  gap: 0;
}

.config-section {
  display: grid;
  gap: 14px;
  padding: 18px 0;
  background-color: transparent;
  border: none;
  border-bottom: 1px solid color-mix(in srgb, var(--border-color) 66%, transparent);
  border-radius: 0;
}

.preference-row,
.theme-actions {
  grid-template-columns: minmax(260px, 1fr) minmax(260px, auto);
  align-items: center;
  gap: 24px;
}

.config-section:first-child {
  padding-top: 0;
}

.config-section:last-child {
  border-bottom: none;
}

.section-copy {
  display: grid;
  gap: 6px;
  min-width: 0;
}

.config-section h5 {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-color);
  margin: 0;
}

.section-copy p {
  max-width: 560px;
  margin: 0;
  color: color-mix(in srgb, var(--text-soft) 92%, var(--text-color));
  font-size: 13px;
  line-height: 1.55;
}

/* 主题模式 */
.theme-modes {
  display: inline-flex;
  justify-self: end;
  min-width: 260px;
  padding: 3px;
  border: 1px solid color-mix(in srgb, var(--border-color) 84%, transparent);
  border-radius: 7px;
  background: color-mix(in srgb, var(--surface-muted) 70%, transparent);
}

.theme-modes button {
  flex: 1;
  min-height: 30px;
  padding: 5px 16px;
  background-color: transparent;
  color: color-mix(in srgb, var(--text-color) 84%, var(--text-soft));
  border: none;
  border-radius: 5px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 600;
  transition: background 0.16s ease, color 0.16s ease;
}

.theme-modes button:hover {
  color: var(--text-color);
  background: color-mix(in srgb, var(--surface-color) 72%, transparent);
}

.theme-modes button.active {
  background: color-mix(in srgb, var(--surface-color) 96%, transparent);
  color: var(--primary-color);
  box-shadow: 0 1px 2px color-mix(in srgb, var(--shadow-color) 18%, transparent);
}

.theme-modes button:focus-visible,
.color-option:focus-visible,
.upload-button:focus-visible,
.remove-bg-button:focus-visible,
.reset-button:focus-visible,
.color-text-input:focus-visible,
.color-input:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--primary-color) 40%, transparent);
  outline-offset: 2px;
}

/* 主题颜色 */
.color-section {
  grid-template-columns: minmax(180px, 0.34fr) minmax(360px, 0.66fr);
  gap: 18px 24px;
  align-items: start;
}

.color-controls {
  display: grid;
  gap: 14px;
  min-width: 0;
}

.color-options {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  min-width: 0;
}

.color-option {
  width: 30px;
  height: 30px;
  border-radius: 6px;
  cursor: pointer;
  border: 1px solid color-mix(in srgb, var(--text-color) 12%, transparent);
  transition: box-shadow 0.16s ease, border-color 0.16s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.color-option:hover {
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--text-color) 10%, transparent);
}

.color-option.active {
  border-color: color-mix(in srgb, var(--text-color) 76%, transparent);
  box-shadow:
    0 0 0 2px color-mix(in srgb, var(--surface-color) 96%, transparent),
    0 0 0 4px color-mix(in srgb, var(--text-color) 62%, transparent);
}

.checkmark {
  color: white;
  font-size: 14px;
  font-weight: bold;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
}

.custom-color-label {
  display: flex;
  align-items: center;
  align-self: stretch;
  min-height: 34px;
  color: color-mix(in srgb, var(--text-color) 88%, var(--text-soft));
  font-size: 13px;
  font-weight: 600;
}

.color-input-wrapper {
  display: grid;
  grid-template-columns: auto 44px minmax(140px, 1fr);
  gap: 8px;
  align-items: center;
  max-width: 420px;
}

.color-input {
  width: 42px;
  height: 34px;
  padding: 2px;
  border: 1px solid color-mix(in srgb, var(--border-color) 86%, transparent);
  border-radius: 7px;
  cursor: pointer;
  background-color: color-mix(in srgb, var(--surface-color) 94%, transparent);
}

.color-input::-webkit-color-swatch-wrapper {
  padding: 0;
}

.color-input::-webkit-color-swatch {
  border: none;
  border-radius: 5px;
}

.color-text-input {
  min-height: 34px;
  padding: 6px 10px;
  border: 1px solid color-mix(in srgb, var(--border-color) 86%, transparent);
  border-radius: 7px;
  background-color: color-mix(in srgb, var(--surface-color) 94%, transparent);
  color: var(--text-color);
  font-size: 13px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}

/* 背景设置 */
.background-section {
  gap: 16px;
}

.background-options {
  display: grid;
  gap: 12px;
}

.background-preview {
  width: 100%;
  min-height: 220px;
  max-height: 360px;
  aspect-ratio: 16 / 9;
  background-size: contain;
  background-position: center;
  background-repeat: no-repeat;
  border-radius: 7px;
  border: 1px solid color-mix(in srgb, var(--border-color) 76%, transparent);
  background-color: color-mix(in srgb, var(--surface-muted) 68%, transparent);
}

.background-preview.empty {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 136px;
  aspect-ratio: auto;
  color: color-mix(in srgb, var(--text-soft) 90%, var(--text-color));
  font-size: 13px;
}

.remove-bg-button {
  min-height: 34px;
  padding: 6px 12px;
  background-color: transparent;
  color: var(--error-color);
  border: 1px solid color-mix(in srgb, var(--error-color) 32%, transparent);
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}

.remove-bg-button:hover {
  background-color: color-mix(in srgb, var(--error-color) 8%, transparent);
}

.upload-section {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 10px;
  align-items: center;
}

.upload-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 34px;
  padding: 6px 12px;
  background-color: color-mix(in srgb, var(--surface-color) 72%, transparent);
  color: var(--text-color);
  border: 1px solid color-mix(in srgb, var(--border-color) 84%, transparent);
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.16s ease, border-color 0.16s ease;
}

.upload-button:hover {
  background-color: var(--surface-muted);
  border-color: color-mix(in srgb, var(--primary-color) 36%, var(--border-color));
}

.hidden-input {
  display: none;
}

.upload-hint {
  color: color-mix(in srgb, var(--text-soft) 92%, var(--text-color));
  font-size: 13px;
  margin: 0;
}

/* 重置按钮 */
.reset-button {
  justify-self: end;
  width: fit-content;
  min-height: 34px;
  padding: 6px 12px;
  background-color: color-mix(in srgb, var(--surface-color) 72%, transparent);
  color: var(--text-color);
  border: 1px solid color-mix(in srgb, var(--border-color) 84%, transparent);
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.16s ease, border-color 0.16s ease;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.reset-button:hover {
  background-color: var(--surface-muted);
  border-color: color-mix(in srgb, var(--primary-color) 36%, var(--border-color));
}

@media (max-width: 768px) {
  .preference-row,
  .color-section,
  .theme-actions {
    grid-template-columns: 1fr;
  }

  .theme-modes,
  .reset-button {
    justify-self: stretch;
  }

  .theme-modes {
    width: 100%;
    min-width: 0;
  }

  .color-input-wrapper {
    grid-template-columns: 1fr 44px minmax(120px, 1fr);
    max-width: none;
  }

  .background-preview {
    min-height: 180px;
  }
}
</style>

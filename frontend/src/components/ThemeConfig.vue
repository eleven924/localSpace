<template>
  <div class="theme-config">
    <section class="set-section">
      <div class="set-section-head">
        <div>
          <h3>主题模式</h3>
          <p>选择适合当前环境的界面明暗模式。</p>
        </div>
      </div>
      <div class="set-modes" role="group" aria-label="主题模式">
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

    <section class="set-section">
      <div class="set-section-head">
        <div>
          <h3>强调色</h3>
          <p>用于导航选中态、主要按钮和关键操作提示。</p>
        </div>
      </div>

      <div class="set-swatches" role="listbox" aria-label="预设强调色">
        <button
          v-for="color in presetColors"
          :key="color"
          type="button"
          class="set-swatch"
          :class="{ active: themeStore.primaryColor === color }"
          :style="{ backgroundColor: color }"
          :aria-label="`选择颜色 ${color}`"
          :aria-selected="themeStore.primaryColor === color"
          @click="handleSetColor(color)"
        >
          <span aria-hidden="true">✓</span>
        </button>
      </div>

      <div class="set-color-input">
        <label for="custom-color">自定义</label>
        <input
          id="custom-color"
          type="color"
          :value="themeStore.primaryColor"
          @input="handleCustomColorChange"
        />
        <input
          class="set-field mono"
          type="text"
          :value="themeStore.primaryColor"
          placeholder="#2196F3"
          aria-label="强调色十六进制值"
          @input="handleColorTextChange"
        />
      </div>
    </section>

    <section class="set-section">
      <div class="set-section-head">
        <div>
          <h3>背景图片</h3>
          <p>背景会在应用外层透出，内容区会保留稳定浅色遮罩以保证可读。</p>
        </div>
      </div>

      <div class="set-bg">
        <div
          class="set-bg-preview"
          :class="{ empty: !themeStore.backgroundImage }"
          :style="
            themeStore.backgroundImage
              ? { backgroundImage: `url(${themeStore.backgroundImage})` }
              : undefined
          "
          role="img"
          :aria-label="themeStore.backgroundImage ? '当前背景图片预览' : '未设置背景图片'"
        >
          <span v-if="!themeStore.backgroundImage">未设置背景图片</span>
        </div>

        <div>
          <div class="set-upload">
            <label class="btn secondary" for="background-upload">上传背景图片</label>
            <input id="background-upload" type="file" accept="image/*" @change="handleBackgroundUpload" />
            <button
              v-if="themeStore.backgroundImage"
              type="button"
              class="set-mini warn"
              @click="handleRemoveBackground"
            >
              移除背景
            </button>
          </div>
          <p v-if="uploadError" class="set-feedback error upload-note">{{ uploadError }}</p>
          <p v-else class="set-hint upload-note">支持 JPG、PNG、GIF、WebP，最大 5MB</p>
        </div>
      </div>
    </section>

    <section class="set-section">
      <div class="set-section-head">
        <div>
          <h3>恢复默认</h3>
          <p>重置主题模式、强调色和背景图片。</p>
        </div>
      </div>
      <button type="button" class="btn secondary set-danger-action" @click="showResetDialog = true">
        重置为默认主题
      </button>
    </section>

    <!-- 重置会一次性丢掉背景图片，那是这一分区里唯一找不回来的东西。 -->
    <div v-if="showResetDialog" class="set-modal" @click.self="showResetDialog = false">
      <div
        class="set-modal-panel danger"
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="theme-reset-title"
      >
        <header class="set-modal-head">
          <span class="set-modal-mark" aria-hidden="true">!</span>
          <div>
            <h3 id="theme-reset-title">重置为默认主题</h3>
            <p>主题模式、强调色和背景图片都会回到出厂状态。</p>
          </div>
        </header>
        <div class="set-modal-body">
          <p class="set-modal-warning">
            <b>背景图片会被清除，且无法撤销。</b>
            如果这张图还要用，请先在本地留一份副本。
          </p>
        </div>
        <footer class="set-modal-foot">
          <button type="button" class="btn secondary" @click="showResetDialog = false">取消</button>
          <button type="button" class="btn secondary set-danger-action" @click="handleResetTheme">
            重置主题
          </button>
        </footer>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useThemeStore } from '@/store/modules/theme'

const emit = defineEmits<{
  themeChanged: []
  themeReset: []
}>()

const themeStore = useThemeStore()

const uploadError = ref('')
const showResetDialog = ref(false)

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

  uploadError.value = ''

  if (!file) return

  // 上传前校验图片类型，避免把非图片内容写入主题配置。
  if (!file.type.match(/^image\/(jpeg|png|gif|webp)$/)) {
    uploadError.value = '请选择有效的图片文件（JPG、PNG、GIF、WebP）'
    target.value = ''
    return
  }

  // 背景图存为 Data URL，限制体积可以避免配置过大影响启动速度。
  if (file.size > 5 * 1024 * 1024) {
    uploadError.value = '图片文件不能超过 5MB'
    target.value = ''
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
      uploadError.value = '读取图片文件失败'
    }
    reader.readAsDataURL(file)
  } catch (err) {
    console.error('Failed to upload background:', err)
    uploadError.value = '上传背景图片失败'
  }

  // 清空 input 以允许重复选择同一文件
  target.value = ''
}

// 移除背景是一步就能撤回的操作（重新上传即可），不值得一个确认框。
const handleRemoveBackground = () => {
  themeStore.setBackgroundImage('')
  uploadError.value = ''
  emit('themeChanged')
}

const handleResetTheme = () => {
  themeStore.resetTheme()
  showResetDialog.value = false
  emit('themeReset')
}
</script>

<style scoped>
.theme-config {
  width: 100%;
}

.upload-note {
  margin-top: 10px;
}
</style>

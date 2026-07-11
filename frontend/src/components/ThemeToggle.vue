<template>
  <div class="theme-toggle">
    <button
      class="theme-mode-button"
      @click="handleToggleTheme"
      :title="`切换到${themeStore.themeMode === 'light' ? '深色' : '浅色'}主题`"
    >
      <span class="theme-icon">{{ themeIcon }}</span>
    </button>

    <div class="color-picker-wrapper">
      <label for="color-picker" class="color-label" title="主题颜色">🎨</label>
      <input
        id="color-picker"
        type="color"
        :value="themeStore.primaryColor"
        @input="handleColorChange"
        class="color-picker"
        title="选择主题颜色"
      />
    </div>

    <button
      class="theme-reset-button"
      @click="handleResetTheme"
      title="重置主题"
    >
      🔄
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useThemeStore } from '@/store/modules/theme'

const themeStore = useThemeStore()

const themeIcon = computed(() => {
  return themeStore.themeMode === 'light' ? '🌙' : '☀️'
})

const handleToggleTheme = () => {
  themeStore.toggleTheme()
}

const handleColorChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  themeStore.setPrimaryColor(target.value)
}

const handleResetTheme = () => {
  if (confirm('确定要重置主题设置吗？')) {
    themeStore.resetTheme()
  }
}
</script>

<style scoped>
.theme-toggle {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  background-color: var(--surface-color);
  border: 1px solid var(--border-color);
  border-radius: 24px;
}

.theme-mode-button {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background-color: var(--bg-color);
  border: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
}

.theme-mode-button:hover {
  background-color: var(--border-color);
  transform: scale(1.1);
}

.theme-mode-button:active {
  transform: scale(0.95);
}

.theme-icon {
  font-size: 18px;
}

.color-picker-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  gap: 4px;
}

.color-label {
  font-size: 16px;
  cursor: pointer;
  transition: transform 0.2s ease;
}

.color-label:hover {
  transform: scale(1.1);
}

.color-picker {
  width: 36px;
  height: 36px;
  padding: 0;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  background: none;
  overflow: hidden;
  -webkit-appearance: none;
}

.color-picker::-webkit-color-swatch-wrapper {
  padding: 0;
}

.color-picker::-webkit-color-swatch {
  border: 1px solid var(--border-color);
  border-radius: 50%;
}

.color-picker::-moz-color-swatch {
  border: 1px solid var(--border-color);
  border-radius: 50%;
}

.color-picker:hover {
  transform: scale(1.1);
  transition: transform 0.2s ease;
}

.theme-reset-button {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background-color: var(--bg-color);
  border: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s ease;
}

.theme-reset-button:hover {
  background-color: var(--border-color);
  transform: rotate(180deg);
}

.theme-reset-button:active {
  transform: rotate(180deg) scale(0.95);
}

@media (max-width: 768px) {
  .theme-toggle {
    gap: 8px;
    padding: 6px 12px;
  }

  .theme-mode-button {
    width: 32px;
    height: 32px;
  }

  .theme-icon {
    font-size: 16px;
  }

  .color-picker {
    width: 32px;
    height: 32px;
  }

  .theme-reset-button {
    width: 28px;
    height: 28px;
    font-size: 12px;
  }
}

@media (max-width: 480px) {
  .theme-toggle {
    gap: 6px;
    padding: 4px 8px;
  }

  .theme-mode-button {
    width: 28px;
    height: 28px;
  }

  .theme-icon {
    font-size: 14px;
  }

  .color-picker {
    width: 28px;
    height: 28px;
  }

  .color-label {
    font-size: 14px;
  }

  .theme-reset-button {
    width: 24px;
    height: 24px;
    font-size: 11px;
  }
}
</style>
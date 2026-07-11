import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api/index'
import type { ThemeConfig } from '@/types'

export const useThemeStore = defineStore('theme', () => {
  const themeMode = ref<'light' | 'dark'>('light')
  const primaryColor = ref('#2196F3')
  const backgroundImage = ref('')
  const loading = ref(false)
  const error = ref<string | null>(null)

  const loadThemeFromBackend = async () => {
    loading.value = true
    error.value = null

    try {
      const config: ThemeConfig = await api.theme.getConfig()
      if (config) {
        themeMode.value = config.themeMode || 'light'
        primaryColor.value = config.primaryColor || '#2196F3'
        backgroundImage.value = config.backgroundImage || ''

        // Apply theme to document
        document.documentElement.setAttribute('data-theme', themeMode.value)
        document.documentElement.style.setProperty('--primary-color', primaryColor.value)
        if (backgroundImage.value) {
          document.documentElement.style.setProperty('--background-image', `url(${backgroundImage.value})`)
          document.documentElement.style.setProperty('--bg-color', 'transparent')
        }
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : '加载主题配置失败'
      console.error('Failed to load theme config:', err)
    } finally {
      loading.value = false
    }
  }

  const setThemeMode = (mode: 'light' | 'dark') => {
    themeMode.value = mode
    document.documentElement.setAttribute('data-theme', mode)
    saveThemeToBackend()
  }

  const setPrimaryColor = (color: string) => {
    primaryColor.value = color
    document.documentElement.style.setProperty('--primary-color', color)
    saveThemeToBackend()
  }

  const setBackgroundImage = (image: string) => {
    backgroundImage.value = image
    document.documentElement.style.setProperty('--background-image', image)
    if (image) {
      document.documentElement.style.setProperty('--bg-color', 'transparent')
    }
    saveThemeToBackend()
  }

  const saveThemeToBackend = async () => {
    try {
      const config: ThemeConfig = {
        id: 1,
        themeMode: themeMode.value,
        primaryColor: primaryColor.value,
        backgroundImage: backgroundImage.value,
      }
      await api.theme.updateConfig(config)
    } catch (err) {
      console.error('Failed to save theme config:', err)
      // Don't show error to user, theme changes still work locally
    }
  }

  const toggleTheme = () => {
    const newMode = themeMode.value === 'light' ? 'dark' : 'light'
    setThemeMode(newMode)
  }

  return {
    themeMode,
    primaryColor,
    backgroundImage,
    loading,
    error,
    loadThemeFromBackend,
    setThemeMode,
    setPrimaryColor,
    setBackgroundImage,
    toggleTheme,
    saveThemeToBackend
  }
})
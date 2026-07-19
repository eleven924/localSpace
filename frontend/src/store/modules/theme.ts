import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api/index'
import type { ThemeConfig } from '@/types'

const DEFAULT_THEME_MODE = 'light'
const DEFAULT_PRIMARY_COLOR = '#2196F3'

export const useThemeStore = defineStore('theme', () => {
  const themeMode = ref<'light' | 'dark'>(DEFAULT_THEME_MODE)
  const primaryColor = ref(DEFAULT_PRIMARY_COLOR)
  const backgroundImage = ref('')
  const loading = ref(false)
  const error = ref<string | null>(null)

  const applyThemeToDocument = () => {
    document.documentElement.setAttribute('data-theme', themeMode.value)
    document.documentElement.style.setProperty('--primary-color', primaryColor.value)

    if (backgroundImage.value) {
      document.documentElement.style.setProperty(
        '--background-image',
        `url(${JSON.stringify(backgroundImage.value)})`
      )
      document.documentElement.style.setProperty('--app-bg-color', 'transparent')
      return
    }

    document.documentElement.style.removeProperty('--background-image')
    document.documentElement.style.removeProperty('--app-bg-color')
  }

  const loadThemeFromBackend = async () => {
    loading.value = true
    error.value = null

    try {
      const config: ThemeConfig = await api.theme.getConfig()
      if (config) {
        themeMode.value = config.themeMode || DEFAULT_THEME_MODE
        primaryColor.value = config.primaryColor || DEFAULT_PRIMARY_COLOR
        backgroundImage.value = config.backgroundImage || ''
      }
      applyThemeToDocument()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load theme config'
      console.error('Failed to load theme config:', err)
    } finally {
      loading.value = false
    }
  }

  const setThemeMode = (mode: 'light' | 'dark') => {
    themeMode.value = mode
    applyThemeToDocument()
    saveThemeToBackend()
  }

  const setPrimaryColor = (color: string) => {
    primaryColor.value = color
    applyThemeToDocument()
    saveThemeToBackend()
  }

  const setBackgroundImage = (image: string) => {
    backgroundImage.value = image
    applyThemeToDocument()
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
    }
  }

  const toggleTheme = () => {
    const newMode = themeMode.value === 'light' ? 'dark' : 'light'
    setThemeMode(newMode)
  }

  const resetTheme = () => {
    themeMode.value = DEFAULT_THEME_MODE
    primaryColor.value = DEFAULT_PRIMARY_COLOR
    backgroundImage.value = ''
    applyThemeToDocument()
    saveThemeToBackend()
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
    resetTheme,
    saveThemeToBackend
  }
})

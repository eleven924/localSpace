<template>
  <div id="app" :class="themeStore.themeMode" :style="appStyle">
    <RouterView />
  </div>
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router'
import { computed, onMounted } from 'vue'
import { useThemeStore } from './store/modules/theme'

const themeStore = useThemeStore()

const appStyle = computed(() => ({
  backgroundImage: themeStore.backgroundImage ? `url(${themeStore.backgroundImage})` : 'none',
  backgroundSize: 'cover',
  backgroundPosition: 'center',
  backgroundRepeat: 'no-repeat',
}))

onMounted(() => {
  // Load theme configuration from backend
  themeStore.loadThemeFromBackend()
})
</script>

<style scoped>
#app {
  position: relative;
  width: 100%;
  height: 100vh;
  overflow: hidden;
  background-color: var(--app-bg-color, var(--bg-color));
}
</style>
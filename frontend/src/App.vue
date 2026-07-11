<template>
  <div id="app" :class="themeStore.themeMode">
    <RouterView />
  </div>
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router'
import { onMounted } from 'vue'
import { useThemeStore } from './store/modules/theme'

const themeStore = useThemeStore()

onMounted(() => {
  // Load theme configuration from backend
  themeStore.loadThemeFromBackend()
})
</script>

<style scoped>
#app {
  width: 100%;
  height: 100vh;
  overflow: hidden;
}

/* Dynamic background image support */
#app[data-background-image] {
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
}
</style>
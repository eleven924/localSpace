<template>
  <div id="app" :class="themeStore.themeMode" :style="appStyle">
    <RouterView />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterView } from 'vue-router'
import { useThemeStore } from './store/modules/theme'

const themeStore = useThemeStore()

const appStyle = computed(() => ({
  backgroundImage: themeStore.backgroundImage ? `url(${themeStore.backgroundImage})` : 'none',
  backgroundSize: 'cover',
  backgroundPosition: 'center',
  backgroundRepeat: 'no-repeat',
}))
</script>

<style scoped>
#app {
  width: 100%;
  height: 100%;
  background-color: var(--app-bg-color, var(--bg-color));
}
</style>

<template>
  <div class="batch-overlay">
    <div class="overlay-card">
      <h3>正在移动 {{ total }} 个文件到「{{ collectionName || '未分配合集' }}」</h3>
      <progress :max="total" :value="completed"></progress>
      <p>{{ message }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  total: number
  completed: number
  message: string
  collectionName: string
}>()
</script>

<style scoped>
.batch-overlay {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 3000;
  background: rgba(15, 23, 42, 0.42);
  backdrop-filter: blur(10px);
}

.overlay-card {
  background: var(--surface-color);
  border-radius: 16px;
  padding: 32px;
  min-width: 320px;
}

.overlay-card h3 {
  margin: 0 0 18px;
  font-size: 16px;
  color: var(--text-color);
}

.overlay-card progress {
  width: 100%;
  height: 8px;
  border-radius: 999px;
  margin-bottom: 14px;
}

.overlay-card p {
  margin: 0;
  font-size: 13px;
  color: var(--text-soft);
}
</style>

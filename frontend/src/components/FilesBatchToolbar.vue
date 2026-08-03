<template>
  <div class="batch-toolbar">
    <span>已选择 {{ selectedIds.length }} 个文件</span>
    <button class="btn secondary" @click="$emit('move')">移动合集</button>
    <button class="btn danger" @click="$emit('delete')">删除</button>
    <button class="btn ghost" @click="$emit('clear')">取消选择</button>
  </div>
</template>

<script setup lang="ts">
defineProps<{ selectedIds: number[] }>()
defineEmits(['move', 'delete', 'clear'])
</script>

<style scoped>
.batch-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.64);
  border: 1px solid rgba(146, 165, 192, 0.18);
}

.batch-toolbar span {
  flex: 1;
  font-size: 14px;
  color: var(--text-color);
}

.batch-toolbar .btn {
  width: auto;
}

.batch-toolbar .btn.danger {
  background: rgba(204, 108, 108, 0.12);
  color: #a85d5d;
  border-color: rgba(204, 108, 108, 0.2);
}
</style>

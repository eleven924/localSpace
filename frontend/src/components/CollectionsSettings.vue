<template>
  <div class="collections-settings">
    <div class="section-block">
      <h5>新增合集</h5>
      <div class="add-row">
        <input v-model="newName" type="text" placeholder="合集名称" />
        <button class="btn primary" :disabled="!newName.trim() || adding" @click="addCollection">添加</button>
      </div>
      <p v-if="error" class="feedback error">{{ error }}</p>
    </div>

    <div class="section-block">
      <h5>已有合集</h5>
      <div v-if="collections.length === 0" class="empty-state">暂无合集</div>
      <div v-else class="collection-list">
        <div v-for="collection in collections" :key="collection.id" class="collection-row">
          <span>{{ collection.name }}</span>
          <span class="count">引用 {{ fileCountMap[collection.id] ?? 0 }}</span>
          <button
            class="btn ghost"
            :disabled="(fileCountMap[collection.id] ?? 0) > 0"
            :title="(fileCountMap[collection.id] ?? 0) > 0 ? '该合集正在被文件使用，无法删除' : '删除合集'"
            @click="removeCollection(collection.id)"
          >删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '@/api'
import type { Collection, File } from '@/types'

const props = defineProps<{
  files: File[]
}>()

const collections = ref<Collection[]>([])
const newName = ref('')
const adding = ref(false)
const error = ref('')

const fileCountMap = computed(() => {
  const map: Record<number, number> = {}
  props.files.forEach((file) => {
    if (file.collectionId) {
      map[file.collectionId] = (map[file.collectionId] || 0) + 1
    }
  })
  return map
})

const load = async () => {
  collections.value = await api.collection.getAll()
}

const addCollection = async () => {
  adding.value = true
  error.value = ''
  try {
    await api.collection.add(newName.value.trim())
    newName.value = ''
    await load()
  } catch (err: any) {
    error.value = err?.message || '添加失败'
  } finally {
    adding.value = false
  }
}

const removeCollection = async (id: number) => {
  if (!window.confirm('确定删除该合集？')) return
  try {
    await api.collection.remove(id)
    await load()
  } catch (err: any) {
    window.alert(err?.message || '删除失败')
  }
}

onMounted(load)
</script>

<style scoped>
.collections-settings {
  display: grid;
  gap: 24px;
}

.section-block h5 {
  margin: 0 0 14px;
  font-size: 15px;
  color: var(--text-color);
}

.add-row {
  display: flex;
  gap: 10px;
}

.add-row input {
  flex: 1;
}

.collection-list {
  display: grid;
  gap: 8px;
}

.collection-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.58);
  border: 1px solid rgba(146, 165, 192, 0.16);
}

.collection-row span:first-child {
  flex: 1;
  font-size: 14px;
  color: var(--text-color);
}

.collection-row .count {
  font-size: 12px;
  color: var(--text-soft);
}

.empty-state {
  color: var(--text-soft);
  font-size: 13px;
  padding: 12px 0;
}

.feedback.error {
  margin-top: 10px;
}
</style>

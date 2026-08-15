<template>
  <div class="collections-settings">
    <section class="set-section">
      <div class="set-section-head">
        <div>
          <h3>新增合集</h3>
          <p>合集用来把同一件事的文件收在一起，导入时可以直接选中。</p>
        </div>
      </div>

      <div class="set-add-row">
        <input
          v-model="newName"
          class="set-field"
          type="text"
          placeholder="合集名称"
          aria-label="合集名称"
          @keyup.enter="addCollection"
        />
        <button
          type="button"
          class="btn primary"
          :disabled="!newName.trim() || adding"
          @click="addCollection"
        >
          {{ adding ? '添加中...' : '添加' }}
        </button>
      </div>
      <p v-if="feedback" class="set-feedback add-feedback" :class="feedback.ok ? 'success' : 'error'">
        {{ feedback.text }}
      </p>
    </section>

    <section class="set-section">
      <div class="set-section-head">
        <div>
          <h3>已有合集</h3>
          <p>被文件引用的合集不能删除，先把文件移出合集再来。</p>
        </div>
      </div>

      <div v-if="collections.length > 0" class="set-collections">
        <div v-for="collection in collections" :key="collection.id" class="set-collection">
          <strong>{{ collection.name }}</strong>
          <span class="set-count">引用 {{ referenceCount(collection.id) }}</span>
          <button
            type="button"
            class="set-mini warn"
            :disabled="referenceCount(collection.id) > 0"
            :title="
              referenceCount(collection.id) > 0 ? '该合集正在被文件使用，无法删除' : '删除合集'
            "
            @click="removeCollection(collection.id)"
          >
            删除
          </button>
        </div>
      </div>

      <div v-else class="set-empty">
        <strong>暂无合集</strong>
        <p>用上面的输入框建一个，导入时就能直接归到这个合集里。</p>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from '@/api'
import type { Collection, CollectionFilterCounts } from '@/types'

const collections = ref<Collection[]>([])
const newName = ref('')
const adding = ref(false)
const feedback = ref<{ text: string; ok: boolean } | null>(null)

// 引用数直接问后端要。之前它是按当前列表页里的文件数出来的，
// 而设置页从不加载文件列表，于是计数恒为 0，「被引用不能删除」这条保护形同虚设。
const counts = ref<CollectionFilterCounts>({ total: 0, unsorted: 0, collections: {} })

const referenceCount = (id: number): number => counts.value.collections[id] ?? 0

const load = async () => {
  const [result, filterCounts] = await Promise.all([
    api.collection.getAll(),
    api.collection.getFilterCounts('all'),
  ])
  collections.value = result || []
  counts.value = filterCounts || { total: 0, unsorted: 0, collections: {} }
}

const addCollection = async () => {
  const name = newName.value.trim()
  if (!name || adding.value) return

  adding.value = true
  feedback.value = null
  try {
    await api.collection.add(name)
    newName.value = ''
    await load()
    feedback.value = { text: `已创建合集「${name}」`, ok: true }
  } catch (err: any) {
    feedback.value = { text: err?.message || '添加失败', ok: false }
  } finally {
    adding.value = false
  }
}

const removeCollection = async (id: number) => {
  // 删除按钮在有引用时是禁用的，走到这里的合集不含文件，一步就能重建，不再拦一次。
  try {
    await api.collection.remove(id)
    await load()
    feedback.value = { text: '合集已删除', ok: true }
  } catch (err: any) {
    feedback.value = { text: err?.message || '删除失败', ok: false }
  }
}

onMounted(load)
</script>

<style scoped>
.collections-settings {
  width: 100%;
}

.add-feedback {
  margin-top: 12px;
}
</style>

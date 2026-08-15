<template>
  <div class="page-shell import-view">
    <AppHeader />

    <div class="page-content">
      <div class="page-stack">
        <!-- 侧栏和面包屑已经写明当前是导入页，标题不再重复；模式切换直接充当页头。 -->
        <header class="imp-header">
          <h1 class="visually-hidden">导入</h1>
          <div class="imp-tabs" role="tablist" aria-label="导入模式">
            <button
              id="import-tab-single"
              type="button"
              role="tab"
              class="imp-tab"
              :class="{ active: mode === 'single' }"
              :aria-selected="mode === 'single'"
              @click="mode = 'single'"
            >
              单文件导入
            </button>
            <button
              id="import-tab-batch"
              type="button"
              role="tab"
              class="imp-tab"
              :class="{ active: mode === 'batch' }"
              :aria-selected="mode === 'batch'"
              @click="mode = 'batch'"
            >
              批量导入任务
            </button>
          </div>
          <p class="imp-lede">
            {{
              mode === 'single'
                ? '一次导入一个文件，顺手把它的名称、标签和描述补齐。'
                : '一次选好多个文件，设一组公共字段，交给后台任务处理。'
            }}
          </p>
        </header>

        <!-- 单文件只有一个主体，所以它是横跨整宽的来源条 -->
        <section
          v-if="mode === 'single'"
          class="import-layout single"
          :class="{ solo: !singleFile || !aiAvailable }"
          role="tabpanel"
          aria-labelledby="import-tab-single"
        >
          <div class="imp-source">
            <button class="imp-pick" type="button" @click="handleSelectSingleFile">
              <span class="imp-pick-tile" aria-hidden="true">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M3 7.5A1.5 1.5 0 0 1 4.5 6h4l2 2.4h7A1.5 1.5 0 0 1 19 9.9v1.1" />
                  <path d="M3.6 12.4h17.2l-1.9 6.1a1.5 1.5 0 0 1-1.4 1H4.6a1.5 1.5 0 0 1-1.5-1.5z" />
                </svg>
              </span>
              <span class="imp-pick-copy">
                <strong>{{ singleFile ? '重新选择文件' : '选择文件' }}</strong>
                <span>打开系统文件窗口，一次选择一个文件</span>
              </span>
              <i class="imp-pick-chevron" aria-hidden="true">›</i>
            </button>

            <article v-if="singleFile" class="imp-file imp-file-lead">
              <span class="imp-type" :class="singleFile.type" aria-hidden="true">{{ singleFile.shortLabel }}</span>
              <div class="imp-file-copy">
                <strong>{{ singleFile.name }}</strong>
                <p>{{ singleFile.typeLabel }} · {{ formatFileSize(singleFile.size) }}</p>
              </div>
              <button class="imp-remove" type="button" @click="clearSingleFile">移除</button>
            </article>

            <div v-else class="imp-empty imp-empty-lead">
              <strong>还没有选择文件</strong>
              <p>选好之后，这份资料的名称、标签和描述会在下面展开。</p>
            </div>
          </div>

          <section class="imp-panel">
            <div class="imp-head">
              <div>
                <h2>元信息</h2>
                <p>补全后这份资料才能被搜到。</p>
              </div>
            </div>

            <div v-if="singleFile" class="imp-form">
              <div class="imp-grid">
                <label class="imp-field">
                  <span>文件名</span>
                  <input v-model="singleForm.fileName" type="text" placeholder="输入文件名" />
                </label>

                <label class="imp-field">
                  <span>关键词<em>供 AI 参考</em></span>
                  <input v-model="singleForm.keywords" type="text" placeholder="例如：课程、会议、剪辑" />
                </label>

                <label class="imp-field wide">
                  <span>合集 / 系列</span>
                  <CollectionSelector v-model="singleForm.collectionId" :collections="collections" />
                </label>
              </div>

              <FileMetadataFields
                :file-name="singleForm.fileName"
                :file-type="singleFile.type"
                :user-keywords="singleForm.keywords"
                :embed-ai="false"
                v-model:modelValueTags="singleForm.tags"
                v-model:modelValueDescription="singleForm.description"
              />
            </div>

            <div v-else class="imp-locked">
              <div>
                <strong>先选择一个文件</strong>
                <p>选中后，这里会展开文件名、关键词、合集、标签和描述。</p>
              </div>
            </div>
          </section>

          <section v-if="singleFile" v-show="aiAvailable" class="imp-panel">
            <FileAISuggestions
              :file-name="singleForm.fileName"
              :file-type="singleFile.type"
              :user-keywords="singleForm.keywords"
              :tags="singleForm.tags"
              :description="singleForm.description"
              @update:tags="singleForm.tags = $event"
              @update:description="singleForm.description = $event"
              @availability="aiAvailable = $event"
            />
          </section>

          <div class="imp-commit">
            <p v-if="singleError" class="imp-feedback error">{{ singleError }}</p>
            <p v-else-if="singleSuccess" class="imp-feedback success">{{ singleSuccess }}</p>
            <button class="btn secondary" type="button" @click="resetSingleForm">重置表单</button>
            <button
              class="btn primary"
              type="button"
              :disabled="singleSubmitting || !singleCanSubmit"
              @click="submitSingleImport"
            >
              {{ singleSubmitting ? '正在导入...' : '立即导入文件' }}
            </button>
          </div>
        </section>

        <!-- 批量是一份清单，所以它占满左栏 -->
        <section v-else class="import-layout batch" role="tabpanel" aria-labelledby="import-tab-batch">
          <section class="imp-panel">
            <div class="imp-head">
              <div>
                <h2>文件</h2>
                <p>可以分多次选择，重复的文件会自动跳过。</p>
              </div>
              <span class="imp-pill" :class="{ on: selectedFiles.length > 0 }">
                {{ selectedFiles.length ? `${selectedFiles.length} 个文件` : '还没有文件' }}
              </span>
            </div>

            <button class="imp-pick" type="button" @click="handleSelectFiles">
              <span class="imp-pick-tile" aria-hidden="true">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M3 7.5A1.5 1.5 0 0 1 4.5 6h4l2 2.4h7A1.5 1.5 0 0 1 19 9.9v1.1" />
                  <path d="M3.6 12.4h17.2l-1.9 6.1a1.5 1.5 0 0 1-1.4 1H4.6a1.5 1.5 0 0 1-1.5-1.5z" />
                </svg>
              </span>
              <span class="imp-pick-copy">
                <strong>添加文件</strong>
                <span>打开系统文件窗口，按住 Ctrl / Shift 可一次选多个</span>
              </span>
              <i class="imp-pick-chevron" aria-hidden="true">›</i>
            </button>

            <template v-if="selectedFiles.length > 0">
              <div class="imp-stats">
                <div class="imp-stat">
                  <strong>{{ selectedFiles.length }}</strong>
                  <span>文件数量</span>
                </div>
                <div class="imp-stat">
                  <strong>{{ totalSizeLabel }}</strong>
                  <span>预计体量</span>
                </div>
              </div>

              <div class="imp-list-head">
                <span>已选文件</span>
                <button class="imp-text-button" type="button" :disabled="batchSubmitting" @click="clearFiles">
                  清空列表
                </button>
              </div>

              <div class="imp-tray scroll-soft">
                <article v-for="file in selectedFiles" :key="file.path" class="imp-file">
                  <span class="imp-type" :class="file.type" aria-hidden="true">{{ file.shortLabel }}</span>
                  <div class="imp-file-copy">
                    <strong>{{ file.name }}</strong>
                    <p>{{ file.typeLabel }} · {{ formatFileSize(file.size) }}</p>
                  </div>
                  <button class="imp-remove" type="button" @click="removeFile(file.path)">移除</button>
                </article>
              </div>
            </template>

            <div v-else class="imp-empty">
              <strong>还没有加入文件</strong>
              <p>用上面的按钮选几个文件，再统一设置标签、合集和描述。</p>
            </div>
          </section>

          <section class="imp-panel">
            <div class="imp-head">
              <div>
                <h2>元信息</h2>
                <p>这里填的内容会写入这一批的每个文件。</p>
              </div>
              <span class="imp-pill">公共字段</span>
            </div>

            <div class="imp-form">
              <div class="imp-grid">
                <label class="imp-field wide">
                  <span>统一标签<em>用逗号分隔</em></span>
                  <input v-model="batchForm.sharedTagsText" type="text" placeholder="例如：课程、待整理、项目A" />
                </label>

                <label class="imp-field wide">
                  <span>合集 / 系列</span>
                  <CollectionSelector v-model="batchForm.collectionId" :collections="collections" />
                </label>

                <label class="imp-field wide">
                  <span>统一描述<em>可选</em></span>
                  <textarea
                    v-model="batchForm.sharedDescription"
                    rows="4"
                    placeholder="为这一批文件补充统一说明"
                  />
                </label>
              </div>

              <div v-if="aiAvailable" class="imp-toggles">
                <label class="imp-toggle">
                  <input v-model="batchForm.enableAIGeneratedTags" type="checkbox" />
                  <div>
                    <strong>AI 追加标签</strong>
                    <p>在统一标签之外，为每个文件再补几个更具体的标签。</p>
                  </div>
                </label>

                <label class="imp-toggle">
                  <input v-model="batchForm.enableAIGeneratedDescription" type="checkbox" />
                  <div>
                    <strong>AI 生成描述</strong>
                    <p>优先为每个文件单独生成描述，失败时退回统一描述。</p>
                  </div>
                </label>
              </div>
            </div>
          </section>

          <div class="imp-commit">
            <p v-if="batchError" class="imp-feedback error">{{ batchError }}</p>
            <p v-else-if="batchSuccess" class="imp-feedback success">{{ batchSuccess }}</p>
            <button class="btn secondary" type="button" :disabled="batchSubmitting" @click="resetBatchForm">
              重置设置
            </button>
            <button
              class="btn primary"
              type="button"
              :disabled="!batchCanSubmit || batchSubmitting"
              @click="submitBatchJob"
            >
              {{ batchSubmitting ? '正在创建任务...' : '创建后台导入任务' }}
            </button>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { api } from '@/api'
import AppHeader from '@/components/AppHeader.vue'
import CollectionSelector from '@/components/CollectionSelector.vue'
import FileAISuggestions from '@/components/FileAISuggestions.vue'
import FileMetadataFields from '@/components/FileMetadataFields.vue'
import type { BatchImportJobRequest, SelectedFile, SingleImportJobRequest } from '@/types/jobs'
import type { Collection } from '@/types'
import { FILE_TYPE_META, formatFileSize } from '@/utils/constants'
import { useJobsStore } from '@/store/modules/jobs'
import { computed, onMounted, reactive, ref } from 'vue'

type ImportMode = 'single' | 'batch'

interface RichSelectedFile extends SelectedFile {
  type: string
  typeLabel: string
  shortLabel: string
}

const EXTENSION_TO_TYPE: Record<string, string> = {
  '.mp4': 'video',
  '.avi': 'video',
  '.mkv': 'video',
  '.mov': 'video',
  '.wmv': 'video',
  '.pdf': 'document',
  '.doc': 'document',
  '.docx': 'document',
  '.xls': 'document',
  '.xlsx': 'document',
  '.ppt': 'document',
  '.pptx': 'document',
  '.txt': 'document',
  '.md': 'document',
  '.mp3': 'music',
  '.wav': 'music',
  '.flac': 'music',
  '.zip': 'archive',
  '.rar': 'archive',
  '.7z': 'archive',
  '.exe': 'installer',
  '.msi': 'installer',
  '.jpg': 'image',
  '.jpeg': 'image',
  '.png': 'image',
  '.gif': 'image',
  '.webp': 'image',
}

const jobsStore = useJobsStore()

const collections = ref<Collection[]>([])
onMounted(async () => {
  collections.value = await api.collection.getAll()
  await loadAIAvailability()
})

const mode = ref<ImportMode>('single')
const singleFile = ref<RichSelectedFile | null>(null)
const selectedFiles = ref<RichSelectedFile[]>([])
// AI 未配置时 FileAISuggestions 不渲染内容，单文件模式收成一栏。
const aiAvailable = ref(false)
const singleSubmitting = ref(false)
const batchSubmitting = ref(false)
const singleError = ref('')
const singleSuccess = ref('')
const batchError = ref('')
const batchSuccess = ref('')

const singleForm = reactive({
  fileName: '',
  keywords: '',
  collectionId: undefined as number | undefined,
  tags: [] as string[],
  description: '',
})

const batchForm = reactive({
  sharedTagsText: '',
  sharedDescription: '',
  collectionId: undefined as number | undefined,
  enableAIGeneratedTags: false,
  enableAIGeneratedDescription: false,
})

const totalSize = computed(() => selectedFiles.value.reduce((sum, file) => sum + file.size, 0))
const totalSizeLabel = computed(() => formatFileSize(totalSize.value))
const singleCanSubmit = computed(() => !!singleFile.value && singleForm.fileName.trim().length > 0)
const batchCanSubmit = computed(() => selectedFiles.value.length > 0)

const loadAIAvailability = async () => {
  try {
    const config = await api.ai.getConfig()
    aiAvailable.value = Boolean(config?.enabled)

    // 未配置 AI 时不仅隐藏能力入口，提交参数也必须保持关闭，避免后台任务误启用 AI。
    if (!aiAvailable.value) {
      batchForm.enableAIGeneratedTags = false
      batchForm.enableAIGeneratedDescription = false
    } else {
      batchForm.enableAIGeneratedTags = true
    }
  } catch (error) {
    console.error('Failed to load AI config for import:', error)
    aiAvailable.value = false
    batchForm.enableAIGeneratedTags = false
    batchForm.enableAIGeneratedDescription = false
  }
}

const toRichFile = async (file: SelectedFile): Promise<RichSelectedFile> => {
  const extension = `.${file.name.split('.').pop()?.toLowerCase() || ''}`
  const parsedType = await api.fileType.parse(extension)
  const fileType = parsedType?.name || EXTENSION_TO_TYPE[extension] || 'other'
  const fileMeta = FILE_TYPE_META[fileType] || FILE_TYPE_META.other

  return {
    ...file,
    type: fileType,
    typeLabel: fileMeta.label,
    shortLabel: fileMeta.shortLabel,
  }
}

const handleSelectSingleFile = async () => {
  singleError.value = ''
  singleSuccess.value = ''

  // 先把文件的类型和大小补齐，右侧表单才能直接拿到完整上下文。
  const path = await api.system.selectFile()
  if (!path || path === 'success') return

  const fileName = path.split(/[/\\]/).pop() || ''
  const extension = `.${fileName.split('.').pop()?.toLowerCase() || ''}`
  const parsedType = await api.fileType.parse(extension)
  const fileType = parsedType?.name || EXTENSION_TO_TYPE[extension] || 'other'

  const metadata = await api.system.getMetadata(path, fileType)
  const file = await toRichFile({
    name: fileName,
    path,
    size: metadata?.size || 0,
  })

  singleFile.value = file
  resetSingleForm()
}

const handleSelectFiles = async () => {
  batchError.value = ''
  batchSuccess.value = ''

  const files = await api.system.selectFiles()
  const seen = new Set(selectedFiles.value.map((file) => file.path))

  // 批量模式只追加未选中的文件，避免重复条目把任务列表撑乱。
  for (const file of files as SelectedFile[]) {
    if (seen.has(file.path)) continue
    selectedFiles.value.push(await toRichFile(file))
    seen.add(file.path)
  }
}

const clearSingleFile = () => {
  singleFile.value = null
  resetSingleForm()
}

const removeFile = (path: string) => {
  selectedFiles.value = selectedFiles.value.filter((file) => file.path !== path)
}

const clearFiles = () => {
  selectedFiles.value = []
  batchError.value = ''
  batchSuccess.value = ''
}

const resetSingleForm = () => {
  singleForm.fileName = singleFile.value?.name || ''
  singleForm.keywords = ''
  singleForm.collectionId = undefined
  singleForm.tags = []
  singleForm.description = ''
  singleError.value = ''
  singleSuccess.value = ''
}

const resetBatchForm = () => {
  batchForm.sharedTagsText = ''
  batchForm.sharedDescription = ''
  batchForm.collectionId = undefined
  batchForm.enableAIGeneratedTags = aiAvailable.value
  batchForm.enableAIGeneratedDescription = false
  batchError.value = ''
  batchSuccess.value = ''
}

const submitSingleImport = async () => {
  if (!singleFile.value || !singleCanSubmit.value) return

  singleSubmitting.value = true
  singleError.value = ''
  singleSuccess.value = ''

  try {
    const payload: SingleImportJobRequest = {
      filePath: singleFile.value.path,
      fileName: singleForm.fileName.trim(),
      description: singleForm.description.trim(),
      tags: singleForm.tags,
      keywords: singleForm.keywords.trim(),
      collectionId: singleForm.collectionId,
    }

    await api.jobs.submitSingleImportJob(payload)
    singleFile.value = null
    // 重置会清空提示，所以成功信息要在重置之后再写。
    resetSingleForm()
    singleSuccess.value = '已创建后台导入任务，可在右上角消息中心或任务中心查看进度。'
  } catch (error) {
    singleError.value = error instanceof Error ? error.message : '提交导入任务失败'
  } finally {
    singleSubmitting.value = false
  }
}

const submitBatchJob = async () => {
  if (!batchCanSubmit.value) return

  batchSubmitting.value = true
  batchError.value = ''
  batchSuccess.value = ''

  const payload: BatchImportJobRequest = {
    files: selectedFiles.value.map((file) => ({
      sourcePath: file.path,
      displayName: file.name,
    })),
    sharedTags: batchForm.sharedTagsText
      .split(',')
      .map((item) => item.trim())
      .filter(Boolean),
    sharedDescription: batchForm.sharedDescription.trim(),
    collectionId: batchForm.collectionId,
    enableAIGeneratedTags: batchForm.enableAIGeneratedTags,
    enableAIGeneratedDescription: batchForm.enableAIGeneratedDescription,
  }

  try {
    await jobsStore.submitBatchImportJob(payload)
    clearFiles()
    // 同上：clearFiles 和 resetBatchForm 都会清空提示。
    resetBatchForm()
    batchSuccess.value = '后台导入任务已创建，可从右上角任务入口或任务中心继续查看。'
  } catch (error) {
    batchError.value = error instanceof Error ? error.message : '提交任务失败'
  } finally {
    batchSubmitting.value = false
  }
}
</script>

<style scoped>
.import-view {
  position: relative;
  --import-surface: rgba(255, 255, 255, 0.92);
  --import-surface-muted: rgba(243, 246, 250, 0.88);
  --import-control: rgba(248, 250, 253, 0.92);
  --import-border: rgba(148, 163, 184, 0.28);
  --import-border-strong: rgba(148, 163, 184, 0.5);
  --import-accent: #2d8cf0;
  --import-accent-strong: #1565c0;
  --import-accent-soft: rgba(45, 140, 240, 0.1);
  --import-text: #223042;
  --import-text-soft: #5f6f82;
  --import-text-faint: #7f8ea3;
  --import-error-bg: rgba(244, 67, 54, 0.08);
  --import-error-text: #c62828;
  --import-success-bg: rgba(76, 175, 80, 0.1);
  --import-success-text: #2e7d32;
}

[data-theme='dark'] .import-view {
  --import-surface: rgba(35, 42, 52, 0.92);
  --import-surface-muted: rgba(42, 50, 61, 0.7);
  --import-control: rgba(30, 37, 47, 0.7);
  --import-border: rgba(173, 191, 214, 0.16);
  --import-border-strong: rgba(173, 191, 214, 0.3);
  --import-accent: #4da3ff;
  --import-accent-strong: #8cc2ff;
  --import-accent-soft: rgba(77, 163, 255, 0.14);
  --import-text: #f2f6fb;
  --import-text-soft: #c2cfdd;
  --import-text-faint: #92a0b1;
  --import-error-bg: rgba(244, 96, 96, 0.14);
  --import-error-text: #ffb7b7;
  --import-success-bg: rgba(82, 184, 149, 0.14);
  --import-success-text: #9fe6c8;
}

.import-view .page-content {
  position: relative;
  padding: 20px 44px 50px;
}

.import-view .page-stack {
  gap: 16px;
}

.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  margin: -1px;
  padding: 0;
  overflow: hidden;
  clip: rect(0 0 0 0);
  white-space: nowrap;
  border: 0;
}

/* 模式切换充当页头：底部这条通栏细线接管了原来大标题的视觉重量。 */
.imp-tabs {
  display: flex;
  gap: 24px;
  border-bottom: 1px solid var(--import-border);
}

.imp-tab {
  position: relative;
  padding: 4px 2px 12px;
  border-radius: 0;
  background: transparent;
  color: var(--import-text-faint);
  font-size: 18px;
  font-weight: 700;
  letter-spacing: -0.025em;
  transition: color 0.2s ease;
}

.imp-tab::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: -1px;
  height: 2px;
  border-radius: 2px;
  background: transparent;
}

.imp-tab:hover {
  color: var(--import-text-soft);
}

.imp-tab.active {
  color: var(--import-accent-strong);
}

.imp-tab.active::after {
  background: var(--import-accent);
}

.imp-lede {
  margin-top: 13px;
  color: var(--import-text-faint);
  font-size: 12px;
  line-height: 1.7;
}

.import-layout {
  display: grid;
  gap: 16px;
  align-items: start;
}

/* 单文件只有一个主体，所以它是页头条；批量是一份清单，所以它占一整栏。 */
.import-layout.single {
  grid-template-columns: minmax(0, 1.3fr) minmax(268px, 0.7fr);
}

/* 没选文件、或没配置 AI 时都只剩一栏，元信息占满整行。 */
.import-layout.single.solo {
  grid-template-columns: minmax(0, 1fr);
}

.import-layout.batch {
  grid-template-columns: minmax(290px, 1fr) minmax(0, 1fr);
}

.imp-source {
  display: grid;
  grid-column: 1 / -1;
  grid-template-columns: minmax(258px, 0.8fr) minmax(0, 1.2fr);
  align-items: stretch;
  gap: 12px;
}

.imp-commit {
  display: flex;
  grid-column: 1 / -1;
  justify-content: flex-end;
  align-items: center;
  gap: 10px;
}

.imp-panel {
  padding: 20px;
  border: 1px solid var(--import-border);
  border-radius: var(--radius-lg);
  background: color-mix(in srgb, var(--import-surface) 82%, transparent);
  box-shadow: 0 10px 28px var(--shadow-color);
}

.imp-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 14px;
  margin-bottom: 16px;
}

.imp-head h2 {
  font-size: 15px;
  font-weight: 700;
  color: var(--import-text);
  letter-spacing: -0.02em;
}

.imp-head p {
  margin-top: 5px;
  color: var(--import-text-faint);
  font-size: 12px;
  line-height: 1.55;
}

.imp-pill {
  flex: none;
  padding: 5px 9px;
  border-radius: 8px;
  color: var(--import-text-faint);
  background: var(--import-surface-muted);
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
}

.imp-pill.on {
  color: var(--import-accent-strong);
  background: var(--import-accent-soft);
}

/* 选择控件保持实线边框和实心底色，读起来是按钮而不是拖放目标。 */
.imp-pick {
  display: grid;
  grid-template-columns: 38px minmax(0, 1fr) 10px;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 13px 14px;
  border: 1px solid var(--import-border-strong);
  border-radius: var(--radius-md);
  background: var(--import-control);
  text-align: left;
  transition: border-color 0.2s ease, background-color 0.2s ease, transform 0.2s ease;
}

.imp-pick:hover {
  border-color: color-mix(in srgb, var(--import-accent) 45%, transparent);
  background: color-mix(in srgb, var(--import-accent) 5%, var(--import-control));
  transform: translateY(-1px);
}

.imp-pick-tile {
  display: grid;
  place-items: center;
  width: 38px;
  height: 38px;
  border-radius: 10px;
  color: var(--import-accent-strong);
  background: var(--import-accent-soft);
}

.imp-pick-tile svg {
  width: 17px;
  height: 17px;
}

.imp-pick-copy {
  min-width: 0;
}

.imp-pick-copy strong {
  display: block;
  color: var(--import-text);
  font-size: 14px;
  font-weight: 700;
}

.imp-pick-copy span {
  display: block;
  margin-top: 4px;
  color: var(--import-text-faint);
  font-size: 12px;
  line-height: 1.5;
}

.imp-pick-chevron {
  color: var(--import-text-faint);
  font-style: normal;
  font-size: 16px;
}

.imp-stats {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  margin-top: 14px;
}

.imp-stat {
  padding: 13px 14px;
  border: 1px solid var(--import-border);
  border-radius: var(--radius-md);
  background: var(--import-surface-muted);
}

.imp-stat strong {
  display: block;
  color: var(--import-text);
  font-size: 20px;
  font-weight: 700;
  letter-spacing: -0.03em;
}

.imp-stat span {
  display: block;
  margin-top: 5px;
  color: var(--import-text-faint);
  font-size: 12px;
}

.imp-list-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
  margin-top: 16px;
  padding-bottom: 9px;
  border-bottom: 1px solid var(--import-border);
  color: var(--import-text-faint);
  font-size: 12px;
  font-weight: 600;
}

.imp-text-button {
  padding: 0;
  background: transparent;
  color: var(--import-accent-strong);
  font-size: 12px;
  font-weight: 600;
}

.imp-tray {
  display: grid;
  gap: 8px;
  margin-top: 10px;
  max-height: 320px;
  overflow-y: auto;
}

.imp-file {
  display: grid;
  grid-template-columns: 36px minmax(0, 1fr) auto;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border: 1px solid var(--import-border);
  border-radius: var(--radius-md);
  background: var(--import-surface-muted);
}

.imp-file-lead {
  padding: 13px 14px;
}

.imp-file-copy {
  min-width: 0;
}

.imp-file-copy strong {
  display: block;
  overflow: hidden;
  color: var(--import-text);
  font-size: 14px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.imp-file-copy p {
  margin-top: 4px;
  color: var(--import-text-faint);
  font-size: 12px;
}

.imp-remove {
  flex: none;
  padding: 6px 10px;
  border-radius: 8px;
  color: var(--import-text-faint);
  background: transparent;
  font-size: 12px;
  font-weight: 600;
}

.imp-remove:hover {
  color: var(--error-color);
  background: var(--import-error-bg);
}

/* 类型色片沿用资料库的类型语言，让导入和归档后是同一套识别方式。 */
.imp-type {
  display: grid;
  place-items: center;
  width: 36px;
  height: 36px;
  border-radius: 9px;
  font-size: 14px;
  font-weight: 700;
  color: #6b7b90;
  background: #eef1f5;
}

.imp-type.video {
  color: #145ab5;
  background: #e6f0ff;
}

.imp-type.document {
  color: #46618a;
  background: #eaf0f7;
}

.imp-type.music {
  color: #5b4a9c;
  background: #ebe6fa;
}

.imp-type.image {
  color: #1f7a5e;
  background: #dff3ec;
}

.imp-type.archive {
  color: #8f6210;
  background: #fff1d6;
}

.imp-type.installer {
  color: #a1414e;
  background: #f8e4e7;
}

[data-theme='dark'] .imp-type {
  color: #a3b2c5;
  background: rgba(140, 160, 185, 0.16);
}

[data-theme='dark'] .imp-type.video {
  color: #9cc4f5;
  background: rgba(45, 120, 220, 0.2);
}

[data-theme='dark'] .imp-type.document {
  color: #a8bcd4;
  background: rgba(120, 150, 190, 0.18);
}

[data-theme='dark'] .imp-type.music {
  color: #b9a9e8;
  background: rgba(130, 105, 215, 0.2);
}

[data-theme='dark'] .imp-type.image {
  color: #7fd3b4;
  background: rgba(45, 170, 130, 0.2);
}

[data-theme='dark'] .imp-type.archive {
  color: #e2bd77;
  background: rgba(190, 140, 40, 0.2);
}

[data-theme='dark'] .imp-type.installer {
  color: #eda3ad;
  background: rgba(210, 90, 110, 0.2);
}

.imp-form {
  display: grid;
  gap: 14px;
}

.imp-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.imp-field {
  display: grid;
  gap: 8px;
}

.imp-field.wide {
  grid-column: 1 / -1;
}

.imp-field > span {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
  color: var(--import-text-soft);
  font-size: 12px;
  font-weight: 600;
}

.imp-field em {
  color: var(--import-text-faint);
  font-style: normal;
  font-weight: 400;
}

.imp-toggles {
  display: grid;
  gap: 10px;
}

.imp-toggle {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 13px 14px;
  border: 1px solid var(--import-border);
  border-radius: var(--radius-md);
  background: var(--import-surface-muted);
  cursor: pointer;
}

.imp-toggle input {
  width: 16px;
  height: 16px;
  margin-top: 2px;
  flex: none;
  accent-color: var(--import-accent);
}

.imp-toggle strong {
  display: block;
  color: var(--import-text);
  font-size: 13px;
}

.imp-toggle p {
  margin-top: 4px;
  color: var(--import-text-faint);
  font-size: 12px;
  line-height: 1.6;
}

.imp-feedback {
  margin-right: auto;
  padding: 9px 13px;
  border-radius: 10px;
  font-size: 13px;
  line-height: 1.5;
}

.imp-feedback.success {
  color: var(--import-success-text);
  background: var(--import-success-bg);
}

.imp-feedback.error {
  color: var(--import-error-text);
  background: var(--import-error-bg);
}

.imp-empty {
  padding: 22px 18px;
  border: 1px solid var(--import-border);
  border-radius: var(--radius-md);
  background: var(--import-surface-muted);
  text-align: center;
}

/* 来源条里的空位要和选择控件同高，所以改成左对齐、垂直居中。 */
.imp-empty-lead {
  display: grid;
  align-content: center;
  padding: 13px 16px;
  text-align: left;
}

.imp-empty strong {
  display: block;
  color: var(--import-text);
  font-size: 13px;
}

.imp-empty p {
  margin-top: 6px;
  color: var(--import-text-faint);
  font-size: 12px;
  line-height: 1.6;
}

.imp-locked {
  display: grid;
  place-items: center;
  min-height: 140px;
  padding: 24px;
  border: 1px solid var(--import-border);
  border-radius: var(--radius-md);
  background: var(--import-surface-muted);
  text-align: center;
}

.imp-locked strong {
  display: block;
  color: var(--import-text);
  font-size: 14px;
}

.imp-locked p {
  max-width: 360px;
  margin-top: 7px;
  color: var(--import-text-faint);
  font-size: 12px;
  line-height: 1.7;
}

@media (max-width: 1080px) {
  .import-layout.single,
  .import-layout.batch {
    grid-template-columns: 1fr;
  }

  .imp-source {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 820px) {
  .import-view .page-content {
    padding: 16px 12px 40px;
  }

  .imp-panel {
    padding: 16px;
  }

  .imp-grid {
    grid-template-columns: 1fr;
  }

  .imp-head {
    flex-direction: column;
    align-items: flex-start;
  }

  .imp-commit {
    flex-direction: column-reverse;
    align-items: stretch;
  }

  .imp-commit .imp-feedback {
    margin-right: 0;
  }
}

@media (max-width: 620px) {
  .imp-tabs {
    gap: 16px;
  }

  .imp-tab {
    font-size: 16px;
  }

  .imp-file,
  .imp-file-lead {
    grid-template-columns: 36px minmax(0, 1fr);
    row-gap: 8px;
  }

  .imp-remove {
    grid-column: 2;
    justify-self: start;
  }
}

@media (prefers-reduced-motion: reduce) {
  .imp-pick:hover {
    transform: none;
  }
}
</style>

<template>
  <div class="page-shell import-view">
    <AppHeader />

    <div class="page-content">
      <div class="page-stack">
        <section class="page-topbar import-topbar">
          <div class="page-topbar-copy">
            <h1 class="page-topbar-title">导入</h1>
            <p class="page-topbar-note">
              {{ mode === 'single' ? '单文件适合精细补充信息。' : '批量导入会创建后台任务，可在右上角继续查看进度。' }}
            </p>
          </div>
          <div class="mode-switch" role="tablist" aria-label="导入模式">
            <button
              type="button"
              class="mode-pill"
              :class="{ active: mode === 'single' }"
              @click="mode = 'single'"
            >
              单文件导入
            </button>
            <button
              type="button"
              class="mode-pill"
              :class="{ active: mode === 'batch' }"
              @click="mode = 'batch'"
            >
              批量导入任务
            </button>
          </div>
        </section>

        <div class="workspace">
          <section class="main-column">
            <template v-if="mode === 'single'">
              <article class="section-panel panel">
                <div class="panel-head">
                  <div>
                    <h2 class="section-title">选择一个文件，完整补充它的资料信息。</h2>
                    <p class="panel-copy">适合精细整理、补标签、加描述、归入合集。</p>
                  </div>
                  <button class="btn secondary" type="button" @click="handleSelectSingleFile">
                    {{ singleFile ? '重新选择' : '选择文件' }}
                  </button>
                </div>

                <div v-if="!singleFile" class="empty-surface">
                  <strong>还没有选择文件</strong>
                  <p>选择后可以在下方编辑名称、标签、关键词、描述和合集。</p>
                </div>

                <div v-else class="single-preview">
                  <div class="file-avatar">{{ singleFile.shortLabel }}</div>
                  <div class="file-copy">
                    <strong>{{ singleFile.name }}</strong>
                    <p>{{ singleFile.typeLabel }} · {{ formatFileSize(singleFile.size) }}</p>
                  </div>
                  <button class="btn ghost preview-action" type="button" @click="clearSingleFile">移除</button>
                </div>
              </article>

              <article v-if="singleFile" class="section-panel panel">
                <div class="panel-head slim">
                  <div>
                    <h2 class="section-title">完善这份文件的元信息。</h2>
                  </div>
                </div>

                <div class="field-grid">
                  <label class="field">
                    <span>文件名</span>
                    <input v-model="singleForm.fileName" type="text" placeholder="输入文件名" />
                  </label>

                  <label class="field">
                    <span>关键词</span>
                    <input v-model="singleForm.keywords" type="text" placeholder="例如：课程、会议、剪辑" />
                  </label>

                  <label class="field field-full">
                    <span>合集 / 系列</span>
                    <input v-model="singleForm.collectionName" type="text" placeholder="例如：2026 课程资料" />
                  </label>
                </div>

                <div class="metadata-shell">
                  <FileMetadataFields
                    :file-name="singleForm.fileName"
                    :file-type="singleFile.type"
                    :user-keywords="singleForm.keywords"
                    v-model:modelValueTags="singleForm.tags"
                    v-model:modelValueDescription="singleForm.description"
                  />
                </div>

                <div v-if="singleError" class="feedback error">{{ singleError }}</div>
                <div v-if="singleSuccess" class="feedback success">{{ singleSuccess }}</div>

                <div class="actions">
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
              </article>
            </template>

            <template v-else>
              <article class="section-panel panel">
                <div class="panel-head">
                  <div>
                    <h2 class="section-title">建立一个后台导入任务。</h2>
                    <p class="panel-copy">适合一次性处理多份资料，提交后可以继续做别的事情。</p>
                  </div>
                  <div class="head-actions">
                    <button class="btn secondary" type="button" @click="handleSelectFiles">添加文件</button>
                    <button
                      class="btn ghost"
                      type="button"
                      :disabled="batchSubmitting || selectedFiles.length === 0"
                      @click="clearFiles"
                    >
                      清空
                    </button>
                  </div>
                </div>

                <div class="summary-grid">
                  <div class="summary-card">
                    <span class="summary-value">{{ selectedFiles.length }}</span>
                    <span class="summary-label">文件数量</span>
                  </div>
                  <div class="summary-card">
                    <span class="summary-value">{{ totalSizeLabel }}</span>
                    <span class="summary-label">预计处理体量</span>
                  </div>
                </div>

                <div v-if="selectedFiles.length === 0" class="empty-surface">
                  <strong>还没有加入批量文件</strong>
                  <p>先选择多个文件，再统一设置标签、描述和合集，然后创建后台任务。</p>
                </div>

                <div v-else class="batch-list scroll-soft">
                  <article v-for="file in selectedFiles" :key="file.path" class="file-row">
                    <div class="file-avatar">{{ file.shortLabel }}</div>
                    <div class="file-copy">
                      <strong>{{ file.name }}</strong>
                      <p>{{ file.typeLabel }} · {{ formatFileSize(file.size) }}</p>
                    </div>
                    <button class="btn ghost preview-action" type="button" @click="removeFile(file.path)">移除</button>
                  </article>
                </div>
              </article>

              <article class="section-panel panel">
                <div class="panel-head slim">
                  <div>
                    <h2 class="section-title">给这一批文件设置统一的补充信息。</h2>
                  </div>
                </div>

                <div class="field-grid">
                  <label class="field">
                    <span>统一标签</span>
                    <input v-model="batchForm.sharedTagsText" type="text" placeholder="例如：课程、待整理、项目A" />
                  </label>

                  <label class="field">
                    <span>合集 / 系列</span>
                    <input v-model="batchForm.collectionName" type="text" placeholder="例如：2026 课程资料" />
                  </label>

                  <label class="field field-full">
                    <span>统一描述</span>
                    <textarea
                      v-model="batchForm.sharedDescription"
                      rows="5"
                      placeholder="为这一批文件补充统一说明"
                    />
                  </label>
                </div>

                <div class="toggle-grid">
                  <label class="toggle-card">
                    <input v-model="batchForm.enableAIGeneratedTags" type="checkbox" />
                    <div>
                      <strong>AI 追加标签</strong>
                      <p>在统一标签的基础上，为每个文件补充更细的标签。</p>
                    </div>
                  </label>

                  <label class="toggle-card">
                    <input v-model="batchForm.enableAIGeneratedDescription" type="checkbox" />
                    <div>
                      <strong>AI 生成描述</strong>
                      <p>优先为每个文件生成单独描述，失败时退回到统一描述。</p>
                    </div>
                  </label>
                </div>

                <div v-if="batchError" class="feedback error">{{ batchError }}</div>
                <div v-if="batchSuccess" class="feedback success">{{ batchSuccess }}</div>

                <div class="actions">
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
              </article>
            </template>
          </section>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import AppHeader from '@/components/AppHeader.vue'
import FileMetadataFields from '@/components/FileMetadataFields.vue'
import { api } from '@/api'
import { formatFileSize } from '@/utils/constants'
import type { BatchImportJobRequest, SelectedFile } from '@/types/jobs'
import { useJobsStore } from '@/store/modules/jobs'

type ImportMode = 'single' | 'batch'

interface RichSelectedFile extends SelectedFile {
  type: string
  typeLabel: string
  shortLabel: string
}

const FILE_TYPE_META: Record<string, { label: string; shortLabel: string }> = {
  video: { label: '视频', shortLabel: '影' },
  document: { label: '文档', shortLabel: '文' },
  music: { label: '音频', shortLabel: '音' },
  archive: { label: '压缩包', shortLabel: '压' },
  installer: { label: '安装包', shortLabel: '装' },
  image: { label: '图片', shortLabel: '图' },
  other: { label: '其他', shortLabel: '其' },
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

const router = useRouter()
const jobsStore = useJobsStore()

const mode = ref<ImportMode>('single')
const singleFile = ref<RichSelectedFile | null>(null)
const selectedFiles = ref<RichSelectedFile[]>([])
const singleSubmitting = ref(false)
const batchSubmitting = ref(false)
const singleError = ref('')
const singleSuccess = ref('')
const batchError = ref('')
const batchSuccess = ref('')

const singleForm = reactive({
  fileName: '',
  keywords: '',
  collectionName: '',
  tags: [] as string[],
  description: '',
})

const batchForm = reactive({
  sharedTagsText: '',
  sharedDescription: '',
  collectionName: '',
  enableAIGeneratedTags: true,
  enableAIGeneratedDescription: false,
})

const totalSize = computed(() => selectedFiles.value.reduce((sum, file) => sum + file.size, 0))
const totalSizeLabel = computed(() => formatFileSize(totalSize.value))
const singleCanSubmit = computed(() => !!singleFile.value && singleForm.fileName.trim().length > 0)
const batchCanSubmit = computed(() => selectedFiles.value.length > 0)

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

  const path = await api.system.selectFile()
  if (!path || path === 'success') return

  const metadata = await api.system.getMetadata(path, 'file')
  const fileName = path.split(/[/\\]/).pop() || ''
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
  singleForm.collectionName = ''
  singleForm.tags = []
  singleForm.description = ''
  singleError.value = ''
  singleSuccess.value = ''
}

const resetBatchForm = () => {
  batchForm.sharedTagsText = ''
  batchForm.sharedDescription = ''
  batchForm.collectionName = ''
  batchForm.enableAIGeneratedTags = true
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
    await api.file.importWithMetadata(
      singleFile.value.path,
      singleForm.fileName.trim(),
      singleForm.description.trim(),
      singleForm.tags,
      singleForm.keywords.trim(),
      singleForm.collectionName.trim()
    )
    singleSuccess.value = '文件已成功导入，稍后会跳转到文件列表。'
    setTimeout(() => {
      router.push('/files')
    }, 400)
  } catch (error) {
    singleError.value = error instanceof Error ? error.message : '导入文件失败'
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
    collectionName: batchForm.collectionName.trim(),
    enableAIGeneratedTags: batchForm.enableAIGeneratedTags,
    enableAIGeneratedDescription: batchForm.enableAIGeneratedDescription,
  }

  try {
    await jobsStore.submitBatchImportJob(payload)
    batchSuccess.value = '后台导入任务已创建，可从右上角任务入口或任务中心继续查看。'
    clearFiles()
    resetBatchForm()
  } catch (error) {
    batchError.value = error instanceof Error ? error.message : '提交任务失败'
  } finally {
    batchSubmitting.value = false
  }
}
</script>

<style scoped>
.import-topbar {
  align-items: center;
}

.mode-switch {
  display: inline-flex;
  gap: 6px;
  padding: 6px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.62);
  border: 1px solid rgba(146, 165, 192, 0.18);
}

.mode-pill {
  min-width: 142px;
  min-height: 42px;
  padding: 10px 16px;
  border-radius: 999px;
  color: var(--text-soft);
  font-size: 13px;
  font-weight: 600;
}

.mode-pill.active {
  background: rgba(255, 255, 255, 0.9);
  color: var(--text-color);
  box-shadow:
    0 12px 24px rgba(44, 62, 94, 0.08),
    inset 0 0 0 1px rgba(111, 143, 216, 0.16);
}

.main-column {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.panel {
  padding: 22px;
}

.panel-head,
.single-preview,
.file-row,
.actions,
.head-actions,
.toggle-card {
  display: flex;
  gap: 14px;
}

.panel-head {
  justify-content: space-between;
  align-items: flex-start;
}

.panel-head.slim {
  margin-bottom: 4px;
}

.panel-copy {
  margin-top: 10px;
  color: var(--text-soft);
  line-height: 1.7;
}

.single-preview,
.file-row,
.toggle-card,
.summary-card {
  border: 1px solid rgba(146, 165, 192, 0.16);
  background: rgba(255, 255, 255, 0.58);
}

.single-preview,
.file-row {
  align-items: center;
  padding: 16px;
  border-radius: 22px;
}

.file-avatar {
  width: 54px;
  height: 54px;
  border-radius: 18px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(235, 241, 250, 0.94);
  color: var(--text-color);
  font-size: 18px;
  font-weight: 700;
}

.file-copy {
  flex: 1;
  min-width: 0;
}

.file-copy strong {
  display: block;
  font-size: 15px;
  color: var(--text-color);
}

.file-copy p {
  margin-top: 6px;
  color: var(--text-faint);
  font-size: 12px;
}

.preview-action {
  width: auto;
}

.summary-grid,
.field-grid,
.toggle-grid {
  display: grid;
  gap: 14px;
}

.summary-grid {
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
}

.summary-card {
  display: grid;
  gap: 6px;
  padding: 16px;
  border-radius: 22px;
}

.summary-value {
  font-size: 22px;
  font-weight: 700;
  color: var(--text-color);
}

.summary-label {
  font-size: 12px;
  color: var(--text-faint);
}

.batch-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-height: 360px;
  overflow-y: auto;
}

.field-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field span {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-color);
}

.field-full {
  grid-column: 1 / -1;
}

.metadata-shell {
  margin-top: 6px;
}

.toggle-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.toggle-card {
  align-items: flex-start;
  padding: 16px;
  border-radius: 22px;
}

.toggle-card input {
  margin-top: 4px;
}

.toggle-card strong {
  display: block;
  color: var(--text-color);
}

.toggle-card p {
  margin-top: 6px;
  color: var(--text-faint);
  font-size: 12px;
  line-height: 1.7;
}

.feedback {
  padding: 12px 14px;
  border-radius: 18px;
  font-size: 13px;
}

.feedback.error {
  background: rgba(204, 108, 108, 0.08);
  border: 1px solid rgba(204, 108, 108, 0.18);
  color: #a85d5d;
}

.feedback.success {
  background: rgba(59, 143, 116, 0.08);
  border: 1px solid rgba(59, 143, 116, 0.16);
  color: var(--success-color);
}

.actions {
  justify-content: flex-end;
}

@media (max-width: 820px) {
  .summary-grid,
  .field-grid,
  .toggle-grid {
    grid-template-columns: 1fr;
  }

  .panel-head,
  .head-actions,
  .actions {
    flex-direction: column;
    align-items: flex-start;
  }
}

@media (max-width: 640px) {
  .panel {
    padding: 16px;
  }

  .mode-switch,
  .head-actions,
  .actions {
    width: 100%;
    flex-direction: column;
  }

  .mode-pill,
  .preview-action {
    width: 100%;
  }

  .single-preview,
  .file-row {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>

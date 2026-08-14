<template>
  <div class="page-shell import-view">
    <AppHeader />

    <div class="page-content">
      <div class="page-stack">
        <section class="page-topbar import-topbar">
          <div class="page-topbar-copy">
            <span class="eyebrow import-eyebrow">导入工作台</span>
            <h1 class="page-topbar-title">导入</h1>
            <p class="page-topbar-note">
              {{ mode === 'single' ? '单文件适合细化整理一份资料。' : '批量导入会创建后台任务，适合一次处理多份文件。' }}
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

        <section v-if="mode === 'single'" class="workspace-single">
          <article class="stage-panel single-panel">
            <div class="panel-head panel-head-compact">
              <div>
                <h2 class="section-title">单文件导入</h2>
                <p class="panel-copy">把一个文件从选择、补全到提交，放在同一条流程里完成。</p>
              </div>
              <span class="status-pill">{{ singleFile ? '已选中 1 个文件' : '等待选择文件' }}</span>
            </div>

            <div class="single-flow">
              <div class="single-selection">
                <div class="single-selection-head">
                  <div>
                    <h3>文件</h3>
                    <p>先把要导入的内容放进来，再补完整的资料。</p>
                  </div>
                  <button class="btn secondary" type="button" @click="handleSelectSingleFile">
                    {{ singleFile ? '重新选择文件' : '选择文件' }}
                  </button>
                </div>

                <div class="picker-shell">
                  <button class="pick-button single-picker" type="button" @click="handleSelectSingleFile">
                    <span class="pick-button-mark">＋</span>
                    <span class="pick-button-copy">
                      <span class="pick-button-text">{{ singleFile ? '重新选择文件' : '选择文件' }}</span>
                      <span class="pick-button-note">支持视频、文档、音频、压缩包、安装包、图片等格式</span>
                    </span>
                  </button>

                  <div v-if="singleFile" class="file-sheet single-file-card">
                    <div class="file-sheet-icon">{{ singleFile.shortLabel }}</div>
                    <div class="file-sheet-copy">
                      <strong class="file-sheet-name">{{ singleFile.name }}</strong>
                      <div class="file-sheet-meta">
                        <span>{{ singleFile.typeLabel }}</span>
                        <span>·</span>
                        <span>{{ formatFileSize(singleFile.size) }}</span>
                      </div>
                    </div>
                    <button class="btn ghost sheet-action" type="button" @click="clearSingleFile">移除</button>
                  </div>

                  <div v-else class="empty-surface import-empty">
                    <strong>还没有选择文件</strong>
                    <p>选择后会展开名称、标签、关键词、描述和合集设置。</p>
                  </div>
                </div>
              </div>

              <div class="single-divider"></div>

              <div v-if="singleFile" class="single-form">
                <div class="single-form-head">
                  <div>
                    <h3>元信息</h3>
                    <p>补全一份资料的可检索信息。</p>
                  </div>
                </div>

                <div class="field-grid single-field-grid">
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
                    <CollectionSelector v-model="singleForm.collectionId" :collections="collections" />
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

                <div class="actions single-actions">
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
              </div>

              <div v-else class="single-form single-form-empty">
                <div class="single-form-head">
                  <div>
                    <h3>元信息</h3>
                    <p>选择文件后这里会展开。</p>
                  </div>
                </div>
                <div class="empty-surface stage-empty">
                  <strong>先选择一个文件</strong>
                  <p>文件选中后，这一栏会展开名称、标签、关键词、描述和合集设置。</p>
                </div>
              </div>
            </div>
          </article>
        </section>

        <section v-else class="workspace-grid mode-batch">
          <article class="stage-panel stage-rail stage-rail-left">
            <div class="panel-head panel-head-compact">
              <div>
                <h2 class="section-title">文件</h2>
                <p class="panel-copy">先把要导入的内容放进来，再补完整的资料。</p>
              </div>
              <div class="panel-meta-row">
                <span class="status-pill">{{ selectedFiles.length }} 个文件</span>
              </div>
            </div>

            <div class="picker-shell">
              <div class="batch-toolbar">
                <button class="btn secondary" type="button" @click="handleSelectFiles">添加文件</button>
                <button class="btn ghost" type="button" :disabled="batchSubmitting || selectedFiles.length === 0" @click="clearFiles">
                  清空
                </button>
              </div>

              <div class="summary-grid">
                <div class="summary-card">
                  <span class="summary-value">{{ selectedFiles.length }}</span>
                  <span class="summary-label">文件数量</span>
                </div>
                <div class="summary-card">
                  <span class="summary-value">{{ totalSizeLabel }}</span>
                  <span class="summary-label">预计体量</span>
                </div>
              </div>

              <div v-if="selectedFiles.length === 0" class="empty-surface import-empty">
                <strong>还没有加入批量文件</strong>
                <p>先选择多个文件，再统一设置标签、描述和合集，最后创建后台任务。</p>
              </div>

              <div v-else class="batch-list scroll-soft">
                <article v-for="file in selectedFiles" :key="file.path" class="file-row">
                  <div class="file-avatar">{{ file.shortLabel }}</div>
                  <div class="file-copy">
                    <strong>{{ file.name }}</strong>
                    <p>{{ file.typeLabel }} · {{ formatFileSize(file.size) }}</p>
                  </div>
                  <button class="btn ghost sheet-action" type="button" @click="removeFile(file.path)">移除</button>
                </article>
              </div>
            </div>
          </article>

          <article class="stage-panel stage-rail stage-rail-right">
            <div class="panel-head panel-head-compact">
              <div>
                <h2 class="section-title">元信息</h2>
                <p class="panel-copy">给这一批文件统一设置公共字段。</p>
              </div>
              <span class="status-pill">批量公共字段</span>
            </div>

            <div class="form-stack">
              <div class="field-grid single-field-grid">
                <label class="field field-full">
                  <span>统一标签</span>
                  <input v-model="batchForm.sharedTagsText" type="text" placeholder="例如：课程、待整理、项目A" />
                </label>

                <label class="field">
                  <span>合集 / 系列</span>
                  <CollectionSelector v-model="batchForm.collectionId" :collections="collections" />
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
            </div>
          </article>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { api } from '@/api'
import AppHeader from '@/components/AppHeader.vue'
import CollectionSelector from '@/components/CollectionSelector.vue'
import FileMetadataFields from '@/components/FileMetadataFields.vue'
import type { BatchImportJobRequest, SelectedFile, SingleImportJobRequest } from '@/types/jobs'
import type { Collection } from '@/types'
import { formatFileSize } from '@/utils/constants'
import { useJobsStore } from '@/store/modules/jobs'
import { computed, onMounted, reactive, ref } from 'vue'

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

const jobsStore = useJobsStore()

const collections = ref<Collection[]>([])
onMounted(async () => {
  collections.value = await api.collection.getAll()
})

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
  collectionId: undefined as number | undefined,
  tags: [] as string[],
  description: '',
})

const batchForm = reactive({
  sharedTagsText: '',
  sharedDescription: '',
  collectionId: undefined as number | undefined,
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
    const payload: SingleImportJobRequest = {
      filePath: singleFile.value.path,
      fileName: singleForm.fileName.trim(),
      description: singleForm.description.trim(),
      tags: singleForm.tags,
      keywords: singleForm.keywords.trim(),
      collectionId: singleForm.collectionId,
    }

    await api.jobs.submitSingleImportJob(payload)
    singleSuccess.value = '已创建后台导入任务，可在右上角消息中心或任务中心查看进度。'
    singleFile.value = null
    resetSingleForm()
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
.import-view {
  position: relative;
  --import-surface: rgba(255, 255, 255, 0.92);
  --import-surface-strong: rgba(248, 250, 253, 0.98);
  --import-surface-muted: rgba(243, 246, 250, 0.88);
  --import-control: rgba(248, 250, 253, 0.92);
  --import-control-hover: rgba(235, 242, 251, 0.98);
  --import-topbar-start: rgba(255, 255, 255, 0.96);
  --import-topbar-end: rgba(244, 248, 253, 0.94);
  --import-highlight: rgba(45, 140, 240, 0.05);
  --import-border: rgba(148, 163, 184, 0.28);
  --import-border-strong: rgba(45, 140, 240, 0.38);
  --import-accent: #2d8cf0;
  --import-accent-strong: #1976d2;
  --import-text: #223042;
  --import-text-soft: #5f6f82;
  --import-text-faint: #7f8ea3;
  --import-on-accent: #ffffff;
  --import-error-bg: rgba(244, 67, 54, 0.08);
  --import-error-border: rgba(244, 67, 54, 0.22);
  --import-error-text: #c62828;
  --import-success-bg: rgba(76, 175, 80, 0.08);
  --import-success-border: rgba(76, 175, 80, 0.22);
  --import-success-text: #2e7d32;
}

:global([data-theme='dark']) .import-view {
  --import-surface: rgba(18, 26, 38, 0.82);
  --import-surface-strong: rgba(23, 32, 45, 0.94);
  --import-surface-muted: rgba(42, 50, 61, 0.82);
  --import-control: rgba(10, 15, 24, 0.45);
  --import-control-hover: rgba(18, 28, 42, 0.44);
  --import-topbar-start: rgba(21, 29, 41, 0.88);
  --import-topbar-end: rgba(16, 23, 34, 0.78);
  --import-highlight: rgba(126, 176, 255, 0.06);
  --import-border: rgba(173, 191, 214, 0.16);
  --import-border-strong: rgba(125, 168, 224, 0.28);
  --import-accent: #7eb0ff;
  --import-accent-strong: #aac9ff;
  --import-text: #f6fbff;
  --import-text-soft: #c4d1df;
  --import-text-faint: #8ea0b6;
  --import-on-accent: #ffffff;
  --import-error-bg: rgba(244, 96, 96, 0.1);
  --import-error-border: rgba(244, 96, 96, 0.22);
  --import-error-text: #ffb7b7;
  --import-success-bg: rgba(82, 184, 149, 0.1);
  --import-success-border: rgba(82, 184, 149, 0.22);
  --import-success-text: #b9f2d9;
}

.import-view::before {
  content: '';
  position: fixed;
  inset: 0;
  pointer-events: none;
  background:
    linear-gradient(180deg, rgba(45, 140, 240, 0.035), transparent),
    linear-gradient(90deg, rgba(45, 140, 240, 0.035) 1px, transparent 1px),
    linear-gradient(rgba(45, 140, 240, 0.035) 1px, transparent 1px);
  background-size: auto, 64px 64px, 64px 64px;
  mix-blend-mode: normal;
  opacity: 0.7;
}

.import-view .page-content {
  position: relative;
  padding: 12px 16px 16px;
}

.import-view .page-stack {
  gap: 14px;
}

.import-topbar {
  align-items: center;
  padding: 16px 18px;
  border-radius: 14px;
  background: linear-gradient(180deg, var(--import-topbar-start), var(--import-topbar-end));
  border: 1px solid var(--import-border);
  box-shadow: 0 18px 44px rgba(0, 0, 0, 0.26);
  backdrop-filter: blur(18px);
}

.import-eyebrow {
  background-color: rgba(126, 176, 255, 0.12);
  color: var(--import-accent-strong);
}

.import-view .page-topbar-title {
  font-size: 22px;
}

.import-view .page-topbar-note {
  color: var(--import-text-soft);
}

.mode-switch {
  display: inline-flex;
  gap: 6px;
  padding: 4px;
  border-radius: 12px;
  background: var(--import-control);
  border: 1px solid var(--import-border);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.03);
}

.mode-pill {
  min-width: 126px;
  min-height: 38px;
  padding: 8px 14px;
  border-radius: 8px;
  color: var(--import-text-faint);
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0;
}

.mode-pill.active {
  background: linear-gradient(180deg, color-mix(in srgb, var(--import-accent) 18%, transparent), color-mix(in srgb, var(--import-accent) 8%, transparent));
  color: var(--import-on-accent);
  border: 1px solid var(--import-border-strong);
  box-shadow: 0 10px 24px rgba(28, 67, 123, 0.22);
}

.workspace-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(340px, 0.96fr);
  gap: 14px;
  align-items: start;
}

.workspace-single {
  width: 100%;
}

.mode-batch {
  grid-template-columns: minmax(0, 1fr) minmax(360px, 1fr);
}

.stage-panel {
  position: relative;
  padding: 18px;
  border-radius: 16px;
  background: linear-gradient(180deg, var(--import-surface), var(--import-surface-strong));
  border: 1px solid var(--import-border);
  box-shadow: 0 18px 40px rgba(0, 0, 0, 0.24);
  backdrop-filter: blur(18px);
  overflow: hidden;
}

.stage-panel::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, var(--import-highlight), transparent 30%, color-mix(in srgb, var(--import-text) 2%, transparent));
  pointer-events: none;
}

.stage-panel > * {
  position: relative;
}

.panel-head {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
}

.panel-head-compact {
  margin-bottom: 14px;
}

.section-title {
  font-size: 18px;
}

.panel-copy {
  margin-top: 6px;
  color: var(--import-text-soft);
  line-height: 1.65;
  font-size: 13px;
}

.single-panel {
  display: grid;
  gap: 16px;
}

.single-flow {
  display: grid;
  gap: 16px;
}

.single-selection,
.single-form {
  display: grid;
  gap: 12px;
}

.single-selection-head,
.single-form-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 14px;
}

.single-selection-head h3,
.single-form-head h3 {
  font-size: 15px;
  font-weight: 700;
  color: var(--import-text);
}

.single-selection-head p,
.single-form-head p {
  margin-top: 4px;
  color: var(--import-text-soft);
  font-size: 12px;
  line-height: 1.6;
}

.single-selection-head .btn {
  width: auto;
}

.panel-meta-row {
  display: flex;
  justify-content: flex-end;
}

.picker-shell,
.form-stack {
  display: grid;
  gap: 14px;
}

.pick-button {
  display: grid;
  justify-items: start;
  gap: 10px;
  width: 100%;
  padding: 22px 20px;
  border-radius: 14px;
  border: 1px dashed var(--import-border);
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--import-text) 3%, transparent), transparent),
    var(--import-control);
  color: var(--import-text);
  text-align: left;
}

.pick-button.single-picker {
  grid-template-columns: 36px minmax(0, 1fr);
  align-items: center;
  justify-items: start;
  min-height: 76px;
  padding: 14px 16px;
}

.pick-button:hover {
  border-color: var(--import-border-strong);
  background-color: var(--import-control-hover);
  transform: translateY(-1px);
}

.pick-button-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: color-mix(in srgb, var(--import-accent) 14%, transparent);
  color: var(--import-accent-strong);
  font-size: 22px;
  font-weight: 700;
}

.pick-button-copy {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.pick-button-text {
  font-size: 15px;
  font-weight: 700;
}

.pick-button-note {
  color: var(--import-text-faint);
  font-size: 13px;
  line-height: 1.6;
}

.file-sheet {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr) auto;
  gap: 14px;
  align-items: center;
  padding: 14px;
  border-radius: 14px;
  background: var(--import-surface-muted);
  border: 1px solid var(--import-border);
}

.single-file-card {
  background: color-mix(in srgb, var(--import-surface-muted) 72%, transparent);
}

.file-sheet-icon,
.file-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: linear-gradient(180deg, color-mix(in srgb, var(--import-accent) 18%, transparent), color-mix(in srgb, var(--import-accent) 8%, transparent));
  color: var(--import-on-accent);
  font-size: 16px;
  font-weight: 700;
}

.file-sheet-copy,
.file-copy {
  min-width: 0;
}

.file-sheet-name,
.file-copy strong {
  display: block;
  font-size: 15px;
  font-weight: 700;
  color: var(--import-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-sheet-meta,
.file-copy p {
  display: flex;
  gap: 6px;
  margin-top: 5px;
  color: var(--import-text-faint);
  font-size: 12px;
  line-height: 1.5;
}

.sheet-action {
  width: auto;
  flex-shrink: 0;
}

.summary-grid,
.field-grid,
.toggle-grid {
  display: grid;
  gap: 12px;
}

.summary-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.summary-card {
  display: grid;
  gap: 6px;
  padding: 14px;
  border-radius: 14px;
  background: var(--import-surface-muted);
  border: 1px solid var(--import-border);
}

.summary-value {
  font-size: 20px;
  font-weight: 700;
  color: var(--import-text);
}

.summary-label {
  font-size: 12px;
  color: var(--import-text-faint);
}

.batch-toolbar {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.batch-toolbar .btn {
  width: auto;
}

.batch-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-height: 360px;
  overflow-y: auto;
}

.file-row {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr) auto;
  gap: 14px;
  align-items: center;
  padding: 14px;
  border-radius: 14px;
  background: color-mix(in srgb, var(--import-surface-muted) 82%, transparent);
  border: 1px solid var(--import-border);
}

.file-copy p {
  margin-top: 4px;
}

.single-field-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field span {
  font-size: 12px;
  font-weight: 600;
  color: var(--import-text-soft);
}

.field-full {
  grid-column: 1 / -1;
}

.metadata-shell {
  padding-top: 2px;
}

.single-divider {
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--import-border), transparent);
}

.toggle-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.toggle-card {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  padding: 14px;
  border-radius: 14px;
  background: var(--import-surface-muted);
  border: 1px solid var(--import-border);
}

.toggle-card input {
  width: 16px;
  height: 16px;
  margin-top: 4px;
  flex-shrink: 0;
}

.toggle-card strong {
  display: block;
  font-size: 13px;
  color: var(--import-text);
}

.toggle-card p {
  margin-top: 4px;
  color: var(--import-text-faint);
  font-size: 12px;
  line-height: 1.65;
}

.feedback {
  padding: 11px 12px;
  border-radius: 12px;
  font-size: 13px;
  line-height: 1.6;
}

.feedback.error {
  background: var(--import-error-bg);
  border: 1px solid var(--import-error-border);
  color: var(--import-error-text);
}

.feedback.success {
  background: var(--import-success-bg);
  border: 1px solid var(--import-success-border);
  color: var(--import-success-text);
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding-top: 2px;
}

.single-actions {
  margin-top: 2px;
}

.empty-surface {
  min-height: 172px;
  padding: 22px;
  border-radius: 14px;
  background: color-mix(in srgb, var(--import-surface-muted) 72%, transparent);
  border: 1px dashed var(--import-border);
}

.import-empty,
.stage-empty {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 8px;
}

.empty-surface strong {
  color: var(--import-text);
  font-size: 15px;
}

.empty-surface p {
  color: var(--import-text-faint);
  line-height: 1.7;
}

@media (max-width: 1080px) {
  .workspace-grid,
  .mode-batch {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 820px) {
  .import-view .page-content {
    padding: 12px;
  }

  .import-topbar,
  .stage-panel {
    padding: 14px;
    border-radius: 14px;
  }

  .panel-head,
  .actions,
  .batch-toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .panel-meta-row {
    justify-content: flex-start;
  }

  .summary-grid,
.single-field-grid,
.toggle-grid {
    grid-template-columns: 1fr;
  }

  .single-selection-head,
  .single-form-head {
    flex-direction: column;
    align-items: stretch;
  }

  .single-selection-head .btn {
    width: 100%;
  }

  .file-sheet,
  .file-row {
    grid-template-columns: 1fr;
    align-items: start;
  }

  .sheet-action {
    width: 100%;
  }

  .mode-switch {
    width: 100%;
  }

  .mode-pill {
    flex: 1;
  }
}

@media (max-width: 640px) {
  .mode-switch {
    flex-direction: column;
  }

  .mode-pill {
    width: 100%;
  }
}
</style>

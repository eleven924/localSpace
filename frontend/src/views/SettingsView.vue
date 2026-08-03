<template>
  <div class="page-shell settings-view">
    <AppHeader />

    <div class="page-content">
      <div class="page-stack">
        <div class="settings-workbench" :class="{ 'nav-collapsed': navCollapsed }">
          <aside class="settings-sidebar" aria-label="设置导航">
            <div class="sidebar-top">
              <div class="sidebar-title">
                <span class="sidebar-mark">LS</span>
                <div class="sidebar-copy">
                  <strong>设置</strong>
                  <span>LocalSpace preferences</span>
                </div>
              </div>

              <button
                type="button"
                class="collapse-button"
                :title="navCollapsed ? '展开设置导航' : '收起设置导航'"
                :aria-label="navCollapsed ? '展开设置导航' : '收起设置导航'"
                :aria-expanded="String(!navCollapsed)"
                @click="toggleNavigation"
              >
                <span>{{ navCollapsed ? '›' : '‹' }}</span>
              </button>
            </div>

            <nav class="settings-nav">
              <button
                v-for="item in settingItems"
                :key="item.id"
                type="button"
                class="settings-nav-item"
                :class="{ active: activeSetting === item.id }"
                :title="item.label"
                @click="activeSetting = item.id"
              >
                <span class="nav-icon" aria-hidden="true">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                    <path
                      v-for="path in item.iconPaths"
                      :key="path"
                      :d="path"
                    />
                  </svg>
                </span>
                <span class="nav-copy">
                  <span>{{ item.label }}</span>
                  <small>{{ item.description }}</small>
                </span>
              </button>
            </nav>
          </aside>

          <section class="settings-detail">
            <header class="settings-detail-header">
              <div class="detail-title-wrap">
                <span class="detail-kicker">Preferences</span>
                <h1>{{ activeSettingMeta.label }}</h1>
                <p>{{ activeSettingMeta.description }}</p>
              </div>

              <button
                type="button"
                class="detail-nav-button"
                :title="navCollapsed ? '展开导航' : '收起导航'"
                :aria-label="navCollapsed ? '展开导航' : '收起导航'"
                @click="toggleNavigation"
              >
                {{ navCollapsed ? '展开导航' : '收起导航' }}
              </button>
            </header>

            <div class="settings-panel scroll-soft">
              <section v-show="activeSetting === 'storage'" class="settings-page">
                <StorageDirSelector
                  @master-dir-added="handleDirAdded"
                  @master-dir-removed="handleDirRemoved"
                  @master-dir-default-changed="handleDirDefaultChanged"
                />
              </section>

              <section v-show="activeSetting === 'layout'" class="settings-page">
                <StorageLayoutConfig />
              </section>

              <section v-show="activeSetting === 'openWith'" class="settings-page">
                <OpenWithConfig />
              </section>

              <section v-show="activeSetting === 'theme'" class="settings-page">
                <ThemeConfig @theme-changed="handleThemeChanged" @theme-reset="handleThemeReset" />
              </section>

              <section v-show="activeSetting === 'ai'" class="settings-page">
                <AIConfigForm @config-saved="handleAIConfigSaved" @config-reset="handleAIConfigReset" />
              </section>

              <section v-show="activeSetting === 'collections'" class="settings-page">
                <CollectionsSettings :files="filesStore.files" />
              </section>

              <section v-show="activeSetting === 'tasks'" class="settings-page">
                <TaskRetentionSettings />
              </section>

              <section v-show="activeSetting === 'about'" class="settings-page about-page">
                <div class="about-content">
                  <div class="app-info">
                    <div class="app-logo">LS</div>
                    <div>
                      <h3>LocalSpace</h3>
                      <p class="app-version">版本 0.1.0</p>
                    </div>
                  </div>

                  <div class="app-description">
                    <p>
                      LocalSpace 是一个面向本地资料整理与回看的桌面工具，适合把图片、文档、音视频和项目素材统一收进一个入口。
                    </p>

                    <ul class="feature-list">
                      <li>支持单文件导入与批量后台导入</li>
                      <li>支持标签、描述、关键词、合集等元信息整理</li>
                      <li>支持 AI 生成标签与描述辅助</li>
                      <li>支持多存储目录、打开方式与存储规则管理</li>
                      <li>内置说明文档，方便交付与演示</li>
                    </ul>
                  </div>

                  <div class="about-action-row">
                    <div>
                      <h4>产品介绍与使用文档</h4>
                      <p>查看核心功能说明、推荐使用流程和页面介绍。</p>
                    </div>
                    <button class="btn primary doc-button" @click="handleViewDocumentation">查看文档</button>
                  </div>

                  <div class="app-links">
                    <button class="btn secondary" @click="handleViewLicense">查看许可</button>
                    <button class="btn secondary" @click="handleCheckUpdates">检查更新</button>
                  </div>
                </div>
              </section>
            </div>
          </section>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import AIConfigForm from '@/components/AIConfigForm.vue'
import AppHeader from '@/components/AppHeader.vue'
import CollectionsSettings from '@/components/CollectionsSettings.vue'
import OpenWithConfig from '@/components/OpenWithConfig.vue'
import StorageDirSelector from '@/components/StorageDirSelector.vue'
import StorageLayoutConfig from '@/components/StorageLayoutConfig.vue'
import TaskRetentionSettings from '@/components/TaskRetentionSettings.vue'
import ThemeConfig from '@/components/ThemeConfig.vue'
import { useFilesStore } from '@/store/modules/files'

type SettingId = 'storage' | 'openWith' | 'theme' | 'ai' | 'layout' | 'collections' | 'tasks' | 'about'

interface SettingItem {
  id: SettingId
  label: string
  description: string
  iconPaths: string[]
}

const filesStore = useFilesStore()

const router = useRouter()
const activeSetting = ref<SettingId>('storage')
const navCollapsed = ref(false)
const hasUserToggledNav = ref(false)

const settingIconPaths = {
  storage: [
    'M3 7.5A2.5 2.5 0 0 1 5.5 5h4l2 2h7A2.5 2.5 0 0 1 21 9.5v7A2.5 2.5 0 0 1 18.5 19h-13A2.5 2.5 0 0 1 3 16.5z',
  ],
  layout: [
    'M4 6h16',
    'M4 12h16',
    'M4 18h10',
    'M8 4v4',
    'M16 10v4',
    'M12 16v4',
  ],
  openWith: [
    'M8 5v14l11-7z',
    'M4 5v14',
  ],
  theme: [
    'M12 3a9 9 0 0 0 0 18h1.5a2 2 0 0 0 0-4H14a2 2 0 0 1 0-4h1a6 6 0 0 0-3-10z',
    'M7.5 10h.01',
    'M10 7.5h.01',
    'M14 7.5h.01',
  ],
  ai: [
    'M12 3v3',
    'M12 18v3',
    'M4.5 12h3',
    'M16.5 12h3',
    'M7.8 7.8l2.1 2.1',
    'M14.1 14.1l2.1 2.1',
    'M16.2 7.8l-2.1 2.1',
    'M9.9 14.1l-2.1 2.1',
    'M12 8.5a3.5 3.5 0 1 1 0 7 3.5 3.5 0 0 1 0-7z',
  ],
  collections: [
    'M4 6h16',
    'M4 12h16',
    'M4 18h10',
  ],
  tasks: [
    'M9 5H7a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-2',
    'M9 5a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2',
    'M12 12h.01',
    'M9 12h.01',
    'M15 12h.01',
  ],
  about: [
    'M12 17v-6',
    'M12 7h.01',
    'M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18z',
  ],
} satisfies Record<SettingId, string[]>

const settingItems: SettingItem[] = [
  { id: 'storage', label: '存储目录', description: '主目录与默认目录', iconPaths: settingIconPaths.storage },
  { id: 'layout', label: '存储规则', description: '归档层级与命名', iconPaths: settingIconPaths.layout },
  { id: 'openWith', label: '打开方式', description: '文件类型与应用', iconPaths: settingIconPaths.openWith },
  { id: 'theme', label: '主题设置', description: '外观、主色、背景', iconPaths: settingIconPaths.theme },
  { id: 'ai', label: 'AI 配置', description: '标签与描述辅助', iconPaths: settingIconPaths.ai },
  { id: 'collections', label: '合集', description: '管理与删除合集', iconPaths: settingIconPaths.collections },
  { id: 'tasks', label: '任务', description: '历史任务保留策略', iconPaths: settingIconPaths.tasks },
  { id: 'about', label: '文档与关于', description: '说明、许可、版本', iconPaths: settingIconPaths.about },
]

const activeSettingMeta = computed(() => {
  return settingItems.find((item) => item.id === activeSetting.value) || settingItems[0]
})

const syncNavigationWidth = () => {
  if (hasUserToggledNav.value) return

  // 窗口较窄时默认收起导航，把横向空间优先留给右侧配置表单。
  navCollapsed.value = window.innerWidth < 980
}

const toggleNavigation = () => {
  hasUserToggledNav.value = true
  navCollapsed.value = !navCollapsed.value
}

const handleDirAdded = () => {
  console.log('Master storage directory added')
}

const handleDirRemoved = () => {
  console.log('Master storage directory removed')
}

const handleDirDefaultChanged = () => {
  console.log('Default master storage directory changed')
}

const handleAIConfigSaved = () => {
  console.log('AI config saved')
}

const handleAIConfigReset = () => {
  console.log('AI config reset')
}

const handleThemeChanged = () => {
  console.log('Theme updated')
}

const handleThemeReset = () => {
  console.log('Theme reset')
}

const handleViewDocumentation = () => {
  router.push('/documentation')
}

const handleViewLicense = () => {
  window.alert('LocalSpace 当前以 MIT 风格开源协议分发，具体文本可在发布包中补充。')
}

const handleCheckUpdates = () => {
  window.alert('当前版本尚未接入在线更新服务，请关注后续发布说明。')
}

onMounted(() => {
  syncNavigationWidth()
  window.addEventListener('resize', syncNavigationWidth)
})

onUnmounted(() => {
  window.removeEventListener('resize', syncNavigationWidth)
})
</script>

<style scoped>
.settings-view {
  background: transparent;
}

.settings-view .page-content {
  overflow: hidden;
  background: transparent;
  padding: 6px 8px 8px;
}

.settings-view .page-stack {
  height: 100%;
  min-height: 0;
  /* 设置页贴近应用边框，并用轻透底色稳定背景图上的文字可读性。 */
  padding: 0;
}

.settings-workbench {
  --settings-divider: rgba(88, 103, 124, 0.2);
  --settings-muted-text: color-mix(in srgb, var(--text-soft) 88%, var(--text-color));

  position: relative;
  display: grid;
  grid-template-columns: 260px minmax(0, 1fr);
  gap: 0;
  height: 100%;
  min-height: 0;
  background: color-mix(in srgb, var(--surface-color) 52%, transparent);
  border: 1px solid color-mix(in srgb, var(--border-color) 34%, transparent);
  border-radius: 12px;
  color: var(--text-color);
  overflow: hidden;
  text-shadow: 0 1px 2px rgba(255, 255, 255, 0.24);
}

.settings-workbench.nav-collapsed {
  grid-template-columns: 66px minmax(0, 1fr);
}

[data-theme='dark'] .settings-workbench {
  --settings-divider: rgba(210, 224, 244, 0.18);
  --settings-muted-text: color-mix(in srgb, var(--text-soft) 86%, var(--text-color));

  background: color-mix(in srgb, var(--surface-color) 30%, transparent);
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.28);
}

.settings-sidebar,
.settings-detail {
  min-height: 0;
  border: none;
  background: transparent;
  box-shadow: none;
}

.settings-sidebar {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-right: 1px solid var(--settings-divider);
}

.sidebar-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  min-height: 64px;
  padding: 10px 14px;
  border-bottom: 1px solid var(--settings-divider);
}

.sidebar-title {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.sidebar-mark,
.app-logo {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 34px;
  height: 34px;
  border: none;
  border-radius: 0;
  background: transparent;
  color: var(--primary-color);
  font-size: 12px;
  font-weight: 800;
}

.sidebar-copy {
  display: grid;
  gap: 2px;
  min-width: 0;
}

.sidebar-copy strong {
  font-size: 15px;
  color: var(--text-color);
}

.sidebar-copy span {
  overflow: hidden;
  color: var(--settings-muted-text);
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collapse-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 30px;
  height: 30px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--text-soft);
  font-size: 20px;
  line-height: 1;
}

.collapse-button:hover {
  color: var(--primary-color);
  background: color-mix(in srgb, var(--primary-color) 8%, transparent);
}

.settings-nav {
  display: grid;
  gap: 4px;
  padding: 10px 12px;
}

.settings-nav-item {
  position: relative;
  display: grid;
  grid-template-columns: 34px minmax(0, 1fr);
  align-items: center;
  gap: 10px;
  width: 100%;
  min-height: 48px;
  padding: 7px 9px;
  border-radius: 6px;
  color: var(--text-soft);
  text-align: left;
}

.settings-nav-item:hover {
  background: color-mix(in srgb, var(--surface-color) 18%, transparent);
  color: var(--text-color);
}

.settings-nav-item.active {
  background: transparent;
  color: var(--text-color);
}

.settings-nav-item.active::before {
  position: absolute;
  top: 10px;
  bottom: 10px;
  left: 0;
  width: 3px;
  border-radius: 999px;
  background: var(--primary-color);
  content: '';
}

.nav-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 0;
  background: transparent;
  color: var(--text-soft);
}

.nav-icon svg {
  width: 17px;
  height: 17px;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 2;
}

.settings-workbench.nav-collapsed .nav-icon svg {
  width: 18px;
  height: 18px;
}

.settings-nav-item.active .nav-icon {
  background: transparent;
  color: var(--primary-color);
}

.nav-copy {
  display: grid;
  gap: 3px;
  min-width: 0;
}

.nav-copy span,
.nav-copy small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.nav-copy span {
  font-size: 13px;
  font-weight: 700;
}

.nav-copy small {
  color: var(--settings-muted-text);
  font-size: 11px;
}

.settings-workbench.nav-collapsed .sidebar-top {
  justify-content: center;
  padding: 12px 8px;
}

.settings-workbench.nav-collapsed .sidebar-copy,
.settings-workbench.nav-collapsed .nav-copy,
.settings-workbench.nav-collapsed .sidebar-mark {
  display: none;
}

.settings-workbench.nav-collapsed .settings-nav-item {
  grid-template-columns: 1fr;
  justify-items: center;
  padding: 7px;
}

.settings-detail {
  position: relative;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.settings-detail::before {
  position: absolute;
  inset: 0;
  background: linear-gradient(
    90deg,
    color-mix(in srgb, var(--surface-color) 24%, transparent),
    color-mix(in srgb, var(--surface-color) 36%, transparent) 34%,
    color-mix(in srgb, var(--surface-color) 24%, transparent)
  );
  content: '';
  pointer-events: none;
}

.settings-detail-header {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  min-height: 104px;
  padding: 18px 38px;
  border-bottom: 1px solid var(--settings-divider);
  background: transparent;
}

.detail-title-wrap {
  display: grid;
  gap: 5px;
  min-width: 0;
}

.detail-kicker {
  color: var(--primary-color);
  font-size: 11px;
  font-weight: 800;
  text-transform: uppercase;
}

.settings-detail-header h1 {
  color: var(--text-color);
  font-size: 23px;
  line-height: 1.25;
}

.settings-detail-header p {
  max-width: 560px;
  color: var(--settings-muted-text);
  font-size: 13px;
  line-height: 1.6;
}

.detail-nav-button {
  display: none;
  flex-shrink: 0;
  min-height: 34px;
  padding: 7px 11px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 8px;
  background: color-mix(in srgb, var(--surface-color) 14%, transparent);
  color: var(--text-soft);
  font-size: 12px;
  font-weight: 700;
}

.settings-panel {
  position: relative;
  z-index: 1;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 24px 38px 36px;
}

.settings-page {
  width: min(940px, 100%);
}

.settings-panel :deep(.storage-dir-selector),
.settings-panel :deep(.ai-config-form),
.settings-panel :deep(.open-with-config),
.settings-panel :deep(.storage-layout-config),
.settings-panel :deep(.theme-config) {
  padding: 0;
  background: transparent;
}

.settings-panel :deep(.selector-header),
.settings-panel :deep(.config-header) {
  display: none;
}

.settings-panel :deep(.config-content),
.settings-panel :deep(.storage-content),
.settings-panel :deep(.theme-content),
.settings-panel :deep(.config-fields) {
  display: grid;
  gap: 0;
}

.settings-panel :deep(.form-group),
.settings-panel :deep(.section-block),
.settings-panel :deep(.config-section),
.settings-panel :deep(.toggle-section),
.settings-panel :deep(.test-section),
.settings-panel :deep(.search-config-fields),
.settings-panel :deep(.disabled-hint),
.settings-panel :deep(.preview-card),
.settings-panel :deep(.rule-card),
.settings-panel :deep(.master-dir-card),
.settings-panel :deep(.theme-preview),
.settings-panel :deep(.no-background),
.settings-panel :deep(.current-background),
.settings-panel :deep(.empty-state) {
  border: none;
  border-bottom: 1px solid var(--settings-divider);
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

.settings-panel :deep(.form-group),
.settings-panel :deep(.toggle-section),
.settings-panel :deep(.test-section),
.settings-panel :deep(.disabled-hint),
.settings-panel :deep(.preview-card),
.settings-panel :deep(.theme-preview),
.settings-panel :deep(.no-background),
.settings-panel :deep(.current-background),
.settings-panel :deep(.empty-state) {
  padding: 16px 0;
}

.settings-panel :deep(.section-block),
.settings-panel :deep(.config-section) {
  padding: 18px 0;
}

.settings-panel :deep(.rule-card) {
  padding: 14px 0;
}

.settings-panel :deep(input),
.settings-panel :deep(textarea),
.settings-panel :deep(select) {
  background: color-mix(in srgb, var(--surface-color) 66%, transparent);
  border-color: color-mix(in srgb, var(--border-color) 72%, transparent);
  text-shadow: none;
}

.settings-panel :deep(label),
.settings-panel :deep(.label-text),
.settings-panel :deep(.section-title h5),
.settings-panel :deep(.toggle-info h5) {
  color: var(--text-color);
}

.settings-panel :deep(.subtitle),
.settings-panel :deep(.form-hint),
.settings-panel :deep(.label-sub),
.settings-panel :deep(.toggle-info p),
.settings-panel :deep(.section-title p),
.settings-panel :deep(.upload-hint) {
  color: var(--settings-muted-text);
}

.settings-panel :deep(.preview-card),
.settings-panel :deep(.rule-card),
.settings-panel :deep(.master-dir-card),
.settings-panel :deep(.theme-preview) {
  background: transparent;
}

.settings-panel :deep(.config-actions) {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  margin-top: 18px;
  padding-top: 16px;
  border-top: 1px solid var(--settings-divider);
}

.settings-panel :deep(.btn) {
  width: auto;
  text-shadow: none;
}

.settings-panel::-webkit-scrollbar {
  width: 8px;
}

.settings-panel::-webkit-scrollbar-track {
  background: transparent;
}

.settings-panel::-webkit-scrollbar-thumb {
  border: 2px solid transparent;
  border-radius: 999px;
  background: color-mix(in srgb, var(--text-faint) 34%, transparent);
  background-clip: padding-box;
}

.settings-panel::-webkit-scrollbar-thumb:hover {
  background: color-mix(in srgb, var(--text-faint) 52%, transparent);
  background-clip: padding-box;
}

.settings-panel :deep(.theme-preview .preview-card),
.settings-panel :deep(.path-editor),
.settings-panel :deep(.usage-preview),
.settings-panel :deep(.background-options),
.settings-panel :deep(.color-input-wrapper),
.settings-panel :deep(.input-wrapper) {
  background: transparent;
}

.about-content {
  display: grid;
  gap: 18px;
}

.app-info {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 16px;
}

.app-logo {
  width: 42px;
  height: 42px;
}

.app-info h3 {
  font-size: 20px;
  color: var(--text-color);
}

.app-version {
  margin-top: 4px;
  font-size: 14px;
  color: var(--text-faint);
}

.app-description p {
  color: var(--text-soft);
  line-height: 1.7;
}

.feature-list {
  list-style: none;
  margin-top: 14px;
}

.feature-list li {
  padding: 8px 0;
  border-bottom: 1px solid var(--border-color);
  font-size: 14px;
  color: var(--text-color);
}

.feature-list li:last-child {
  border-bottom: none;
}

.about-action-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 15px 0;
  border-top: 1px solid rgba(148, 163, 184, 0.18);
  border-bottom: 1px solid rgba(148, 163, 184, 0.18);
}

.about-action-row h4 {
  font-size: 16px;
  color: var(--text-color);
}

.about-action-row p {
  margin-top: 5px;
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-faint);
}

.doc-button {
  width: auto;
}

.app-links {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

@media (max-width: 980px) {
  .settings-workbench {
    grid-template-columns: 224px minmax(0, 1fr);
  }

  .settings-workbench.nav-collapsed {
    grid-template-columns: 60px minmax(0, 1fr);
  }

  .detail-nav-button {
    display: inline-flex;
  }
}

@media (max-width: 640px) {
  .settings-view .page-content {
    padding: 4px;
  }

  .settings-view .page-stack {
    padding: 0;
  }

  .settings-workbench {
    gap: 0;
    grid-template-columns: 206px minmax(0, 1fr);
    border-radius: 10px;
  }

  .settings-workbench.nav-collapsed {
    grid-template-columns: 56px minmax(0, 1fr);
  }

  .settings-detail-header {
    align-items: flex-start;
    flex-direction: column;
    min-height: 0;
    padding: 16px;
  }

  .settings-panel {
    padding: 16px;
  }

  .about-action-row {
    flex-direction: column;
    align-items: flex-start;
  }

  .app-links {
    grid-template-columns: 1fr;
  }
}
</style>

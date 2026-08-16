<template>
  <div class="page-shell settings-view">
    <AppHeader />

    <div class="page-content">
      <div class="page-stack">
        <h1 class="visually-hidden">设置</h1>

        <div class="set-workbench" :class="{ 'nav-collapsed': navCollapsed }">
          <!-- 导航沿用说明页的页内目录写法：画布上的列表 + 左侧色条，分组名做小标签。 -->
          <aside class="set-nav" aria-label="设置导航">
            <nav aria-label="设置分类导航">
              <section v-for="group in settingGroups" :key="group.id" class="set-nav-group">
                <span class="set-nav-label">{{ group.label }}</span>
                <button
                  v-for="item in group.items"
                  :key="item.id"
                  type="button"
                  class="set-nav-item"
                  :class="{ active: activeSetting === item.id }"
                  :title="item.label"
                  :aria-current="activeSetting === item.id ? 'page' : undefined"
                  @click="activateSetting(item)"
                >
                  <svg
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="1.7"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    aria-hidden="true"
                  >
                    <path v-for="path in item.iconPaths" :key="path" :d="path" />
                  </svg>
                  <span>{{ item.label }}</span>
                </button>
              </section>
            </nav>

            <button
              type="button"
              class="set-nav-collapse"
              :title="navCollapsed ? '展开设置导航' : '收起设置导航'"
              :aria-label="navCollapsed ? '展开设置导航' : '收起设置导航'"
              :aria-expanded="String(!navCollapsed)"
              @click="toggleNavigation"
            >
              <i aria-hidden="true">{{ navCollapsed ? '›' : '‹' }}</i>
              <span>收起设置导航</span>
            </button>
          </aside>

          <div class="set-detail">
            <!-- 当前分区就是页头：左导航和面包屑已经写过「设置」，标题再写一次是冗余。 -->
            <header class="set-head">
              <div>
                <h2>{{ activeItem.label }}</h2>
                <p>{{ activeItem.context }}</p>
              </div>
              <span v-if="activeItem.instantNote" class="set-head-note">
                {{ activeItem.instantNote }}
              </span>
            </header>

            <section
              v-if="activeSetting === 'storage'"
              class="set-panel"
              role="tabpanel"
              aria-label="存储目录"
            >
              <StorageDirSelector
                @master-dir-added="handleDirAdded"
                @master-dir-removed="handleDirRemoved"
                @master-dir-default-changed="handleDirDefaultChanged"
              />
            </section>

            <section
              v-else-if="activeSetting === 'layout'"
              class="set-panel"
              role="tabpanel"
              aria-label="存储规则"
            >
              <StorageLayoutConfig />
            </section>

            <section
              v-else-if="activeSetting === 'openWith'"
              class="set-panel"
              role="tabpanel"
              aria-label="打开方式"
            >
              <OpenWithConfig />
            </section>

            <section
              v-else-if="activeSetting === 'theme'"
              class="set-panel"
              role="tabpanel"
              aria-label="主题"
            >
              <ThemeConfig @theme-changed="handleThemeChanged" @theme-reset="handleThemeReset" />
            </section>

            <section
              v-else-if="activeSetting === 'ai'"
              class="set-panel"
              role="tabpanel"
              aria-label="AI"
            >
              <AIConfigForm @config-saved="handleAIConfigSaved" @config-reset="handleAIConfigReset" />
            </section>

            <section
              v-else-if="activeSetting === 'collections'"
              class="set-panel"
              role="tabpanel"
              aria-label="合集"
            >
              <CollectionsSettings />
            </section>

            <section
              v-else-if="activeSetting === 'tasks'"
              class="set-panel"
              role="tabpanel"
              aria-label="任务"
            >
              <TaskRetentionSettings />
            </section>

            <section v-else class="set-panel" role="tabpanel" aria-label="关于">
              <div class="set-about-id">
                <span class="set-about-mark" aria-hidden="true">LS</span>
                <div>
                  <h3>LocalSpace</h3>
                  <p>版本 0.3.0</p>
                </div>
              </div>

              <p class="set-about-body">
                LocalSpace 是一个面向本地资料整理与回看的桌面工具，适合把图片、文档、音视频和项目素材统一收进一个入口。
              </p>

              <ul class="set-features">
                <li>支持单文件导入与批量后台导入</li>
                <li>支持标签、描述、关键词、合集等元信息整理</li>
                <li>支持 AI 生成标签与描述辅助</li>
                <li>支持多存储目录、打开方式与存储规则管理</li>
                <li>内置说明文档，方便交付与演示</li>
              </ul>

              <div class="set-section">
                <div class="set-doc-row">
                  <div>
                    <strong>产品介绍与使用文档</strong>
                    <p>查看核心功能说明、推荐使用流程和页面介绍。</p>
                  </div>
                  <button type="button" class="btn primary" @click="handleViewDocumentation">
                    查看文档
                  </button>
                </div>
                <div class="set-add-row set-about-links">
                  <button type="button" class="btn secondary" @click="handleViewLicense">
                    查看许可
                  </button>
                  <button type="button" class="btn secondary" @click="handleCheckUpdates">
                    检查更新
                  </button>
                </div>
              </div>
            </section>
          </div>
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

type SettingId = 'storage' | 'openWith' | 'theme' | 'ai' | 'layout' | 'collections' | 'tasks' | 'about'

interface SettingItem {
  id: SettingId
  label: string
  /** 分区标题下的一句说明，作为页面副标题使用。 */
  context: string
  /** 没有保存按钮的分区在页头写明即时生效，避免用户找一个不存在的按钮。 */
  instantNote?: string
  iconPaths: string[]
}

interface SettingGroup {
  id: 'general' | 'theme' | 'ai' | 'special' | 'about'
  label: string
  items: SettingItem[]
}

const router = useRouter()
const activeSetting = ref<SettingId>('storage')
const navCollapsed = ref(false)
const hasUserToggledNav = ref(false)

const circle = 'M12 3.6a8.4 8.4 0 1 1 0 16.8 8.4 8.4 0 0 1 0-16.8z'

const settingIconPaths = {
  storage: ['M3 7.2A1.5 1.5 0 0 1 4.5 5.7h4l1.9 2.3h9.1v9.8a1.5 1.5 0 0 1-1.5 1.5h-14a1.5 1.5 0 0 1-1.5-1.5z'],
  layout: ['M5.2 4.4h13.6a2 2 0 0 1 2 2v1.6a2 2 0 0 1-2 2H5.2a2 2 0 0 1-2-2V6.4a2 2 0 0 1 2-2z', 'M5.2 14h6.5a2 2 0 0 1 2 2v1.6a2 2 0 0 1-2 2H5.2a2 2 0 0 1-2-2V16a2 2 0 0 1 2-2z'],
  openWith: ['M14 4.2h5.8V10', 'M19.8 4.2 11.8 12.2', 'M17.6 13.6v5a1.6 1.6 0 0 1-1.6 1.6H5.6A1.6 1.6 0 0 1 4 18.6V8.2a1.6 1.6 0 0 1 1.6-1.6h5'],
  theme: [circle, 'M12 3.6v16.8', 'M15 8.6h3', 'M16 12h3.2', 'M15 15.4h3'],
  ai: ['M10.4 3.4 12 7.9l4.5 1.6-4.5 1.6-1.6 4.5-1.6-4.5L4.3 9.5 8.8 7.9z', 'M17.4 14.2l.85 2.35 2.35.85-2.35.85-.85 2.35-.85-2.35-2.35-.85 2.35-.85z'],
  collections: ['M3.6 3.6h7v7h-7z', 'M13.4 3.6h7v7h-7z', 'M3.6 13.4h7v7h-7z', 'M13.4 13.4h7v7h-7z'],
  tasks: [circle, 'M12 7.2V12l3.2 1.9'],
  about: [circle, 'M12 11.2v5', 'M12 7.7h.01'],
} satisfies Record<SettingId, string[]>

const settingItems: SettingItem[] = [
  { id: 'storage', label: '存储目录', context: '配置本地资料的入口目录，导入的文件会落在这里。', instantNote: '改动立即生效', iconPaths: settingIconPaths.storage },
  { id: 'layout', label: '存储规则', context: '控制导入后在主目录中的落盘层级结构。', iconPaths: settingIconPaths.layout },
  { id: 'openWith', label: '打开方式', context: '指定常用文件在 LocalSpace 内优先使用的软件。', iconPaths: settingIconPaths.openWith },
  { id: 'theme', label: '主题', context: '调整界面明暗、强调色与背景图片。', instantNote: '修改后立即生效', iconPaths: settingIconPaths.theme },
  { id: 'ai', label: 'AI', context: '配置模型、密钥与工具增强分析能力。', iconPaths: settingIconPaths.ai },
  { id: 'collections', label: '合集', context: '维护可复用的文件分组。', instantNote: '改动立即生效', iconPaths: settingIconPaths.collections },
  { id: 'tasks', label: '任务', context: '管理任务记录的保留时长和清理策略。', iconPaths: settingIconPaths.tasks },
  { id: 'about', label: '关于', context: '查看产品信息、许可与文档入口。', iconPaths: settingIconPaths.about },
]

const pickItems = (...ids: SettingId[]) =>
  ids.map((id) => settingItems.find((item) => item.id === id)!)

// 分组只做标签，不再折叠：它说明「存储目录 / 存储规则 / 打开方式」属于同一族，
// 这条信息值几个字，但不值一层交互。
const settingGroups: SettingGroup[] = [
  { id: 'general', label: '通用配置', items: pickItems('storage', 'layout', 'openWith') },
  { id: 'theme', label: '主题', items: pickItems('theme') },
  { id: 'ai', label: 'AI 配置', items: pickItems('ai') },
  { id: 'special', label: '专项配置', items: pickItems('collections', 'tasks') },
  { id: 'about', label: '关于', items: pickItems('about') },
]

const activeItem = computed(
  () => settingItems.find((item) => item.id === activeSetting.value) ?? settingItems[0],
)

const activateSetting = (item: SettingItem) => {
  activeSetting.value = item.id
}

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

<!-- set- 前缀的共用装置（设置行、控件、色片、对话框）由八个分区共享，放在单独文件里。 -->
<style src="../assets/styles/settings.css"></style>

<style scoped>
.settings-view .page-content {
  padding: 20px 44px 50px;
}

.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  margin: -1px;
  padding: 0;
  overflow: hidden;
  border: 0;
  clip: rect(0 0 0 0);
  white-space: nowrap;
}

/* 关于是内容型分区，允许比设置表单松一档。 */
.set-about-id {
  display: flex;
  align-items: center;
  gap: 15px;
}

.set-about-mark {
  display: grid;
  place-items: center;
  width: 52px;
  height: 52px;
  border-radius: 15px;
  background: linear-gradient(145deg, var(--primary-color), var(--primary-hover));
  box-shadow: 0 9px 18px var(--shadow-color);
  color: #fff;
  font-size: 18px;
  font-weight: 700;
}

.set-about-id h3 {
  color: var(--text-color);
  font-size: 21px;
  font-weight: 700;
  letter-spacing: -0.03em;
}

.set-about-id p {
  margin-top: 5px;
  color: var(--text-faint);
  font-size: 12px;
}

.set-about-body {
  max-width: 620px;
  margin-top: 18px;
  color: var(--text-soft);
  font-size: 13px;
  line-height: 1.85;
}

.set-features {
  display: grid;
  gap: 7px;
  margin: 16px 0 0;
  padding: 0;
  list-style: none;
}

.set-features li {
  position: relative;
  padding: 10px 13px 10px 28px;
  border-radius: 9px;
  background: color-mix(in srgb, var(--surface-color) 58%, transparent);
  color: var(--text-soft);
  font-size: 12px;
  line-height: 1.6;
}

.set-features li::before {
  position: absolute;
  top: 17px;
  left: 13px;
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--primary-color);
  content: '';
}

.set-doc-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 15px 16px;
  border-left: 3px solid var(--primary-color);
  border-radius: 0 var(--radius-md) var(--radius-md) 0;
  background: color-mix(in srgb, var(--primary-color) 9%, transparent);
}

.set-doc-row strong {
  display: block;
  color: var(--text-color);
  font-size: 13px;
}

.set-doc-row p {
  margin-top: 5px;
  color: var(--text-soft);
  font-size: 12px;
}

.set-about-links {
  margin-top: 12px;
}

/* 关于分区的最后一块紧跟在特性列表后面，需要自己拉开间距。 */
.set-panel > .set-section {
  margin-top: 24px;
}

@media (max-width: 820px) {
  .settings-view .page-content {
    padding: 16px 12px 40px;
  }

  .set-doc-row {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>

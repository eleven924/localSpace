<template>
  <div class="documentation-view">
    <header class="header">
      <div class="header-left">
        <router-link to="/settings" class="back-link">
          <span class="back-icon">←</span>
          返回设置
        </router-link>
        <div class="header-info">
          <h1>LocalSpace 介绍与使用文档</h1>
          <p class="subtitle">随应用一起打包，可在桌面版中离线查看。</p>
        </div>
      </div>
    </header>

    <div ref="contentRef" class="content">
      <div class="layout">
        <aside class="toc">
          <div class="toc-card">
            <h2>目录</h2>
            <button
              v-for="section in sections"
              :key="section.id"
              type="button"
              class="toc-link"
              :class="{ active: activeSectionId === section.id }"
              @click="scrollToSection(section.id)"
            >
              <span class="toc-index">{{ section.index }}</span>
              <span>{{ section.title }}</span>
            </button>
          </div>
        </aside>

        <main class="document">
          <section class="hero">
            <p class="eyebrow">产品概览</p>
            <h2>围绕本地资料整理、检索和复用打造的轻量桌面工作台</h2>
            <p class="hero-text">
              LocalSpace 用于管理本地文件资料，支持导入归档、标签描述、搜索筛选、
              缩略图预览、AI 辅助标注以及目录与主题配置，适合个人资料库和项目素材库使用。
            </p>
            <div class="hero-stats">
              <div class="stat-card">
                <span class="stat-value">4+</span>
                <span class="stat-label">核心工作流</span>
              </div>
              <div class="stat-card">
                <span class="stat-value">6</span>
                <span class="stat-label">主要功能模块</span>
              </div>
              <div class="stat-card">
                <span class="stat-value">内置</span>
                <span class="stat-label">离线文档查看</span>
              </div>
            </div>
          </section>

          <section
            v-for="section in sections"
            :id="section.id"
            :key="section.id"
            :ref="(el) => setSectionRef(section.id, el)"
            class="doc-section"
          >
            <div class="section-heading">
              <span class="section-index">{{ section.index }}</span>
              <div>
                <h2>{{ section.title }}</h2>
                <p>{{ section.summary }}</p>
              </div>
            </div>

            <div v-if="section.paragraphs?.length" class="paragraphs">
              <p v-for="paragraph in section.paragraphs" :key="paragraph">
                {{ paragraph }}
              </p>
            </div>

            <ul v-if="section.bullets?.length" class="bullet-list">
              <li v-for="bullet in section.bullets" :key="bullet">
                {{ bullet }}
              </li>
            </ul>

            <ol v-if="section.steps?.length" class="steps-list">
              <li v-for="step in section.steps" :key="step.title">
                <strong>{{ step.title }}</strong>
                <span>{{ step.description }}</span>
              </li>
            </ol>

            <div v-if="section.screenshots?.length" class="screenshot-grid">
              <figure
                v-for="shot in section.screenshots"
                :key="shot.src"
                class="shot-card"
              >
                <img :src="shot.src" :alt="shot.alt" />
                <figcaption>
                  <strong>{{ shot.title }}</strong>
                  <span>{{ shot.caption }}</span>
                </figcaption>
              </figure>
            </div>
          </section>
        </main>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

interface SectionStep {
  title: string
  description: string
}

interface SectionScreenshot {
  src: string
  alt: string
  title: string
  caption: string
}

interface DocSection {
  id: string
  index: string
  title: string
  summary: string
  paragraphs?: string[]
  bullets?: string[]
  steps?: SectionStep[]
  screenshots?: SectionScreenshot[]
}

const sections: DocSection[] = [
  {
    id: 'core-features',
    index: '01',
    title: '软件核心功能',
    summary: '这一部分对应当前项目已经落地的主要能力，聚焦导入、整理、搜索、展示与配置。',
    bullets: [
      '文件导入与归档：从本地选择文件后导入 LocalSpace，并记录名称、描述、标签和关键词。',
      '统一索引与浏览：在主列表中按文件类型筛选，集中查看已导入文件的缩略图、元信息和状态。',
      '搜索与检索：支持按关键字搜索，并结合描述、标签和文件类型快速定位资料。',
      'AI 元数据辅助：可在设置中配置模型参数，为文件自动生成标签与描述。',
      '多主目录管理：支持设置主存储目录、默认目录和容量约束，便于规划资料落盘结构。',
      '缩略图与元信息提取：针对图片、文档、视频等文件生成预览并提取基础元信息。'
    ],
    screenshots: [
      {
        src: '/docs/home-overview.png',
        alt: 'LocalSpace 主列表页截图',
        title: '主列表页',
        caption: '主界面聚合文件类型筛选、搜索和文件列表，是日常检索与回看资料的主要入口。'
      }
    ]
  },
  {
    id: 'quick-start',
    index: '02',
    title: '推荐上手流程',
    summary: '如果是第一次使用，建议按下面的顺序完成初始化。',
    steps: [
      {
        title: '先配置主存储目录',
        description: '打开“设置”，在“存储目录”中新增主目录，并根据容量规划决定是否设置默认目录。'
      },
      {
        title: '按需开启 AI 配置',
        description: '如果你希望导入时自动生成标签与描述，可填写 API Key、模型名、Base URL 等参数。'
      },
      {
        title: '导入第一批文件',
        description: '进入“导入文件”页面，选择本地文件后补充名称、标签、描述与关键词，再执行导入。'
      },
      {
        title: '回到文件列表检查结果',
        description: '在主列表中按类型筛选、搜索和打开文件，确认目录规划与元数据是否符合预期。'
      }
    ],
    screenshots: [
      {
        src: '/docs/import-workflow.png',
        alt: 'LocalSpace 导入页面截图',
        title: '导入流程页',
        caption: '导入页聚焦单文件处理，适合在落库前补全描述、标签和关键词。'
      }
    ]
  },
  {
    id: 'manual',
    index: '03',
    title: '使用手册',
    summary: '下面按当前页面路径整理成可直接操作的说明。',
    paragraphs: [
      '进入“文件”页后，可以通过顶部按钮进入导入页或设置页；列表区域支持按文件类型筛选，并结合搜索栏快速查找资料。',
      '在“导入文件”页，先选择本地文件，系统会尝试识别文件类型。随后在元数据表单中补充标签、描述与关键词，完成后导入。',
      '在文件列表中点击文件可触发打开；如果后续文件在系统外被修改，页面会周期性检查并尝试刷新显示内容。',
      '在“设置”页中，存储目录负责管理落盘位置，AI 配置负责自动标签与描述生成，主题设置负责界面风格个性化。'
    ],
    steps: [
      {
        title: '查看文件',
        description: '进入主列表，使用文件类型筛选和搜索框缩小范围，再从文件卡片打开或定位文件。'
      },
      {
        title: '管理目录',
        description: '在设置页新增主目录，避免路径冲突，并根据使用习惯指定默认目录。'
      },
      {
        title: '优化搜索效果',
        description: '导入时尽量补全标签、描述和关键词，能显著提升后续模糊检索的命中率。'
      },
      {
        title: '启用 AI',
        description: '确认模型服务可用后再开启 AI 配置，先用少量样本验证生成质量。'
      }
    ]
  },
  {
    id: 'settings-docs',
    index: '04',
    title: '设置中的文档入口',
    summary: '本次新增了“设置 -> 文档与关于 -> 查看文档”入口，方便在发布版中直接查看说明。',
    paragraphs: [
      '文档页面属于前端静态资源的一部分，构建后会进入 frontend/dist，再由 Wails 使用 embed 打包进桌面程序，因此最终生成的 exe 可以离线查看文档。',
      '为了增强可读性，文档页除了结构化说明外，还会展示主界面、导入页和设置页截图，帮助用户快速建立操作心智。'
    ],
    screenshots: [
      {
        src: '/docs/settings-documentation-entry.png',
        alt: '设置页中的文档入口截图',
        title: '设置页文档入口',
        caption: '用户可从设置页直接进入文档，不需要额外打开外部网页或独立说明文件。'
      }
    ]
  },
  {
    id: 'tips',
    index: '05',
    title: '使用建议',
    summary: '这些建议有助于让资料库更稳定、更容易检索。',
    bullets: [
      '主目录尽量放在容量充足、路径稳定的位置，避免频繁迁移。',
      '如果资料类型很多，建议在导入时统一标签命名规则，例如按项目名、年份或主题分组。',
      '首次启用 AI 配置时先用少量文件试跑，观察标签和描述是否符合你的检索习惯。',
      '对于常用文件，可优先关注缩略图与描述质量，它们决定了列表浏览时的辨识效率。'
    ]
  }
]

const contentRef = ref<HTMLElement | null>(null)
const activeSectionId = ref(sections[0].id)
const sectionRefs = new Map<string, HTMLElement>()

const setSectionRef = (id: string, el: Element | null) => {
  if (el instanceof HTMLElement) {
    sectionRefs.set(id, el)
  } else {
    sectionRefs.delete(id)
  }
}

const scrollToSection = (id: string) => {
  const target = sectionRefs.get(id)
  if (!target) return

  activeSectionId.value = id
  target.scrollIntoView({
    behavior: 'smooth',
    block: 'start',
  })
}

const updateActiveSection = () => {
  const container = contentRef.value
  if (!container) return

  const containerTop = container.getBoundingClientRect().top + 120
  let currentSectionId = sections[0].id

  for (const section of sections) {
    const element = sectionRefs.get(section.id)
    if (!element) continue

    const rect = element.getBoundingClientRect()
    if (rect.top <= containerTop) {
      currentSectionId = section.id
    }
  }

  activeSectionId.value = currentSectionId
}

onMounted(() => {
  contentRef.value?.addEventListener('scroll', updateActiveSection, { passive: true })
  updateActiveSection()
})

onBeforeUnmount(() => {
  contentRef.value?.removeEventListener('scroll', updateActiveSection)
})
</script>

<style scoped>
.documentation-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  background-color: var(--app-bg-color, var(--bg-color));
}

.header {
  display: flex;
  align-items: center;
  padding: 14px 20px;
  border-bottom: 1px solid var(--border-color);
  background-color: var(--surface-color);
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 14px;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 11px;
  border-radius: 8px;
  background-color: var(--bg-color);
  color: var(--text-color);
  border: 1px solid var(--border-color);
}

.back-icon {
  font-size: 14px;
}

.header-info h1 {
  margin: 0 0 4px 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-color);
}

.subtitle {
  margin: 0;
  font-size: 12px;
  line-height: 1.5;
  color: var(--text-color);
  opacity: 0.68;
}

.content {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

.layout {
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr);
  gap: 20px;
  width: min(1220px, calc(100% - 32px));
  margin: 0 auto;
  padding: 18px 0 24px;
}

.toc {
  position: relative;
}

.toc-card {
  position: sticky;
  top: 18px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 16px;
  border: 1px solid var(--border-color);
  border-radius: 12px;
  background-color: var(--surface-color);
}

.toc-card h2 {
  margin: 0 0 4px 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color);
}

.toc-link {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 9px 10px;
  border: 1px solid transparent;
  border-radius: 8px;
  background-color: transparent;
  color: var(--text-color);
  text-align: left;
  font-size: 13px;
}

.toc-link:hover {
  background-color: var(--bg-color);
  border-color: var(--border-color);
}

.toc-link.active {
  background-color: rgba(33, 150, 243, 0.08);
  border-color: rgba(33, 150, 243, 0.18);
  color: var(--primary-color);
}

.toc-index {
  flex-shrink: 0;
  width: 28px;
  font-size: 11px;
  font-weight: 700;
}

.document {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.hero,
.doc-section {
  scroll-margin-top: 18px;
  padding: 18px;
  border: 1px solid var(--border-color);
  border-radius: 12px;
  background-color: var(--surface-color);
}

.hero {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.eyebrow {
  display: inline-block;
  width: fit-content;
  padding: 4px 8px;
  border-radius: 999px;
  background-color: rgba(33, 150, 243, 0.08);
  color: var(--primary-color);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.hero h2 {
  margin: 0;
  font-size: 24px;
  line-height: 1.4;
  font-weight: 600;
  color: var(--text-color);
}

.hero-text {
  margin: 0;
  font-size: 13px;
  line-height: 1.8;
  color: var(--text-color);
  opacity: 0.82;
}

.hero-stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.stat-card {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 14px;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  background-color: var(--bg-color);
}

.stat-value {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-color);
}

.stat-label {
  font-size: 12px;
  color: var(--text-color);
  opacity: 0.68;
}

.section-heading {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr);
  gap: 12px;
  margin-bottom: 14px;
}

.section-index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 10px;
  background-color: rgba(33, 150, 243, 0.08);
  color: var(--primary-color);
  font-size: 12px;
  font-weight: 700;
}

.section-heading h2 {
  margin: 0 0 6px 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-color);
}

.section-heading p,
.paragraphs p {
  margin: 0;
  font-size: 13px;
  line-height: 1.8;
  color: var(--text-color);
  opacity: 0.8;
}

.paragraphs {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.bullet-list,
.steps-list {
  display: grid;
  gap: 10px;
  margin: 0;
  padding-left: 20px;
}

.bullet-list li,
.steps-list li {
  font-size: 13px;
  line-height: 1.8;
  color: var(--text-color);
}

.steps-list li {
  display: grid;
  gap: 3px;
}

.steps-list strong {
  font-size: 13px;
  font-weight: 600;
}

.steps-list span {
  font-size: 13px;
  color: var(--text-color);
  opacity: 0.8;
}

.screenshot-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 14px;
  margin-top: 14px;
}

.shot-card {
  margin: 0;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  overflow: hidden;
  background-color: var(--bg-color);
}

.shot-card img {
  display: block;
  width: 100%;
  height: auto;
  object-fit: contain;
  background-color: #fff;
}

.shot-card figcaption {
  display: grid;
  gap: 5px;
  padding: 12px;
}

.shot-card strong {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-color);
}

.shot-card span {
  font-size: 12px;
  line-height: 1.7;
  color: var(--text-color);
  opacity: 0.76;
}

@media (max-width: 1024px) {
  .layout {
    grid-template-columns: 1fr;
  }

  .toc-card {
    position: static;
  }
}

@media (max-width: 768px) {
  .header {
    padding: 12px 16px;
  }

  .header-left {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }

  .layout {
    width: calc(100% - 24px);
    padding: 12px 0 20px;
    gap: 14px;
  }

  .hero,
  .doc-section {
    padding: 16px;
  }

  .hero h2 {
    font-size: 20px;
  }

  .hero-stats {
    grid-template-columns: 1fr;
  }

  .section-heading {
    grid-template-columns: 1fr;
  }
}
</style>

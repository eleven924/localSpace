<template>
  <div class="documentation-view">
    <AppHeader />

    <div ref="contentRef" class="content">
      <section class="page-intro">
        <div class="page-intro-copy">
          <router-link to="/settings" class="doc-back-link">返回设置</router-link>
          <p class="eyebrow">内置文档</p>
          <h1>LocalSpace 介绍与使用文档</h1>
          <p class="subtitle">随应用一起打包，可以在桌面版中离线查看，适合交付给最终使用者直接阅读。</p>
        </div>

        <p class="intro-meta">你可以把这里理解成产品手册、操作说明和展示页的结合体。</p>
      </section>

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
          <section class="hero-shell">
            <div class="hero-copy">
              <p class="eyebrow">产品概览</p>
              <h2>把本地资料整理、检索和复用这条链路放到一个轻量桌面工作台里</h2>
              <p class="hero-text">
                LocalSpace 用于管理本地文件资料，支持导入归档、标签描述、搜索筛选、缩略图预览、AI
                辅助标注，以及目录、打开方式和主题配置，适合个人资料库和项目素材库使用。
              </p>

              <div class="hero-highlights">
                <div class="highlight-card">
                  <span class="highlight-label">本地优先</span>
                  <strong>资料不需要依赖网页后台就能整理和回看</strong>
                </div>
                <div class="highlight-card">
                  <span class="highlight-label">AI 辅助</span>
                  <strong>在导入阶段补足标签与描述，提升后续搜索效率</strong>
                </div>
              </div>
            </div>

            <div class="hero-aside">
              <div class="hero-note">
                <span class="hero-note-label">适合什么场景</span>
                <p>个人知识库、课程资料、项目素材、视频剪辑素材、截图归档、文档与图片收纳。</p>
              </div>

              <div class="hero-stats">
                <div class="stat-card">
                  <span class="stat-value">3</span>
                  <span class="stat-label">核心入口页面</span>
                </div>
                <div class="stat-card">
                  <span class="stat-value">4</span>
                  <span class="stat-label">文档章节</span>
                </div>
                <div class="stat-card">
                  <span class="stat-value">离线</span>
                  <span class="stat-label">可直接查看</span>
                </div>
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
            <div class="section-surface">
              <div class="section-heading">
                <span class="section-index">{{ section.index }}</span>
                <div class="section-heading-copy">
                  <span class="section-kicker">Documentation</span>
                  <h2>{{ section.title }}</h2>
                  <p>{{ section.summary }}</p>
                </div>
              </div>

              <div v-if="section.paragraphs?.length" class="paragraphs">
                <p v-for="paragraph in section.paragraphs" :key="paragraph">
                  {{ paragraph }}
                </p>
              </div>

              <ul v-if="section.bullets?.length" class="bullet-grid">
                <li v-for="bullet in section.bullets" :key="bullet" class="bullet-card">
                  <span class="bullet-dot"></span>
                  <span>{{ bullet }}</span>
                </li>
              </ul>

              <ol v-if="section.steps?.length" class="steps-grid">
                <li v-for="step in section.steps" :key="step.title" class="step-card">
                  <span class="step-badge">{{ step.title }}</span>
                  <p>{{ step.description }}</p>
                </li>
              </ol>

              <div v-if="section.screenshots?.length" class="screenshot-grid">
                <figure v-for="shot in section.screenshots" :key="shot.src" class="shot-card">
                  <div class="shot-frame">
                    <img :src="shot.src" :alt="shot.alt" />
                  </div>
                  <figcaption>
                    <span class="shot-meta">界面截图</span>
                    <strong>{{ shot.title }}</strong>
                    <span>{{ shot.caption }}</span>
                  </figcaption>
                </figure>
              </div>
            </div>
          </section>
        </main>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import AppHeader from '@/components/AppHeader.vue'

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
    id: 'basic-intro',
    index: '01',
    title: '软件基本介绍',
    summary: 'LocalSpace 的定位不是单纯选文件，而是把本地资料整理成可搜索、可回看、可持续维护的资料库。',
    paragraphs: [
      'LocalSpace 是一款“本地文件管理 + AI 辅助整理”的桌面应用，适合个人资料库、学习资料库和项目素材库。',
      '它把导入、归档、检索、打开和持续整理串成一条完整链路，让图片、文档、音视频、压缩包、安装包等资料都能归到同一个入口中。',
      '当你配置了 AI 后，还可以在导入时获得标签和描述建议，降低整理成本，并提升后续搜索命中率。'
    ],
    bullets: [
      '本地优先：文件由本地桌面应用管理，适合长期沉淀个人资料。',
      '统一入口：不同类型文件集中管理，减少在多个文件夹之间反复切换。',
      'AI 辅助：在导入阶段补全标签与描述，帮助资料更容易被找到。',
      '可持续整理：支持存储目录、打开方式、存储规则、主题与内置文档配置。'
    ]
  },
  {
    id: 'features',
    index: '02',
    title: '功能说明',
    summary: '当前版本的核心能力主要集中在文件展示、文件导入和设置三个维度。',
    paragraphs: [
      '文件展示页是日常使用的主入口，支持按全部、视频、文档、音乐、压缩包、安装包、图片等类型切换查看，并结合搜索栏查找文件名、标签和描述。',
      '导入页负责把本地文件正式纳入资料库结构中。用户先选文件，再补充名称、标签、关键词、描述和合集等信息，最后再执行导入。',
      '设置页负责软件初始化和长期使用中的关键配置，包含存储目录、打开方式、AI 配置、存储规则、主题设置以及文档与关于。'
    ],
    bullets: [
      '文件展示：支持类型筛选、搜索、缩略图与元信息浏览，并可直接打开文件。',
      '文件导入：支持类型识别、元数据补充，以及导入前的资料整理。',
      'AI 辅助：在导入流程中生成标签与描述建议，再由用户决定是否采用。',
      '设置中心：统一管理目录、打开方式、存储策略、主题和说明文档。'
    ],
    screenshots: [
      {
        src: '/docs/home-overview.png',
        alt: 'LocalSpace 主列表页截图',
        title: '主列表页',
        caption: '新版首页已经包含主导航、类型筛选、搜索栏、结果摘要和展示模式切换，是日常检索与回看资料的主要入口。'
      },
      {
        src: '/docs/import-workflow.png',
        alt: 'LocalSpace 导入页截图',
        title: '导入流程页',
        caption: '新版导入页增加了顶部导入说明卡片，更像一个完整的导入中心，适合在落库前补全描述、标签、关键词和合集信息。'
      },
      {
        src: '/docs/settings-documentation-entry.png',
        alt: '设置页中的文档入口截图',
        title: '设置页',
        caption: '新版设置页已经把存储目录、打开方式、AI 配置、存储规则、主题设置和文档入口完整集中到一个页面中。'
      }
    ]
  },
  {
    id: 'user-guide',
    index: '03',
    title: '用户指南',
    summary: '首次使用时，建议按固定顺序完成初始化，这样最容易建立清晰的使用心智。',
    steps: [
      {
        title: '先配置主存储目录',
        description: '打开“设置”，在“存储目录”中新增主目录，并根据容量规划决定是否设置默认目录。'
      },
      {
        title: '补充打开方式和存储规则',
        description: '如果你希望文件按更清晰的层级归档，或用固定软件打开文件，可以继续完善这两部分设置。'
      },
      {
        title: '按需开启 AI 配置',
        description: '如果你希望导入时自动生成标签与描述，可以填写 API Key、模型名、Base URL 等参数。'
      },
      {
        title: '导入第一批文件',
        description: '进入“导入文件”页面，选择本地文件后补充名称、标签、关键词、描述和合集，再执行导入。'
      },
      {
        title: '按需使用 AI 建议',
        description: '在导入表单中生成标签和描述建议，确认内容合适后再采用，避免无效元数据进入资料库。'
      },
      {
        title: '回到文件列表检查结果',
        description: '在主列表中按类型筛选、搜索和打开文件，确认目录规划与元数据是否符合预期。'
      }
    ],
    paragraphs: [
      '如果是交付给终端用户使用，最重要的是先教会他们配置存储目录，再决定是否启用 AI。',
      'AI 的主要使用位置在导入流程中，而不是文件列表页。也就是说，用户先整理元数据，再享受后续搜索收益。',
      '当资料量增长后，标签、关键词、描述和合集的命名一致性，会比单次导入速度更重要。'
    ],
    bullets: [
      '首次使用时优先把存储目录配置稳定，避免后续频繁迁移。',
      '标签、关键词和描述尽量统一命名规则，这会直接影响搜索效果。',
      'AI 建议先用少量文件试跑，确认生成风格符合预期后再扩大使用范围。',
      '如果资料较多，建议把合集作为轻量分组，用来区分项目、课程、专题或来源。'
    ]
  },
  {
    id: 'roadmap',
    index: '04',
    title: '后续计划',
    summary: '下一阶段会继续围绕导入效率、整理效率和去重可信度三个方向推进。',
    paragraphs: [
      '基于 2026-07-19 的产品审查文档，LocalSpace 的下一步不建议把功能做散，而是优先把资料库闭环做完整。',
      '短期重点会放在批量导入、目录扫描导入、重复文件中心和整理工作台，让用户从“查找资料”逐步过渡到“经营资料库”。',
      '中期则会扩展高级搜索、资料合集 / 收藏夹、导入规则模板、元数据质量面板，以及更可控的 AI 生成体验。'
    ],
    bullets: [
      '批量导入与导入队列：支持多文件选择、公共字段套用、进度展示、失败重试与跳过重复项。',
      '目录扫描 / 监听导入：先做手动扫描，后做自动发现新增文件。',
      '重复文件中心：把已有重复检测能力前置，并提供集中处理入口。',
      '整理工作台：补足批量改标签、待整理视图、最近导入与最近修改等能力。',
      '高级搜索与规则模板：增强筛选、保存视图和自动化导入能力。',
      'AI 质量增强：提供简单理由、低置信度提示，以及仅补标签或描述等更细粒度操作。'
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

  const containerTop = container.getBoundingClientRect().top + 140
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
  height: 100vh;
  min-height: 0;
  background:
    radial-gradient(circle at top left, rgba(33, 150, 243, 0.08), transparent 22%),
    var(--app-bg-color, var(--bg-color));
}

.content {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

.page-intro {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  width: min(1220px, calc(100% - 32px));
  margin: 18px auto 0;
  padding: 16px 18px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 18px;
  background-color: color-mix(in srgb, var(--surface-color) 90%, transparent);
}

.page-intro-copy {
  min-width: 0;
}

.doc-back-link {
  display: inline-flex;
  align-items: center;
  margin-bottom: 10px;
  padding: 6px 10px;
  border-radius: 999px;
  background-color: color-mix(in srgb, var(--bg-color) 84%, var(--surface-color));
  border: 1px solid rgba(148, 163, 184, 0.16);
  color: var(--text-color);
  font-size: 12px;
  font-weight: 600;
}

.eyebrow {
  display: inline-block;
  width: fit-content;
  margin: 0 0 6px;
  padding: 4px 8px;
  border-radius: 999px;
  background-color: rgba(33, 150, 243, 0.08);
  color: var(--primary-color);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.page-intro h1 {
  margin: 0;
  font-size: 24px;
  line-height: 1.2;
  color: var(--text-color);
}

.subtitle {
  margin: 8px 0 0;
  font-size: 13px;
  line-height: 1.7;
  color: var(--text-color);
  opacity: 0.74;
}

.intro-meta {
  margin: 0;
  max-width: 280px;
  text-align: right;
  font-size: 12px;
  line-height: 1.7;
  color: var(--text-color);
  opacity: 0.66;
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
  margin: 0 0 4px;
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
  gap: 18px;
}

.hero-shell,
.doc-section {
  scroll-margin-top: 18px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 22px;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--surface-color) 94%, transparent), var(--surface-color));
  box-shadow: 0 14px 36px rgba(15, 23, 42, 0.08);
}

.hero-shell {
  display: grid;
  grid-template-columns: minmax(0, 1.15fr) minmax(280px, 0.85fr);
  gap: 18px;
  padding: 20px;
  overflow: hidden;
  background:
    radial-gradient(circle at top right, rgba(33, 150, 243, 0.12), transparent 24%),
    linear-gradient(180deg, color-mix(in srgb, var(--surface-color) 96%, transparent), var(--surface-color));
}

.hero-copy {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.hero-shell h2 {
  margin: 0;
  font-size: 28px;
  line-height: 1.35;
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

.hero-highlights {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.highlight-card {
  display: grid;
  gap: 8px;
  padding: 16px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 16px;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--bg-color) 78%, var(--surface-color)), transparent);
}

.highlight-label,
.section-kicker,
.shot-meta,
.hero-note-label {
  display: inline-flex;
  width: fit-content;
  align-items: center;
  border-radius: 999px;
  padding: 4px 9px;
  background-color: rgba(33, 150, 243, 0.1);
  color: var(--primary-color);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.highlight-card strong {
  font-size: 14px;
  line-height: 1.7;
  color: var(--text-color);
}

.hero-aside {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.hero-note {
  display: grid;
  gap: 10px;
  padding: 16px;
  border-radius: 18px;
  background:
    linear-gradient(135deg, rgba(33, 150, 243, 0.12), rgba(15, 23, 42, 0.03));
  border: 1px solid rgba(33, 150, 243, 0.12);
}

.hero-note p {
  margin: 0;
  font-size: 13px;
  line-height: 1.8;
  color: var(--text-color);
  opacity: 0.8;
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
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 16px;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--bg-color) 86%, var(--surface-color)), transparent);
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

.section-surface {
  padding: 20px;
}

.section-heading {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr);
  gap: 12px;
  margin-bottom: 16px;
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

.section-heading-copy {
  display: grid;
  gap: 7px;
}

.section-heading h2 {
  margin: 0;
  font-size: 21px;
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
  gap: 12px;
}

.bullet-grid,
.steps-grid {
  display: grid;
  gap: 12px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.bullet-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.bullet-card,
.step-card {
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 16px;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--bg-color) 88%, var(--surface-color)), transparent);
  color: var(--text-color);
}

.bullet-card {
  display: grid;
  grid-template-columns: 14px minmax(0, 1fr);
  gap: 10px;
  align-items: start;
  padding: 15px 16px;
  font-size: 13px;
  line-height: 1.8;
}

.bullet-dot {
  width: 9px;
  height: 9px;
  margin-top: 7px;
  border-radius: 999px;
  background: linear-gradient(135deg, var(--primary-color), color-mix(in srgb, var(--primary-color) 68%, #0f172a));
  box-shadow: 0 0 0 5px rgba(33, 150, 243, 0.12);
}

.steps-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.step-card {
  display: grid;
  gap: 10px;
  padding: 16px;
}

.step-badge {
  display: inline-flex;
  width: fit-content;
  padding: 6px 10px;
  border-radius: 999px;
  background-color: color-mix(in srgb, var(--primary-color) 14%, var(--surface-color));
  color: var(--primary-color);
  font-size: 12px;
  font-weight: 700;
}

.step-card p {
  margin: 0;
  font-size: 13px;
  line-height: 1.8;
  color: var(--text-color);
  opacity: 0.8;
}

.screenshot-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
  margin-top: 16px;
}

.shot-card {
  margin: 0;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 18px;
  overflow: hidden;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--bg-color) 88%, var(--surface-color)), transparent);
}

.shot-card:first-child {
  grid-column: 1 / -1;
}

.shot-frame {
  padding: 12px 12px 0;
}

.shot-card img {
  display: block;
  width: 100%;
  height: auto;
  object-fit: contain;
  background-color: #fff;
  border-radius: 14px;
  border: 1px solid rgba(148, 163, 184, 0.14);
}

.shot-card figcaption {
  display: grid;
  gap: 7px;
  padding: 14px 14px 16px;
}

.shot-card strong {
  font-size: 15px;
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
  .page-intro {
    flex-direction: column;
    align-items: flex-start;
  }

  .intro-meta {
    max-width: none;
    text-align: left;
  }

  .layout {
    grid-template-columns: 1fr;
  }

  .toc-card {
    position: static;
  }

  .hero-shell {
    grid-template-columns: 1fr;
  }

  .bullet-grid,
  .steps-grid,
  .screenshot-grid {
    grid-template-columns: 1fr;
  }

  .shot-card:first-child {
    grid-column: auto;
  }
}

@media (max-width: 768px) {
  .page-intro,
  .layout {
    width: calc(100% - 24px);
  }

  .page-intro {
    margin-top: 12px;
    padding: 14px;
  }

  .page-intro h1 {
    font-size: 22px;
  }

  .layout {
    padding: 12px 0 20px;
    gap: 14px;
  }

  .hero-shell,
  .section-surface {
    padding: 16px;
  }

  .hero-shell h2 {
    font-size: 20px;
  }

  .hero-highlights,
  .hero-stats {
    grid-template-columns: 1fr;
  }

  .section-heading {
    grid-template-columns: 1fr;
  }
}
</style>

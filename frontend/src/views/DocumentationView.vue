<template>
  <div class="page-shell documentation-view">
    <AppHeader />

    <div ref="contentRef" class="page-content">
      <div class="page-stack doc-stack">
        <section class="page-topbar documentation-topbar">
          <div class="page-topbar-copy">
            <router-link to="/settings" class="doc-back-link">返回设置</router-link>
            <h1 class="page-topbar-title">LocalSpace 使用说明</h1>
            <p class="page-topbar-note">
              离线查看核心功能说明、推荐使用流程和后续规划。
            </p>
          </div>

          <div class="page-topbar-meta">
            <div class="metric-badge">
              <strong>4</strong>
              <span>核心章节</span>
            </div>
            <div class="metric-badge">
              <strong>离线</strong>
              <span>可直接查看</span>
            </div>
          </div>
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
            <section class="section-panel intro-panel">
              <div class="intro-grid">
                <div class="intro-card">
                  <span class="eyebrow">定位</span>
                  <h2>把本地资料整理、检索、归档和复用串成同一条工作流。</h2>
                  <p>
                    LocalSpace 更像一个个人资料工作台，而不是简单的文件搬运工具。它适合课程资料、项目素材、截图归档和日常文档整理。
                  </p>
                </div>
                <div class="intro-card">
                  <span class="eyebrow">适用场景</span>
                  <p>个人知识库、课程材料、项目素材库、图片归档、视频剪辑素材、文档集中管理。</p>
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
              <div class="section-panel doc-panel">
                <div class="section-heading">
                  <span class="section-index">{{ section.index }}</span>
                  <div class="section-heading-copy">
                    <span class="eyebrow">{{ section.kicker }}</span>
                    <h2 class="section-title">{{ section.title }}</h2>
                    <p class="section-copy">{{ section.summary }}</p>
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
                      <span class="shot-meta">页面截图</span>
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
  kicker: string
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
    kicker: 'Overview',
    title: '软件基本介绍',
    summary: 'LocalSpace 的重点不在于“存文件”，而在于把本地资料整理成可长期维护、可回看、可检索的资料库。',
    paragraphs: [
      'LocalSpace 是一个“本地文件管理 + AI 辅助整理”的桌面应用，适合个人知识库、课程资料库和项目素材库。',
      '它把导入、归档、检索、打开和持续整理串成一条完整路径，让图片、文档、音视频、压缩包、安装包等内容都能进入同一个入口。',
      '当你接入 AI 配置后，还可以在导入时获得标签与描述建议，降低整理成本，提升后续搜索命中率。',
    ],
    bullets: [
      '本地优先：资料不依赖云端服务，适合长期沉淀。',
      '统一入口：不同类型文件在同一套界面里管理。',
      'AI 辅助：在导入阶段补充标签与描述，帮助后续查找。',
      '持续整理：支持存储目录、打开方式、主题和文档配置。',
    ],
  },
  {
    id: 'features',
    index: '02',
    kicker: 'Features',
    title: '功能说明',
    summary: '当前版本的核心能力集中在文件浏览、导入整理、任务管理和设置配置四个部分。',
    paragraphs: [
      '文件页是最常用的入口，支持按类型筛选、按合集浏览、按关键词搜索，并结合缩略图与元信息快速回看内容。',
      '导入页将单文件导入与批量导入任务统一到一个界面中，既适合精细整理，也适合长时间后台处理。',
      '任务中心用于查看批量导入的运行状态、恢复中断任务、回看历史记录，并支持分页浏览。',
    ],
    bullets: [
      '文件浏览：支持类型筛选、搜索、缩略图与详情预览。',
      '导入整理：支持单文件导入与批量后台导入。',
      '任务中心：支持历史记录、分页查询、异常任务恢复。',
      '设置中心：统一管理目录、打开方式、AI、主题和存储规则。',
    ],
    screenshots: [
      {
        src: '/docs/home-overview.png',
        alt: '文件页截图',
        title: '文件页',
        caption: '用于日常搜索、筛选、查看和打开资料，是最常驻的工作页面。',
      },
      {
        src: '/docs/import-workflow.png',
        alt: '导入页截图',
        title: '导入页',
        caption: '在单文件导入与批量导入之间切换，并在导入前补全关键信息。',
      },
      {
        src: '/docs/settings-documentation-entry.png',
        alt: '设置页截图',
        title: '设置页',
        caption: '集中管理目录、主题、AI 与内置文档入口。',
      },
    ],
  },
  {
    id: 'user-guide',
    index: '03',
    kicker: 'Guide',
    title: '推荐使用流程',
    summary: '首次使用时，建议先完成目录与打开方式的基础配置，再开始导入文件。',
    steps: [
      {
        title: '先配置主存储目录',
        description: '在设置页中新增主目录，并根据容量和使用习惯确定默认目录。',
      },
      {
        title: '补充打开方式和存储规则',
        description: '如果你希望文件按更清晰的结构归档，或使用固定应用打开文件，可以继续完善这两项配置。',
      },
      {
        title: '按需开启 AI 配置',
        description: '当你希望导入时自动补全标签与描述，可以填写模型、接口地址与 API Key。',
      },
      {
        title: '导入第一批文件',
        description: '进入导入页，选择单文件或批量模式，补充名称、标签、描述和合集等信息。',
      },
      {
        title: '检查结果并持续整理',
        description: '回到文件页查看导入结果，确认标签和合集命名是否一致，后续搜索会更稳定。',
      },
    ],
    bullets: [
      '先把目录规划稳定，再大批量导入，能减少后续迁移成本。',
      '标签、关键词和合集命名越统一，搜索体验越好。',
      'AI 建议适合先小范围试用，确认风格后再扩展使用。',
      '资料量增大后，合集会比单次导入速度更影响长期管理效率。',
    ],
  },
  {
    id: 'roadmap',
    index: '04',
    kicker: 'Roadmap',
    title: '后续计划',
    summary: '下一阶段将继续围绕导入效率、资料整理效率和重复内容治理来演进。',
    paragraphs: [
      '基于当前产品目标，下一步更适合做深而不是做散，优先把资料库闭环打磨完整。',
      '短期重点会放在批量导入、目录扫描导入、重复文件中心和整理工作台，让用户从“收纳资料”逐步过渡到“经营资料库”。',
      '中期可以扩展高级搜索、导入模板、资料质量面板以及更细粒度的 AI 辅助能力。',
    ],
    bullets: [
      '批量导入增强：失败重试、跳过重复、任务队列优化。',
      '目录扫描导入：先手动扫描，再扩展到自动监听。',
      '重复文件中心：把重复检测前置，并提供集中处理入口。',
      '整理工作台：补充待整理、最近导入、最近修改等视图。',
      '高级搜索与模板：增强筛选能力和可复用的导入规则。',
    ],
  },
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
    if (rect.top <= containerTop) currentSectionId = section.id
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
.doc-stack {
  gap: 18px;
}

.documentation-topbar {
  align-items: center;
}

.doc-back-link {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  min-height: 36px;
  padding: 8px 12px;
  border-radius: 999px;
  border: 1px solid rgba(146, 165, 192, 0.18);
  background: rgba(255, 255, 255, 0.58);
  color: var(--text-soft);
  font-size: 13px;
  font-weight: 600;
}

.layout {
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr);
  gap: 18px;
}

.toc {
  position: relative;
}

.toc-card {
  position: sticky;
  top: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 16px;
  border: 1px solid rgba(146, 165, 192, 0.18);
  border-radius: 26px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.88), rgba(244, 247, 252, 0.84));
}

.toc-card h2 {
  margin-bottom: 4px;
  font-size: 15px;
  color: var(--text-color);
}

.toc-link {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 10px 12px;
  border-radius: 16px;
  text-align: left;
  color: var(--text-soft);
  font-size: 13px;
}

.toc-link.active {
  background: rgba(255, 255, 255, 0.92);
  color: var(--text-color);
  box-shadow:
    0 12px 26px rgba(34, 52, 84, 0.08),
    inset 0 0 0 1px rgba(111, 143, 216, 0.16);
}

.toc-index {
  width: 28px;
  color: var(--text-faint);
  font-size: 11px;
  font-family: var(--font-mono);
}

.document {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.intro-panel,
.doc-panel {
  padding: 22px;
}

.intro-grid {
  display: grid;
  grid-template-columns: 1.2fr 0.8fr;
  gap: 14px;
}

.intro-card {
  display: grid;
  gap: 12px;
  padding: 18px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.56);
}

.intro-card h2 {
  font-size: 28px;
  line-height: 1.2;
  color: var(--text-color);
}

.intro-card p {
  color: var(--text-soft);
  line-height: 1.8;
}

.section-heading {
  display: grid;
  grid-template-columns: 52px minmax(0, 1fr);
  gap: 14px;
  margin-bottom: 18px;
}

.section-index {
  width: 52px;
  height: 52px;
  border-radius: 18px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(235, 241, 250, 0.92);
  color: var(--text-color);
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 700;
}

.section-heading-copy {
  display: grid;
  gap: 10px;
}

.paragraphs {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.paragraphs p {
  color: var(--text-soft);
  line-height: 1.8;
}

.bullet-grid,
.steps-grid {
  display: grid;
  gap: 12px;
  margin-top: 16px;
  list-style: none;
}

.bullet-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.bullet-card,
.step-card {
  border-radius: 22px;
  border: 1px solid rgba(146, 165, 192, 0.16);
  background: rgba(255, 255, 255, 0.56);
}

.bullet-card {
  display: grid;
  grid-template-columns: 14px minmax(0, 1fr);
  gap: 10px;
  padding: 16px;
  color: var(--text-soft);
  line-height: 1.8;
}

.bullet-dot {
  width: 8px;
  height: 8px;
  margin-top: 8px;
  border-radius: 999px;
  background: var(--primary-color);
  box-shadow: 0 0 0 5px rgba(111, 143, 216, 0.12);
}

.steps-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
  padding: 0;
}

.step-card {
  display: grid;
  gap: 12px;
  padding: 16px;
}

.step-badge {
  display: inline-flex;
  width: fit-content;
  padding: 6px 10px;
  border-radius: 999px;
  background: rgba(235, 241, 250, 0.92);
  color: var(--text-color);
  font-size: 12px;
  font-weight: 700;
}

.step-card p {
  color: var(--text-soft);
  line-height: 1.8;
}

.screenshot-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
  margin-top: 18px;
}

.shot-card {
  margin: 0;
  overflow: hidden;
  border-radius: 24px;
  border: 1px solid rgba(146, 165, 192, 0.16);
  background: rgba(255, 255, 255, 0.56);
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
  border-radius: 18px;
  border: 1px solid rgba(146, 165, 192, 0.14);
  background: #fff;
}

.shot-card figcaption {
  display: grid;
  gap: 8px;
  padding: 14px 14px 16px;
}

.shot-meta {
  color: var(--text-faint);
  font-size: 11px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.shot-card strong {
  font-size: 16px;
  color: var(--text-color);
}

.shot-card span:last-child {
  color: var(--text-soft);
  line-height: 1.7;
}

@media (max-width: 1100px) {
  .layout,
  .intro-grid,
  .bullet-grid,
  .steps-grid,
  .screenshot-grid {
    grid-template-columns: 1fr;
  }

  .toc-card {
    position: static;
  }

  .shot-card:first-child {
    grid-column: auto;
  }
}

@media (max-width: 640px) {
  .intro-panel,
  .doc-panel {
    padding: 16px;
  }

  .section-heading {
    grid-template-columns: 1fr;
  }

  .section-index {
    width: 44px;
    height: 44px;
    border-radius: 14px;
  }
}
</style>

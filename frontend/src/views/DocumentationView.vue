<template>
  <div class="page-shell documentation-view">
    <AppHeader />

    <div ref="contentRef" class="page-content">
      <div class="guide-shell">
        <header id="overview" :ref="(el) => setSectionRef('overview', el)" class="guide-hero">
          <div class="hero-copy">
            <span class="guide-kicker">产品概览</span>
            <h1>LocalSpace 使用说明</h1>
            <p>
              LocalSpace 是一个面向个人资料整理与回看的本地桌面工具。它把散落的图片、文档、
              音视频、压缩包和项目素材收进同一套资料库，再用合集、标签、描述和缩略图保留上下文。
            </p>
          </div>

          <div class="hero-notes" aria-label="产品特点">
            <div v-for="note in productNotes" :key="note.title" class="hero-note">
              <strong>{{ note.title }}</strong>
              <span>{{ note.description }}</span>
            </div>
          </div>

          <div class="workflow-map" aria-label="LocalSpace 推荐使用路径">
            <div class="workflow-heading">
              <span>一条完整路径</span>
              <strong>资料不是被“搬进去”，而是从此有了可回看的上下文。</strong>
            </div>
            <ol class="workflow-steps">
              <li v-for="(step, index) in workflowSteps" :key="step.title">
                <span class="workflow-index">{{ String(index + 1).padStart(2, '0') }}</span>
                <div><strong>{{ step.title }}</strong><p>{{ step.note }}</p></div>
                <svg v-if="index < workflowSteps.length - 1" viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M5 12h13M14 8l4 4-4 4" />
                </svg>
              </li>
            </ol>
          </div>
        </header>

        <div class="guide-layout">
          <aside class="guide-toc" aria-label="使用说明目录">
            <div class="toc-inner">
              <span class="toc-label">本页目录</span>
              <button
                v-for="item in navigationItems"
                :key="item.id"
                type="button"
                class="toc-link"
                :class="{ active: activeSectionId === item.id }"
                :aria-current="activeSectionId === item.id ? 'location' : undefined"
                @click="scrollToSection(item.id)"
              >
                <span class="toc-marker"></span><span>{{ item.label }}</span>
              </button>
              <div class="toc-tip"><span>快速入口</span><router-link to="/import">去导入资料</router-link></div>
            </div>
          </aside>

          <main class="guide-content">
            <section id="quick-start" :ref="(el) => setSectionRef('quick-start', el)" class="guide-section">
              <div class="section-heading">
                <span class="section-kicker">第一次使用</span>
                <h2>先把资料放在哪里想清楚。</h2>
                <p>完成下面四件事就可以开始使用。AI、主题和自定义打开方式都可以稍后再配置。</p>
              </div>

              <ol class="start-list">
                <li v-for="(step, index) in quickStartSteps" :key="step.title">
                  <span class="start-number">{{ index + 1 }}</span>
                  <div><strong>{{ step.title }}</strong><p>{{ step.description }}</p></div>
                  <router-link v-if="step.route" :to="step.route">{{ step.action }}</router-link>
                </li>
              </ol>

              <div class="quiet-note">
                <svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="9" /><path d="M12 10.5v6M12 7.4h.01" /></svg>
                <p><strong>建议先小批量试用。</strong>确认存储层级、合集命名和 AI 输出符合习惯后，再导入大量资料。</p>
              </div>
            </section>

            <section
              v-for="module in modules"
              :id="module.id"
              :key="module.id"
              :ref="(el) => setSectionRef(module.id, el)"
              class="guide-section module-section"
            >
              <div class="module-heading">
                <div class="module-title-row">
                  <span class="module-icon" aria-hidden="true">
                    <svg viewBox="0 0 24 24"><path v-for="path in module.iconPaths" :key="path" :d="path" /></svg>
                  </span>
                  <div><span class="section-kicker">{{ module.kicker }}</span><h2>{{ module.title }}</h2></div>
                </div>
                <router-link :to="module.route" class="module-link">
                  {{ module.action }}
                  <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M5 12h13M14 8l4 4-4 4" /></svg>
                </router-link>
              </div>

              <p class="module-lede">{{ module.summary }}</p>

              <div class="capability-list">
                <article v-for="feature in module.features" :key="feature.title" class="capability-item">
                  <strong>{{ feature.title }}</strong><p>{{ feature.description }}</p>
                </article>
              </div>

              <div v-if="module.tips.length" class="tip-strip">
                <span>使用要点</span>
                <ul><li v-for="tip in module.tips" :key="tip">{{ tip }}</li></ul>
              </div>

              <figure v-if="module.screenshot" class="screen-figure">
                <div class="screen-frame"><img :src="module.screenshot.src" :alt="module.screenshot.alt" loading="lazy" /></div>
                <figcaption>
                  <span>界面截图</span><strong>{{ module.screenshot.title }}</strong><p>{{ module.screenshot.caption }}</p>
                </figcaption>
              </figure>

              <div v-if="module.id === 'tasks'" class="exit-guide">
                <div class="exit-copy">
                  <span class="section-kicker">退出前检查</span>
                  <h3>有任务运行时，关闭按钮会先打开退出确认。</h3>
                  <p>
                    弹窗会列出受影响任务、当前状态和是否可恢复。优先选择“返回应用”等待完成，
                    或选择“查看任务”处理；只有确认不再等待时才使用“仍然退出”。
                  </p>
                  <div class="exit-actions" aria-label="退出确认中的三个操作">
                    <div><strong>返回应用</strong><span>关闭弹窗，任务继续运行</span></div>
                    <div><strong>查看任务</strong><span>跳转到任务页确认进度</span></div>
                    <div class="danger"><strong>仍然退出</strong><span>中断当前处理后关闭应用</span></div>
                  </div>
                </div>
                <figure class="exit-screen">
                  <img src="/docs/exit-guard.png" alt="运行任务时关闭 LocalSpace 出现的退出确认框" loading="lazy" />
                  <figcaption>退出前先看清“可恢复 / 不可恢复”标记。</figcaption>
                </figure>
              </div>
            </section>

            <section id="faq" :ref="(el) => setSectionRef('faq', el)" class="guide-section faq-section">
              <div class="section-heading compact"><span class="section-kicker">遇到问题时</span><h2>先检查这几处。</h2></div>
              <div class="faq-list">
                <details v-for="(item, index) in faqs" :key="item.question" :open="index === 0">
                  <summary>{{ item.question }}</summary><p>{{ item.answer }}</p>
                </details>
              </div>
              <footer class="guide-footer">
                <span class="footer-mark">LS</span>
                <div><strong>LocalSpace 使用说明</strong><p>适用于当前 0.3.0 版本。功能入口以应用内实际界面为准。</p></div>
                <button type="button" @click="scrollToSection('overview')">回到顶部 ↑</button>
              </footer>
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

interface Feature {
  title: string
  description: string
}

interface Screenshot {
  src: string
  alt: string
  title: string
  caption: string
}

interface GuideModule {
  id: string
  kicker: string
  title: string
  summary: string
  route: string
  action: string
  iconPaths: string[]
  features: Feature[]
  tips: string[]
  screenshot?: Screenshot
}

const navigationItems = [
  { id: 'overview', label: '认识 LocalSpace' },
  { id: 'quick-start', label: '第一次使用' },
  { id: 'library', label: '资料库' },
  { id: 'collections', label: '合集' },
  { id: 'import', label: '导入' },
  { id: 'tasks', label: '任务管理' },
  { id: 'settings', label: '设置' },
  { id: 'faq', label: '常见问题' },
]

const productNotes = [
  { title: '本地优先', description: '文件与资料库由你自己的存储目录承载。' },
  { title: '可检索', description: '文件名、标签和描述会共同帮助你重新找到资料。' },
  { title: '持续整理', description: '先导入，之后仍可调整元信息、合集与打开方式。' },
]

const workflowSteps = [
  { title: '设置', note: '确定存储目录与规则' },
  { title: '导入', note: '选择文件并补充信息' },
  { title: '任务', note: '观察后台进度与异常' },
  { title: '资料库', note: '搜索、查看与继续整理' },
  { title: '合集', note: '按上下文重新组织资料' },
]

const quickStartSteps = [
  {
    title: '添加主存储目录',
    description: '进入设置 → 存储目录，添加至少一个主目录，并确认默认目录。没有主目录时，导入的资料没有落脚位置。',
    route: '/settings',
    action: '打开设置',
  },
  {
    title: '确认存储规则',
    description: '决定文件按类型、合集等层级如何落盘。已有资料较多时，建议先用少量文件验证目录结构。',
    route: '/settings',
    action: '查看规则',
  },
  {
    title: '导入一份测试资料',
    description: '选择单文件导入，填写文件名、标签、描述和合集，检查资料在资料库中的呈现是否符合预期。',
    route: '/import',
    action: '开始导入',
  },
  {
    title: '形成自己的命名习惯',
    description: '合集用自然语言命名，标签保持简短一致；以后搜索时，使用你当初会自然想到的词。',
    route: '',
    action: '',
  },
]

const modules: GuideModule[] = [
  {
    id: 'library',
    kicker: '日常入口',
    title: '资料库',
    summary: '资料库是最常驻的工作页面，用来重新找到资料、查看上下文，并继续完成打开、编辑、移动或删除等操作。',
    route: '/files',
    action: '打开资料库',
    iconPaths: ['M3 7.5A2.5 2.5 0 0 1 5.5 5h4l2 2h7A2.5 2.5 0 0 1 21 9.5v7a2.5 2.5 0 0 1-2.5 2.5h-13A2.5 2.5 0 0 1 3 16.5z'],
    features: [
      { title: '搜索资料', description: '输入文件名、标签或描述中的词，结果会随搜索条件更新。' },
      { title: '缩小范围', description: '通过文件类型与合集筛选收窄结果；当前条件会以轻量标签保留。' },
      { title: '查看与打开', description: '通过缩略图和详情确认内容，可按偏好应用、系统默认方式打开，或定位到文件位置。' },
      { title: '继续整理', description: '编辑名称、标签、描述和合集；勾选多份资料后还可批量移动或删除。' },
    ],
    tips: ['搜索词尽量使用当时写入的标签或描述语言。', '删除资料前确认目标；删除动作会影响资料库记录与本地内容。'],
    screenshot: {
      src: '/docs/library-overview.png',
      alt: 'LocalSpace 资料库页面，包含搜索、筛选和文件卡片',
      title: '资料库的主要阅读顺序',
      caption: '先用顶部搜索与筛选缩小范围，再从文件卡片读取类型、名称、描述、标签和大小。',
    },
  },
  {
    id: 'collections',
    kicker: '按上下文浏览',
    title: '合集',
    summary: '合集不是物理文件夹的替代品，而是一种面向“事情与用途”的分组。项目资料、课程材料或灵感参考都适合成为合集。',
    route: '/collections',
    action: '浏览合集',
    iconPaths: ['M4 6h6l2 2h8v10H4z', 'M4 10h16'],
    features: [
      { title: '按名称查找', description: '在合集页输入名称，快速定位已有分组。' },
      { title: '按文件类型筛选', description: '只查看包含指定类型资料的合集，例如图片、文档或视频。' },
      { title: '进入合集', description: '点击合集后会回到资料库，并自动带上对应的合集筛选条件。' },
      { title: '维护合集', description: '在设置 → 合集中新增或删除合集；被资料引用的合集需要先解除引用。' },
    ],
    tips: ['用“设计参考”“课程资料”这类自然名称，比临时项目编号更容易在未来理解。', '还没想好归属时可以先留在未分类，之后再统一整理。'],
  },
  {
    id: 'import',
    kicker: '资料进入工作台的入口',
    title: '导入',
    summary: '导入页分为单文件与批量任务两种模式。两者都通过系统文件窗口选择资料，不支持把文件拖进页面。',
    route: '/import',
    action: '前往导入',
    iconPaths: ['M12 3v12', 'm7 10 5 5 5-5', 'M4 19h16'],
    features: [
      { title: '单文件导入', description: '适合精细整理一份资料。可修改文件名、关键词、标签、描述和合集，再立即导入。' },
      { title: 'AI 辅助', description: '启用 AI 后可生成标签与描述建议；先检查内容，再选择应用描述或应用全部。' },
      { title: '批量导入任务', description: '按住 Ctrl / Shift 一次选择多份文件，设置公共字段后创建后台任务。' },
      { title: '重复与反馈', description: '批量选择可分多次追加，重复文件会跳过；提交后的进度与结果到任务页查看。' },
    ],
    tips: ['单文件适合“现在就整理好”，批量模式适合“先把同一批资料收进来”。', '批量公共字段会写入每个文件，提交前确认它们适用于整批资料。'],
    screenshot: {
      src: '/docs/import-workflow.png',
      alt: 'LocalSpace 单文件导入页面，包含文件来源、元信息与 AI 分析区域',
      title: '单文件导入',
      caption: '从上到下确认文件来源与元信息，AI 建议位于右侧；最后在页面底部提交导入。',
    },
  },
  {
    id: 'tasks',
    kicker: '后台进度与异常处理',
    title: '任务管理',
    summary: '批量导入、单文件导入和清理操作都会留下任务记录。这里首先回答“有没有仍在处理或需要我介入的任务”。',
    route: '/tasks',
    action: '查看任务',
    iconPaths: ['M5 5h14v14H5z', 'm8 12 2.5 2.5L16 9'],
    features: [
      { title: '识别状态', description: '进行中、排队、等待继续与失败任务会使用状态色提示；已完成记录保持安静。' },
      { title: '观察进度', description: '运行中的任务显示已处理数量与进度，结束后改为展示成功、失败等结果。' },
      { title: '处理异常', description: '可取消活动任务、继续可恢复任务，并展开详情查看错误与单文件结果。' },
      { title: '整理历史', description: '终态任务可以删除；设置 → 任务中还可以配置保留数量、天数并手动清理。' },
    ],
    tips: ['右上角的任务状态入口可快速查看后台是否仍在工作。', '任务失败时先展开“详情”，错误原因通常比重新提交更有帮助。'],
    screenshot: {
      src: '/docs/task-center.png',
      alt: 'LocalSpace 任务页面，展示运行中、排队、等待继续、失败和已完成任务',
      title: '任务状态集中在同一条扫描轴上',
      caption: '先看行首状态色与状态标签，再看进度或结果；需要介入的动作始终位于右侧。',
    },
  },
  {
    id: 'settings',
    kicker: '决定资料如何落盘与呈现',
    title: '设置',
    summary: '设置页按配置对象分区。第一次使用先处理存储目录与存储规则，其余选项可以随使用习惯逐步完善。',
    route: '/settings',
    action: '打开设置',
    iconPaths: ['M12 3.6a8.4 8.4 0 1 1 0 16.8 8.4 8.4 0 0 1 0-16.8z', 'M12 7.2V12l3.2 1.9'],
    features: [
      { title: '存储目录与规则', description: '添加多个主目录、指定默认目录，并控制导入后的目录层级。' },
      { title: '打开方式', description: '为常见文件类型或特定扩展名指定优先使用的本地应用。' },
      { title: '主题与 AI', description: '调整明暗、强调色与背景；配置 AI 模型、接口、密钥和工具能力。' },
      { title: '合集、任务与关于', description: '维护合集、设置任务保留策略，并查看版本、许可和内置说明入口。' },
    ],
    tips: ['标记“立即生效”的设置无需再寻找保存按钮。', '移除主存储目录是高风险操作，请先阅读确认框中的路径与后果。'],
    screenshot: {
      src: '/docs/settings-overview.png',
      alt: 'LocalSpace 设置页面，左侧为设置分区，右侧为存储目录配置',
      title: '设置按对象分区',
      caption: '左侧切换存储、主题、AI、合集和任务等分区，右侧只呈现当前配置。',
    },
  },
]

const faqs = [
  { question: '为什么导入按钮不可用？', answer: '先确认已经选择文件，并填写必需的文件名；如果仍不可用，再到设置中检查是否已添加可用的主存储目录。' },
  { question: '为什么合集不能删除？', answer: '正在被资料引用的合集会受到保护。先在资料库中把这些资料移动到其他合集或未分类，再返回设置删除。' },
  { question: '任务失败后应该直接重新导入吗？', answer: '先在任务页展开详情，确认是存储空间、文件损坏还是配置问题。修复原因后再继续或重新提交。' },
  { question: 'AI 是使用 LocalSpace 的必需项吗？', answer: '不是。AI 只用于辅助生成标签和描述；关闭 AI 后，导入、搜索、合集和任务管理仍可正常使用。' },
]

const contentRef = ref<HTMLElement | null>(null)
const activeSectionId = ref('overview')
const sectionRefs = new Map<string, HTMLElement>()

const setSectionRef = (id: string, element: Element | null) => {
  if (element instanceof HTMLElement) sectionRefs.set(id, element)
  else sectionRefs.delete(id)
}

const scrollToSection = (id: string) => {
  const target = sectionRefs.get(id)
  if (!target) return
  activeSectionId.value = id
  target.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

const updateActiveSection = () => {
  const container = contentRef.value
  if (!container) return

  // 页面滚动发生在内容画布内，以画布顶部下方的阅读线判断当前章节。
  const readingLine = container.getBoundingClientRect().top + 150
  let currentId = navigationItems[0].id

  for (const item of navigationItems) {
    const section = sectionRefs.get(item.id)
    if (!section) continue
    if (section.getBoundingClientRect().top <= readingLine) currentId = item.id
  }
  activeSectionId.value = currentId
}

onMounted(() => {
  contentRef.value?.addEventListener('scroll', updateActiveSection, { passive: true })
  updateActiveSection()
})

onBeforeUnmount(() => contentRef.value?.removeEventListener('scroll', updateActiveSection))
</script>

<style scoped>
.documentation-view {
  --g-ink: var(--text-color);
  --g-soft: var(--text-soft);
  --g-faint: var(--text-faint);
  --g-line: color-mix(in srgb, var(--border-color) 82%, transparent);
  --g-paper: color-mix(in srgb, var(--surface-color) 88%, transparent);
  --g-blue-soft: color-mix(in srgb, var(--primary-color) 10%, var(--surface-color));
  --g-warn: #ad7216;
  --g-warn-soft: color-mix(in srgb, #efbd5c 14%, var(--surface-color));
}
.documentation-view .page-content { padding-top: 26px; scroll-behavior: smooth; }
.guide-shell { width: min(1220px, 100%); margin: 0 auto; }
.guide-hero {
  display: grid; grid-template-columns: minmax(0, 1.25fr) minmax(300px, .75fr);
  gap: 26px 48px; padding: 24px 0 32px; scroll-margin-top: 24px;
}
.guide-kicker, .section-kicker, .toc-label {
  display: block; color: var(--g-faint); font-size: 10px; font-weight: 800;
  letter-spacing: .16em; text-transform: uppercase;
}
.hero-copy { max-width: 760px; }
.hero-copy h1 {
  margin-top: 9px; color: var(--g-ink); font-family: Avenir, "Segoe UI Variable Display", "Microsoft YaHei UI", sans-serif;
  font-size: 30px; font-weight: 720; letter-spacing: -.035em; line-height: 1.25;
}
.hero-copy p { max-width: 720px; margin-top: 14px; color: var(--g-soft); font-size: 13px; line-height: 1.85; }
.hero-notes { align-self: end; border-top: 1px solid var(--g-line); }
.hero-note { display: grid; grid-template-columns: 72px 1fr; gap: 14px; padding: 15px 0; border-bottom: 1px solid var(--g-line); }
.hero-note strong { color: var(--g-ink); font-size: 12px; }
.hero-note span { color: var(--g-faint); font-size: 11px; line-height: 1.65; }
.workflow-map { grid-column: 1 / -1; overflow: hidden; border: 1px solid var(--g-line); border-radius: 18px; background: var(--g-paper); box-shadow: 0 18px 42px var(--shadow-color); }
.workflow-heading { display: flex; justify-content: space-between; gap: 24px; padding: 15px 18px; border-bottom: 1px solid var(--g-line); }
.workflow-heading span { color: var(--primary-color); font-size: 10px; font-weight: 800; letter-spacing: .12em; }
.workflow-heading strong { color: var(--g-soft); font-size: 11px; }
.workflow-steps { display: grid; grid-template-columns: repeat(5, 1fr); list-style: none; }
.workflow-steps li { position: relative; display: grid; grid-template-columns: 30px 1fr; gap: 10px; min-height: 96px; padding: 20px 26px 18px 18px; border-right: 1px solid var(--g-line); }
.workflow-steps li:last-child { border-right: 0; }
.workflow-index { display: grid; place-items: center; width: 30px; height: 30px; border-radius: 9px; background: var(--g-blue-soft); color: var(--primary-color); font: 800 10px Consolas, monospace; }
.workflow-steps strong { color: var(--g-ink); font-size: 13px; }
.workflow-steps p { margin-top: 5px; color: var(--g-faint); font-size: 10px; line-height: 1.55; }
.workflow-steps svg { position: absolute; z-index: 1; top: 27px; right: -10px; width: 19px; height: 19px; padding: 4px; border-radius: 50%; background: var(--surface-color); fill: none; stroke: var(--g-faint); stroke-width: 1.8; }
.guide-layout { display: grid; grid-template-columns: 184px minmax(0, 1fr); gap: 48px; padding-bottom: 72px; }
.toc-inner { position: sticky; top: 0; display: grid; padding-top: 22px; }
.toc-label { padding: 0 12px 10px; }
.toc-link { display: grid; grid-template-columns: 3px 1fr; align-items: center; gap: 11px; min-height: 38px; padding: 8px 11px 8px 0; border-radius: 0 9px 9px 0; color: var(--g-faint); text-align: left; font-size: 12px; font-weight: 620; }
.toc-link:hover, .toc-link.active { color: var(--g-ink); background: color-mix(in srgb, var(--surface-color) 68%, transparent); }
.toc-marker { align-self: stretch; border-radius: 0 3px 3px 0; }
.toc-link.active .toc-marker { background: var(--primary-color); }
.toc-tip { display: grid; gap: 7px; margin-top: 22px; padding: 14px 12px; border-top: 1px solid var(--g-line); }
.toc-tip span { color: var(--g-faint); font-size: 10px; }
.toc-tip a { color: var(--primary-color); font-size: 12px; font-weight: 700; }
.guide-content { min-width: 0; }
.guide-section { padding: 58px 0; border-top: 1px solid var(--g-line); scroll-margin-top: 18px; }
.section-heading { max-width: 760px; }
.section-heading h2, .module-heading h2 { margin-top: 10px; color: var(--g-ink); font-family: Avenir, "Segoe UI Variable Display", "Microsoft YaHei UI", sans-serif; font-size: clamp(30px, 4vw, 46px); font-weight: 730; letter-spacing: -.045em; line-height: 1.15; }
.section-heading > p { margin-top: 15px; color: var(--g-soft); font-size: 14px; line-height: 1.8; }
.start-list { margin-top: 32px; border-top: 1px solid var(--g-line); list-style: none; }
.start-list li { display: grid; grid-template-columns: 36px 1fr auto; align-items: center; gap: 16px; padding: 20px 4px; border-bottom: 1px solid var(--g-line); }
.start-number { display: grid; place-items: center; width: 30px; height: 30px; border: 1px solid var(--g-line); border-radius: 9px; color: var(--g-faint); font: 10px Consolas, monospace; }
.start-list strong, .capability-item strong { color: var(--g-ink); font-size: 13px; }
.start-list p, .capability-item p { margin-top: 6px; color: var(--g-soft); font-size: 12px; line-height: 1.75; }
.start-list a { color: var(--primary-color); font-size: 11px; font-weight: 700; }
.quiet-note { display: flex; gap: 12px; margin-top: 22px; padding: 15px 17px; border-radius: 12px; background: var(--g-blue-soft); color: var(--g-soft); }
.quiet-note svg { width: 18px; height: 18px; flex: none; fill: none; stroke: var(--primary-color); stroke-width: 1.8; }
.quiet-note p { font-size: 12px; line-height: 1.75; } .quiet-note strong { color: var(--g-ink); }
.module-heading { display: flex; align-items: flex-end; justify-content: space-between; gap: 24px; }
.module-title-row { display: flex; align-items: center; gap: 17px; }
.module-icon { display: grid; place-items: center; width: 48px; height: 48px; flex: none; border: 1px solid color-mix(in srgb, var(--primary-color) 20%, var(--g-line)); border-radius: 14px; background: var(--g-blue-soft); color: var(--primary-color); }
.module-icon svg { width: 21px; height: 21px; fill: none; stroke: currentColor; stroke-width: 1.7; stroke-linecap: round; stroke-linejoin: round; }
.module-heading h2 { margin-top: 5px; }
.module-link { display: inline-flex; align-items: center; gap: 7px; padding: 9px 12px; border: 1px solid var(--g-line); border-radius: 9px; background: var(--g-paper); color: var(--g-soft); font-size: 11px; font-weight: 700; }
.module-link:hover { color: var(--primary-color); border-color: color-mix(in srgb, var(--primary-color) 34%, var(--g-line)); }
.module-link svg { width: 15px; height: 15px; fill: none; stroke: currentColor; stroke-width: 1.8; }
.module-lede { max-width: 780px; margin-top: 22px; color: var(--g-soft); font-size: 14px; line-height: 1.9; }
.capability-list { display: grid; grid-template-columns: repeat(2, 1fr); margin-top: 30px; border-top: 1px solid var(--g-line); }
.capability-item { padding: 20px 22px 20px 0; border-bottom: 1px solid var(--g-line); }
.capability-item:nth-child(odd) { padding-right: 32px; border-right: 1px solid var(--g-line); }
.capability-item:nth-child(even) { padding-left: 32px; }
.tip-strip { display: grid; grid-template-columns: 86px 1fr; gap: 18px; margin-top: 22px; padding: 16px 18px; border-left: 3px solid var(--primary-color); background: color-mix(in srgb, var(--g-blue-soft) 62%, transparent); }
.tip-strip > span { color: var(--primary-color); font-size: 11px; font-weight: 800; }
.tip-strip ul { display: grid; gap: 7px; padding-left: 17px; color: var(--g-soft); font-size: 11px; line-height: 1.65; }
.screen-figure { display: grid; grid-template-columns: minmax(0, 1fr) 210px; margin: 30px 0 0; overflow: hidden; border: 1px solid var(--g-line); border-radius: 16px; background: var(--g-paper); box-shadow: 0 18px 40px var(--shadow-color); }
.screen-frame { min-width: 0; padding: 10px; border-right: 1px solid var(--g-line); }
.screen-frame img, .exit-screen img { display: block; width: 100%; height: auto; border-radius: 10px; background: #eef3f9; }
.screen-figure figcaption { display: flex; flex-direction: column; justify-content: flex-end; padding: 22px 18px; }
.screen-figure figcaption > span { color: var(--g-faint); font-size: 9px; font-weight: 800; letter-spacing: .14em; }
.screen-figure figcaption strong { margin-top: 10px; color: var(--g-ink); font-size: 14px; line-height: 1.45; }
.screen-figure figcaption p { margin-top: 8px; color: var(--g-soft); font-size: 11px; line-height: 1.7; }
.exit-guide { display: grid; grid-template-columns: minmax(0, .9fr) minmax(360px, 1.1fr); gap: 28px; margin-top: 30px; padding: 24px; border: 1px solid color-mix(in srgb, #d89b34 40%, var(--g-line)); border-radius: 16px; background: var(--g-warn-soft); }
.exit-copy h3 { margin-top: 10px; color: var(--g-ink); font-size: 22px; letter-spacing: -.025em; line-height: 1.35; }
.exit-copy > p { margin-top: 12px; color: var(--g-soft); font-size: 12px; line-height: 1.8; }
.exit-actions { display: grid; margin-top: 18px; border-top: 1px solid color-mix(in srgb, #d89b34 24%, var(--g-line)); }
.exit-actions div { display: grid; grid-template-columns: 76px 1fr; gap: 10px; padding: 10px 0; border-bottom: 1px solid color-mix(in srgb, #d89b34 24%, var(--g-line)); }
.exit-actions strong { color: var(--g-ink); font-size: 11px; } .exit-actions span { color: var(--g-faint); font-size: 10px; } .exit-actions .danger strong { color: #b95757; }
.exit-screen { align-self: center; margin: 0; padding: 8px; border: 1px solid color-mix(in srgb, #d89b34 24%, var(--g-line)); border-radius: 13px; background: color-mix(in srgb, var(--surface-color) 84%, transparent); }
.exit-screen figcaption { padding: 9px 5px 3px; color: var(--g-faint); font-size: 10px; text-align: center; }
.section-heading.compact { margin-bottom: 26px; }
.faq-list { border-top: 1px solid var(--g-line); }
.faq-list details { border-bottom: 1px solid var(--g-line); }
.faq-list summary { position: relative; padding: 18px 36px 18px 2px; color: var(--g-ink); cursor: pointer; font-size: 13px; font-weight: 700; list-style: none; }
.faq-list summary::-webkit-details-marker { display: none; }
.faq-list summary::after { content: '+'; position: absolute; right: 4px; color: var(--g-faint); font-size: 18px; }
.faq-list details[open] summary::after { content: '−'; }
.faq-list details p { max-width: 720px; padding: 0 36px 19px 2px; color: var(--g-soft); font-size: 12px; line-height: 1.8; }
.guide-footer { display: grid; grid-template-columns: 40px 1fr auto; align-items: center; gap: 13px; margin-top: 48px; padding-top: 22px; border-top: 1px solid var(--g-line); }
.footer-mark { display: grid; place-items: center; width: 36px; height: 36px; border-radius: 11px; background: #2d405e; color: #fff; font-size: 10px; font-weight: 800; }
.guide-footer strong { color: var(--g-ink); font-size: 12px; } .guide-footer p { margin-top: 3px; color: var(--g-faint); font-size: 10px; }
.guide-footer button { color: var(--g-soft); font-size: 11px; font-weight: 700; } .guide-footer button:hover { color: var(--primary-color); }
@media (max-width: 1080px) {
  .guide-hero { grid-template-columns: 1fr; }
  .hero-notes { display: grid; grid-template-columns: repeat(3, 1fr); }
  .hero-note { grid-template-columns: 1fr; border-right: 1px solid var(--g-line); }
  .workflow-steps { grid-template-columns: repeat(3, 1fr); }
  .guide-layout { grid-template-columns: 160px 1fr; gap: 30px; }
  .screen-figure, .exit-guide { grid-template-columns: 1fr; }
  .screen-frame { border-right: 0; border-bottom: 1px solid var(--g-line); }
}
@media (max-width: 760px) {
  .documentation-view .page-content { padding-inline: 18px; }
  .guide-hero { padding-top: 20px; }
  .hero-copy h1 { font-size: 26px; }
  .hero-notes, .workflow-steps, .capability-list { grid-template-columns: 1fr; }
  .hero-note, .workflow-steps li, .capability-item:nth-child(odd) { border-right: 0; }
  .workflow-steps li { border-bottom: 1px solid var(--g-line); } .workflow-steps svg { display: none; }
  .workflow-heading, .module-heading { align-items: flex-start; flex-direction: column; }
  .guide-layout { grid-template-columns: 1fr; } .guide-toc { display: none; }
  .guide-section { padding: 42px 0; }
  .start-list li { grid-template-columns: 32px 1fr; } .start-list a { grid-column: 2; }
  .module-title-row { align-items: flex-start; }
  .capability-item, .capability-item:nth-child(odd), .capability-item:nth-child(even) { padding: 18px 0; }
  .tip-strip { grid-template-columns: 1fr; gap: 9px; }
  .exit-guide { padding: 17px; }
  .guide-footer { grid-template-columns: 40px 1fr; } .guide-footer button { grid-column: 2; justify-self: start; }
}
@media (prefers-reduced-motion: reduce) { .documentation-view .page-content { scroll-behavior: auto; } }
</style>

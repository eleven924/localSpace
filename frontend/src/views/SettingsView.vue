<template>
  <div class="settings-view">
    <header class="header">
      <div class="header-left">
        <router-link to="/files" class="back-link">
          <span class="back-icon">←</span>
          返回
        </router-link>
        <div class="header-info">
          <h1>设置</h1>
          <p class="subtitle">配置 LocalSpace 的存储、AI、主题与内置文档。</p>
        </div>
      </div>
    </header>

    <div class="content">
      <div class="settings-grid">
        <section class="settings-section storage-section">
          <div class="section-header">
            <div class="section-icon">📁</div>
            <h2>存储目录</h2>
          </div>
          <StorageDirSelector
            @master-dir-added="handleDirAdded"
            @master-dir-removed="handleDirRemoved"
            @master-dir-default-changed="handleDirDefaultChanged"
          />
        </section>

        <section class="settings-section ai-section">
          <div class="section-header">
            <div class="section-icon">🤖</div>
            <h2>AI 配置</h2>
          </div>
          <AIConfigForm
            @config-saved="handleAIConfigSaved"
            @config-reset="handleAIConfigReset"
          />
        </section>

        <section class="settings-section open-section">
          <div class="section-header">
            <div class="section-icon">▶️</div>
            <h2>打开方式</h2>
          </div>
          <OpenWithConfig />
        </section>

        <section class="settings-section theme-section">
          <div class="section-header">
            <div class="section-icon">🎨</div>
            <h2>主题设置</h2>
          </div>
          <ThemeConfig
            @theme-changed="handleThemeChanged"
            @theme-reset="handleThemeReset"
          />
        </section>

        <section class="settings-section about-section">
          <div class="section-header">
            <div class="section-icon">ℹ️</div>
            <h2>文档与关于</h2>
          </div>

          <div class="about-content">
            <div class="app-info">
              <div class="app-logo">🗂️</div>
              <div>
                <h3>LocalSpace</h3>
                <p class="app-version">版本 1.0.0</p>
              </div>
            </div>

            <div class="app-description">
              <p>
                LocalSpace 是一款本地文件索引与整理工具，帮助你统一管理不同类型的素材，
                并通过标签、描述、缩略图与 AI 辅助提升检索效率。
              </p>
              <ul class="feature-list">
                <li>统一导入并归档图片、文档、音视频等常见文件</li>
                <li>基于标签、描述、关键词与文件类型进行检索</li>
                <li>支持 AI 自动补全标签与描述</li>
                <li>支持多主目录、多主题与缩略图缓存管理</li>
                <li>内置使用文档，随应用一起打包发布</li>
              </ul>
            </div>

            <div class="doc-card">
              <div>
                <h4>产品介绍与使用文档</h4>
                <p>包含核心功能说明、推荐使用流程和关键页面截图。</p>
              </div>
              <button class="doc-button" @click="handleViewDocumentation">
                查看文档
              </button>
            </div>

            <div class="app-links">
              <button class="app-link" @click="handleViewLicense">
                查看许可
              </button>
              <button class="app-link" @click="handleCheckUpdates">
                检查更新
              </button>
            </div>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import StorageDirSelector from '@/components/StorageDirSelector.vue'
import AIConfigForm from '@/components/AIConfigForm.vue'
import ThemeConfig from '@/components/ThemeConfig.vue'
import OpenWithConfig from '@/components/OpenWithConfig.vue'

const router = useRouter()

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
  window.alert('LocalSpace 当前以 MIT 风格开源协议进行分发，具体文本可在发布包中补充。')
}

const handleCheckUpdates = () => {
  window.alert('当前版本未接入在线更新服务，请关注后续发布说明。')
}
</script>

<style scoped>
.settings-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: var(--app-bg-color, var(--bg-color));
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background-color: var(--surface-color);
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.back-link {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  color: var(--text-color);
  text-decoration: none;
  border-radius: 8px;
  transition: all 0.2s ease;
}

.back-link:hover {
  background-color: var(--border-color);
}

.back-icon {
  font-size: 16px;
}

.header-info h1 {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color);
  margin: 0 0 4px 0;
}

.subtitle {
  font-size: 14px;
  color: var(--text-color);
  opacity: 0.75;
  margin: 0;
}

.content {
  flex: 1;
  padding: 20px 24px 24px;
  overflow-y: auto;
}

.settings-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  grid-template-areas:
    "storage ai"
    "open ai"
    "theme ai"
    "about ai";
  align-items: start;
  gap: 18px;
  max-width: 1400px;
  margin: 0 auto;
}

.storage-section {
  grid-area: storage;
}

.ai-section {
  grid-area: ai;
}

.open-section {
  grid-area: open;
}

.theme-section {
  grid-area: theme;
}

.about-section {
  grid-area: about;
}

.settings-section {
  background-color: var(--surface-color);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.06);
}

.section-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 18px;
  border-bottom: 1px solid var(--border-color);
  background-color: var(--bg-color);
}

.section-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: 8px;
  background-color: var(--surface-color);
  font-size: 19px;
}

.section-header h2 {
  font-size: 17px;
  font-weight: 600;
  color: var(--text-color);
  margin: 0;
}

.settings-section :deep(.storage-dir-selector),
.settings-section :deep(.ai-config-form),
.settings-section :deep(.open-with-config),
.settings-section :deep(.theme-config) {
  padding: 18px;
}

.settings-section :deep(.selector-header),
.settings-section :deep(.config-header) {
  margin-bottom: 14px;
}

.settings-section :deep(.selector-header h4),
.settings-section :deep(.config-header h4) {
  display: none;
}

.about-content {
  padding: 18px;
}

.app-info {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 16px;
}

.app-logo {
  font-size: 36px;
}

.app-info h3 {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-color);
  margin: 0 0 4px 0;
}

.app-version {
  font-size: 14px;
  color: var(--text-color);
  opacity: 0.7;
  margin: 0;
}

.app-description {
  margin-bottom: 16px;
}

.app-description p {
  font-size: 14px;
  color: var(--text-color);
  line-height: 1.7;
  margin: 0 0 14px 0;
}

.feature-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.feature-list li {
  font-size: 14px;
  color: var(--text-color);
  padding: 8px 0;
  border-bottom: 1px solid var(--border-color);
}

.feature-list li:last-child {
  border-bottom: none;
}

.doc-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
  padding: 16px;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  background: linear-gradient(135deg, rgba(33, 150, 243, 0.08), rgba(76, 175, 80, 0.05));
}

.doc-card h4 {
  margin: 0 0 6px 0;
  font-size: 16px;
  color: var(--text-color);
}

.doc-card p {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-color);
  opacity: 0.78;
}

.doc-button {
  min-width: 108px;
  padding: 10px 16px;
  border: none;
  border-radius: 8px;
  background-color: var(--primary-color);
  color: #fff;
  font-size: 14px;
  font-weight: 600;
}

.doc-button:hover {
  opacity: 0.92;
  box-shadow: 0 2px 8px var(--shadow-color);
}

.app-links {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}

.app-link {
  justify-content: center;
  padding: 10px 12px;
  background-color: var(--bg-color);
  color: var(--text-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 8px;
}

.app-link:hover {
  background-color: var(--border-color);
  transform: translateY(-1px);
}

@media (max-width: 1024px) {
  .settings-grid {
    grid-template-columns: 1fr;
    grid-template-areas:
      "storage"
      "ai"
      "theme"
      "about";
  }
}

@media (max-width: 768px) {
  .header {
    padding: 12px 16px;
  }

  .header-left {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .header-info h1 {
    font-size: 20px;
  }

  .content {
    padding: 16px;
  }

  .section-header {
    padding: 14px 16px;
  }

  .section-icon {
    width: 32px;
    height: 32px;
    font-size: 18px;
  }

  .section-header h2 {
    font-size: 16px;
  }

  .about-content {
    padding: 16px;
  }

  .doc-card {
    flex-direction: column;
    align-items: flex-start;
  }

  .doc-button {
    width: 100%;
  }
}

@media (max-width: 480px) {
  .header {
    padding: 10px 12px;
  }

  .content {
    padding: 12px;
  }

  .settings-section :deep(.storage-dir-selector),
  .settings-section :deep(.ai-config-form),
  .settings-section :deep(.theme-config),
  .about-content {
    padding: 14px;
  }

  .app-links {
    grid-template-columns: 1fr;
  }
}
</style>

<template>
  <div class="page-shell settings-view">
    <AppHeader />

    <div class="page-content">
      <div class="page-stack">
        <div class="settings-grid">
          <div class="settings-column">
            <section class="settings-section">
              <div class="section-header">
                <div class="section-icon">📁</div>
                <div>
                  <h2>存储目录</h2>
                  <p>管理主目录与默认目录</p>
                </div>
              </div>
              <StorageDirSelector
                @master-dir-added="handleDirAdded"
                @master-dir-removed="handleDirRemoved"
                @master-dir-default-changed="handleDirDefaultChanged"
              />
            </section>

            <section class="settings-section">
              <div class="section-header">
                <div class="section-icon">▶</div>
                <div>
                  <h2>打开方式</h2>
                  <p>为不同类型的文件指定打开应用</p>
                </div>
              </div>
              <OpenWithConfig />
            </section>

            <section class="settings-section">
              <div class="section-header">
                <div class="section-icon">🎨</div>
                <div>
                  <h2>主题设置</h2>
                  <p>控制主题模式、主色与背景</p>
                </div>
              </div>
              <ThemeConfig @theme-changed="handleThemeChanged" @theme-reset="handleThemeReset" />
            </section>
          </div>

          <div class="settings-column">
            <section class="settings-section">
              <div class="section-header">
                <div class="section-icon">🤖</div>
                <div>
                  <h2>AI 配置</h2>
                  <p>为导入流程提供标签和描述辅助</p>
                </div>
              </div>
              <AIConfigForm @config-saved="handleAIConfigSaved" @config-reset="handleAIConfigReset" />
            </section>

            <section class="settings-section">
              <div class="section-header">
                <div class="section-icon">🗂</div>
                <div>
                  <h2>存储规则</h2>
                  <p>控制归档结构与文件夹命名</p>
                </div>
              </div>
              <StorageLayoutConfig />
            </section>

            <section class="settings-section">
              <div class="section-header">
                <div class="section-icon">ℹ</div>
                <div>
                  <h2>文档与关于</h2>
                  <p>查看说明、许可与版本信息</p>
                </div>
              </div>

              <div class="about-content">
                <div class="app-info">
                  <div class="app-logo">🛰</div>
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

                <div class="doc-card">
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
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import AIConfigForm from '@/components/AIConfigForm.vue'
import AppHeader from '@/components/AppHeader.vue'
import OpenWithConfig from '@/components/OpenWithConfig.vue'
import StorageDirSelector from '@/components/StorageDirSelector.vue'
import StorageLayoutConfig from '@/components/StorageLayoutConfig.vue'
import ThemeConfig from '@/components/ThemeConfig.vue'

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
  window.alert('LocalSpace 当前以 MIT 风格开源协议分发，具体文本可在发布包中补充。')
}

const handleCheckUpdates = () => {
  window.alert('当前版本尚未接入在线更新服务，请关注后续发布说明。')
}
</script>

<style scoped>
.settings-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.08fr) minmax(0, 0.92fr);
  gap: 18px;
}

.settings-column {
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-width: 0;
}

.settings-section {
  overflow: hidden;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 16px;
  background-color: color-mix(in srgb, var(--surface-color) 92%, transparent);
  box-shadow: 0 10px 28px rgba(15, 23, 42, 0.06);
}

.section-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 18px;
  border-bottom: 1px solid var(--border-color);
  background-color: color-mix(in srgb, var(--bg-color) 80%, var(--surface-color));
}

.section-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: 10px;
  background-color: color-mix(in srgb, var(--surface-color) 92%, transparent);
  font-size: 18px;
}

.section-header h2 {
  font-size: 17px;
  color: var(--text-color);
}

.section-header p {
  margin-top: 4px;
  color: var(--text-faint);
  font-size: 12px;
}

.settings-section :deep(.storage-dir-selector),
.settings-section :deep(.ai-config-form),
.settings-section :deep(.open-with-config),
.settings-section :deep(.storage-layout-config),
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
  font-size: 34px;
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

.doc-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin: 16px 0;
  padding: 16px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 12px;
  background: linear-gradient(135deg, rgba(45, 140, 240, 0.08), rgba(76, 175, 80, 0.05));
}

.doc-card h4 {
  font-size: 16px;
  color: var(--text-color);
}

.doc-card p {
  margin-top: 6px;
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-faint);
}

.doc-button {
  width: auto;
}

.app-links {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

@media (max-width: 1080px) {
  .settings-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .section-header,
  .about-content,
  .settings-section :deep(.storage-dir-selector),
  .settings-section :deep(.ai-config-form),
  .settings-section :deep(.open-with-config),
  .settings-section :deep(.storage-layout-config),
  .settings-section :deep(.theme-config) {
    padding: 16px;
  }

  .doc-card {
    flex-direction: column;
    align-items: flex-start;
  }

  .app-links {
    grid-template-columns: 1fr;
  }
}
</style>

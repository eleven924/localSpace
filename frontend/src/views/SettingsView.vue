<template>
  <div class="settings-view">
    <AppHeader />

    <div class="content">
      <div class="settings-grid">
        <div class="settings-column">
          <section class="settings-section">
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

          <section class="settings-section">
            <div class="section-header">
              <div class="section-icon">▶</div>
              <h2>打开方式</h2>
            </div>
            <OpenWithConfig />
          </section>

          <section class="settings-section">
            <div class="section-header">
              <div class="section-icon">🎨</div>
              <h2>主题设置</h2>
            </div>
            <ThemeConfig @theme-changed="handleThemeChanged" @theme-reset="handleThemeReset" />
          </section>
        </div>

        <div class="settings-column">
          <section class="settings-section">
            <div class="section-header">
              <div class="section-icon">🤖</div>
              <h2>AI 配置</h2>
            </div>
            <AIConfigForm @config-saved="handleAIConfigSaved" @config-reset="handleAIConfigReset" />
          </section>

          <section class="settings-section">
            <div class="section-header">
              <div class="section-icon">🗂</div>
              <h2>存储规则</h2>
            </div>
            <StorageLayoutConfig />
          </section>

          <section class="settings-section">
            <div class="section-header">
              <div class="section-icon">ℹ</div>
              <h2>文档与关于</h2>
            </div>

            <div class="about-content">
              <div class="app-info">
                <div class="app-logo">🛰</div>
                <div>
                  <h3>LocalSpace</h3>
                  <p class="app-version">版本 1.0.0</p>
                </div>
              </div>

              <div class="app-description">
                <p>
                  LocalSpace 是一个本地文件整理与检索工具，帮助你把不同类型的素材放到统一入口中，
                  再通过标签、简介、合集和缩略图快速回看与管理。
                </p>

                <ul class="feature-list">
                  <li>统一导入并归档图片、文档、音视频等常见文件</li>
                  <li>基于标签、简介、关键词、文件类型和合集进行检索</li>
                  <li>支持 AI 自动补全标签与描述信息</li>
                  <li>支持多主目录、默认打开方式和缩略图缓存管理</li>
                  <li>内置使用文档，便于后续打包和交付</li>
                </ul>
              </div>

              <div class="doc-card">
                <div>
                  <h4>产品介绍与使用文档</h4>
                  <p>查看核心功能说明、推荐使用流程和页面示意。</p>
                </div>
                <button class="doc-button" @click="handleViewDocumentation">查看文档</button>
              </div>

              <div class="app-links">
                <button class="app-link" @click="handleViewLicense">查看许可</button>
                <button class="app-link" @click="handleCheckUpdates">检查更新</button>
              </div>
            </div>
          </section>
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
  window.alert('LocalSpace 当前以 MIT 风格开源协议进行分发，具体文本可在发布包中补充。')
}

const handleCheckUpdates = () => {
  window.alert('当前版本尚未接入在线更新服务，请关注后续发布说明。')
}
</script>

<style scoped>
.settings-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background:
    radial-gradient(circle at top left, rgba(33, 150, 243, 0.08), transparent 22%),
    var(--app-bg-color, var(--bg-color));
}

.content {
  flex: 1;
  padding: 12px 18px 18px;
  overflow-y: auto;
}

.settings-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.08fr) minmax(0, 0.92fr);
  gap: 18px;
  max-width: 1400px;
  margin: 0 auto;
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
  margin: 0;
  font-size: 17px;
  font-weight: 600;
  color: var(--text-color);
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
  margin: 0 0 4px;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-color);
}

.app-version {
  margin: 0;
  font-size: 14px;
  color: var(--text-color);
  opacity: 0.7;
}

.app-description {
  margin-bottom: 16px;
}

.app-description p {
  margin: 0 0 14px;
  font-size: 14px;
  line-height: 1.7;
  color: var(--text-color);
}

.feature-list {
  list-style: none;
  padding: 0;
  margin: 0;
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
  margin-bottom: 16px;
  padding: 16px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 12px;
  background: linear-gradient(135deg, rgba(33, 150, 243, 0.08), rgba(76, 175, 80, 0.05));
}

.doc-card h4 {
  margin: 0 0 6px;
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
  border-radius: 10px;
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
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.app-link {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  background-color: color-mix(in srgb, var(--bg-color) 78%, var(--surface-color));
  color: var(--text-color);
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.app-link:hover {
  background-color: var(--border-color);
  transform: translateY(-1px);
}

@media (max-width: 1024px) {
  .settings-grid {
    grid-template-columns: 1fr;
  }

  .settings-column {
    gap: 18px;
  }
}

@media (max-width: 768px) {
  .content {
    padding: 12px 14px 14px;
  }

  .section-header {
    padding: 14px 16px;
  }

  .section-icon {
    width: 32px;
    height: 32px;
    font-size: 17px;
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

@media (max-width: 640px) {
  .content {
    padding: 12px;
  }

  .settings-section :deep(.storage-dir-selector),
  .settings-section :deep(.ai-config-form),
  .settings-section :deep(.open-with-config),
  .settings-section :deep(.storage-layout-config),
  .settings-section :deep(.theme-config),
  .about-content {
    padding: 14px;
  }

  .app-links {
    grid-template-columns: 1fr;
  }
}
</style>

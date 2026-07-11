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
          <p class="subtitle">配置 LocalSpace 的各项功能</p>
        </div>
      </div>
    </header>

    <div class="content">
      <div class="settings-grid">
        <!-- 存储目录设置 -->
        <div class="settings-section">
          <div class="section-header">
            <div class="section-icon">📁</div>
            <h2>存储目录</h2>
          </div>
          <StorageDirSelector
            @dir-added="handleDirAdded"
            @dir-removed="handleDirRemoved"
            @dir-toggled="handleDirToggled"
          />
        </div>

        <!-- AI 配置 -->
        <div class="settings-section">
          <div class="section-header">
            <div class="section-icon">🤖</div>
            <h2>AI 配置</h2>
          </div>
          <AIConfigForm
            @config-saved="handleAIConfigSaved"
            @config-reset="handleAIConfigReset"
          />
        </div>

        <!-- 主题设置 -->
        <div class="settings-section">
          <div class="section-header">
            <div class="section-icon">🎨</div>
            <h2>主题设置</h2>
          </div>
          <ThemeConfig
            @theme-changed="handleThemeChanged"
            @theme-reset="handleThemeReset"
          />
        </div>

        <!-- 关于信息 -->
        <div class="settings-section">
          <div class="section-header">
            <div class="section-icon">ℹ️</div>
            <h2>关于</h2>
          </div>
          <div class="about-content">
            <div class="app-info">
              <div class="app-logo">🚀</div>
              <h3>LocalSpace</h3>
              <p class="app-version">版本 1.0.0</p>
            </div>
            <div class="app-description">
              <p>LocalSpace 是一个功能强大的本地文件索引管理软件，帮助您高效管理和查找本地文件。</p>
              <ul class="feature-list">
                <li>📁 多类型文件支持</li>
                <li>🔍 智能搜索功能</li>
                <li>🤖 AI 标签生成</li>
                <li>🎨 可自定义主题</li>
                <li>📦 存储目录管理</li>
              </ul>
            </div>
            <div class="app-links">
              <a href="#" class="app-link" @click.prevent="handleViewDocumentation">
                📚 文档
              </a>
              <a href="#" class="app-link" @click.prevent="handleViewLicense">
                ⚖️ 许可证
              </a>
              <a href="#" class="app-link" @click.prevent="handleCheckUpdates">
                🔄 检查更新
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useThemeStore } from '@/store/modules/theme'
import StorageDirSelector from '@/components/StorageDirSelector.vue'
import AIConfigForm from '@/components/AIConfigForm.vue'
import ThemeConfig from '@/components/ThemeConfig.vue'

const themeStore = useThemeStore()

// 事件处理函数
const handleDirAdded = () => {
  console.log('存储目录已添加')
}

const handleDirRemoved = () => {
  console.log('存储目录已删除')
}

const handleDirToggled = () => {
  console.log('存储目录状态已切换')
}

const handleAIConfigSaved = () => {
  console.log('AI 配置已保存')
}

const handleAIConfigReset = () => {
  console.log('AI 配置已重置')
}

const handleThemeChanged = () => {
  console.log('主题已更改')
}

const handleThemeReset = () => {
  console.log('主题已重置')
}

const handleViewDocumentation = () => {
  alert('文档功能即将推出')
}

const handleViewLicense = () => {
  alert('LocalSpace 是开源软件，使用 MIT 许可证')
}

const handleCheckUpdates = () => {
  alert('当前已是最新版本')
}
</script>

<style scoped>
.settings-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: var(--bg-color);
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
  opacity: 0.7;
  margin: 0;
}

.content {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}

.settings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(500px, 1fr));
  gap: 24px;
  max-width: 1400px;
  margin: 0 auto;
}

.settings-section {
  background-color: var(--surface-color);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  overflow: hidden;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-color);
  background-color: var(--bg-color);
}

.section-icon {
  font-size: 24px;
}

.section-header h2 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-color);
  margin: 0;
}

/* 关于内容样式 */
.about-content {
  padding: 24px;
}

.app-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 24px;
}

.app-logo {
  font-size: 64px;
  margin-bottom: 16px;
}

.app-info h3 {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color);
  margin: 0 0 8px 0;
}

.app-version {
  font-size: 14px;
  color: var(--text-color);
  opacity: 0.7;
  margin: 0;
}

.app-description {
  margin-bottom: 24px;
}

.app-description p {
  font-size: 14px;
  color: var(--text-color);
  line-height: 1.6;
  margin: 0 0 16px 0;
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

.app-links {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.app-link {
  padding: 12px 16px;
  background-color: var(--bg-color);
  color: var(--text-color);
  text-decoration: none;
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
  transform: translateX(4px);
}

@media (max-width: 1024px) {
  .settings-grid {
    grid-template-columns: 1fr;
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
    padding: 16px 20px;
  }

  .section-icon {
    font-size: 20px;
  }

  .section-header h2 {
    font-size: 16px;
  }

  .about-content {
    padding: 20px;
  }

  .app-logo {
    font-size: 48px;
  }

  .app-info h3 {
    font-size: 20px;
  }
}

@media (max-width: 480px) {
  .header {
    padding: 10px 12px;
  }

  .header-actions {
    display: none;
  }

  .content {
    padding: 12px;
  }

  .section-header {
    padding: 12px 16px;
    flex-direction: column;
    align-items: flex-start;
  }

  .section-icon {
    font-size: 18px;
  }

  .section-header h2 {
    font-size: 14px;
  }
}
</style>
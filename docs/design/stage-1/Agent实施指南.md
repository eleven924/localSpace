# LocalSpace 项目 - Agent 实施指南

## 项目概述

LocalSpace 是一个本地文件索引软件，使用 Golang + Vue + SQLite + Wails 开发。

## 当前状态

### 已完成 (Phase 1-3, 5-9)
✅ 项目基础搭建
✅ 目录结构创建
✅ Go 后端框架
✅ Vue 前端框架
✅ 数据库设计
✅ API 接口定义
✅ 主题系统
- ✅ Phase 2: 数据库详细实现（已完成，所有测试通过）
- Phase 3: 后端核心功能
✅ 前端核心组件 (Phase 5)
✅ 文件导入功能 (Phase 6)
✅ 设置页面 (Phase 7)
✅ 元数据提取 (Phase 8)
✅ 文件预览 (Phase 9)

### 待实施 (Phase 10-12)
- ✅ Phase 4: 前端基础框架（已完成，包括UI组件库）
- Phase 10: 跨平台打包
- Phase 11: 测试和优化
- Phase 12: 发布和文档

## Phase 5 完成总结

### 已实现组件

#### 1. FileCard.vue
- 文件卡片显示组件
- 支持缩略图/图标显示
- 显示文件名、标签、描述、大小、创建时间
- 响应式设计

#### 2. FileList.vue
- 响应式网格布局（2-5列自适应）
- 自动调整列数（基于窗口宽度）
- 空状态显示
- 自定义滚动条

#### 3. SearchBar.vue
- 实时搜索输入
- 可配置防抖延迟
- 清除搜索按钮
- 回车键搜索

#### 4. FileTypeFilter.vue
- 文件类型筛选按钮组
- 显示各类型文件数量
- 高亮当前选中类型

#### 5. ThemeToggle.vue
- 主题模式切换（浅色/深色）
- 主题颜色选择器
- 重置主题功能

### 视图更新
- FilesView.vue 重构，集成所有新组件
- main.ts 添加主题初始化逻辑

详细内容见：`Phase5_完成总结.md`

## Phase 6 完成总结

### 已实现组件

#### 1. FileSelector.vue
- 文件选择组件
- 调用 Wails 文件选择器
- 显示文件详细信息
- 文件类型自动识别
- 文件大小获取
- 移除文件功能

#### 2. FileMetaForm.vue
- 文件元数据表单
- 文件名、标签、描述编辑
- 实时标签预览
- AI 分析集成
- 进度提示弹窗
- 表单验证

#### 3. ImportView.vue 重构
- 集成 FileSelector 和 FileMetaForm
- 两阶段导入流程
- 文件摘要显示
- 更换文件功能
- 自动 AI 分析

### 工具函数

#### fileValidator.ts
- 文件基本信息验证
- 文件元数据验证
- 完整导入验证
- 文件类型识别
- 扩展名支持检查

详细内容见：`Phase6_完成总结.md`

## Phase 7 完成总结

### 已实现组件

#### 1. StorageDirSelector.vue
- 存储目录管理组件
- 目录列表显示
- 添加/删除/启用目录
- 目录浏览功能
- 容量限制设置

#### 2. AIConfigForm.vue
- AI 配置表单
- AI 功能开关
- API Key 管理
- 模型选择
- 连接测试
- 配置验证

#### 3. ThemeConfig.vue
- 主题配置组件
- 主题模式切换
- 预设颜色选择
- 自定义颜色
- 背景图片上传
- 实时预览

#### 4. SettingsView.vue 重构
- 集成所有设置组件
- 模块化布局
- 关于信息页面
- 外部链接

详细内容见：`Phase7_完成总结.md`

## Phase 8 完成总结

### 已实现工具

#### 1. image_metadata.go
- 图片元数据提取工具
- 尺寸和格式检测
- 色彩空间识别
- 支持 JPG, PNG, GIF, BMP, WEBP

#### 2. video_metadata.go
- 视频元数据提取工具
- 时长和帧率提取
- 编解码器识别
- 依赖 ffprobe 工具

#### 3. document_metadata.go
- 文档元数据提取工具
- 页数和标题提取
- 支持多种文档格式
- 文本和 Markdown 支持

### 服务集成

#### file_service.go
- 更新元数据提取方法
- 类型特定的提取逻辑
- 统一的元数据接口
- 错误处理和容错

详细内容见：`Phase8_完成总结.md`

## 为后续 Agent 的说明

### 重要注意事项

1. **依赖安装**
   - 在开始任何开发前，先运行：
     ```bash
     go mod download
     cd frontend && npm install
     ```

2. **开发模式**
   - 使用 `wails dev` 启动开发服务器
   - 前端修改会自动热重载
   - 后端修改需要重启

3. **数据库位置**
   - 开发环境：`storage/data/localspace.db`
   - 可以直接使用 SQLite 工具查看和调试

4. **系统依赖**
   - **ffprobe**: Phase 8 视频元数据提取需要
   - 安装方法见 Phase 8 完成总结

5. **导出方法规则**
   - Go 中需要导出给前端的方法必须使用大写字母开头
   - 例如：`Startup` 而不是 `startup`

### 项目结构理解

```
LocalSpace/
├── main.go              # Wails 入口，绑定 App 方法
├── app/
│   ├── app.go           # App 结构体，包含所有导出方法
│   ├── database/        # 数据库初始化和表结构
│   ├── models/          # 数据模型定义
│   ├── repositories/    # 数据访问层（CRUD）
│   ├── services/        # 业务逻辑层
│   └── utils/           # 工具函数 (Phase 8 新增)
│       ├── image_metadata.go
│       ├── video_metadata.go
│       └── document_metadata.go
├── frontend/
│   └── src/
│       ├── components/  # Vue 组件 (Phase 5-7)
│       ├── views/       # 页面组件
│       ├── store/       # Pinia 状态管理
│       ├── utils/       # 工具函数
│       └── types/       # TypeScript 类型定义
└── wails.json           # Wails 配置
```

### API 调用模式

**前端调用后端**：
```typescript
// 通过 window.wails.go.main.App 调用
await window.wails.go.main.App.GetFiles(1, 50, 'video')
```

**后端导出方法**：
```go
// 在 app.go 中定义导出方法
func (a *App) GetFiles(page, pageSize int, fileType string) ([]*models.File, error) {
    return a.fileService.ListFiles(services.FileFilter{
        Page:     page,
        PageSize: pageSize,
        FileType: fileType,
    })
}
```

### 分层架构模式

1. **Handler 层 (app.go)**
   - 处理前端请求
   - 参数验证
   - 调用 Service 层

2. **Service 层 (services/)**
   - 业务逻辑
   - 调用多个 Repository
   - 调用工具函数 (Phase 8 新增)
   - 事务管理

3. **Repository 层 (repositories/)**
   - 数据库 CRUD 操作
   - SQL 查询
   - 数据映射

4. **Model 层 (models/)**
   - 数据结构定义
   - JSON 序列化标签

5. **Utils 层 (utils/)**
   - 元数据提取 (Phase 8 新增)
   - 工具函数

### 错误处理规范

```go
// Service 层：包装错误
func (s *Service) DoSomething() error {
    if err != nil {
        return fmt.Errorf("failed to do something: %w", err)
    }
    return nil
}

// Repository 层：返回原始错误或明确错误
func (r *Repo) Create() error {
    if err != nil {
        return fmt.Errorf("failed to create: %w", err)
    }
    return nil
}

// Utils 层：返回详细错误
func (u *Utils) ExtractMetadata() error {
    if err != nil {
        return fmt.Errorf("failed to extract metadata: %w", err)
    }
    return nil
}
```

### 前端状态管理模式

```typescript
// 使用 Pinia store
const store = useFilesStore()
await store.loadFiles('all')

// Store 中处理错误
try {
    const result = await api.file.list(...)
    files.value = result
} catch (err) {
    error.value = err.message
}
```

### CSS 变量使用

```css
/* 使用 CSS 变量实现主题 */
background-color: var(--bg-color);
color: var(--text-color);
border-color: var(--border-color);
```

### 元数据提取

```go
// 自动提取元数据
metadata, err := fileService.ExtractMetadata(filePath, fileType)
if err != nil {
    // 元数据提取失败不影响文件导入
    log.Printf("Warning: Failed to extract metadata: %v", err)
}
```

### 存储目录管理

```go
// 根据文件类型获取存储路径
path, err := storageService.GetStoragePath(fileType)

// 检查存储空间
hasSpace, err := storageService.CheckStorageSpace(fileSize)
```

## 各 Phase 实施要点

### Phase 2: 数据库详细实现
- ✅ 添加数据库连接池配置
- ✅ 实现数据库迁移脚本
- ✅ 添加数据验证
- ✅ 编写单元测试
- ✅ 所有测试通过（10/10）

### Phase 3: 后端核心功能
- ✅ 完善 FileService 的所有方法 (Phase 6-8 完成)
- ✅ 实现文件复制/移动逻辑 (Phase 6 完成)
- ✅ 添加错误处理和日志
- ✅ 实现元数据提取 (Phase 8 完成)
- ✅ 实现文件去重检测 (Phase 3 完成)

### Phase 4: 前端基础框架
- ✅ Vue Router 路由配置 (Phase 4 完成)
- ✅ Pinia 状态管理 (Phase 4 完成)
- ✅ API 封装 (Phase 4 完成)
- ✅ 主题样式系统 (Phase 4 完成)
- ✅ 响应式布局优化 (Phase 4 完成)
- ✅ 基础UI组件库 (Phase 4 完成)
- ✅ 移除有问题的 TypeScript 配置 (Phase 4 完成)

### Phase 5-6: 文件导入和核心组件 (已完成)
- ✅ 文件选择对话框 (Phase 6 完成)
- ✅ 完善导入流程 (Phase 6 完成)
- ✅ 添加进度提示 (Phase 6 完成)
- ✅ 实现文件验证 (Phase 6 完成)
- ✅ 前端核心组件 (Phase 5 完成)

### Phase 7: 设置页面 (已完成)
- ✅ 存储目录管理 (Phase 7 完成)
- ✅ AI 配置界面 (Phase 7 完成)
- ✅ 主题设置 (Phase 7 完成)
- ✅ 配置验证 (Phase 7 完成)

### Phase 9 完成总结

### 已实现组件

#### 1. 缩略图生成工具
- 图片缩略图生成（支持 JPG、PNG、GIF、BMP、WEBP）
- 视频缩略图生成（使用 ffmpeg）
- 文档缩略图生成（占位图）
- 音频缩略图生成（图标占位）
- 缓存管理和清理功能

#### 2. 缩略图缓存服务
- LRU 缓存淘汰机制
- 自动清理过期缓存
- 缓存大小限制（默认 100MB）
- 缓存时间限制（默认 7 天）
- 异步缩略图生成

#### 3. 文件服务集成
- 文件导入时自动生成缩略图
- 文件列表时异步生成缺失缩略图
- 文件删除时自动清理缩略图
- 缩略图路径更新机制

#### 4. App 导出方法
- GenerateThumbnail(fileID)
- GetThumbnailCacheInfo()
- ClearThumbnailCache()
- GetThumbnailCacheStats()

#### 5. 前端 API 集成
- 缩略图 API 调用封装
- TypeScript 类型定义
- FileCard 组件支持缩略图显示

详细内容见：`Phase9_完成总结.md`

### Phase 10: 跨平台打包
- ✅ 缩略图功能完成
- 配置构建脚本
- 测试不同平台
- 优化打包体积
- 生成安装程序

### Phase 11: 测试和优化
- 编写单元测试
- 编写集成测试
- 性能优化
- 内存泄漏检测

### Phase 12: 发布和文档
- 编写用户手册
- 编写开发文档
- 准备发布版本
- 创建更新日志

## 常见问题

### Q: 如何添加新的导出方法？
A: 在 `app/app.go` 中添加方法，确保方法名首字母大写。

### Q: 如何添加新的数据库表？
A: 在 `app/database/sqlite.go` 的 `createTables` 函数中添加 SQL。

### Q: 如何修改主题？
A: 使用 `themeStore.setThemeMode('dark' | 'light')` 切换主题。

### Q: 如何调用元数据提取？
A: 文件导入时会自动调用 `fileService.ExtractMetadata()`。

### Q: 如何调试数据库？
A: 使用 SQLite 工具打开 `storage/data/localspace.db`。

### Q: ffprobe 未找到？
A: 需要安装 FFmpeg，详见 Phase 8 完成总结。

### Q: 如何使用 Phase 5-8 的新功能？
A: 参考各 Phase 完成总结文档中的使用示例。

## 推荐开发顺序

1. ✅ 先完成后端核心功能，确保数据流正常 (Phase 1-3, 6-8)
2. ✅ 再完成前端界面，确保用户交互流畅 (Phase 4-7)
3. 最后集成预览和高级功能 (Phase 9)

## 代码规范

### 1. Go 代码
- 使用 `gofmt` 格式化
- 添加错误处理
- 添加必要的注释
- Utils 层函数应该是无状态的

### 2. Vue 代码
- 使用 Composition API
- TypeScript 类型定义
- 组件单一职责

### 3. CSS 代码
- 使用 CSS 变量
- 响应式设计
- 主题适配

## 测试策略

### 1. 单元测试
- Repository 层测试
- Service 层测试
- Utils 层测试 (Phase 8 新增)

### 2. 集成测试
- API 接口测试
- 数据流测试
- 用户流程测试
- 元数据提取测试 (Phase 8 新增)

### 3. E2E 测试
- 完整流程测试
- 跨平台测试
- 性能测试

## 联系和反馈

如有问题，请查看：
- 分阶段开发方案：`分阶段开发方案.md`
- Phase 5 总结：`Phase5_完成总结.md`
- Phase 6 总结：`Phase6_完成总结.md`
- Phase 7 总结：`Phase7_完成总结.md`
- Phase 8 总结：`Phase8_完成总结.md`
- Phase 1 总结：`Phase1_完成总结.md`
- 原始设计：`oriDesign/LocalSpace设计方案.md`
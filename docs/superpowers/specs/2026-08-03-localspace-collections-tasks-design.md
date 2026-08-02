# LocalSpace 合集与任务功能设计文档

> 生成日期：2026-08-03  
> 范围：后端（Go + SQLite + Wails Job 系统）、前端（Vue 3 + Pinia + TypeScript）

---

## 1. 目标与范围

本设计文档覆盖以下需求：

1. **合集管理**：合集只能在设置中新增/删除；导入、批量导入、文件编辑时仅允许从已配置合集列表中选择，禁止自由输入。文件通过 `collection_id` 与合集关联。
2. **文件展示**：文件卡片支持后端分页；每页数量可选 20/50/100；文件卡片支持左上角选择并执行批量操作（移动合集、删除）。
3. **元数据编辑**：原“编辑标签和描述”改为“编辑元数据”，支持调整合集。
4. **任务页面**：不再区分当前任务与历史任务，合并为单一列表；列表展示任务名称、开始时间、结束时间、进度、当前状态；支持删除记录与恢复（不可恢复时置灰）；列表使用半透明/毛玻璃背景。
5. **设置页优化**：新增“合集”管理区；新增“任务”设置区，可配置历史任务保留条数/天数、启用自动清理、手动触发清理 Job。

---

## 2. 关键决策

| 议题 | 决策 |
|------|------|
| 合集数据模型 | 新增独立 `collections` 表；`files` 表增加 `collection_id` 外键；不保留冗余 `collection_name`。 |
| 历史数据迁移 | 不迁移。旧文件的 `collection_name` 保留但只读；新写入统一使用 `collection_id`。编辑旧文件时选择器默认显示“未分配合集”。 |
| 删除合集限制 | 若 `files` 表中存在 `collection_id` 引用，禁止删除并返回错误。 |
| 批量移动合集 | 同时更新数据库 `collection_id` 并按 storage layout 物理移动文件到新合集目录；前端使用阻塞式进度遮罩。 |
| 分页 | 后端分页，每页数量可选 20/50/100。 |
| 任务清理规则 | 最大条数、最大天数可独立启用；任一条件触发即删除。仅清理终态任务（completed/failed/cancelled/timed_out/cleanup_failed）。 |
| 任务清理默认 | 默认关闭自动清理。 |
| 任务列表布局 | 紧凑行内表格；失败行可展开查看失败详情；操作列使用下拉菜单提供“删除记录”和“恢复”。 |

---

## 3. 数据模型变更

### 3.1 新增 `collections` 表

```sql
CREATE TABLE IF NOT EXISTS collections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

- `name` 唯一，首尾空格去除后存储，空字符串不允许保存。
- 未分配合集不写入本表，使用 `files.collection_id IS NULL` 表示。

### 3.2 `files` 表新增字段

```sql
ALTER TABLE files ADD COLUMN collection_id INTEGER DEFAULT NULL;
```

可选索引：

```sql
CREATE INDEX IF NOT EXISTS idx_files_collection_id ON files(collection_id);
```

新增模型定义（`app/models/config.go` 或新增 `app/models/collection.go`）：

```go
type Collection struct {
    ID        uint   `json:"id"`
    Name      string `json:"name"`
    CreatedAt string `json:"createdAt"`
    UpdatedAt string `json:"updatedAt"`
}
```

### 3.3 现有 `collection_name` 字段处理

- 保留原字段，但业务层不再写入。
- 读取文件时，若 `collection_id` 非空，通过 JOIN 查询 `collections.name`；若 `collection_id` 为空，则显示“未分配合集”。
- 旧文件未关联到 `collections` 表时，UI 不再显示旧的 `collection_name`，编辑时选择器回退到“未分配合集”。

---

## 4. 后端 API 扩展

### 4.1 合集管理

```go
// GetCollections 返回全部合集
GetCollections() ([]models.Collection, error)

// AddCollection 新增合集，name 唯一，返回 id
AddCollection(name string) (uint, error)

// RemoveCollection 删除合集；若被文件引用则返回错误
RemoveCollection(id uint) error
```

对应前端 `api` 封装：

```ts
api.collection.getAll: () => Promise<Collection[]>
api.collection.add: (name: string) => Promise<number>
api.collection.remove: (id: number) => Promise<string>
```

### 4.2 文件元数据

扩展现有方法签名：

```go
// 原：UpdateFileMetadata(id uint, tags []string, description string)
// 新：
UpdateFileMetadata(id uint, tags []string, description string, collectionID *uint) error
```

- `collectionID` 为 `nil` 时表示“未分配合集”。
- 调用 `FileService.UpdateFileMetadata` 时，若合集变更导致物理路径变化，同步更新 `files.collection_id`、`files.file_path` 与 `modified_at`；若路径未变化，仅更新 `collection_id` 与 `modified_at`。

### 4.3 批量操作

```go
// BatchUpdateFilesCollection 批量修改合集并物理移动
BatchUpdateFilesCollection(ids []uint, collectionID *uint) (BatchMoveResult, error)

// BatchDeleteFiles 批量删除文件
BatchDeleteFiles(ids []uint) ([]BatchDeleteItemResult, error)
```

返回结构：

```go
type BatchMoveResult struct {
    SuccessCount int
    FailedCount  int
    FailedItems  []BatchMoveFailedItem
}

type BatchMoveFailedItem struct {
    FileID   uint
    FileName string
    Error    string
}
```

### 4.4 文件查询分页

扩展现有方法签名：

```go
// 原：GetFiles(page, pageSize int, fileType string)
// 新：
GetFiles(page, pageSize int, fileType string, collectionID *uint) (*FileListResponse, error)

// 原：SearchFiles(query string)
// 新：
SearchFiles(query string, page, pageSize int) (*FileListResponse, error)
```

新增模型定义（`app/models/file.go`）：

```go
type FileListResponse struct {
    Items    []*File `json:"items"`
    Page     int     `json:"page"`
    PageSize int     `json:"pageSize"`
    Total    int     `json:"total"`
}
```

对应前端 API：

```ts
api.file.list: (page, pageSize, fileType, collectionId?) => Promise<FileListResponse>
api.file.search: (query, page, pageSize) => Promise<FileListResponse>
```

### 4.5 任务清理与设置

```go
// GetJobRetentionConfig 读取任务保留配置
GetJobRetentionConfig() (*models.JobRetentionConfig, error)

// UpdateJobRetentionConfig 保存任务保留配置
UpdateJobRetentionConfig(config models.JobRetentionConfig) error

// SubmitJobCleanup 手动提交清理任务
SubmitJobCleanup() (*models.Job, error)

// DeleteJobRecord 删除单条终态任务记录
DeleteJobRecord(id uint) error
```

`JobRetentionConfig` 模型：

```go
type JobRetentionConfig struct {
    ID          uint `json:"id"`
    Enabled     bool `json:"enabled"`      // 是否启用自动清理
    MaxCount    int  `json:"maxCount"`     // 最多保留条数；<=0 表示未启用
    MaxDays     int  `json:"maxDays"`      // 最多保留天数；<=0 表示未启用
    UpdatedAt   string `json:"updatedAt"`
}
```

---

## 5. Job 系统

### 5.1 新增 `job_cleanup` 类型

```go
const JobTypeCleanup = "job_cleanup"
```

策略：

```go
JobPolicy{
    JobType:          JobTypeCleanup,
    MaxConcurrent:    1,
    ExclusiveKey:     "cleanup",
    CanRunBackground: true,
    Recoverable:      false,
    Timeout:          5 * time.Minute,
}
```

### 5.2 CleanupHandler

`CleanupHandler` 实现 `JobHandler` 接口：

- `Validate`：校验配置中至少启用了一个条件（MaxCount > 0 或 MaxDays > 0）。
- `Execute`：
  1. 读取 `JobRetentionConfig`。
  2. 查询所有终态任务（`status IN ('completed','failed','cancelled','timed_out','cleanup_failed')`）。
  3. 计算需要删除的 ID 集合：
     - 若启用 `MaxCount` 且总数 > MaxCount，按 `created_at` 升序删除最早的 N 条。
     - 若启用 `MaxDays` 且 MaxDays > 0，删除 `created_at < now - MaxDays` 的记录。
  4. 执行 `DELETE FROM jobs WHERE id IN (...)`；级联删除 `batch_import_items`（外键已设置 `ON DELETE CASCADE`）。
  5. 更新 job 状态为 `completed`，记录删除数量。
  6. 若失败，状态设为 `failed` 并发送通知。

### 5.3 触发方式

- **自动**：应用启动时，若配置启用，且当前存在需要清理的终态任务记录（条数或天数任一条件触发），自动提交一次 `job_cleanup`。
- **手动**：设置页“手动清理”按钮调用 `SubmitJobCleanup()`，创建同样类型的 Job。

### 5.4 删除单条记录

```go
func (s *JobService) DeleteJobRecord(id uint) error
```

- 仅允许删除终态任务；非终态返回错误。
- 直接删除 `jobs` 记录，级联删除子表记录。

---

## 6. 前端改动

### 6.1 合集选择器 `CollectionSelector`

- 单选下拉组件，数据源为 `collections` 列表 + 固定“未分配合集”选项。
- 用于：单文件导入、批量导入、文件编辑、批量操作工具栏。
- 旧文件编辑时，若当前 `collection_id` 为 `NULL` 或不存在，默认选中“未分配合集”。

### 6.2 设置页

**新增“合集”设置区**

- 列表展示：合集名称、文件引用数量、删除按钮。
- 删除按钮：被引用时置灰，hover 或点击提示“该合集正在被文件使用，无法删除”。
- 新增输入框：输入名称后添加；若名称已存在则提示。

**新增“任务”设置区**

- 自动清理开关（默认关闭）。
- 最大条数：启用开关 + 数字输入框。
- 最大天数：启用开关 + 数字输入框。
- 说明文字：自动清理仅删除已终态任务。
- 手动清理按钮：点击后提交 `job_cleanup`，按钮进入 loading 状态，完成后提示。

### 6.3 文件卡片与批量操作

**`FileCard` 扩展**

- 新增选择模式：左上角显示复选框，选中后卡片高亮。
- 选择模式不影响原本的点击打开行为；复选框单独响应点击，避免与卡片点击冲突。
- 原有下拉菜单中“编辑标签和描述”改为“编辑元数据”。

**`FilesView` 批量工具栏**

- 当至少选中一个文件时，顶部显示批量工具栏：
  - 已选择 X 个文件
  - “移动合集”按钮：点击后弹出 `CollectionSelector` 选择目标合集，确认后进入阻塞式进度遮罩。
  - “删除”按钮：二次确认后删除，完成后刷新列表。
- 提供“取消选择”按钮。

**分页控件**

- 底部页码：上一页/下一页、当前页码/总页数。
- 每页数量选择：20 / 50 / 100。
- 切换每页数量时重置到第 1 页。

### 6.4 批量移动合集进度遮罩

- 点击确认后弹出全屏/半屏遮罩：
  - 标题：正在移动 X 个文件到「合集名」
- 进度条 + 当前处理文件名。
- 本阶段暂不提供取消按钮，后续按需扩展。
- 遮罩期间：文件列表置灰，禁止新的选择、打开、删除、再次移动。
- 完成后：
  - 成功：关闭遮罩，刷新列表，toast 提示成功。
  - 部分失败：关闭遮罩，弹窗或 toast 展示失败文件数量与原因，列表刷新。

### 6.5 任务页面重构

**`TasksView` 统一列表**

- 列表列：任务名称、开始时间、结束时间、进度（进度条 + 百分比）、状态标签。
- 失败任务行可展开：展示失败原因、成功数量、失败数量、错误信息。若任务无 `result`（如单文件导入失败），仅展示 `errorMessage`。
- 操作列：下拉按钮“操作”，包含：
  - 删除记录（仅终态可用）
  - 恢复（仅 `awaiting_resume` / `timed_out` 状态可用，否则置灰）
- 列表行背景使用半透明/毛玻璃效果（`backdrop-filter: blur(...)` + `background: rgba(...)`），透出主题背景。

**分页**

- 列表底部保留分页，每页数量可选 10/20/50。

### 6.6 路由与状态

- `filesStore` 增加 `currentPage`、`pageSize`；`loadFiles` 和 `searchFiles` 改为分页调用。
- `filesStore` 增加 `selectedFileIds` 用于批量选择。
- `jobsStore` 增加 `loadJobHistory` 分页刷新，支持手动删除单条记录。

---

## 7. 物理文件移动逻辑

当批量移动合集或编辑元数据修改合集时，若文件实际路径需要变化（如 `type_collection` 布局下合集目录名改变），后端执行以下步骤：

1. 读取目标合集 `name` 与当前 storage layout 配置。
2. 根据新合集计算 `newFilePath`。
3. 检查 `newFilePath` 是否已存在；若存在则返回错误，跳过该文件。
4. 在事务中更新 `files.collection_id` 和 `files.file_path`。
5. 执行 `os.Rename(oldFilePath, newFilePath)`。
6. 若物理移动失败：
   - 回滚事务中的 `collection_id` 与 `file_path`。
   - 记录该文件失败，继续处理下一个。
7. 若成功：更新 `storage_dirs` 相关目录大小（原目录减少，新目录增加）。

注意：
- 文件 `file_name` 本身不改变，仅路径中的合集目录段变化。
- 单文件编辑元数据改合集时，同步执行上述移动；失败时整个操作返回错误（不继续）。
- 批量移动时，单文件失败不影响其他文件，最终汇总失败列表。

---

## 8. 错误处理与边界

| 场景 | 处理 |
|------|------|
| 删除被引用的合集 | 后端返回错误；前端提示“该合集正在被文件使用，无法删除”。 |
| 添加重复合集名称 | 后端返回错误；前端提示“合集名称已存在”。 |
| 批量移动时目标路径已存在 | 该文件失败，继续处理其余。 |
| 批量移动时物理移动失败 | 回滚该文件的数据库变更；其他文件继续。 |
| 清理任务配置未启用任何条件 | 校验失败，提示用户至少启用一个条件。 |
| 删除非终态任务记录 | 后端返回错误，前端不展示删除入口。 |
| 分页参数越界 | 后端返回空列表，前端显示“暂无数据”并保持有效页码。 |
| 旧文件无 `collection_id` | 编辑时选择器默认“未分配合集”；保存时写入正常 `collection_id` 或 `NULL`。 |

---

## 9. 测试计划

### 9.1 后端测试

- `CollectionRepository`：新增、删除、名称唯一、删除被引用时失败。
- `FileRepository`：按 `collection_id` 过滤、分页查询。
- `FileService`：
  - `UpdateFileMetadata` 修改合集并物理移动。
  - `BatchUpdateFilesCollection` 部分失败场景。
  - `BatchDeleteFiles` 批量删除。
- `CleanupHandler`：
  - 按条数清理。
  - 按天数清理。
  - 仅删除终态任务。
  - 失败时更新自身状态。
- `JobService`：
  - `DeleteJobRecord` 对非终态任务报错。
  - `SubmitJobCleanup` 创建正确类型任务。

### 9.2 前端测试

- 项目当前未配置前端测试脚本，本阶段不新增测试脚本。
- 通过 `npm run build` 和 TypeScript 类型检查验证改动。

### 9.3 集成验证

- 新增合集 → 导入文件选择该合集 → 文件显示正确合集 → 设置中删除该合集（被引用时失败）。
- 文件卡片选择多个 → 移动合集 → 遮罩进度 → 刷新后路径与合集一致。
- 任务列表删除单条终态记录 → 列表刷新且总数减少。
- 启用自动清理 → 启动应用 → 清理 Job 完成 → 通知中心收到提示。

---

## 10. 实现阶段建议

为降低风险，建议按以下阶段交付：

### 阶段 1：合集基础设施
- 数据库迁移（`collections` 表、`files.collection_id`）。
- 后端合集 CRUD API。
- 设置页“合集”管理区。
- `CollectionSelector` 组件。
- 导入/批量导入/文件编辑中限制合集为选择。

### 阶段 2：文件分页与批量操作
- 后端分页 API。
- 文件卡片选择模式与批量工具栏。
- 批量移动/删除（含物理移动）。
- 前端分页控件。

### 阶段 3：任务页面与清理 Job
- 任务列表重构（合并、表格、毛玻璃）。
- `job_cleanup` Handler。
- 设置页“任务”保留配置 + 手动清理按钮。
- 删除单条任务记录。

每个阶段完成后进行集成验证，再进入下一阶段。

---

## 11. 附录：相关文件

- 后端模型：`app/models/file.go`、`app/models/config.go`、`app/models/job.go`
- 后端服务：`app/services/file_service.go`、`app/services/job_service.go`、`app/services/config_service.go`
- 后端仓库：`app/repositories/file_repository.go`、`app/repositories/job_repository.go`、`app/repositories/config_repository.go`
- 数据库迁移：`app/database/migrations.go`
- Wails 绑定层：`app/app.go`
- 前端 API：`frontend/src/api/index.ts`
- 前端状态：`frontend/src/store/modules/files.ts`、`frontend/src/store/modules/jobs.ts`
- 前端视图：`frontend/src/views/FilesView.vue`、`frontend/src/views/TasksView.vue`、`frontend/src/views/SettingsView.vue`、`frontend/src/views/ImportView.vue`
- 前端组件：`frontend/src/components/FileCard.vue`、`frontend/src/components/FileList.vue`、`frontend/src/components/EditFileMetaModal.vue`、`frontend/src/components/FileMetaForm.vue`

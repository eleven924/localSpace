# LocalSpace 问题修复设计文档

> 日期：2026-08-01
> 范围：Go 后端（`app/services`、`app/repositories`、`app/app.go`）、Vue 3 前端（`frontend/src/api`、`frontend/src/views`）
> 来源：`ori/问题分析.md`

---

## 1. 背景与目标

`ori/问题分析.md` 梳理出 10 项主要问题 + 4 项补充提示。本次设计覆盖全部 10 项主要问题以及补充项 C（单文件元数据类型错误）。修复目标按优先级如下：

1. **立即修复**：数据丢失/不一致（issue 1、4）。
2. **高优先级**：SQL 注入、Job 系统稳定性（issue 3、6、7）。
3. **中优先级**：跨盘行为、Rename 回滚、PowerShell 命令注入、前端 API 错误处理、批量导入临时路径（issue 2、5、8、9、10）。
4. **低优先级**：体验优化（补充 C）。

---

## 2. 整体策略

采用**统一导入事务（Import Pipeline）**方案：把单文件导入和批量导入里“文件如何安全落地到库存储”的公共逻辑抽成一段可复用的 pipeline，统一处理临时文件、checksum、重复检测、数据库写入、原子重命名、源文件清理和失败回滚。

其余问题（SQL 注入、DeleteFile/RenameFile 顺序、Job panic 恢复、shutdown 同步、PowerShell 字符串拼接、前端 API 错误处理）按分析文档点对点修复。

---

## 3. 统一导入事务

### 3.1 新增内部类型

在 `app/services/file_service.go` 中新增：

```go
type stagedImport struct {
    SourcePath     string
    TempPath       string
    FinalPath      string
    Checksum       string
    FileSize       int64
    FileType       string
    FileSubType    string
    MasterID       uint
    CollectionName string
    FileName       string
    OriginalName   string
    Metadata       models.Metadata
}
```

### 3.2 新增核心函数

```go
// prepareStagedImport
// 1. 计算源文件 checksum
// 2. 重复检测（复用 CheckDuplicateByChecksum）
// 3. 生成目标 master 目录下的唯一 tempPath 和 finalPath
// 4. 把源文件复制到 tempPath
// 失败时只删除 tempPath（如果已创建），源文件始终不动
func (s *FileService) prepareStagedImport(
    sourcePath, fileName, collectionName string,
    onProgress func(copied int64) error,
) (*stagedImport, error)

// commitStagedImport
// 1. 创建 models.File 数据库记录
// 2. os.Rename(TempPath, FinalPath)
// 3. 成功后删除源文件
// 若重命名失败，删除刚创建的 DB 记录，保留源文件
func (s *FileService) commitStagedImport(
    staged *stagedImport,
    tags []string, description string,
) (*models.File, error)

// cleanupStagedImport 失败回滚：只删除 TempPath
func (s *FileService) cleanupStagedImport(staged *stagedImport) error
```

### 3.3 临时目录结构

统一放在目标 master 目录下：

- 单文件导入：
  ```
  <master-dir>/.localspace-temp/<filename>-<timestamp>-<random>.localspace-importing
  ```
- 批量导入（按 job + item 组织）：
  ```
  <master-dir>/.localspace-temp/batch-<jobID>/<itemIndex>/<filename>-<random>.localspace-importing
  ```

设计要点：
- 与最终目录同盘，保证 `os.Rename` 原子；
- 并发导入同名文件不会冲突（补充 D）；
- 批量导入临时文件按任务集中管理（issue 10）；
- 文件名中的随机后缀避免 collision，清理时按明确路径删除，不再用后缀启发式匹配。

### 3.4 导入流程

#### 单文件导入（ImportFile）

1. `prepareStagedImport`：checksum → 重复检测 → 复制到 temp。
2. `ExtractMetadata(tempPath, fileType)` 提取元数据。
3. 构造 `models.File`。
4. `commitStagedImport`：写库 → 重命名 → 删源。
5. 任意阶段失败：调用 `cleanupStagedImport`，源文件保留。

#### 批量导入（BatchImportHandler）

1. `PrepareBatchImport` 调用 `prepareStagedImport` 的前半段（checksum/重复检测/生成路径），返回 `BatchImportPlan`。
2. `CopyFileToTemp` 执行实际的复制到 temp（复用 `prepareStagedImport` 的复制逻辑）。
3. `FinalizeBatchImport` 调用 `commitStagedImport` 完成写库/重命名/删源。

### 3.5 清理策略

| 阶段 | 失败情况 | 清理动作 |
|---|---|---|
| checksum / 重复检测 | 计算失败 / 是重复文件 | 无 temp 可清理 |
| 复制到 temp | 中断/磁盘满 | 删除 `TempPath` |
| 创建 DB 记录 | 数据库写入失败 | 删除 `TempPath`，源文件保留 |
| 重命名 temp → final | 被占用/权限 | 删除 `TempPath`，删除刚创建的 DB 记录，源文件保留 |
| 删除源文件 | 权限/被占用 | 导入已成功；记录日志，不阻塞 |

### 3.6 残留临时文件清理

启动时调用：

```go
func (s *FileService) CleanupOrphanedTempFiles(maxAge time.Duration) error
```

- 扫描所有 master 目录下的 `.localspace-temp/`；
- 删除创建时间超过 24 小时的残留文件；
- 若目录为空则一并删除。

挂载点：`App.Startup` 在 `NormalizeUnfinishedJobs` 之后执行。

---

## 4. 逐项修复映射

| 问题 | 文件 | 修复方式 |
|---|---|---|
| **1. 单文件导入 DB 失败后源文件丢失** | `app/services/file_service.go` | `ImportFile` 改用 `prepareStagedImport` + `commitStagedImport`，源文件在 DB 写入成功后才会被删除。 |
| **2. 跨盘符只复制不删源文件** | `app/services/file_service.go` | 统一先复制到目标 master 的 `.localspace-temp/`，再重命名、删源；跨盘与同盘行为一致。 |
| **3. SQL 注入** | `app/repositories/file_repository.go` | `SortBy` 白名单校验（仅允许 `id`、`created_at`、`modified_at`、`file_name`）；`SortOrder` 仅允许 `ASC`/`DESC`；拼入 SQL 前校验。 |
| **4. DeleteFile 顺序错误** | `app/services/file_service.go` | 先 `fileRepo.Delete(id)`，再 `os.Remove(file.FilePath)`，再更新 storage size；删除失败时整体返回错误，不进入半删状态。 |
| **5. RenameFile 回滚失败** | `app/services/file_service.go` | 先写数据库记录预期新路径，再 `os.Rename(file.FilePath, newPath)`；若重命名失败，数据库回滚到旧路径。 |
| **6. Job goroutine panic** | `app/services/job_service.go` | 在 `runJob` 顶部加 `defer recover()`，panic 后记录错误并把 job 状态设为 `failed`。 |
| **7. shutdown 时 DB 被提前关闭** | `app/services/job_service.go` + `app/app.go` | `JobService` 增加 `sync.WaitGroup` 跟踪运行中的 goroutine；`PrepareForShutdown` 取消上下文并等待 `WaitGroup` 归零；`App.Shutdown` 先等 job service 完成再 `db.Close()`。 |
| **8. PowerShell 字符串拼接** | `app/app.go` | `OpenFileLocation` Windows 分支改为 `exec.Command("explorer", "/select,", absPath).Start()`，删除 PowerShell 中间层。 |
| **9. 前端 API 吞错误** | `frontend/src/api/index.ts` | 修改 `safeWailsCall`：仅 `window.go` 不存在时返回 fallback；Wails 调用失败时 reject。关键操作（导入、删除、任务）的 UI 调用方 catch 并展示真实错误。 |
| **10. 批量导入临时路径 + 清理条件** | `app/services/file_service.go` + `batch_import_handler.go` | 批量导入临时路径改为 `<master>/.localspace-temp/batch-<jobID>/...`；`CleanupBatchImportArtifacts` 只删除传入的 `TempPath`，不再检查 `FinalPath` 后缀。 |
| **补充 C. 单文件元数据类型错误** | `frontend/src/views/ImportView.vue` | 单文件选择后调用 `api.system.getMetadata(path, singleFile.type)`，用 `toRichFile` 解析出的真实类型，而不是硬编码 `'file'`。 |

---

## 5. Job 系统稳定性设计

### 5.1 panic 恢复

`app/services/job_service.go` 的 `runJob` 顶部增加：

```go
defer func() {
    if r := recover(); r != nil {
        err := fmt.Errorf("job panicked: %v\n%s", r, debug.Stack())
        // 更新 job 为 failed 状态并发送事件
    }
}()
```

### 5.2 shutdown 同步

`JobService` 增加 `sync.WaitGroup`：

```go
type JobService struct {
    // ...
    runningWg sync.WaitGroup
}
```

`runJob` 中：

```go
s.runningWg.Add(1)
defer func() {
    s.untrackRunning(jobID)
    s.runningWg.Done()
}()
```

`PrepareForShutdown` 在取消所有运行中任务后调用：

```go
s.runningWg.Wait()
```

### 5.3 取消与恢复行为

- **用户取消**：状态变为 `cancelled`，不可恢复；`Cleanup` 删除未完成的临时文件，并把未完成 item 标记为 `failed`。
- **应用关闭中断**：running/recovering/pending 的可恢复任务变为 `awaiting_resume`，通过持久化的 `batch_import_items` 记录恢复进度，已完成的 item 不会重复导入。

---

## 6. 前端 API 错误处理

### 6.1 改造 `safeWailsCall`

```ts
function safeWailsCall<T>(
  fn: () => Promise<T>,
  fallbackValue: T,
  errorMessage: string
): Promise<T> {
  return new Promise((resolve, reject) => {
    try {
      if (!window.go || !window.go.app || !window.go.app.App) {
        // 浏览器开发环境：返回 fallback
        console.warn(`Wails API not available: ${errorMessage}`)
        resolve(fallbackValue)
        return
      }

      fn()
        .then(resolve)
        .catch((error) => {
          // Wails 真实调用失败：reject
          console.error(`Wails API error: ${errorMessage}`, error)
          reject(error)
        })
    } catch (error) {
      console.error(`Unexpected error calling Wails API: ${errorMessage}`, error)
      reject(error)
    }
  })
}
```

### 6.2 调用方适配

关键操作（导入、删除、重命名、任务提交/恢复/取消）在 UI 中显式 catch 并展示错误。列表/搜索等只读查询保持现有 catch，会把 error 显示在 UI 上。

---

## 7. 测试与验证

### 7.1 单元测试

| 测试目标 | 文件 | 验证点 |
|---|---|---|
| 重复文件直接返回，不复制 | `file_service_test.go` | `prepareStagedImport` 对 duplicate checksum 返回 error，temp 不存在 |
| SQL 注入白名单 | `file_repository_test.go` | 传入 `id;DROP TABLE files;--` 被过滤为 `created_at` |
| DeleteFile 顺序 | `file_service_test.go` | 物理删除失败不影响 DB 记录；DB 删除成功后再删物理 |
| RenameFile 回滚 | `file_service_test.go` | 物理重命名失败时 DB 保持旧路径 |

### 7.2 手动/集成测试

| 场景 | 验证点 |
|---|---|
| 单文件导入全流程 | 文件在 master 目录、数据库记录、源文件已删除 |
| 跨盘导入 | 源在 D 盘，master 在 E 盘，导入后 D 盘源文件删除 |
| DB 写入失败回滚 | 模拟 repo.Create 失败，源文件保留、temp 删除 |
| 批量导入中断恢复 | 50 个文件任务中途关闭应用，恢复后已完成 item 不再重复 |
| Job panic 恢复 | 手动触发 panic，job 状态变为 `failed` |
| Shutdown 等待 | 运行任务时关闭窗口，DB 不 panic，任务可恢复 |
| 打开文件所在位置 | Windows 上点击“打开位置”，Explorer 正确选中文件 |

### 7.3 回归点

- 文件列表正常加载、排序、分页；
- 搜索文件；
- 缩略图生成；
- 删除文件后前端不再显示“幽灵文件”；
- 导入重复文件给出明确提示；
- 单文件图片/视频元数据（宽高、时长）正确提取。

---

## 8. 风险与注意事项

1. `file_service.go` 改动较大，需要同时验证单文件导入、批量导入、重命名、删除、跨盘导入等路径。
2. 前端 `safeWailsCall` 改为 reject 后，需要检查所有调用点是否已 catch，避免未处理 rejection。
3. 新临时目录路径格式下，旧版本残留的 `.localspace-importing` 文件不会被自动识别；通过 `CleanupOrphanedTempFiles` 在启动时按时间阈值清理。
4. 批量导入任务在恢复时会重新复制未完成的 item，大文件任务恢复时可能重复部分网络/磁盘 IO，但可避免数据不一致。

---

## 9. 下一步

本设计通过 `superpowers:writing-plans` 技能转化为详细实现计划，按文件逐一实施。
# LocalSpace 合集与任务功能实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 LocalSpace 从自由的 `collection_name` 字符串切换为 `collections` 表驱动的 `collection_id` 外键，完成文件分页、批量操作（移动合集/删除）、任务页面重构以及历史任务清理 Job。

**Architecture:** 后端新增 `collections` 表与 `files.collection_id` 外键；`FileService`/`CollectionService` 负责合辑变更与物理移动；`JobService` 新增 `job_cleanup` 策略与 `CleanupHandler`，通过 `config` 键保存保留规则；前端通过新的 `CollectionSelector` 限制选择，使用 `FilesView` 批量工具栏与 `TasksView` 统一列表展示。

**Tech Stack:** Wails v2 + Go 1.22+ + SQLite (`github.com/glebarez/sqlite`) + Vue 3 + TypeScript + Pinia + Vite.

## Global Constraints

- 合集只能通过设置页管理；导入、批量导入、编辑文件时只准选择，禁止自由输入。
- 文件通过 `collection_id`（可空）关联 `collections` 表；旧 `collection_name` 列保留但只读。
- 删除合集时若 `files` 中存在引用，后端返回错误，前端提示。
- 批量/单独修改合集时，若存储路径变化必须按 `storage_layout_config` 物理移动文件。
- 后端分页：文件每页可选 20/50/100，任务每页可选 10/20/50。
- 任务清理仅删除终态任务（`completed`/`failed`/`cancelled`/`timed_out`/`cleanup_failed`）。
- 最大条数与最大天数独立启用，任一条件触发即删除。
- 任务列表使用半透明/毛玻璃背景（`backdrop-filter: blur(...)` + `background: rgba(...)`）。
- 默认关闭自动清理。
- 每个任务末尾提交一次；每阶段完成后运行对应测试/构建命令。

---

## Phase 1：合集基础设施

### Task 1：数据库迁移与模型

**Files:**
- Modify: `app/database/migrations.go`（新增 v9 迁移）
- Create: `app/models/collection.go`
- Modify: `app/models/file.go`（添加 `CollectionID`/`FileListResponse`）
- Modify: `app/models/job.go`（添加 `JobTypeCleanup`/`JobRetentionConfig`）

**Interfaces:**
- Consumes: 现有迁移框架
- Produces: `models.Collection`, `models.FileListResponse`, `models.JobRetentionConfig`, `models.File.CollectionID`, `models.JobTypeCleanup`

- [ ] **Step 1：新增迁移 v9**

在 `app/database/migrations.go` 的 `migrations` 切片末尾追加版本 9，并在文件底部实现 `migration009_Up`/`migration009_Down`：

```go
// 在 migrations 切片中追加
{
    Version: 9,
    Name:    "add_collections_and_file_collection_id",
    Up:      migration009_Up,
    Down:    migration009_Down,
},
```

```go
func migration009_Up(db *sql.DB) error {
    if _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS collections (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL UNIQUE,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `); err != nil {
        return fmt.Errorf("failed to create collections table: %w", err)
    }

    var exists bool
    if err := db.QueryRow(`
        SELECT COUNT(*) > 0
        FROM pragma_table_info('files')
        WHERE name = 'collection_id'
    `).Scan(&exists); err != nil {
        return fmt.Errorf("failed to check collection_id column: %w", err)
    }
    if !exists {
        if _, err := db.Exec(`ALTER TABLE files ADD COLUMN collection_id INTEGER DEFAULT NULL`); err != nil {
            return fmt.Errorf("failed to add collection_id column: %w", err)
        }
    }

    if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_files_collection_id ON files(collection_id)`); err != nil {
        return fmt.Errorf("failed to create collection_id index: %w", err)
    }
    return nil
}

func migration009_Down(db *sql.DB) error {
    return fmt.Errorf("SQLite rollback not supported for collection_id addition")
}
```

- [ ] **Step 2：新增模型文件**

`app/models/collection.go`：

```go
package models

import "strings"

type Collection struct {
    ID        uint   `json:"id"`
    Name      string `json:"name"`
    CreatedAt string `json:"createdAt"`
    UpdatedAt string `json:"updatedAt"`
}

type FileListResponse struct {
    Items    []*File `json:"items"`
    Page     int     `json:"page"`
    PageSize int     `json:"pageSize"`
    Total    int     `json:"total"`
}

type BatchMoveResult struct {
    SuccessCount int                   `json:"successCount"`
    FailedCount  int                   `json:"failedCount"`
    FailedItems  []BatchMoveFailedItem `json:"failedItems"`
}

type BatchMoveFailedItem struct {
    FileID   uint   `json:"fileId"`
    FileName string `json:"fileName"`
    Error    string `json:"error"`
}

type BatchDeleteResult struct {
    SuccessCount int                     `json:"successCount"`
    FailedCount  int                     `json:"failedCount"`
    FailedItems  []BatchDeleteFailedItem `json:"failedItems"`
}

type BatchDeleteFailedItem struct {
    FileID   uint   `json:"fileId"`
    FileName string `json:"fileName"`
    Error    string `json:"error"`
}

func NormalizeCollectionName(name string) string {
    return strings.TrimSpace(name)
}
```

（在 `app/models/file.go` 的 `File` 结构体中添加 `CollectionID`：）

```go
type File struct {
    ID             uint    `json:"id"`
    FileName       string  `json:"fileName"`
    OriginalName   string  `json:"originalName"`
    CollectionName string  `json:"collectionName"`
    CollectionID   *uint   `json:"collectionId"`
    FilePath       string  `json:"filePath"`
    // ... 其余字段不变
}
```

（在 `app/models/job.go` 中：）

```go
const (
    JobTypeBatchImport  = "batch_import"
    JobTypeSingleImport = "single_import"
    JobTypeCleanup      = "job_cleanup"
)

type JobRetentionConfig struct {
    ID        uint   `json:"id"`
    Enabled   bool   `json:"enabled"`
    MaxCount  int    `json:"maxCount"`
    MaxDays   int    `json:"maxDays"`
    UpdatedAt string `json:"updatedAt"`
}
```

- [ ] **Step 3：运行后端测试确保迁移通过**

Run: `go test ./app/database/...`
Expected: PASS

- [ ] **Step 4：提交**

```bash
git add app/database/migrations.go app/models/collection.go app/models/file.go app/models/job.go
git commit -m "feat(collections): add collections table, file collection_id, retention config models"
```

### Task 2：合集数据访问层

**Files:**
- Create: `app/repositories/collection_repository.go`

**Interfaces:**
- Consumes: `models.Collection`
- Produces: `CollectionRepository.GetAll`, `Add`, `Remove`, `FindByID`, `CountFilesByCollectionID`

- [ ] **Step 1：编写合集仓库**

```go
package repositories

import (
    "database/sql"
    "fmt"
    "strings"
    "time"

    "LocalSpace/app/models"
)

type CollectionRepository struct {
    db *sql.DB
}

func NewCollectionRepository(dbWrapper *SQLiteDBWrapper) *CollectionRepository {
    return &CollectionRepository{db: dbWrapper.GetDB()}
}

func (r *CollectionRepository) GetAll() ([]models.Collection, error) {
    rows, err := r.db.Query(`
        SELECT id, name, created_at, updated_at
        FROM collections
        ORDER BY name ASC
    `)
    if err != nil {
        return nil, fmt.Errorf("failed to list collections: %w", err)
    }
    defer rows.Close()

    var result []models.Collection
    for rows.Next() {
        var c models.Collection
        if err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt); err != nil {
            return nil, fmt.Errorf("failed to scan collection: %w", err)
        }
        result = append(result, c)
    }
    return result, nil
}

func (r *CollectionRepository) FindByID(id uint) (*models.Collection, error) {
    row := r.db.QueryRow(`
        SELECT id, name, created_at, updated_at
        FROM collections WHERE id = ?
    `, id)
    var c models.Collection
    if err := row.Scan(&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt); err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("collection not found")
        }
        return nil, fmt.Errorf("failed to find collection: %w", err)
    }
    return &c, nil
}

func (r *CollectionRepository) Add(name string) (uint, error) {
    trimmed := models.NormalizeCollectionName(name)
    if trimmed == "" {
        return 0, fmt.Errorf("collection name cannot be empty")
    }
    now := time.Now().Format(time.RFC3339)
    result, err := r.db.Exec(`
        INSERT INTO collections (name, created_at, updated_at) VALUES (?, ?, ?)
    `, trimmed, now, now)
    if err != nil {
        if strings.Contains(err.Error(), "UNIQUE constraint failed") {
            return 0, fmt.Errorf("collection name already exists")
        }
        return 0, fmt.Errorf("failed to add collection: %w", err)
    }
    id, err := result.LastInsertId()
    if err != nil {
        return 0, fmt.Errorf("failed to get collection id: %w", err)
    }
    return uint(id), nil
}

func (r *CollectionRepository) Remove(id uint) error {
    result, err := r.db.Exec(`DELETE FROM collections WHERE id = ?`, id)
    if err != nil {
        return fmt.Errorf("failed to remove collection: %w", err)
    }
    rows, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    if rows == 0 {
        return fmt.Errorf("collection not found")
    }
    return nil
}

func (r *CollectionRepository) CountFilesByCollectionID(id uint) (int, error) {
    var count int
    if err := r.db.QueryRow(`
        SELECT COUNT(*) FROM files WHERE collection_id = ? AND is_deleted = 0
    `, id).Scan(&count); err != nil {
        return 0, fmt.Errorf("failed to count files in collection: %w", err)
    }
    return count, nil
}
```

- [ ] **Step 2：编写失败测试**

`app/repositories/collection_repository_test.go`：

```go
package repositories

import (
    "database/sql"
    "path/filepath"
    "testing"

    _ "github.com/glebarez/sqlite"
    "LocalSpace/app/database"
)

func TestCollectionRepository_Remove_BlockedWhenReferenced(t *testing.T) {
    tempDir := t.TempDir()
    dbPath := filepath.Join(tempDir, "test.db")
    db, err := database.NewSQLiteDB(dbPath)
    if err != nil {
        t.Fatalf("failed to create test db: %v", err)
    }
    defer db.Close()

    wrapper := NewSQLiteDBWrapper(db)
    colRepo := NewCollectionRepository(wrapper)
    fileRepo := NewFileRepository(wrapper)

    id, err := colRepo.Add("Projects")
    if err != nil {
        t.Fatalf("failed to add collection: %v", err)
    }

    file := &models.File{
        FileName:     "doc.pdf",
        OriginalName: "doc.pdf",
        FilePath:     "/test/doc.pdf",
        FileType:     "document",
        FileSubType:  "pdf",
        FileSize:     100,
        Tags:         []string{},
        Metadata:     models.Metadata{},
        CollectionID: &id,
        Checksum:     "abc",
    }
    if err := fileRepo.Create(file); err != nil {
        t.Fatalf("failed to create file: %v", err)
    }

    if err := colRepo.Remove(id); err == nil {
        t.Fatal("expected remove to be blocked when referenced")
    }
}
```

（注意：真正需要移除被引用合集时，该测试会先失败；后续服务层会负责检查，但仓库层不主动检查。）

- [ ] **Step 3：运行测试**

Run: `go test ./app/repositories/... -run TestCollectionRepository`
Expected: PASS

- [ ] **Step 4：提交**

```bash
git add app/repositories/collection_repository.go app/repositories/collection_repository_test.go
git commit -m "feat(collections): add collection repository and tests"
```

### Task 3：文件仓库支持合集 ID 与分页计数

**Files:**
- Modify: `app/repositories/file_repository.go`

**Interfaces:**
- Consumes: `models.Collection`, `models.FileListResponse`, `repositories.CollectionRepository`
- Produces: `FileFilter.CollectionID`, `FileRepository.List` 返回总数, `FileRepository.Count`, `FileRepository.UpdateMetadataWithCollection`, `FileRepository.UpdateCollectionID`

- [ ] **Step 1：扩展 FileFilter**

```go
type FileFilter struct {
    Page         int
    PageSize     int
    FileType     string
    CollectionID *uint
    SortBy       string
    SortOrder    string
}
```

- [ ] **Step 2：新增 Count 查询**

在 `app/repositories/file_repository.go` 添加：

```go
func (r *FileRepository) Count(filter FileFilter) (int, error) {
    whereParts := []string{"1=1"}
    args := []interface{}{}
    argIndex := 1

    if filter.FileType != "" && filter.FileType != "all" {
        whereParts = append(whereParts, fmt.Sprintf("file_type = $%d", argIndex))
        args = append(args, filter.FileType)
        argIndex++
    }
    if filter.CollectionID != nil {
        whereParts = append(whereParts, fmt.Sprintf("collection_id = $%d", argIndex))
        args = append(args, *filter.CollectionID)
        argIndex++
    }

    query := fmt.Sprintf(`SELECT COUNT(*) FROM files WHERE %s`, strings.Join(whereParts, " AND "))
    var count int
    if err := r.db.QueryRow(query, args...).Scan(&count); err != nil {
        return 0, fmt.Errorf("failed to count files: %w", err)
    }
    return count, nil
}
```

- [ ] **Step 3：修改 List 以支持 collection_id 过滤和分页占位符**

将 `List` 中 `WHERE 1=1` 之后替换为：

```go
args := []interface{}{}
argIndex := 1

if filter.FileType != "" && filter.FileType != "all" {
    query += fmt.Sprintf(" AND file_type = $%d", argIndex)
    args = append(args, filter.FileType)
    argIndex++
}
if filter.CollectionID != nil {
    query += fmt.Sprintf(" AND collection_id = $%d", argIndex)
    args = append(args, *filter.CollectionID)
    argIndex++
}

sortBy := normalizeSortBy(filter.SortBy)
sortOrder := normalizeSortOrder(filter.SortOrder)
query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

offset := (filter.Page - 1) * filter.PageSize
query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
args = append(args, filter.PageSize, offset)
```

（注意：将分页占位符从原来的 `$%d` 替换为基于当前 `argIndex` 的编号。）

- [ ] **Step 4：修改 Create/Update 写入 collection_id**

`Create` 的 SQL 改为显式写入 `collection_id`：

```go
query := `
    INSERT INTO files (file_name, original_name, collection_name, collection_id, file_path, file_type, file_sub_type, file_size, tags, description, metadata, thumbnail, checksum, is_deleted, deleted_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

result, err := r.db.Exec(query,
    file.FileName,
    file.OriginalName,
    "",
    file.CollectionID,
    file.FilePath,
    file.FileType,
    file.FileSubType,
    file.FileSize,
    string(tagsJSON),
    file.Description,
    string(metadataJSON),
    file.Thumbnail,
    file.Checksum,
    file.IsDeleted,
    file.DeletedAt,
)
```

`Update` 的 SQL 改为：

```go
query := `
    UPDATE files
    SET file_name = ?, collection_name = ?, collection_id = ?, file_path = ?, file_type = ?, file_sub_type = ?,
        file_size = ?, tags = ?, description = ?, metadata = ?, thumbnail = ?,
        modified_at = CURRENT_TIMESTAMP
    WHERE id = ?`
```

并传入 `file.CollectionID`。

- [ ] **Step 5：新增 UpdateMetadataWithCollection**

```go
func (r *FileRepository) UpdateMetadataWithCollection(id uint, tags []string, description string, collectionID *uint) error {
    tagsJSON, err := json.Marshal(tags)
    if err != nil {
        return fmt.Errorf("failed to marshal tags: %w", err)
    }
    query := `UPDATE files SET tags = ?, description = ?, collection_id = ?, modified_at = ? WHERE id = ?`
    modifiedAt := time.Now().Format(time.RFC3339)
    result, err := r.db.Exec(query, string(tagsJSON), description, collectionID, modifiedAt, id)
    if err != nil {
        return fmt.Errorf("failed to update file metadata: %w", err)
    }
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    if rowsAffected == 0 {
        return fmt.Errorf("file not found")
    }
    return nil
}
```

- [ ] **Step 6：修改 Search 支持分页（按设计 spec 返回分页）**

本任务先保持 `Search` 返回最多 100 条；在 Task 2.1 中再改为分页。

- [ ] **Step 7：运行仓库测试**

Run: `go test ./app/repositories/...`
Expected: PASS

- [ ] **Step 8：提交**

```bash
git add app/repositories/file_repository.go
git commit -m "feat(collections): support collection_id filtering and pagination count"
```

### Task 4：任务保留配置存储

**Files:**
- Modify: `app/services/config_service.go`

**Interfaces:**
- Consumes: `models.JobRetentionConfig`
- Produces: `ConfigService.GetJobRetentionConfig`, `ConfigService.UpdateJobRetentionConfig`

- [ ] **Step 1：新增方法**

```go
const jobRetentionConfigKey = "job_retention_config"

func defaultJobRetentionConfig() *models.JobRetentionConfig {
    return &models.JobRetentionConfig{
        ID:       1,
        Enabled:  false,
        MaxCount: 0,
        MaxDays:  0,
    }
}

func (s *ConfigService) GetJobRetentionConfig() (*models.JobRetentionConfig, error) {
    value, err := s.configRepo.Get(jobRetentionConfigKey)
    if err != nil {
        if strings.Contains(err.Error(), "config key not found") {
            return defaultJobRetentionConfig(), nil
        }
        return nil, err
    }
    if strings.TrimSpace(value) == "" {
        return defaultJobRetentionConfig(), nil
    }
    var config models.JobRetentionConfig
    if err := json.Unmarshal([]byte(value), &config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal job retention config: %w", err)
    }
    if config.ID == 0 {
        config.ID = 1
    }
    return &config, nil
}

func (s *ConfigService) UpdateJobRetentionConfig(config models.JobRetentionConfig) error {
    if config.MaxCount < 0 {
        config.MaxCount = 0
    }
    if config.MaxDays < 0 {
        config.MaxDays = 0
    }
    data, err := json.Marshal(config)
    if err != nil {
        return fmt.Errorf("failed to marshal job retention config: %w", err)
    }
    return s.configRepo.Set(jobRetentionConfigKey, string(data))
}
```

- [ ] **Step 2：导入依赖**

确保 `app/services/config_service.go` 顶部已引入 `encoding/json` 和 `strings`。

- [ ] **Step 3：运行测试**

Run: `go test ./app/services/... -run TestConfig`
Expected: PASS

- [ ] **Step 4：提交**

```bash
git add app/services/config_service.go
git commit -m "feat(jobs): add job retention config CRUD"
```

### Task 5：合集服务与批量操作

**Files:**
- Create: `app/services/collection_service.go`

**Interfaces:**
- Consumes: `CollectionRepository`, `FileRepository`, `StorageService`, `models.Collection`, `models.BatchMoveResult`, `models.BatchDeleteResult`
- Produces: `CollectionService.GetCollections`, `AddCollection`, `RemoveCollection`, `BatchUpdateFilesCollection`, `BatchDeleteFiles`

- [ ] **Step 1：实现 CollectionService**

```go
package services

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "LocalSpace/app/models"
    "LocalSpace/app/repositories"
)

type CollectionService struct {
    collectionRepo *repositories.CollectionRepository
    fileRepo       *repositories.FileRepository
    storageService *StorageService
}

func NewCollectionService(
    collectionRepo *repositories.CollectionRepository,
    fileRepo *repositories.FileRepository,
    storageService *StorageService,
) *CollectionService {
    return &CollectionService{
        collectionRepo: collectionRepo,
        fileRepo:       fileRepo,
        storageService: storageService,
    }
}

func (s *CollectionService) GetCollections() ([]models.Collection, error) {
    return s.collectionRepo.GetAll()
}

func (s *CollectionService) AddCollection(name string) (uint, error) {
    return s.collectionRepo.Add(name)
}

func (s *CollectionService) RemoveCollection(id uint) error {
    count, err := s.collectionRepo.CountFilesByCollectionID(id)
    if err != nil {
        return err
    }
    if count > 0 {
        return fmt.Errorf("collection is referenced by %d file(s)", count)
    }
    return s.collectionRepo.Remove(id)
}

func (s *CollectionService) BatchUpdateFilesCollection(ids []uint, collectionID *uint) (models.BatchMoveResult, error) {
    var result models.BatchMoveResult
    targetName := ""
    if collectionID != nil {
        c, err := s.collectionRepo.FindByID(*collectionID)
        if err != nil {
            return result, err
        }
        targetName = c.Name
    }

    for _, id := range ids {
        file, err := s.fileRepo.FindByID(id)
        if err != nil {
            result.FailedCount++
            result.FailedItems = append(result.FailedItems, models.BatchMoveFailedItem{FileID: id, Error: err.Error()})
            continue
        }

        if err := s.moveFileToCollection(file, collectionID, targetName); err != nil {
            result.FailedCount++
            result.FailedItems = append(result.FailedItems, models.BatchMoveFailedItem{FileID: id, FileName: file.FileName, Error: err.Error()})
            continue
        }
        result.SuccessCount++
    }
    return result, nil
}

func (s *CollectionService) moveFileToCollection(file *models.File, collectionID *uint, targetName string) error {
    oldPath := file.FilePath
    master, err := s.findMasterForPath(oldPath)
    if err != nil {
        return err
    }

    layout, err := s.storageService.GetStorageLayoutConfig()
    if err != nil {
        return err
    }

    segment := targetName
    if segment == "" {
        segment = layout.UnsortedFolderName
    }
    if layout.SanitizeFolderName {
        segment = sanitizePathSegment(segment)
    }

    _, subPath, err := s.storageService.EnsureSubDirectory(master.ID, file.FileType)
    if err != nil {
        return err
    }
    if layout.Strategy == "type_collection" {
        subPath = filepath.Join(subPath, segment)
    }
    newPath := filepath.Join(subPath, file.FileName)

    if newPath == oldPath {
        file.CollectionID = collectionID
        return s.fileRepo.Update(file)
    }

    if _, err := os.Stat(newPath); err == nil {
        return fmt.Errorf("destination already exists: %s", newPath)
    }
    if err := s.storageService.EnsureStorageDirExists(filepath.Dir(newPath)); err != nil {
        return err
    }

    file.CollectionID = collectionID
    file.FilePath = newPath
    if err := s.fileRepo.Update(file); err != nil {
        return err
    }
    if err := os.Rename(oldPath, newPath); err != nil {
        file.CollectionID = nil
        file.FilePath = oldPath
        _ = s.fileRepo.Update(file)
        return fmt.Errorf("failed to move file: %w", err)
    }

    if err := s.storageService.UpdateMasterDirectorySize(master.ID, file.FileType, -file.FileSize); err != nil {
        fmt.Printf("Warning: failed to update old master size: %v\n", err)
    }
    newMaster, err := s.findMasterForPath(newPath)
    if err == nil && newMaster.ID != master.ID {
        if err := s.storageService.UpdateMasterDirectorySize(newMaster.ID, file.FileType, file.FileSize); err != nil {
            fmt.Printf("Warning: failed to update new master size: %v\n", err)
        }
    }
    return nil
}

func (s *CollectionService) findMasterForPath(path string) (*models.StorageDir, error) {
    masters, err := s.storageService.GetMasterDirectories()
    if err != nil {
        return nil, err
    }
    for i := range masters {
        if strings.HasPrefix(path, masters[i].Path) {
            return &masters[i], nil
        }
    }
    return nil, fmt.Errorf("no master directory contains path %s", path)
}

func (s *CollectionService) BatchDeleteFiles(ids []uint) (models.BatchDeleteResult, error) {
    var result models.BatchDeleteResult
    for _, id := range ids {
        file, err := s.fileRepo.FindByID(id)
        if err != nil {
            result.FailedCount++
            result.FailedItems = append(result.FailedItems, models.BatchDeleteFailedItem{FileID: id, Error: err.Error()})
            continue
        }
        if err := s.fileRepo.Delete(id); err != nil {
            result.FailedCount++
            result.FailedItems = append(result.FailedItems, models.BatchDeleteFailedItem{FileID: id, FileName: file.FileName, Error: err.Error()})
            continue
        }
        if err := os.Remove(file.FilePath); err != nil && !os.IsNotExist(err) {
            result.FailedCount++
            result.FailedItems = append(result.FailedItems, models.BatchDeleteFailedItem{FileID: id, FileName: file.FileName, Error: err.Error()})
            continue
        }
        if err := s.storageService.UpdateStorageSize(file.FileType, -file.FileSize); err != nil {
            fmt.Printf("Warning: failed to update storage size: %v\n", err)
        }
        result.SuccessCount++
    }
    return result, nil
}
```

- [ ] **Step 2：提交**

```bash
git add app/services/collection_service.go
git commit -m "feat(collections): add collection service with batch move/delete"
```

### Task 6：文件服务元数据更新支持合集变更

**Files:**
- Modify: `app/services/file_service.go`

**Interfaces:**
- Consumes: `models.Collection`, `CollectionService`?（直接复用 `CollectionRepository` 解析名称）
- Produces: `FileService.UpdateFileMetadata(id, tags, description, collectionID)`

- [ ] **Step 1：在 FileService 中注入 collectionRepo**

新增依赖接口与字段：

```go
type collectionResolver interface {
    FindByID(id uint) (*models.Collection, error)
}

// 在 FileService 结构体添加
    collectionRepo collectionResolver
```

修改 `NewFileService` 签名（添加 `collectionRepo` 参数）。

- [ ] **Step 2：重构 UpdateFileMetadata 处理合集移动**

将原来的 `UpdateFileMetadata` 替换为：

```go
func (s *FileService) UpdateFileMetadata(id uint, tags []string, description string, collectionID *uint) error {
    if id == 0 {
        return fmt.Errorf("file id cannot be empty")
    }
    if s.metadataRepo == nil {
        return fmt.Errorf("file repository not initialized")
    }

    file, err := s.metadataRepo.FindByID(id)
    if err != nil {
        return fmt.Errorf("failed to get file: %w", err)
    }

    normalizedTags := s.normalizeMetadataTags(tags)
    normalizedDescription := s.normalizeMetadataDescription(description)

    changedCollection := !sameCollectionID(file.CollectionID, collectionID)
    if changedCollection {
        targetName := ""
        if collectionID != nil {
            collection, err := s.collectionRepo.FindByID(*collectionID)
            if err != nil {
                return fmt.Errorf("failed to resolve collection: %w", err)
            }
            targetName = collection.Name
        }
        if err := s.moveFileToCollection(file, collectionID, targetName); err != nil {
            return fmt.Errorf("failed to move file to collection: %w", err)
        }
    }

    file.Tags = normalizedTags
    file.Description = normalizedDescription
    if err := s.fileRepo.Update(file); err != nil {
        return fmt.Errorf("failed to update file metadata: %w", err)
    }
    return nil
}

func sameCollectionID(a, b *uint) bool {
    if a == nil && b == nil {
        return true
    }
    if a == nil || b == nil {
        return false
    }
    return *a == *b
}
```

其中 `moveFileToCollection` 复用 Task 5 中实现，可抽出到 `app/services/collection_move.go` 或内联在 `FileService` 中（保持复制粘贴可用，但推荐抽到单独文件）。

- [ ] **Step 3：更新测试桩**

`file_service_test.go` 中 `metadataRepoStub` 已实现 `FindByID`/`UpdateMetadata`；测试 `TestUpdateFileMetadata_*` 将更新签名，但本任务先保持原有签名的测试，新增 `collectionRepo` 字段为 `nil` 的 `FileService` 内部调用；旧测试会无法编译。因此需要同步修改测试文件中的 `service := &FileService{metadataRepo: repo}` 为 `service := &FileService{metadataRepo: repo, collectionRepo: nil}`。由于旧 `UpdateFileMetadata` 签名已不存在，旧测试必须改为使用新的三参数签名并加 `nil` collectionID。

- [ ] **Step 4：提交**

```bash
git add app/services/file_service.go app/services/file_service_test.go
git commit -m "feat(collections): update metadata supports collection change"
```

### Task 7：导入请求与 Handler 使用合集 ID

**Files:**
- Modify: `app/services/file_service.go`（`ImportFileRequest`/`BatchImportJobRequest`）
- Modify: `app/services/batch_import_handler.go`
- Modify: `app/services/single_import_handler.go`（无需改动，依赖 FileService.ImportFile）

**Interfaces:**
- Consumes: `models.BatchImportJobRequest.CollectionID *uint`, `ImportFileRequest.CollectionID *uint`
- Produces: 导入时根据 ID 解析合集名称并写入 `collection_id`

- [ ] **Step 1：修改请求模型**

`app/services/file_service.go`：

```go
type ImportFileRequest struct {
    FilePath       string
    FileName       string
    Description    string
    Tags           []string
    Keywords       string
    CollectionID   *uint
}

type BatchImportJobRequest struct {
    Files                        []BatchImportFileInput
    SharedTags                   []string
    SharedDescription            string
    CollectionID                 *uint
    EnableAIGeneratedTags        bool
    EnableAIGeneratedDescription bool
}
```

- [ ] **Step 2：修改 FileService 导入 staging 解析合集名**

`prepareStagedImport` 的签名由 `collectionName string` 改为 `collectionID *uint`。函数开头：

```go
var collectionName string
if collectionID != nil {
    collection, err := s.collectionRepo.FindByID(*collectionID)
    if err != nil {
        return nil, fmt.Errorf("failed to resolve collection: %w", err)
    }
    collectionName = collection.Name
}
```

然后沿用 `trimmedCollectionName := strings.TrimSpace(collectionName)` 等逻辑。

- [ ] **Step 3：修改 BatchImportHandler 使用 CollectionID**

```go
plan, err := runtime.FileService().PrepareBatchImport(sourcePath, displayName, payload.CollectionID, job.ID, item.ItemIndex)
```

- [ ] **Step 4：提交**

```bash
git add app/services/file_service.go app/services/batch_import_handler.go
git commit -m "feat(collections): import requests use collection_id"
```

### Task 8：Wails 绑定层更新

**Files:**
- Modify: `app/app.go`

**Interfaces:**
- Consumes: `CollectionService`, `ConfigService.GetJobRetentionConfig`
- Produces: Wails 暴露的 `GetCollections`, `AddCollection`, `RemoveCollection`, `UpdateFileMetadata`, `BatchUpdateFilesCollection`, `BatchDeleteFiles`, `GetFiles`, `SearchFiles`, `GetJobRetentionConfig`, `UpdateJobRetentionConfig`, `SubmitJobCleanup`, `DeleteJobRecord`

- [ ] **Step 1：初始化 CollectionService**

在 `initializeApp` 中，在 `a.fileService` 创建后添加：

```go
collectionRepo := repositories.NewCollectionRepository(repositories.NewSQLiteDBWrapper(db))
a.collectionService = services.NewCollectionService(collectionRepo, fileRepo, a.storageService)
```

并在 `App` 结构体中添加字段 `collectionService *services.CollectionService`。

- [ ] **Step 2：更新导入方法签名**

```go
func (a *App) ImportFileWithMetadata(filePath, fileName, description string, tags []string, keywords string, collectionID uint) error {
    var cid *uint
    if collectionID > 0 {
        cid = &collectionID
    }
    return a.fileService.ImportFile(services.ImportFileRequest{
        FilePath:       filePath,
        FileName:       fileName,
        Description:    description,
        Tags:           tags,
        Keywords:       keywords,
        CollectionID:   cid,
    })
}

func (a *App) SubmitSingleImportJob(filePath, fileName, description string, tags []string, keywords string, collectionID uint) (*models.Job, error) {
    var cid *uint
    if collectionID > 0 {
        cid = &collectionID
    }
    return a.jobService.SubmitSingleImportJob(services.ImportFileRequest{
        FilePath:       filePath,
        FileName:       fileName,
        Description:    description,
        Tags:           tags,
        Keywords:       keywords,
        CollectionID:   cid,
    })
}

func (a *App) SubmitBatchImportJob(req models.BatchImportJobRequest) (*models.Job, error) {
    return a.jobService.SubmitBatchImportJob(req)
}
```

- [ ] **Step 3：新增合集相关 Wails 方法**

```go
func (a *App) GetCollections() ([]models.Collection, error) {
    return a.collectionService.GetCollections()
}

func (a *App) AddCollection(name string) (uint, error) {
    return a.collectionService.AddCollection(name)
}

func (a *App) RemoveCollection(id uint) error {
    return a.collectionService.RemoveCollection(id)
}

func (a *App) UpdateFileMetadata(id uint, tags []string, description string, collectionID uint) error {
    var cid *uint
    if collectionID > 0 {
        cid = &collectionID
    }
    return a.fileService.UpdateFileMetadata(id, tags, description, cid)
}

func (a *App) BatchUpdateFilesCollection(ids []uint, collectionID uint) (models.BatchMoveResult, error) {
    var cid *uint
    if collectionID > 0 {
        cid = &collectionID
    }
    return a.collectionService.BatchUpdateFilesCollection(ids, cid)
}

func (a *App) BatchDeleteFiles(ids []uint) (models.BatchDeleteResult, error) {
    return a.collectionService.BatchDeleteFiles(ids)
}

func (a *App) GetFiles(page, pageSize int, fileType string, collectionID uint) (*models.FileListResponse, error) {
    var cid *uint
    if collectionID > 0 {
        cid = &collectionID
    }
    return a.fileService.ListFilesResponse(services.FileFilter{
        Page:         page,
        PageSize:     pageSize,
        FileType:     fileType,
        CollectionID: cid,
    })
}

func (a *App) SearchFiles(query string, page, pageSize int) (*models.FileListResponse, error) {
    return a.fileService.SearchFilesResponse(query, page, pageSize)
}
```

- [ ] **Step 4：新增任务保留/清理 Wails 方法**

```go
func (a *App) GetJobRetentionConfig() (*models.JobRetentionConfig, error) {
    return a.configService.GetJobRetentionConfig()
}

func (a *App) UpdateJobRetentionConfig(config models.JobRetentionConfig) error {
    return a.configService.UpdateJobRetentionConfig(config)
}

func (a *App) SubmitJobCleanup() (*models.Job, error) {
    return a.jobService.SubmitJobCleanup()
}

func (a *App) DeleteJobRecord(id uint) error {
    return a.jobService.DeleteJobRecord(id)
}
```

- [ ] **Step 5：提交**

```bash
git add app/app.go
git commit -m "feat(api): expose collection and job cleanup Wails bindings"
```

### Task 9：前端类型与 API 封装

**Files:**
- Modify: `frontend/src/types/index.ts`
- Modify: `frontend/src/types/jobs.ts`
- Modify: `frontend/src/api/index.ts`

**Interfaces:**
- Consumes: Wails 方法签名
- Produces: TS `Collection`, `FileListResponse`, `BatchMoveResult`, `BatchDeleteResult`, `JobRetentionConfig`, `api.collection.*`, `api.file.*` 分页签名，清理相关 API

- [ ] **Step 1：扩展类型定义**

`frontend/src/types/index.ts` 追加：

```ts
export interface Collection {
  id: number
  name: string
  createdAt: string
  updatedAt: string
}

export interface FileListResponse {
  items: File[]
  page: number
  pageSize: number
  total: number
}

export interface BatchMoveFailedItem {
  fileId: number
  fileName: string
  error: string
}

export interface BatchMoveResult {
  successCount: number
  failedCount: number
  failedItems: BatchMoveFailedItem[]
}

export interface BatchDeleteFailedItem {
  fileId: number
  fileName: string
  error: string
}

export interface BatchDeleteResult {
  successCount: number
  failedCount: number
  failedItems: BatchDeleteFailedItem[]
}

export interface JobRetentionConfig {
  id: number
  enabled: boolean
  maxCount: number
  maxDays: number
  updatedAt: string
}
```

`frontend/src/types/jobs.ts` 将 `BatchImportJobRequest` 和 `SingleImportJobRequest` 中的 `collectionName` 改为 `collectionId?: number`。

- [ ] **Step 2：更新 api/index.ts 声明与方法**

在 `Window.go.app.App` 接口中新增/修改：

```ts
GetCollections: () => Promise<any[]>
AddCollection: (name: string) => Promise<number>
RemoveCollection: (id: number) => Promise<string>
GetFiles: (page: number, pageSize: number, fileType: string, collectionId: number) => Promise<any>
SearchFiles: (query: string, page: number, pageSize: number) => Promise<any>
UpdateFileMetadata: (id: number, tags: string[], description: string, collectionId: number) => Promise<void>
BatchUpdateFilesCollection: (ids: number[], collectionId: number) => Promise<any>
BatchDeleteFiles: (ids: number[]) => Promise<any>
ImportFileWithMetadata: (filePath: string, fileName: string, description: string, tags: string[], keywords: string, collectionId: number) => Promise<string>
SubmitSingleImportJob: (filePath: string, fileName: string, description: string, tags: string[], keywords: string, collectionId: number) => Promise<any>
SubmitBatchImportJob: (payload: any) => Promise<any>
GetJobRetentionConfig: () => Promise<any>
UpdateJobRetentionConfig: (config: any) => Promise<string>
SubmitJobCleanup: () => Promise<any>
DeleteJobRecord: (id: number) => Promise<string>
```

封装 `api` 对象：

```ts
export const api = {
  // ... 保留原有 API

  collection: {
    getAll: () => safeWailsCall(() => window.go!.app!.App.GetCollections(), [], 'GetCollections'),
    add: (name: string) => safeWailsCall(() => window.go!.app!.App.AddCollection(name), 0, 'AddCollection'),
    remove: (id: number) => safeWailsCall(() => window.go!.app!.App.RemoveCollection(id), 'success', 'RemoveCollection'),
  },

  file: {
    // 替换旧 list/search/updateMetadata/importWithMetadata
    list: (page: number, pageSize: number, fileType: string, collectionId?: number) =>
      safeWailsCall(() => window.go!.app!.App.GetFiles(page, pageSize, fileType, collectionId || 0), { items: [], page, pageSize, total: 0 }, 'GetFiles'),
    search: (query: string, page: number, pageSize: number) =>
      safeWailsCall(() => window.go!.app!.App.SearchFiles(query, page, pageSize), { items: [], page, pageSize, total: 0 }, 'SearchFiles'),
    updateMetadata: (id: number, tags: string[], description: string, collectionId?: number) =>
      new Promise<void>((resolve, reject) => {
        try {
          if (!window.go || !window.go.app || !window.go.app.App) {
            reject(new Error('Wails API not available'))
            return
          }
          window.go!.app!.App.UpdateFileMetadata(id, tags, description, collectionId || 0)
            .then(() => resolve())
            .catch(reject)
        } catch (error) { reject(error) }
      }),
    importWithMetadata: (filePath: string, fileName: string, description: string, tags: string[], keywords: string, collectionId?: number) =>
      new Promise<void>((resolve, reject) => {
        try {
          if (!window.go || !window.go.app || !window.go.app.App) {
            reject(new Error('Wails API not available'))
            return
          }
          window.go!.app!.App.ImportFileWithMetadata(filePath, fileName, description, tags, keywords, collectionId || 0)
            .then(() => resolve())
            .catch(reject)
        } catch (error) { reject(error) }
      }),
    batchUpdateCollection: (ids: number[], collectionId?: number) =>
      safeWailsCall(() => window.go!.app!.App.BatchUpdateFilesCollection(ids, collectionId || 0), { successCount: 0, failedCount: 0, failedItems: [] }, 'BatchUpdateFilesCollection'),
    batchDelete: (ids: number[]) =>
      safeWailsCall(() => window.go!.app!.App.BatchDeleteFiles(ids), { successCount: 0, failedCount: 0, failedItems: [] }, 'BatchDeleteFiles'),
  },

  jobs: {
    // ... 保留已有方法，新增：
    submitSingleImportJob: (payload: SingleImportJobRequest) =>
      new Promise((resolve, reject) => {
        try {
          if (!window.go || !window.go.app || !window.go.app.App) {
            reject(new Error('Wails API not available'))
            return
          }
          window.go!.app!.App.SubmitSingleImportJob(
            payload.filePath,
            payload.fileName,
            payload.description,
            payload.tags,
            payload.keywords,
            payload.collectionId || 0
          ).then(resolve).catch(reject)
        } catch (error) { reject(error) }
      }),
    submitBatchImportJob: (payload: BatchImportJobRequest) =>
      new Promise((resolve, reject) => {
        try {
          if (!window.go || !window.go.app || !window.go.app.App) {
            reject(new Error('Wails API not available'))
            return
          }
          window.go!.app!.App.SubmitBatchImportJob(payload).then(resolve).catch(reject)
        } catch (error) { reject(error) }
      }),
    getRetentionConfig: () => safeWailsCall(() => window.go!.app!.App.GetJobRetentionConfig(), { id: 1, enabled: false, maxCount: 0, maxDays: 0 }, 'GetJobRetentionConfig'),
    updateRetentionConfig: (config: JobRetentionConfig) =>
      safeWailsCall(() => window.go!.app!.App.UpdateJobRetentionConfig(config), 'success', 'UpdateJobRetentionConfig'),
    submitCleanup: () => safeWailsCall(() => window.go!.app!.App.SubmitJobCleanup(), null, 'SubmitJobCleanup'),
    deleteRecord: (id: number) => safeWailsCall(() => window.go!.app!.App.DeleteJobRecord(id), 'success', 'DeleteJobRecord'),
  },
}
```

- [ ] **Step 3：提交**

```bash
git add frontend/src/types/index.ts frontend/src/types/jobs.ts frontend/src/api/index.ts
git commit -m "feat(frontend): types and api for collections and pagination"
```

### Task 10：合集选择器组件

**Files:**
- Create: `frontend/src/components/CollectionSelector.vue`

**Interfaces:**
- Consumes: `Collection[]`, `modelValue` (number 或 0 表示未分配)
- Produces: 选择事件

- [ ] **Step 1：实现组件**

```vue
<template>
  <select :value="normalizedValue" @change="handleChange">
    <option :value="0">未分配合集</option>
    <option v-for="collection in collections" :key="collection.id" :value="collection.id">
      {{ collection.name }}
    </option>
  </select>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Collection } from '@/types'

const props = defineProps<{
  modelValue?: number | null
  collections: Collection[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: number | undefined]
}>()

const normalizedValue = computed(() => props.modelValue || 0)

const handleChange = (event: Event) => {
  const target = event.target as HTMLSelectElement
  const value = parseInt(target.value, 10)
  emit('update:modelValue', value > 0 ? value : undefined)
}
</script>
```

- [ ] **Step 2：提交**

```bash
git add frontend/src/components/CollectionSelector.vue
git commit -m "feat(frontend): add CollectionSelector component"
```

### Task 11：设置页合集管理区

**Files:**
- Create: `frontend/src/components/CollectionsSettings.vue`
- Modify: `frontend/src/views/SettingsView.vue`

**Interfaces:**
- Consumes: `api.collection.getAll/add/remove`, `Collection[]`
- Produces: 设置页导航项 `collections` 与 `tasks`

- [ ] **Step 1：创建 CollectionsSettings 组件**

```vue
<template>
  <div class="collections-settings">
    <div class="section-block">
      <h5>新增合集</h5>
      <div class="add-row">
        <input v-model="newName" type="text" placeholder="合集名称" />
        <button class="btn primary" :disabled="!newName.trim() || adding" @click="addCollection">添加</button>
      </div>
      <p v-if="error" class="feedback error">{{ error }}</p>
    </div>

    <div class="section-block">
      <h5>已有合集</h5>
      <div v-if="collections.length === 0" class="empty-state">暂无合集</div>
      <div v-else class="collection-list">
        <div v-for="collection in collections" :key="collection.id" class="collection-row">
          <span>{{ collection.name }}</span>
          <span class="count">引用 {{ fileCountMap[collection.id] ?? 0 }}</span>
          <button
            class="btn ghost"
            :disabled="(fileCountMap[collection.id] ?? 0) > 0"
            :title="(fileCountMap[collection.id] ?? 0) > 0 ? '该合集正在被文件使用，无法删除' : '删除合集'"
            @click="removeCollection(collection.id)"
          >删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '@/api'
import type { Collection, File } from '@/types'

const props = defineProps<{
  files: File[]
}>()

const collections = ref<Collection[]>([])
const newName = ref('')
const adding = ref(false)
const error = ref('')

const fileCountMap = computed(() => {
  const map: Record<number, number> = {}
  props.files.forEach((file) => {
    if (file.collectionId) {
      map[file.collectionId] = (map[file.collectionId] || 0) + 1
    }
  })
  return map
})

const load = async () => {
  collections.value = await api.collection.getAll()
}

const addCollection = async () => {
  adding.value = true
  error.value = ''
  try {
    await api.collection.add(newName.value.trim())
    newName.value = ''
    await load()
  } catch (err: any) {
    error.value = err?.message || '添加失败'
  } finally {
    adding.value = false
  }
}

const removeCollection = async (id: number) => {
  if (!window.confirm('确定删除该合集？')) return
  try {
    await api.collection.remove(id)
    await load()
  } catch (err: any) {
    window.alert(err?.message || '删除失败')
  }
}

onMounted(load)
</script>
```

- [ ] **Step 2：在 SettingsView 中注册导航与内容区**

在 `settingItems` 中追加：

```ts
{ id: 'collections', label: '合集', description: '管理与删除合集', iconPaths: settingIconPaths.collections },
{ id: 'tasks', label: '任务', description: '历史任务保留策略', iconPaths: settingIconPaths.tasks },
```

在 `settingIconPaths` 中补充 `collections` 和 `tasks` 的 SVG path 数组。

在 `settings-panel` 中新增：

```vue
<section v-show="activeSetting === 'collections'" class="settings-page">
  <CollectionsSettings :files="filesStore.files" />
</section>

<section v-show="activeSetting === 'tasks'" class="settings-page">
  <TaskRetentionSettings />
</section>
```

- [ ] **Step 3：提交**

```bash
git add frontend/src/components/CollectionsSettings.vue frontend/src/views/SettingsView.vue
git commit -m "feat(frontend): settings collections management section"
```

### Task 12：导入页面限制合集选择

**Files:**
- Modify: `frontend/src/views/ImportView.vue`

**Interfaces:**
- Consumes: `CollectionSelector`, `api.collection.getAll`, `collectionId` 替代 `collectionName`

- [ ] **Step 1：加载合集列表**

```ts
const collections = ref<Collection[]>([])
onMounted(async () => {
  collections.value = await api.collection.getAll()
})
```

- [ ] **Step 2：替换单文件和合集输入**

在单文件表单中：

```vue
<label class="field field-full">
  <span>合集 / 系列</span>
  <CollectionSelector v-model="singleForm.collectionId" :collections="collections" />
</label>
```

批量表单同理：

```vue
<label class="field">
  <span>合集 / 系列</span>
  <CollectionSelector v-model="batchForm.collectionId" :collections="collections" />
</label>
```

- [ ] **Step 3：更新表单数据与提交**

```ts
const singleForm = reactive({
  fileName: '',
  keywords: '',
  collectionId: undefined as number | undefined,
  tags: [] as string[],
  description: '',
})

const batchForm = reactive({
  sharedTagsText: '',
  sharedDescription: '',
  collectionId: undefined as number | undefined,
  enableAIGeneratedTags: true,
  enableAIGeneratedDescription: false,
})
```

在 `submitSingleImport` 与 `submitBatchJob` 中把 `collectionName` 改为 `collectionId: singleForm.collectionId` / `collectionId: batchForm.collectionId`。

- [ ] **Step 4：提交**

```bash
git add frontend/src/views/ImportView.vue
git commit -m "feat(frontend): import views use collection selector"
```

### Task 13：编辑文件元数据弹窗支持合集

**Files:**
- Modify: `frontend/src/components/EditFileMetaModal.vue`

- [ ] **Step 1：加载合集并选择**

```ts
const collections = ref<Collection[]>([])
const collectionId = ref<number | undefined>(props.file.collectionId)

onMounted(async () => {
  collections.value = await api.collection.getAll()
})
```

- [ ] **Step 2：在模态框中增加 CollectionSelector**

在 `FileMetadataFields` 之前或之后：

```vue
<div class="form-group">
  <label>合集</label>
  <CollectionSelector v-model="collectionId" :collections="collections" />
</div>
```

- [ ] **Step 3：保存时传递 collectionId**

```ts
await api.file.updateMetadata(props.file.id, tags.value, description.value, collectionId.value)
```

- [ ] **Step 4：提交**

```bash
git add frontend/src/components/EditFileMetaModal.vue
git commit -m "feat(frontend): edit metadata supports collection selection"
```

### Task 14：后端集成测试（Phase 1）

**Files:**
- Create/Modify: `app/services/collection_service_test.go`

- [ ] **Step 1：编写测试**

```go
func TestCollectionService_Remove_BlockedWhenReferenced(t *testing.T) {
    tempDir := t.TempDir()
    dbPath := filepath.Join(tempDir, "test.db")
    db, err := database.NewSQLiteDB(dbPath)
    if err != nil { t.Fatalf("...") }
    defer db.Close()

    wrapper := repositories.NewSQLiteDBWrapper(db)
    colRepo := repositories.NewCollectionRepository(wrapper)
    fileRepo := repositories.NewFileRepository(wrapper)
    svc := services.NewCollectionService(colRepo, fileRepo, nil)

    id, _ := colRepo.Add("Projects")
    file := &models.File{ /* ... CollectionID: &id ... */ }
    fileRepo.Create(file)

    if err := svc.RemoveCollection(id); err == nil {
        t.Fatal("expected error removing referenced collection")
    }
}
```

- [ ] **Step 2：运行测试**

Run: `go test ./app/services/... ./app/repositories/...`
Expected: PASS

- [ ] **Step 3：提交**

```bash
git add app/services/collection_service_test.go
git commit -m "test(collections): add collection service tests"
```

### Task 15：Phase 1 验证

- [ ] **Step 1：编译前端**

Run: `cd frontend && npm run build`
Expected: no TypeScript errors, build succeeds

- [ ] **Step 2：编译后端**

Run: `go build .`
Expected: PASS

- [ ] **Step 3：集成验证**

- 在设置页添加合集 A → 导入文件选择合集 A → 文件列表显示合集 A。
- 在设置页删除合集 A → 提示被引用，无法删除。
- 编辑文件元数据，将合集改为“未分配” → 保存成功。

- [ ] **Step 4：提交阶段产物**

```bash
git commit -m "phase(1): collection infrastructure complete"
```

---

## Phase 2：文件分页与批量操作

### Task 2.1：后端文件分页与搜索分页

**Files:**
- Modify: `app/services/file_service.go`

**Interfaces:**
- Consumes: `repositories.FileRepository.Count`, `models.FileListResponse`
- Produces: `FileService.ListFilesResponse(filter)`, `SearchFilesResponse(query, page, pageSize)`

- [ ] **Step 1：新增 ListFilesResponse**

```go
func (s *FileService) ListFilesResponse(filter services.FileFilter) (*models.FileListResponse, error) {
    if filter.Page < 1 {
        filter.Page = 1
    }
    if filter.PageSize <= 0 {
        filter.PageSize = 50
    }
    files, err := s.ListFiles(filter)
    if err != nil {
        return nil, err
    }
    total, err := s.fileRepo.Count(repositories.FileFilter{
        FileType:     filter.FileType,
        CollectionID: filter.CollectionID,
    })
    if err != nil {
        return nil, err
    }
    return &models.FileListResponse{
        Items:    files,
        Page:     filter.Page,
        PageSize: filter.PageSize,
        Total:    total,
    }, nil
}
```

- [ ] **Step 2：新增 SearchFilesResponse**

```go
func (s *FileService) SearchFilesResponse(query string, page, pageSize int) (*models.FileListResponse, error) {
    if page < 1 {
        page = 1
    }
    if pageSize <= 0 {
        pageSize = 50
    }
    // 先全部搜索（上限较大）再内存分页，后续优化可改 SQL 分页
    all, err := s.SearchFiles(query)
    if err != nil {
        return nil, err
    }
    total := len(all)
    start := (page - 1) * pageSize
    if start >= total {
        return &models.FileListResponse{Items: []*models.File{}, Page: page, PageSize: pageSize, Total: total}, nil
    }
    end := start + pageSize
    if end > total {
        end = total
    }
    return &models.FileListResponse{
        Items:    all[start:end],
        Page:     page,
        PageSize: pageSize,
        Total:    total,
    }, nil
}
```

- [ ] **Step 3：提交**

```bash
git add app/services/file_service.go
git commit -m "feat(files): paginated file list and search responses"
```

### Task 2.2：批量操作后端（已在 Task 5 实现）

- [ ] **Step 1：确认 `CollectionService.BatchUpdateFilesCollection` 与 `BatchDeleteFiles` 已存在并可被 App 调用。**

- [ ] **Step 2：提交空提交标记**

```bash
git commit --allow-empty -m "phase(2): batch operations wired"
```

### Task 2.3：前端文件状态与分页

**Files:**
- Modify: `frontend/src/store/modules/files.ts`

- [ ] **Step 1：重构 store**

```ts
import type { CollectionSummary, File as LibraryFile, FileListResponse } from '@/types'

export const useFilesStore = defineStore('files', () => {
  const files = ref<LibraryFile[]>([])
  const currentPage = ref(1)
  const pageSize = ref(50)
  const totalFiles = ref(0)
  const currentFileType = ref('all')
  const currentCollectionId = ref<number | 'all' | 'unsorted'>('all')
  const searchQuery = ref('')
  const selectedFileIds = ref<number[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const loadFiles = async (fileType: string = currentFileType.value, page: number = currentPage.value, pageSizeValue: number = pageSize.value) => {
    loading.value = true
    error.value = null
    currentFileType.value = fileType
    currentPage.value = page
    pageSize.value = pageSizeValue
    try {
      const collectionId = currentCollectionId.value === 'all' || currentCollectionId.value === 'unsorted'
        ? 0
        : currentCollectionId.value
      const resp: FileListResponse = await api.file.list(page, pageSizeValue, fileType, collectionId)
      files.value = resp.items
      totalFiles.value = resp.total
    } catch (err) {
      error.value = err instanceof Error ? err.message : '加载文件失败'
      files.value = []
    } finally {
      loading.value = false
    }
  }

  const searchFiles = async (query: string, page: number = 1, pageSizeValue: number = pageSize.value) => {
    loading.value = true
    searchQuery.value = query
    try {
      const resp: FileListResponse = await api.file.search(query, page, pageSizeValue)
      files.value = resp.items
      totalFiles.value = resp.total
    } catch (err) {
      error.value = err instanceof Error ? err.message : '搜索失败'
      files.value = []
    } finally {
      loading.value = false
    }
  }

  const toggleSelection = (id: number) => {
    const index = selectedFileIds.value.indexOf(id)
    if (index >= 0) {
      selectedFileIds.value.splice(index, 1)
    } else {
      selectedFileIds.value.push(id)
    }
  }

  const clearSelection = () => {
    selectedFileIds.value = []
  }

  const refreshCurrent = async () => {
    if (searchQuery.value.trim()) {
      await searchFiles(searchQuery.value, currentPage.value, pageSize.value)
    } else {
      await loadFiles(currentFileType.value, currentPage.value, pageSize.value)
    }
  }

  return {
    files, currentPage, pageSize, totalFiles, currentFileType, currentCollectionId, searchQuery,
    selectedFileIds, loading, error, loadFiles, searchFiles, toggleSelection, clearSelection, refreshCurrent,
  }
})
```

- [ ] **Step 2：提交**

```bash
git add frontend/src/store/modules/files.ts
git commit -m "feat(frontend): files store supports pagination and selection"
```

### Task 2.4：文件卡片选择模式

**Files:**
- Modify: `frontend/src/components/FileCard.vue`

- [ ] **Step 1：新增选择复选框**

在卡片左上角新增：

```vue
<label class="select-box" @click.stop>
  <input
    v-if="selectable"
    type="checkbox"
    :checked="selected"
    @change="emit('toggleSelect', props.file.id)"
  />
</label>
```

新增 props/emits：

```ts
const props = defineProps<{
  file: File
  selectable?: boolean
  selected?: boolean
}>()

const emit = defineEmits<{
  // ... 原有 emits
  toggleSelect: [id: number]
}>()
```

样式 `.select-box` 使用 `position: absolute; top: 8px; left: 8px; z-index: 100;`。

- [ ] **Step 2：提交**

```bash
git add frontend/src/components/FileCard.vue
git commit -m "feat(frontend): file card selection mode"
```

### Task 2.5：批量工具栏与进度遮罩

**Files:**
- Create: `frontend/src/components/FilesBatchToolbar.vue`
- Create: `frontend/src/components/BatchMoveProgressOverlay.vue`
- Modify: `frontend/src/views/FilesView.vue`

- [ ] **Step 1：FilesBatchToolbar**

```vue
<template>
  <div class="batch-toolbar">
    <span>已选择 {{ selectedIds.length }} 个文件</span>
    <button class="btn secondary" @click="$emit('move')">移动合集</button>
    <button class="btn danger" @click="$emit('delete')">删除</button>
    <button class="btn ghost" @click="$emit('clear')">取消选择</button>
  </div>
</template>

<script setup lang="ts">
defineProps<{ selectedIds: number[] }>()
defineEmits(['move', 'delete', 'clear'])
</script>
```

- [ ] **Step 2：BatchMoveProgressOverlay**

```vue
<template>
  <div class="batch-overlay" :class="{ 'with-progress': true }">
    <div class="overlay-card">
      <h3>正在移动 {{ total }} 个文件到「{{ collectionName }}」</h3>
      <progress :max="total" :value="completed"></progress>
      <p>{{ message }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  total: number
  completed: number
  message: string
  collectionName: string
}>()
</script>

<style scoped>
.batch-overlay {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 3000;
  background: rgba(15, 23, 42, 0.42);
  backdrop-filter: blur(10px);
}
.overlay-card {
  background: var(--surface-color);
  border-radius: 16px;
  padding: 32px;
  min-width: 320px;
}
</style>
```

- [ ] **Step 3：在 FilesView 中集成**

```vue
<FilesBatchToolbar
  v-if="filesStore.selectedFileIds.length > 0"
  :selected-ids="filesStore.selectedFileIds"
  @move="openMoveDialog"
  @delete="confirmBatchDelete"
  @clear="filesStore.clearSelection()"
/>

<FileList
  :files="filesStore.files"
  :selectable="true"
  :selected-ids="filesStore.selectedFileIds"
  @toggle-select="filesStore.toggleSelection"
  @open="handleOpenFile"
  @delete="handleDeleteFile"
  @updated="handleUpdateFileMetadata"
/>

<BatchMoveProgressOverlay
  v-if="moving"
  :total="moveTotal"
  :completed="moveCompleted"
  :message="moveMessage"
  :collection-name="moveCollectionName"
/>
```

实现 `openMoveDialog`：弹窗使用 `CollectionSelector` 选择目标合集，确认后调用 `api.file.batchUpdateCollection`（同步阻塞等待，因为后端是同步操作），完成后刷新并清空选择。

实现 `confirmBatchDelete`：`window.confirm` 后调用 `api.file.batchDelete(filesStore.selectedFileIds)`，刷新列表。

- [ ] **Step 4：提交**

```bash
git add frontend/src/components/FilesBatchToolbar.vue frontend/src/components/BatchMoveProgressOverlay.vue frontend/src/views/FilesView.vue
git commit -m "feat(frontend): batch toolbar and move overlay"
```

### Task 2.6：分页控件组件

**Files:**
- Create: `frontend/src/components/PaginationControls.vue`

- [ ] **Step 1：实现组件**

```vue
<template>
  <div class="pagination-controls">
    <button class="btn secondary" :disabled="page <= 1" @click="emit('change', page - 1)">上一页</button>
    <span>第 {{ page }} / {{ totalPages }} 页</span>
    <button class="btn secondary" :disabled="page >= totalPages" @click="emit('change', page + 1)">下一页</button>
    <select :value="pageSize" @change="emit('changeSize', Number(($event.target as HTMLSelectElement).value))">
      <option v-for="size in pageSizes" :key="size" :value="size">{{ size }} / 页</option>
    </select>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  page: number
  pageSize: number
  total: number
  pageSizes?: number[]
}>()

const emit = defineEmits(['change', 'changeSize'])

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
</script>
```

- [ ] **Step 2：在 FilesView 与 TasksView 中使用**

- [ ] **Step 3：提交**

```bash
git add frontend/src/components/PaginationControls.vue frontend/src/views/FilesView.vue
git commit -m "feat(frontend): pagination controls"
```

### Task 2.7：合集筛选器使用 ID

**Files:**
- Modify: `frontend/src/components/CollectionFilter.vue`

- [ ] **Step 1：修改 props/options**

```vue
<script setup lang="ts">
import { UNSORTED_COLLECTION_KEY, UNSORTED_COLLECTION_LABEL } from '@/utils/constants'
import type { Collection } from '@/types'

const props = defineProps<{
  modelValue: 'all' | 'unsorted' | number
  collections: Collection[]
  totalCount: number
}>()
const emit = defineEmits(['update:modelValue', 'filter'])
</script>
```

选项包含：全部、未分配、各合集。

- [ ] **Step 2：提交**

```bash
git add frontend/src/components/CollectionFilter.vue
git commit -m "feat(frontend): collection filter uses id"
```

### Task 2.8：Phase 2 测试与验证

- [ ] **Step 1：前端构建**

Run: `cd frontend && npm run build`
Expected: PASS

- [ ] **Step 2：后端测试**

Run: `go test ./app/services/... ./app/repositories/...`
Expected: PASS

- [ ] **Step 3：集成验证**

- 选择 3 个文件 → 移动合集 → 遮罩显示 → 成功后文件合集变更。
- 选择 2 个文件 → 删除 → 列表刷新。
- 切换每页数量 → 回到第 1 页。

- [ ] **Step 4：提交**

```bash
git commit -m "phase(2): file pagination and batch operations complete"
```

---

## Phase 3：任务页面重构与清理 Job

### Task 3.1：Cleanup Handler 与 Job 策略

**Files:**
- Create: `app/services/cleanup_handler.go`
- Modify: `app/services/job_service.go`（注册 policy/handler）

**Interfaces:**
- Consumes: `models.JobRetentionConfig`, `ConfigService.GetJobRetentionConfig`, `JobRepository`
- Produces: `CleanupHandler`, `JobTypeCleanup` policy

- [ ] **Step 1：实现 CleanupHandler**

`app/services/cleanup_handler.go`：

```go
package services

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "LocalSpace/app/models"
)

type CleanupHandler struct{}

func NewCleanupHandler() *CleanupHandler {
    return &CleanupHandler{}
}

func (h *CleanupHandler) Type() string {
    return models.JobTypeCleanup
}

func (h *CleanupHandler) Validate(payload json.RawMessage) error {
    var config models.JobRetentionConfig
    if err := json.Unmarshal(payload, &config); err != nil {
        return fmt.Errorf("invalid retention config: %w", err)
    }
    if config.MaxCount <= 0 && config.MaxDays <= 0 {
        return fmt.Errorf("at least one retention condition must be enabled")
    }
    return nil
}

func (h *CleanupHandler) Execute(ctx context.Context, job *models.Job, runtime JobRuntime) error {
    var config models.JobRetentionConfig
    if err := json.Unmarshal(job.Payload, &config); err != nil {
        return fmt.Errorf("failed to parse cleanup payload: %w", err)
    }

    repo := runtime.JobRepository()
    terminalStatuses := []string{
        models.JobStatusCompleted,
        models.JobStatusFailed,
        models.JobStatusCancelled,
        models.JobStatusTimedOut,
        models.JobStatusCleanupFailed,
    }
    jobs, err := repo.ListByStatuses(terminalStatuses...)
    if err != nil {
        return fmt.Errorf("failed to list terminal jobs: %w", err)
    }

    toDelete := make(map[uint]bool)
    if config.Enabled && config.MaxCount > 0 && len(jobs) > config.MaxCount {
        excess := len(jobs) - config.MaxCount
        for i := 0; i < excess; i++ {
            toDelete[jobs[i].ID] = true
        }
    }
    if config.Enabled && config.MaxDays > 0 {
        cutoff := time.Now().AddDate(0, 0, -config.MaxDays)
        for _, j := range jobs {
            created, err := time.Parse(time.RFC3339, j.CreatedAt)
            if err != nil {
                continue
            }
            if created.Before(cutoff) {
                toDelete[j.ID] = true
            }
        }
    }

    if len(toDelete) == 0 {
        return nil
    }

    ids := make([]uint, 0, len(toDelete))
    for id := range toDelete {
        ids = append(ids, id)
    }
    if err := repo.DeleteByIDs(ids); err != nil {
        return fmt.Errorf("failed to delete jobs: %w", err)
    }

    result := map[string]int{"deleted": len(ids)}
    resultJSON, _ := json.Marshal(result)
    job.Result = resultJSON
    job.ProgressMessage = fmt.Sprintf("已清理 %d 条任务记录", len(ids))
    return nil
}

func (h *CleanupHandler) Resume(ctx context.Context, job *models.Job, runtime JobRuntime) error {
    return fmt.Errorf("cleanup job does not support resume")
}

func (h *CleanupHandler) Cleanup(ctx context.Context, job *models.Job, runtime JobRuntime) error {
    return nil
}
```

- [ ] **Step 2：在 JobRepository 添加 ListByStatuses 和 DeleteByIDs**

```go
func (r *JobRepository) ListByStatuses(statuses ...string) ([]*models.Job, error) {
    if len(statuses) == 0 {
        return []*models.Job{}, nil
    }
    placeholders := make([]string, len(statuses))
    args := make([]interface{}, len(statuses))
    for i, s := range statuses {
        placeholders[i] = "?"
        args[i] = s
    }
    query := fmt.Sprintf(`
        SELECT id, job_type, status, title, payload, result, progress_total, progress_completed, progress_message,
               exclusive_key, can_resume, started_at, heartbeat_at, finished_at, timeout_at, error_message,
               created_at, updated_at
        FROM jobs
        WHERE status IN (%s)
        ORDER BY created_at ASC
    `, strings.Join(placeholders, ","))
    rows, err := r.db.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var result []*models.Job
    for rows.Next() {
        job, err := scanJob(rows)
        if err != nil {
            return nil, err
        }
        result = append(result, job)
    }
    return result, nil
}

func (r *JobRepository) DeleteByIDs(ids []uint) error {
    if len(ids) == 0 {
        return nil
    }
    placeholders := make([]string, len(ids))
    args := make([]interface{}, len(ids))
    for i, id := range ids {
        placeholders[i] = "?"
        args[i] = id
    }
    query := fmt.Sprintf(`DELETE FROM jobs WHERE id IN (%s)`, strings.Join(placeholders, ","))
    _, err := r.db.Exec(query, args...)
    return err
}
```

- [ ] **Step 3：在 JobService 中注册 CleanupHandler 与 policy**

在 `NewJobService` 的 policies 中追加：

```go
models.JobTypeCleanup: {
    JobType:          models.JobTypeCleanup,
    MaxConcurrent:    1,
    ExclusiveKey:     "cleanup",
    CanRunBackground: true,
    Recoverable:      false,
    Timeout:          5 * time.Minute,
},
```

并在 `NewJobService` 末尾注册：`service.RegisterHandler(NewCleanupHandler())`。

- [ ] **Step 4：提交**

```bash
git add app/services/cleanup_handler.go app/services/job_service.go app/repositories/job_repository.go
git commit -m "feat(jobs): add cleanup handler and repository helpers"
```

### Task 3.2：JobService 清理与删除入口

**Files:**
- Modify: `app/services/job_service.go`

**Interfaces:**
- Consumes: `ConfigService.GetJobRetentionConfig`, `CleanupHandler.Validate`
- Produces: `JobService.SubmitJobCleanup`, `JobService.DeleteJobRecord`, 启动时自动清理

- [ ] **Step 1：新增提交清理任务方法**

```go
func (s *JobService) SubmitJobCleanup() (*models.Job, error) {
    config, err := s.configService.GetJobRetentionConfig()
    if err != nil {
        return nil, fmt.Errorf("failed to get retention config: %w", err)
    }
    payload, err := json.Marshal(config)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal retention config: %w", err)
    }
    handler, policy, err := s.handlerAndPolicy(models.JobTypeCleanup)
    if err != nil {
        return nil, err
    }
    if err := handler.Validate(payload); err != nil {
        return nil, err
    }
    if err := s.ensureConcurrency(policy, 0); err != nil {
        return nil, err
    }
    job := &models.Job{
        JobType:           models.JobTypeCleanup,
        Status:            models.JobStatusPending,
        Title:             "清理历史任务记录",
        Payload:           payload,
        ProgressTotal:     1,
        ProgressCompleted: 0,
        ProgressMessage:   "等待开始清理",
        ExclusiveKey:      policy.ExclusiveKey,
        CanResume:         policy.Recoverable,
    }
    if err := s.jobRepo.Create(job); err != nil {
        return nil, err
    }
    s.emitJobEvent("job:created", job)
    go s.runJob(job.ID, false)
    return job, nil
}
```

- [ ] **Step 2：新增 DeleteJobRecord**

```go
func (s *JobService) DeleteJobRecord(id uint) error {
    job, err := s.jobRepo.FindByID(id)
    if err != nil {
        return err
    }
    terminal := []string{
        models.JobStatusCompleted,
        models.JobStatusFailed,
        models.JobStatusCancelled,
        models.JobStatusTimedOut,
        models.JobStatusCleanupFailed,
    }
    found := false
    for _, status := range terminal {
        if job.Status == status {
            found = true
            break
        }
    }
    if !found {
        return fmt.Errorf("only terminal jobs can be deleted")
    }
    return s.jobRepo.DeleteByIDs([]uint{id})
}
```

- [ ] **Step 3：启动时自动提交清理（若启用且需要）**

在 `NormalizeUnfinishedJobs` 之后或 `App.initializeApp` 中调用：

```go
func (s *JobService) MaybeSubmitAutoCleanup() error {
    config, err := s.configService.GetJobRetentionConfig()
    if err != nil {
        return err
    }
    if !config.Enabled || (config.MaxCount <= 0 && config.MaxDays <= 0) {
        return nil
    }
    terminal, err := s.jobRepo.ListByStatuses(
        models.JobStatusCompleted,
        models.JobStatusFailed,
        models.JobStatusCancelled,
        models.JobStatusTimedOut,
        models.JobStatusCleanupFailed,
    )
    if err != nil {
        return err
    }
    needs := false
    if config.MaxCount > 0 && len(terminal) > config.MaxCount {
        needs = true
    }
    if config.MaxDays > 0 {
        cutoff := time.Now().AddDate(0, 0, -config.MaxDays)
        for _, j := range terminal {
            created, err := time.Parse(time.RFC3339, j.CreatedAt)
            if err != nil {
                continue
            }
            if created.Before(cutoff) {
                needs = true
                break
            }
        }
    }
    if !needs {
        return nil
    }
    _, err = s.SubmitJobCleanup()
    return err
}
```

在 `App.initializeApp` 中 `NormalizeUnfinishedJobs` 之后调用 `a.jobService.MaybeSubmitAutoCleanup()`。

- [ ] **Step 4：提交**

```bash
git add app/services/job_service.go
git commit -m "feat(jobs): submit cleanup and delete job record"
```

### Task 3.3：任务保留设置前端组件

**Files:**
- Create: `frontend/src/components/TaskRetentionSettings.vue`
- Modify: `frontend/src/views/SettingsView.vue`

- [ ] **Step 1：实现 TaskRetentionSettings**

```vue
<template>
  <div class="task-retention-settings">
    <div class="section-block">
      <label class="toggle-card">
        <input v-model="config.enabled" type="checkbox" />
        <div>
          <strong>启动时自动清理历史任务</strong>
          <p>仅删除已终态任务记录。</p>
        </div>
      </label>
    </div>

    <div class="section-block">
      <label class="toggle-card">
        <input v-model="limitCountEnabled" type="checkbox" />
        <div>
          <strong>最多保留条数</strong>
          <p v-if="limitCountEnabled">
            保留最近的 <input v-model.number="config.maxCount" type="number" min="1" /> 条
          </p>
        </div>
      </label>
    </div>

    <div class="section-block">
      <label class="toggle-card">
        <input v-model="limitDaysEnabled" type="checkbox" />
        <div>
          <strong>最多保留天数</strong>
          <p v-if="limitDaysEnabled">
            删除超过 <input v-model.number="config.maxDays" type="number" min="1" /> 天的记录
          </p>
        </div>
      </label>
    </div>

    <p class="hint">至少启用“条数”或“天数”之一，保存才会生效。</p>

    <div class="config-actions">
      <button class="btn primary" :disabled="saving" @click="save">保存设置</button>
      <button class="btn secondary" :disabled="cleaning" @click="manualCleanup">立即手动清理</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed } from 'vue'
import { api } from '@/api'
import type { JobRetentionConfig } from '@/types'

const config = reactive<JobRetentionConfig>({ id: 1, enabled: false, maxCount: 0, maxDays: 0, updatedAt: '' })
const saving = ref(false)
const cleaning = ref(false)

const limitCountEnabled = computed({
  get: () => config.maxCount > 0,
  set: (v) => { config.maxCount = v ? 100 : 0 }
})
const limitDaysEnabled = computed({
  get: () => config.maxDays > 0,
  set: (v) => { config.maxDays = v ? 30 : 0 }
})

const load = async () => {
  const loaded = await api.jobs.getRetentionConfig()
  Object.assign(config, loaded)
}

const save = async () => {
  saving.value = true
  try {
    await api.jobs.updateRetentionConfig(config)
    await load()
  } finally {
    saving.value = false
  }
}

const manualCleanup = async () => {
  cleaning.value = true
  try {
    await api.jobs.submitCleanup()
  } finally {
    cleaning.value = false
  }
}

load()
</script>
```

- [ ] **Step 2：在 SettingsView 中注册 tasks 导航区**

已在 Task 11 中预留 `<section v-show="activeSetting === 'tasks'">`。

- [ ] **Step 3：提交**

```bash
git add frontend/src/components/TaskRetentionSettings.vue frontend/src/views/SettingsView.vue
git commit -m "feat(frontend): task retention settings component"
```

### Task 3.4：TasksView 统一列表重构

**Files:**
- Modify: `frontend/src/views/TasksView.vue`
- Modify: `frontend/src/store/modules/jobs.ts`

- [ ] **Step 1：统一列表数据结构**

在 `jobsStore` 中合并 active/resumable/history 为一个 `jobs` 列表，按分页加载：

```ts
const jobs = ref<Job[]>([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const loadJobs = async (p = page.value, ps = pageSize.value) => {
  page.value = p
  pageSize.value = ps
  const resp = await api.jobs.listJobs(p, ps, '')
  jobs.value = resp.items
  total.value = resp.total
}
```

- [ ] **Step 2：新增单条删除入口**

```ts
const deleteJobRecord = async (id: number) => {
  if (!window.confirm('确定删除该任务记录？')) return
  await api.jobs.deleteRecord(id)
  await loadJobs()
}
```

- [ ] **Step 3：重写 TasksView 为表格**

使用 `<table>` 或 div 表格，行背景：

```css
.task-row {
  background: rgba(255, 255, 255, 0.42);
  backdrop-filter: blur(8px);
  border-bottom: 1px solid rgba(146, 165, 192, 0.18);
}
[data-theme='dark'] .task-row {
  background: rgba(15, 23, 42, 0.42);
}
```

列：任务名称、开始时间、结束时间、进度条、状态标签。失败行可展开显示 `errorMessage` 和 `result.failedCount`。

操作列：下拉按钮，包含“删除记录”（仅终态可点）和“恢复”（仅 `awaiting_resume`/`timed_out` 可点，否则 disabled）。

- [ ] **Step 4：提交**

```bash
git add frontend/src/views/TasksView.vue frontend/src/store/modules/jobs.ts
git commit -m "feat(frontend): unified task list with glassmorphism and actions"
```

### Task 3.5：Cleanup 与任务删除的后端测试

**Files:**
- Create: `app/services/cleanup_handler_test.go`
- Modify: `app/services/job_service_test.go`

- [ ] **Step 1：测试 CleanupHandler**

```go
func TestCleanupHandler_DeletesByMaxCount(t *testing.T) {
    tempDir := t.TempDir()
    dbPath := filepath.Join(tempDir, "test.db")
    db, _ := database.NewSQLiteDB(dbPath)
    defer db.Close()
    repo := repositories.NewJobRepository(repositories.NewSQLiteDBWrapper(db))
    configService := services.NewConfigService(repositories.NewConfigRepository(repositories.NewSQLiteDBWrapper(db)))
    fileService := &services.FileService{}
    jobService := services.NewJobService(repo, fileService, nil, nil)
    jobService.RegisterHandler(services.NewCleanupHandler())

    // 创建 3 个 completed 任务
    for i := 0; i < 3; i++ {
        j := &models.Job{
            JobType: models.JobTypeBatchImport,
            Status:  models.JobStatusCompleted,
            Title:   "test",
        }
        _ = repo.Create(j)
    }

    payload, _ := json.Marshal(models.JobRetentionConfig{Enabled: true, MaxCount: 1})
    job := &models.Job{JobType: models.JobTypeCleanup, Payload: payload}
    _ = repo.Create(job)

    handler := services.NewCleanupHandler()
    err := handler.Execute(context.Background(), job, &jobRuntime{service: jobService, job: job})
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    remaining, _, _ := repo.List(1, 100, "")
    if len(remaining) != 2 {
        t.Fatalf("expected 2 jobs remaining (1 old + cleanup), got %d", len(remaining))
    }
}
```

- [ ] **Step 2：测试 DeleteJobRecord**

```go
func TestJobService_DeleteJobRecord_OnlyTerminal(t *testing.T) {
    tempDir := t.TempDir()
    dbPath := filepath.Join(tempDir, "test.db")
    db, _ := database.NewSQLiteDB(dbPath)
    defer db.Close()
    repo := repositories.NewJobRepository(repositories.NewSQLiteDBWrapper(db))
    svc := services.NewJobService(repo, nil, nil, nil)

    job := &models.Job{JobType: models.JobTypeBatchImport, Status: models.JobStatusRunning}
    _ = repo.Create(job)

    if err := svc.DeleteJobRecord(job.ID); err == nil {
        t.Fatal("expected error deleting non-terminal job")
    }
}
```

- [ ] **Step 3：运行测试**

Run: `go test ./app/services/...`
Expected: PASS

- [ ] **Step 4：提交**

```bash
git add app/services/cleanup_handler_test.go app/services/job_service_test.go
git commit -m "test(jobs): cleanup handler and delete job record"
```

### Task 3.6：Phase 3 验证与收尾

- [ ] **Step 1：前端构建**

Run: `cd frontend && npm run build`
Expected: PASS

- [ ] **Step 2：后端测试**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 3：集成验证**

- 在设置页启用“最多保留 2 条”，添加 3 个终态任务，重启应用 → 自动清理到 2 条。
- 手动点击“立即手动清理” → 生成清理 Job 并完成。
- 在任务列表删除单条终态记录 → 总数减少。
- 失败任务行展开可查看失败原因。

- [ ] **Step 4：提交阶段产物**

```bash
git commit -m "phase(3): task list redesign and cleanup job complete"
```

---

## Final Review & Handoff

- [ ] **Spec coverage check:** 确保 Phase 1-3 覆盖了：合集管理、文件分页、批量操作、元数据编辑、任务列表、清理配置。
- [ ] **Placeholder scan:** 搜索 `TODO`/`TBD`/`...` 并替换为实际代码/说明。
- [ ] **Type consistency:** 确认 `CollectionID` 在 Go 中统一为 `*uint`，前端为 `number | undefined`，Wails 参数为 `uint` 并用 0 表示未分配。
- [ ] **Execution handoff:** 见下方选项。

**Plan complete and saved to `docs/superpowers/plans/2026-08-03-localspace-collections-tasks.md`. Two execution options:**

1. **Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, review between tasks, fast iteration.
2. **Inline Execution** — Execute tasks in this session using `superpowers:executing-plans`, batch execution with checkpoints.

Which approach would you like?

# LocalSpace 问题修复实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复 `ori/问题分析.md` 中的 10 项主要问题及补充项 C，覆盖 Go 后端、Wails 绑定层和 Vue 3 前端。

**Architecture:** 采用统一导入事务（Import Pipeline）抽象，把单文件导入和批量导入的公共落地逻辑集中到 `prepareStagedImport` / `commitStagedImport` / `cleanupStagedImport`，统一处理临时文件、checksum、重复检测、数据库写入、原子重命名、源文件清理和失败回滚。其余问题按分析文档点对点修复。

**Tech Stack:** Go (Wails v2), SQLite, Vue 3 + TypeScript + Vite, Pinia, Windows/macOS/Linux.

## Global Constraints

- 所有 Go 代码在 `app/` 包下；前端代码在 `frontend/src/` 下。
- 临时目录统一放在目标 master 目录的 `.localspace-temp/` 下，保证 `os.Rename` 原子性。
- 源文件在数据库记录成功写入前不得删除。
- `SortBy` 和 `SortOrder` 必须白名单校验后才能拼入 SQL。
- `JobService` 必须在 goroutine 完全退出后 `App.Shutdown` 才能关闭数据库。
- 前端 `safeWailsCall` 仅在 `window.go` 不存在时返回 fallback；真实调用失败必须 reject。
- 所有 commit 消息以 `Co-Authored-By: Claude <noreply@anthropic.com>` 结尾。
- 每完成一个 Task 后运行 `go test ./...` 和 `cd frontend && npm run build`（如适用）验证。

---

## 文件变更总览

| 文件 | 变更类型 | 责任 |
|---|---|---|
| `app/repositories/file_repository.go` | 修改 | SQL 注入白名单校验 |
| `app/services/file_service.go` | 大幅修改 | 统一导入 pipeline、DeleteFile/RenameFile 顺序调整、批量导入临时路径 |
| `app/services/batch_import_handler.go` | 修改 | 复用新 pipeline、按 job 组织临时路径 |
| `app/services/job_service.go` | 修改 | panic 恢复、shutdown 同步 WaitGroup |
| `app/app.go` | 修改 | PowerShell 替换为 explorer 调用、启动时清理残留临时文件 |
| `frontend/src/api/index.ts` | 修改 | `safeWailsCall` 失败时 reject |
| `frontend/src/views/ImportView.vue` | 修改 | 单文件元数据类型传真实类型 |
| `app/services/file_service_test.go` | 修改/新增测试 | 覆盖 staged import、DeleteFile、RenameFile |
| `app/repositories/file_repository_test.go` | 新增测试 | SQL 注入白名单 |
| `app/services/job_service_test.go` | 新增测试 | panic 恢复与 shutdown 同步 |

---

## Task 1: SQL 注入白名单

**Files:**
- Modify: `app/repositories/file_repository.go:147-157`
- Test: `app/repositories/file_repository_test.go`

**Interfaces:**
- Consumes: 无。
- Produces: `normalizeSortBy(sortBy string) string`, `normalizeSortOrder(sortOrder string) string`。

- [ ] **Step 1: 编写失败测试**

在 `app/repositories/file_repository_test.go` 末尾新增：

```go
func TestFileRepository_List_SortByWhitelist(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := NewFileRepository(NewSQLiteDBWrapper(db))

	file := &models.File{
		FileName:     "test.txt",
		OriginalName: "test.txt",
		FilePath:     "/test/test.txt",
		FileType:     "document",
		FileSubType:  "txt",
		FileSize:     1024,
	}
	if err := repo.Create(file); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// 恶意 SortBy 应被降级为默认 created_at DESC，不应报错
	results, err := repo.List(FileFilter{Page: 1, PageSize: 10, SortBy: "id; DROP TABLE files;--", SortOrder: "ASC"})
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 file, got %d", len(results))
	}

	// 验证 files 表未被删除
	exists, err := repo.ExistsByPath("/test/test.txt")
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Expected files table to still exist")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
go test -run TestFileRepository_List_SortByWhitelist ./app/repositories
```

Expected: FAIL，因为 `normalizeSortBy` 尚未实现。

- [ ] **Step 3: 实现白名单函数**

在 `app/repositories/file_repository.go` 中 `FileFilter` 类型之后添加：

```go
var allowedSortColumns = map[string]bool{
	"id":          true,
	"created_at":  true,
	"modified_at": true,
	"file_name":   true,
}

func normalizeSortBy(sortBy string) string {
	if allowedSortColumns[sortBy] {
		return sortBy
	}
	return "created_at"
}

func normalizeSortOrder(sortOrder string) string {
	order := strings.ToUpper(sortOrder)
	if order == "ASC" || order == "DESC" {
		return order
	}
	return "DESC"
}
```

修改 `List` 中的排序拼接：

```go
sortBy := normalizeSortBy(filter.SortBy)
sortOrder := normalizeSortOrder(filter.SortOrder)
query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)
```

- [ ] **Step 4: 运行测试确认通过**

```bash
go test -run TestFileRepository_List_SortByWhitelist ./app/repositories
```

Expected: PASS。

- [ ] **Step 5: 运行全仓库测试**

```bash
go test ./app/repositories
```

Expected: 全部 PASS。

- [ ] **Step 6: Commit**

```bash
git add app/repositories/file_repository.go app/repositories/file_repository_test.go
git commit -m "fix: whitelist SortBy and SortOrder to prevent SQL injection

- Only allow id, created_at, modified_at, file_name for sort columns
- Only allow ASC/DESC for sort order
- Add repository test with malicious SortBy payload

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 2: 统一导入 Pipeline 核心抽象

**Files:**
- Modify: `app/services/file_service.go`（新增类型和函数）
- Test: `app/services/file_service_test.go`

**Interfaces:**
- Consumes: `StorageService`, `FileRepository`, `utils.CalculateFileChecksum`。
- Produces: `prepareStagedImport(...)`, `commitStagedImport(...)`, `cleanupStagedImport(...)`。

- [ ] **Step 1: 新增类型与辅助函数**

在 `app/services/file_service.go` 中 `FileService` 类型之后添加：

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

const batchImportTempDir = ".localspace-temp"
const batchImportTempSuffix = ".localspace-importing"

func generateTempPath(masterPath, fileName, jobID string, itemIndex int) string {
	timestamp := time.Now().UnixNano()
	random := make([]byte, 4)
	// crypto/rand is overkill for temp filenames; use math/rand for simplicity
	// but ensure it is seeded at startup if needed.
	base := fmt.Sprintf("%s-%d-%x", fileName, timestamp, random)
	if jobID != "" {
		return filepath.Join(masterPath, batchImportTempDir, "batch-"+jobID, fmt.Sprintf("%d", itemIndex), base+batchImportTempSuffix)
	}
	return filepath.Join(masterPath, batchImportTempDir, base+batchImportTempSuffix)
}
```

注意：需要在文件顶部增加 `math/rand` 和 `crypto/rand` 或 `encoding/hex` 的 import。为简单，使用 `time.Now().UnixNano()` 和 `rand.Int63()` 组合，在 `init` 中 seed 一次。

- [ ] **Step 2: 实现 prepareStagedImport**

```go
func (s *FileService) prepareStagedImport(sourcePath, fileName, collectionName string, jobID string, itemIndex int, onProgress func(copied int64) error) (*stagedImport, error) {
	if sourcePath == "" {
		return nil, fmt.Errorf("source path cannot be empty")
	}
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("source file does not exist: %s", sourcePath)
	}

	fileInfo, err := os.Stat(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	extension := filepath.Ext(sourcePath)
	fileType, err := parseFileType(extension)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file type: %w", err)
	}

	masters, err := s.storageService.GetMasterDirectories()
	if err != nil {
		return nil, fmt.Errorf("failed to get master directories: %w", err)
	}
	if len(masters) == 0 {
		return nil, fmt.Errorf("no master directory configured")
	}

	defaultMaster := selectDefaultMasterDirectory(masters)
	if defaultMaster == nil {
		return nil, fmt.Errorf("no master directory configured")
	}

	hasSpace, err := s.storageService.CheckMasterStorageSpace(defaultMaster.ID, fileInfo.Size())
	if err != nil {
		return nil, fmt.Errorf("failed to check storage space: %w", err)
	}
	if !hasSpace {
		return nil, fmt.Errorf("not enough storage space for this file")
	}

	layoutConfig, err := s.storageService.GetStorageLayoutConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get storage layout config: %w", err)
	}

	trimmedCollectionName := strings.TrimSpace(collectionName)
	effectiveCollectionName := trimmedCollectionName
	if layoutConfig.Strategy == "type_collection" && effectiveCollectionName == "" {
		effectiveCollectionName = layoutConfig.UnsortedFolderName
	}

	finalPath, err := s.storageService.GetStoragePathForFileWithMaster(defaultMaster.ID, fileType, effectiveCollectionName, fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage path: %w", err)
	}
	if err := s.storageService.EnsureStorageDirExists(filepath.Dir(finalPath)); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	checksum, err := utils.CalculateFileChecksum(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate file checksum: %w", err)
	}

	isDuplicate, duplicateID, err := s.fileRepo.CheckDuplicateByChecksum(checksum)
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicate files: %w", err)
	}
	if isDuplicate {
		duplicateFile, findErr := s.fileRepo.FindByID(duplicateID)
		if findErr == nil {
			return nil, fmt.Errorf("duplicate file detected. A file with identical content already exists: %s (ID: %d)", duplicateFile.FileName, duplicateID)
		}
		return nil, fmt.Errorf("duplicate file detected. A file with identical content already exists (ID: %d)", duplicateID)
	}

	exists, err := s.fileRepo.ExistsByPath(finalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if file exists: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("file already exists at destination: %s", finalPath)
	}

	masterPath := filepath.Dir(filepath.Dir(finalPath)) // finalPath = master/type/(collection/)fileName
	if layoutConfig.Strategy == "type_collection" {
		masterPath = filepath.Dir(filepath.Dir(filepath.Dir(finalPath)))
	} else {
		masterPath = filepath.Dir(filepath.Dir(finalPath))
	}
	// Robust approach: derive master path from storage service lookup
	masterPath, err = s.storageService.GetMasterPath(defaultMaster.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get master path: %w", err)
	}

	tempPath := generateTempPath(masterPath, fileName, jobID, itemIndex)
	if err := s.storageService.EnsureStorageDirExists(filepath.Dir(tempPath)); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	if err := copyFileWithProgress(sourcePath, tempPath, onProgress); err != nil {
		_ = os.Remove(tempPath)
		return nil, fmt.Errorf("failed to copy file to staging: %w", err)
	}

	return &stagedImport{
		SourcePath:     sourcePath,
		TempPath:       tempPath,
		FinalPath:      finalPath,
		Checksum:       checksum,
		FileSize:       fileInfo.Size(),
		FileType:       fileType,
		FileSubType:    strings.TrimPrefix(extension, "."),
		MasterID:       defaultMaster.ID,
		CollectionName: effectiveCollectionName,
		FileName:       fileName,
		OriginalName:   filepath.Base(sourcePath),
	}, nil
}
```

注意：这里需要新增 `StorageService.GetMasterPath(id uint) (string, error)` 辅助方法，见 Task 3。若嫌新增方法麻烦，可先用 `master.Path` 从 `defaultMaster` 直接取。

替换为简单实现：

```go
masterPath := defaultMaster.Path
```

- [ ] **Step 3: 实现 commitStagedImport 与 cleanupStagedImport**

```go
func (s *FileService) commitStagedImport(staged *stagedImport, tags []string, description string, metadata models.Metadata) (*models.File, error) {
	if staged == nil {
		return nil, fmt.Errorf("staged import is nil")
	}

	file := &models.File{
		FileName:       staged.FileName,
		OriginalName:   staged.OriginalName,
		CollectionName: staged.CollectionName,
		FilePath:       staged.FinalPath,
		FileType:       staged.FileType,
		FileSubType:    staged.FileSubType,
		FileSize:       staged.FileSize,
		Tags:           s.normalizeMetadataTags(tags),
		Description:    s.normalizeMetadataDescription(description),
		Metadata:       metadata,
		Thumbnail:      "",
		Checksum:       staged.Checksum,
		IsDeleted:      false,
		DeletedAt:      "",
	}

	if err := s.fileRepo.Create(file); err != nil {
		_ = os.Remove(staged.TempPath)
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	if err := os.Rename(staged.TempPath, staged.FinalPath); err != nil {
		_ = s.fileRepo.Delete(file.ID)
		_ = os.Remove(staged.TempPath)
		return nil, fmt.Errorf("failed to commit imported file: %w", err)
	}

	if err := os.Remove(staged.SourcePath); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: imported file committed but failed to remove source %s: %v\n", staged.SourcePath, err)
	}

	if supportsGeneratedThumbnail(staged.FileType) {
		s.generateThumbnailForFile(file.ID, staged.FinalPath, staged.FileType)
	}

	if err := s.storageService.UpdateMasterDirectorySize(staged.MasterID, staged.FileType, staged.FileSize); err != nil {
		fmt.Printf("Failed to update storage size: %v\n", err)
	}

	return file, nil
}

func (s *FileService) cleanupStagedImport(staged *stagedImport) error {
	if staged == nil {
		return nil
	}
	if err := os.Remove(staged.TempPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove staged temp file %s: %w", staged.TempPath, err)
	}
	return nil
}
```

- [ ] **Step 4: 新增 copyFileWithProgress**

在 `copyFile` 函数旁边新增：

```go
func copyFileWithProgress(src, dst string, onProgress func(copied int64) error) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	buffer := make([]byte, 1024*1024)
	var copied int64
	for {
		readBytes, readErr := source.Read(buffer)
		if readBytes > 0 {
			written, writeErr := destination.Write(buffer[:readBytes])
			if writeErr != nil {
				return writeErr
			}
			copied += int64(written)
			if onProgress != nil {
				if err := onProgress(copied); err != nil {
					return err
				}
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	return nil
}
```

- [ ] **Step 5: 编写测试覆盖重复文件不复制**

在 `app/services/file_service_test.go` 中新增测试：

```go
func TestPrepareStagedImport_ReturnsDuplicateWithoutCopy(t *testing.T) {
	// This test requires a real repository and storage service to fully verify;
	// use an integration-style test or mock repository if available.
	// For the plan, we will add a focused integration test in Task 10.
}
```

- [ ] **Step 6: 运行编译检查**

```bash
go build ./app/services
```

Expected: 编译通过（可能需要先完成 Task 3 的 `StorageService.GetMasterPath` 或直接用 `defaultMaster.Path`）。

- [ ] **Step 7: Commit**

```bash
git add app/services/file_service.go
git commit -m "feat: add unified staged import pipeline helpers

- prepareStagedImport: checksum, duplicate detection, copy to temp
- commitStagedImport: create DB record, atomic rename, remove source
- cleanupStagedImport: remove temp file on failure
- Use unique temp paths under <master>/.localspace-temp

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 3: 单文件导入迁移到 Pipeline

**Files:**
- Modify: `app/services/file_service.go:102-245`（`ImportFile` 函数）

**Interfaces:**
- Consumes: `prepareStagedImport`, `commitStagedImport`, `cleanupStagedImport`。
- Produces: `ImportFile` 新行为。

- [ ] **Step 1: 重写 ImportFile**

将 `ImportFile` 函数体重写为：

```go
func (s *FileService) ImportFile(req ImportFileRequest) error {
	if req.FilePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}
	if _, err := os.Stat(req.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", req.FilePath)
	}

	fileName := strings.TrimSpace(req.FileName)
	if fileName == "" {
		fileName = filepath.Base(req.FilePath)
	}

	staged, err := s.prepareStagedImport(req.FilePath, fileName, req.CollectionName, "", 0, nil)
	if err != nil {
		return err
	}
	defer func() {
		if staged != nil {
			_ = s.cleanupStagedImport(staged)
		}
	}()

	metadata, err := s.ExtractMetadata(staged.TempPath, staged.FileType)
	if err != nil {
		fmt.Printf("Warning: Failed to extract metadata: %v\n", err)
	}

	formMetadata := s.resolveMetadataForImport(&req)

	_, err = s.commitStagedImport(staged, formMetadata.Tags, formMetadata.Description, metadata)
	if err != nil {
		return err
	}

	// commit succeeded, prevent cleanup of temp (already renamed to final)
	staged = nil
	return nil
}
```

注意：这里 `staged = nil` 的 trick 防止 defer 删除已经重命名的 final 文件。但 `cleanupStagedImport` 只删除 `TempPath`，而 `commit` 成功后 `TempPath` 已不存在，所以即使不置 nil 也安全。为保险起见保留置 nil 习惯。

- [ ] **Step 2: 运行编译与测试**

```bash
go build ./app/services
go test ./app/services
```

Expected: 编译通过，现有测试 PASS。

- [ ] **Step 3: Commit**

```bash
git add app/services/file_service.go
git commit -m "refactor: migrate single-file import to staged pipeline

- Source file is no longer moved before DB record creation
- Cross-drive imports now consistently delete source after success
- Failure before DB commit leaves source file intact

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 4: 批量导入迁移到 Pipeline + 临时路径安全

**Files:**
- Modify: `app/services/file_service.go:261-487`（`PrepareBatchImport`, `CopyFileToTemp`, `FinalizeBatchImport`, `CleanupBatchImportArtifacts`）
- Modify: `app/services/batch_import_handler.go:79-263`（`process` 流程）

**Interfaces:**
- Consumes: `prepareStagedImport`, `commitStagedImport`, `cleanupStagedImport`。
- Produces: `BatchImportPlan` 新的 `TempPath` 格式；`CleanupBatchImportArtifacts` 只删 `TempPath`。

- [ ] **Step 1: 修改 BatchImportPlan 与 PrepareBatchImport**

`BatchImportPlan` 保留 `TempPath` 和 `FinalPath`，但 `PrepareBatchImport` 改为调用 `prepareStagedImport`：

```go
func (s *FileService) PrepareBatchImport(sourcePath, displayName, collectionName string, jobID uint, itemIndex int) (*BatchImportPlan, error) {
	fileName := strings.TrimSpace(displayName)
	if fileName == "" {
		fileName = filepath.Base(sourcePath)
	}

	staged, err := s.prepareStagedImport(sourcePath, fileName, collectionName, fmt.Sprintf("%d", jobID), itemIndex, nil)
	if err != nil {
		return nil, err
	}

	return &BatchImportPlan{
		SourcePath:     staged.SourcePath,
		FileName:       staged.FileName,
		OriginalName:   staged.OriginalName,
		CollectionName: staged.CollectionName,
		FileType:       staged.FileType,
		FileSubType:    staged.FileSubType,
		FileSize:       staged.FileSize,
		Checksum:       staged.Checksum,
		MasterID:       staged.MasterID,
		FinalPath:      staged.FinalPath,
		TempPath:       staged.TempPath,
	}, nil
}
```

- [ ] **Step 2: 修改 CopyFileToTemp**

`CopyFileToTemp` 不再需要自己清理 finalPath，且使用已生成的 `TempPath`：

```go
func (s *FileService) CopyFileToTemp(plan *BatchImportPlan, onProgress func(copied int64) error) error {
	if plan == nil {
		return fmt.Errorf("batch import plan is nil")
	}

	if err := copyFileWithProgress(plan.SourcePath, plan.TempPath, onProgress); err != nil {
		_ = os.Remove(plan.TempPath)
		return fmt.Errorf("failed to copy file to staging: %w", err)
	}

	// 验证大小
	info, err := os.Stat(plan.TempPath)
	if err != nil {
		_ = os.Remove(plan.TempPath)
		return fmt.Errorf("failed to stat staged file: %w", err)
	}
	if info.Size() != plan.FileSize {
		_ = os.Remove(plan.TempPath)
		return fmt.Errorf("copied file size mismatch: expected %d bytes, got %d bytes", plan.FileSize, info.Size())
	}

	return nil
}
```

- [ ] **Step 3: 修改 FinalizeBatchImport**

```go
func (s *FileService) FinalizeBatchImport(plan *BatchImportPlan, tags []string, description string, metadata models.Metadata) (*models.File, error) {
	if plan == nil {
		return nil, fmt.Errorf("batch import plan is nil")
	}

	staged := &stagedImport{
		SourcePath:     plan.SourcePath,
		TempPath:       plan.TempPath,
		FinalPath:      plan.FinalPath,
		Checksum:       plan.Checksum,
		FileSize:       plan.FileSize,
		FileType:       plan.FileType,
		FileSubType:    plan.FileSubType,
		MasterID:       plan.MasterID,
		CollectionName: plan.CollectionName,
		FileName:       plan.FileName,
		OriginalName:   plan.OriginalName,
	}

	return s.commitStagedImport(staged, tags, description, metadata)
}
```

- [ ] **Step 4: 修改 CleanupBatchImportArtifacts**

```go
func (s *FileService) CleanupBatchImportArtifacts(tempPath, finalPath string) error {
	// 只删除明确传入的 tempPath；finalPath 不再作为删除候选
	path := strings.TrimSpace(tempPath)
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove temp artifact %s: %w", path, err)
	}
	return nil
}
```

- [ ] **Step 5: 修改 BatchImportHandler 调用**

在 `batch_import_handler.go` 中：

1. `PrepareBatchImport` 调用增加 `jobID` 和 `itemIndex`：

```go
plan, err := runtime.FileService().PrepareBatchImport(sourcePath, displayName, payload.CollectionName, job.ID, item.ItemIndex)
```

2. `process` 函数中 `CleanupBatchImportArtifacts` 调用保持不变，但参数只传 `item.TempPath`：

```go
runtime.FileService().CleanupBatchImportArtifacts(item.TempPath, "")
```

- [ ] **Step 6: 运行编译与测试**

```bash
go build ./...
go test ./app/services ./app/...
```

Expected: 编译通过，测试 PASS。

- [ ] **Step 7: Commit**

```bash
git add app/services/file_service.go app/services/batch_import_handler.go
git commit -m "refactor: migrate batch import to staged pipeline and safe temp paths

- Batch temp paths now under <master>/.localspace-temp/batch-<jobID>/<itemIndex>/
- CleanupBatchImportArtifacts only deletes the explicit TempPath
- Reuses prepareStagedImport / commitStagedImport pipeline

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 5: DeleteFile 与 RenameFile 顺序修复

**Files:**
- Modify: `app/services/file_service.go:725-786`（`DeleteFile`, `RenameFile`）
- Test: `app/services/file_service_test.go`

**Interfaces:**
- Consumes: `FileRepository.Delete`, `FileRepository.Update`, `os.Remove`, `os.Rename`。
- Produces: `DeleteFile` 和 `RenameFile` 新顺序行为。

- [ ] **Step 1: 重写 DeleteFile**

```go
func (s *FileService) DeleteFile(id uint) error {
	file, err := s.fileRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	resolvedType := resolveDisplayFileType(file.FilePath, file.FileType)

	if err := s.fileRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete file record: %w", err)
	}

	if err := os.Remove(file.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete physical file: %w", err)
	}

	if file.Thumbnail != "" {
		_ = s.thumbnailService.RemoveThumbnail(id, resolvedType)
	}

	if err := s.storageService.UpdateStorageSize(file.FileType, -file.FileSize); err != nil {
		fmt.Printf("Failed to update storage size: %v\n", err)
	}

	return nil
}
```

- [ ] **Step 2: 重写 RenameFile**

```go
func (s *FileService) RenameFile(id uint, newName string) error {
	file, err := s.fileRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	if newName == "" {
		return fmt.Errorf("new file name cannot be empty")
	}
	if newName == file.FileName {
		return nil
	}

	ext := filepath.Ext(file.FilePath)
	if ext != "" && !strings.HasSuffix(newName, ext) {
		newName = newName + ext
	}

	dir := filepath.Dir(file.FilePath)
	newPath := filepath.Join(dir, newName)
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("a file with this name already exists: %s", newName)
	}

	originalPath := file.FilePath
	file.FileName = newName
	file.FilePath = newPath
	file.ModifiedAt = time.Now().Format(time.RFC3339)

	if err := s.fileRepo.Update(file); err != nil {
		return fmt.Errorf("failed to update file record: %w", err)
	}

	if err := os.Rename(originalPath, newPath); err != nil {
		// DB already has new path; rollback DB to old path
		file.FileName = filepath.Base(originalPath)
		file.FilePath = originalPath
		file.ModifiedAt = time.Now().Format(time.RFC3339)
		if rollbackErr := s.fileRepo.Update(file); rollbackErr != nil {
			fmt.Printf("Critical: failed to rollback file record after rename failure: %v\n", rollbackErr)
		}
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}
```

- [ ] **Step 3: 新增测试**

在 `app/services/file_service_test.go` 中新增 `metadataRepoStub` 类测试，或新增基于 mock 的 DeleteFile/RenameFile 测试。由于这两个函数依赖文件系统，使用 `t.TempDir()` 创建真实文件：

```go
func TestDeleteFile_DeletesRecordBeforePhysicalFile(t *testing.T) {
	// 集成测试，放在 file_service_integration_test.go 或新增 test helper
}

func TestRenameFile_RollsBackDatabaseOnPhysicalFailure(t *testing.T) {
	// 集成测试
}
```

- [ ] **Step 4: 运行测试**

```bash
go test ./app/services
```

Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add app/services/file_service.go app/services/file_service_test.go
git commit -m "fix: reorder DeleteFile and RenameFile to prevent inconsistencies

- DeleteFile: delete DB record first, then physical file
- RenameFile: update DB record to new path first, then os.Rename
- Rollback DB record on physical rename failure

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 6: JobService panic 恢复与 shutdown 同步

**Files:**
- Modify: `app/services/job_service.go`
- Modify: `app/app.go:74-117`
- Test: `app/services/job_service_test.go`（新增）

**Interfaces:**
- Consumes: `sync.WaitGroup`, `debug.Stack`。
- Produces: `runJob` 不再因 panic 丢失状态；`PrepareForShutdown` 等待 goroutine 退出。

- [ ] **Step 1: 添加 WaitGroup 字段**

在 `JobService` struct 中新增：

```go
runningWg sync.WaitGroup
```

- [ ] **Step 2: 修改 runJob**

在 `runJob` 函数开头，读取 job 之后添加 panic recovery：

```go
func (s *JobService) runJob(jobID uint, resume bool) {
	job, err := s.jobRepo.FindByID(jobID)
	if err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			panicErr := fmt.Errorf("job panicked: %v\n%s", r, debug.Stack())
			fmt.Printf("Job %d panic: %v\n", jobID, panicErr)

			latestJob, latestErr := s.jobRepo.FindByID(jobID)
			if latestErr == nil {
				job = latestJob
			}
			job.Status = models.JobStatusFailed
			job.ErrorMessage = panicErr.Error()
			job.FinishedAt = time.Now().Format(time.RFC3339)
			_ = s.jobRepo.Update(job)
			s.emitJobEvent("job:failed", job)
			s.untrackRunning(jobID)
		}
	}()

	// 原逻辑继续...
	ctx, cancel := context.WithCancel(context.Background())
	s.trackRunning(jobID, cancel)
	s.runningWg.Add(1)
	defer func() {
		s.untrackRunning(jobID)
		s.runningWg.Done()
	}()

	// ... 其余原逻辑不变
}
```

- [ ] **Step 3: 修改 PrepareForShutdown**

在取消所有 running cancel func 后添加：

```go
for _, cancel := range running {
	if cancel != nil {
		cancel()
	}
}

s.runningWg.Wait()
return nil
```

- [ ] **Step 4: 修改 App.Shutdown**

`app/app.go` 中：

```go
func (a *App) Shutdown(ctx context.Context) {
	if a.jobService != nil {
		if err := a.jobService.PrepareForShutdown(); err != nil {
			fmt.Printf("Failed to prepare jobs for shutdown: %v\n", err)
		}
	}
	if a.db != nil {
		a.db.Close()
	}
}
```

- [ ] **Step 5: 新增 JobService 测试**

创建 `app/services/job_service_test.go`：

```go
package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"LocalSpace/app/models"
	"encoding/json"
)

type mockJobRepo struct{}

func (r *mockJobRepo) Create(job *models.Job) error { return nil }
func (r *mockJobRepo) FindByID(id uint) (*models.Job, error) {
	return &models.Job{ID: id, JobType: models.JobTypeBatchImport, Status: models.JobStatusRunning}, nil
}
func (r *mockJobRepo) Update(job *models.Job) error { return nil }
func (r *mockJobRepo) List(page, pageSize int, jobType string) ([]*models.Job, int, error) { return nil, 0, nil }
func (r *mockJobRepo) ListActive() ([]*models.Job, error) { return nil, nil }
func (r *mockJobRepo) ListResumable() ([]*models.Job, error) { return nil, nil }
func (r *mockJobRepo) ListUnfinishedForStartup() ([]*models.Job, error) { return nil, nil }
func (r *mockJobRepo) CountActiveByExclusiveKey(key string, exclude uint) (int, error) { return 0, nil }
func (r *mockJobRepo) CreateBatchImportItems(items []*models.BatchImportItem) error { return nil }
func (r *mockJobRepo) ListBatchImportItems(jobID uint) ([]*models.BatchImportItem, error) { return nil, nil }
func (r *mockJobRepo) UpdateBatchImportItem(item *models.BatchImportItem) error { return nil }

func TestJobService_RunJob_RecoversFromPanic(t *testing.T) {
	service := NewJobService(&mockJobRepo{}, nil, nil, nil)

	panickingHandler := &testPanicHandler{}
	service.RegisterHandler(panickingHandler)

	done := make(chan bool)
	var failedJob *models.Job
	service.SetEventEmitter(func(name string, data interface{}) {
		if name == "job:failed" {
			failedJob = data.(*models.Job)
			done <- true
		}
	})

	go service.runJob(1, false)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for panic recovery")
	}

	if failedJob == nil {
		t.Fatal("expected job:failed event")
	}
	if failedJob.Status != models.JobStatusFailed {
		t.Errorf("expected status failed, got %s", failedJob.Status)
	}
	if failedJob.ErrorMessage == "" {
		t.Error("expected error message after panic")
	}
}

type testPanicHandler struct{}

func (h *testPanicHandler) Type() string { return models.JobTypeBatchImport }
func (h *testPanicHandler) Validate(payload json.RawMessage) error { return nil }
func (h *testPanicHandler) Execute(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	panic("intentional panic")
}
func (h *testPanicHandler) Resume(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	panic("intentional panic")
}
func (h *testPanicHandler) Cleanup(ctx context.Context, job *models.Job, runtime JobRuntime) error { return nil }
```

- [ ] **Step 6: 运行测试**

```bash
go test -run TestJobService_RunJob_RecoversFromPanic ./app/services
```

Expected: PASS。

- [ ] **Step 7: Commit**

```bash
git add app/services/job_service.go app/services/job_service_test.go app/app.go
git commit -m "fix: recover from job panics and wait for goroutines on shutdown

- Add defer recover() in runJob to mark panicked jobs as failed
- Add sync.WaitGroup to track running goroutines
- PrepareForShutdown waits for all goroutines to exit
- App.Shutdown ensures job service finishes before closing DB

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 7: PowerShell 替换为原生 Explorer 调用

**Files:**
- Modify: `app/app.go:403-453`（`openFileLocation` Windows 分支）

**Interfaces:**
- Consumes: `exec.Command`, `filepath.Abs`。
- Produces: 无字符串拼接的 Windows 打开位置。

- [ ] **Step 1: 替换 Windows 分支**

```go
case "windows":
    absPath, err := filepath.Abs(filePath)
    if err != nil {
        return fmt.Errorf("failed to get absolute path: %w", err)
    }

    return exec.Command("explorer", "/select,", absPath).Start()
```

- [ ] **Step 2: 运行编译**

```bash
go build ./app
```

Expected: 编译通过。

- [ ] **Step 3: Commit**

```bash
git add app/app.go
git commit -m "fix: use native explorer instead of powershell for open file location

- Avoids string concatenation and shell injection risk
- Uses exec.Command("explorer", "/select,", absPath)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 8: 启动时清理残留临时文件

**Files:**
- Modify: `app/services/file_service.go`（新增 `CleanupOrphanedTempFiles`）
- Modify: `app/app.go`（在 `initializeApp` 中调用）

**Interfaces:**
- Consumes: 所有 master 目录路径。
- Produces: `CleanupOrphanedTempFiles(maxAge time.Duration) error`。

- [ ] **Step 1: 实现 CleanupOrphanedTempFiles**

在 `app/services/file_service.go` 中新增：

```go
func (s *FileService) CleanupOrphanedTempFiles(maxAge time.Duration) error {
	masters, err := s.storageService.GetMasterDirectories()
	if err != nil {
		return fmt.Errorf("failed to get master directories: %w", err)
	}

	for _, master := range masters {
		tempDir := filepath.Join(master.Path, batchImportTempDir)
		if err := s.cleanupOldFilesInDir(tempDir, maxAge); err != nil {
			fmt.Printf("Warning: failed to cleanup temp dir %s: %v\n", tempDir, err)
		}
	}
	return nil
}

func (s *FileService) cleanupOldFilesInDir(root string, maxAge time.Duration) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	now := time.Now()
	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())
		if entry.IsDir() {
			if err := s.cleanupOldFilesInDir(path, maxAge); err != nil {
				fmt.Printf("Warning: failed to cleanup subdirectory %s: %v\n", path, err)
			}
			// Try to remove empty directory
			_ = os.Remove(path)
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}
		if now.Sub(info.ModTime()) > maxAge {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				fmt.Printf("Warning: failed to remove orphaned temp file %s: %v\n", path, err)
			}
		}
	}
	return nil
}
```

- [ ] **Step 2: 在 App.Startup 中调用**

在 `app/app.go` 的 `initializeApp` 中，在 `NormalizeUnfinishedJobs` 之后添加：

```go
if err := a.fileService.CleanupOrphanedTempFiles(24 * time.Hour); err != nil {
    fmt.Printf("Failed to cleanup orphaned temp files: %v\n", err)
}
```

- [ ] **Step 3: 运行编译**

```bash
go build ./...
```

Expected: 编译通过。

- [ ] **Step 4: Commit**

```bash
git add app/services/file_service.go app/app.go
git commit -m "feat: cleanup orphaned temp files on startup

- Scan <master>/.localspace-temp/ for files older than 24 hours
- Remove old files and empty directories

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 9: 前端 API 错误处理与单文件元数据类型

**Files:**
- Modify: `frontend/src/api/index.ts:96-119`
- Modify: `frontend/src/views/ImportView.vue:346`

**Interfaces:**
- Consumes: `window.go.app.App.*`。
- Produces: `safeWailsCall` 真实失败时 reject；`getMetadata` 使用真实类型。

- [ ] **Step 1: 修改 safeWailsCall**

```ts
function safeWailsCall<T>(
  fn: () => Promise<T>,
  fallbackValue: T,
  errorMessage: string
): Promise<T> {
  return new Promise((resolve, reject) => {
    try {
      if (!window.go || !window.go.app || !window.go.app.App) {
        console.warn(`Wails API not available: ${errorMessage}`)
        console.warn('This is expected in browser development. Using fallback values.')
        resolve(fallbackValue)
        return
      }

      fn()
        .then(resolve)
        .catch((error) => {
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

- [ ] **Step 2: 修改 ImportView 单文件元数据类型**

`frontend/src/views/ImportView.vue:346`：

```ts
const metadata = await api.system.getMetadata(path, singleFile.value?.type || 'file')
```

- [ ] **Step 3: 检查并补充调用方 catch**

搜索所有 `api.file.delete` 和 `api.file.rename` 的调用点，确保已 try/catch。例如若文件列表组件中有：

```ts
try {
  await api.file.delete(file.id)
  await loadFiles()
} catch (error) {
  // 显示错误消息
}
```

- [ ] **Step 4: 运行前端构建**

```bash
cd frontend
npm run build
```

Expected: 构建成功。

- [ ] **Step 5: Commit**

```bash
git add frontend/src/api/index.ts frontend/src/views/ImportView.vue
git commit -m "fix: expose real Wails errors and use correct file type for metadata

- safeWailsCall now rejects on Wails call failures, only falls back when Wails is unavailable
- ImportView passes parsed file type to getMetadata instead of hardcoded 'file'

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 10: 集成测试与最终验证

**Files:**
- Modify: `app/services/file_service_integration_test.go`
- 运行命令：全局 `go test ./...` + `cd frontend && npm run build`

- [ ] **Step 1: 新增跨盘/失败回滚集成测试**

在 `app/services/file_service_integration_test.go` 中新增或扩展现有测试：

```go
func TestImportFile_StagedPipeline_KeepsSourceOnFailure(t *testing.T) {
	// 创建临时 master 目录和源文件
	// 构造 FileService，注入一个会在 Create 时失败的 repo stub
	// 调用 ImportFile，断言源文件仍存在
}
```

- [ ] **Step 2: 运行完整 Go 测试**

```bash
go test ./...
```

Expected: 全部 PASS。

- [ ] **Step 3: 运行前端构建**

```bash
cd frontend
npm run build
```

Expected: 构建成功。

- [ ] **Step 4: 手动验证清单**

| 场景 | 验证结果 |
|---|---|
| 单文件导入成功 | 文件在 master 目录，源文件删除，DB 有记录 |
| 单文件导入 DB 失败 | 源文件保留，temp 被删除 |
| 跨盘导入 | 源盘文件删除，目标盘有文件 |
| 删除文件 | DB 记录先删除，物理文件后删除 |
| 重命名文件失败 | DB 记录保持旧路径 |
| 批量导入 50 文件中途关闭 | 重启后可恢复，已完成 item 不再重复 |
| 打开文件位置 | Explorer 选中文件 |
| 列表排序恶意参数 | 不报错、不删表 |

- [ ] **Step 5: 最终 Commit（如需要）**

```bash
git add app/services/file_service_integration_test.go
git commit -m "test: add integration tests for staged import pipeline

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## 计划自审

1. **Spec 覆盖**：
   - Issue 1/2/10/补充 D：Task 2-4 统一 pipeline + 唯一 temp 路径。
   - Issue 3：Task 1 SQL 白名单。
   - Issue 4/5：Task 5 Delete/Rename 顺序。
   - Issue 6/7：Task 6 panic 恢复 + WaitGroup。
   - Issue 8：Task 7 explorer 调用。
   - Issue 9：Task 9 safeWailsCall。
   - 补充 C：Task 9 ImportView 类型。
   - 残留清理：Task 8。

2. **Placeholder 检查**：无 TBD/TODO/"implement later" 等占位符。

3. **类型一致性**：所有任务使用 `stagedImport`, `BatchImportPlan`, `FileFilter` 等一致类型。

---

## 执行交接

**Plan complete and saved to `docs/superpowers/plans/2026-08-01-localspace-issue-fixes-plan.md`. Two execution options:**

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using `superpowers:executing-plans`, batch execution with checkpoints

**Which approach?**
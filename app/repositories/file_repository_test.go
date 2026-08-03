package repositories

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/glebarez/sqlite"

	"LocalSpace/app/database"
	"LocalSpace/app/models"
)

// setupTestDB creates a test database
func setupTestDB(t *testing.T) *sql.DB {
	// Create temp directory for test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Create test database
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	return db
}

// cleanupTestDB closes the database connection
func cleanupTestDB(db *sql.DB) {
	if db != nil {
		db.Close()
	}
}

func TestFileRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := NewFileRepository(NewSQLiteDBWrapper(db))

	// Create test file
	file := &models.File{
		FileName:     "test.mp4",
		OriginalName: "original_test.mp4",
		CollectionName: "电视剧A",
		FilePath:     "/test/path/test.mp4",
		FileType:     "video",
		FileSubType:  "mp4",
		FileSize:     1024000,
		Tags:         []string{"test", "video"},
		Description:  "Test video file",
		Metadata:     models.Metadata{Width: 1920, Height: 1080},
		Thumbnail:    "",
		Checksum:     "abc123",
		IsDeleted:    false,
		DeletedAt:    "",
	}

	// Test create
	err := repo.Create(file)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Verify ID was set
	if file.ID == 0 {
		t.Error("File ID was not set after create")
	}
}

func TestFileRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := NewFileRepository(NewSQLiteDBWrapper(db))

	// Create test file
	file := &models.File{
		FileName:     "test.pdf",
		OriginalName: "original_test.pdf",
		CollectionName: "项目资料",
		FilePath:     "/test/path/test.pdf",
		FileType:     "document",
		FileSubType:  "pdf",
		FileSize:     512000,
		Tags:         []string{"test", "document"},
		Description:  "Test document file",
		Metadata:     models.Metadata{PageCount: 10},
		Thumbnail:    "",
		Checksum:     "def456",
		IsDeleted:    false,
		DeletedAt:    "",
	}

	err := repo.Create(file)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Test find by ID
	found, err := repo.FindByID(file.ID)
	if err != nil {
		t.Fatalf("Failed to find file: %v", err)
	}

	// Verify fields
	if found.FileName != file.FileName {
		t.Errorf("Expected FileName %s, got %s", file.FileName, found.FileName)
	}
	if found.OriginalName != file.OriginalName {
		t.Errorf("Expected OriginalName %s, got %s", file.OriginalName, found.OriginalName)
	}
	if found.CollectionName != file.CollectionName {
		t.Errorf("Expected CollectionName %s, got %s", file.CollectionName, found.CollectionName)
	}
	if found.FileType != file.FileType {
		t.Errorf("Expected FileType %s, got %s", file.FileType, found.FileType)
	}
	if len(found.Tags) != len(file.Tags) {
		t.Errorf("Expected %d tags, got %d", len(file.Tags), len(found.Tags))
	}
}

func TestFileRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := NewFileRepository(NewSQLiteDBWrapper(db))

	// Create test files
	files := []*models.File{
		{
			FileName:     "video1.mp4",
			OriginalName: "video1.mp4",
			CollectionName: "剧集A",
			FilePath:     "/test/video1.mp4",
			FileType:     "video",
			FileSubType:  "mp4",
			FileSize:     1024000,
			Tags:         []string{"test"},
			Description:  "Test video 1",
		},
		{
			FileName:     "video2.mp4",
			OriginalName: "video2.mp4",
			CollectionName: "剧集A",
			FilePath:     "/test/video2.mp4",
			FileType:     "video",
			FileSubType:  "mp4",
			FileSize:     2048000,
			Tags:         []string{"test"},
			Description:  "Test video 2",
		},
		{
			FileName:     "doc.pdf",
			OriginalName: "doc.pdf",
			CollectionName: "项目A",
			FilePath:     "/test/doc.pdf",
			FileType:     "document",
			FileSubType:  "pdf",
			FileSize:     512000,
			Tags:         []string{"test"},
			Description:  "Test document",
		},
	}

	for _, file := range files {
		if err := repo.Create(file); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Test list all files
	allFiles, err := repo.List(FileFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}
	if len(allFiles) != 3 {
		t.Errorf("Expected 3 files, got %d", len(allFiles))
	}

	// Test list filtered by file type
	videoFiles, err := repo.List(FileFilter{Page: 1, PageSize: 10, FileType: "video"})
	if err != nil {
		t.Fatalf("Failed to list video files: %v", err)
	}
	if len(videoFiles) != 2 {
		t.Errorf("Expected 2 video files, got %d", len(videoFiles))
	}
}

func TestFileRepository_Search(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := NewFileRepository(NewSQLiteDBWrapper(db))

	// Create test files
	files := []*models.File{
		{
			FileName:     "tutorial_video.mp4",
			OriginalName: "tutorial_video.mp4",
			CollectionName: "课程",
			FilePath:     "/test/tutorial.mp4",
			FileType:     "video",
			FileSubType:  "mp4",
			FileSize:     1024000,
			Tags:         []string{"tutorial", "education"},
			Description:  "Learn programming tutorial",
		},
		{
			FileName:     "entertainment_video.mp4",
			OriginalName: "entertainment_video.mp4",
			CollectionName: "娱乐",
			FilePath:     "/test/entertainment.mp4",
			FileType:     "video",
			FileSubType:  "mp4",
			FileSize:     2048000,
			Tags:         []string{"fun", "entertainment"},
			Description:  "Funny cat video",
		},
	}

	for _, file := range files {
		if err := repo.Create(file); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Test search by file name
	results, err := repo.Search("tutorial")
	if err != nil {
		t.Fatalf("Failed to search files: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'tutorial', got %d", len(results))
	}

	// Test search by description
	results, err = repo.Search("programming")
	if err != nil {
		t.Fatalf("Failed to search files: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'programming', got %d", len(results))
	}

	// Test search by tags
	results, err = repo.Search("education")
	if err != nil {
		t.Fatalf("Failed to search files: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'education', got %d", len(results))
	}

	results, err = repo.Search("课程")
	if err != nil {
		t.Fatalf("Failed to search files by collection: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result for collection search, got %d", len(results))
	}
}

func TestFileRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := NewFileRepository(NewSQLiteDBWrapper(db))

	// Create test file
	file := &models.File{
		FileName:     "to_delete.mp4",
		OriginalName: "to_delete.mp4",
		CollectionName: "临时合集",
		FilePath:     "/test/to_delete.mp4",
		FileType:     "video",
		FileSubType:  "mp4",
		FileSize:     1024000,
		Tags:         []string{"test"},
		Description:  "File to delete",
	}

	err := repo.Create(file)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Test delete
	err = repo.Delete(file.ID)
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// Verify file is deleted
	_, err = repo.FindByID(file.ID)
	if err == nil {
		t.Error("Expected error when finding deleted file, got nil")
	}
}

func TestFileRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := NewFileRepository(NewSQLiteDBWrapper(db))

	// Create test file
	file := &models.File{
		FileName:     "to_update.mp4",
		OriginalName: "to_update.mp4",
		CollectionName: "旧合集",
		FilePath:     "/test/to_update.mp4",
		FileType:     "video",
		FileSubType:  "mp4",
		FileSize:     1024000,
		Tags:         []string{"old"},
		Description:  "Old description",
	}

	err := repo.Create(file)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Update file
	file.Tags = []string{"new", "updated"}
	file.Description = "New description"
	file.CollectionName = "新合集"
	cid := uint(1)
	file.CollectionID = &cid
	err = repo.Update(file)
	if err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}

	// Verify update
	updated, err := repo.FindByID(file.ID)
	if err != nil {
		t.Fatalf("Failed to find updated file: %v", err)
	}

	if len(updated.Tags) != 2 {
		t.Errorf("Expected 2 tags after update, got %d", len(updated.Tags))
	}
	if updated.Description != "New description" {
		t.Errorf("Expected description 'New description', got '%s'", updated.Description)
	}
	if updated.CollectionName != "旧合集" {
		t.Errorf("Expected collection name to be preserved '旧合集', got '%s'", updated.CollectionName)
	}
	if updated.CollectionID == nil || *updated.CollectionID != 1 {
		t.Errorf("Expected collection id 1, got %v", updated.CollectionID)
	}
}

func TestFileRepository_ExistsByPath(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := NewFileRepository(NewSQLiteDBWrapper(db))

	// Create test file
	file := &models.File{
		FileName:     "check_exist.mp4",
		OriginalName: "check_exist.mp4",
		CollectionName: "测试合集",
		FilePath:     "/test/check_exist.mp4",
		FileType:     "video",
		FileSubType:  "mp4",
		FileSize:     1024000,
		Tags:         []string{"test"},
		Description:  "Test file",
	}

	err := repo.Create(file)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Test existing path
	exists, err := repo.ExistsByPath("/test/check_exist.mp4")
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Expected file to exist, got false")
	}

	// Test non-existing path
	exists, err = repo.ExistsByPath("/test/nonexistent.mp4")
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if exists {
		t.Error("Expected file to not exist, got true")
	}
}

func TestFileRepository_FindByChecksum(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewFileRepository(&SQLiteDBWrapper{db: db})

	// Create a file with specific checksum
	testChecksum := "abc123def456"
	file := &models.File{
		FileName:     "test_file.txt",
		OriginalName: "test_file.txt",
		FilePath:     "/test/test_file.txt",
		FileType:     "document",
		FileSubType:  "txt",
		FileSize:     1024,
		Checksum:     testChecksum,
		Tags:         []string{"test"},
		Description:  "Test file with checksum",
	}

	err := repo.Create(file)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Test finding by checksum
	foundFile, err := repo.FindByChecksum(testChecksum)
	if err != nil {
		t.Fatalf("Failed to find file by checksum: %v", err)
	}
	if foundFile == nil {
		t.Fatal("Expected to find file, got nil")
	}
	if foundFile.ID != file.ID {
		t.Errorf("Expected file ID %d, got %d", file.ID, foundFile.ID)
	}
	if foundFile.Checksum != testChecksum {
		t.Errorf("Expected checksum %s, got %s", testChecksum, foundFile.Checksum)
	}

	// Test finding by non-existent checksum
	notFoundFile, err := repo.FindByChecksum("nonexistent")
	if err != nil {
		t.Fatalf("Failed to query by checksum: %v", err)
	}
	if notFoundFile != nil {
		t.Error("Expected nil for non-existent checksum, got file")
	}
}

func TestFileRepository_CheckDuplicateByChecksum(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewFileRepository(&SQLiteDBWrapper{db: db})

	// Test with no files initially
	isDuplicate, duplicateID, err := repo.CheckDuplicateByChecksum("test123")
	if err != nil {
		t.Fatalf("Failed to check duplicate: %v", err)
	}
	if isDuplicate {
		t.Error("Expected no duplicate for empty database")
	}
	if duplicateID != 0 {
		t.Errorf("Expected duplicate ID 0, got %d", duplicateID)
	}

	// Create a file with specific checksum
	testChecksum := "xyz789"
	file := &models.File{
		FileName:     "original_file.txt",
		OriginalName: "original_file.txt",
		CollectionName: "文档合集",
		FilePath:     "/test/original.txt",
		FileType:     "document",
		FileSubType:  "txt",
		FileSize:     1024,
		Checksum:     testChecksum,
		Tags:         []string{"original"},
		Description:  "Original file",
	}

	err = repo.Create(file)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Test checking for duplicate with same checksum
	isDuplicate, duplicateID, err = repo.CheckDuplicateByChecksum(testChecksum)
	if err != nil {
		t.Fatalf("Failed to check duplicate: %v", err)
	}
	if !isDuplicate {
		t.Error("Expected duplicate to be found")
	}
	if duplicateID != file.ID {
		t.Errorf("Expected duplicate ID %d, got %d", file.ID, duplicateID)
	}

	// Test checking for different checksum
	isDuplicate, duplicateID, err = repo.CheckDuplicateByChecksum("different123")
	if err != nil {
		t.Fatalf("Failed to check duplicate: %v", err)
	}
	if isDuplicate {
		t.Error("Expected no duplicate for different checksum")
	}
	if duplicateID != 0 {
		t.Errorf("Expected duplicate ID 0, got %d", duplicateID)
	}
}

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

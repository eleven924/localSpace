package services

import (
	"os"
	"path/filepath"
	"testing"

	"LocalSpace/app/database"
	"LocalSpace/app/repositories"
	"LocalSpace/app/utils"
)

func TestImportFileRequestWithAgentFields(t *testing.T) {
	// Test ImportFileRequest with Keywords field
	req := ImportFileRequest{
		FilePath:    "/tmp/test.txt",
		FileName:    "test.txt",
		Keywords:    "test document",
		Tags:        []string{"test"},
		Description: "A test file",
	}

	// Verify all fields are set correctly
	if req.Keywords != "test document" {
		t.Errorf("Expected keywords to be 'test document', got '%s'", req.Keywords)
	}

	if len(req.Tags) != 1 || req.Tags[0] != "test" {
		t.Errorf("Expected tags to be ['test'], got %v", req.Tags)
	}

	if req.Description != "A test file" {
		t.Errorf("Expected description to be 'A test file', got '%s'", req.Description)
	}

	// Test empty defaults
	emptyReq := ImportFileRequest{
		FilePath: "/tmp/test.txt",
		FileName: "test.txt",
	}

	if emptyReq.Keywords != "" {
		t.Errorf("Expected default Keywords to be empty string, got '%s'", emptyReq.Keywords)
	}

	if len(emptyReq.Tags) != 0 {
		t.Errorf("Expected default Tags to be empty, got %v", emptyReq.Tags)
	}

	if emptyReq.Description != "" {
		t.Errorf("Expected default Description to be empty string, got '%s'", emptyReq.Description)
	}
}

func TestFileServiceStructure(t *testing.T) {
	// Test that FileService can be created with agent service integration
	// This test verifies the structure without requiring full dependencies

	// Verify ImportFileRequest structure supports agent integration
	req := ImportFileRequest{
		FilePath:    "/path/to/file.mp4",
		FileName:    "movie.mp4",
		Keywords:    "action thriller",
		Tags:        []string{"entertainment"},
		Description: "An action thriller movie",
	}

	// Test that all required fields are present
	requiredFields := map[string]string{
		"FilePath":    req.FilePath,
		"FileName":    req.FileName,
		"Keywords":    req.Keywords,
		"Description": req.Description,
	}

	for field, value := range requiredFields {
		if value == "" {
			t.Errorf("Required field '%s' is empty", field)
		}
	}

	if len(req.Tags) == 0 {
		t.Error("Tags field should not be empty")
	}
}

func TestFileServiceResolveMetadataForImport_UsesOnlyFormMetadata(t *testing.T) {
	service := &FileService{}

	metadata := service.resolveMetadataForImport(&ImportFileRequest{
		FileName:    "movie.mp4",
		Keywords:    "action thriller",
		Tags:        []string{"entertainment"},
		Description: "  manual description  ",
	})
	if len(metadata.Tags) != 1 || metadata.Tags[0] != "entertainment" {
		t.Fatalf("expected form tags to be preserved, got %v", metadata.Tags)
	}
	if metadata.Description != "manual description" {
		t.Fatalf("expected trimmed form description, got %q", metadata.Description)
	}
}

func TestFileServiceResolveMetadataForImport_PreservesExplicitUserMetadata(t *testing.T) {
	service := &FileService{}

	metadata := service.resolveMetadataForImport(&ImportFileRequest{
		FileName:    "movie.mp4",
		Tags:        []string{"用户标签"},
		Description: "用户描述",
	})

	if len(metadata.Tags) != 1 || metadata.Tags[0] != "用户标签" {
		t.Fatalf("expected explicit user tags to be imported, got %v", metadata.Tags)
	}
	if metadata.Description != "用户描述" {
		t.Fatalf("expected explicit user description to be imported, got %q", metadata.Description)
	}
}

func TestImportFile_StagedPipeline_MovesFileAndDeletesSource(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	configRepo := repositories.NewConfigRepository(repositories.NewSQLiteDBWrapper(db))
	storageService := NewStorageService(configRepo)
	fileRepo := repositories.NewFileRepository(repositories.NewSQLiteDBWrapper(db))
	aiService := NewAIService(configRepo)
	thumbnailService := NewThumbnailService(filepath.Join(tempDir, "thumbnails"))

	collectionRepo := repositories.NewCollectionRepository(repositories.NewSQLiteDBWrapper(db))
	fileService := NewFileService(fileRepo, collectionRepo, storageService, aiService, thumbnailService)

	masterDir := filepath.Join(tempDir, "master")
	if err := os.MkdirAll(masterDir, 0755); err != nil {
		t.Fatalf("failed to create master directory: %v", err)
	}
	if err := storageService.AddMasterDirectory(masterDir, 0); err != nil {
		t.Fatalf("failed to add master directory: %v", err)
	}

	sourcePath := filepath.Join(tempDir, "source.txt")
	content := "hello world"
	if err := writeTestFile(sourcePath, content); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	expectedChecksum, err := utils.CalculateFileChecksum(sourcePath)
	if err != nil {
		t.Fatalf("failed to calculate expected checksum: %v", err)
	}

	if err := fileService.ImportFile(ImportFileRequest{
		FilePath: sourcePath,
		FileName: "imported.txt",
		Tags:     []string{"test"},
	}); err != nil {
		t.Fatalf("import failed: %v", err)
	}

	if _, err := fileService.fileRepo.FindByChecksum(expectedChecksum); err != nil {
		t.Fatalf("failed to find imported file by checksum: %v", err)
	}

	files, err := fileService.fileRepo.List(repositories.FileFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("failed to list files: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file in db, got %d", len(files))
	}
	imported := files[0]
	if imported.FileName != "imported.txt" {
		t.Errorf("expected file name imported.txt, got %s", imported.FileName)
	}
	if imported.Checksum != expectedChecksum {
		t.Errorf("expected checksum %s, got %s", expectedChecksum, imported.Checksum)
	}

	if _, err := fileService.fileRepo.ExistsByPath(imported.FilePath); err != nil {
		t.Errorf("failed to verify final path: %v", err)
	}

	if _, err := os.ReadFile(sourcePath); err == nil {
		t.Error("expected source file to be deleted after successful import")
	}
}

func writeTestFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}


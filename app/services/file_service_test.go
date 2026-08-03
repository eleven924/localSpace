package services

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"LocalSpace/app/database"
	"LocalSpace/app/models"
	"LocalSpace/app/repositories"
)

func TestImportFileRequestKeywords(t *testing.T) {
	cid := uint(3)
	req := ImportFileRequest{
		FilePath:     "/test/path/file.txt",
		FileName:     "file.txt",
		Description:  "test description",
		Tags:         []string{"tag1", "tag2"},
		Keywords:     "test keywords",
		CollectionID: &cid,
	}

	if req.Keywords != "test keywords" {
		t.Errorf("Expected Keywords to be 'test keywords', got '%s'", req.Keywords)
	}
	if req.CollectionID == nil || *req.CollectionID != 3 {
		t.Errorf("Expected CollectionID to be 3, got %v", req.CollectionID)
	}

	// Test default empty value
	emptyReq := ImportFileRequest{
		FilePath: "/test/path/file.txt",
		FileName: "file.txt",
	}

	if emptyReq.Keywords != "" {
		t.Errorf("Expected default Keywords to be empty string, got '%s'", emptyReq.Keywords)
	}
	if emptyReq.CollectionID != nil {
		t.Errorf("Expected default CollectionID to be nil, got %v", emptyReq.CollectionID)
	}
}

func TestNormalizeMetadataTags_TrimsDedupesAndDropsEmpty(t *testing.T) {
	service := &FileService{}

	got := service.normalizeMetadataTags([]string{" 工作 ", "", "重要", "工作", "  ", "归档 "})
	want := []string{"工作", "重要", "归档"}

	if len(got) != len(want) {
		t.Fatalf("expected %d tags, got %d (%v)", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected tag %d to be %q, got %q", i, want[i], got[i])
		}
	}
}

func TestNormalizeMetadataDescription_TrimsWhitespace(t *testing.T) {
	service := &FileService{}

	got := service.normalizeMetadataDescription("  这是描述  ")
	if got != "这是描述" {
		t.Fatalf("expected trimmed description, got %q", got)
	}
}

type metadataRepoStub struct {
	findByIDCalledWith uint
	findFile           *models.File
	findErr            error
	updateErr          error
}

func (r *metadataRepoStub) FindByID(id uint) (*models.File, error) {
	r.findByIDCalledWith = id
	if r.findErr != nil {
		return nil, r.findErr
	}
	if r.findFile != nil {
		copy := *r.findFile
		copy.ID = id
		return &copy, nil
	}
	return &models.File{ID: id}, nil
}

func (r *metadataRepoStub) UpdateMetadata(id uint, tags []string, description string) error {
	return r.updateErr
}

func TestUpdateFileMetadata_NormalizesInputBeforePersisting(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	fileRepo := repositories.NewFileRepository(repositories.NewSQLiteDBWrapper(db))
	service := &FileService{metadataRepo: fileRepo, fileRepo: fileRepo}

	file := &models.File{
		FileName:     "test.txt",
		OriginalName: "test.txt",
		FilePath:     filepath.Join(tempDir, "test.txt"),
		FileType:     "document",
		FileSubType:  "txt",
		FileSize:     5,
		Checksum:     "abc",
	}
	if err := fileRepo.Create(file); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	err = service.UpdateFileMetadata(file.ID, []string{" 工作 ", "工作", "重要", ""}, "  新描述  ", nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	updated, err := fileRepo.FindByID(file.ID)
	if err != nil {
		t.Fatalf("failed to find updated file: %v", err)
	}

	wantTags := []string{"工作", "重要"}
	if len(updated.Tags) != len(wantTags) {
		t.Fatalf("expected %d tags, got %d (%v)", len(wantTags), len(updated.Tags), updated.Tags)
	}
	for i := range wantTags {
		if updated.Tags[i] != wantTags[i] {
			t.Fatalf("expected tag %d to be %q, got %q", i, wantTags[i], updated.Tags[i])
		}
	}

	if updated.Description != "新描述" {
		t.Fatalf("expected trimmed description, got %q", updated.Description)
	}
}

func TestUpdateFileMetadata_PropagatesRepositoryErrors(t *testing.T) {
	repo := &metadataRepoStub{findErr: errors.New("boom")}
	service := &FileService{metadataRepo: repo}

	err := service.UpdateFileMetadata(9, []string{"tag"}, "desc", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got, want := err.Error(), "failed to get file: boom"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRenameFile_UpdatesDatabaseBeforeRenaming(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	repo := repositories.NewFileRepository(repositories.NewSQLiteDBWrapper(db))

	originalPath := filepath.Join(tempDir, "old.txt")
	if err := os.WriteFile(originalPath, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	file := &models.File{
		FileName:     "old.txt",
		OriginalName: "old.txt",
		FilePath:     originalPath,
		FileType:     "document",
		FileSubType:  "txt",
		FileSize:     5,
		Checksum:     "abc",
	}
	if err := repo.Create(file); err != nil {
		t.Fatalf("failed to create file record: %v", err)
	}

	service := &FileService{fileRepo: repo}

	if err := service.RenameFile(file.ID, "new.txt"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, err := repo.FindByID(file.ID)
	if err != nil {
		t.Fatalf("failed to find updated file: %v", err)
	}
	if updated.FileName != "new.txt" {
		t.Errorf("expected file name new.txt, got %s", updated.FileName)
	}
	if !strings.Contains(updated.FilePath, "new.txt") {
		t.Errorf("expected updated path to contain new.txt, got %s", updated.FilePath)
	}

	if _, err := os.Stat(updated.FilePath); err != nil {
		t.Errorf("expected renamed file to exist at %s: %v", updated.FilePath, err)
	}
}

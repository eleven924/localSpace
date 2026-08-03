package services

import (
	"os"
	"path/filepath"
	"testing"

	"LocalSpace/app/database"
	"LocalSpace/app/models"
	"LocalSpace/app/repositories"
	_ "github.com/glebarez/sqlite"
)

func TestCollectionService_Remove_BlockedWhenReferenced(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer db.Close()

	wrapper := repositories.NewSQLiteDBWrapper(db)
	colRepo := repositories.NewCollectionRepository(wrapper)
	fileRepo := repositories.NewFileRepository(wrapper)
	svc := NewCollectionService(colRepo, fileRepo, nil)

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

	if err := svc.RemoveCollection(id); err == nil {
		t.Fatal("expected error removing referenced collection")
	}
}

func TestCollectionService_BatchUpdateFilesCollectionMovesFileOnDisk(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer db.Close()

	wrapper := repositories.NewSQLiteDBWrapper(db)
	colRepo := repositories.NewCollectionRepository(wrapper)
	fileRepo := repositories.NewFileRepository(wrapper)
	configRepo := repositories.NewConfigRepository(wrapper)
	storageService := NewStorageService(configRepo)
	svc := NewCollectionService(colRepo, fileRepo, storageService)

	masterDir := filepath.Join(tempDir, "master")
	if err := os.MkdirAll(masterDir, 0755); err != nil {
		t.Fatalf("failed to create master dir: %v", err)
	}
	if err := storageService.AddMasterDirectory(masterDir, 0); err != nil {
		t.Fatalf("failed to add master directory: %v", err)
	}
	masters, err := storageService.GetMasterDirectories()
	if err != nil {
		t.Fatalf("failed to get master directories: %v", err)
	}
	if len(masters) == 0 {
		t.Fatal("expected one master directory")
	}

	collectionID, err := colRepo.Add("V世代")
	if err != nil {
		t.Fatalf("failed to add collection: %v", err)
	}

	oldPath, err := storageService.GetStoragePathForFileWithMaster(masters[0].ID, "document", "", "move.txt")
	if err != nil {
		t.Fatalf("failed to build old path: %v", err)
	}
	if err := storageService.EnsureStorageDirExists(filepath.Dir(oldPath)); err != nil {
		t.Fatalf("failed to create old dir: %v", err)
	}
	if err := os.WriteFile(oldPath, []byte("move me"), 0644); err != nil {
		t.Fatalf("failed to create old file: %v", err)
	}

	file := &models.File{
		FileName:     "move.txt",
		OriginalName: "move.txt",
		FilePath:     oldPath,
		FileType:     "document",
		FileSubType:  "txt",
		FileSize:     7,
		Tags:         []string{},
		Metadata:     models.Metadata{},
		Checksum:     "move-checksum",
	}
	if err := fileRepo.Create(file); err != nil {
		t.Fatalf("failed to create file record: %v", err)
	}

	result, err := svc.BatchUpdateFilesCollection([]uint{file.ID}, &collectionID)
	if err != nil {
		t.Fatalf("batch move returned error: %v", err)
	}
	if result.SuccessCount != 1 || result.FailedCount != 0 {
		t.Fatalf("expected one successful move, got success=%d failed=%d", result.SuccessCount, result.FailedCount)
	}

	updated, err := fileRepo.FindByID(file.ID)
	if err != nil {
		t.Fatalf("failed to find moved file: %v", err)
	}
	if updated.CollectionID == nil || *updated.CollectionID != collectionID {
		t.Fatalf("expected collection id %d, got %v", collectionID, updated.CollectionID)
	}
	if updated.CollectionName != "V世代" {
		t.Fatalf("expected collection name V世代, got %q", updated.CollectionName)
	}

	wantPath := filepath.Join(masterDir, "document", "V世代", "move.txt")
	if updated.FilePath != wantPath {
		t.Fatalf("expected moved path %q, got %q", wantPath, updated.FilePath)
	}
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("expected moved file at %s: %v", wantPath, err)
	}
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("expected old path to be removed, stat err: %v", err)
	}
}

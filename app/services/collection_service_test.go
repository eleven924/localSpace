package services

import (
	"path/filepath"
	"testing"

	_ "github.com/glebarez/sqlite"
	"LocalSpace/app/database"
	"LocalSpace/app/models"
	"LocalSpace/app/repositories"
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

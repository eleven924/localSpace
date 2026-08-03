package repositories

import (
	"path/filepath"
	"testing"

	"LocalSpace/app/database"
)

func TestCollectionRepository_CRUD(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer db.Close()

	wrapper := NewSQLiteDBWrapper(db)
	repo := NewCollectionRepository(wrapper)

	id, err := repo.Add("Projects")
	if err != nil {
		t.Fatalf("failed to add collection: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero collection id")
	}

	all, err := repo.GetAll()
	if err != nil {
		t.Fatalf("failed to list collections: %v", err)
	}
	if len(all) != 1 || all[0].Name != "Projects" {
		t.Fatalf("expected one collection named Projects, got %v", all)
	}

	found, err := repo.FindByID(id)
	if err != nil {
		t.Fatalf("failed to find collection: %v", err)
	}
	if found.Name != "Projects" {
		t.Fatalf("expected Projects, got %s", found.Name)
	}

	count, err := repo.CountFilesByCollectionID(id)
	if err != nil {
		t.Fatalf("failed to count files: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 files, got %d", count)
	}

	if err := repo.Remove(id); err != nil {
		t.Fatalf("failed to remove collection: %v", err)
	}

	_, err = repo.FindByID(id)
	if err == nil {
		t.Fatal("expected collection not found after remove")
	}
}

func TestCollectionRepository_Add_DuplicateBlocked(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer db.Close()

	wrapper := NewSQLiteDBWrapper(db)
	repo := NewCollectionRepository(wrapper)

	if _, err := repo.Add("Projects"); err != nil {
		t.Fatalf("failed to add collection: %v", err)
	}
	if _, err := repo.Add("Projects"); err == nil {
		t.Fatal("expected duplicate name to be blocked")
	}
}

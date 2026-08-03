package services

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	_ "github.com/glebarez/sqlite"
	"LocalSpace/app/database"
	"LocalSpace/app/models"
	"LocalSpace/app/repositories"
)

func TestCleanupHandler_DeletesByMaxCount(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer db.Close()

	repo := repositories.NewJobRepository(repositories.NewSQLiteDBWrapper(db))
	configService := NewConfigService(repositories.NewConfigRepository(repositories.NewSQLiteDBWrapper(db)))
	jobService := NewJobService(repo, nil, nil, nil, configService)
	jobService.RegisterHandler(NewCleanupHandler())

	for i := 0; i < 3; i++ {
		j := &models.Job{
			JobType: models.JobTypeBatchImport,
			Status:  models.JobStatusCompleted,
			Title:   "test",
		}
		if err := repo.Create(j); err != nil {
			t.Fatalf("failed to create job: %v", err)
		}
	}

	payload, _ := json.Marshal(models.JobRetentionConfig{Enabled: true, MaxCount: 1})
	job := &models.Job{JobType: models.JobTypeCleanup, Payload: payload}
	if err := repo.Create(job); err != nil {
		t.Fatalf("failed to create cleanup job: %v", err)
	}

	handler := NewCleanupHandler()
	err = handler.Execute(context.Background(), job, &jobRuntime{service: jobService, job: job})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	remaining, _, err := repo.List(1, 100, "")
	if err != nil {
		t.Fatalf("failed to list jobs: %v", err)
	}
	if len(remaining) != 2 {
		t.Fatalf("expected 2 jobs remaining (1 old + cleanup), got %d", len(remaining))
	}
}

func TestCleanupHandler_ValidateRequiresCondition(t *testing.T) {
	handler := NewCleanupHandler()
	payload, _ := json.Marshal(models.JobRetentionConfig{Enabled: true})
	if err := handler.Validate(payload); err == nil {
		t.Fatal("expected validation error when no retention condition is set")
	}
}

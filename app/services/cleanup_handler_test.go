package services

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"LocalSpace/app/database"
	"LocalSpace/app/models"
	"LocalSpace/app/repositories"
	_ "github.com/glebarez/sqlite"
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
		createdAt := time.Date(2026, 8, 4, 9+i, 0, 0, 0, time.UTC).Format(time.RFC3339)
		if _, err := db.Exec("UPDATE jobs SET created_at = ?, updated_at = ? WHERE id = ?", createdAt, createdAt, j.ID); err != nil {
			t.Fatalf("failed to set job timestamp: %v", err)
		}
	}

	payload, _ := json.Marshal(models.JobRetentionConfig{Enabled: false, MaxCount: 1})
	job := &models.Job{
		JobType:       models.JobTypeCleanup,
		Status:        models.JobStatusRunning,
		Title:         "cleanup",
		Payload:       payload,
		ProgressTotal: 1,
	}
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

	var keptCompleted *models.Job
	for _, j := range remaining {
		if j.JobType == models.JobTypeBatchImport {
			keptCompleted = j
		}
	}
	if keptCompleted == nil {
		t.Fatal("expected newest completed job to be kept")
	}
	if keptCompleted.CreatedAt != time.Date(2026, 8, 4, 11, 0, 0, 0, time.UTC).Format(time.RFC3339) {
		t.Fatalf("expected newest job to be kept, got created_at %s", keptCompleted.CreatedAt)
	}

	cleanupJob, err := repo.FindByID(job.ID)
	if err != nil {
		t.Fatalf("failed to reload cleanup job: %v", err)
	}
	if cleanupJob.ProgressMessage != "已清理 2 条任务记录" {
		t.Fatalf("expected cleanup progress message to be persisted, got %q", cleanupJob.ProgressMessage)
	}
	var result map[string]int
	if err := json.Unmarshal(cleanupJob.Result, &result); err != nil {
		t.Fatalf("failed to parse cleanup result: %v", err)
	}
	if result["deleted"] != 2 {
		t.Fatalf("expected deleted result to be 2, got %d", result["deleted"])
	}
}

func TestCleanupHandler_ValidateRequiresCondition(t *testing.T) {
	handler := NewCleanupHandler()
	payload, _ := json.Marshal(models.JobRetentionConfig{Enabled: true})
	if err := handler.Validate(payload); err == nil {
		t.Fatal("expected validation error when no retention condition is set")
	}
}

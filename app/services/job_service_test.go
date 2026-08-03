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
)

type panicJobHandler struct{}

func (h *panicJobHandler) Type() string { return "panic_test" }
func (h *panicJobHandler) Validate(payload json.RawMessage) error { return nil }
func (h *panicJobHandler) Execute(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	panic("intentional panic")
}
func (h *panicJobHandler) Resume(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	panic("intentional panic")
}
func (h *panicJobHandler) Cleanup(ctx context.Context, job *models.Job, runtime JobRuntime) error { return nil }

type successTestHandler struct{}

func (h *successTestHandler) Type() string { return "success_test" }
func (h *successTestHandler) Validate(payload json.RawMessage) error { return nil }
func (h *successTestHandler) Execute(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	return nil
}
func (h *successTestHandler) Resume(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	return nil
}
func (h *successTestHandler) Cleanup(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	return nil
}

func TestJobService_EmitsNotificationOnCompletion(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	jobRepo := repositories.NewJobRepository(repositories.NewSQLiteDBWrapper(db))
	service := NewJobService(jobRepo, nil, nil, nil, nil)
	service.RegisterHandler(&successTestHandler{})
	service.policies["success_test"] = JobPolicy{
		JobType:          "success_test",
		MaxConcurrent:    1,
		CanRunBackground: true,
		Recoverable:      false,
		Timeout:          time.Minute,
	}

	var emitted []models.NotificationEvent
	service.SetEventEmitter(func(eventName string, data interface{}) {
		if eventName == "notification:new" {
			if n, ok := data.(models.NotificationEvent); ok {
				emitted = append(emitted, n)
			}
		}
	})

	job := &models.Job{
		JobType:       "success_test",
		Status:        models.JobStatusPending,
		Title:         "test",
		Payload:       json.RawMessage("{}"),
		ProgressTotal: 1,
		CanResume:     false,
	}
	if err := jobRepo.Create(job); err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	service.runJob(job.ID, false)

	if len(emitted) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(emitted))
	}
	if emitted[0].Type != models.NotificationTypeJobCompleted {
		t.Errorf("expected type completed, got %s", emitted[0].Type)
	}
	if emitted[0].ID == "" {
		t.Error("expected notification id to be set")
	}
}

func TestJobService_SubmitSingleImportJob_CreatesJob(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	jobRepo := repositories.NewJobRepository(repositories.NewSQLiteDBWrapper(db))
	jobService := NewJobService(jobRepo, nil, nil, nil, nil)

	var createdJob *models.Job
	jobService.SetEventEmitter(func(eventName string, data interface{}) {
		if eventName == "job:created" {
			if j, ok := data.(*models.Job); ok {
				createdJob = j
			}
		}
	})

	req := ImportFileRequest{
		FilePath: "/tmp/test.txt",
		FileName: "test.txt",
		Tags:     []string{"test"},
	}
	job, err := jobService.SubmitSingleImportJob(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if job.JobType != models.JobTypeSingleImport {
		t.Errorf("expected job type %s, got %s", models.JobTypeSingleImport, job.JobType)
	}
	if createdJob == nil {
		t.Error("expected job:created event to be emitted")
	}
}

func TestJobService_RunJob_RecoversFromPanic(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	jobRepo := repositories.NewJobRepository(repositories.NewSQLiteDBWrapper(db))
	service := NewJobService(jobRepo, nil, nil, nil, nil)
	service.RegisterHandler(&panicJobHandler{})

	job := &models.Job{
		JobType:       "panic_test",
		Status:        models.JobStatusPending,
		Title:         "panic test",
		Payload:       json.RawMessage("{}"),
		ProgressTotal: 1,
		CanResume:     true,
	}
	if err := jobRepo.Create(job); err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	done := make(chan *models.Job, 1)
	service.SetEventEmitter(func(name string, data interface{}) {
		if name == "job:failed" {
			if j, ok := data.(*models.Job); ok {
				done <- j
			}
		}
	})

	go service.runJob(job.ID, false)

	select {
	case failedJob := <-done:
		if failedJob.Status != models.JobStatusFailed {
			t.Errorf("expected status failed, got %s", failedJob.Status)
		}
		if failedJob.ErrorMessage == "" {
			t.Error("expected error message after panic")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for panic recovery")
	}
}

type slowJobHandler struct {
	block chan struct{}
}

func (h *slowJobHandler) Type() string { return "slow_test" }
func (h *slowJobHandler) Validate(payload json.RawMessage) error { return nil }
func (h *slowJobHandler) Execute(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	<-h.block
	return ctx.Err()
}
func (h *slowJobHandler) Resume(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	<-h.block
	return ctx.Err()
}
func (h *slowJobHandler) Cleanup(ctx context.Context, job *models.Job, runtime JobRuntime) error { return nil }

func TestJobService_DeleteJobRecord_OnlyTerminal(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	repo := repositories.NewJobRepository(repositories.NewSQLiteDBWrapper(db))
	svc := NewJobService(repo, nil, nil, nil, nil)

	job := &models.Job{JobType: models.JobTypeBatchImport, Status: models.JobStatusRunning}
	if err := repo.Create(job); err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	if err := svc.DeleteJobRecord(job.ID); err == nil {
		t.Fatal("expected error deleting non-terminal job")
	}

	job.Status = models.JobStatusCompleted
	if err := repo.Update(job); err != nil {
		t.Fatalf("failed to update job: %v", err)
	}

	if err := svc.DeleteJobRecord(job.ID); err != nil {
		t.Fatalf("expected no error deleting terminal job: %v", err)
	}

	_, err = repo.FindByID(job.ID)
	if err == nil {
		t.Fatal("expected job to be deleted")
	}
}

func TestJobService_PrepareForShutdown_WaitsForRunningJob(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	jobRepo := repositories.NewJobRepository(repositories.NewSQLiteDBWrapper(db))
	service := NewJobService(jobRepo, nil, nil, nil, nil)
	handler := &slowJobHandler{block: make(chan struct{})}
	service.RegisterHandler(handler)
	service.policies["slow_test"] = JobPolicy{
		JobType:       "slow_test",
		MaxConcurrent: 1,
		Recoverable:   true,
	}

	job := &models.Job{
		JobType:       "slow_test",
		Status:        models.JobStatusPending,
		Title:         "slow test",
		Payload:       json.RawMessage("{}"),
		ProgressTotal: 1,
		CanResume:     true,
	}
	if err := jobRepo.Create(job); err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	go service.runJob(job.ID, false)
	// Give runJob time to start and block
	time.Sleep(100 * time.Millisecond)

	shutdownDone := make(chan struct{})
	go func() {
		if err := service.PrepareForShutdown(); err != nil {
			t.Errorf("PrepareForShutdown failed: %v", err)
		}
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		t.Fatal("PrepareForShutdown returned before job finished")
	case <-time.After(200 * time.Millisecond):
		// Expected: still waiting
	}

	close(handler.block)

	select {
	case <-shutdownDone:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for PrepareForShutdown to return")
	}
}

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

func TestJobService_RunJob_RecoversFromPanic(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	jobRepo := repositories.NewJobRepository(repositories.NewSQLiteDBWrapper(db))
	service := NewJobService(jobRepo, nil, nil, nil)
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

func TestJobService_PrepareForShutdown_WaitsForRunningJob(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	jobRepo := repositories.NewJobRepository(repositories.NewSQLiteDBWrapper(db))
	service := NewJobService(jobRepo, nil, nil, nil)
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

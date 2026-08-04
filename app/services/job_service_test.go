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

func (h *panicJobHandler) Type() string                           { return "panic_test" }
func (h *panicJobHandler) Validate(payload json.RawMessage) error { return nil }
func (h *panicJobHandler) Execute(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	panic("intentional panic")
}
func (h *panicJobHandler) Resume(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	panic("intentional panic")
}
func (h *panicJobHandler) Cleanup(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	return nil
}

type successTestHandler struct{}

func (h *successTestHandler) Type() string                           { return "success_test" }
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

func (h *slowJobHandler) Type() string                           { return "slow_test" }
func (h *slowJobHandler) Validate(payload json.RawMessage) error { return nil }
func (h *slowJobHandler) Execute(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	<-h.block
	return ctx.Err()
}
func (h *slowJobHandler) Resume(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	<-h.block
	return ctx.Err()
}
func (h *slowJobHandler) Cleanup(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	return nil
}

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

func TestJobService_GetExitGuardSnapshot_ProtectsOnlyConfiguredActiveStatuses(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	jobRepo := repositories.NewJobRepository(repositories.NewSQLiteDBWrapper(db))
	service := NewJobService(jobRepo, nil, nil, nil, nil)

	jobs := []*models.Job{
		{JobType: models.JobTypeBatchImport, Status: models.JobStatusPending, Title: "pending import"},
		{JobType: models.JobTypeBatchImport, Status: models.JobStatusRunning, Title: "running import"},
		{JobType: models.JobTypeBatchImport, Status: models.JobStatusRecovering, Title: "recovering import"},
		{JobType: models.JobTypeBatchImport, Status: models.JobStatusAwaitingResume, Title: "waiting import"},
		{JobType: models.JobTypeBatchImport, Status: models.JobStatusCompleted, Title: "completed import"},
	}
	for _, job := range jobs {
		if err := jobRepo.Create(job); err != nil {
			t.Fatalf("failed to create job %q: %v", job.Title, err)
		}
	}

	snapshot, err := service.GetExitGuardSnapshot()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !snapshot.HasProtectedJobs {
		t.Fatal("expected protected jobs")
	}
	if snapshot.Total != 3 {
		t.Fatalf("expected 3 protected jobs, got %d", snapshot.Total)
	}
	if snapshot.StatusCounts[models.JobStatusPending] != 1 {
		t.Fatalf("expected 1 pending job, got %d", snapshot.StatusCounts[models.JobStatusPending])
	}
	if snapshot.StatusCounts[models.JobStatusRunning] != 1 {
		t.Fatalf("expected 1 running job, got %d", snapshot.StatusCounts[models.JobStatusRunning])
	}
	if snapshot.StatusCounts[models.JobStatusRecovering] != 1 {
		t.Fatalf("expected 1 recovering job, got %d", snapshot.StatusCounts[models.JobStatusRecovering])
	}
	for _, job := range snapshot.Jobs {
		if job.Status == models.JobStatusAwaitingResume || job.Status == models.JobStatusCompleted {
			t.Fatalf("status %s should not be close-protected", job.Status)
		}
	}
}

func TestJobService_GetExitGuardSnapshot_IgnoresCleanupByPolicy(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	jobRepo := repositories.NewJobRepository(repositories.NewSQLiteDBWrapper(db))
	service := NewJobService(jobRepo, nil, nil, nil, nil)

	for _, status := range []string{models.JobStatusPending, models.JobStatusRunning, models.JobStatusRecovering} {
		job := &models.Job{JobType: models.JobTypeCleanup, Status: status, Title: "cleanup"}
		if err := jobRepo.Create(job); err != nil {
			t.Fatalf("failed to create cleanup job: %v", err)
		}
	}

	snapshot, err := service.GetExitGuardSnapshot()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if snapshot.HasProtectedJobs || snapshot.Total != 0 {
		t.Fatalf("cleanup jobs should not be protected, got total %d", snapshot.Total)
	}
}

func TestJobService_GetExitGuardSnapshot_NewTypeUsesPolicyCloseProtection(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	jobRepo := repositories.NewJobRepository(repositories.NewSQLiteDBWrapper(db))
	service := NewJobService(jobRepo, nil, nil, nil, nil)
	service.policies["custom_protected"] = JobPolicy{JobType: "custom_protected", CloseProtection: CloseProtectionActive}
	service.policies["custom_unprotected"] = JobPolicy{JobType: "custom_unprotected", CloseProtection: CloseProtectionNone}

	protectedJob := &models.Job{JobType: "custom_protected", Status: models.JobStatusRunning, Title: "protected"}
	unprotectedJob := &models.Job{JobType: "custom_unprotected", Status: models.JobStatusRunning, Title: "unprotected"}
	if err := jobRepo.Create(protectedJob); err != nil {
		t.Fatalf("failed to create protected job: %v", err)
	}
	if err := jobRepo.Create(unprotectedJob); err != nil {
		t.Fatalf("failed to create unprotected job: %v", err)
	}

	snapshot, err := service.GetExitGuardSnapshot()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if snapshot.Total != 1 {
		t.Fatalf("expected only custom protected job, got %d", snapshot.Total)
	}
	if len(snapshot.Jobs) != 1 || snapshot.Jobs[0].JobType != "custom_protected" {
		t.Fatalf("expected custom_protected snapshot, got %#v", snapshot.Jobs)
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

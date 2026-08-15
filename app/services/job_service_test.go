package services

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"sync/atomic"
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

// blockingCancelHandler 模拟"取消后不会立刻停"的真实情况：
// 复制过程不接收 ctx，所以 Execute 要等当前工作做完才返回。
type blockingCancelHandler struct {
	started      chan struct{}
	release      chan struct{}
	cleanupCalls int32
	cleanupErr   error
}

func (h *blockingCancelHandler) Type() string                           { return "blocking_test" }
func (h *blockingCancelHandler) Validate(payload json.RawMessage) error { return nil }

func (h *blockingCancelHandler) Execute(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	close(h.started)
	<-ctx.Done()
	// 取消信号到达后仍要"写完当前文件"，此时清理绝不能开始。
	<-h.release
	return ctx.Err()
}

func (h *blockingCancelHandler) Resume(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	return h.Execute(ctx, job, runtime)
}

func (h *blockingCancelHandler) Cleanup(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	atomic.AddInt32(&h.cleanupCalls, 1)
	return h.cleanupErr
}

func newCancelTestService(t *testing.T, handler JobHandler, jobType string) (*JobService, *repositories.JobRepository) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	jobRepo := repositories.NewJobRepository(repositories.NewSQLiteDBWrapper(db))
	service := NewJobService(jobRepo, nil, nil, nil, nil)
	service.RegisterHandler(handler)
	service.policies[jobType] = JobPolicy{
		JobType:          jobType,
		MaxConcurrent:    1,
		CanRunBackground: true,
		Recoverable:      true,
		Timeout:          time.Minute,
	}
	return service, jobRepo
}

func waitForStatus(t *testing.T, repo *repositories.JobRepository, jobID uint, want string) *models.Job {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	var last string
	for time.Now().Before(deadline) {
		job, err := repo.FindByID(jobID)
		if err == nil {
			last = job.Status
			if job.Status == want {
				return job
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("job %d never reached status %q (last %q)", jobID, want, last)
	return nil
}

// 取消运行中的任务时也要清理中间产物，并且必须等 Execute 真正退出后再清理。
func TestJobService_CancelRunningJob_RunsCleanupAfterRunnerExits(t *testing.T) {
	handler := &blockingCancelHandler{started: make(chan struct{}), release: make(chan struct{})}
	service, jobRepo := newCancelTestService(t, handler, "blocking_test")

	job := &models.Job{
		JobType:       "blocking_test",
		Status:        models.JobStatusPending,
		Title:         "cancel running",
		Payload:       json.RawMessage("{}"),
		ProgressTotal: 1,
		CanResume:     true,
	}
	if err := jobRepo.Create(job); err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	go service.runJob(job.ID, false)
	<-handler.started

	if err := service.CancelJob(job.ID); err != nil {
		t.Fatalf("CancelJob returned error: %v", err)
	}

	// runner 还卡在"写完当前文件"，此时清理必须还没发生。
	time.Sleep(80 * time.Millisecond)
	if got := atomic.LoadInt32(&handler.cleanupCalls); got != 0 {
		t.Fatalf("cleanup ran while the job was still executing (calls=%d)", got)
	}

	// CancelJob 不应该阻塞到 runner 退出，状态要能立刻反馈给界面。
	cancelled, err := jobRepo.FindByID(job.ID)
	if err != nil {
		t.Fatalf("failed to reload job: %v", err)
	}
	if cancelled.Status != models.JobStatusCancelled {
		t.Fatalf("expected status cancelled right after CancelJob, got %q", cancelled.Status)
	}

	close(handler.release)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&handler.cleanupCalls) == 1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("cleanup never ran after the runner exited (calls=%d)", atomic.LoadInt32(&handler.cleanupCalls))
}

// 清理失败要落到 cleanup_failed，而不是停在 cancelled 假装成功。
func TestJobService_CancelRunningJob_CleanupFailureMarksJob(t *testing.T) {
	handler := &blockingCancelHandler{
		started:    make(chan struct{}),
		release:    make(chan struct{}),
		cleanupErr: errors.New("temp file locked"),
	}
	service, jobRepo := newCancelTestService(t, handler, "blocking_test")

	job := &models.Job{
		JobType:       "blocking_test",
		Status:        models.JobStatusPending,
		Title:         "cleanup fails",
		Payload:       json.RawMessage("{}"),
		ProgressTotal: 1,
		CanResume:     true,
	}
	if err := jobRepo.Create(job); err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	go service.runJob(job.ID, false)
	<-handler.started
	if err := service.CancelJob(job.ID); err != nil {
		t.Fatalf("CancelJob returned error: %v", err)
	}
	close(handler.release)

	failed := waitForStatus(t, jobRepo, job.ID, models.JobStatusCleanupFailed)
	if failed.ErrorMessage != "temp file locked" {
		t.Errorf("expected cleanup error to be recorded, got %q", failed.ErrorMessage)
	}
}

// 关停走的是"等待恢复"，暂存文件要留给下次继续，绝不能被清理掉。
func TestJobService_ShutdownDoesNotCleanupResumableJob(t *testing.T) {
	handler := &blockingCancelHandler{started: make(chan struct{}), release: make(chan struct{})}
	service, jobRepo := newCancelTestService(t, handler, "blocking_test")

	job := &models.Job{
		JobType:       "blocking_test",
		Status:        models.JobStatusPending,
		Title:         "shutdown keeps artifacts",
		Payload:       json.RawMessage("{}"),
		ProgressTotal: 1,
		CanResume:     true,
	}
	if err := jobRepo.Create(job); err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	go service.runJob(job.ID, false)
	<-handler.started
	waitForStatus(t, jobRepo, job.ID, models.JobStatusRunning)

	close(handler.release)
	if err := service.PrepareForShutdown(); err != nil {
		t.Fatalf("PrepareForShutdown returned error: %v", err)
	}

	if got := atomic.LoadInt32(&handler.cleanupCalls); got != 0 {
		t.Fatalf("shutdown must not clean up a resumable job's artifacts (calls=%d)", got)
	}

	final, err := jobRepo.FindByID(job.ID)
	if err != nil {
		t.Fatalf("failed to reload job: %v", err)
	}
	if final.Status != models.JobStatusAwaitingResume {
		t.Errorf("expected awaiting_resume after shutdown, got %q", final.Status)
	}
}

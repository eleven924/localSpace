package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"LocalSpace/app/models"
	"LocalSpace/app/repositories"
)

type JobPolicy struct {
	JobType          string
	MaxConcurrent    int
	ExclusiveKey     string
	CanRunBackground bool
	Recoverable      bool
	Timeout          time.Duration
}

type JobHandler interface {
	Type() string
	Validate(payload json.RawMessage) error
	Execute(ctx context.Context, job *models.Job, runtime JobRuntime) error
	Resume(ctx context.Context, job *models.Job, runtime JobRuntime) error
	Cleanup(ctx context.Context, job *models.Job, runtime JobRuntime) error
}

type JobService struct {
	jobRepo      *repositories.JobRepository
	fileService  *FileService
	aiService    *AIService
	agentService *AgentService

	handlers map[string]JobHandler
	policies map[string]JobPolicy

	emitter func(eventName string, data interface{})

	mu           sync.Mutex
	running      map[uint]context.CancelFunc
	runningWg    sync.WaitGroup
	shuttingDown bool
}

func NewJobService(
	jobRepo *repositories.JobRepository,
	fileService *FileService,
	aiService *AIService,
	agentService *AgentService,
) *JobService {
	service := &JobService{
		jobRepo:      jobRepo,
		fileService:  fileService,
		aiService:    aiService,
		agentService: agentService,
		handlers:     make(map[string]JobHandler),
		policies: map[string]JobPolicy{
			models.JobTypeBatchImport: {
				JobType:          models.JobTypeBatchImport,
				MaxConcurrent:    1,
				ExclusiveKey:     "import",
				CanRunBackground: true,
				Recoverable:      true,
				Timeout:          2 * time.Hour,
			},
			models.JobTypeSingleImport: {
				JobType:          models.JobTypeSingleImport,
				MaxConcurrent:    1,
				ExclusiveKey:     "import",
				CanRunBackground: true,
				Recoverable:      false,
				Timeout:          30 * time.Minute,
			},
		},
		running: make(map[uint]context.CancelFunc),
	}

	service.RegisterHandler(NewBatchImportHandler())
	service.RegisterHandler(NewSingleImportHandler())
	go service.watchTimeouts()

	return service
}

func (s *JobService) SetEventEmitter(emitter func(eventName string, data interface{})) {
	s.emitter = emitter
}

func (s *JobService) RegisterHandler(handler JobHandler) {
	if handler == nil {
		return
	}
	s.handlers[handler.Type()] = handler
}

func (s *JobService) SubmitBatchImportJob(req models.BatchImportJobRequest) (*models.Job, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch import payload: %w", err)
	}

	handler, policy, err := s.handlerAndPolicy(models.JobTypeBatchImport)
	if err != nil {
		return nil, err
	}
	if err := handler.Validate(payload); err != nil {
		return nil, err
	}

	if err := s.ensureConcurrency(policy, 0); err != nil {
		return nil, err
	}

	job := &models.Job{
		JobType:           models.JobTypeBatchImport,
		Status:            models.JobStatusPending,
		Title:             fmt.Sprintf("批量导入 %d 个文件", len(req.Files)),
		Payload:           payload,
		ProgressTotal:     len(req.Files),
		ProgressCompleted: 0,
		ProgressMessage:   "等待开始导入",
		ExclusiveKey:      policy.ExclusiveKey,
		CanResume:         policy.Recoverable,
	}
	if err := s.jobRepo.Create(job); err != nil {
		return nil, err
	}

	items := make([]*models.BatchImportItem, 0, len(req.Files))
	for index, file := range req.Files {
		items = append(items, &models.BatchImportItem{
			JobID:       job.ID,
			SourcePath:  file.SourcePath,
			DisplayName: file.DisplayName,
			Status:      models.BatchImportItemStatusPending,
			ItemIndex:   index,
		})
	}
	if err := s.jobRepo.CreateBatchImportItems(items); err != nil {
		return nil, err
	}

	s.emitJobEvent("job:created", job)
	go s.runJob(job.ID, false)

	return job, nil
}

func (s *JobService) SubmitSingleImportJob(req ImportFileRequest) (*models.Job, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal single import payload: %w", err)
	}

	handler, policy, err := s.handlerAndPolicy(models.JobTypeSingleImport)
	if err != nil {
		return nil, err
	}
	if err := handler.Validate(payload); err != nil {
		return nil, err
	}

	if err := s.ensureConcurrency(policy, 0); err != nil {
		return nil, err
	}

	job := &models.Job{
		JobType:           models.JobTypeSingleImport,
		Status:            models.JobStatusPending,
		Title:             fmt.Sprintf("导入 %s", req.FileName),
		Payload:           payload,
		ProgressTotal:     1,
		ProgressCompleted: 0,
		ProgressMessage:   "等待开始导入",
		ExclusiveKey:      policy.ExclusiveKey,
		CanResume:         policy.Recoverable,
	}
	if err := s.jobRepo.Create(job); err != nil {
		return nil, err
	}

	s.emitJobEvent("job:created", job)
	go s.runJob(job.ID, false)

	return job, nil
}

func (s *JobService) GetActiveJobs() ([]*models.Job, error) {
	return s.jobRepo.ListActive()
}

func (s *JobService) GetResumableJobs() ([]*models.Job, error) {
	return s.jobRepo.ListResumable()
}

func (s *JobService) GetJob(jobID uint) (*models.Job, error) {
	return s.jobRepo.FindByID(jobID)
}

func (s *JobService) ListJobs(page, pageSize int, jobType string) (*models.JobListResponse, error) {
	items, total, err := s.jobRepo.List(page, pageSize, jobType)
	if err != nil {
		return nil, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	return &models.JobListResponse{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (s *JobService) ResumeJob(jobID uint) error {
	job, err := s.jobRepo.FindByID(jobID)
	if err != nil {
		return err
	}
	if job.Status != models.JobStatusAwaitingResume && job.Status != models.JobStatusTimedOut {
		return fmt.Errorf("job %d is not resumable", jobID)
	}
	if err := s.ensureConcurrency(s.policies[job.JobType], jobID); err != nil {
		return err
	}
	go s.runJob(jobID, true)
	return nil
}

func (s *JobService) CancelJob(jobID uint) error {
	job, err := s.jobRepo.FindByID(jobID)
	if err != nil {
		return err
	}

	if job.Status == models.JobStatusAwaitingResume || job.Status == models.JobStatusTimedOut {
		handler, _, err := s.handlerAndPolicy(job.JobType)
		if err != nil {
			return err
		}
		runtime := &jobRuntime{service: s, job: job}
		if err := handler.Cleanup(context.Background(), job, runtime); err != nil {
			job.Status = models.JobStatusCleanupFailed
			job.ErrorMessage = err.Error()
			job.FinishedAt = time.Now().Format(time.RFC3339)
			if updateErr := s.jobRepo.Update(job); updateErr != nil {
				return updateErr
			}
			s.emitJobEvent("job:failed", job)
			return err
		}
		job.Status = models.JobStatusCancelled
		job.ErrorMessage = ""
		job.FinishedAt = time.Now().Format(time.RFC3339)
		if err := s.jobRepo.Update(job); err != nil {
			return err
		}
		s.emitJobEvent("job:updated", job)
		return nil
	}

	s.mu.Lock()
	cancel, ok := s.running[jobID]
	s.mu.Unlock()
	if ok && cancel != nil {
		cancel()
	}

	job.Status = models.JobStatusCancelled
	job.ErrorMessage = ""
	job.FinishedAt = time.Now().Format(time.RFC3339)
	if err := s.jobRepo.Update(job); err != nil {
		return err
	}
	s.emitJobEvent("job:updated", job)
	return nil
}

func (s *JobService) NormalizeUnfinishedJobs() error {
	jobs, err := s.jobRepo.ListUnfinishedForStartup()
	if err != nil {
		return err
	}

	for _, job := range jobs {
		if !job.CanResume {
			job.Status = models.JobStatusFailed
			job.ErrorMessage = "任务在应用退出后无法自动恢复"
			job.FinishedAt = time.Now().Format(time.RFC3339)
		} else if job.Status == models.JobStatusRunning || job.Status == models.JobStatusRecovering {
			job.Status = models.JobStatusAwaitingResume
			job.ErrorMessage = "任务在上次退出后等待恢复"
		}
		if err := s.jobRepo.Update(job); err != nil {
			return err
		}
	}

	return nil
}

func (s *JobService) HasBackgroundJobsForCloseProtection() (bool, int, error) {
	jobs, err := s.jobRepo.ListActive()
	if err != nil {
		return false, 0, err
	}

	count := 0
	for _, job := range jobs {
		policy, ok := s.policies[job.JobType]
		if !ok || !policy.CanRunBackground {
			continue
		}
		if job.Status == models.JobStatusRunning || job.Status == models.JobStatusRecovering || job.Status == models.JobStatusPending {
			count++
		}
	}

	return count > 0, count, nil
}

func (s *JobService) PrepareForShutdown() error {
	s.mu.Lock()
	s.shuttingDown = true
	running := make(map[uint]context.CancelFunc, len(s.running))
	for id, cancel := range s.running {
		running[id] = cancel
	}
	s.mu.Unlock()

	jobs, err := s.jobRepo.ListActive()
	if err != nil {
		return err
	}

	for _, job := range jobs {
		policy, ok := s.policies[job.JobType]
		if !ok || !policy.Recoverable {
			continue
		}
		if job.Status == models.JobStatusRunning || job.Status == models.JobStatusRecovering || job.Status == models.JobStatusPending {
			job.Status = models.JobStatusAwaitingResume
			job.ErrorMessage = "应用关闭后等待恢复"
			job.HeartbeatAt = time.Now().Format(time.RFC3339)
			if err := s.jobRepo.Update(job); err != nil {
				return err
			}
			s.emitJobEvent("job:needs-resume", job)
		}
	}

	for _, cancel := range running {
		if cancel != nil {
			cancel()
		}
	}

	s.runningWg.Wait()
	return nil
}

func (s *JobService) IsShuttingDown() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.shuttingDown
}

func (s *JobService) emitJobEvent(eventName string, data interface{}) {
	if s.emitter != nil {
		s.emitter(eventName, data)
	}
}

func (s *JobService) emitNotificationEvent(job *models.Job) {
	if s.emitter == nil {
		return
	}

	notification := models.NotificationEvent{
		ID:        fmt.Sprintf("notif-%d-%d", job.ID, time.Now().UnixMilli()),
		Type:      s.notificationTypeForJob(job),
		Title:     s.notificationTitleForJob(job),
		Message:   s.notificationMessageForJob(job),
		Payload:   map[string]interface{}{"jobId": job.ID, "jobType": job.JobType},
		CreatedAt: time.Now().Format(time.RFC3339),
		Read:      false,
	}

	s.emitter("notification:new", notification)
}

func (s *JobService) notificationTypeForJob(job *models.Job) models.NotificationType {
	switch job.Status {
	case models.JobStatusCompleted:
		return models.NotificationTypeJobCompleted
	case models.JobStatusFailed:
		return models.NotificationTypeJobFailed
	case models.JobStatusCancelled:
		return models.NotificationTypeJobCancelled
	default:
		return models.NotificationTypeJobCompleted
	}
}

func (s *JobService) notificationTitleForJob(job *models.Job) string {
	isSingle := job.JobType == models.JobTypeSingleImport
	switch job.Status {
	case models.JobStatusCompleted:
		if isSingle {
			return "文件导入成功"
		}
		return "批量导入完成"
	case models.JobStatusFailed:
		if isSingle {
			return "文件导入失败"
		}
		return "批量导入失败"
	case models.JobStatusCancelled:
		if isSingle {
			return "文件导入已取消"
		}
		return "批量导入已取消"
	default:
		return "任务状态更新"
	}
}

func (s *JobService) notificationMessageForJob(job *models.Job) string {
	if job.Status == models.JobStatusCompleted && job.JobType == models.JobTypeSingleImport {
		if job.Title != "" {
			return job.Title
		}
		return "文件导入成功"
	}
	if job.Status == models.JobStatusCompleted && job.JobType == models.JobTypeBatchImport {
		var result models.BatchImportResult
		if err := json.Unmarshal(job.Result, &result); err == nil {
			return fmt.Sprintf("成功导入 %d 个文件，失败 %d 个", result.SuccessCount, result.FailedCount)
		}
		return "批量导入已完成"
	}
	if job.ErrorMessage != "" {
		return job.ErrorMessage
	}
	return job.ProgressMessage
}

func (s *JobService) runJob(jobID uint, resume bool) {
	job, err := s.jobRepo.FindByID(jobID)
	if err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			panicErr := fmt.Errorf("job panicked: %v\n%s", r, debug.Stack())
			fmt.Printf("Job %d panic: %v\n", jobID, panicErr)

			latestJob, latestErr := s.jobRepo.FindByID(jobID)
			if latestErr == nil {
				job = latestJob
			}
			job.Status = models.JobStatusFailed
			job.ErrorMessage = panicErr.Error()
			job.FinishedAt = time.Now().Format(time.RFC3339)
			_ = s.jobRepo.Update(job)
			s.emitJobEvent("job:failed", job)
			s.emitNotificationEvent(job)
		}
	}()

	handler, policy, err := s.handlerAndPolicy(job.JobType)
	if err != nil {
		job.Status = models.JobStatusFailed
		job.ErrorMessage = err.Error()
		job.FinishedAt = time.Now().Format(time.RFC3339)
		_ = s.jobRepo.Update(job)
		s.emitJobEvent("job:failed", job)
		return
	}

	if err := s.ensureConcurrency(policy, jobID); err != nil {
		job.Status = models.JobStatusFailed
		job.ErrorMessage = err.Error()
		job.FinishedAt = time.Now().Format(time.RFC3339)
		_ = s.jobRepo.Update(job)
		s.emitJobEvent("job:failed", job)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.trackRunning(jobID, cancel)
	s.runningWg.Add(1)
	defer func() {
		s.untrackRunning(jobID)
		s.runningWg.Done()
	}()

	now := time.Now().Format(time.RFC3339)
	if resume {
		job.Status = models.JobStatusRecovering
		job.ProgressMessage = "正在恢复任务"
	} else {
		job.Status = models.JobStatusRunning
		if job.ProgressMessage == "" {
			job.ProgressMessage = "任务开始执行"
		}
	}
	if job.StartedAt == "" {
		job.StartedAt = now
	}
	job.HeartbeatAt = now
	if policy.Timeout > 0 {
		job.TimeoutAt = time.Now().Add(policy.Timeout).Format(time.RFC3339)
	}
	job.ErrorMessage = ""
	if err := s.jobRepo.Update(job); err != nil {
		return
	}
	s.emitJobEvent("job:updated", job)

	runtime := &jobRuntime{service: s, job: job}
	if resume {
		err = handler.Resume(ctx, job, runtime)
	} else {
		err = handler.Execute(ctx, job, runtime)
	}

	latestJob, latestErr := s.jobRepo.FindByID(jobID)
	if latestErr == nil {
		job = latestJob
	}

	if errors.Is(err, context.Canceled) {
		switch job.Status {
		case models.JobStatusCancelled, models.JobStatusAwaitingResume, models.JobStatusTimedOut:
			s.emitJobEvent("job:updated", job)
			if job.Status == models.JobStatusCancelled {
				s.emitNotificationEvent(job)
			}
			return
		default:
			if s.IsShuttingDown() && policy.Recoverable {
				job.Status = models.JobStatusAwaitingResume
				job.ErrorMessage = "应用关闭后等待恢复"
				job.FinishedAt = ""
				_ = s.jobRepo.Update(job)
				s.emitJobEvent("job:needs-resume", job)
				return
			}
			job.Status = models.JobStatusCancelled
			job.FinishedAt = time.Now().Format(time.RFC3339)
			_ = s.jobRepo.Update(job)
			s.emitJobEvent("job:updated", job)
			return
		}
	}

	if err != nil {
		job.Status = models.JobStatusFailed
		job.ErrorMessage = err.Error()
		job.FinishedAt = time.Now().Format(time.RFC3339)
		_ = s.jobRepo.Update(job)
		s.emitJobEvent("job:failed", job)
		s.emitNotificationEvent(job)
		return
	}

	job.Status = models.JobStatusCompleted
	job.ErrorMessage = ""
	job.ProgressCompleted = job.ProgressTotal
	job.FinishedAt = time.Now().Format(time.RFC3339)
	if err := s.jobRepo.Update(job); err != nil {
		return
	}
	s.emitJobEvent("job:completed", job)
	s.emitNotificationEvent(job)
}

func (s *JobService) watchTimeouts() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		jobs, err := s.jobRepo.ListActive()
		if err != nil {
			continue
		}

		now := time.Now()
		for _, job := range jobs {
			if job.Status != models.JobStatusRunning && job.Status != models.JobStatusRecovering {
				continue
			}
			timeout := s.policyTimeout(job.JobType)
			if timeout <= 0 {
				continue
			}

			heartbeatAt, err := time.Parse(time.RFC3339, job.HeartbeatAt)
			if err != nil {
				heartbeatAt = now
			}
			timeoutAt, err := time.Parse(time.RFC3339, job.TimeoutAt)
			if err != nil || timeoutAt.IsZero() {
				timeoutAt = heartbeatAt.Add(timeout)
			}

			if now.After(timeoutAt) || now.After(heartbeatAt.Add(timeout)) {
				job.Status = models.JobStatusTimedOut
				job.ErrorMessage = "任务执行超时"
				job.FinishedAt = now.Format(time.RFC3339)
				if err := s.jobRepo.Update(job); err == nil {
					s.emitJobEvent("job:failed", job)
					s.emitNotificationEvent(job)
				}
				s.mu.Lock()
				cancel := s.running[job.ID]
				s.mu.Unlock()
				if cancel != nil {
					cancel()
				}
			}
		}
	}
}

func (s *JobService) handlerAndPolicy(jobType string) (JobHandler, JobPolicy, error) {
	handler, ok := s.handlers[jobType]
	if !ok {
		return nil, JobPolicy{}, fmt.Errorf("job handler not registered for type %s", jobType)
	}
	policy, ok := s.policies[jobType]
	if !ok {
		return nil, JobPolicy{}, fmt.Errorf("job policy not registered for type %s", jobType)
	}
	return handler, policy, nil
}

func (s *JobService) ensureConcurrency(policy JobPolicy, excludeJobID uint) error {
	if policy.ExclusiveKey == "" || policy.MaxConcurrent <= 0 {
		return nil
	}
	activeCount, err := s.jobRepo.CountActiveByExclusiveKey(policy.ExclusiveKey, excludeJobID)
	if err != nil {
		return err
	}
	if activeCount >= policy.MaxConcurrent {
		return fmt.Errorf("已有同类型任务正在运行，请等待当前任务完成后再试")
	}
	return nil
}

func (s *JobService) policyTimeout(jobType string) time.Duration {
	policy, ok := s.policies[jobType]
	if !ok {
		return 0
	}
	return policy.Timeout
}

func (s *JobService) trackRunning(jobID uint, cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running[jobID] = cancel
}

func (s *JobService) untrackRunning(jobID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.running, jobID)
}

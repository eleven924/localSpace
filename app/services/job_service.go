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
	CloseProtection  string
	Timeout          time.Duration
}

const (
	CloseProtectionNone   = "none"
	CloseProtectionActive = "active"
)

type JobHandler interface {
	Type() string
	Validate(payload json.RawMessage) error
	Execute(ctx context.Context, job *models.Job, runtime JobRuntime) error
	Resume(ctx context.Context, job *models.Job, runtime JobRuntime) error
	Cleanup(ctx context.Context, job *models.Job, runtime JobRuntime) error
}

type JobService struct {
	jobRepo       *repositories.JobRepository
	fileService   *FileService
	aiService     *AIService
	agentService  *AgentService
	configService *ConfigService

	handlers map[string]JobHandler
	policies map[string]JobPolicy

	emitter func(eventName string, data interface{})

	mu           sync.Mutex
	running      map[uint]*jobRun
	runningWg    sync.WaitGroup
	shuttingDown bool
}

// jobRun 记录一次运行的取消函数和退出信号。
// 需要 done 是因为取消运行中的任务后，必须等 Execute 真正返回才能清理暂存文件，
// 否则会和仍在写盘的复制过程打架。
type jobRun struct {
	cancel context.CancelFunc
	done   chan struct{}
}

func NewJobService(
	jobRepo *repositories.JobRepository,
	fileService *FileService,
	aiService *AIService,
	agentService *AgentService,
	configService *ConfigService,
) *JobService {
	service := &JobService{
		jobRepo:       jobRepo,
		fileService:   fileService,
		aiService:     aiService,
		agentService:  agentService,
		configService: configService,
		handlers:      make(map[string]JobHandler),
		policies: map[string]JobPolicy{
			models.JobTypeBatchImport: {
				JobType:          models.JobTypeBatchImport,
				MaxConcurrent:    1,
				ExclusiveKey:     "import",
				CanRunBackground: true,
				Recoverable:      true,
				CloseProtection:  CloseProtectionActive,
				Timeout:          2 * time.Hour,
			},
			models.JobTypeSingleImport: {
				JobType:          models.JobTypeSingleImport,
				MaxConcurrent:    1,
				ExclusiveKey:     "import",
				CanRunBackground: true,
				Recoverable:      false,
				CloseProtection:  CloseProtectionActive,
				Timeout:          30 * time.Minute,
			},
			models.JobTypeCleanup: {
				JobType:          models.JobTypeCleanup,
				MaxConcurrent:    1,
				ExclusiveKey:     "cleanup",
				CanRunBackground: true,
				Recoverable:      false,
				CloseProtection:  CloseProtectionNone,
				Timeout:          5 * time.Minute,
			},
		},
		running: make(map[uint]*jobRun),
	}

	service.RegisterHandler(NewBatchImportHandler())
	service.RegisterHandler(NewSingleImportHandler())
	service.RegisterHandler(NewCleanupHandler())
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

// GetBatchImportItems 返回批量导入任务的逐文件状态，供任务中心展示取消后的导入结果。
func (s *JobService) GetBatchImportItems(jobID uint) ([]*models.BatchImportItem, error) {
	return s.jobRepo.ListBatchImportItems(jobID)
}

// populateBatchImportResult 根据逐文件状态生成取消任务的最终汇总，避免把处理中进度误当成成功数量。
func (s *JobService) populateBatchImportResult(job *models.Job) error {
	if job == nil || job.JobType != models.JobTypeBatchImport {
		return nil
	}

	items, err := s.jobRepo.ListBatchImportItems(job.ID)
	if err != nil {
		return err
	}

	result := models.BatchImportResult{FailedItems: make([]models.BatchImportFailedItem, 0)}
	for _, item := range items {
		if item.Status == models.BatchImportItemStatusCompleted {
			result.SuccessCount++
			continue
		}

		result.FailedCount++
		errorMessage := item.ErrorMessage
		if errorMessage == "" {
			errorMessage = "文件未完成导入"
		}
		result.FailedItems = append(result.FailedItems, models.BatchImportFailedItem{
			SourcePath:  item.SourcePath,
			DisplayName: item.DisplayName,
			Error:       errorMessage,
		})
	}

	payload, err := json.Marshal(result)
	if err != nil {
		return err
	}
	job.Result = payload
	job.ProgressCompleted = result.SuccessCount + result.FailedCount
	job.ProgressMessage = fmt.Sprintf("任务已取消：成功导入 %d 个，未导入 %d 个", result.SuccessCount, result.FailedCount)
	return nil
}

func (s *JobService) ListJobs(page, pageSize int, jobType string) (*models.JobListResponse, error) {
	items, total, err := s.jobRepo.List(page, pageSize, jobType)
	if err != nil {
		return nil, err
	}

	// 兼容旧版本已经取消的批量任务：首次打开任务中心时补算一次最终汇总。
	for _, job := range items {
		if job.Status != models.JobStatusCancelled || job.JobType != models.JobTypeBatchImport {
			continue
		}
		if err := s.populateBatchImportResult(job); err != nil {
			return nil, err
		}
		if err := s.jobRepo.Update(job); err != nil {
			return nil, err
		}
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
		if err := s.populateBatchImportResult(job); err != nil {
			return err
		}
		if err := s.jobRepo.Update(job); err != nil {
			return err
		}
		s.emitJobEvent("job:updated", job)
		return nil
	}

	s.mu.Lock()
	run, ok := s.running[jobID]
	s.mu.Unlock()
	if ok && run != nil && run.cancel != nil {
		run.cancel()
	}

	job.Status = models.JobStatusCancelled
	job.ErrorMessage = ""
	job.FinishedAt = time.Now().Format(time.RFC3339)
	if err := s.jobRepo.Update(job); err != nil {
		return err
	}
	s.emitJobEvent("job:updated", job)

	// Execute 收到取消后直接返回，不会自己走 Cleanup。这里补上，否则批量导入的
	// 暂存文件会一直留在磁盘上——同一个任务在"等待继续"状态取消却会清理干净，
	// 两条路径必须一致。复制过程不接收 ctx，所以要等 runner 真正退出再清理。
	if ok && run != nil {
		go s.cleanupAfterCancel(jobID, run.done)
	}
	return nil
}

// cleanupAfterCancel 等运行中的任务退出后清理它留下的中间产物。
// 只处理确实停在 cancelled 的任务：关停走的是"等待恢复"，那些暂存文件要留给下次继续。
func (s *JobService) cleanupAfterCancel(jobID uint, done <-chan struct{}) {
	<-done

	job, err := s.jobRepo.FindByID(jobID)
	if err != nil || job.Status != models.JobStatusCancelled {
		return
	}

	handler, _, err := s.handlerAndPolicy(job.JobType)
	if err != nil {
		return
	}

	runtime := &jobRuntime{service: s, job: job}
	if cleanupErr := handler.Cleanup(context.Background(), job, runtime); cleanupErr != nil {
		latest, findErr := s.jobRepo.FindByID(jobID)
		if findErr != nil || latest.Status != models.JobStatusCancelled {
			return
		}
		latest.Status = models.JobStatusCleanupFailed
		latest.ErrorMessage = cleanupErr.Error()
		latest.FinishedAt = time.Now().Format(time.RFC3339)
		if updateErr := s.jobRepo.Update(latest); updateErr != nil {
			return
		}
		s.emitJobEvent("job:failed", latest)
		return
	}

	if err := s.populateBatchImportResult(job); err != nil {
		return
	}
	if err := s.jobRepo.Update(job); err != nil {
		return
	}
	s.emitJobEvent("job:updated", job)
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

func (s *JobService) SubmitJobCleanup() (*models.Job, error) {
	if s.configService == nil {
		return nil, fmt.Errorf("config service not initialized")
	}
	config, err := s.configService.GetJobRetentionConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get retention config: %w", err)
	}
	payload, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal retention config: %w", err)
	}
	handler, policy, err := s.handlerAndPolicy(models.JobTypeCleanup)
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
		JobType:           models.JobTypeCleanup,
		Status:            models.JobStatusPending,
		Title:             "清理历史任务记录",
		Payload:           payload,
		ProgressTotal:     1,
		ProgressCompleted: 0,
		ProgressMessage:   "等待开始清理",
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

func (s *JobService) DeleteJobRecord(id uint) error {
	job, err := s.jobRepo.FindByID(id)
	if err != nil {
		return err
	}
	terminal := []string{
		models.JobStatusCompleted,
		models.JobStatusFailed,
		models.JobStatusCancelled,
		models.JobStatusTimedOut,
		models.JobStatusCleanupFailed,
	}
	found := false
	for _, status := range terminal {
		if job.Status == status {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("only terminal jobs can be deleted")
	}
	return s.jobRepo.DeleteByIDs([]uint{id})
}

func (s *JobService) MaybeSubmitAutoCleanup() error {
	if s.configService == nil {
		return nil
	}
	config, err := s.configService.GetJobRetentionConfig()
	if err != nil {
		return err
	}
	if !config.Enabled || (config.MaxCount <= 0 && config.MaxDays <= 0) {
		return nil
	}
	terminal, err := s.jobRepo.ListByStatuses(
		models.JobStatusCompleted,
		models.JobStatusFailed,
		models.JobStatusCancelled,
		models.JobStatusTimedOut,
		models.JobStatusCleanupFailed,
	)
	if err != nil {
		return err
	}
	needs := false
	if config.MaxCount > 0 && len(terminal) > config.MaxCount {
		needs = true
	}
	if config.MaxDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -config.MaxDays)
		for _, j := range terminal {
			created, err := time.Parse(time.RFC3339, j.CreatedAt)
			if err != nil {
				continue
			}
			if created.Before(cutoff) {
				needs = true
				break
			}
		}
	}
	if !needs {
		return nil
	}
	_, err = s.SubmitJobCleanup()
	return err
}

func (s *JobService) HasBackgroundJobsForCloseProtection() (bool, int, error) {
	snapshot, err := s.GetExitGuardSnapshot()
	if err != nil {
		return false, 0, err
	}
	return snapshot.HasProtectedJobs, snapshot.Total, nil
}

func (s *JobService) GetExitGuardSnapshot() (*models.ExitGuardSnapshot, error) {
	jobs, err := s.jobRepo.ListActive()
	if err != nil {
		return nil, err
	}

	snapshot := &models.ExitGuardSnapshot{
		StatusCounts: map[string]int{
			models.JobStatusPending:    0,
			models.JobStatusRunning:    0,
			models.JobStatusRecovering: 0,
		},
		Jobs: []*models.ExitGuardJobSnapshot{},
	}

	for _, job := range jobs {
		if !s.shouldProtectJobOnClose(job) {
			continue
		}

		// 退出保护只统计还在执行链路上的任务，等待恢复的任务已经停住，不需要再次拦截退出。
		snapshot.Total++
		snapshot.StatusCounts[job.Status]++
		snapshot.Jobs = append(snapshot.Jobs, &models.ExitGuardJobSnapshot{
			ID:                job.ID,
			JobType:           job.JobType,
			Status:            job.Status,
			Title:             job.Title,
			ProgressTotal:     job.ProgressTotal,
			ProgressCompleted: job.ProgressCompleted,
			ProgressMessage:   job.ProgressMessage,
			CanResume:         job.CanResume,
		})
	}

	snapshot.HasProtectedJobs = snapshot.Total > 0
	return snapshot, nil
}

func (s *JobService) shouldProtectJobOnClose(job *models.Job) bool {
	if job == nil {
		return false
	}
	policy, ok := s.policies[job.JobType]
	if !ok {
		return false
	}
	if policy.CloseProtection != CloseProtectionActive {
		return false
	}
	return job.Status == models.JobStatusPending ||
		job.Status == models.JobStatusRunning ||
		job.Status == models.JobStatusRecovering
}

func (s *JobService) PrepareForShutdown() error {
	s.mu.Lock()
	s.shuttingDown = true
	running := make(map[uint]*jobRun, len(s.running))
	for id, run := range s.running {
		running[id] = run
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

	for _, run := range running {
		if run != nil && run.cancel != nil {
			run.cancel()
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
	run := s.trackRunning(jobID, cancel)
	s.runningWg.Add(1)
	defer func() {
		s.untrackRunning(jobID)
		close(run.done)
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
				run := s.running[job.ID]
				s.mu.Unlock()
				if run != nil && run.cancel != nil {
					run.cancel()
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

func (s *JobService) trackRunning(jobID uint, cancel context.CancelFunc) *jobRun {
	s.mu.Lock()
	defer s.mu.Unlock()
	run := &jobRun{cancel: cancel, done: make(chan struct{})}
	s.running[jobID] = run
	return run
}

func (s *JobService) untrackRunning(jobID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.running, jobID)
}

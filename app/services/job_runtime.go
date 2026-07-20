package services

import (
	"encoding/json"
	"fmt"
	"time"

	"LocalSpace/app/models"
	"LocalSpace/app/repositories"
)

type JobRuntime interface {
	UpdateProgress(completed, total int, message string) error
	UpdateHeartbeat(message string) error
	SetResult(result interface{}) error
	JobRepository() *repositories.JobRepository
	FileService() *FileService
	AgentService() *AgentService
	AIService() *AIService
	IsShuttingDown() bool
}

type jobRuntime struct {
	service *JobService
	job     *models.Job
}

func (r *jobRuntime) UpdateProgress(completed, total int, message string) error {
	if r == nil || r.service == nil || r.job == nil {
		return fmt.Errorf("job runtime is not initialized")
	}

	r.job.ProgressCompleted = completed
	r.job.ProgressTotal = total
	r.job.ProgressMessage = message

	now := time.Now().Format(time.RFC3339)
	r.job.HeartbeatAt = now
	if timeout := r.service.policyTimeout(r.job.JobType); timeout > 0 {
		r.job.TimeoutAt = time.Now().Add(timeout).Format(time.RFC3339)
	}

	if err := r.service.jobRepo.Update(r.job); err != nil {
		return err
	}
	r.service.emitJobEvent("job:updated", r.job)
	return nil
}

func (r *jobRuntime) UpdateHeartbeat(message string) error {
	if r == nil || r.service == nil || r.job == nil {
		return fmt.Errorf("job runtime is not initialized")
	}
	if message != "" {
		r.job.ProgressMessage = message
	}
	now := time.Now().Format(time.RFC3339)
	r.job.HeartbeatAt = now
	if timeout := r.service.policyTimeout(r.job.JobType); timeout > 0 {
		r.job.TimeoutAt = time.Now().Add(timeout).Format(time.RFC3339)
	}
	if err := r.service.jobRepo.Update(r.job); err != nil {
		return err
	}
	r.service.emitJobEvent("job:updated", r.job)
	return nil
}

func (r *jobRuntime) SetResult(result interface{}) error {
	if r == nil || r.service == nil || r.job == nil {
		return fmt.Errorf("job runtime is not initialized")
	}

	if result == nil {
		r.job.Result = nil
	} else {
		payload, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal job result: %w", err)
		}
		r.job.Result = payload
	}

	if err := r.service.jobRepo.Update(r.job); err != nil {
		return err
	}
	r.service.emitJobEvent("job:updated", r.job)
	return nil
}

func (r *jobRuntime) JobRepository() *repositories.JobRepository {
	return r.service.jobRepo
}

func (r *jobRuntime) FileService() *FileService {
	return r.service.fileService
}

func (r *jobRuntime) AgentService() *AgentService {
	return r.service.agentService
}

func (r *jobRuntime) AIService() *AIService {
	return r.service.aiService
}

func (r *jobRuntime) IsShuttingDown() bool {
	return r.service.IsShuttingDown()
}

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"LocalSpace/app/models"
)

type CleanupHandler struct{}

func NewCleanupHandler() *CleanupHandler {
	return &CleanupHandler{}
}

func (h *CleanupHandler) Type() string {
	return models.JobTypeCleanup
}

func (h *CleanupHandler) Validate(payload json.RawMessage) error {
	var config models.JobRetentionConfig
	if err := json.Unmarshal(payload, &config); err != nil {
		return fmt.Errorf("invalid retention config: %w", err)
	}
	if config.MaxCount <= 0 && config.MaxDays <= 0 {
		return fmt.Errorf("at least one retention condition must be enabled")
	}
	return nil
}

func (h *CleanupHandler) Execute(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	var config models.JobRetentionConfig
	if err := json.Unmarshal(job.Payload, &config); err != nil {
		return fmt.Errorf("failed to parse cleanup payload: %w", err)
	}

	repo := runtime.JobRepository()
	terminalStatuses := []string{
		models.JobStatusCompleted,
		models.JobStatusFailed,
		models.JobStatusCancelled,
		models.JobStatusTimedOut,
		models.JobStatusCleanupFailed,
	}
	jobs, err := repo.ListByStatuses(terminalStatuses...)
	if err != nil {
		return fmt.Errorf("failed to list terminal jobs: %w", err)
	}

	toDelete := make(map[uint]bool)
	if config.Enabled && config.MaxCount > 0 && len(jobs) > config.MaxCount {
		excess := len(jobs) - config.MaxCount
		for i := 0; i < excess; i++ {
			toDelete[jobs[i].ID] = true
		}
	}
	if config.Enabled && config.MaxDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -config.MaxDays)
		for _, j := range jobs {
			created, err := time.Parse(time.RFC3339, j.CreatedAt)
			if err != nil {
				continue
			}
			if created.Before(cutoff) {
				toDelete[j.ID] = true
			}
		}
	}

	if len(toDelete) == 0 {
		return nil
	}

	ids := make([]uint, 0, len(toDelete))
	for id := range toDelete {
		ids = append(ids, id)
	}
	if err := repo.DeleteByIDs(ids); err != nil {
		return fmt.Errorf("failed to delete jobs: %w", err)
	}

	result := map[string]int{"deleted": len(ids)}
	resultJSON, _ := json.Marshal(result)
	job.Result = resultJSON
	job.ProgressMessage = fmt.Sprintf("已清理 %d 条任务记录", len(ids))
	return nil
}

func (h *CleanupHandler) Resume(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	return fmt.Errorf("cleanup job does not support resume")
}

func (h *CleanupHandler) Cleanup(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	return nil
}

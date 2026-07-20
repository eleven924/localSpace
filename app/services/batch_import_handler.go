package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"LocalSpace/app/models"
)

type BatchImportHandler struct{}

func NewBatchImportHandler() *BatchImportHandler {
	return &BatchImportHandler{}
}

func (h *BatchImportHandler) Type() string {
	return models.JobTypeBatchImport
}

func (h *BatchImportHandler) Validate(payload json.RawMessage) error {
	var req models.BatchImportJobRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("invalid batch import payload: %w", err)
	}
	if len(req.Files) == 0 {
		return fmt.Errorf("至少需要选择一个文件")
	}
	for index, file := range req.Files {
		if strings.TrimSpace(file.SourcePath) == "" {
			return fmt.Errorf("第 %d 个文件缺少源路径", index+1)
		}
		displayName := strings.TrimSpace(file.DisplayName)
		if displayName == "" {
			return fmt.Errorf("第 %d 个文件缺少显示名称", index+1)
		}
	}
	return nil
}

func (h *BatchImportHandler) Execute(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	return h.process(ctx, job, runtime, false)
}

func (h *BatchImportHandler) Resume(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	return h.process(ctx, job, runtime, true)
}

func (h *BatchImportHandler) Cleanup(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	items, err := runtime.JobRepository().ListBatchImportItems(job.ID)
	if err != nil {
		return err
	}

	for _, item := range items {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := runtime.FileService().CleanupBatchImportArtifacts(item.TempPath, item.FinalPath); err != nil {
			return err
		}
		if item.Status != models.BatchImportItemStatusCompleted {
			item.Status = models.BatchImportItemStatusFailed
			item.ErrorMessage = "任务已取消并清理中间文件"
			item.FinishedAt = time.Now().Format(time.RFC3339)
			if err := runtime.JobRepository().UpdateBatchImportItem(item); err != nil {
				return err
			}
		}
	}

	return runtime.UpdateHeartbeat("已清理未完成的导入文件")
}

func (h *BatchImportHandler) process(ctx context.Context, job *models.Job, runtime JobRuntime, resuming bool) error {
	var payload models.BatchImportJobRequest
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("failed to parse batch import payload: %w", err)
	}

	items, err := runtime.JobRepository().ListBatchImportItems(job.ID)
	if err != nil {
		return err
	}

	processed := countProcessedItems(items)
	if err := runtime.UpdateProgress(processed, len(items), progressMessage(job, processed, len(items), "")); err != nil {
		return err
	}

	result := models.BatchImportResult{
		FailedItems: make([]models.BatchImportFailedItem, 0),
	}

	for _, item := range items {
		if item.Status == models.BatchImportItemStatusCompleted {
			result.SuccessCount++
			continue
		}
		if item.Status == models.BatchImportItemStatusFailed {
			result.FailedCount++
			result.FailedItems = append(result.FailedItems, models.BatchImportFailedItem{
				SourcePath:  item.SourcePath,
				DisplayName: item.DisplayName,
				Error:       item.ErrorMessage,
			})
			continue
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if resuming || item.Status == models.BatchImportItemStatusRecovering || item.Status == models.BatchImportItemStatusProcessing {
			if err := runtime.FileService().CleanupBatchImportArtifacts(item.TempPath, item.FinalPath); err != nil {
				return err
			}
		}

		item.Status = models.BatchImportItemStatusProcessing
		item.ErrorMessage = ""
		item.BytesCopied = 0
		item.StartedAt = time.Now().Format(time.RFC3339)
		if err := runtime.JobRepository().UpdateBatchImportItem(item); err != nil {
			return err
		}

		sourcePath := item.SourcePath
		displayName := item.DisplayName
		if strings.TrimSpace(displayName) == "" {
			displayName = filepath.Base(sourcePath)
			item.DisplayName = displayName
		}

		if err := runtime.UpdateProgress(processed, len(items), progressMessage(job, processed, len(items), displayName)); err != nil {
			return err
		}

		plan, err := runtime.FileService().PrepareBatchImport(sourcePath, displayName, payload.CollectionName)
		if err != nil {
			finalizeFailedItem(item, err)
			if updateErr := runtime.JobRepository().UpdateBatchImportItem(item); updateErr != nil {
				return updateErr
			}
			result.FailedCount++
			result.FailedItems = append(result.FailedItems, models.BatchImportFailedItem{
				SourcePath:  sourcePath,
				DisplayName: displayName,
				Error:       err.Error(),
			})
			processed++
			if progressErr := runtime.UpdateProgress(processed, len(items), progressMessage(job, processed, len(items), displayName)); progressErr != nil {
				return progressErr
			}
			continue
		}

		item.DetectedFileType = plan.FileType
		item.TempPath = plan.TempPath
		item.FinalPath = plan.FinalPath
		item.ExpectedSize = plan.FileSize
		item.Checksum = plan.Checksum
		if err := runtime.JobRepository().UpdateBatchImportItem(item); err != nil {
			return err
		}

		copyErr := runtime.FileService().CopyFileToTemp(plan, func(copied int64) error {
			item.BytesCopied = copied
			if err := runtime.JobRepository().UpdateBatchImportItem(item); err != nil {
				return err
			}
			return runtime.UpdateHeartbeat(fmt.Sprintf("正在复制 %s", displayName))
		})
		if copyErr != nil {
			_ = runtime.FileService().CleanupBatchImportArtifacts(plan.TempPath, plan.FinalPath)
			finalizeFailedItem(item, copyErr)
			if err := runtime.JobRepository().UpdateBatchImportItem(item); err != nil {
				return err
			}
			result.FailedCount++
			result.FailedItems = append(result.FailedItems, models.BatchImportFailedItem{
				SourcePath:  sourcePath,
				DisplayName: displayName,
				Error:       copyErr.Error(),
			})
			processed++
			if err := runtime.UpdateProgress(processed, len(items), progressMessage(job, processed, len(items), displayName)); err != nil {
				return err
			}
			continue
		}

		metadata, err := runtime.FileService().ExtractMetadata(plan.TempPath, plan.FileType)
		if err != nil && !os.IsNotExist(err) {
			metadata = models.Metadata{}
		}

		tags, description, err := runtime.FileService().ResolveBatchImportMetadata(BatchImportMetadataRequest{
			FileName:                     displayName,
			FileType:                     plan.FileType,
			SharedTags:                   payload.SharedTags,
			SharedDescription:            payload.SharedDescription,
			EnableAIGeneratedTags:        payload.EnableAIGeneratedTags,
			EnableAIGeneratedDescription: payload.EnableAIGeneratedDescription,
		})
		if err != nil {
			_ = runtime.FileService().CleanupBatchImportArtifacts(plan.TempPath, plan.FinalPath)
			finalizeFailedItem(item, err)
			if err := runtime.JobRepository().UpdateBatchImportItem(item); err != nil {
				return err
			}
			result.FailedCount++
			result.FailedItems = append(result.FailedItems, models.BatchImportFailedItem{
				SourcePath:  sourcePath,
				DisplayName: displayName,
				Error:       err.Error(),
			})
			processed++
			if err := runtime.UpdateProgress(processed, len(items), progressMessage(job, processed, len(items), displayName)); err != nil {
				return err
			}
			continue
		}

		if _, err := runtime.FileService().FinalizeBatchImport(plan, tags, description, metadata); err != nil {
			_ = runtime.FileService().CleanupBatchImportArtifacts(plan.TempPath, plan.FinalPath)
			finalizeFailedItem(item, err)
			if err := runtime.JobRepository().UpdateBatchImportItem(item); err != nil {
				return err
			}
			result.FailedCount++
			result.FailedItems = append(result.FailedItems, models.BatchImportFailedItem{
				SourcePath:  sourcePath,
				DisplayName: displayName,
				Error:       err.Error(),
			})
			processed++
			if err := runtime.UpdateProgress(processed, len(items), progressMessage(job, processed, len(items), displayName)); err != nil {
				return err
			}
			continue
		}

		item.Status = models.BatchImportItemStatusCompleted
		item.ErrorMessage = ""
		item.BytesCopied = plan.FileSize
		item.FinishedAt = time.Now().Format(time.RFC3339)
		if err := runtime.JobRepository().UpdateBatchImportItem(item); err != nil {
			return err
		}

		result.SuccessCount++
		processed++
		if err := runtime.UpdateProgress(processed, len(items), progressMessage(job, processed, len(items), displayName)); err != nil {
			return err
		}
	}

	return runtime.SetResult(result)
}

func finalizeFailedItem(item *models.BatchImportItem, err error) {
	item.Status = models.BatchImportItemStatusFailed
	item.ErrorMessage = err.Error()
	item.FinishedAt = time.Now().Format(time.RFC3339)
}

func countProcessedItems(items []*models.BatchImportItem) int {
	count := 0
	for _, item := range items {
		if item.Status == models.BatchImportItemStatusCompleted || item.Status == models.BatchImportItemStatusFailed {
			count++
		}
	}
	return count
}

func progressMessage(job *models.Job, completed, total int, currentName string) string {
	if currentName == "" {
		return fmt.Sprintf("%s（%d / %d）", job.Title, completed, total)
	}
	return fmt.Sprintf("正在导入 %s（%d / %d）", currentName, completed, total)
}

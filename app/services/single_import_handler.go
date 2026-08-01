package services

import (
	"context"
	"encoding/json"
	"fmt"

	"LocalSpace/app/models"
)

type SingleImportHandler struct{}

func NewSingleImportHandler() *SingleImportHandler {
	return &SingleImportHandler{}
}

func (h *SingleImportHandler) Type() string {
	return models.JobTypeSingleImport
}

func (h *SingleImportHandler) Validate(payload json.RawMessage) error {
	var req ImportFileRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("invalid single import payload: %w", err)
	}
	if req.FilePath == "" {
		return fmt.Errorf("file path is required")
	}
	if req.FileName == "" {
		return fmt.Errorf("file name is required")
	}
	return nil
}

func (h *SingleImportHandler) Execute(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	var req ImportFileRequest
	if err := json.Unmarshal(job.Payload, &req); err != nil {
		return fmt.Errorf("failed to unmarshal single import payload: %w", err)
	}

	fileService := runtime.FileService()
	if fileService == nil {
		return fmt.Errorf("file service is not available")
	}

	_ = runtime.UpdateHeartbeat("正在导入文件")
	if err := fileService.ImportFile(req); err != nil {
		return err
	}

	_ = runtime.UpdateHeartbeat("导入完成")
	return nil
}

func (h *SingleImportHandler) Resume(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	return fmt.Errorf("single import does not support resume")
}

func (h *SingleImportHandler) Cleanup(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	// SingleImportHandler uses FileService.ImportFile which already handles staging cleanup.
	return nil
}

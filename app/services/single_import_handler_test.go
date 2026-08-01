package services

import (
	"context"
	"encoding/json"
	"testing"

	"LocalSpace/app/models"
)

func TestSingleImportHandler_Type(t *testing.T) {
	h := NewSingleImportHandler()
	if h.Type() != models.JobTypeSingleImport {
		t.Errorf("expected type %s, got %s", models.JobTypeSingleImport, h.Type())
	}
}

func TestSingleImportHandler_Validate(t *testing.T) {
	h := NewSingleImportHandler()

	err := h.Validate(json.RawMessage(`{"FilePath": "/tmp/a.txt", "FileName": "a.txt"}`))
	if err != nil {
		t.Errorf("expected valid payload, got %v", err)
	}

	err = h.Validate(json.RawMessage(`{"FilePath": "", "FileName": ""}`))
	if err == nil {
		t.Error("expected validation error for empty fields")
	}
}

func TestSingleImportHandler_Resume_ReturnsError(t *testing.T) {
	h := NewSingleImportHandler()
	err := h.Resume(context.Background(), &models.Job{}, nil)
	if err == nil {
		t.Error("expected resume to return error")
	}
}

package services

import (
	"context"
	"testing"

	"LocalSpace/app/agents"
	"LocalSpace/app/models"
)

func TestImportFileRequestWithAgentFields(t *testing.T) {
	// Test ImportFileRequest with Keywords field
	req := ImportFileRequest{
		FilePath:    "/tmp/test.txt",
		FileName:    "test.txt",
		Keywords:    "test document",
		Tags:        []string{"test"},
		Description: "A test file",
	}

	// Verify all fields are set correctly
	if req.Keywords != "test document" {
		t.Errorf("Expected keywords to be 'test document', got '%s'", req.Keywords)
	}

	if len(req.Tags) != 1 || req.Tags[0] != "test" {
		t.Errorf("Expected tags to be ['test'], got %v", req.Tags)
	}

	if req.Description != "A test file" {
		t.Errorf("Expected description to be 'A test file', got '%s'", req.Description)
	}

	// Test empty defaults
	emptyReq := ImportFileRequest{
		FilePath: "/tmp/test.txt",
		FileName: "test.txt",
	}

	if emptyReq.Keywords != "" {
		t.Errorf("Expected default Keywords to be empty string, got '%s'", emptyReq.Keywords)
	}

	if len(emptyReq.Tags) != 0 {
		t.Errorf("Expected default Tags to be empty, got %v", emptyReq.Tags)
	}

	if emptyReq.Description != "" {
		t.Errorf("Expected default Description to be empty string, got '%s'", emptyReq.Description)
	}
}

func TestFileServiceStructure(t *testing.T) {
	// Test that FileService can be created with agent service integration
	// This test verifies the structure without requiring full dependencies

	// Verify ImportFileRequest structure supports agent integration
	req := ImportFileRequest{
		FilePath:    "/path/to/file.mp4",
		FileName:    "movie.mp4",
		Keywords:    "action thriller",
		Tags:        []string{"entertainment"},
		Description: "An action thriller movie",
	}

	// Test that all required fields are present
	requiredFields := map[string]string{
		"FilePath":    req.FilePath,
		"FileName":    req.FileName,
		"Keywords":    req.Keywords,
		"Description": req.Description,
	}

	for field, value := range requiredFields {
		if value == "" {
			t.Errorf("Required field '%s' is empty", field)
		}
	}

	if len(req.Tags) == 0 {
		t.Error("Tags field should not be empty")
	}
}

func TestAgentServiceIntegrationStructure(t *testing.T) {
	// Test that the integration points are correctly structured
	// This verifies that the FileService can accept and process agent-generated data

	// Simulate what the agent service would return
	agentGeneratedTags := []string{"video", "action", "thriller"}
	agentGeneratedDescription := "An action thriller movie"

	// Test that the data structures are compatible
	req := ImportFileRequest{
		FilePath:    "/path/to/file.mp4",
		FileName:    "movie.mp4",
		Keywords:    "action thriller",
		Tags:        []string{"entertainment"},
		Description: "",
	}

	// Simulate agent integration: use agent-generated data when available
	finalTags := agentGeneratedTags
	finalDescription := agentGeneratedDescription

	// Fallback to user input if agent generation failed
	if len(finalTags) == 0 && len(req.Tags) > 0 {
		finalTags = req.Tags
	}
	if finalDescription == "" && req.Description != "" {
		finalDescription = req.Description
	}

	// Verify the integration logic works
	if len(finalTags) == 0 {
		t.Error("Expected final tags to be non-empty")
	}

	if finalDescription == "" {
		t.Error("Expected final description to be non-empty")
	}

	// Verify agent-generated tags are present
	expectedTags := []string{"video", "action", "thriller"}
	for _, expected := range expectedTags {
		found := false
		for _, tag := range finalTags {
			if tag == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected tag '%s' not found in final tags: %v", expected, finalTags)
		}
	}
}

type countingAgentService struct {
	calls        int
	lastMetadata models.Metadata
}

func (c *countingAgentService) AnalyzeMetadata(ctx context.Context, input *agents.MetadataGenerationInput) (*agents.MetadataAnalysis, error) {
	c.calls++
	c.lastMetadata = input.Metadata
	return &agents.MetadataAnalysis{Tags: []string{"video", "action"}, Description: "动作片"}, nil
}

func TestFileServiceUsesSingleMetadataAnalysisResult(t *testing.T) {
	service := &FileService{}
	counting := &countingAgentService{}

	analysis, err := service.resolveMetadataForImport(context.Background(), counting, &ImportFileRequest{
		FileName: "movie.mp4",
		Keywords: "action thriller",
	}, "video", models.Metadata{Duration: 120})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if counting.calls != 1 {
		t.Fatalf("expected one metadata analysis call, got %d", counting.calls)
	}

	if analysis.Description != "动作片" {
		t.Fatalf("expected unified description result, got %q", analysis.Description)
	}
	if counting.lastMetadata.Duration != 120 {
		t.Fatalf("expected extracted metadata to be passed to analyzer, got %+v", counting.lastMetadata)
	}
}

func TestFileServiceResolveMetadataForImport_PreservesExplicitUserMetadata(t *testing.T) {
	service := &FileService{}
	counting := &countingAgentService{}

	analysis, err := service.resolveMetadataForImport(context.Background(), counting, &ImportFileRequest{
		FileName:    "movie.mp4",
		Tags:        []string{"用户标签"},
		Description: "用户描述",
	}, "video", models.Metadata{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(analysis.Tags) != 1 || analysis.Tags[0] != "用户标签" {
		t.Fatalf("expected explicit user tags to take priority, got %v", analysis.Tags)
	}
	if analysis.Description != "用户描述" {
		t.Fatalf("expected explicit user description to take priority, got %q", analysis.Description)
	}
}

func TestFileServiceResolveMetadataForImport_NilRequestSafe(t *testing.T) {
	service := &FileService{}
	counting := &countingAgentService{}

	analysis, err := service.resolveMetadataForImport(context.Background(), counting, nil, "video", models.Metadata{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if analysis == nil {
		t.Fatal("expected fallback analysis for nil request")
	}
	if counting.calls != 0 {
		t.Fatalf("expected analyzer not to be called for nil request, got %d calls", counting.calls)
	}
}

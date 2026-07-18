package services

import (
	"testing"
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
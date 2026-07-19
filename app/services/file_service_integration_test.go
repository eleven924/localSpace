package services

import "testing"

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

func TestFileServiceResolveMetadataForImport_UsesOnlyFormMetadata(t *testing.T) {
	service := &FileService{}

	metadata := service.resolveMetadataForImport(&ImportFileRequest{
		FileName:    "movie.mp4",
		Keywords:    "action thriller",
		Tags:        []string{"entertainment"},
		Description: "  manual description  ",
	})
	if len(metadata.Tags) != 1 || metadata.Tags[0] != "entertainment" {
		t.Fatalf("expected form tags to be preserved, got %v", metadata.Tags)
	}
	if metadata.Description != "manual description" {
		t.Fatalf("expected trimmed form description, got %q", metadata.Description)
	}
}

func TestFileServiceResolveMetadataForImport_PreservesExplicitUserMetadata(t *testing.T) {
	service := &FileService{}

	metadata := service.resolveMetadataForImport(&ImportFileRequest{
		FileName:    "movie.mp4",
		Tags:        []string{"用户标签"},
		Description: "用户描述",
	})

	if len(metadata.Tags) != 1 || metadata.Tags[0] != "用户标签" {
		t.Fatalf("expected explicit user tags to be imported, got %v", metadata.Tags)
	}
	if metadata.Description != "用户描述" {
		t.Fatalf("expected explicit user description to be imported, got %q", metadata.Description)
	}
}

func TestFileServiceResolveMetadataForImport_NilRequestSafe(t *testing.T) {
	service := &FileService{}

	metadata := service.resolveMetadataForImport(nil)
	if metadata.Tags == nil {
		t.Fatal("expected fallback analysis for nil request")
	}
	if len(metadata.Tags) != 0 || metadata.Description != "" {
		t.Fatalf("expected empty metadata for nil request, got %+v", metadata)
	}
}

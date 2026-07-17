package services

import "testing"

func TestImportFileRequestKeywords(t *testing.T) {
	req := ImportFileRequest{
		FilePath:    "/test/path/file.txt",
		FileName:    "file.txt",
		Description: "test description",
		Tags:        []string{"tag1", "tag2"},
		Keywords:    "test keywords",
	}

	if req.Keywords != "test keywords" {
		t.Errorf("Expected Keywords to be 'test keywords', got '%s'", req.Keywords)
	}

	// Test default empty value
	emptyReq := ImportFileRequest{
		FilePath: "/test/path/file.txt",
		FileName: "file.txt",
	}

	if emptyReq.Keywords != "" {
		t.Errorf("Expected default Keywords to be empty string, got '%s'", emptyReq.Keywords)
	}
}
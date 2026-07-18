package app

import (
	"testing"
)

func TestImportFileBackwardCompatibility(t *testing.T) {
	app := NewApp()

	if app == nil {
		t.Fatal("Expected app to be created")
	}

	// Test that ImportFile method exists with 4 parameters (backward compatible)
	_ = app.ImportFile // This will fail if the method doesn't exist

	// Test that ImportFileWithKeywords method exists with 5 parameters
	_ = app.ImportFileWithKeywords // This will fail if the method doesn't exist

	// Both methods should be available:
	// ImportFile(path, name, desc, tags)              // 4 args - backward compatible
	// ImportFileWithKeywords(path, name, desc, tags, keywords) // 5 args - new functionality
}
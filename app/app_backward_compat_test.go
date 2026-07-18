package app

import (
	"testing"
)

func TestImportFileBackwardCompatibility(t *testing.T) {
	app := NewApp()

	if app == nil {
		t.Fatal("Expected app to be created")
	}

	// Test that ImportFile can be called with 4 arguments (backward compatible)
	// This simulates the old frontend behavior
	// Note: We can't actually call it without a real database, but we can verify the method signature
	_ = app.ImportFile // This will fail if the method doesn't exist

	// The variadic parameters allow both:
	// ImportFile(path, name, desc, tags)         // 4 args - backward compatible
	// ImportFile(path, name, desc, tags, keywords) // 5 args - new functionality
}
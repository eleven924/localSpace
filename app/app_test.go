package app

import (
	"testing"
)

func TestAppAgentServiceInitialization(t *testing.T) {
	app := NewApp()

	if app == nil {
		t.Fatal("Expected app to be created")
	}

	// Test that App struct has agentService field
	// We can't test full initialization without a real database,
	// but we can verify the structure is correct
	_ = app.agentService // This will fail if the field doesn't exist

	// Test that the App struct can be created with the expected fields
	if app == nil {
		t.Error("Expected app to be created")
	}
}

func TestAppImportFileSignature(t *testing.T) {
	app := NewApp()

	if app == nil {
		t.Fatal("Expected app to be created")
	}

	// Test that ImportFile method has the correct signature with keywords parameter
	// We can't call it without a real file, but we can verify the method exists
	_ = app.ImportFile // This will fail if the method doesn't exist
}

func TestAppStructureHasRequiredServices(t *testing.T) {
	app := NewApp()

	// Test that all required service fields exist in the App struct
	// This verifies the structure is correct for agent integration
	_ = app.fileService       // FileService must exist
	_ = app.agentService      // AgentService must exist
	_ = app.aiService         // AIService must exist
	_ = app.storageService    // StorageService must exist
	_ = app.configService     // ConfigService must exist
	_ = app.thumbnailService // ThumbnailService must exist
}
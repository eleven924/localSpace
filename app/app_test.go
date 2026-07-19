package app

import (
	"testing"
	"time"

	"LocalSpace/app/services"
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
	_ = app.fileService      // FileService must exist
	_ = app.agentService     // AgentService must exist
	_ = app.aiService        // AIService must exist
	_ = app.storageService   // StorageService must exist
	_ = app.configService    // ConfigService must exist
	_ = app.thumbnailService // ThumbnailService must exist
}

func TestWaitForInitializationReturnsTrueWhenServicesBecomeReady(t *testing.T) {
	app := NewApp()

	go func() {
		time.Sleep(100 * time.Millisecond)
		app.fileService = &services.FileService{}
		app.configService = &services.ConfigService{}
	}()

	start := time.Now()
	if !app.waitForInitialization(500 * time.Millisecond) {
		t.Fatal("expected waitForInitialization to detect initialized services")
	}

	if waited := time.Since(start); waited < 100*time.Millisecond {
		t.Fatalf("expected waitForInitialization to wait for startup, only waited %v", waited)
	}
}

func TestWaitForInitializationTimesOutWhenServicesStayNil(t *testing.T) {
	app := NewApp()

	start := time.Now()
	if app.waitForInitialization(120 * time.Millisecond) {
		t.Fatal("expected waitForInitialization to time out")
	}

	if waited := time.Since(start); waited < 100*time.Millisecond {
		t.Fatalf("expected waitForInitialization to keep waiting before timing out, only waited %v", waited)
	}
}

package services

import (
	"testing"

	"LocalSpace/app/repositories"
)

func TestNewAgentService(t *testing.T) {
	// Create a mock database wrapper for testing
	// In real implementation, this would use a proper mock interface
	dbWrapper := &repositories.SQLiteDBWrapper{}

	configRepo := repositories.NewConfigRepository(dbWrapper)

	service := NewAgentService(configRepo)
	if service == nil {
		t.Fatal("Expected agent service to be created")
	}

	if service.tagAgent == nil {
		t.Error("Expected tag agent to be initialized")
	}

	if service.descriptionAgent == nil {
		t.Error("Expected description agent to be initialized")
	}
}

// Note: Tests that require database access are skipped for now
// In a real implementation, we would use a proper mock interface or test database
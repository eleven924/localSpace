package agents

import (
	"context"
	"testing"
	"time"
)

func TestBaseAgent(t *testing.T) {
	// Test creating a base agent
	config := &AgentConfig{
		Name:    "test-agent",
		Timeout: 30 * time.Second,
	}

	agent := NewBaseAgent(config)
	if agent == nil {
		t.Fatal("Expected agent to be created")
	}

	if agent.Name() != "test-agent" {
		t.Errorf("Expected agent name to be 'test-agent', got '%s'", agent.Name())
	}

	// Test context with timeout
	ctx, cancel := agent.CreateContext(context.Background())
	defer cancel()

	if ctx == nil {
		t.Error("Expected context to be created")
	}

	select {
	case <-ctx.Done():
	case <-time.After(35 * time.Second):
		t.Error("Expected context to timeout within 30 seconds")
	}
}
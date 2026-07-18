package tools_test

import (
	"context"
	"testing"

	"LocalSpace/app/tools"
)

func TestToolRegistry(t *testing.T) {
	registry := tools.NewToolRegistry()

	if registry == nil {
		t.Fatal("Expected registry to be created")
	}

	mockTool := &MockTool{name: "test-tool"}
	err := registry.Register("test-tool", mockTool)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	tool, err := registry.Get("test-tool")
	if err != nil {
		t.Fatalf("Failed to get tool: %v", err)
	}

	if tool.Name() != "test-tool" {
		t.Errorf("Expected tool name to be 'test-tool', got '%s'", tool.Name())
	}

	_, err = registry.Get("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent tool")
	}
}

type MockTool struct {
	name string
}

func (m *MockTool) Name() string {
	return m.name
}

func (m *MockTool) Description() string {
	return "Mock tool for testing"
}

func (m *MockTool) Execute(ctx context.Context, input string) (string, error) {
	return "mock result", nil
}

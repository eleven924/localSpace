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

	tool, ok := registry.Get("test-tool")
	if !ok {
		t.Fatal("Failed to get tool")
	}

	if tool.Name() != "test-tool" {
		t.Errorf("Expected tool name to be 'test-tool', got '%s'", tool.Name())
	}

	_, ok = registry.Get("non-existent")
	if ok {
		t.Error("Expected missing tool lookup to return ok=false")
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

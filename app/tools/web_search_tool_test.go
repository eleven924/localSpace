package tools

import (
	"context"
	"testing"
	"time"
)

func TestWebSearchTool(t *testing.T) {
	tool := NewWebSearchTool(10 * time.Second)

	// Test tool properties
	if tool.Name() != "web_search" {
		t.Errorf("Expected tool name to be 'web_search', got '%s'", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Expected tool description to be non-empty")
	}

	// Test search functionality (with mock)
	// Note: This test would require mocking the search client
	// For now, test error handling with invalid input
	ctx := context.Background()

	// Test with empty query
	_, err := tool.Search(ctx, "")
	if err == nil {
		t.Error("Expected error for empty query")
	}

	// Test with query length validation
	longQuery := string(make([]byte, 501))
	_, err = tool.Search(ctx, longQuery)
	if err == nil {
		t.Error("Expected error for query that's too long")
	}
}
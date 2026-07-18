package tools

import (
	"context"
	"testing"
	"time"
)

func TestWebSearchTool(t *testing.T) {
	tool := NewWebSearchTool(10 * time.Second)

	if tool.Name() != "web_search" {
		t.Errorf("Expected tool name to be 'web_search', got '%s'", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Expected tool description to be non-empty")
	}

	ctx := context.Background()

	_, err := tool.Search(ctx, "")
	if err == nil {
		t.Error("Expected error for empty query")
	}

	longQuery := string(make([]byte, 501))
	_, err = tool.Search(ctx, longQuery)
	if err == nil {
		t.Error("Expected error for query that's too long")
	}
}

func TestWebSearchToolExecuteBoundsResults(t *testing.T) {
	tool := NewWebSearchTool(10 * time.Second)

	result, err := tool.Execute(context.Background(), "movie")
	if err != nil {
		t.Fatalf("execute returned error: %v", err)
	}

	if result != "query=movie results=0" {
		t.Fatalf("expected bounded execute summary, got %q", result)
	}
}

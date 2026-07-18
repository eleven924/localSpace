package agents

import (
	"context"
	"testing"

	"LocalSpace/app/models"
	"LocalSpace/app/tools"

	"github.com/cloudwego/eino/schema"
)

type runtimeTestTool struct {
	name   string
	calls  []string
	result string
}

func (t *runtimeTestTool) Name() string        { return t.name }
func (t *runtimeTestTool) Description() string { return "test tool" }
func (t *runtimeTestTool) Execute(ctx context.Context, input string) (string, error) {
	t.calls = append(t.calls, input)
	return t.result, nil
}

func TestExecuteToolCallRecordsUsageAndSearchQuery(t *testing.T) {
	tool := &runtimeTestTool{name: "web_search", result: "query: movie\nresults:\n1. Title: Movie"}
	runtime := NewDefaultAgentRuntime()

	output, toolsUsed, searchQueries, err := runtime.executeToolCall(context.Background(), map[string]tools.Tool{
		"web_search": tool,
	}, "web_search", "movie", nil, nil)
	if err != nil {
		t.Fatalf("executeToolCall returned error: %v", err)
	}
	if output != "query: movie\nresults:\n1. Title: Movie" {
		t.Fatalf("unexpected tool output %q", output)
	}
	if len(toolsUsed) != 1 || toolsUsed[0] != "web_search" {
		t.Fatalf("expected web_search recorded in ToolsUsed, got %v", toolsUsed)
	}
	if len(searchQueries) != 1 || searchQueries[0] != "movie" {
		t.Fatalf("expected movie recorded in SearchQueries, got %v", searchQueries)
	}
}

func TestDefaultAgentRuntimeRun_WithoutToolsReturnsDirectOutput(t *testing.T) {
	runtime := &DefaultAgentRuntime{
		generate: func(ctx context.Context, messages []*schema.Message, req *AgentRunRequest) (*schema.Message, error) {
			return schema.AssistantMessage(`{"tags":["video"],"description":"影片"}`, nil), nil
		},
	}

	resp, err := runtime.Run(context.Background(), &AgentRunRequest{
		SystemPrompt: "sys",
		UserPrompt:   "user",
		AIConfig:     &models.AIConfig{APIKey: "key", Model: "model", BaseURL: "https://api.example.com/v1"},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if resp.Output != `{"tags":["video"],"description":"影片"}` {
		t.Fatalf("unexpected output %q", resp.Output)
	}
	if len(resp.ToolsUsed) != 0 {
		t.Fatalf("expected no tools used, got %v", resp.ToolsUsed)
	}
}

func TestDefaultAgentRuntimeRun_ExecutesToolLoop(t *testing.T) {
	tool := &runtimeTestTool{name: "web_search", result: "query: movie\nresults:\n1. Title: Movie"}
	calls := 0
	runtime := &DefaultAgentRuntime{
		generate: func(ctx context.Context, messages []*schema.Message, req *AgentRunRequest) (*schema.Message, error) {
			calls++
			if calls == 1 {
				return schema.AssistantMessage(`{"tool":"web_search","input":"movie"}`, nil), nil
			}
			return schema.AssistantMessage(`{"tags":["video","movie"],"description":"Movie metadata"}`, nil), nil
		},
	}

	resp, err := runtime.Run(context.Background(), &AgentRunRequest{
		SystemPrompt: "sys",
		UserPrompt:   "user",
		Tools:        []tools.Tool{tool},
		AIConfig:     &models.AIConfig{APIKey: "key", Model: "model", BaseURL: "https://api.example.com/v1"},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if resp.Output != `{"tags":["video","movie"],"description":"Movie metadata"}` {
		t.Fatalf("unexpected output %q", resp.Output)
	}
	if len(resp.ToolsUsed) != 1 || resp.ToolsUsed[0] != "web_search" {
		t.Fatalf("expected web_search in ToolsUsed, got %v", resp.ToolsUsed)
	}
	if len(resp.SearchQueries) != 1 || resp.SearchQueries[0] != "movie" {
		t.Fatalf("expected movie in SearchQueries, got %v", resp.SearchQueries)
	}
}

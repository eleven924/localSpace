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

func TestDecodeToolInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "json envelope", input: `{"input":"movie"}`, want: "movie"},
		{name: "plain string fallback", input: "movie", want: "movie"},
		{name: "missing input", input: `{"query":"movie"}`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeToolInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("decodeToolInput() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Fatalf("decodeToolInput() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDefaultAgentRuntimeRun_WithoutToolsReturnsDirectOutput(t *testing.T) {
	runtime := &DefaultAgentRuntime{
		generate: func(ctx context.Context, messages []*schema.Message, req *AgentRunRequest, toolInfos []*schema.ToolInfo) (*schema.Message, error) {
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
	tool := &runtimeTestTool{name: "web_search", result: `{"type":"search_results","query":"movie","results":[{"title":"Movie","url":"http://example.com","snippet":"A movie"}]}`}
	calls := 0
	runtime := &DefaultAgentRuntime{
		generate: func(ctx context.Context, messages []*schema.Message, req *AgentRunRequest, toolInfos []*schema.ToolInfo) (*schema.Message, error) {
			calls++
			if calls == 1 {
				return &schema.Message{
					Role: schema.Assistant,
					ToolCalls: []schema.ToolCall{{
						ID:   "call_123",
						Type: "function",
						Function: schema.FunctionCall{
							Name:      "web_search",
							Arguments: `{"input":"movie"}`,
						},
					}},
				}, nil
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
	if len(tool.calls) != 1 || tool.calls[0] != "movie" {
		t.Fatalf("expected decoded tool input, got %v", tool.calls)
	}
}

func TestDefaultAgentRuntimeRun_LimitsWebSearchToTwoCalls(t *testing.T) {
	tool := &runtimeTestTool{name: "web_search", result: `{"type":"search_results","query":"movie","results":[]}`}
	calls := 0
	runtime := &DefaultAgentRuntime{
		generate: func(ctx context.Context, messages []*schema.Message, req *AgentRunRequest, toolInfos []*schema.ToolInfo) (*schema.Message, error) {
			calls++
			switch calls {
			case 1:
				return &schema.Message{Role: schema.Assistant, ToolCalls: []schema.ToolCall{{ID: "call_1", Type: "function", Function: schema.FunctionCall{Name: "web_search", Arguments: `{"input":"movie one"}`}}}}, nil
			case 2:
				return &schema.Message{Role: schema.Assistant, ToolCalls: []schema.ToolCall{{ID: "call_2", Type: "function", Function: schema.FunctionCall{Name: "web_search", Arguments: `{"input":"movie two"}`}}}}, nil
			default:
				if len(toolInfos) != 0 {
					return &schema.Message{Role: schema.Assistant, ToolCalls: []schema.ToolCall{{ID: "call_3", Type: "function", Function: schema.FunctionCall{Name: "web_search", Arguments: `{"input":"movie three"}`}}}}, nil
				}
				return schema.AssistantMessage(`{"tags":["video"],"description":"Final metadata"}`, nil), nil
			}
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
	if resp.Output != `{"tags":["video"],"description":"Final metadata"}` {
		t.Fatalf("unexpected output %q", resp.Output)
	}
	if len(tool.calls) != 2 {
		t.Fatalf("expected exactly two web_search executions, got %v", tool.calls)
	}
	if len(resp.SearchQueries) != 2 || resp.SearchQueries[0] != "movie one" || resp.SearchQueries[1] != "movie two" {
		t.Fatalf("unexpected search queries %v", resp.SearchQueries)
	}
}

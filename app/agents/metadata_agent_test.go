package agents

import (
	"context"
	"strings"
	"testing"

	"LocalSpace/app/models"
	"LocalSpace/app/tools"
)

type fakeRuntime struct {
	response *AgentRunResponse
	err      error
}

func (f *fakeRuntime) Run(ctx context.Context, req *AgentRunRequest) (*AgentRunResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.response, nil
}

type fakeTool struct {
	name string
}

func (f *fakeTool) Name() string        { return f.name }
func (f *fakeTool) Description() string { return "fake tool" }
func (f *fakeTool) Execute(ctx context.Context, input string) (string, error) {
	return "", nil
}

func TestEinoMetadataAgentAnalyze_PropagatesRuntimeToolTrace(t *testing.T) {
	agent := NewEinoMetadataAgent(&fakeRuntime{
		response: &AgentRunResponse{
			Output:        `{"tags":["video","movie"],"description":"Movie metadata"}`,
			ToolsUsed:     []string{"web_search"},
			SearchQueries: []string{"movie"},
		},
	})

	result, err := agent.Analyze(
		context.Background(),
		&MetadataGenerationInput{FileName: "movie.mp4", FileType: "video"},
		&models.AIConfig{Enabled: true, EnableAgent: true, APIKey: "key", Model: "model", BaseURL: "https://api.example.com/v1"},
		[]tools.Tool{&fakeTool{name: "web_search"}},
	)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}
	if len(result.Trace.ToolsAvailable) != 1 || result.Trace.ToolsAvailable[0] != "web_search" {
		t.Fatalf("expected ToolsAvailable to include web_search, got %v", result.Trace.ToolsAvailable)
	}
	if len(result.Trace.ToolsUsed) != 1 || result.Trace.ToolsUsed[0] != "web_search" {
		t.Fatalf("expected ToolsUsed to include web_search, got %v", result.Trace.ToolsUsed)
	}
	if len(result.Trace.SearchQueries) != 1 || result.Trace.SearchQueries[0] != "movie" {
		t.Fatalf("expected SearchQueries to include movie, got %v", result.Trace.SearchQueries)
	}
}

func TestEinoMetadataAgentAnalyze_ParsesUnicodeAndWrappedJSON(t *testing.T) {
	agent := NewEinoMetadataAgent(&fakeRuntime{
		response: &AgentRunResponse{
			Output:        "```json\n{\"tags\":[\"drama\",\"æon\"],\"description\":\"彭昱畅 æ 冒险\"}\n```",
			ToolsUsed:     []string{"web_search"},
			SearchQueries: []string{"movie"},
		},
	})

	result, err := agent.Analyze(
		context.Background(),
		&MetadataGenerationInput{FileName: "movie.mp4", FileType: "video"},
		&models.AIConfig{Enabled: true, EnableAgent: true, APIKey: "key", Model: "model", BaseURL: "https://api.example.com/v1"},
		[]tools.Tool{&fakeTool{name: "web_search"}},
	)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}
	if result.Analysis == nil {
		t.Fatalf("expected analysis")
	}
	if len(result.Analysis.Tags) != 2 || result.Analysis.Tags[1] != "æon" {
		t.Fatalf("expected unicode tags preserved, got %v", result.Analysis.Tags)
	}
	if result.Analysis.Description != "彭昱畅 æ 冒险" {
		t.Fatalf("expected unicode description preserved, got %q", result.Analysis.Description)
	}
	if result.Trace == nil || result.Trace.RawOutput != strings.TrimSpace("```json\n{\"tags\":[\"drama\",\"æon\"],\"description\":\"彭昱畅 æ 冒险\"}\n```") {
		t.Fatalf("expected raw output preserved, got %#v", result.Trace)
	}
}

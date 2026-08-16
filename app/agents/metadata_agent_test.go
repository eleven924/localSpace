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

type sequenceRuntime struct {
	responses []*AgentRunResponse
	requests  []*AgentRunRequest
}

func (r *sequenceRuntime) Run(ctx context.Context, req *AgentRunRequest) (*AgentRunResponse, error) {
	r.requests = append(r.requests, req)
	index := len(r.requests) - 1
	if index >= len(r.responses) {
		return nil, nil
	}
	return r.responses[index], nil
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
			ToolCalls:     []models.AIToolCall{{Name: "web_search", Input: "movie", Output: `{"results":[]}`, Status: "success"}},
		},
	})

	result, err := agent.Analyze(
		context.Background(),
		&MetadataGenerationInput{
			FileName: "movie.mp4",
			FileType: "video",
			Evidence: []tools.EvidenceItem{{Source: "local:file_metadata", Kind: "file_metadata"}},
		},
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
	if len(result.Trace.EvidenceSources) != 1 || result.Trace.EvidenceSources[0] != "local:file_metadata" {
		t.Fatalf("expected evidence source in trace, got %v", result.Trace.EvidenceSources)
	}
}

func TestEinoMetadataAgentAnalyze_PreservesToolTraceWhenOutputCannotBeParsed(t *testing.T) {
	agent := NewEinoMetadataAgent(&fakeRuntime{
		response: &AgentRunResponse{
			Output:        "",
			ToolsUsed:     []string{"web_search"},
			SearchQueries: []string{"shameless us season 1"},
			ToolCalls:     []models.AIToolCall{{Name: "web_search", Input: "shameless us season 1", Output: `{"results":[{"title":"Shameless"}]}`, Status: "success"}},
		},
	})

	result, err := agent.Analyze(
		context.Background(),
		&MetadataGenerationInput{FileName: "Shameless.mp4", FileType: "video"},
		&models.AIConfig{Enabled: true, EnableAgent: true, APIKey: "key", Model: "model", BaseURL: "https://api.example.com/v1"},
		[]tools.Tool{&fakeTool{name: "web_search"}},
	)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}
	if result == nil || result.Trace == nil || len(result.Trace.ToolCalls) != 1 {
		t.Fatalf("expected tool trace to survive parse fallback, got %#v", result)
	}
	if result.Trace.FallbackReason == "" || result.Analysis == nil || result.Analysis.Description == "" {
		t.Fatalf("expected a reviewable fallback result, got %#v", result)
	}
}

func TestEinoMetadataAgentAnalyze_RepairsMalformedOutputUsingToolEvidence(t *testing.T) {
	runtime := &sequenceRuntime{responses: []*AgentRunResponse{
		{
			Output:    "搜索结果显示这是一部电视剧。",
			ToolsUsed: []string{"web_search"},
			ToolCalls: []models.AIToolCall{{
				Name:   "web_search",
				Input:  "shameless us season 1",
				Output: `{"type":"search_results","results":[{"title":"Shameless"}]}`,
				Status: "success",
			}},
		},
		{
			Output: `{"tags":["电视剧","喜剧","剧情"],"description":"Shameless 是一部喜剧剧情电视剧。"}`,
		},
	}}
	agent := NewEinoMetadataAgent(runtime)

	result, err := agent.Analyze(
		context.Background(),
		&MetadataGenerationInput{FileName: "Shameless.mp4", FileType: "video"},
		&models.AIConfig{Enabled: true, EnableAgent: true, APIKey: "key", Model: "model", BaseURL: "https://api.example.com/v1"},
		[]tools.Tool{&fakeTool{name: "web_search"}},
	)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}
	if result == nil || result.Analysis == nil || result.Analysis.Description != "Shameless 是一部喜剧剧情电视剧。" {
		t.Fatalf("expected repaired analysis, got %#v", result)
	}
	if result.Trace == nil || result.Trace.FallbackReason != "" {
		t.Fatalf("expected repaired result without fallback, got %#v", result.Trace)
	}
	if len(runtime.requests) != 2 {
		t.Fatalf("expected initial request and one repair request, got %d", len(runtime.requests))
	}
	if len(runtime.requests[1].Tools) != 0 || !runtime.requests[1].StructuredOutput || !strings.Contains(runtime.requests[1].UserPrompt, "Shameless") {
		t.Fatalf("expected repair request to include tool evidence and no tools, got %#v", runtime.requests[1])
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

func TestBuildMetadataUserPromptIncludesSanitizedNativeMetadata(t *testing.T) {
	prompt := buildMetadataUserPrompt(&AnalysisRequest{
		FileName:       "photo.jpg",
		FileType:       "image",
		FileSubType:    "jpg",
		NativeMetadata: models.Metadata{Width: 1920, Height: 1080, Size: 2048},
	}, true)

	if !strings.Contains(prompt, `"width":1920`) || !strings.Contains(prompt, `"height":1080`) {
		t.Fatalf("expected native metadata in prompt, got %q", prompt)
	}
	if strings.Contains(prompt, "必须使用web_search") || strings.Contains(prompt, "请使用web_search工具") {
		t.Fatalf("prompt should not force web search, got %q", prompt)
	}
}

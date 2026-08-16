package services

import (
	"context"
	"errors"
	"testing"

	"LocalSpace/app/agents"
	"LocalSpace/app/models"
	"LocalSpace/app/tools"
)

type fakeMetadataAgent struct {
	result *agents.MetadataAnalysisResult
	err    error
}

func (f *fakeMetadataAgent) Analyze(ctx context.Context, input *agents.MetadataGenerationInput, aiConfig *models.AIConfig, availableTools []tools.Tool) (*agents.MetadataAnalysisResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.result != nil {
		return f.result, nil
	}
	return &agents.MetadataAnalysisResult{Analysis: &agents.MetadataAnalysis{Tags: []string{"video"}, Description: "metadata"}}, nil
}

type errorTool struct {
	name string
}

func (e *errorTool) Name() string        { return e.name }
func (e *errorTool) Description() string { return "error tool" }
func (e *errorTool) Execute(ctx context.Context, input string) (string, error) {
	return "", errors.New("search backend unavailable")
}

func TestNewAgentService(t *testing.T) {
	service := NewAgentService(nil)
	if service == nil {
		t.Fatal("expected service")
	}
	if service.toolRegistry == nil {
		t.Fatal("expected tool registry")
	}
	if service.metadataGraph == nil {
		t.Fatal("expected metadata graph")
	}
	if len(service.toolRegistry.List()) != 0 {
		t.Fatalf("expected no eagerly registered tools, got %v", service.toolRegistry.List())
	}
}

func TestAgentServiceAnalyzeMetadataUsesGraphWhenConfigured(t *testing.T) {
	agent := &fakeMetadataAgent{result: &agents.MetadataAnalysisResult{
		Analysis: &agents.MetadataAnalysis{Tags: []string{"Document"}, Description: "from graph"},
	}}
	service := &AgentService{
		metadataAgent: agent,
		metadataGraph: agents.NewMetadataAnalysisGraph(agent),
		toolRegistry:  tools.NewToolRegistry(),
	}

	result, err := service.analyzeMetadataWithConfig(
		context.Background(),
		&agents.AnalysisRequest{FileName: "notes.md", FileType: "document"},
		&models.AIConfig{Enabled: true, EnableAgent: true, Model: "test-model"},
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result == nil || result.Trace == nil {
		t.Fatal("expected graph result with trace")
	}
	if result.Trace.CurrentStage != "quality_needs_review" {
		t.Fatalf("expected graph quality stage, got %q", result.Trace.CurrentStage)
	}
	if len(result.Trace.StageHistory) == 0 {
		t.Fatal("expected graph stage history")
	}
}

func TestResolveMetadataToolsUsesConfiguredWebSearchTool(t *testing.T) {
	service := &AgentService{toolRegistry: tools.NewToolRegistry()}
	config := &models.AIConfig{
		Enabled:             true,
		EnableAgent:         true,
		EnableWebSearch:     true,
		WebSearchProvider:   "mock-http",
		WebSearchBaseURL:    "https://search.example.com",
		WebSearchAPIKey:     "search-key",
		WebSearchTimeout:    10,
		WebSearchMaxResults: 3,
	}

	service.ensureConfiguredTools(config)
	resolved := service.resolveMetadataTools(config, &agents.MetadataGenerationInput{FileName: "MyAwesomeMovie.mp4", FileType: "video"})
	if len(resolved) != 1 {
		t.Fatalf("expected one resolved tool, got %d", len(resolved))
	}
	if resolved[0].Name() != "web_search" {
		t.Fatalf("expected web_search tool, got %q", resolved[0].Name())
	}
}

func TestResolveMetadataToolsReturnsNoneWhenDisabled(t *testing.T) {
	service := &AgentService{toolRegistry: tools.NewToolRegistry()}
	service.ensureConfiguredTools(&models.AIConfig{})

	resolved := service.resolveMetadataTools(&models.AIConfig{EnableAgent: true, EnableWebSearch: false}, &agents.MetadataGenerationInput{FileName: "movie.mp4", FileType: "video"})
	if len(resolved) != 0 {
		t.Fatalf("expected no resolved tools, got %d", len(resolved))
	}
}

func TestAgentServiceAnalyzeMetadata_FallbackIncludesResolvedToolsWhenRuntimeUsesFailingTool(t *testing.T) {
	service := &AgentService{
		metadataAgent: &fakeMetadataAgent{err: errors.New("execute tool web_search: search backend unavailable")},
		toolRegistry:  tools.NewToolRegistry(),
	}
	if err := service.toolRegistry.Register("web_search", &errorTool{name: "web_search"}); err != nil {
		t.Fatalf("register tool: %v", err)
	}

	result, err := service.analyzeMetadataWithConfig(
		context.Background(),
		&agents.MetadataGenerationInput{FileName: "MyAwesomeMovie.mp4", FileType: "video"},
		&models.AIConfig{Enabled: true, EnableAgent: true, EnableWebSearch: true},
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result.Trace.FallbackReason != "execute tool web_search: search backend unavailable" {
		t.Fatalf("unexpected fallback reason %q", result.Trace.FallbackReason)
	}
	if len(result.Trace.ToolsAvailable) != 1 || result.Trace.ToolsAvailable[0] != "web_search" {
		t.Fatalf("expected resolved tool captured in fallback trace, got %v", result.Trace.ToolsAvailable)
	}
}

func TestAgentServiceAnalyzeMetadataWithTraceAddsRunMetadata(t *testing.T) {
	service := &AgentService{
		metadataAgent: &fakeMetadataAgent{
			result: &agents.MetadataAnalysisResult{
				Analysis: &agents.MetadataAnalysis{Tags: []string{"video"}, Description: "metadata"},
			},
		},
		toolRegistry: tools.NewToolRegistry(),
	}

	result, err := service.analyzeMetadataWithConfig(
		context.Background(),
		&agents.AnalysisRequest{FileName: "movie.mp4", FileType: "video"},
		&models.AIConfig{Enabled: true, EnableAgent: true, Model: "test-model"},
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result == nil || result.Trace == nil {
		t.Fatal("expected trace result")
	}
	if result.Trace.RunID == "" {
		t.Fatal("expected generated run id")
	}
	if result.Trace.Status != "completed" {
		t.Fatalf("expected completed status, got %q", result.Trace.Status)
	}
	if result.Trace.Model != "test-model" {
		t.Fatalf("expected model in trace, got %q", result.Trace.Model)
	}
}

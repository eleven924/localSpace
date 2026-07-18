package services

import (
	"context"
	"errors"
	"testing"

	"LocalSpace/app/agents"
	"LocalSpace/app/models"
	"LocalSpace/app/repositories"
	"LocalSpace/app/tools"
)

func TestNewAgentService(t *testing.T) {
	// Create a mock database wrapper for testing
	// In real implementation, this would use a proper mock interface
	dbWrapper := &repositories.SQLiteDBWrapper{}

	configRepo := repositories.NewConfigRepository(dbWrapper)

	service := NewAgentService(configRepo)
	if service == nil {
		t.Fatal("Expected agent service to be created")
	}

	if service.metadataAgent == nil {
		t.Error("Expected metadata agent to be initialized")
	}

	if service.toolRegistry == nil {
		t.Error("Expected tool registry to be initialized")
	}
}

func TestShouldExposeWebSearch(t *testing.T) {
	enabledConfig := &models.AIConfig{Enabled: true, EnableAgent: true, EnableWebSearch: true}
	sparseInput := &agents.MetadataGenerationInput{FileName: "MyAwesomeMovie.mp4", FileType: "video"}
	richInput := &agents.MetadataGenerationInput{FileName: "movie.mp4", FileType: "video", UserKeywords: "action thriller", UserTags: []string{"娱乐", "动作"}, UserDescription: "一部动作电影"}
	disabledConfig := &models.AIConfig{Enabled: true, EnableAgent: true, EnableWebSearch: false}
	blankNameInput := &agents.MetadataGenerationInput{FileName: "   ", FileType: "video"}

	if !shouldExposeWebSearch(enabledConfig, sparseInput) {
		t.Fatal("expected sparse input to expose web search")
	}

	if shouldExposeWebSearch(enabledConfig, richInput) {
		t.Fatal("expected rich input to skip web search exposure")
	}

	if shouldExposeWebSearch(disabledConfig, sparseInput) {
		t.Fatal("expected disabled web search config to skip exposure")
	}

	if shouldExposeWebSearch(enabledConfig, blankNameInput) {
		t.Fatal("expected blank file name to skip exposure")
	}
}

func TestResolveMetadataTools(t *testing.T) {
	service := &AgentService{toolRegistry: tools.NewToolRegistry()}
	searchTool := &testTool{name: "web_search"}
	if err := service.toolRegistry.Register(searchTool.Name(), searchTool); err != nil {
		t.Fatalf("register tool: %v", err)
	}

	resolved := service.resolveMetadataTools(
		&models.AIConfig{Enabled: true, EnableAgent: true, EnableWebSearch: true},
		&agents.MetadataGenerationInput{FileName: "MyAwesomeMovie.mp4", FileType: "video"},
	)
	if len(resolved) != 1 || resolved[0].Name() != "web_search" {
		t.Fatalf("expected web_search to be resolved, got %v", resolved)
	}

	notResolved := service.resolveMetadataTools(
		&models.AIConfig{Enabled: true, EnableAgent: true, EnableWebSearch: true},
		&agents.MetadataGenerationInput{FileName: "movie.mp4", FileType: "video", UserKeywords: "action thriller"},
	)
	if len(notResolved) != 0 {
		t.Fatalf("expected rich input to resolve no tools, got %v", notResolved)
	}
}

func TestAgentServiceAnalyzeMetadata_ReturnsUnifiedResult(t *testing.T) {
	service := &AgentService{
		metadataAgent: &fakeMetadataAgent{
			result: &agents.MetadataAnalysisResult{
				Analysis: &agents.MetadataAnalysis{Tags: []string{"video", "action"}, Description: "动作片"},
				Trace:    &agents.MetadataTrace{AgentEnabled: true},
			},
		},
		toolRegistry: tools.NewToolRegistry(),
	}

	result, err := service.analyzeMetadataWithConfig(context.Background(), &agents.MetadataGenerationInput{FileName: "movie.mp4", FileType: "video"}, &models.AIConfig{Enabled: true, EnableAgent: true})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if result.Analysis.Description != "动作片" {
		t.Fatalf("expected description to be returned, got %q", result.Analysis.Description)
	}
	if !result.Trace.AgentEnabled {
		t.Fatal("expected trace to preserve agent-enabled result")
	}
}

func TestAgentServiceAnalyzeMetadata_FallbackOnDisabledAgent(t *testing.T) {
	service := &AgentService{toolRegistry: tools.NewToolRegistry()}

	result, err := service.analyzeMetadataWithConfig(context.Background(), &agents.MetadataGenerationInput{FileName: "movie.mp4", FileType: "video", UserTags: []string{"娱乐"}}, &models.AIConfig{Enabled: true, EnableAgent: false})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(result.Analysis.Tags) == 0 {
		t.Fatal("expected fallback tags to be populated")
	}

	if result.Trace.FallbackReason == "" {
		t.Fatal("expected fallback reason to be recorded")
	}
}

func TestAgentServiceAnalyzeMetadataWithTrace_FallbackOnMissingConfigRepo(t *testing.T) {
	service := &AgentService{}

	result, err := service.AnalyzeMetadataWithTrace(context.Background(), &agents.MetadataGenerationInput{FileType: "video"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result.Trace.FallbackReason != "failed to load ai config" {
		t.Fatalf("expected missing config repo to fallback as failed to load ai config, got %q", result.Trace.FallbackReason)
	}
}

func TestAgentServiceGenerateCompatibilityMethods_UseUnifiedMetadataAnalysis(t *testing.T) {
	service := &AgentService{
		metadataAgent: &fakeMetadataAgent{
			result: &agents.MetadataAnalysisResult{
				Analysis: &agents.MetadataAnalysis{Tags: []string{"video", "剧情"}, Description: "剧情影片"},
				Trace:    &agents.MetadataTrace{AgentEnabled: true},
			},
		},
		toolRegistry: tools.NewToolRegistry(),
	}

	analysis, err := service.AnalyzeMetadata(context.Background(), &agents.MetadataGenerationInput{FileName: "movie.mp4", FileType: "video"})
	if err != nil {
		t.Fatalf("expected nil error from AnalyzeMetadata, got %v", err)
	}

	tags, err := service.GenerateTags(context.Background(), "movie.mp4", "video", "", nil, "")
	if err != nil {
		t.Fatalf("expected nil error from GenerateTags, got %v", err)
	}
	if len(tags) != len(analysis.Tags) {
		t.Fatalf("expected GenerateTags to match AnalyzeMetadata output size, got %v vs %v", tags, analysis.Tags)
	}
	for i := range tags {
		if tags[i] != analysis.Tags[i] {
			t.Fatalf("expected GenerateTags to match AnalyzeMetadata output, got %v vs %v", tags, analysis.Tags)
		}
	}

	description, err := service.GenerateDescription(context.Background(), "movie.mp4", "video", "", nil, "")
	if err != nil {
		t.Fatalf("expected nil error from GenerateDescription, got %v", err)
	}
	if description != analysis.Description {
		t.Fatalf("expected GenerateDescription to match AnalyzeMetadata output, got %q vs %q", description, analysis.Description)
	}
}

func TestAgentServiceAnalyzeMetadata_FallbackIncludesResolvedToolsOnAgentError(t *testing.T) {
	service := &AgentService{
		metadataAgent: &fakeMetadataAgent{err: errors.New("agent failed")},
		toolRegistry:  tools.NewToolRegistry(),
	}
	searchTool := &testTool{name: "web_search"}
	if err := service.toolRegistry.Register(searchTool.Name(), searchTool); err != nil {
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

	if result.Trace.FallbackReason != "agent failed" {
		t.Fatalf("expected fallback reason to capture agent error, got %q", result.Trace.FallbackReason)
	}
	if len(result.Trace.ToolsAvailable) != 1 || result.Trace.ToolsAvailable[0] != "web_search" {
		t.Fatalf("expected resolved tools to be captured in fallback trace, got %v", result.Trace.ToolsAvailable)
	}
}

type fakeMetadataAgent struct {
	result *agents.MetadataAnalysisResult
	err    error
}

func (f *fakeMetadataAgent) Analyze(ctx context.Context, input *agents.MetadataGenerationInput, aiConfig *models.AIConfig, availableTools []tools.Tool) (*agents.MetadataAnalysisResult, error) {
	return f.result, f.err
}

type testTool struct {
	name string
}

func (m *testTool) Name() string {
	return m.name
}

func (m *testTool) Description() string {
	return "test tool"
}

func (m *testTool) Execute(ctx context.Context, input string) (string, error) {
	return "test result", nil
}

// Note: Tests that require database access are skipped for now
// In a real implementation, we would use a proper mock interface or test database

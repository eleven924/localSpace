package services

import (
	"context"
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

	if service.tagAgent == nil {
		t.Error("Expected tag agent to be initialized")
	}

	if service.descriptionAgent == nil {
		t.Error("Expected description agent to be initialized")
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

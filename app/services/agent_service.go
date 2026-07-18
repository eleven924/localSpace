package services

import (
	"context"
	"strings"
	"time"

	"LocalSpace/app/agents"
	"LocalSpace/app/models"
	"LocalSpace/app/repositories"
	"LocalSpace/app/tools"
)

// AgentService manages AI agents for tag and description generation
type AgentService struct {
	metadataAgent agents.MetadataAgent
	configRepo    *repositories.ConfigRepository
	toolRegistry  *tools.ToolRegistry
}

// NewAgentService creates a new agent service with phase-one metadata wiring.
//
// The service layer owns tool registration and gating. The default runtime remains a single
// direct chat-model call seam so the phase-one architecture is explicit in code.
func NewAgentService(configRepo *repositories.ConfigRepository) *AgentService {
	toolRegistry := tools.NewToolRegistry()

	webSearchTool := tools.NewWebSearchTool(10 * time.Second)
	toolRegistry.Register(webSearchTool.Name(), webSearchTool)

	metadataRuntime := agents.NewDefaultAgentRuntime()
	metadataAgent := agents.NewEinoMetadataAgent(metadataRuntime)

	return &AgentService{
		metadataAgent: metadataAgent,
		configRepo:    configRepo,
		toolRegistry:  toolRegistry,
	}
}

// GenerateTags generates tags for a file
func (s *AgentService) GenerateTags(
	ctx context.Context,
	fileName, fileType string,
	userKeywords string,
	userTags []string,
	userDescription string,
) ([]string, error) {
	analysis, err := s.AnalyzeMetadata(ctx, &agents.MetadataGenerationInput{
		FileName:        fileName,
		FileType:        fileType,
		UserKeywords:    userKeywords,
		UserTags:        userTags,
		UserDescription: userDescription,
	})
	if err != nil {
		return []string{}, err
	}
	if analysis == nil {
		return []string{}, nil
	}

	return analysis.Tags, nil
}

// GenerateDescription generates a description for a file
func (s *AgentService) GenerateDescription(
	ctx context.Context,
	fileName, fileType string,
	userKeywords string,
	userTags []string,
	userDescription string,
) (string, error) {
	analysis, err := s.AnalyzeMetadata(ctx, &agents.MetadataGenerationInput{
		FileName:        fileName,
		FileType:        fileType,
		UserKeywords:    userKeywords,
		UserTags:        userTags,
		UserDescription: userDescription,
	})
	if err != nil {
		return "", err
	}
	if analysis == nil {
		return "", nil
	}

	return analysis.Description, nil
}

// GenerateBatch generates tags and descriptions for multiple files (reserved for future use)
func (s *AgentService) GenerateBatch(
	ctx context.Context,
	files []FileContext,
) ([]*models.AIAnalysis, error) {
	// Reserved for future implementation
	return []*models.AIAnalysis{}, nil
}

func shouldExposeWebSearch(config *models.AIConfig, input *agents.MetadataGenerationInput) bool {
	if config == nil || input == nil {
		return false
	}

	if !config.Enabled || !config.EnableAgent || !config.EnableWebSearch {
		return false
	}

	if strings.TrimSpace(input.FileName) == "" {
		return false
	}

	hasRichUserContext := strings.TrimSpace(input.UserDescription) != "" || len(input.UserTags) >= 2 || strings.TrimSpace(input.UserKeywords) != ""
	return !hasRichUserContext
}

func (s *AgentService) resolveMetadataTools(config *models.AIConfig, input *agents.MetadataGenerationInput) []tools.Tool {
	if s == nil || s.toolRegistry == nil || !shouldExposeWebSearch(config, input) {
		return []tools.Tool{}
	}

	tool, err := s.toolRegistry.Get("web_search")
	if err != nil {
		return []tools.Tool{}
	}

	return []tools.Tool{tool}
}

// AnalyzeMetadata returns unified metadata analysis without trace details.
func (s *AgentService) AnalyzeMetadata(ctx context.Context, input *agents.MetadataGenerationInput) (*agents.MetadataAnalysis, error) {
	result, err := s.AnalyzeMetadataWithTrace(ctx, input)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	return result.Analysis, nil
}

// AnalyzeMetadataWithTrace returns unified metadata analysis with diagnostics.
func (s *AgentService) AnalyzeMetadataWithTrace(ctx context.Context, input *agents.MetadataGenerationInput) (*agents.MetadataAnalysisResult, error) {
	if s == nil {
		return fallbackMetadataResult(input, "agent service unavailable"), nil
	}
	if s.configRepo == nil {
		return fallbackMetadataResult(input, "failed to load ai config"), nil
	}

	config, err := s.configRepo.GetAIConfig()
	if err != nil {
		return fallbackMetadataResult(input, "failed to load ai config"), nil
	}

	return s.analyzeMetadataWithConfig(ctx, input, config)
}

func (s *AgentService) analyzeMetadataWithConfig(ctx context.Context, input *agents.MetadataGenerationInput, config *models.AIConfig) (*agents.MetadataAnalysisResult, error) {
	if config == nil || !config.Enabled || !config.EnableAgent {
		return fallbackMetadataResult(input, "agent disabled"), nil
	}
	if s == nil || s.metadataAgent == nil {
		return fallbackMetadataResult(input, "metadata agent unavailable"), nil
	}

	availableTools := s.resolveMetadataTools(config, input)
	result, err := s.metadataAgent.Analyze(ctx, input, config, availableTools)
	if err != nil {
		fallback := fallbackMetadataResult(input, err.Error())
		fallback.Trace.ToolsAvailable = namesForTools(availableTools)
		return fallback, nil
	}
	if result == nil {
		fallback := fallbackMetadataResult(input, "metadata agent returned nil result")
		fallback.Trace.ToolsAvailable = namesForTools(availableTools)
		return fallback, nil
	}
	if result.Trace == nil {
		result.Trace = &agents.MetadataTrace{}
	}
	if len(result.Trace.ToolsAvailable) == 0 {
		result.Trace.ToolsAvailable = namesForTools(availableTools)
	}

	return result, nil
}

func fallbackMetadataResult(input *agents.MetadataGenerationInput, reason string) *agents.MetadataAnalysisResult {
	tags := []string{}
	description := ""
	fileType := ""
	if input != nil {
		tags = append(tags, input.UserTags...)
		description = input.UserDescription
		fileType = strings.TrimSpace(input.FileType)
	}

	if len(tags) == 0 && fileType != "" {
		tags = []string{fileType}
	}
	if description == "" && fileType != "" {
		description = fileType + "文件"
	}

	return &agents.MetadataAnalysisResult{
		Analysis: &agents.MetadataAnalysis{
			Tags:        tags,
			Description: description,
		},
		Trace: &agents.MetadataTrace{
			AgentEnabled:   false,
			FallbackReason: reason,
		},
	}
}

func namesForTools(availableTools []tools.Tool) []string {
	names := make([]string, 0, len(availableTools))
	for _, tool := range availableTools {
		if tool == nil {
			continue
		}
		names = append(names, tool.Name())
	}
	return names
}

// FileContext represents a file for batch generation
type FileContext struct {
	FileName        string
	FileType        string
	UserKeywords    string
	UserTags        []string
	UserDescription string
}
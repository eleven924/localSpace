package services

import (
	"context"
	"path/filepath"
	"strings"

	"LocalSpace/app/agents"
	"LocalSpace/app/models"
	"LocalSpace/app/repositories"
	"LocalSpace/app/tools"
)

// AgentService coordinates metadata generation with runtime-configured tools.
type AgentService struct {
	metadataAgent agents.MetadataAgent
	configRepo    *repositories.ConfigRepository
	toolRegistry  *tools.ToolRegistry
}

// NewAgentService creates a new AgentService without eagerly registering tools.
func NewAgentService(configRepo *repositories.ConfigRepository) *AgentService {
	return &AgentService{
		metadataAgent: agents.NewEinoMetadataAgent(agents.NewDefaultAgentRuntime()),
		configRepo:    configRepo,
		toolRegistry:  tools.NewToolRegistry(),
	}
}

func (s *AgentService) ensureConfiguredTools(config *models.AIConfig) {
	if s == nil || s.toolRegistry == nil {
		return
	}
	webSearchTool := tools.NewConfiguredWebSearchTool(config)
	_ = s.toolRegistry.Register(webSearchTool.Name(), webSearchTool)
}

func (s *AgentService) resolveMetadataTools(config *models.AIConfig, input *agents.MetadataGenerationInput) []tools.Tool {
	if s == nil || s.toolRegistry == nil || config == nil || input == nil {
		return nil
	}
	if !config.EnableAgent || !config.EnableWebSearch {
		return nil
	}
	if !isMetadataSearchCandidate(input) {
		return nil
	}

	tool, ok := s.toolRegistry.Get("web_search")
	if !ok || tool == nil {
		return nil
	}
	return []tools.Tool{tool}
}

func (s *AgentService) analyzeMetadataWithConfig(ctx context.Context, input *agents.MetadataGenerationInput, config *models.AIConfig) (*agents.MetadataAnalysisResult, error) {
	if s == nil || s.metadataAgent == nil {
		return fallbackMetadataResult(input, nil, "metadata agent is not configured"), nil
	}
	if input == nil {
		return fallbackMetadataResult(nil, nil, "metadata input is nil"), nil
	}
	if config == nil {
		config = &models.AIConfig{}
	}

	s.ensureConfiguredTools(config)
	availableTools := s.resolveMetadataTools(config, input)
	result, err := s.metadataAgent.Analyze(ctx, input, config, availableTools)
	if err != nil {
		return fallbackMetadataResult(input, availableTools, err.Error()), nil
	}
	if result.Trace == nil {
		result.Trace = &agents.MetadataTrace{}
	}
	if len(result.Trace.ToolsAvailable) == 0 {
		result.Trace.ToolsAvailable = toolNames(availableTools)
	}
	return result, nil
}

func isMetadataSearchCandidate(input *agents.MetadataGenerationInput) bool {
	if input == nil {
		return false
	}
	fileType := strings.ToLower(strings.TrimSpace(input.FileType))
	switch fileType {
	case "video", "document", "music", "game", "installer", "image":
		return true
	}

	ext := strings.ToLower(filepath.Ext(input.FileName))
	switch ext {
	case ".mp4", ".mkv", ".avi", ".mov", ".pdf", ".doc", ".docx", ".mp3", ".flac":
		return true
	default:
		return false
	}
}

func fallbackMetadataResult(input *agents.MetadataGenerationInput, availableTools []tools.Tool, reason string) *agents.MetadataAnalysisResult {
	result := &agents.MetadataAnalysisResult{
		Tags:        []string{},
		Description: "",
		Trace: &agents.MetadataTrace{
			ToolsAvailable: toolNames(availableTools),
			FallbackReason: reason,
		},
	}
	if input != nil {
		result.Tags = fallbackTags(input)
		result.Description = fallbackDescription(input)
	}
	return result
}

func toolNames(availableTools []tools.Tool) []string {
	names := make([]string, 0, len(availableTools))
	for _, tool := range availableTools {
		if tool == nil {
			continue
		}
		names = append(names, tool.Name())
	}
	return names
}

func fallbackTags(input *agents.MetadataGenerationInput) []string {
	if input == nil {
		return []string{}
	}
	name := strings.TrimSuffix(input.FileName, filepath.Ext(input.FileName))
	name = strings.TrimSpace(name)
	if name == "" {
		return []string{}
	}
	if input.FileType != "" {
		return []string{strings.ToLower(input.FileType), name}
	}
	return []string{name}
}

func fallbackDescription(input *agents.MetadataGenerationInput) string {
	if input == nil {
		return ""
	}
	if input.FileType == "" {
		return input.FileName
	}
	return input.FileName + " (" + input.FileType + ")"
}

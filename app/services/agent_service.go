package services

import (
	"context"
	"fmt"
	"time"

	"LocalSpace/app/agents"
	"LocalSpace/app/models"
	"LocalSpace/app/repositories"
	"LocalSpace/app/tools"
)

// AgentService manages AI agents for tag and description generation
type AgentService struct {
	tagAgent         *agents.TagAgent
	descriptionAgent *agents.DescriptionAgent
	configRepo       *repositories.ConfigRepository
	toolRegistry     *tools.ToolRegistry
}

// NewAgentService creates a new agent service
func NewAgentService(configRepo *repositories.ConfigRepository) *AgentService {
	// Create tool registry
	toolRegistry := tools.NewToolRegistry()

	// Register web search tool
	webSearchTool := tools.NewWebSearchTool(10 * time.Second)
	toolRegistry.Register(webSearchTool.Name(), webSearchTool)

	// Create agent configurations
	tagAgentConfig := &agents.AgentConfig{
		Name:    "tag-agent",
		Timeout: 30 * time.Second,
		Tools:   []tools.Tool{webSearchTool},
	}

	descriptionAgentConfig := &agents.AgentConfig{
		Name:    "description-agent",
		Timeout: 30 * time.Second,
		Tools:   []tools.Tool{webSearchTool},
	}

	// Create agents
	tagAgent := agents.NewTagAgent(tagAgentConfig)
	descriptionAgent := agents.NewDescriptionAgent(descriptionAgentConfig)

	return &AgentService{
		tagAgent:         tagAgent,
		descriptionAgent: descriptionAgent,
		configRepo:       configRepo,
		toolRegistry:     toolRegistry,
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
	// Get AI configuration
	config, err := s.configRepo.GetAIConfig()
	if err != nil {
		return []string{}, nil // Return empty tags, don't block import
	}

	// Check if AI is enabled
	if !config.Enabled {
		return []string{}, nil
	}

	// Create input
	input := &agents.TagGenerationInput{
		FileName:        fileName,
		FileType:        fileType,
		UserKeywords:    userKeywords,
		UserTags:        userTags,
		UserDescription: userDescription,
	}

	// Generate tags
	tags, err := s.tagAgent.Generate(ctx, input, config)
	if err != nil {
		// Log error but don't block import
		fmt.Printf("Warning: Failed to generate tags: %v\n", err)
		return []string{}, nil
	}

	return tags, nil
}

// GenerateDescription generates a description for a file
func (s *AgentService) GenerateDescription(
	ctx context.Context,
	fileName, fileType string,
	userKeywords string,
	userTags []string,
	userDescription string,
) (string, error) {
	// Get AI configuration
	config, err := s.configRepo.GetAIConfig()
	if err != nil {
		return "", nil // Return empty description, don't block import
	}

	// Check if AI is enabled
	if !config.Enabled {
		return "", nil
	}

	// Create input
	input := &agents.DescriptionGenerationInput{
		FileName:        fileName,
		FileType:        fileType,
		UserKeywords:    userKeywords,
		UserTags:        userTags,
		UserDescription: userDescription,
	}

	// Generate description
	description, err := s.descriptionAgent.Generate(ctx, input, config)
	if err != nil {
		// Log error but don't block import
		fmt.Printf("Warning: Failed to generate description: %v\n", err)
		return "", nil
	}

	return description, nil
}

// GenerateBatch generates tags and descriptions for multiple files (reserved for future use)
func (s *AgentService) GenerateBatch(
	ctx context.Context,
	files []FileContext,
) ([]*models.AIAnalysis, error) {
	// Reserved for future implementation
	return []*models.AIAnalysis{}, nil
}

// FileContext represents a file for batch generation
type FileContext struct {
	FileName        string
	FileType        string
	UserKeywords    string
	UserTags        []string
	UserDescription string
}
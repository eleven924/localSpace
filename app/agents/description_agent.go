package agents

import (
	"context"
	"strings"

	"LocalSpace/app/models"
)

// DescriptionGenerationInput represents input for description generation
type DescriptionGenerationInput struct {
	FileName         string
	FileType         string
	UserKeywords     string
	UserTags         []string
	UserDescription  string
	Metadata         models.Metadata
}

// DescriptionAgent generates file descriptions using AI
type DescriptionAgent struct {
	*BaseAgent
	promptBuilder *PromptBuilder
}

// NewDescriptionAgent creates a new description agent
func NewDescriptionAgent(config *AgentConfig) *DescriptionAgent {
	baseAgent := NewBaseAgent(config)

	return &DescriptionAgent{
		BaseAgent:     baseAgent,
		promptBuilder: NewPromptBuilder(),
	}
}

// Generate generates a description for a file
func (a *DescriptionAgent) Generate(
	ctx context.Context,
	input *DescriptionGenerationInput,
	aiConfig *models.AIConfig,
) (string, error) {
	// Validate AI config
	if err := a.ValidateAIConfig(aiConfig); err != nil {
		return "", nil // Return empty description, don't block import
	}

	// Check if agent is enabled
	if !aiConfig.EnableAgent {
		return "", nil
	}

	// If user provided a description, use it
	if input.UserDescription != "" {
		// Truncate if too long
		if len(input.UserDescription) > 50 {
			return input.UserDescription[:50] + "...", nil
		}
		return input.UserDescription, nil
	}

	// Build prompt (prepared for future eino agent framework integration)
	_ = a.promptBuilder.BuildDescriptionPrompt(
		input.FileName,
		input.FileType,
		input.UserKeywords,
		input.UserTags,
		input.UserDescription,
	)

	// For now, return basic description
	description := a.generateBasicDescription(input)

	return description, nil
}

// generateBasicDescription generates a basic description without AI
func (a *DescriptionAgent) generateBasicDescription(input *DescriptionGenerationInput) string {
	var parts []string

	// Add file type
	if input.FileType != "" {
		typeNames := map[string]string{
			"video":     "视频",
			"document":  "文档",
			"music":     "音乐",
			"game":      "游戏",
			"installer": "安装包",
			"image":     "镜像",
		}
		if name, ok := typeNames[input.FileType]; ok {
			parts = append(parts, name)
		}
	}

	// Add keywords if available
	if input.UserKeywords != "" {
		parts = append(parts, input.UserKeywords)
	}

	// Add tags info
	if len(input.UserTags) > 0 {
		parts = append(parts, strings.Join(input.UserTags, "、"))
	}

	if len(parts) == 0 {
		return input.FileType + "文件"
	}

	description := strings.Join(parts, "，")
	if len(description) > 50 {
		return description[:50] + "..."
	}

	return description
}

// decideNeedSearch decides if web search is needed
func (a *DescriptionAgent) decideNeedSearch(
	ctx context.Context,
	input *DescriptionGenerationInput,
	aiConfig *models.AIConfig,
) (bool, error) {
	// Check if web search is enabled
	if !aiConfig.EnableWebSearch {
		return false, nil
	}

	// If user provided description, no need to search
	if input.UserDescription != "" {
		return false, nil
	}

	// Check if we have enough information
	hasUserInput := input.UserKeywords != "" || len(input.UserTags) > 0
	hasDescriptiveName := len(input.FileName) > 5

	// If we lack information and have a descriptive filename, search might help
	return !hasUserInput && hasDescriptiveName, nil
}
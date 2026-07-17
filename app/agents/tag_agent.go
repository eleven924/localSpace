package agents

import (
	"context"
	"strings"

	"LocalSpace/app/models"
)

// TagGenerationInput represents input for tag generation
type TagGenerationInput struct {
	FileName         string
	FileType         string
	UserKeywords     string
	UserTags         []string
	UserDescription  string
	Metadata         models.Metadata
}

// TagAgent generates file tags using AI
type TagAgent struct {
	*BaseAgent
	promptBuilder *PromptBuilder
}

// NewTagAgent creates a new tag agent
func NewTagAgent(config *AgentConfig) *TagAgent {
	baseAgent := NewBaseAgent(config)

	return &TagAgent{
		BaseAgent:     baseAgent,
		promptBuilder: NewPromptBuilder(),
	}
}

// Generate generates tags for a file
func (a *TagAgent) Generate(
	ctx context.Context,
	input *TagGenerationInput,
	aiConfig *models.AIConfig,
) ([]string, error) {
	// Validate AI config
	if err := a.ValidateAIConfig(aiConfig); err != nil {
		return []string{}, nil // Return empty tags, don't block import
	}

	// Check if agent is enabled
	if !aiConfig.EnableAgent {
		return []string{}, nil
	}

	// Build prompt (prepared for future eino agent framework integration)
	_ = a.promptBuilder.BuildTagPrompt(
		input.FileName,
		input.FileType,
		input.UserKeywords,
		input.UserTags,
		input.UserDescription,
	)

	// For now, return basic tags based on input
	// Full implementation will integrate with eino agent framework
	tags := a.generateBasicTags(input)

	return tags, nil
}

// generateBasicTags generates basic tags without AI
func (a *TagAgent) generateBasicTags(input *TagGenerationInput) []string {
	var tags []string

	// Add file type as tag
	if input.FileType != "" {
		tags = append(tags, input.FileType)
	}

	// Add user tags
	tags = append(tags, input.UserTags...)

	// Extract tags from keywords
	if input.UserKeywords != "" {
		keywords := strings.Fields(input.UserKeywords)
		for _, keyword := range keywords {
			if len(keyword) <= 4 { // Only short keywords as tags
				tags = append(tags, keyword)
			}
		}
	}

	// Limit to 5 tags
	if len(tags) > 5 {
		tags = tags[:5]
	}

	return tags
}

// decideNeedSearch decides if web search is needed
func (a *TagAgent) decideNeedSearch(
	ctx context.Context,
	input *TagGenerationInput,
	aiConfig *models.AIConfig,
) (bool, error) {
	// Check if web search is enabled
	if !aiConfig.EnableWebSearch {
		return false, nil
	}

	// Check if we have enough information
	hasUserInput := input.UserKeywords != "" || len(input.UserTags) > 0 || input.UserDescription != ""
	hasDescriptiveName := len(input.FileName) > 5

	// If we lack information and have a descriptive filename, search might help
	return !hasUserInput && hasDescriptiveName, nil
}
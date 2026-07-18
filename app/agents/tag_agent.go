package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
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

// TagAgent generates file tags using AI with direct OpenAI API calls
type TagAgent struct {
	*BaseAgent
	promptBuilder *PromptBuilder
	model         *openai.ChatModel
}

// NewTagAgent creates a new tag agent with OpenAI integration
func NewTagAgent(config *AgentConfig) *TagAgent {
	baseAgent := NewBaseAgent(config)

	return &TagAgent{
		BaseAgent:     baseAgent,
		promptBuilder: NewPromptBuilder(),
	}
}

// Generate generates tags for a file using OpenAI API
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
		return a.generateBasicTags(input), nil
	}

	// Check if we have enough information to skip AI
	hasUserInput := input.UserKeywords != "" || len(input.UserTags) > 0 || input.UserDescription != ""
	if hasUserInput && len(input.UserTags) >= 3 {
		return input.UserTags, nil
	}

	// Initialize OpenAI model if not already done
	if a.model == nil {
		if err := a.initializeModel(aiConfig); err != nil {
			// Fall back to basic generation on initialization error
			return a.generateBasicTags(input), nil
		}
	}

	// Build prompt
	prompt := a.promptBuilder.BuildTagPrompt(
		input.FileName,
		input.FileType,
		input.UserKeywords,
		input.UserTags,
		input.UserDescription,
	)

	// Call OpenAI API
	tags, err := a.callOpenAIForTags(ctx, prompt)
	if err != nil {
		// Log error but don't block import
		fmt.Printf("Warning: Failed to generate tags via AI: %v\n", err)
		return a.generateBasicTags(input), nil
	}

	return tags, nil
}

// initializeModel initializes the OpenAI chat model
func (a *TagAgent) initializeModel(aiConfig *models.AIConfig) error {
	// Create OpenAI chat model config
	openaiConfig := &openai.ChatModelConfig{
		APIKey:  aiConfig.APIKey,
		BaseURL: aiConfig.BaseURL,
		Model:   aiConfig.Model,
	}

	// Set timeout and max tokens if configured
	if aiConfig.Timeout > 0 {
		openaiConfig.Timeout = time.Duration(aiConfig.Timeout) * time.Second
	}
	if aiConfig.MaxTokens > 0 {
		maxTokens := aiConfig.MaxTokens
		openaiConfig.MaxTokens = &maxTokens
	}

	// Set temperature for consistent tag generation
	temperature := float32(0.7)
	openaiConfig.Temperature = &temperature

	// Create OpenAI model
	model, err := openai.NewChatModel(context.Background(), openaiConfig)
	if err != nil {
		return err
	}

	a.model = model
	return nil
}

// callOpenAIForTags calls OpenAI API to generate tags
func (a *TagAgent) callOpenAIForTags(ctx context.Context, prompt string) ([]string, error) {
	if a.model == nil {
		return []string{}, nil
	}

	// Create system message
	systemMsg := &schema.Message{
		Role:    schema.System,
		Content: "你是一个专业的文件标签生成助手。你的任务是根据文件信息生成3-5个精准、简洁的标签。每个标签2-4个字，用逗号分隔。只返回标签，不要有任何其他文字。",
	}

	// Create user message
	userMsg := &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}

	// Generate response
	response, err := a.model.Generate(ctx, []*schema.Message{systemMsg, userMsg})
	if err != nil {
		return []string{}, err
	}

	// Parse response into tags
	return a.parseTagResponse(response.Content), nil
}

// parseTagResponse parses the AI response into tags
func (a *TagAgent) parseTagResponse(response string) []string {
	// Clean up response
	response = strings.TrimSpace(response)

	// Remove common prefixes
	prefixes := []string{"生成的标签：", "标签：", "Tags:", "Tags", "Tag", "生成标签"}
	for _, prefix := range prefixes {
		response = strings.TrimPrefix(response, prefix)
		response = strings.TrimSpace(response)
	}

	// Remove quotes and brackets
	response = strings.Trim(response, `"'"` + `""'`)
	response = strings.ReplaceAll(response, "[", "")
	response = strings.ReplaceAll(response, "]", "")

	// Split by common delimiters
	delimiters := []string{"，", ",", "、", " ", "\n", "\t"}
	var tags []string

	for _, delimiter := range delimiters {
		if strings.Contains(response, delimiter) {
			parts := strings.Split(response, delimiter)
			for _, part := range parts {
				tag := strings.TrimSpace(part)
				// Remove quotes from individual tags
				tag = strings.Trim(tag, `"'"` + `""'`)
				if tag != "" && len(tag) <= 8 && len(tag) >= 1 { // Limit tag length
					tags = append(tags, tag)
				}
			}
			if len(tags) > 0 {
				break
			}
		}
	}

	// If no delimiters found, try JSON parsing
	if len(tags) == 0 {
		var result struct {
			Tags []string `json:"tags"`
		}
		if err := json.Unmarshal([]byte(response), &result); err == nil && len(result.Tags) > 0 {
			tags = result.Tags
		}
	}

	// If still no tags, try to split the whole string
	if len(tags) == 0 && len(response) > 0 {
		tags = []string{strings.TrimSpace(response)}
	}

	// Limit to 5 tags
	if len(tags) > 5 {
		tags = tags[:5]
	}

	// Remove duplicates
	uniqueTags := make([]string, 0, len(tags))
	seen := make(map[string]bool)
	for _, tag := range tags {
		if !seen[tag] {
			seen[tag] = true
			uniqueTags = append(uniqueTags, tag)
		}
	}

	return uniqueTags
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
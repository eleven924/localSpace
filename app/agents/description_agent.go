package agents

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
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

// DescriptionAgent generates file descriptions using AI with direct OpenAI API calls
type DescriptionAgent struct {
	*BaseAgent
	promptBuilder *PromptBuilder
	model         *openai.ChatModel
}

// NewDescriptionAgent creates a new description agent with OpenAI integration
func NewDescriptionAgent(config *AgentConfig) *DescriptionAgent {
	baseAgent := NewBaseAgent(config)

	return &DescriptionAgent{
		BaseAgent:     baseAgent,
		promptBuilder: NewPromptBuilder(),
	}
}

// Generate generates a description for a file using OpenAI API
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
		return a.generateBasicDescription(input), nil
	}

	// If user provided a description, use it
	if input.UserDescription != "" {
		// Truncate if too long
		if len(input.UserDescription) > 50 {
			return input.UserDescription[:50] + "...", nil
		}
		return input.UserDescription, nil
	}

	// Initialize OpenAI model if not already done
	if a.model == nil {
		if err := a.initializeModel(aiConfig); err != nil {
			// Fall back to basic generation on initialization error
			return a.generateBasicDescription(input), nil
		}
	}

	// Build prompt
	prompt := a.promptBuilder.BuildDescriptionPrompt(
		input.FileName,
		input.FileType,
		input.UserKeywords,
		input.UserTags,
		input.UserDescription,
	)

	// Call OpenAI API
	description, err := a.callOpenAIForDescription(ctx, prompt)
	if err != nil {
		// Log error but don't block import
		fmt.Printf("Warning: Failed to generate description via AI: %v\n", err)
		return a.generateBasicDescription(input), nil
	}

	return description, nil
}

// initializeModel initializes the OpenAI chat model
func (a *DescriptionAgent) initializeModel(aiConfig *models.AIConfig) error {
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

	// Set temperature for consistent description generation
	temperature := float32(0.6)
	openaiConfig.Temperature = &temperature

	// Create OpenAI model
	model, err := openai.NewChatModel(context.Background(), openaiConfig)
	if err != nil {
		return err
	}

	a.model = model
	return nil
}

// callOpenAIForDescription calls OpenAI API to generate description
func (a *DescriptionAgent) callOpenAIForDescription(ctx context.Context, prompt string) (string, error) {
	if a.model == nil {
		return "", nil
	}

	// Create system message
	systemMsg := &schema.Message{
		Role:    schema.System,
		Content: "你是一个专业的文件描述生成助手。你的任务是根据文件信息生成一个简洁准确的描述（1-2句话，不超过50个字）。描述应该概括文件的用途、内容或特点。只返回描述，不要有任何其他文字。",
	}

	// Create user message
	userMsg := &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}

	// Generate response
	response, err := a.model.Generate(ctx, []*schema.Message{systemMsg, userMsg})
	if err != nil {
		return "", err
	}

	// Clean up description
	return a.cleanDescription(response.Content), nil
}

// cleanDescription cleans up the AI response
func (a *DescriptionAgent) cleanDescription(response string) string {
	// Clean up response
	response = strings.TrimSpace(response)

	// Remove common prefixes
	prefixes := []string{"生成的描述：", "描述：", "Description:", "描述为：", "生成描述"}
	for _, prefix := range prefixes {
		response = strings.TrimPrefix(response, prefix)
		response = strings.TrimSpace(response)
	}

	// Remove quotes and brackets
	response = strings.Trim(response, `"'"` + `""'`)
	response = strings.ReplaceAll(response, "[", "")
	response = strings.ReplaceAll(response, "]", "")

	// Limit to 50 characters
	if len(response) > 50 {
		response = response[:50] + "..."
	}

	return response
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
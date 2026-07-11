package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"LocalSpace/app/models"
	"LocalSpace/app/repositories"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// AIService handles AI-related operations
type AIService struct {
	configRepo *repositories.ConfigRepository
	chatModel  model.ChatModel
}

// NewAIService creates a new AIService
func NewAIService(configRepo *repositories.ConfigRepository) *AIService {
	return &AIService{
		configRepo: configRepo,
	}
}

// initChatModel initializes or reinitializes the chat model with current config
func (s *AIService) initChatModel(config *models.AIConfig) error {
	ctx := context.Background()

	// Create OpenAI chat model config
	chatConfig := &openai.ChatModelConfig{
		APIKey:  config.APIKey,
		Model:   config.Model,
		BaseURL: config.BaseURL,
		Timeout: 30 * time.Second,
	}

	// Create the chat model
	chatModel, err := openai.NewChatModel(ctx, chatConfig)
	if err != nil {
		return fmt.Errorf("failed to create chat model: %w", err)
	}

	s.chatModel = chatModel
	return nil
}

// ensureChatModelInitialized ensures the chat model is initialized
func (s *AIService) ensureChatModelInitialized() error {
	if s.chatModel != nil {
		return nil
	}

	config, err := s.configRepo.GetAIConfig()
	if err != nil {
		return fmt.Errorf("failed to get AI config: %w", err)
	}

	return s.initChatModel(config)
}

// GenerateTags generates tags for a file using AI
func (s *AIService) GenerateTags(fileName, fileType string) ([]string, error) {
	config, err := s.configRepo.GetAIConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get AI config: %w", err)
	}

	if !config.Enabled || config.APIKey == "" {
		return []string{}, nil // AI not enabled, return empty tags
	}

	prompt := s.buildTagsPrompt(fileName, fileType)
	result, err := s.callAI(config, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call AI: %w", err)
	}

	return s.parseTags(result), nil
}

// GenerateDescription generates a description for a file using AI
func (s *AIService) GenerateDescription(fileName, fileType string) (string, error) {
	config, err := s.configRepo.GetAIConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get AI config: %w", err)
	}

	if !config.Enabled || config.APIKey == "" {
		return "", nil // AI not enabled, return empty description
	}

	prompt := s.buildDescriptionPrompt(fileName, fileType)
	result, err := s.callAI(config, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to call AI: %w", err)
	}

	return strings.TrimSpace(result), nil
}

// AnalyzeFile performs both tag and description generation
func (s *AIService) AnalyzeFile(fileName, fileType string) (*models.AIAnalysis, error) {
	tags, err := s.GenerateTags(fileName, fileType)
	if err != nil {
		tags = []string{}
	}

	description, err := s.GenerateDescription(fileName, fileType)
	if err != nil {
		description = ""
	}

	return &models.AIAnalysis{
		Tags:        tags,
		Description: description,
	}, nil
}

// buildTagsPrompt builds a prompt for tag generation (reserved for future use)
func (s *AIService) buildTagsPrompt(fileName, fileType string) string {
	prompts := map[string]string{
		"video":     fmt.Sprintf("文件名：%s\n类型：视频\n请为这个视频文件生成 3-5 个相关标签，用逗号分隔。标签应该反映视频的内容、风格或类型。", fileName),
		"document":  fmt.Sprintf("文件名：%s\n类型：文档\n请为这个文档文件生成 3-5 个相关标签，用逗号分隔。标签应该反映文档的主题或内容类别。", fileName),
		"music":     fmt.Sprintf("文件名：%s\n类型：音乐\n请为这个音乐文件生成 3-5 个相关标签，用逗号分隔。标签应该反映音乐的风格、情绪或类型。", fileName),
		"game":      fmt.Sprintf("文件名：%s\n类型：游戏\n请为这个游戏文件生成 3-5 个相关标签，用逗号分隔。标签应该反映游戏的类型或风格。", fileName),
		"installer": fmt.Sprintf("文件名：%s\n类型：安装包\n请为这个安装包文件生成 3-5 个相关标签，用逗号分隔。标签应该反映软件的类型或用途。", fileName),
		"image":     fmt.Sprintf("文件名：%s\n类型：镜像\n请为这个镜像文件生成 3-5 个相关标签，用逗号分隔。标签应该反映镜像的用途或类型。", fileName),
	}

	if prompt, ok := prompts[fileType]; ok {
		return prompt
	}

	return fmt.Sprintf("文件名：%s\n类型：%s\n请为这个文件生成 3-5 个相关标签，用逗号分隔。", fileName, fileType)
}

// buildDescriptionPrompt builds a prompt for description generation (reserved for future use)
func (s *AIService) buildDescriptionPrompt(fileName, fileType string) string {
	prompts := map[string]string{
		"video":     fmt.Sprintf("文件名：%s\n类型：视频\n请根据文件名为这个视频生成一个简短的描述（1-2句话）。", fileName),
		"document":  fmt.Sprintf("文件名：%s\n类型：文档\n请根据文件名为这个文档生成一个简短的描述（1-2句话）。", fileName),
		"music":     fmt.Sprintf("文件名：%s\n类型：音乐\n请根据文件名为这个音乐生成一个简短的描述（1-2句话）。", fileName),
		"game":      fmt.Sprintf("文件名：%s\n类型：游戏\n请根据文件名为这个游戏生成一个简短的描述（1-2句话）。", fileName),
		"installer": fmt.Sprintf("文件名：%s\n类型：安装包\n请根据文件名为这个安装包生成一个简短的描述（1-2句话）。", fileName),
		"image":     fmt.Sprintf("文件名：%s\n类型：镜像\n请根据文件名为这个镜像生成一个简短的描述（1-2句话）。", fileName),
	}

	if prompt, ok := prompts[fileType]; ok {
		return prompt
	}

	return fmt.Sprintf("文件名：%s\n类型：%s\n请根据文件名生成一个简短的描述（1-2句话）。", fileName, fileType)
}

// parseTags parses tags from AI response (reserved for future use)
func (s *AIService) parseTags(result string) []string {
	// Clean up the result
	result = strings.TrimSpace(result)

	// Split by common separators
	separators := []string{",", "，", ";", "；", "|", "、"}

	var tags []string
	for _, sep := range separators {
		if strings.Contains(result, sep) {
			tags = strings.Split(result, sep)
			break
		}
	}

	// If no separator found, treat whole result as single tag
	if len(tags) == 0 {
		tags = []string{result}
	}

	// Clean up each tag
	cleanTags := make([]string, 0)
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		tag = strings.Trim(tag, "\"'`")
		if tag != "" {
			cleanTags = append(cleanTags, tag)
		}
	}

	// Limit to 5 tags
	if len(cleanTags) > 5 {
		cleanTags = cleanTags[:5]
	}

	return cleanTags
}

// callAI makes a request to the AI API using eino chatmodel
func (s *AIService) callAI(config *models.AIConfig, prompt string) (string, error) {
	ctx := context.Background()

	// Ensure chat model is initialized with current config
	if err := s.initChatModel(config); err != nil {
		return "", fmt.Errorf("failed to initialize chat model: %w", err)
	}

	// Prepare message
	messages := []*schema.Message{
		schema.SystemMessage(prompt),
	}

	// Generate response
	resp, err := s.chatModel.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	// Extract content from response
	if resp.Content == "" {
		return "", fmt.Errorf("empty response from AI")
	}

	return resp.Content, nil
}
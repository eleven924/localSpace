package agents

import (
	"context"
	"fmt"
	"time"

	"LocalSpace/app/models"
	"LocalSpace/app/tools"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

// AgentRunRequest describes a single runtime execution request.
type AgentRunRequest struct {
	SystemPrompt string
	UserPrompt   string
	Tools        []tools.Tool
	AIConfig     *models.AIConfig
}

// AgentRunResponse captures runtime output and tool usage details.
type AgentRunResponse struct {
	Output        string
	ToolsUsed     []string
	SearchQueries []string
}

// AgentRuntime executes a single agent request.
type AgentRuntime interface {
	Run(ctx context.Context, req *AgentRunRequest) (*AgentRunResponse, error)
}

// OpenAIRuntime executes metadata-agent requests with the existing Eino/OpenAI seam.
type OpenAIRuntime struct{}

// NewOpenAIRuntime creates the default production runtime for metadata agents.
func NewOpenAIRuntime() *OpenAIRuntime {
	return &OpenAIRuntime{}
}

// Run performs a single chat-model execution and returns trace-aware runtime output.
func (r *OpenAIRuntime) Run(ctx context.Context, req *AgentRunRequest) (*AgentRunResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("agent run request is nil")
	}
	if req.AIConfig == nil {
		return nil, fmt.Errorf("ai config is nil")
	}

	chatConfig := &openai.ChatModelConfig{
		APIKey:  req.AIConfig.APIKey,
		Model:   req.AIConfig.Model,
		BaseURL: req.AIConfig.BaseURL,
	}
	if req.AIConfig.Timeout > 0 {
		chatConfig.Timeout = time.Duration(req.AIConfig.Timeout) * time.Second
	}
	if req.AIConfig.MaxTokens > 0 {
		maxTokens := req.AIConfig.MaxTokens
		chatConfig.MaxTokens = &maxTokens
	}

	temperature := float32(0.2)
	chatConfig.Temperature = &temperature

	chatModel, err := openai.NewChatModel(ctx, chatConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}

	messages := []*schema.Message{
		schema.SystemMessage(req.SystemPrompt),
		schema.UserMessage(req.UserPrompt),
	}

	resp, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}
	if resp.Content == "" {
		return nil, fmt.Errorf("empty response from AI")
	}

	return &AgentRunResponse{
		Output:    resp.Content,
		ToolsUsed: namesForRuntimeTools(req.Tools),
	}, nil
}

func namesForRuntimeTools(tools []tools.Tool) []string {
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		if tool == nil {
			continue
		}
		names = append(names, tool.Name())
	}
	return names
}

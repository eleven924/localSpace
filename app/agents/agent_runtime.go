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

// AgentRunResponse captures runtime output and tool execution details.
type AgentRunResponse struct {
	Output        string
	ToolsUsed     []string
	SearchQueries []string
}

// AgentRuntime executes a single agent request.
type AgentRuntime interface {
	Run(ctx context.Context, req *AgentRunRequest) (*AgentRunResponse, error)
}

// DefaultAgentRuntime executes phase-one metadata-agent requests through a direct chat-model seam.
//
// Phase one intentionally keeps tool gating and tool selection in AgentService. The runtime still
// performs a single direct model call rather than a multi-step agent/tool loop, so it must not
// report available tools as used unless a future runtime actually executes them.
type DefaultAgentRuntime struct{}

// NewDefaultAgentRuntime creates the default production runtime for phase-one metadata agents.
func NewDefaultAgentRuntime() *DefaultAgentRuntime {
	return &DefaultAgentRuntime{}
}

// Run performs a single direct chat-model execution and returns trace-aware runtime output.
func (r *DefaultAgentRuntime) Run(ctx context.Context, req *AgentRunRequest) (*AgentRunResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("agent run request is nil")
	}
	if req.AIConfig == nil {
		return nil, fmt.Errorf("ai config is nil")
	}
	if req.AIConfig.APIKey == "" || req.AIConfig.Model == "" || req.AIConfig.BaseURL == "" {
		return nil, fmt.Errorf("invalid ai config")
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
		Output: resp.Content,
	}, nil
}

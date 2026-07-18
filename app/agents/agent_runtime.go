package agents

import (
	"context"

	"LocalSpace/app/models"
	"LocalSpace/app/tools"
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

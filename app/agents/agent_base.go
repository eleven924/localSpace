package agents

import (
	"context"
	"fmt"
	"time"

	"LocalSpace/app/models"
	"LocalSpace/app/tools"
)

// AgentConfig represents agent configuration
type AgentConfig struct {
	Name      string
	Model     string
	Timeout   time.Duration
	MaxTokens int
	Tools     []tools.Tool
}

// BaseAgent provides common agent functionality
type BaseAgent struct {
	config       *AgentConfig
	toolRegistry *tools.ToolRegistry
}

// NewBaseAgent creates a new base agent
func NewBaseAgent(config *AgentConfig) *BaseAgent {
	registry := tools.NewToolRegistry()

	// Register tools
	for _, tool := range config.Tools {
		registry.Register(tool.Name(), tool)
	}

	return &BaseAgent{
		config:       config,
		toolRegistry: registry,
	}
}

// Name returns the agent name
func (a *BaseAgent) Name() string {
	return a.config.Name
}

// CreateContext creates a context with timeout
func (a *BaseAgent) CreateContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, a.config.Timeout)
}

// GetTool retrieves a tool by name
func (a *BaseAgent) GetTool(name string) (tools.Tool, error) {
	tool, ok := a.toolRegistry.Get(name)
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return tool, nil
}

// ListTools returns all available tool names
func (a *BaseAgent) ListTools() []string {
	return a.toolRegistry.List()
}

// ValidateAIConfig validates AI configuration for agent use
func (a *BaseAgent) ValidateAIConfig(config *models.AIConfig) error {
	if config == nil {
		return fmt.Errorf("AI config cannot be nil")
	}

	if !config.Enabled {
		return fmt.Errorf("AI is not enabled")
	}

	if config.APIKey == "" {
		return fmt.Errorf("AI API key cannot be empty")
	}

	if config.Model == "" {
		return fmt.Errorf("AI model cannot be empty")
	}

	return nil
}

package tools

import (
	"context"
	"errors"
	"sync"
)

// Tool represents a tool that can be called by agents
type Tool interface {
	// Name returns the tool name
	Name() string

	// Description returns a description of what the tool does
	Description() string

	// Execute executes the tool with the given input
	Execute(ctx context.Context, input string) (string, error)
}

// ToolRegistry manages available tools
type ToolRegistry struct {
	tools map[string]Tool
	mu    sync.RWMutex
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register registers a tool with the given name
func (r *ToolRegistry) Register(name string, tool Tool) error {
	if name == "" {
		return errors.New("tool name cannot be empty")
	}

	if tool == nil {
		return errors.New("tool cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools[name] = tool
	return nil
}

// Get retrieves a tool by name
func (r *ToolRegistry) Get(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		return nil, errors.New("tool not found: " + name)
	}

	return tool, nil
}

// List returns all registered tool names
func (r *ToolRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}

	return names
}


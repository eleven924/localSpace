package tools

import "fmt"

// ToolRegistry stores runtime tools by stable name.
type ToolRegistry struct {
	tools map[string]Tool
}

// NewToolRegistry creates an empty registry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{tools: make(map[string]Tool)}
}

// Register stores or replaces a tool by name.
func (r *ToolRegistry) Register(name string, tool Tool) error {
	if r == nil {
		return fmt.Errorf("tool registry is nil")
	}
	if tool == nil {
		return fmt.Errorf("tool is nil")
	}
	if name == "" {
		name = tool.Name()
	}
	if name == "" {
		return fmt.Errorf("tool name is empty")
	}
	if r.tools == nil {
		r.tools = make(map[string]Tool)
	}

	r.tools[name] = tool
	return nil
}

// Get returns one tool by name.
func (r *ToolRegistry) Get(name string) (Tool, bool) {
	if r == nil || r.tools == nil {
		return nil, false
	}
	tool, ok := r.tools[name]
	return tool, ok
}

// List returns all registered tool names.
func (r *ToolRegistry) List() []string {
	if r == nil || r.tools == nil {
		return nil
	}
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

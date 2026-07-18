package agents

import (
	"context"

	"LocalSpace/app/models"
	"LocalSpace/app/tools"
)

// MetadataAgent analyzes metadata generation input through a runtime seam.
type MetadataAgent interface {
	Analyze(ctx context.Context, input *MetadataGenerationInput, aiConfig *models.AIConfig, availableTools []tools.Tool) (*MetadataAnalysisResult, error)
}

// EinoMetadataAgent is the first metadata agent wrapper over AgentRuntime.
type EinoMetadataAgent struct {
	runtime       AgentRuntime
	promptBuilder *PromptBuilder
}

// NewEinoMetadataAgent creates a metadata agent backed by the provided runtime.
func NewEinoMetadataAgent(runtime AgentRuntime) *EinoMetadataAgent {
	return &EinoMetadataAgent{
		runtime:       runtime,
		promptBuilder: NewPromptBuilder(),
	}
}

// Analyze builds prompts, executes the runtime, and parses structured metadata output.
func (a *EinoMetadataAgent) Analyze(ctx context.Context, input *MetadataGenerationInput, aiConfig *models.AIConfig, availableTools []tools.Tool) (*MetadataAnalysisResult, error) {
	response, err := a.runtime.Run(ctx, &AgentRunRequest{
		SystemPrompt: a.promptBuilder.BuildMetadataSystemPrompt(),
		UserPrompt:   a.promptBuilder.BuildMetadataUserPrompt(input),
		Tools:        availableTools,
		AIConfig:     aiConfig,
	})
	if err != nil {
		return nil, err
	}

	analysis, err := ParseMetadataOutput(response.Output)
	if err != nil {
		return nil, err
	}

	toolNames := make([]string, 0, len(availableTools))
	for _, tool := range availableTools {
		toolNames = append(toolNames, tool.Name())
	}

	return &MetadataAnalysisResult{
		Analysis: analysis,
		Trace: &MetadataTrace{
			AgentEnabled:   true,
			ToolsAvailable: toolNames,
			ToolsUsed:      response.ToolsUsed,
			SearchQueries:  response.SearchQueries,
			RawOutput:      response.Output,
		},
	}, nil
}

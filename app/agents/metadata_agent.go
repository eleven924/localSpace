package agents

import (
	"context"
	"errors"

	"LocalSpace/app/models"
	"LocalSpace/app/tools"
)

// MetadataAgent analyzes metadata generation input through a runtime seam.
type MetadataAgent interface {
	Analyze(ctx context.Context, input *MetadataGenerationInput, aiConfig *models.AIConfig, availableTools []tools.Tool) (*MetadataAnalysisResult, error)
}

// EinoMetadataAgent is the phase-one metadata-agent wrapper over AgentRuntime.
//
// It owns prompt construction, delegates one runtime call, and parses structured output. It does
// not yet orchestrate a richer tool-calling loop inside the runtime.
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
	if a == nil || a.runtime == nil {
		return nil, errors.New("metadata agent runtime is nil")
	}
	promptBuilder := a.promptBuilder
	if promptBuilder == nil {
		promptBuilder = NewPromptBuilder()
	}

	response, err := a.runtime.Run(ctx, &AgentRunRequest{
		SystemPrompt: promptBuilder.BuildMetadataSystemPrompt(),
		UserPrompt:   promptBuilder.BuildMetadataUserPrompt(input),
		Tools:        availableTools,
		AIConfig:     aiConfig,
	})
	if err != nil {
		return nil, err
	}

	if response == nil {
		return nil, errors.New("metadata agent runtime returned nil response")
	}

	analysis, err := ParseMetadataOutput(response.Output)
	if err != nil {
		return nil, err
	}

	toolNames := make([]string, 0, len(availableTools))
	for _, tool := range availableTools {
		if tool == nil {
			continue
		}
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

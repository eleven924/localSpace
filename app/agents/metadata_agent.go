package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"LocalSpace/app/models"
	"LocalSpace/app/tools"
)

// MetadataGenerationInput describes the file under analysis.
type MetadataGenerationInput struct {
	FileName string
	FileType string
}

// MetadataTrace captures tool availability, runtime usage, and fallback reason.
type MetadataTrace struct {
	ToolsAvailable []string `json:"toolsAvailable,omitempty"`
	ToolsUsed      []string `json:"toolsUsed,omitempty"`
	SearchQueries  []string `json:"searchQueries,omitempty"`
	FallbackReason string   `json:"fallbackReason,omitempty"`
}

// MetadataAnalysisResult is the agent-facing metadata analysis payload.
type MetadataAnalysisResult struct {
	Tags        []string       `json:"tags"`
	Description string         `json:"description"`
	Trace       *MetadataTrace `json:"trace,omitempty"`
}

// MetadataAgent analyzes metadata using the configured runtime and optional tools.
type MetadataAgent interface {
	Analyze(ctx context.Context, input *MetadataGenerationInput, aiConfig *models.AIConfig, availableTools []tools.Tool) (*MetadataAnalysisResult, error)
}

// EinoMetadataAgent is a metadata-focused agent wrapper around AgentRuntime.
type EinoMetadataAgent struct {
	runtime AgentRuntime
}

// NewEinoMetadataAgent creates a metadata agent with the provided runtime.
func NewEinoMetadataAgent(runtime AgentRuntime) *EinoMetadataAgent {
	if runtime == nil {
		runtime = NewDefaultAgentRuntime()
	}
	return &EinoMetadataAgent{runtime: runtime}
}

// Analyze executes metadata generation and propagates runtime tool trace information.
func (a *EinoMetadataAgent) Analyze(ctx context.Context, input *MetadataGenerationInput, aiConfig *models.AIConfig, availableTools []tools.Tool) (*MetadataAnalysisResult, error) {
	if a == nil || a.runtime == nil {
		return nil, fmt.Errorf("metadata agent runtime is not configured")
	}
	if input == nil {
		return nil, fmt.Errorf("metadata generation input is nil")
	}
	if aiConfig == nil {
		return nil, fmt.Errorf("ai config is nil")
	}

	availableNames := make([]string, 0, len(availableTools))
	for _, tool := range availableTools {
		if tool == nil {
			continue
		}
		availableNames = append(availableNames, tool.Name())
	}

	resp, err := a.runtime.Run(ctx, &AgentRunRequest{
		SystemPrompt: buildMetadataSystemPrompt(),
		UserPrompt:   buildMetadataUserPrompt(input),
		Tools:        availableTools,
		AIConfig:     aiConfig,
	})
	if err != nil {
		return nil, err
	}

	result := &MetadataAnalysisResult{Trace: &MetadataTrace{ToolsAvailable: availableNames, ToolsUsed: resp.ToolsUsed, SearchQueries: resp.SearchQueries}}
	if err := json.Unmarshal([]byte(resp.Output), result); err != nil {
		return nil, fmt.Errorf("parse metadata analysis output: %w", err)
	}
	if result.Trace == nil {
		result.Trace = &MetadataTrace{}
	}
	result.Trace.ToolsAvailable = availableNames
	result.Trace.ToolsUsed = resp.ToolsUsed
	result.Trace.SearchQueries = resp.SearchQueries
	return result, nil
}

func buildMetadataSystemPrompt() string {
	return strings.TrimSpace(`You analyze one file and return strict JSON with keys "tags" and "description".
If you need external factual context, call a tool by returning JSON like {"tool":"web_search","input":"..."}.
When you have enough information, return final JSON like {"tags":["tag1"],"description":"..."} with no extra text.`)
}

func buildMetadataUserPrompt(input *MetadataGenerationInput) string {
	return fmt.Sprintf("file name: %s\nfile type: %s\nGenerate 3-5 relevant tags and a short description.", input.FileName, input.FileType)
}

package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"LocalSpace/app/models"
	"LocalSpace/app/tools"
)

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

	analysis := &MetadataAnalysis{}
	if err := json.Unmarshal([]byte(resp.Output), analysis); err != nil {
		return nil, fmt.Errorf("parse metadata analysis output: %w", err)
	}

	return &MetadataAnalysisResult{
		Analysis: analysis,
		Trace: &MetadataTrace{
			AgentEnabled:   aiConfig.Enabled && aiConfig.EnableAgent,
			ToolsAvailable: availableNames,
			ToolsUsed:      resp.ToolsUsed,
			SearchQueries:  resp.SearchQueries,
			RawOutput:      strings.TrimSpace(resp.Output),
		},
	}, nil
}

func buildMetadataSystemPrompt() string {
	return strings.TrimSpace(`You analyze one file and return strict JSON with keys "tags" and "description".
If you need external factual context, call a tool by returning JSON like {"tool":"web_search","input":"..."}.
When you have enough information, return final JSON like {"tags":["tag1"],"description":"..."} with no extra text.`)
}

func buildMetadataUserPrompt(input *MetadataGenerationInput) string {
	tagsStr := ""
	if input != nil {
		tagsStr = strings.Join(input.UserTags, ", ")
		return fmt.Sprintf(
			"文件名：%s\n文件类型：%s\n用户关键词：%s\n用户标签：%s\n用户描述：%s\n请输出 metadata JSON。",
			input.FileName,
			input.FileType,
			input.UserKeywords,
			tagsStr,
			input.UserDescription,
		)
	}

	return fmt.Sprintf(
		"文件名：%s\n文件类型：%s\n用户关键词：%s\n用户标签：%s\n用户描述：%s\n请输出 metadata JSON。",
		"", "", "", tagsStr, "",
	)
}

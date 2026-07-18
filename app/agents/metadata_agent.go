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

	// 检查是否启用了web search工具
	hasWebSearch := hasTool(availableTools, "web_search")

	resp, err := a.runtime.Run(ctx, &AgentRunRequest{
		SystemPrompt: buildMetadataSystemPrompt(hasWebSearch),
		UserPrompt:   buildMetadataUserPrompt(input, hasWebSearch),
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

func buildMetadataSystemPrompt(hasWebSearch bool) string {
	if hasWebSearch {
		return strings.TrimSpace(`You analyze one file and return strict JSON with keys "tags" and "description".

IMPORTANT: For video, document, music, game, installer, or image files, you MUST use the web search tool first to gather accurate information before generating tags and description.

WORKFLOW:
1. First, call the web search tool by returning JSON like: {"tool":"web_search","input":"文件名或关键词"}
2. Wait for the search results
3. After receiving search results, return final JSON like: {"tags":["tag1","tag2"],"description":"detailed description"}

The web search tool is available and should be used to get accurate information about the file content, especially for media files where the filename alone is insufficient.

Return STRICT JSON format only, no other text.`)
	}

	// 没有启用web search时的简化prompt
	return strings.TrimSpace(`You analyze one file and return strict JSON with keys "tags" and "description".

Analyze the filename and file type to generate appropriate tags and a brief description.
- Tags should be 1-5 relevant keywords or categories
- Description should be a concise summary of what the file likely contains

Return STRICT JSON format only, no other text: {"tags":["tag1","tag2"],"description":"..."}`)
}

func buildMetadataUserPrompt(input *MetadataGenerationInput, hasWebSearch bool) string {
	tagsStr := ""
	if input != nil {
		tagsStr = strings.Join(input.UserTags, ", ")
		hint := ""
		// 只有在启用web search且文件类型需要时，才添加提示
		if hasWebSearch && (input.FileType == "video" || input.FileType == "document" || input.FileType == "music") {
			hint = "\n提示：文件名可能不足以描述文件内容，请先使用web_search工具搜索文件名相关信息，获取准确的描述后再生成标签。"
		}

		if hasWebSearch {
			return fmt.Sprintf(
				"文件名：%s\n文件类型：%s\n用户关键词：%s\n用户标签：%s\n用户描述：%s%s\n请先使用web_search工具搜索文件信息，然后输出 metadata JSON。",
				input.FileName,
				input.FileType,
				input.UserKeywords,
				tagsStr,
				input.UserDescription,
				hint,
			)
		}

		return fmt.Sprintf(
			"文件名：%s\n文件类型：%s\n用户关键词：%s\n用户标签：%s\n用户描述：%s\n请直接输出 metadata JSON。",
			input.FileName,
			input.FileType,
			input.UserKeywords,
			tagsStr,
			input.UserDescription,
		)
	}

	if hasWebSearch {
		return fmt.Sprintf(
			"文件名：%s\n文件类型：%s\n用户关键词：%s\n用户标签：%s\n用户描述：%s\n请先使用web_search工具搜索文件信息，然后输出 metadata JSON。",
			"", "", "", tagsStr, "",
		)
	}

	return fmt.Sprintf(
		"文件名：%s\n文件类型：%s\n用户关键词：%s\n用户标签：%s\n用户描述：%s\n请直接输出 metadata JSON。",
		"", "", "", tagsStr, "",
	)
}

// hasTool 检查availableTools中是否包含指定名称的工具
func hasTool(availableTools []tools.Tool, name string) bool {
	for _, tool := range availableTools {
		if tool != nil && tool.Name() == name {
			return true
		}
	}
	return false
}

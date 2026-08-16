package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
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

	// 记录本地证据和 Web Search 是否可用，让 Prompt 明确证据优先级。
	hasWebSearch := hasTool(availableTools, "web_search")
	hasLocalEvidence := hasTool(availableTools, "get_local_file_metadata") || hasTool(availableTools, "extract_document_text") || hasTool(availableTools, "list_archive_entries")

	resp, err := a.runtime.Run(ctx, &AgentRunRequest{
		SystemPrompt: buildMetadataSystemPrompt(hasWebSearch, hasLocalEvidence),
		UserPrompt:   buildMetadataUserPrompt(input, hasWebSearch),
		Tools:        availableTools,
		AIConfig:     aiConfig,
	})
	if err != nil {
		// Eino 可能已经完成工具调用后才在最终轮次失败；优先保留已有 trace，
		// 避免服务层降级时把“实际调用过的工具”一起丢掉。
		if resp != nil {
			// 如果失败发生在最终回答阶段，仍尝试利用已完成的工具结果生成结构化 JSON。
			repairedResp, repairErr := a.repairMetadataOutput(ctx, input, aiConfig, resp)
			if repairErr == nil && repairedResp != nil {
				if repairedAnalysis, parseRepairErr := ParseMetadataOutput(repairedResp.Output); parseRepairErr == nil {
					trace := buildMetadataTrace(input, aiConfig, availableNames, resp)
					trace.RawOutput = strings.TrimSpace(repairedResp.Output)
					return &MetadataAnalysisResult{
						Analysis: repairedAnalysis,
						Trace:    trace,
					}, nil
				} else {
					repairErr = fmt.Errorf("parse repaired metadata output: %w", parseRepairErr)
				}
			}
			trace := buildMetadataTrace(input, aiConfig, availableNames, resp)
			trace.FallbackReason = err.Error()
			if repairErr != nil {
				trace.FallbackReason += "; structured repair failed: " + repairErr.Error()
			}
			if resp.OutputDiagnostic != "" {
				trace.FallbackReason += "; model output: " + resp.OutputDiagnostic
			}
			return &MetadataAnalysisResult{
				Analysis: fallbackMetadataAnalysis(input),
				Trace:    trace,
			}, nil
		}
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("metadata runtime returned empty response")
	}

	analysis, err := ParseMetadataOutput(resp.Output)
	if err != nil {
		// 首轮输出可能夹杂解释文字或 Markdown；优先把搜索结果和首轮输出交给一次无工具修复请求，避免直接丢弃已获取的外部证据。
		repairedResp, repairErr := a.repairMetadataOutput(ctx, input, aiConfig, resp)
		if repairErr == nil && repairedResp != nil {
			if repairedAnalysis, parseRepairErr := ParseMetadataOutput(repairedResp.Output); parseRepairErr == nil {
				trace := buildMetadataTrace(input, aiConfig, availableNames, resp)
				trace.RawOutput = strings.TrimSpace(repairedResp.Output)
				return &MetadataAnalysisResult{
					Analysis: repairedAnalysis,
					Trace:    trace,
				}, nil
			} else {
				repairErr = fmt.Errorf("parse repaired metadata output: %w", parseRepairErr)
			}
		}

		fmt.Printf("[DEBUG] Web Search Import - Failed to parse output: %v\n", err)
		trace := buildMetadataTrace(input, aiConfig, availableNames, resp)
		trace.FallbackReason = fmt.Sprintf("parse metadata analysis output: %v", err)
		if repairErr != nil {
			trace.FallbackReason += "; structured repair failed: " + repairErr.Error()
		}
		if resp.OutputDiagnostic != "" {
			trace.FallbackReason += "; model output: " + resp.OutputDiagnostic
		}
		return &MetadataAnalysisResult{
			Analysis: fallbackMetadataAnalysis(input),
			Trace:    trace,
		}, nil
	}

	return &MetadataAnalysisResult{
		Analysis: analysis,
		Trace:    buildMetadataTrace(input, aiConfig, availableNames, resp),
	}, nil
}

// repairMetadataOutput performs one bounded, no-tool pass that converts the model's
// natural-language response plus tool evidence into the required metadata JSON.
func (a *EinoMetadataAgent) repairMetadataOutput(ctx context.Context, input *MetadataGenerationInput, aiConfig *models.AIConfig, original *AgentRunResponse) (*AgentRunResponse, error) {
	if original == nil {
		return nil, fmt.Errorf("original metadata response is nil")
	}

	toolEvidence := make([]string, 0, len(original.ToolCalls))
	for _, call := range original.ToolCalls {
		if call.Status != "success" || strings.TrimSpace(call.Output) == "" {
			continue
		}
		// 工具输出已经在运行时做过脱敏和长度限制，这里再次限制只为避免修复请求过大。
		toolEvidence = append(toolEvidence, fmt.Sprintf("tool=%s\ninput=%s\noutput=%s", call.Name, call.Input, call.Output))
	}
	if strings.TrimSpace(original.Output) == "" && len(toolEvidence) == 0 {
		return nil, fmt.Errorf("no model output or successful tool evidence to repair")
	}

	return a.runtime.Run(ctx, &AgentRunRequest{
		SystemPrompt: `You are a strict JSON repair step for file metadata analysis.
Use the supplied file context, original model response, and successful tool results.
Preserve facts from the tool results, but do not invent details that are not supported.
Return ONLY one valid JSON object with exactly these required keys:
{"tags":["tag1"],"description":"..."}
Do not use Markdown fences, explanations, citations outside the JSON, or extra keys.`,
		UserPrompt: buildMetadataRepairPrompt(input, original.Output, toolEvidence),
		AIConfig:   aiConfig,
		// 修复阶段不再需要工具选择，开启 JSON 模式让兼容的模型直接返回可解析对象。
		StructuredOutput: true,
	})
}

func buildMetadataRepairPrompt(input *MetadataGenerationInput, originalOutput string, toolEvidence []string) string {
	fileName, fileType, fileSubType := "", "", ""
	if input != nil {
		fileName = input.FileName
		fileType = input.FileType
		fileSubType = input.FileSubType
	}
	return fmt.Sprintf(`File context:
- name: %s
- type: %s
- subtype: %s

Original model response:
%s

Successful tool evidence:
%s

Convert the above into the required metadata JSON. The tool evidence is available evidence and must be considered before deciding the tags and description.`,
		fileName,
		fileType,
		fileSubType,
		strings.TrimSpace(originalOutput),
		strings.Join(toolEvidence, "\n\n"),
	)
}

func buildMetadataTrace(input *MetadataGenerationInput, aiConfig *models.AIConfig, availableNames []string, resp *AgentRunResponse) *MetadataTrace {
	evidenceSources := make([]string, 0)
	if input != nil {
		evidenceSources = make([]string, 0, len(input.Evidence))
		for _, item := range input.Evidence {
			if strings.TrimSpace(item.Source) == "" {
				continue
			}
			evidenceSources = append(evidenceSources, item.Source)
		}
	}

	trace := &MetadataTrace{
		ToolsAvailable:  availableNames,
		EvidenceSources: evidenceSources,
	}
	if aiConfig != nil {
		trace.AgentEnabled = aiConfig.Enabled && aiConfig.EnableAgent
	}
	if resp != nil {
		trace.ToolsUsed = resp.ToolsUsed
		trace.SearchQueries = resp.SearchQueries
		trace.ToolCalls = resp.ToolCalls
		trace.RawOutput = strings.TrimSpace(resp.Output)
		trace.OutputDiagnostic = resp.OutputDiagnostic
	}
	return trace
}

// fallbackMetadataAnalysis keeps the preview useful when the model's final
// response is malformed, while the trace still explains what actually happened.
func fallbackMetadataAnalysis(input *MetadataGenerationInput) *MetadataAnalysis {
	if input == nil {
		return &MetadataAnalysis{Tags: []string{}, Description: ""}
	}
	baseName := strings.TrimSpace(strings.TrimSuffix(input.FileName, filepath.Ext(input.FileName)))
	tags := make([]string, 0, 2)
	if strings.TrimSpace(input.FileType) != "" {
		tags = append(tags, strings.ToLower(strings.TrimSpace(input.FileType)))
	}
	if baseName != "" {
		tags = append(tags, baseName)
	}
	description := strings.TrimSpace(input.FileName)
	if input.FileType != "" && description != "" {
		description += " (" + strings.TrimSpace(input.FileType) + ")"
	}
	return &MetadataAnalysis{Tags: tags, Description: description}
}

func buildMetadataSystemPrompt(hasWebSearch, hasLocalEvidence bool) string {
	if hasWebSearch || hasLocalEvidence {
		return strings.TrimSpace(`You analyze one file and return strict JSON with keys "tags" and "description".

IMPORTANT: Prefer local file metadata, user-provided context, and extracted local evidence.
When you call a tool, you MUST inspect its returned result before producing the final answer. Treat successful web_search results as supplemental evidence: use relevant titles and content in the tags and description, and do not silently ignore them.
Use local evidence tools when the supplied evidence is insufficient. Use web_search when local evidence and filename are still insufficient, and never include local absolute paths or secrets in a query.

WORKFLOW:
1. Inspect the filename, file type, user context, and local metadata
2. If important information is still missing, use web_search with a sanitized query
3. Read and analyze every successful tool result; distinguish facts from guesses
4. Return final JSON like: {"tags":["tag1","tag2"],"description":"detailed description"}

Return STRICT JSON format only, no Markdown, no explanation, and no other text.`)
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
	metadataStr := "{}"
	evidenceStr := "[]"
	if input != nil {
		tagsStr = strings.Join(input.UserTags, ", ")
		// 只将经过筛选的本地元数据放入 Prompt，不暴露 FilePath、FileID 等后端字段。
		if nativeMetadata, err := json.Marshal(input.NativeFileMetadata()); err == nil {
			metadataStr = string(nativeMetadata)
		}
		if len(input.Evidence) > 0 {
			// Evidence 已在本地采集阶段限长，这里只做序列化，不把原始路径带入模型。
			if evidence, err := json.Marshal(input.Evidence); err == nil {
				evidenceStr = string(evidence)
			}
		}
		hint := ""
		// 只有在启用 web search 时提示模型按需补充证据，而不是强制搜索。
		if hasWebSearch && (input.FileType == "video" || input.FileType == "document" || input.FileType == "music") {
			hint = "\n提示：如果本地元数据和用户上下文仍不足以判断内容，再使用 web_search 补充证据。"
		}

		if hasWebSearch {
			return fmt.Sprintf(
				"文件名：%s\n文件类型：%s\n文件子类型：%s\n本地元数据：%s\n本地证据：%s\n用户关键词：%s\n用户标签：%s\n用户描述：%s%s\n请基于现有证据输出 metadata JSON。",
				input.FileName,
				input.FileType,
				input.FileSubType,
				metadataStr,
				evidenceStr,
				input.UserKeywords,
				tagsStr,
				input.UserDescription,
				hint,
			)
		}

		return fmt.Sprintf(
			"文件名：%s\n文件类型：%s\n文件子类型：%s\n本地元数据：%s\n本地证据：%s\n用户关键词：%s\n用户标签：%s\n用户描述：%s\n请直接输出 metadata JSON。",
			input.FileName,
			input.FileType,
			input.FileSubType,
			metadataStr,
			evidenceStr,
			input.UserKeywords,
			tagsStr,
			input.UserDescription,
		)
	}

	if hasWebSearch {
		return fmt.Sprintf(
			"文件名：%s\n文件类型：%s\n文件子类型：%s\n本地元数据：%s\n本地证据：%s\n用户关键词：%s\n用户标签：%s\n用户描述：%s\n请基于现有证据输出 metadata JSON。",
			"", "", "", metadataStr, evidenceStr, "", tagsStr, "",
		)
	}

	return fmt.Sprintf(
		"文件名：%s\n文件类型：%s\n文件子类型：%s\n本地元数据：%s\n本地证据：%s\n用户关键词：%s\n用户标签：%s\n用户描述：%s\n请直接输出 metadata JSON。",
		"", "", "", metadataStr, evidenceStr, "", tagsStr, "",
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

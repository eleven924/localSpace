package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"LocalSpace/app/agents"
	"LocalSpace/app/models"
	"LocalSpace/app/repositories"
	"LocalSpace/app/tools"
)

// AgentService coordinates metadata generation with runtime-configured tools.
type AgentService struct {
	metadataAgent agents.MetadataAgent
	metadataGraph *agents.MetadataAnalysisGraph
	configRepo    *repositories.ConfigRepository
	toolRegistry  *tools.ToolRegistry
}

// NewAgentService creates a new AgentService without eagerly registering tools.
func NewAgentService(configRepo *repositories.ConfigRepository) *AgentService {
	metadataAgent := agents.NewEinoMetadataAgent(agents.NewEinoReactAgentRuntime())
	return &AgentService{
		// 新入口默认使用 Eino ReAct Runtime；旧 Runtime 仍保留用于兼容和回归测试。
		// Graph 负责业务流程，Eino Agent 作为其中的局部分析节点。
		metadataAgent: metadataAgent,
		metadataGraph: agents.NewMetadataAnalysisGraph(metadataAgent),
		configRepo:    configRepo,
		toolRegistry:  tools.NewToolRegistry(),
	}
}

func (s *AgentService) ensureConfiguredTools(config *models.AIConfig) {
	if s == nil || s.toolRegistry == nil {
		return
	}

	webSearchTool := tools.NewConfiguredWebSearchTool(config)
	if err := s.toolRegistry.Register(webSearchTool.Name(), webSearchTool); err != nil {
		fmt.Printf("[DEBUG] Web Search Import - Tool registration error: %v\n", err)
	}
}

// Helper to mask API key in logs
func maskAPIKey(key string) string {
	if len(key) <= 4 {
		return "***"
	}
	return key[:2] + "***" + key[len(key)-2:]
}

func (s *AgentService) resolveMetadataTools(config *models.AIConfig, input *agents.MetadataGenerationInput) []tools.Tool {
	if s == nil || s.toolRegistry == nil || config == nil || input == nil {
		return nil
	}
	if !config.EnableAgent {
		return nil
	}

	resolved := make([]tools.Tool, 0, 4)
	// 本地证据工具绑定后端已校验的文件路径，模型只能选择证据类型，不能传入任意路径。
	if strings.TrimSpace(input.FilePath) != "" {
		resolved = append(resolved, tools.NewLocalEvidenceTools(input.FilePath, input.FileType, input.FileSubType)...)
	}

	if config.EnableWebSearch && isMetadataSearchCandidate(input) {
		if tool, ok := s.toolRegistry.Get("web_search"); ok && tool != nil {
			resolved = append(resolved, tool)
		}
	}
	return resolved
}

func (s *AgentService) analyzeMetadataWithConfig(ctx context.Context, input *agents.MetadataGenerationInput, config *models.AIConfig) (*agents.MetadataAnalysisResult, error) {
	runID := newMetadataRunID()
	if s == nil || s.metadataAgent == nil {
		return withMetadataTrace(fallbackMetadataResult(input, nil, "metadata agent is not configured"), runID, config, nil), nil
	}
	if input == nil {
		return withMetadataTrace(fallbackMetadataResult(nil, nil, "metadata input is nil"), runID, config, nil), nil
	}
	if config == nil {
		if s.configRepo == nil {
			return withMetadataTrace(fallbackMetadataResult(input, nil, "ai config repository is not configured"), runID, nil, nil), nil
		}
		var err error
		config, err = s.configRepo.GetAIConfig()
		if err != nil {
			return withMetadataTrace(fallbackMetadataResult(input, nil, "failed to get AI config: "+err.Error()), runID, nil, nil), nil
		}
	}

	s.ensureConfiguredTools(config)
	availableTools := s.resolveMetadataTools(config, input)
	var result *agents.MetadataAnalysisResult
	var err error
	if s.metadataGraph != nil {
		// Graph 负责路由、阶段记录和结果校验，Agent 只负责局部分析决策。
		result, err = s.metadataGraph.Run(ctx, input, config, availableTools)
	} else {
		// 保留可注入的旧路径，兼容已有测试和 Graph 不可用时的降级。
		result, err = s.metadataAgent.Analyze(ctx, input, config, availableTools)
	}
	if err != nil {
		return withMetadataTrace(fallbackMetadataResult(input, availableTools, err.Error()), runID, config, availableTools), nil
	}
	if result == nil {
		return withMetadataTrace(fallbackMetadataResult(input, availableTools, "metadata agent returned empty result"), runID, config, availableTools), nil
	}
	result = withMetadataTrace(result, runID, config, availableTools)
	return result, nil
}

// AnalyzeMetadata adapts agent-facing analysis into the file import contract.
func (s *AgentService) AnalyzeMetadata(ctx context.Context, input *agents.MetadataGenerationInput) (*agents.MetadataAnalysis, error) {
	result, err := s.AnalyzeMetadataWithTrace(ctx, input)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.Analysis, nil
}

// AnalyzeMetadataWithTrace is the unified metadata analysis entry for preview and import flows.
func (s *AgentService) AnalyzeMetadataWithTrace(ctx context.Context, input *agents.AnalysisRequest) (*agents.MetadataAnalysisResult, error) {
	return s.analyzeMetadataWithConfig(ctx, input, nil)
}

func newMetadataRunID() string {
	return fmt.Sprintf("metadata-%d", time.Now().UnixNano())
}

// withMetadataTrace adds run-level diagnostics without changing the existing fallback contract.
func withMetadataTrace(result *agents.MetadataAnalysisResult, runID string, config *models.AIConfig, availableTools []tools.Tool) *agents.MetadataAnalysisResult {
	if result == nil {
		result = fallbackMetadataResult(nil, availableTools, "metadata result is nil")
	}
	if result.Trace == nil {
		result.Trace = &agents.MetadataTrace{}
	}
	if result.Trace.RunID == "" {
		result.Trace.RunID = runID
	}
	if result.Trace.Status == "" {
		if result.Trace.FallbackReason != "" {
			result.Trace.Status = "fallback"
		} else {
			result.Trace.Status = "completed"
		}
	}
	if result.Trace.FallbackReason != "" {
		// fallback 结果可以帮助导入继续，但不能被质量门禁当作可信 AI 建议自动应用。
		result.Trace.Confidence = 0
		result.Trace.NeedsReview = true
		result.Trace.QualityStatus = "low_confidence"
	}
	if config != nil {
		result.Trace.Model = config.Model
		result.Trace.AgentEnabled = config.Enabled && config.EnableAgent
	}
	if len(result.Trace.ToolsAvailable) == 0 {
		result.Trace.ToolsAvailable = toolNames(availableTools)
	}
	return result
}

func isMetadataSearchCandidate(input *agents.MetadataGenerationInput) bool {
	if input == nil {
		return false
	}
	fileType := strings.ToLower(strings.TrimSpace(input.FileType))
	switch fileType {
	case "video", "document", "music", "archive", "installer", "image":
		return true
	}

	ext := strings.ToLower(filepath.Ext(input.FileName))
	switch ext {
	case ".mp4", ".mkv", ".avi", ".mov", ".pdf", ".doc", ".docx", ".mp3", ".flac":
		return true
	default:
		return false
	}
}

func fallbackMetadataResult(input *agents.MetadataGenerationInput, availableTools []tools.Tool, reason string) *agents.MetadataAnalysisResult {
	evidenceSources := make([]string, 0)
	if input != nil {
		for _, item := range input.Evidence {
			if strings.TrimSpace(item.Source) != "" {
				evidenceSources = append(evidenceSources, item.Source)
			}
		}
	}
	result := &agents.MetadataAnalysisResult{
		Analysis: &agents.MetadataAnalysis{Tags: []string{}, Description: ""},
		Trace: &agents.MetadataTrace{
			ToolsAvailable:  toolNames(availableTools),
			EvidenceSources: evidenceSources,
			FallbackReason:  reason,
		},
	}
	if input != nil {
		result.Analysis.Tags = fallbackTags(input)
		result.Analysis.Description = fallbackDescription(input)
	}
	return result
}

func toolNames(availableTools []tools.Tool) []string {
	names := make([]string, 0, len(availableTools))
	for _, tool := range availableTools {
		if tool == nil {
			continue
		}
		names = append(names, tool.Name())
	}
	return names
}

func fallbackTags(input *agents.MetadataGenerationInput) []string {
	if input == nil {
		return []string{}
	}
	name := strings.TrimSuffix(input.FileName, filepath.Ext(input.FileName))
	name = strings.TrimSpace(name)
	if name == "" {
		return []string{}
	}
	if input.FileType != "" {
		return []string{strings.ToLower(input.FileType), name}
	}
	return []string{name}
}

func fallbackDescription(input *agents.MetadataGenerationInput) string {
	if input == nil {
		return ""
	}
	if input.FileType == "" {
		return input.FileName
	}
	return input.FileName + " (" + input.FileType + ")"
}

package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"LocalSpace/app/models"
	"LocalSpace/app/tools"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

// EinoReactAgentRuntime uses Eino's ReAct agent and ToolsNode for tool orchestration.
// DefaultAgentRuntime remains available as a compatibility implementation during migration.
type EinoReactAgentRuntime struct {
	MaxStep           int
	MaxWebSearchCalls int
}

// NewEinoReactAgentRuntime creates the runtime used by the new AgentService path.
func NewEinoReactAgentRuntime() *EinoReactAgentRuntime {
	return &EinoReactAgentRuntime{
		MaxStep:           8,
		MaxWebSearchCalls: 2,
	}
}

// Run executes a bounded Eino ReAct workflow and returns the final assistant message.
func (r *EinoReactAgentRuntime) Run(ctx context.Context, req *AgentRunRequest) (*AgentRunResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("agent run request is nil")
	}
	if req.AIConfig == nil {
		return nil, fmt.Errorf("ai config is nil")
	}
	if req.AIConfig.APIKey == "" || req.AIConfig.Model == "" || req.AIConfig.BaseURL == "" {
		return nil, fmt.Errorf("invalid ai config")
	}

	chatModel, err := newReactChatModel(ctx, req.AIConfig, req.StructuredOutput)
	if err != nil {
		return nil, err
	}

	messages := []*schema.Message{
		schema.SystemMessage(req.SystemPrompt),
		schema.UserMessage(req.UserPrompt),
	}
	if len(req.Tools) == 0 {
		// 没有工具时直接调用模型，避免为纯文本请求构造不必要的 Agent 图。
		message, err := chatModel.Generate(ctx, messages)
		// 一些 OpenAI 兼容服务不支持 response_format；结构化修复仍回退一次普通请求，避免直接丢弃搜索证据。
		if err != nil && req.StructuredOutput {
			chatModel, fallbackErr := newReactChatModel(ctx, req.AIConfig, false)
			if fallbackErr == nil {
				message, err = chatModel.Generate(ctx, messages)
			}
		}
		if err != nil {
			return nil, fmt.Errorf("failed to generate response: %w", err)
		}
		response := reactAgentResponse(message, nil, nil)
		if strings.TrimSpace(response.Output) == "" {
			return response, fmt.Errorf("model returned empty assistant content: %s", response.OutputDiagnostic)
		}
		return response, nil
	}

	einoTools := make([]tool.BaseTool, 0, len(req.Tools))
	for _, localTool := range req.Tools {
		if localTool == nil {
			continue
		}
		einoTools = append(einoTools, &localToolAdapter{local: localTool})
	}
	if len(einoTools) == 0 {
		return nil, fmt.Errorf("no valid tools configured")
	}

	trace := &reactToolTrace{}
	maxStep := r.MaxStep
	if maxStep <= 0 {
		maxStep = 8
	}
	maxWebSearchCalls := r.MaxWebSearchCalls
	if maxWebSearchCalls <= 0 {
		maxWebSearchCalls = 2
	}

	// 通过 Eino middleware 记录工具调用，并在业务层保留按工具限流策略。
	middleware := compose.ToolMiddleware{
		Invokable: func(next compose.InvokableToolEndpoint) compose.InvokableToolEndpoint {
			return func(callCtx context.Context, input *compose.ToolInput) (*compose.ToolOutput, error) {
				startedAt := time.Now()
				if input.Name == "web_search" && trace.webSearchCalls() >= maxWebSearchCalls {
					err := fmt.Errorf("tool call limit reached for %s", input.Name)
					trace.record(input.Name, input.Arguments, "", err, time.Since(startedAt))
					return nil, err
				}
				output, err := next(callCtx, input)
				if err != nil {
					trace.record(input.Name, input.Arguments, "", err, time.Since(startedAt))
					return nil, err
				}
				trace.record(input.Name, input.Arguments, output.Result, nil, time.Since(startedAt))
				return output, nil
			}
		},
	}

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools:               einoTools,
			ExecuteSequentially: true,
			ToolCallMiddlewares: []compose.ToolMiddleware{middleware},
		},
		MaxStep: maxStep,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Eino ReAct agent: %w", err)
	}

	// Agent 内部负责模型、ToolCall、ToolMessage 和循环状态推进。
	message, err := agent.Generate(ctx, messages)
	if err != nil {
		// 返回已完成的 trace，即使最终一轮失败，调用详情仍可被上层展示。
		return reactAgentResponse(nil, trace.toolsUsed(), trace.searchQueries(), trace.toolCalls()), fmt.Errorf("failed to run Eino ReAct agent: %w", err)
	}
	response := reactAgentResponse(message, trace.toolsUsed(), trace.searchQueries(), trace.toolCalls())
	if strings.TrimSpace(response.Output) == "" {
		return response, fmt.Errorf("Eino ReAct returned empty assistant content: %s", response.OutputDiagnostic)
	}
	return response, nil
}

func newReactChatModel(ctx context.Context, config *models.AIConfig, structuredOutput bool) (*openai.ChatModel, error) {
	chatConfig := &openai.ChatModelConfig{
		APIKey:  config.APIKey,
		Model:   config.Model,
		BaseURL: config.BaseURL,
	}
	if config.Timeout > 0 {
		chatConfig.Timeout = time.Duration(config.Timeout) * time.Second
	}
	if config.MaxTokens > 0 {
		maxTokens := config.MaxTokens
		chatConfig.MaxTokens = &maxTokens
	}
	temperature := float32(0.2)
	chatConfig.Temperature = &temperature
	if structuredOutput {
		// 修复请求只需要 JSON 对象，使用 provider 支持度最高的 json_object 模式。
		chatConfig.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		}
	}

	chatModel, err := openai.NewChatModel(ctx, chatConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}
	return chatModel, nil
}

func reactAgentResponse(message *schema.Message, toolsUsed, searchQueries []string, toolCalls ...[]models.AIToolCall) *AgentRunResponse {
	var calls []models.AIToolCall
	if len(toolCalls) > 0 {
		calls = toolCalls[0]
	}
	if message == nil {
		return &AgentRunResponse{ToolsUsed: toolsUsed, SearchQueries: searchQueries, ToolCalls: calls}
	}
	output := assistantOutputText(message)
	return &AgentRunResponse{
		Output:           output,
		OutputDiagnostic: assistantOutputDiagnostic(message, output),
		ToolsUsed:        toolsUsed,
		SearchQueries:    searchQueries,
		ToolCalls:        calls,
	}
}

// assistantOutputText extracts assistant text; reasoning content is diagnostic data,
// not a safe substitute for the metadata JSON that the application must parse.
func assistantOutputText(message *schema.Message) string {
	if message == nil {
		return ""
	}
	if strings.TrimSpace(message.Content) != "" {
		return strings.TrimSpace(message.Content)
	}
	parts := make([]string, 0, len(message.AssistantGenMultiContent))
	for _, part := range message.AssistantGenMultiContent {
		if part.Type == schema.ChatMessagePartTypeText && strings.TrimSpace(part.Text) != "" {
			parts = append(parts, strings.TrimSpace(part.Text))
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func assistantOutputDiagnostic(message *schema.Message, output string) string {
	if message == nil || strings.TrimSpace(output) != "" {
		return ""
	}
	parts := []string{"content_empty"}
	if strings.TrimSpace(message.ReasoningContent) != "" {
		parts = append(parts, "reasoning_content_present")
	}
	if len(message.AssistantGenMultiContent) > 0 {
		parts = append(parts, fmt.Sprintf("multi_content_parts=%d", len(message.AssistantGenMultiContent)))
	}
	if len(message.ToolCalls) > 0 {
		parts = append(parts, fmt.Sprintf("tool_calls=%d", len(message.ToolCalls)))
	}
	if message.ResponseMeta != nil && message.ResponseMeta.FinishReason != "" {
		parts = append(parts, "finish_reason="+message.ResponseMeta.FinishReason)
	}
	return strings.Join(parts, ", ")
}

// localToolAdapter bridges the existing LocalSpace tool contract to Eino's InvokableTool.
type localToolAdapter struct {
	local tools.Tool
}

func (a *localToolAdapter) Info(context.Context) (*schema.ToolInfo, error) {
	if a == nil || a.local == nil {
		return nil, fmt.Errorf("local tool is nil")
	}
	if a.local.Name() == "web_search" {
		return &schema.ToolInfo{
			Name: a.local.Name(),
			Desc: a.local.Description(),
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"query": {Type: schema.String, Desc: "The web search query", Required: true},
				"limit": {Type: schema.Integer, Desc: "Maximum number of results", Required: false},
			}),
		}, nil
	}
	return &schema.ToolInfo{
		Name: a.local.Name(),
		Desc: a.local.Description(),
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"input": {Type: schema.String, Desc: "The input for the tool", Required: true},
		}),
	}, nil
}

func (a *localToolAdapter) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	if a == nil || a.local == nil {
		return "", fmt.Errorf("local tool is nil")
	}
	input, err := normalizeToolArguments(a.local.Name(), argumentsInJSON)
	if err != nil {
		return "", err
	}
	return a.local.Execute(ctx, input)
}

func normalizeToolArguments(toolName, arguments string) (string, error) {
	trimmed := strings.TrimSpace(arguments)
	if trimmed == "" {
		return "", fmt.Errorf("tool arguments are empty")
	}

	if toolName == "web_search" {
		var payload struct {
			Query string `json:"query"`
			Input string `json:"input"`
		}
		if json.Unmarshal([]byte(trimmed), &payload) == nil {
			query := strings.TrimSpace(payload.Query)
			if query == "" {
				query = strings.TrimSpace(payload.Input)
			}
			if query == "" {
				return "", fmt.Errorf("web_search arguments missing query")
			}
			return query, nil
		}
	}

	var payload struct {
		Input string `json:"input"`
	}
	if json.Unmarshal([]byte(trimmed), &payload) == nil && strings.TrimSpace(payload.Input) != "" {
		return strings.TrimSpace(payload.Input), nil
	}
	return trimmed, nil
}

type reactToolTrace struct {
	mu      sync.Mutex
	used    []string
	queries []string
	calls   []models.AIToolCall
}

// record stores a bounded tool call after Eino has completed the endpoint.
// 这里同时记录失败调用，确保“调用过但失败”的网络搜索也可被定位。
func (t *reactToolTrace) record(name, arguments, output string, callErr error, duration time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	call := models.AIToolCall{
		Name:       name,
		Input:      sanitizeToolTraceValue(toolTraceInput(name, arguments)),
		Output:     sanitizeToolTraceValue(output),
		Status:     "success",
		DurationMs: duration.Milliseconds(),
	}
	if callErr != nil {
		call.Status = "failed"
		call.Error = sanitizeToolTraceValue(callErr.Error())
	} else {
		t.used = append(t.used, name)
	}
	t.calls = append(t.calls, call)
	if name == "web_search" {
		if query, err := normalizeToolArguments(name, arguments); err == nil && strings.TrimSpace(query) != "" {
			t.queries = append(t.queries, query)
		}
	}
}

func toolTraceInput(name, arguments string) string {
	if input, err := normalizeToolArguments(name, arguments); err == nil {
		return input
	}
	return arguments
}

func (t *reactToolTrace) webSearchCalls() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	count := 0
	for _, name := range t.used {
		if name == "web_search" {
			count++
		}
	}
	return count
}

func (t *reactToolTrace) toolsUsed() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]string(nil), t.used...)
}

func (t *reactToolTrace) searchQueries() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]string(nil), t.queries...)
}

func (t *reactToolTrace) toolCalls() []models.AIToolCall {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]models.AIToolCall(nil), t.calls...)
}

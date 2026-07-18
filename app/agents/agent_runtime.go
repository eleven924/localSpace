package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"LocalSpace/app/models"
	"LocalSpace/app/tools"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

// AgentRunRequest describes one metadata runtime invocation.
type AgentRunRequest struct {
	SystemPrompt string
	UserPrompt   string
	Tools        []tools.Tool
	AIConfig     *models.AIConfig
}

// AgentRunResponse captures the final model output and runtime tool trace.
type AgentRunResponse struct {
	Output        string
	ToolsUsed     []string
	SearchQueries []string
}

// AgentRuntime is the execution contract used by metadata agents.
type AgentRuntime interface {
	Run(ctx context.Context, req *AgentRunRequest) (*AgentRunResponse, error)
}

// DefaultAgentRuntime runs a bounded model -> tool -> model loop.
type DefaultAgentRuntime struct {
	generate func(ctx context.Context, messages []*schema.Message, req *AgentRunRequest, toolInfos []*schema.ToolInfo) (*schema.Message, error)
}

// NewDefaultAgentRuntime creates the default metadata runtime.
func NewDefaultAgentRuntime() *DefaultAgentRuntime {
	runtime := &DefaultAgentRuntime{}
	runtime.generate = runtime.generateWithChatModel
	return runtime
}

func (r *DefaultAgentRuntime) generateWithChatModel(ctx context.Context, messages []*schema.Message, req *AgentRunRequest, toolInfos []*schema.ToolInfo) (*schema.Message, error) {
	chatConfig := &openai.ChatModelConfig{
		APIKey:  req.AIConfig.APIKey,
		Model:   req.AIConfig.Model,
		BaseURL: req.AIConfig.BaseURL,
	}
	if req.AIConfig.Timeout > 0 {
		chatConfig.Timeout = time.Duration(req.AIConfig.Timeout) * time.Second
	}
	if req.AIConfig.MaxTokens > 0 {
		maxTokens := req.AIConfig.MaxTokens
		chatConfig.MaxTokens = &maxTokens
	}
	temperature := float32(0.2)
	chatConfig.Temperature = &temperature

	chatModel, err := openai.NewChatModel(ctx, chatConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}

	var resp *schema.Message
	if len(toolInfos) > 0 {
		// Use tool calling model when tools are available
		toolModel, err := chatModel.WithTools(toolInfos)
		if err != nil {
			return nil, fmt.Errorf("failed to create tool calling model: %w", err)
		}
		resp, err = toolModel.Generate(ctx, messages)
		if err != nil {
			return nil, fmt.Errorf("failed to generate response with tools: %w", err)
		}
	} else {
		// Use regular model when no tools are available
		resp, err = chatModel.Generate(ctx, messages)
		if err != nil {
			return nil, fmt.Errorf("failed to generate response: %w", err)
		}
	}

	// Return the raw AssistantMessage (may contain ToolCalls)
	return resp, nil
}

// Run executes a bounded runtime loop for metadata generation.
func (r *DefaultAgentRuntime) Run(ctx context.Context, req *AgentRunRequest) (*AgentRunResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("agent run request is nil")
	}
	if req.AIConfig == nil {
		return nil, fmt.Errorf("ai config is nil")
	}
	if req.AIConfig.APIKey == "" || req.AIConfig.Model == "" || req.AIConfig.BaseURL == "" {
		return nil, fmt.Errorf("invalid ai config")
	}
	if r.generate == nil {
		r.generate = r.generateWithChatModel
	}

	messages := []*schema.Message{
		schema.SystemMessage(req.SystemPrompt),
		schema.UserMessage(req.UserPrompt),
	}
	toolMap := buildToolMap(req.Tools)
	toolInfos := buildToolInfos(req.Tools)
	toolsUsed := []string{}
	searchQueries := []string{}

	assistantMsg, err := r.generate(ctx, messages, req, toolInfos)
	if err != nil {
		return nil, err
	}
	messages = append(messages, assistantMsg)

	maxRounds := 4
	webSearchCalls := 0
	const maxWebSearchCalls = 2
	for len(assistantMsg.ToolCalls) > 0 && maxRounds > 0 {
		// Process all tool calls in this batch
		for _, toolCall := range assistantMsg.ToolCalls {
			toolName := toolCall.Function.Name
			toolInput, err := decodeToolInput(toolCall.Function.Arguments)
			if err != nil {
				return nil, fmt.Errorf("decode tool input for %s: %w", toolName, err)
			}

			if toolName == "web_search" && webSearchCalls >= maxWebSearchCalls {
				fmt.Printf("[DEBUG] Web Search Import - Web search limit reached after 2 calls; forcing final response\n")
				continue
			}

			toolOutput, updatedToolsUsed, updatedSearchQueries, err := r.executeToolCall(ctx, toolMap, toolName, toolInput, toolsUsed, searchQueries)
			if err != nil {
				return nil, err
			}
			toolsUsed = updatedToolsUsed
			searchQueries = updatedSearchQueries
			if toolName == "web_search" {
				webSearchCalls++
			}

			messages = append(messages, schema.ToolMessage(toolOutput, toolCall.ID))
		}

		currentToolInfos := toolInfos
		if webSearchCalls >= maxWebSearchCalls {
			currentToolInfos = nil
		}
		assistantMsg, err = r.generate(ctx, messages, req, currentToolInfos)
		if err != nil {
			return nil, err
		}
		messages = append(messages, assistantMsg)
		maxRounds--

		if webSearchCalls >= maxWebSearchCalls {
			break
		}
	}

	if len(assistantMsg.ToolCalls) > 0 && webSearchCalls < maxWebSearchCalls {
		return nil, fmt.Errorf("tool loop exceeded max rounds")
	}
	if assistantMsg.Content == "" {
		return nil, fmt.Errorf("empty response from AI")
	}

	return &AgentRunResponse{
		Output:        assistantMsg.Content,
		ToolsUsed:     toolsUsed,
		SearchQueries: searchQueries,
	}, nil
}

func (r *DefaultAgentRuntime) executeToolCall(
	ctx context.Context,
	toolMap map[string]tools.Tool,
	toolName string,
	input string,
	toolsUsed []string,
	searchQueries []string,
) (string, []string, []string, error) {
	tool, ok := toolMap[toolName]
	if !ok {
		return "", toolsUsed, searchQueries, fmt.Errorf("tool not found: %s", toolName)
	}

	output, err := tool.Execute(ctx, input)
	if err != nil {
		return "", toolsUsed, searchQueries, fmt.Errorf("execute tool %s: %w", toolName, err)
	}

	toolsUsed = append(toolsUsed, toolName)
	if toolName == "web_search" && strings.TrimSpace(input) != "" {
		searchQueries = append(searchQueries, strings.TrimSpace(input))
	}

	return output, toolsUsed, searchQueries, nil
}

func decodeToolInput(arguments string) (string, error) {
	trimmed := strings.TrimSpace(arguments)
	if trimmed == "" {
		return "", fmt.Errorf("tool arguments are empty")
	}

	var payload struct {
		Input string `json:"input"`
	}
	if json.Unmarshal([]byte(trimmed), &payload) == nil {
		payload.Input = strings.TrimSpace(payload.Input)
		if payload.Input == "" {
			return "", fmt.Errorf("tool arguments missing input")
		}
		return payload.Input, nil
	}

	return trimmed, nil
}

func buildToolMap(availableTools []tools.Tool) map[string]tools.Tool {
	toolMap := make(map[string]tools.Tool, len(availableTools))
	for _, tool := range availableTools {
		if tool == nil {
			continue
		}
		toolMap[tool.Name()] = tool
	}
	return toolMap
}

func buildToolInfos(availableTools []tools.Tool) []*schema.ToolInfo {
	toolInfos := make([]*schema.ToolInfo, 0, len(availableTools))
	for _, tool := range availableTools {
		if tool == nil {
			continue
		}

		// Create parameters for tool
		params := map[string]*schema.ParameterInfo{
			"input": {
				Type:     schema.String,
				Desc:     "The input query for the tool",
				Required: true,
			},
		}

		toolInfo := &schema.ToolInfo{
			Name:        tool.Name(),
			Desc:        tool.Description(),
			ParamsOneOf: schema.NewParamsOneOfByParams(params),
		}
		toolInfos = append(toolInfos, toolInfo)
	}
	return toolInfos
}

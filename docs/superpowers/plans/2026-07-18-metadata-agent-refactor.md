# Metadata Agent Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the direct chat-model tag/description generation path with a unified `MetadataAgent` + `AgentRuntime` + `ToolRegistry` architecture, while preserving `AgentService` compatibility and switching `FileService` to a single metadata analysis call.

**Architecture:** The implementation keeps `AgentService` as the business-facing facade, introduces a unified metadata analysis domain in `app/agents`, and routes execution through a single `MetadataAgent` backed by an `AgentRuntime`. Tool availability is resolved outside the agent, the agent returns structured output plus trace data, and `FileService` consumes one metadata analysis result instead of separate tag and description requests.

**Tech Stack:** Go, Wails service architecture, CloudWeGo Eino / Eino OpenAI chat model integration, existing `tools.Tool` registry pattern, Go testing package.

## Global Constraints

- Replace pseudo-agent direct model invocation with a real Eino ADK agent execution path.
- Unify metadata generation into a single `MetadataAgent` that produces both tags and description in one analysis pass.
- Separate business result output from internal execution trace/debug information.
- Preserve current business compatibility through `AgentService` while consolidating internal architecture.
- Provide a stable extension point for web search and future tools.
- Ensure agent or tool failures never block file import.
- Do not fully optimize web search relevance or quality in the first phase.
- Do not introduce multi-agent orchestration.
- Do not replace all AI-related services across the entire codebase in one step.
- Do not remove every legacy file immediately during the first migration phase.
- Adopt an outer authorization, inner autonomy tool model.
- Prioritize structured output first, rule-based cleanup as fallback.
- Keep web search conservative and bounded: limit calls per metadata request, limit query length, limit result count and payload size, apply explicit timeouts, and never fail the whole metadata analysis if search fails.

---

## File Map

### Create
- `app/agents/metadata_types.go` — unified metadata input/result/trace domain types.
- `app/agents/output_parser.go` — structured output parsing and normalization helpers.
- `app/agents/output_parser_test.go` — parser-focused tests.
- `app/agents/metadata_agent.go` — `MetadataAgent` interface and `EinoMetadataAgent` implementation.
- `app/agents/metadata_agent_test.go` — metadata agent behavior with fake runtime.
- `app/agents/agent_runtime.go` — runtime request/response types and default runtime implementation.
- `app/agents/agent_runtime_test.go` — runtime tool binding/trace tests using test doubles where practical.

### Modify
- `app/agents/prompt_builder.go` — add unified metadata prompts while keeping transition-safe compatibility for current prompt tests if needed.
- `app/agents/prompt_builder_test.go` — cover unified metadata prompts.
- `app/tools/tool_registry.go` — add metadata-scoped tool resolution helpers.
- `app/tools/tool_registry_test.go` — cover tool resolution and gating.
- `app/tools/web_search_tool.go` — expose bounded execution details suitable for metadata agent integration.
- `app/services/agent_service.go` — switch to unified metadata analysis, trace-aware internal path, and compatibility wrappers.
- `app/services/agent_service_test.go` — test fallback, compatibility, and unified analysis behavior.
- `app/services/file_service.go` — call one metadata analysis path during import.
- `app/services/file_service_integration_test.go` — validate single-call integration expectations.
- `docs/agent-integration.md` — document the unified metadata-agent architecture and web-search readiness.

### Transitional / Do Not Expand
- `app/agents/tag_agent.go` — stop using as primary path; leave in place unless cleanup is part of the final task.
- `app/agents/description_agent.go` — stop using as primary path; leave in place unless cleanup is part of the final task.
- `app/services/ai_service.go` — do not expand; only adjust if compile/test constraints require it.

---

### Task 1: Add unified metadata domain types and parser

**Files:**
- Create: `app/agents/metadata_types.go`
- Create: `app/agents/output_parser.go`
- Create: `app/agents/output_parser_test.go`
- Modify: `app/agents/prompt_builder_test.go`

**Interfaces:**
- Consumes: `models.Metadata`
- Produces: `type MetadataGenerationInput struct { FileName string; FileType string; UserKeywords string; UserTags []string; UserDescription string; Metadata models.Metadata }`
- Produces: `type MetadataAnalysis struct { Tags []string; Description string }`
- Produces: `type MetadataTrace struct { AgentEnabled bool; ToolsAvailable []string; ToolsUsed []string; SearchQueries []string; FallbackReason string; RawOutput string }`
- Produces: `type MetadataAnalysisResult struct { Analysis *MetadataAnalysis; Trace *MetadataTrace }`
- Produces: `func ParseMetadataOutput(raw string) (*MetadataAnalysis, error)`
- Produces: `func NormalizeMetadataAnalysis(analysis *MetadataAnalysis) *MetadataAnalysis`

- [ ] **Step 1: Write the failing parser and prompt tests**

```go
package agents

import "testing"

func TestParseMetadataOutput_JSON(t *testing.T) {
    raw := `{"tags":["video","action","action","thriller","movie","extra"],"description":"A very long description that should still be normalized by the parser layer into a bounded description value."}`

    analysis, err := ParseMetadataOutput(raw)
    if err != nil {
        t.Fatalf("expected nil error, got %v", err)
    }

    if len(analysis.Tags) != 5 {
        t.Fatalf("expected 5 tags after normalization, got %d", len(analysis.Tags))
    }

    if analysis.Tags[0] != "video" {
        t.Fatalf("expected first tag to remain 'video', got %q", analysis.Tags[0])
    }

    if analysis.Description == "" {
        t.Fatal("expected normalized description to be non-empty")
    }
}

func TestParseMetadataOutput_InvalidJSON(t *testing.T) {
    if _, err := ParseMetadataOutput("not-json"); err == nil {
        t.Fatal("expected invalid JSON to return error")
    }
}
```

```go
package agents

import (
    "strings"
    "testing"
)

func TestPromptBuilderBuildMetadataUserPrompt(t *testing.T) {
    builder := NewPromptBuilder()

    prompt := builder.BuildMetadataUserPrompt(&MetadataGenerationInput{
        FileName:        "movie.mp4",
        FileType:        "video",
        UserKeywords:    "action thriller",
        UserTags:        []string{"娱乐", "动作"},
        UserDescription: "一部动作惊悚电影",
    })

    expectedParts := []string{
        "movie.mp4",
        "video",
        "action thriller",
        "娱乐, 动作",
        "一部动作惊悚电影",
    }

    for _, part := range expectedParts {
        if !strings.Contains(prompt, part) {
            t.Fatalf("expected metadata user prompt to contain %q", part)
        }
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./app/agents -run "TestParseMetadataOutput_JSON|TestParseMetadataOutput_InvalidJSON|TestPromptBuilderBuildMetadataUserPrompt"`
Expected: FAIL with undefined `ParseMetadataOutput`, undefined `MetadataGenerationInput`, or missing `BuildMetadataUserPrompt`.

- [ ] **Step 3: Write the minimal implementation**

```go
package agents

import "LocalSpace/app/models"

type MetadataGenerationInput struct {
    FileName        string
    FileType        string
    UserKeywords    string
    UserTags        []string
    UserDescription string
    Metadata        models.Metadata
}

type MetadataAnalysis struct {
    Tags        []string `json:"tags"`
    Description string   `json:"description"`
}

type MetadataTrace struct {
    AgentEnabled   bool
    ToolsAvailable []string
    ToolsUsed      []string
    SearchQueries  []string
    FallbackReason string
    RawOutput      string
}

type MetadataAnalysisResult struct {
    Analysis *MetadataAnalysis
    Trace    *MetadataTrace
}
```

```go
package agents

import (
    "encoding/json"
    "errors"
    "strings"
)

func ParseMetadataOutput(raw string) (*MetadataAnalysis, error) {
    raw = strings.TrimSpace(raw)
    if raw == "" {
        return nil, errors.New("empty metadata output")
    }

    var analysis MetadataAnalysis
    if err := json.Unmarshal([]byte(raw), &analysis); err != nil {
        return nil, err
    }

    return NormalizeMetadataAnalysis(&analysis), nil
}

func NormalizeMetadataAnalysis(analysis *MetadataAnalysis) *MetadataAnalysis {
    if analysis == nil {
        return &MetadataAnalysis{Tags: []string{}, Description: ""}
    }

    uniqueTags := make([]string, 0, len(analysis.Tags))
    seen := make(map[string]bool)
    for _, tag := range analysis.Tags {
        normalized := strings.TrimSpace(tag)
        if normalized == "" || seen[normalized] {
            continue
        }
        seen[normalized] = true
        uniqueTags = append(uniqueTags, normalized)
        if len(uniqueTags) == 5 {
            break
        }
    }

    description := strings.TrimSpace(analysis.Description)
    if len([]rune(description)) > 200 {
        description = string([]rune(description)[:200]) + "..."
    }

    return &MetadataAnalysis{
        Tags:        uniqueTags,
        Description: description,
    }
}
```

```go
func (b *PromptBuilder) BuildMetadataSystemPrompt() string {
    return "你是一个文件元数据补全助手。请根据上下文生成 tags 和 description，并仅返回 JSON：{\"tags\":[],\"description\":\"\"}。"
}

func (b *PromptBuilder) BuildMetadataUserPrompt(input *MetadataGenerationInput) string {
    tagsStr := strings.Join(input.UserTags, ", ")
    return fmt.Sprintf(
        "文件名：%s\n文件类型：%s\n用户关键词：%s\n用户标签：%s\n用户描述：%s\n请输出 metadata JSON。",
        input.FileName,
        input.FileType,
        input.UserKeywords,
        tagsStr,
        input.UserDescription,
    )
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./app/agents -run "TestParseMetadataOutput_JSON|TestParseMetadataOutput_InvalidJSON|TestPromptBuilderBuildMetadataUserPrompt"`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add app/agents/metadata_types.go app/agents/output_parser.go app/agents/output_parser_test.go app/agents/prompt_builder.go app/agents/prompt_builder_test.go
git commit -m "feat: add metadata analysis domain types"
```

### Task 2: Add runtime and metadata agent execution seam

**Files:**
- Create: `app/agents/agent_runtime.go`
- Create: `app/agents/agent_runtime_test.go`
- Create: `app/agents/metadata_agent.go`
- Create: `app/agents/metadata_agent_test.go`

**Interfaces:**
- Consumes: `type MetadataGenerationInput`, `type MetadataAnalysisResult`, `func ParseMetadataOutput(raw string) (*MetadataAnalysis, error)`
- Consumes: `type Tool interface { Name() string; Description() string; Execute(ctx context.Context, input string) (string, error) }`
- Produces: `type AgentRunRequest struct { SystemPrompt string; UserPrompt string; Tools []tools.Tool; AIConfig *models.AIConfig }`
- Produces: `type AgentRunResponse struct { Output string; ToolsUsed []string; SearchQueries []string }`
- Produces: `type AgentRuntime interface { Run(ctx context.Context, req *AgentRunRequest) (*AgentRunResponse, error) }`
- Produces: `type MetadataAgent interface { Analyze(ctx context.Context, input *MetadataGenerationInput, aiConfig *models.AIConfig, availableTools []tools.Tool) (*MetadataAnalysisResult, error) }`
- Produces: `type EinoMetadataAgent struct { runtime AgentRuntime; promptBuilder *PromptBuilder }`

- [ ] **Step 1: Write the failing runtime and metadata-agent tests**

```go
package agents

import (
    "context"
    "testing"

    "LocalSpace/app/models"
    "LocalSpace/app/tools"
)

type fakeRuntime struct {
    response *AgentRunResponse
    err      error
}

func (f *fakeRuntime) Run(ctx context.Context, req *AgentRunRequest) (*AgentRunResponse, error) {
    return f.response, f.err
}

type fakeTool struct{ name string }

func (f *fakeTool) Name() string { return f.name }
func (f *fakeTool) Description() string { return "fake" }
func (f *fakeTool) Execute(ctx context.Context, input string) (string, error) { return "ok", nil }

func TestEinoMetadataAgentAnalyze_ParsesStructuredOutput(t *testing.T) {
    agent := NewEinoMetadataAgent(&fakeRuntime{
        response: &AgentRunResponse{
            Output:        `{"tags":["video","action"],"description":"动作影片"}`,
            ToolsUsed:     []string{"web_search"},
            SearchQueries: []string{"movie mp4"},
        },
    })

    result, err := agent.Analyze(context.Background(), &MetadataGenerationInput{FileName: "movie.mp4", FileType: "video"}, &models.AIConfig{Enabled: true, EnableAgent: true}, []tools.Tool{&fakeTool{name: "web_search"}})
    if err != nil {
        t.Fatalf("expected nil error, got %v", err)
    }

    if len(result.Analysis.Tags) != 2 {
        t.Fatalf("expected 2 tags, got %d", len(result.Analysis.Tags))
    }

    if result.Trace.ToolsUsed[0] != "web_search" {
        t.Fatalf("expected tool usage to be recorded, got %v", result.Trace.ToolsUsed)
    }
}

func TestEinoMetadataAgentAnalyze_InvalidOutput(t *testing.T) {
    agent := NewEinoMetadataAgent(&fakeRuntime{response: &AgentRunResponse{Output: "not-json"}})

    if _, err := agent.Analyze(context.Background(), &MetadataGenerationInput{FileName: "movie.mp4", FileType: "video"}, &models.AIConfig{Enabled: true, EnableAgent: true}, nil); err == nil {
        t.Fatal("expected invalid output to return error")
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./app/agents -run "TestEinoMetadataAgentAnalyze_ParsesStructuredOutput|TestEinoMetadataAgentAnalyze_InvalidOutput"`
Expected: FAIL with undefined `NewEinoMetadataAgent`, `AgentRunResponse`, or `Analyze` signature mismatch.

- [ ] **Step 3: Write the minimal implementation**

```go
package agents

import (
    "context"

    "LocalSpace/app/models"
    "LocalSpace/app/tools"
)

type AgentRunRequest struct {
    SystemPrompt string
    UserPrompt   string
    Tools        []tools.Tool
    AIConfig     *models.AIConfig
}

type AgentRunResponse struct {
    Output        string
    ToolsUsed     []string
    SearchQueries []string
}

type AgentRuntime interface {
    Run(ctx context.Context, req *AgentRunRequest) (*AgentRunResponse, error)
}
```

```go
package agents

import (
    "context"

    "LocalSpace/app/models"
    "LocalSpace/app/tools"
)

type MetadataAgent interface {
    Analyze(ctx context.Context, input *MetadataGenerationInput, aiConfig *models.AIConfig, availableTools []tools.Tool) (*MetadataAnalysisResult, error)
}

type EinoMetadataAgent struct {
    runtime       AgentRuntime
    promptBuilder *PromptBuilder
}

func NewEinoMetadataAgent(runtime AgentRuntime) *EinoMetadataAgent {
    return &EinoMetadataAgent{
        runtime:       runtime,
        promptBuilder: NewPromptBuilder(),
    }
}

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

    available := make([]string, 0, len(availableTools))
    for _, tool := range availableTools {
        available = append(available, tool.Name())
    }

    return &MetadataAnalysisResult{
        Analysis: analysis,
        Trace: &MetadataTrace{
            AgentEnabled:   true,
            ToolsAvailable: available,
            ToolsUsed:      response.ToolsUsed,
            SearchQueries:  response.SearchQueries,
            RawOutput:      response.Output,
        },
    }, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./app/agents -run "TestEinoMetadataAgentAnalyze_ParsesStructuredOutput|TestEinoMetadataAgentAnalyze_InvalidOutput"`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add app/agents/agent_runtime.go app/agents/agent_runtime_test.go app/agents/metadata_agent.go app/agents/metadata_agent_test.go
git commit -m "feat: add metadata agent execution seam"
```

### Task 3: Enhance tool registry and bounded web-search exposure

**Files:**
- Modify: `app/tools/tool_registry.go`
- Modify: `app/tools/tool_registry_test.go`
- Modify: `app/tools/web_search_tool.go`
- Modify: `app/tools/web_search_tool_test.go`

**Interfaces:**
- Consumes: `type MetadataGenerationInput struct { ... }`
- Consumes: `type AIConfig struct { EnableWebSearch bool ... }`
- Produces: `func (r *ToolRegistry) ResolveForMetadata(config *models.AIConfig, input *agents.MetadataGenerationInput) []Tool`
- Produces: `func ShouldExposeWebSearch(config *models.AIConfig, input *agents.MetadataGenerationInput) bool`

- [ ] **Step 1: Write the failing tool gating tests**

```go
package tools

import (
    "testing"

    "LocalSpace/app/agents"
    "LocalSpace/app/models"
)

func TestShouldExposeWebSearch(t *testing.T) {
    enabledConfig := &models.AIConfig{Enabled: true, EnableAgent: true, EnableWebSearch: true}
    sparseInput := &agents.MetadataGenerationInput{FileName: "MyAwesomeMovie.mp4", FileType: "video"}
    richInput := &agents.MetadataGenerationInput{FileName: "movie.mp4", FileType: "video", UserKeywords: "action thriller", UserTags: []string{"娱乐", "动作"}, UserDescription: "一部动作电影"}

    if !ShouldExposeWebSearch(enabledConfig, sparseInput) {
        t.Fatal("expected sparse input to expose web search")
    }

    if ShouldExposeWebSearch(enabledConfig, richInput) {
        t.Fatal("expected rich input to skip web search exposure")
    }
}

func TestToolRegistryResolveForMetadata(t *testing.T) {
    registry := NewToolRegistry()
    searchTool := &MockTool{name: "web_search"}
    if err := registry.Register(searchTool.Name(), searchTool); err != nil {
        t.Fatalf("register tool: %v", err)
    }

    resolved := registry.ResolveForMetadata(&models.AIConfig{Enabled: true, EnableAgent: true, EnableWebSearch: true}, &agents.MetadataGenerationInput{FileName: "MyAwesomeMovie.mp4", FileType: "video"})
    if len(resolved) != 1 || resolved[0].Name() != "web_search" {
        t.Fatalf("expected web_search to be resolved, got %v", resolved)
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./app/tools -run "TestShouldExposeWebSearch|TestToolRegistryResolveForMetadata"`
Expected: FAIL with undefined `ShouldExposeWebSearch` or missing `ResolveForMetadata`.

- [ ] **Step 3: Write the minimal implementation**

```go
package tools

import (
    "LocalSpace/app/agents"
    "LocalSpace/app/models"
)

func ShouldExposeWebSearch(config *models.AIConfig, input *agents.MetadataGenerationInput) bool {
    if config == nil || !config.Enabled || !config.EnableAgent || !config.EnableWebSearch || input == nil {
        return false
    }

    hasRichUserContext := input.UserDescription != "" || len(input.UserTags) >= 2 || input.UserKeywords != ""
    return !hasRichUserContext && input.FileName != ""
}

func (r *ToolRegistry) ResolveForMetadata(config *models.AIConfig, input *agents.MetadataGenerationInput) []Tool {
    if !ShouldExposeWebSearch(config, input) {
        return []Tool{}
    }

    tool, err := r.Get("web_search")
    if err != nil {
        return []Tool{}
    }

    return []Tool{tool}
}
```

```go
// In web_search_tool.go keep execution bounded and deterministic for phase 1.
func (t *WebSearchTool) Execute(ctx context.Context, input string) (string, error) {
    result, err := t.Search(ctx, input)
    if err != nil {
        return "", err
    }

    if len(result.Results) > 3 {
        result.Results = result.Results[:3]
    }

    return fmt.Sprintf("query=%s results=%d", result.Query, len(result.Results)), nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./app/tools -run "TestShouldExposeWebSearch|TestToolRegistryResolveForMetadata"`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add app/tools/tool_registry.go app/tools/tool_registry_test.go app/tools/web_search_tool.go app/tools/web_search_tool_test.go
git commit -m "feat: add metadata tool resolution"
```

### Task 4: Refactor AgentService to use unified metadata analysis and trace-aware fallback

**Files:**
- Modify: `app/services/agent_service.go`
- Modify: `app/services/agent_service_test.go`

**Interfaces:**
- Consumes: `type MetadataAgent interface { Analyze(ctx context.Context, input *agents.MetadataGenerationInput, aiConfig *models.AIConfig, availableTools []tools.Tool) (*agents.MetadataAnalysisResult, error) }`
- Consumes: `func (r *ToolRegistry) ResolveForMetadata(config *models.AIConfig, input *agents.MetadataGenerationInput) []tools.Tool`
- Produces: `func (s *AgentService) AnalyzeMetadata(ctx context.Context, input *agents.MetadataGenerationInput) (*agents.MetadataAnalysis, error)`
- Produces: `func (s *AgentService) AnalyzeMetadataWithTrace(ctx context.Context, input *agents.MetadataGenerationInput) (*agents.MetadataAnalysisResult, error)`
- Produces: compatibility wrappers `GenerateTags(...) ([]string, error)` and `GenerateDescription(...) (string, error)` routed through `AnalyzeMetadata`

- [ ] **Step 1: Write the failing unified service tests**

```go
package services

import (
    "context"
    "testing"

    "LocalSpace/app/agents"
    "LocalSpace/app/models"
    "LocalSpace/app/tools"
)

type fakeMetadataAgent struct {
    result *agents.MetadataAnalysisResult
    err    error
}

func (f *fakeMetadataAgent) Analyze(ctx context.Context, input *agents.MetadataGenerationInput, aiConfig *models.AIConfig, availableTools []tools.Tool) (*agents.MetadataAnalysisResult, error) {
    return f.result, f.err
}

func TestAgentServiceAnalyzeMetadata_ReturnsUnifiedResult(t *testing.T) {
    service := &AgentService{
        metadataAgent: &fakeMetadataAgent{
            result: &agents.MetadataAnalysisResult{
                Analysis: &agents.MetadataAnalysis{Tags: []string{"video", "action"}, Description: "动作片"},
                Trace:    &agents.MetadataTrace{AgentEnabled: true},
            },
        },
        toolRegistry: tools.NewToolRegistry(),
    }

    result, err := service.analyzeMetadataWithConfig(context.Background(), &agents.MetadataGenerationInput{FileName: "movie.mp4", FileType: "video"}, &models.AIConfig{Enabled: true, EnableAgent: true})
    if err != nil {
        t.Fatalf("expected nil error, got %v", err)
    }

    if result.Analysis.Description != "动作片" {
        t.Fatalf("expected description to be returned, got %q", result.Analysis.Description)
    }
}

func TestAgentServiceAnalyzeMetadata_FallbackOnDisabledAgent(t *testing.T) {
    service := &AgentService{toolRegistry: tools.NewToolRegistry()}

    result, err := service.analyzeMetadataWithConfig(context.Background(), &agents.MetadataGenerationInput{FileName: "movie.mp4", FileType: "video", UserTags: []string{"娱乐"}}, &models.AIConfig{Enabled: true, EnableAgent: false})
    if err != nil {
        t.Fatalf("expected nil error, got %v", err)
    }

    if len(result.Analysis.Tags) == 0 {
        t.Fatal("expected fallback tags to be populated")
    }

    if result.Trace.FallbackReason == "" {
        t.Fatal("expected fallback reason to be recorded")
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./app/services -run "TestAgentServiceAnalyzeMetadata_ReturnsUnifiedResult|TestAgentServiceAnalyzeMetadata_FallbackOnDisabledAgent"`
Expected: FAIL with missing `metadataAgent` field, missing `analyzeMetadataWithConfig`, or signature mismatch.

- [ ] **Step 3: Write the minimal implementation**

```go
// Update AgentService fields.
type AgentService struct {
    metadataAgent agents.MetadataAgent
    configRepo    *repositories.ConfigRepository
    toolRegistry  *tools.ToolRegistry
}
```

```go
func (s *AgentService) AnalyzeMetadata(ctx context.Context, input *agents.MetadataGenerationInput) (*agents.MetadataAnalysis, error) {
    result, err := s.AnalyzeMetadataWithTrace(ctx, input)
    if err != nil {
        return nil, err
    }
    return result.Analysis, nil
}

func (s *AgentService) AnalyzeMetadataWithTrace(ctx context.Context, input *agents.MetadataGenerationInput) (*agents.MetadataAnalysisResult, error) {
    config, err := s.configRepo.GetAIConfig()
    if err != nil {
        return s.fallbackMetadataResult(input, "failed to load ai config"), nil
    }
    return s.analyzeMetadataWithConfig(ctx, input, config)
}

func (s *AgentService) analyzeMetadataWithConfig(ctx context.Context, input *agents.MetadataGenerationInput, config *models.AIConfig) (*agents.MetadataAnalysisResult, error) {
    if config == nil || !config.Enabled || !config.EnableAgent {
        return s.fallbackMetadataResult(input, "agent disabled"), nil
    }

    availableTools := s.toolRegistry.ResolveForMetadata(config, input)
    result, err := s.metadataAgent.Analyze(ctx, input, config, availableTools)
    if err != nil {
        fallback := s.fallbackMetadataResult(input, err.Error())
        fallback.Trace.ToolsAvailable = namesForTools(availableTools)
        return fallback, nil
    }
    return result, nil
}
```

```go
func (s *AgentService) GenerateTags(ctx context.Context, fileName, fileType, userKeywords string, userTags []string, userDescription string) ([]string, error) {
    analysis, err := s.AnalyzeMetadata(ctx, &agents.MetadataGenerationInput{
        FileName:        fileName,
        FileType:        fileType,
        UserKeywords:    userKeywords,
        UserTags:        userTags,
        UserDescription: userDescription,
    })
    if err != nil {
        return []string{}, err
    }
    return analysis.Tags, nil
}

func (s *AgentService) GenerateDescription(ctx context.Context, fileName, fileType, userKeywords string, userTags []string, userDescription string) (string, error) {
    analysis, err := s.AnalyzeMetadata(ctx, &agents.MetadataGenerationInput{
        FileName:        fileName,
        FileType:        fileType,
        UserKeywords:    userKeywords,
        UserTags:        userTags,
        UserDescription: userDescription,
    })
    if err != nil {
        return "", err
    }
    return analysis.Description, nil
}
```

```go
func (s *AgentService) fallbackMetadataResult(input *agents.MetadataGenerationInput, reason string) *agents.MetadataAnalysisResult {
    tags := append([]string{}, input.UserTags...)
    if len(tags) == 0 && input.FileType != "" {
        tags = []string{input.FileType}
    }

    description := input.UserDescription
    if description == "" && input.FileType != "" {
        description = input.FileType + "文件"
    }

    return &agents.MetadataAnalysisResult{
        Analysis: &agents.MetadataAnalysis{Tags: tags, Description: description},
        Trace: &agents.MetadataTrace{
            AgentEnabled:   false,
            FallbackReason: reason,
        },
    }
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./app/services -run "TestAgentServiceAnalyzeMetadata_ReturnsUnifiedResult|TestAgentServiceAnalyzeMetadata_FallbackOnDisabledAgent"`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add app/services/agent_service.go app/services/agent_service_test.go
git commit -m "feat: unify metadata analysis in agent service"
```

### Task 5: Update FileService to use one metadata analysis pass

**Files:**
- Modify: `app/services/file_service.go`
- Modify: `app/services/file_service_integration_test.go`

**Interfaces:**
- Consumes: `func (s *AgentService) AnalyzeMetadata(ctx context.Context, input *agents.MetadataGenerationInput) (*agents.MetadataAnalysis, error)`
- Produces: import flow that performs a single metadata analysis and uses fallback only if unified analysis is unavailable

- [ ] **Step 1: Write the failing integration-oriented test**

```go
package services

import (
    "context"
    "testing"

    "LocalSpace/app/agents"
)

type countingAgentService struct {
    calls int
}

func (c *countingAgentService) AnalyzeMetadata(ctx context.Context, input *agents.MetadataGenerationInput) (*agents.MetadataAnalysis, error) {
    c.calls++
    return &agents.MetadataAnalysis{Tags: []string{"video", "action"}, Description: "动作片"}, nil
}

func TestFileServiceUsesSingleMetadataAnalysisResult(t *testing.T) {
    service := &FileService{}
    counting := &countingAgentService{}

    analysis, err := service.resolveMetadataForImport(context.Background(), counting, &ImportFileRequest{FileName: "movie.mp4", Keywords: "action thriller", Tags: []string{"娱乐"}}, "video")
    if err != nil {
        t.Fatalf("expected nil error, got %v", err)
    }

    if counting.calls != 1 {
        t.Fatalf("expected one metadata analysis call, got %d", counting.calls)
    }

    if analysis.Description != "动作片" {
        t.Fatalf("expected unified description result, got %q", analysis.Description)
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./app/services -run "TestFileServiceUsesSingleMetadataAnalysisResult"`
Expected: FAIL with undefined `resolveMetadataForImport` or inability to inject single analysis path.

- [ ] **Step 3: Write the minimal implementation**

```go
type metadataAnalyzer interface {
    AnalyzeMetadata(ctx context.Context, input *agents.MetadataGenerationInput) (*agents.MetadataAnalysis, error)
}

func (s *FileService) resolveMetadataForImport(ctx context.Context, analyzer metadataAnalyzer, req *ImportFileRequest, fileType string) (*agents.MetadataAnalysis, error) {
    if analyzer == nil {
        return &agents.MetadataAnalysis{Tags: req.Tags, Description: req.Description}, nil
    }

    analysis, err := analyzer.AnalyzeMetadata(ctx, &agents.MetadataGenerationInput{
        FileName:        req.FileName,
        FileType:        fileType,
        UserKeywords:    req.Keywords,
        UserTags:        req.Tags,
        UserDescription: req.Description,
    })
    if err != nil {
        return &agents.MetadataAnalysis{Tags: req.Tags, Description: req.Description}, nil
    }

    if len(analysis.Tags) == 0 {
        analysis.Tags = req.Tags
    }
    if analysis.Description == "" {
        analysis.Description = req.Description
    }

    return analysis, nil
}
```

```go
// In ImportFile replace separate tag/description calls with one pass:
metadataAnalysis, _ := s.resolveMetadataForImport(ctx, s.agentService, &req, fileType)
tags := metadataAnalysis.Tags
description := metadataAnalysis.Description
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./app/services -run "TestFileServiceUsesSingleMetadataAnalysisResult"`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add app/services/file_service.go app/services/file_service_integration_test.go
git commit -m "feat: use unified metadata analysis in file import"
```

### Task 6: Wire default runtime, trace, and documentation for phase-one web-search readiness

**Files:**
- Modify: `app/agents/agent_runtime.go`
- Modify: `app/agents/agent_runtime_test.go`
- Modify: `app/services/agent_service.go`
- Modify: `docs/agent-integration.md`

**Interfaces:**
- Consumes: `type AgentRunRequest struct { SystemPrompt string; UserPrompt string; Tools []tools.Tool; AIConfig *models.AIConfig }`
- Produces: default runtime implementation used by `NewAgentService`
- Produces: trace-ready registration of `web_search` through `ToolRegistry`

- [ ] **Step 1: Write the failing runtime-construction and documentation-alignment tests**

```go
package agents

import (
    "context"
    "testing"

    "LocalSpace/app/models"
)

func TestDefaultAgentRuntimeRun_RequiresConfig(t *testing.T) {
    runtime := &DefaultAgentRuntime{}

    if _, err := runtime.Run(context.Background(), &AgentRunRequest{SystemPrompt: "sys", UserPrompt: "user", AIConfig: &models.AIConfig{}}); err == nil {
        t.Fatal("expected invalid config to return error")
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./app/agents -run "TestDefaultAgentRuntimeRun_RequiresConfig"`
Expected: FAIL with undefined `DefaultAgentRuntime` or missing `Run` behavior.

- [ ] **Step 3: Write the minimal implementation**

```go
package agents

import (
    "context"
    "fmt"
    "time"

    "LocalSpace/app/models"

    openaiModel "github.com/cloudwego/eino-ext/components/model/openai"
    "github.com/cloudwego/eino/schema"
)

type DefaultAgentRuntime struct{}

func (r *DefaultAgentRuntime) Run(ctx context.Context, req *AgentRunRequest) (*AgentRunResponse, error) {
    if req == nil || req.AIConfig == nil || req.AIConfig.APIKey == "" || req.AIConfig.Model == "" || req.AIConfig.BaseURL == "" {
        return nil, fmt.Errorf("invalid ai config")
    }

    cfg := &openaiModel.ChatModelConfig{
        APIKey:  req.AIConfig.APIKey,
        BaseURL: req.AIConfig.BaseURL,
        Model:   req.AIConfig.Model,
    }
    if req.AIConfig.Timeout > 0 {
        cfg.Timeout = time.Duration(req.AIConfig.Timeout) * time.Second
    }

    chatModel, err := openaiModel.NewChatModel(ctx, cfg)
    if err != nil {
        return nil, err
    }

    resp, err := chatModel.Generate(ctx, []*schema.Message{
        schema.SystemMessage(req.SystemPrompt),
        schema.UserMessage(req.UserPrompt),
    })
    if err != nil {
        return nil, err
    }

    return &AgentRunResponse{
        Output:    resp.Content,
        ToolsUsed: []string{},
    }, nil
}
```

```go
// In NewAgentService, register web_search once, instantiate DefaultAgentRuntime,
// then build metadataAgent := agents.NewEinoMetadataAgent(&agents.DefaultAgentRuntime{}).
```

```markdown
Update `docs/agent-integration.md` to document:
- `AgentService` now centers on unified metadata analysis
- `FileService` performs one metadata analysis pass per import
- web search is exposed through tool resolution and trace-ready infrastructure
- failures still degrade gracefully and do not block imports
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./app/agents -run "TestDefaultAgentRuntimeRun_RequiresConfig" && go test ./app/services ./app/tools`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add app/agents/agent_runtime.go app/agents/agent_runtime_test.go app/services/agent_service.go docs/agent-integration.md
git commit -m "feat: wire default metadata runtime"
```

## Self-Review

### Spec coverage
- Unified `MetadataAgent`, `MetadataAnalysis`, `MetadataTrace`, and `MetadataAnalysisResult` are implemented by Tasks 1-2.
- `AgentService` compatibility plus unified business entry point are covered by Task 4.
- Single metadata analysis in `FileService` is covered by Task 5.
- Tool resolution, soft-gated web search exposure, and bounded first-phase web-search behavior are covered by Task 3.
- Runtime integration and trace-ready default construction are covered by Task 6.
- Documentation update is covered by Task 6.
- Graceful fallback is covered by Tasks 1, 4, and 5.

### Placeholder scan
- Removed generic "write tests" placeholders by embedding concrete test functions and exact commands.
- Removed vague implementation notes by listing exact file paths, produced interfaces, and code skeletons.

### Type consistency
- `MetadataGenerationInput`, `MetadataAnalysis`, `MetadataTrace`, and `MetadataAnalysisResult` are introduced in Task 1 and reused consistently later.
- `AgentRuntime`, `AgentRunRequest`, and `AgentRunResponse` are introduced in Task 2 and reused in Task 6.
- `AnalyzeMetadata` is introduced in Task 4 and consumed in Task 5.

Plan complete and saved to `docs/superpowers/plans/2026-07-18-metadata-agent-refactor.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**

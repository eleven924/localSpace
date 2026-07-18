# Metadata Agent Refactor Design

**Date:** 2026-07-18  
**Status:** Ready for user review  
**Scope:** Refactor AI metadata generation from direct chat-model calls to a unified Eino ADK agent + tool architecture

## 1. Background

The current AI metadata generation path in LocalSpace uses `AgentService`, `TagAgent`, and `DescriptionAgent`, but the actual execution model is still based on direct `chatModel.Generate(...)`-style request/response prompting rather than a true Eino ADK agent loop with tool calling.

This creates several problems:

- `TagAgent` and `DescriptionAgent` duplicate model initialization, prompting, parsing, and fallback logic.
- Web search exists as a configuration concept and placeholder tool, but it is not integrated into a real agent tool-calling loop.
- Tags and descriptions are generated as two loosely related tasks even though they are derived from the same file context.
- Debugging relies heavily on scattered logging rather than structured execution trace data.

The desired direction is to redesign this capability around a single metadata-oriented agent that can use tools, produce unified results, and support future expansion such as web search and additional metadata enrichment tools.

## 2. Goals

This design should achieve the following goals:

1. Replace pseudo-agent direct model invocation with a real Eino ADK agent execution path.
2. Unify metadata generation into a single `MetadataAgent` that produces both tags and description in one analysis pass.
3. Separate business result output from internal execution trace/debug information.
4. Preserve current business compatibility through `AgentService` while consolidating internal architecture.
5. Provide a stable extension point for web search and future tools.
6. Ensure agent or tool failures never block file import.

## 3. Non-Goals

This refactor does **not** attempt to:

- Fully optimize web search relevance or quality in the first phase.
- Introduce multi-agent orchestration.
- Replace all AI-related services across the entire codebase in one step.
- Remove every legacy file immediately during the first migration phase.

## 4. High-Level Architecture

### 4.1 Business Entry Point

`AgentService` remains the service-layer facade used by `FileService` and other business code.

Responsibilities:

- Read and validate AI configuration.
- Decide whether agent execution is enabled.
- Resolve which tools are available for the current request.
- Call `MetadataAgent`.
- Apply graceful fallback when agent execution fails or is disabled.
- Expose a clean business-facing result.
- Preserve trace information for internal use.

### 4.2 Unified Intelligent Execution Core

A new `MetadataAgent` becomes the single intelligent execution entry point for metadata enrichment.

Responsibilities:

- Accept a unified metadata-generation input model.
- Build agent prompts using shared prompt infrastructure.
- Execute a real Eino ADK agent loop.
- Use available tools when needed.
- Return structured metadata output plus trace data.

This replaces the current dual-agent split between `TagAgent` and `DescriptionAgent` as the primary execution path.

### 4.3 Agent Runtime Layer

A dedicated `AgentRuntime` encapsulates Eino ADK integration details.

Responsibilities:

- Initialize the model/provider runtime.
- Bind the configured tools.
- Configure timeout, max tokens, temperature, and related runtime controls.
- Execute the agent loop.
- Return raw output and tool-usage information.

This prevents Eino-specific execution details from leaking into business-layer code.

### 4.4 Tool Layer

The tool layer is centered around `ToolRegistry` and concrete tools such as `WebSearchTool`.

Responsibilities:

- Register all candidate tools.
- Resolve the subset of tools exposed to the agent for a given request/configuration.
- Keep tool policy and availability outside the main agent logic.

### 4.5 Prompt and Output Protocol Layer

Prompt and output handling remain explicit and structured.

Responsibilities:

- Build stable system instructions for metadata analysis.
- Build request-specific user prompt content.
- Define the expected structured output format.
- Parse and normalize returned results.

## 5. Core Domain Models

### 5.1 Unified Input Model

A new unified input model replaces the split between tag-generation input and description-generation input.

```go
type MetadataGenerationInput struct {
    FileName        string
    FileType        string
    UserKeywords    string
    UserTags        []string
    UserDescription string
    Metadata        models.Metadata
}
```

This model represents the complete input context needed for metadata enrichment.

### 5.2 Business Result Model

```go
type MetadataAnalysis struct {
    Tags        []string
    Description string
}
```

This is the clean business-facing output returned to main service consumers.

### 5.3 Execution Trace Model

```go
type MetadataTrace struct {
    AgentEnabled   bool
    ToolsAvailable []string
    ToolsUsed      []string
    SearchQueries  []string
    FallbackReason string
    RawOutput      string
}
```

This trace is intended for internal debugging, diagnostics, and testing rather than normal business consumption.

### 5.4 Internal Combined Result

```go
type MetadataAnalysisResult struct {
    Analysis *MetadataAnalysis
    Trace    *MetadataTrace
}
```

This provides one internal return type that can carry both business results and execution details.

## 6. Service and Agent Interfaces

### 6.1 AgentService Public Interfaces

Primary interface:

```go
func (s *AgentService) AnalyzeMetadata(
    ctx context.Context,
    input *agents.MetadataGenerationInput,
) (*agents.MetadataAnalysis, error)
```

Internal/debug interface:

```go
func (s *AgentService) AnalyzeMetadataWithTrace(
    ctx context.Context,
    input *agents.MetadataGenerationInput,
) (*agents.MetadataAnalysisResult, error)
```

Compatibility interfaces retained:

```go
func (s *AgentService) GenerateTags(...)
func (s *AgentService) GenerateDescription(...)
```

These compatibility methods should internally delegate to the unified metadata analysis path rather than invoking separate agent executions.

### 6.2 MetadataAgent Interface

```go
type MetadataAgent interface {
    Analyze(
        ctx context.Context,
        input *MetadataGenerationInput,
        aiConfig *models.AIConfig,
    ) (*MetadataAnalysisResult, error)
}
```

A concrete implementation such as `EinoMetadataAgent` is recommended.

### 6.3 AgentRuntime Interface

```go
type AgentRuntime interface {
    Run(
        ctx context.Context,
        req *AgentRunRequest,
    ) (*AgentRunResponse, error)
}
```

Representative request/response structures:

```go
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
```

## 7. End-to-End Execution Flow

A single metadata analysis request should execute as follows:

1. `FileService` collects file context.
2. `FileService` calls `AgentService.AnalyzeMetadata(...)`.
3. `AgentService` validates AI configuration.
4. `AgentService` decides whether agent execution is enabled.
5. `ToolRegistry` resolves the tools available for this request.
6. `PromptBuilder` builds system and user prompts.
7. `MetadataAgent` invokes `AgentRuntime`.
8. `AgentRuntime` runs a real Eino ADK agent loop.
9. The agent may invoke tools such as web search if exposed.
10. Raw output is returned.
11. `OutputParser` parses and normalizes the result.
12. `AgentService` applies fallback if needed.
13. `FileService` persists tags and description into the file record.

## 8. Tool Exposure Strategy

The design adopts an **outer authorization, inner autonomy** model.

- The outer service/runtime layers decide whether a tool is available to the agent.
- Once a tool is exposed, the agent decides whether and when to use it.

This preserves agent autonomy while keeping runtime cost and capability boundaries under service control.

### 8.1 Web Search Exposure Policy

`EnableWebSearch` should become a real capability gate instead of a passive config field.

However, a soft exposure policy is recommended. The system may choose not to expose web search when available context is already sufficient, for example when:

- The user has already provided a strong description.
- User tags and keywords are sufficiently rich.
- File metadata already makes the subject clear.

This is a resource/permission decision, not a substitute for agent reasoning.

### 8.2 Web Search Operational Constraints

In the first phase, web search should be conservative and bounded:

- Limit calls per metadata request.
- Limit query length.
- Limit result count and payload size.
- Apply explicit timeouts.
- Never fail the whole metadata analysis if search fails.

The first objective is to make tool usage architecturally correct and observable, not maximally sophisticated.

## 9. Prompt and Output Strategy

### 9.1 Prompt Structure

`PromptBuilder` should evolve from separate tag and description prompts to a unified metadata prompt strategy.

Recommended split:

- `BuildSystemPrompt()` for persistent instructions.
- `BuildUserPrompt(input *MetadataGenerationInput)` for request-specific context.

System prompt should define:

- The agent role as a metadata enrichment assistant.
- The requirement to produce tags and description together.
- Preference for preserving user-provided context.
- Permission to use tools when helpful.
- Output structure and constraints.

### 9.2 Structured Output First

The new design should prioritize structured output, ideally JSON-shaped output such as:

```json
{
  "tags": ["动作", "冒险", "电影"],
  "description": "这是一部动作冒险题材的视频文件，适合娱乐观看。"
}
```

`OutputParser` should then:

- Validate the structure.
- Remove duplicates.
- Enforce tag count and length limits.
- Enforce description length constraints.
- Handle malformed output gracefully.

The strategy is **structured output first, rule-based cleanup as fallback**.

## 10. Fallback Policy

Fallback remains a service-level responsibility.

### 10.1 Pre-Agent Fallback

Do not enter the agent path when:

- AI is disabled.
- Agent execution is disabled.
- Required AI config is invalid.
- Runtime prerequisites are missing.

### 10.2 Execution Fallback

Fallback should occur when:

- Runtime initialization fails.
- Agent execution fails.
- Tool execution fails in a way that prevents useful output.
- Timeout occurs.
- Output parsing fails.

### 10.3 Result Fallback

Even after successful execution, fallback/repair logic may apply if results are poor or empty.

### 10.4 Fallback Priority Order

Tags:

1. Agent output
2. User-provided tags
3. Basic rule-generated tags
4. Empty array

Description:

1. Agent output
2. User-provided description
3. Basic rule-generated description
4. Empty string

This preserves user intent and ensures imports are never blocked by AI failure.

## 11. File-Level Refactor Plan

### 11.1 Files to Keep and Refactor

- `app/services/agent_service.go`
- `app/agents/prompt_builder.go`
- `app/tools/web_search_tool.go`

### 11.2 New Files to Add

- `app/agents/metadata_agent.go`
- `app/agents/metadata_types.go`
- `app/agents/agent_runtime.go`
- `app/agents/output_parser.go`

### 11.3 Files to Keep in Transitional or Reduced Role

- `app/agents/agent_base.go`
- `app/services/ai_service.go`

### 11.4 Files to Remove from Primary Execution Path

- `app/agents/tag_agent.go`
- `app/agents/description_agent.go`

They may remain temporarily for compatibility or phased migration, but should no longer represent the main metadata-generation path.

## 12. Migration Plan

### Phase 1: Build New Skeleton

- Add new metadata-agent-related files.
- Introduce unified models and parser.
- Enhance tool registry.
- Keep legacy code present but secondary.

### Phase 2: Refactor AgentService

- Introduce `AnalyzeMetadata(...)` and `AnalyzeMetadataWithTrace(...)`.
- Route compatibility methods through the unified path.

### Phase 3: Update FileService

- Replace separate tag/description generation calls with a single metadata analysis call.
- Ensure only one analysis pass occurs per import.

### Phase 4: Activate Tool Policy

- Connect `EnableWebSearch` to actual tool resolution.
- Make web search available to the agent through the runtime.
- Record tool usage in trace.

### Phase 5: Retire Legacy Primary Path

- Remove `TagAgent` and `DescriptionAgent` from the main execution path.
- Clarify whether `AIService` remains as a lower-level helper or is removed.

## 13. Testing Strategy

### 13.1 Unit Tests

Priority targets:

- `OutputParser`
- Fallback logic
- Tool exposure policy

### 13.2 Service Tests

Use mocks/fakes for the metadata agent to validate:

- Disabled configuration behavior
- Fallback behavior
- Compatibility method behavior
- Trace propagation

### 13.3 Runtime/Integration Tests

Validate:

- Runtime tool binding
- Agent execution path correctness
- Parsing of returned structured output
- Safe handling of tool failures

### 13.4 FileService Integration Tests

Validate:

- A single metadata analysis path is used during import.
- AI failures do not block import.
- Returned metadata is persisted correctly.

## 14. Risks and Controls

### Risk 1: Superficial Refactor

The implementation may rename structures without truly switching to Eino ADK agent execution.

**Control:** Verify that the new main path uses an actual ADK agent loop with tool binding.

### Risk 2: Duplicate Execution

Compatibility wrappers may still cause multiple full agent runs.

**Control:** Move `FileService` to a single unified metadata analysis call as part of the migration.

### Risk 3: Structured Output Instability

Model output may be malformed or partially valid.

**Control:** Use a dedicated parser plus fallback cleanup.

### Risk 4: Search Latency and Noise

Search can increase latency and cost.

**Control:** Bound tool usage, timeout, payload size, and failure impact.

### Risk 5: Poor Observability

Execution may remain difficult to diagnose if trace is not preserved.

**Control:** Keep business results and trace explicitly separated.

## 15. Success Criteria

This refactor is successful when all of the following are true:

1. `AgentService` exposes unified metadata analysis as the primary business interface.
2. `FileService` uses one metadata analysis pass to generate both tags and description.
3. The primary metadata path runs through a real Eino ADK agent execution model.
4. Tools can be registered and exposed by configuration.
5. Web search is architecturally integrated into the agent tool path.
6. Business results are separated from execution trace.
7. Agent/tool failures never block file import.
8. Legacy tag/description direct-execution paths no longer serve as the main architecture.

## 16. Recommendation

Proceed with a unified `MetadataAgent` design rather than preserving separate `TagAgent` and `DescriptionAgent` as first-class execution units.

This approach best matches the actual business problem—file metadata enrichment as a single intelligent task—and creates the cleanest foundation for future tool integrations such as web search and additional metadata assistance capabilities.

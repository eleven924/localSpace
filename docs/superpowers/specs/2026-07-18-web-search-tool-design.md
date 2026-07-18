# Web Search Tool Implementation Design

**Date:** 2026-07-18  
**Status:** Ready for user review  
**Scope:** Implement a real `web_search` tool and connect the metadata agent runtime so the agent can actually call it during metadata analysis.

## 1. Background

The project now has the core architectural seams for agent-facing tool usage:

- `AgentService` registers `web_search` through `ToolRegistry`.
- `AgentService` conditionally exposes `web_search` when AI, agent mode, and web search are enabled for eligible metadata inputs.
- `EinoMetadataAgent` separates `ToolsAvailable` from `ToolsUsed` in trace output.
- `WebSearchTool` is implemented as a bounded HTTP-backed search tool.
- `DefaultAgentRuntime` performs a bounded tool-calling loop for metadata generation.

Implementation note: The first shipped runtime loop uses a small JSON tool-call envelope (`{"tool":"...","input":"..."}`) for deterministic testing before migrating to richer provider-native tool schemas.

As a result, the codebase currently models web search as “available in principle” rather than “usable in practice.” The agent can be told a tool exists, but it cannot actually invoke it and receive real search results.

## 2. Goals

This change should achieve the following:

1. Make `app/tools/web_search_tool.go` perform real HTTP-backed search requests.
2. Keep the search implementation provider-agnostic at the tool boundary by introducing a client abstraction.
3. Allow the metadata agent runtime to expose tools to the model and execute a bounded tool-calling loop.
4. Preserve the existing service-layer gating rule for whether `web_search` is available.
5. Record accurate tool trace information, including which tools were available, which were used, and which search queries were issued.
6. Ensure search failures never block metadata generation or file import.

## 3. Non-Goals

This design does not attempt to:

- Build a general multi-tool orchestration platform.
- Introduce webpage fetching, crawling, or full-document reading.
- Optimize search ranking quality beyond basic provider passthrough.
- Replace the existing tool exposure policy in `AgentService`.
- Redesign the overall metadata prompt system beyond what is needed for correct tool usage.

## 4. Recommended Approach

Use a **tool-owned search client abstraction backed by a configurable HTTP provider**, and extend the metadata runtime from a single direct model call to a **bounded tool-calling loop**.

This is the smallest change that makes the current architecture actually work:

- `AgentService` continues to own registration and policy.
- `WebSearchTool` becomes real instead of placeholder.
- `DefaultAgentRuntime` becomes capable of running one model → tool → model loop when tools are available.
- The runtime remains bounded and metadata-focused rather than expanding into a broad agent framework.

## 5. Architecture

### 5.1 Service Layer Responsibilities Stay Where They Are

`AgentService` should continue to own:

- AI config loading
- tool registration
- tool exposure gating
- fallback behavior
- final result shaping

This preserves the current clean boundary: the runtime does not decide whether a tool is allowed; it only executes tools it has already been given.

### 5.2 Web Search Tool Becomes a Thin Orchestrator

`WebSearchTool` should stop owning the placeholder search logic directly. Instead, it should:

- validate search input
- normalize/bound the query
- apply tool-level limits
- call a `WebSearchClient`
- format a bounded result payload for the agent

This keeps the tool easy to reason about and easy to test.

### 5.3 Search Client Abstraction

Introduce a client interface behind the tool, for example:

```go
type WebSearchClient interface {
    Search(ctx context.Context, query string, limit int) ([]SearchItem, error)
}
```

Responsibilities of the client abstraction:

- isolate provider-specific HTTP details
- normalize provider responses into internal `SearchItem` values
- allow unit tests to inject a fake client
- make future provider replacement cheap

### 5.4 Default HTTP Search Client

The first production implementation should be an HTTP client, for example `HTTPWebSearchClient`.

Responsibilities:

- build the outbound request from configured provider settings
- attach authentication headers or query parameters
- decode JSON responses
- map provider-specific fields into internal `SearchItem`
- respect context cancellation and timeout

The abstraction is provider-oriented, but the first concrete implementation still needs a real expected wire format. That means the system is extensible in structure while still shipping a concrete default behavior.

## 6. Data Model and Tool Contract

### 6.1 Search Result Shape

The internal result shape should remain deliberately small:

```go
type SearchItem struct {
    Title   string
    URL     string
    Snippet string
}
```

This is enough for metadata enrichment while avoiding excessive prompt bloat.

### 6.2 Tool Output Shape

The tool should return a bounded, agent-readable payload containing:

- the normalized query
- the number of results returned
- a short list of search results with title, URL, and snippet

The output can still be a string at the `Tool` interface boundary, but the content should be structured and deterministic enough that the runtime can also record the query used.

Recommended output style:

```text
query: MyAwesomeMovie mp4
results:
1. Title: ...
   URL: ...
   Snippet: ...
```

### 6.3 Bounds

The first implementation should keep hard bounds:

- empty query rejected
- query length capped
- default max results limited to 3–5
- snippet length truncated
- total returned payload bounded

These limits protect both latency and model context budget.

## 7. Configuration Strategy

The search provider should be configured independently from the LLM model settings.

Recommended new `AIConfig` fields:

- `WebSearchProvider string`
- `WebSearchBaseURL string`
- `WebSearchAPIKey string`
- `WebSearchTimeout int`
- `WebSearchMaxResults int`

The existing `EnableWebSearch bool` remains the capability gate.

### 7.1 Why Not Reuse the LLM Config?

The model provider and the search provider are often different systems with different:

- endpoints
- authentication
- timeout expectations
- usage limits

Keeping search config separate prevents hidden coupling and makes the tool actually deployable.

### 7.2 Defaults

The repository layer should supply safe defaults when fields are missing:

- `WebSearchTimeout`: e.g. 10 seconds
- `WebSearchMaxResults`: e.g. 3
- provider/base URL/API key: empty until configured

If search is enabled but required search provider config is missing, the tool should fail clearly and the service should still fall back safely.

## 8. Runtime Tool-Calling Design

### 8.1 Current Gap

Today `DefaultAgentRuntime` does this:

1. build chat model
2. send system + user message once
3. return output

That path cannot support actual tool use.

### 8.2 New Runtime Behavior

When tools are available, the runtime should switch to a bounded loop:

1. Build the model request with tool definitions.
2. Ask the model for the next step.
3. If the model returns final content, stop.
4. If the model requests a tool call:
   - locate the named tool
   - execute it with the provided input
   - append the tool result to the conversation
   - record the tool name in `ToolsUsed`
   - record the search query in `SearchQueries` when `web_search` runs
5. Ask the model again for the final structured metadata output.
6. Stop after a small maximum number of tool rounds.

### 8.3 Loop Boundaries

For metadata generation, the loop should stay intentionally tight:

- at most one or two tool rounds
- no recursive or open-ended planning
- fail closed on malformed tool calls

This keeps the runtime aligned with the project’s current narrow metadata use case.

### 8.4 Tool Definition Mapping

The runtime must map internal `tools.Tool` instances into model-visible tool definitions:

- name from `Tool.Name()`
- description from `Tool.Description()`
- a simple single-string input contract for initial compatibility

The first version does not need a rich JSON arguments schema if the current model/runtime integration works cleanly with a single input string. The important requirement is correctness and observability.

## 9. Trace and Observability

The existing split between service-layer availability and runtime usage should be preserved and made accurate.

### 9.1 `ToolsAvailable`

Still determined in `AgentService` from config and request context.

### 9.2 `ToolsUsed`

Now determined only by actual runtime execution.

### 9.3 `SearchQueries`

Should capture the actual queries passed to `web_search`.

This allows diagnostics to answer three different questions clearly:

1. Was the tool allowed?
2. Did the model choose to use it?
3. What did it search for?

## 10. Error Handling and Fallback

### 10.1 Search Tool Errors

Search should return an error when:

- the query is empty
- the query exceeds limits
- the HTTP provider is unconfigured
- the provider returns non-200
- the provider response is malformed
- the request times out

### 10.2 Runtime Behavior on Tool Failure

If tool execution fails, the runtime should surface the error cleanly to the caller. `AgentService` already owns fallback behavior, so this keeps failure handling centralized.

### 10.3 Service-Level Fallback

`AgentService` should preserve the current guarantees:

- metadata generation failure does not block import
- fallback reasons remain traceable
- user-provided metadata stays preferred where applicable

## 11. File Changes

### 11.1 Files to Modify

- `app/tools/web_search_tool.go`
- `app/tools/web_search_tool_test.go`
- `app/agents/agent_runtime.go`
- `app/agents/agent_runtime_test.go`
- `app/models/config.go`
- `app/models/config_test.go`
- `app/repositories/config_repository.go`
- `app/database/migrations.go`
- `frontend/src/components/AIConfigForm.vue`
- `docs/agent-integration.md`

### 11.2 New Files Likely Needed

- a search-client implementation file under `app/tools/` (for example `web_search_client.go`)
- possibly provider-specific tests under `app/tools/`

## 12. Execution Flow After This Change

The metadata path should work like this:

1. `AgentService` loads AI config.
2. `AgentService` decides whether `web_search` should be exposed.
3. `EinoMetadataAgent` passes the available tools to the runtime.
4. `DefaultAgentRuntime` sends the prompt plus tool definitions to the model.
5. The model may request `web_search`.
6. The runtime executes `WebSearchTool`, which calls the HTTP search client.
7. The runtime appends the tool result to the model conversation.
8. The model returns structured metadata JSON.
9. `EinoMetadataAgent` parses the output.
10. `AgentService` returns the analysis, or falls back if anything failed.

## 13. Testing Strategy

### 13.1 Tool Tests

Add or update tests to cover:

- empty query rejection
- overlong query rejection
- provider config missing
- successful search response mapping
- result-count truncation
- snippet truncation if implemented at tool level
- timeout/cancel propagation

These should primarily use a fake `WebSearchClient` rather than real network access.

### 13.2 Runtime Tests

Add tests for:

- direct no-tool path still working
- tool definitions included when tools are present
- tool call execution updating `ToolsUsed`
- `web_search` query capture updating `SearchQueries`
- runtime stop behavior after bounded tool rounds
- tool execution failure propagating as an error

### 13.3 Service Tests

Keep and extend tests ensuring:

- exposure gating still works
- fallback trace still includes `ToolsAvailable`
- agent-facing result still behaves correctly when the runtime uses tools

### 13.4 Configuration Tests

Add tests for:

- default search config values
- repository round-trip of new search config fields
- migration behavior adding new search config columns

## 14. Risks and Controls

### Risk 1: Provider coupling sneaks back into the tool

**Control:** keep provider-specific JSON parsing inside the client implementation, not inside `WebSearchTool` itself.

### Risk 2: Tool output is too verbose for metadata generation

**Control:** cap result count, snippet length, and total output size.

### Risk 3: Runtime becomes an unbounded agent framework

**Control:** implement only a small bounded loop tuned for metadata generation.

### Risk 4: Search misconfiguration causes hard failures

**Control:** let the tool/runtime fail clearly, but keep service-layer fallback intact.

### Risk 5: Trace becomes misleading again

**Control:** only runtime execution writes `ToolsUsed` and `SearchQueries`; service only writes `ToolsAvailable`.

## 15. Success Criteria

This work is successful when all of the following are true:

1. `web_search` returns real HTTP-backed results instead of a placeholder empty list.
2. The metadata agent runtime can expose and execute `web_search`.
3. The agent can complete a model → tool → model flow during metadata generation.
4. Trace accurately distinguishes available tools from used tools.
5. Search queries are recorded when `web_search` runs.
6. Search or runtime failures do not block import-time metadata fallback behavior.
7. The configuration layer can persist the required search-provider settings.

## 16. Recommendation

Proceed with a **search-client abstraction + default HTTP provider + bounded runtime tool loop**.

This matches the codebase’s current architecture, avoids a large runtime redesign, and turns `web_search` from a placeholder capability into a real agent-usable tool with clear testing and fallback boundaries.

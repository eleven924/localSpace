# Agent Integration Guide

## Overview

LocalSpace now routes import-time AI enrichment through one metadata-agent path. During file import, `FileService` performs a single metadata analysis pass and reuses the result for both tags and description generation. Phase one also prepares bounded web-search readiness, but the current runtime still executes one direct chat-model call per metadata request instead of a richer multi-step agent loop.

## Phase-One Architecture

The delivered phase-one architecture is intentionally narrow and explicit:

- **FileService**: Builds metadata input once during import and reuses one analysis result for tag and description updates.
- **AgentService**: Owns AI configuration lookup, fallback behavior, `ToolRegistry` setup, and metadata tool gating.
- **EinoMetadataAgent**: Builds prompts, delegates a single runtime invocation, parses structured JSON output, and records trace data.
- **DefaultAgentRuntime**: Implements the production runtime seam used by `NewAgentService`; today it wraps one direct Eino/OpenAI-compatible chat-model call.
- **ToolRegistry**: Holds optional tools such as `web_search` so the service layer can expose them deliberately.
- **WebSearchTool**: Exists as a trace-ready, service-gated capability for future expansion.

This is an architecture change, not just a rename of direct model invocation: prompt building, runtime execution, trace capture, and service-layer tool exposure now sit behind explicit metadata-agent seams even though the phase-one runtime remains a single direct model call internally.

## Unified Metadata Analysis Flow

1. `FileService` assembles `MetadataGenerationInput` during import.
2. `AgentService.AnalyzeMetadata` or `AnalyzeMetadataWithTrace` loads AI configuration and decides whether phase-one agent execution is enabled.
3. `AgentService` resolves optional tools from `ToolRegistry` in the service layer. For phase one, this may expose `web_search`.
4. `EinoMetadataAgent` builds prompts and invokes `DefaultAgentRuntime` once.
5. `DefaultAgentRuntime` performs one direct chat-model request and returns output without marking available tools as used.
6. `EinoMetadataAgent` parses the structured response into one `MetadataAnalysisResult`.
7. `FileService` reuses that single analysis result for both tags and description updates, while explicit user tags and descriptions keep priority over generated values.
8. If any AI or runtime step fails, the service falls back to safe metadata defaults without blocking the import.

## Runtime and Tool Boundaries

The important phase-one seams are:

- Tool registration happens once in `NewAgentService`.
- Tool gating stays in `AgentService`, not in `DefaultAgentRuntime`.
- `DefaultAgentRuntime` does not decide whether tools are allowed.
- `DefaultAgentRuntime` does not yet execute a tool-calling loop.
- The runtime only receives already-resolved tools for future extensibility; phase one records them as available, not used.

That makes the current behavior accurate in both code and docs: web search is architecturally visible and trace-ready, while the runtime remains a direct model invocation seam for now.

## Web-Search Readiness

Phase-one web-search support is intentionally bounded:

- `NewAgentService` registers `web_search` once through `ToolRegistry`.
- Tool exposure is resolved in the service layer, not inside the runtime.
- `web_search` is only offered when AI is enabled, agent mode is enabled, web search is enabled, and the file lacks enough user-supplied context.
- Trace data records which tools were available separately from which tools were actually used, making the flow ready for later observability and expansion.

This keeps metadata tool gating in one place while preparing the branch for richer search-backed analysis later.

## Error Resilience

AI generation failures never block file imports:

- Configuration-load failures fall back safely.
- Runtime or parsing failures fall back safely.
- User-provided metadata is preserved where possible; explicit import descriptions and tags take priority over generated values.
- Basic file-type-derived defaults are used when generated metadata is unavailable.
- Trace data retains fallback reasons for diagnostics.

## Configuration

### AI Configuration

Agent-backed metadata analysis depends on AI configuration like the following:

```json
{
  "ai": {
    "enabled": true,
    "apiKey": "your-api-key",
    "model": "gpt-4",
    "baseURL": "https://api.openai.com/v1",
    "enableAgent": true,
    "enableWebSearch": true,
    "maxTokens": 500,
    "timeout": 30
  }
}
```

### Configuration Options

- `enabled`: Enables or disables AI features globally.
- `enableAgent`: Enables unified metadata-agent analysis.
- `enableWebSearch`: Allows service-layer exposure of the `web_search` tool.
- `maxTokens`: Caps model output size.
- `timeout`: Sets the runtime timeout in seconds.

## Operational Notes

- Imports continue even when AI configuration is missing or invalid.
- `AnalyzeMetadataWithTrace` is the diagnostics-friendly entry point when callers need trace information.
- `AnalyzeMetadata`, `GenerateTags`, and `GenerateDescription` all rely on the same unified metadata-analysis path.
- Phase one prepares the architecture for future tool expansion without widening runtime responsibilities.
- The current runtime seam is intentionally direct-chat based; future work can evolve that seam without moving metadata tool gating out of the service layer.

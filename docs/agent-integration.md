# Agent Integration Guide

## Overview

LocalSpace now routes import-time AI enrichment through one metadata-agent path. During file import, `FileService` performs a single metadata analysis pass and reuses the result for both tags and description generation. Phase-one web-search readiness is wired into this flow through trace-aware tool resolution, while failures still degrade gracefully so imports continue.

## Architecture

The phase-one metadata-agent architecture centers on these components:

- **FileService**: Performs one unified metadata analysis pass per import and applies the returned tags and description.
- **AgentService**: Owns metadata analysis orchestration, AI configuration lookup, fallback behavior, and service-layer tool gating.
- **EinoMetadataAgent**: Builds prompts, invokes the runtime, parses structured output, and records execution trace details.
- **DefaultAgentRuntime**: Provides the default production runtime used by `NewAgentService` for model execution.
- **ToolRegistry**: Registers trace-ready tools such as `web_search` for optional metadata-agent use.
- **WebSearchTool**: Supplies bounded phase-one web-search capability when the service layer decides it should be exposed.

## Unified Metadata Analysis Flow

1. `FileService` assembles metadata-generation input during import.
2. `AgentService.AnalyzeMetadata` or `AnalyzeMetadataWithTrace` loads AI configuration and decides whether agent execution is enabled.
3. `AgentService` resolves optional tools from `ToolRegistry` in the service layer. For phase one, this includes gated exposure of `web_search`.
4. `EinoMetadataAgent` runs once through `DefaultAgentRuntime` and returns one structured metadata result.
5. `FileService` reuses that single analysis result for both tags and description updates.
6. If any AI or runtime step fails, the service falls back to safe metadata defaults without blocking the import.

## Web-Search Readiness

Phase-one web-search support is intentionally bounded:

- `NewAgentService` registers `web_search` once through `ToolRegistry`.
- Tool exposure is resolved in the service layer, not inside the runtime.
- `web_search` is only offered when AI is enabled, agent mode is enabled, web search is enabled, and the file lacks enough user-supplied context.
- Trace data records which tools were available and which were used, making the flow ready for later observability and expansion.

This keeps metadata tool gating in one place while preparing the branch for richer search-backed analysis later.

## Error Resilience

AI generation failures never block file imports:

- Configuration-load failures fall back safely.
- Runtime or parsing failures fall back safely.
- User-provided metadata is preserved where possible.
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

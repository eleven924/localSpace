# Agent Integration

## Overview

This project exposes a metadata-focused agent path that can optionally use a bounded `web_search` tool during metadata analysis.

## Runtime Components

- **EinoMetadataAgent**: Builds metadata prompts, invokes the runtime, and maps the JSON result into tag/description output.
- **DefaultAgentRuntime**: Executes a bounded metadata-focused tool loop when tools are available.
- **WebSearchTool**: A real service-gated capability that can execute bounded HTTP-backed search requests during metadata analysis.
- **ToolRegistry**: Stores runtime tools by stable name and allows config-aware registration.
- **AgentService**: Decides whether tools are available from service-layer policy and handles safe fallback behavior.

## Execution Flow

1. `AgentService` loads AI configuration and lazily registers a config-aware `web_search` tool.
2. `AgentService.resolveMetadataTools` decides whether the current file is eligible to expose `web_search`.
3. `EinoMetadataAgent` sends a system/user prompt to `DefaultAgentRuntime`.
4. `DefaultAgentRuntime` may return a direct JSON response, or it may request `web_search` by emitting a small JSON tool envelope.
5. If a tool call is requested, the runtime executes `web_search`, appends tool output, asks the model again, and records `ToolsUsed` / `SearchQueries`.
6. If runtime or tool execution fails, `AgentService` returns a fallback metadata result so file import remains unblocked.

## Trace Semantics

- `ToolsAvailable` is derived by `AgentService` from service gating.
- `ToolsUsed` reflects only tools actually executed by the runtime.
- `SearchQueries` reflects only search inputs actually sent through the runtime.
- `FallbackReason` is populated only when safe fallback behavior was required.

## Constraints

- The initial runtime loop is intentionally bounded.
- Search output is bounded and snippet-heavy, not full-page content.
- Search provider configuration is independent from model configuration.
- Search failures must not block metadata generation or file import.

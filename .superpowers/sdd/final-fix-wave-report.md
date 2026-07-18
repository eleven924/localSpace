Final fix wave report

Files changed
- H:/mySpace/myGoSpace/localSpace/app/agents/agent_runtime.go
- H:/mySpace/myGoSpace/localSpace/app/agents/agent_runtime_test.go
- H:/mySpace/myGoSpace/localSpace/app/agents/metadata_agent.go
- H:/mySpace/myGoSpace/localSpace/app/agents/output_parser.go
- H:/mySpace/myGoSpace/localSpace/app/agents/output_parser_test.go
- H:/mySpace/myGoSpace/localSpace/app/agents/tag_agent_test.go
- H:/mySpace/myGoSpace/localSpace/app/app.go
- H:/mySpace/myGoSpace/localSpace/app/services/file_service.go
- H:/mySpace/myGoSpace/localSpace/app/services/file_service_integration_test.go
- H:/mySpace/myGoSpace/localSpace/docs/agent-integration.md
- H:/mySpace/myGoSpace/localSpace/.superpowers/sdd/final-fix-wave-report.md

Fixes implemented
- Preserved explicit import description priority and explicit user tag priority in FileService metadata resolution.
- Updated App.GetAIAnalysis to use one AgentService.AnalyzeMetadata call instead of separate GenerateTags and GenerateDescription calls.
- Stopped DefaultAgentRuntime from reporting available tools as used when the phase-one runtime does not execute tools.
- Made ParseMetadataOutput accept strict JSON, fenced JSON, and recoverable JSON embedded behind common prefixes/suffixes.
- Passed extracted models.Metadata into agents.MetadataGenerationInput from FileService.
- Updated TestTagAgentGenerate to match the refactored safe-fallback behavior for disabled/unconfigured AI.
- Added nil protection for EinoMetadataAgent runtime, runtime response, prompt builder, and available-tool trace construction.
- Added nil-request protection in FileService.resolveMetadataForImport.
- Updated docs to distinguish available tools from actually used tools and to document explicit user metadata priority.

Exact test commands and outputs

1. Command:
   cd H:/mySpace/myGoSpace/localSpace && go test ./app/services ./app/agents

   Output:
   ok  	LocalSpace/app/services	(cached)
   ok  	LocalSpace/app/agents	(cached)

2. Command:
   cd H:/mySpace/myGoSpace/localSpace && go test ./app

   Output:
   ok  	LocalSpace/app	(cached)

3. Final combined sanity command:
   cd H:/mySpace/myGoSpace/localSpace && go test ./app/services ./app/agents && go test ./app

   Output:
   ok  	LocalSpace/app/services	(cached)
   ok  	LocalSpace/app/agents	(cached)
   ok  	LocalSpace/app	(cached)

Self-review notes
- Verified import safety is preserved: resolveMetadataForImport returns user-provided fallback metadata when the analyzer is nil, the request is nil, analysis fails, or analysis is nil.
- Verified app/tools remains decoupled from app/agents; final fixes stay in app, app/services, app/agents, and docs.
- Verified phase-one runtime behavior is accurately traced: ToolsAvailable is populated by EinoMetadataAgent from resolved service-layer tools, while ToolsUsed remains runtime-reported and empty for the direct chat runtime.
- Verified explicit user tags and descriptions override generated metadata during import rather than being overwritten.
- Verified fenced and embedded metadata JSON parsing is covered by tests.
- Verified extracted file metadata is passed through to MetadataGenerationInput and covered by tests.

Concerns
- No open test concerns. go test ./app/services ./app/agents and go test ./app are green.

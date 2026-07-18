Task 3 implementation report

Files changed
- H:/mySpace/myGoSpace/localSpace/app/services/agent_service.go
- H:/mySpace/myGoSpace/localSpace/app/services/agent_service_test.go
- H:/mySpace/myGoSpace/localSpace/app/tools/tool_registry.go
- H:/mySpace/myGoSpace/localSpace/app/tools/tool_registry_test.go
- H:/mySpace/myGoSpace/localSpace/app/tools/web_search_tool.go
- H:/mySpace/myGoSpace/localSpace/app/tools/web_search_tool_test.go

Summary
- Removed metadata-specific web-search policy from H:/mySpace/myGoSpace/localSpace/app/tools/tool_registry.go so app/tools stays generic and no longer imports app/agents or app/models.
- Added service-level helpers shouldExposeWebSearch and resolveMetadataTools in H:/mySpace/myGoSpace/localSpace/app/services/agent_service.go.
- Moved metadata tool-gating tests to H:/mySpace/myGoSpace/localSpace/app/services/agent_service_test.go and expanded them to cover disabled web search and blank filename cases.
- Kept bounded generic web-search behavior in H:/mySpace/myGoSpace/localSpace/app/tools/web_search_tool.go by returning a deterministic capped summary.
- Adjusted H:/mySpace/myGoSpace/localSpace/app/services/agent_service.go to avoid passing []tools.Tool through AgentConfig for the legacy TagAgent and DescriptionAgent path, keeping the current seam compiling cleanly without metadata runtime wiring.
- Reduced H:/mySpace/myGoSpace/localSpace/app/tools/tool_registry_test.go back to generic registry coverage only.

TDD notes
1. Wrote service-level tests first in H:/mySpace/myGoSpace/localSpace/app/services/agent_service_test.go for shouldExposeWebSearch and resolveMetadataTools.
2. Ran the targeted service tests before implementation and observed the expected failure:
   Command:
   go test ./app/services -run "TestShouldExposeWebSearch|TestResolveMetadataTools"
   Output:
   # LocalSpace/app/services [LocalSpace/app/services.test]
   app\services\agent_service_test.go:39:6: undefined: shouldExposeWebSearch
   app\services\agent_service_test.go:43:5: undefined: shouldExposeWebSearch
   app\services\agent_service_test.go:55:22: service.resolveMetadataTools undefined (type *AgentService has no field or method resolveMetadataTools)
   FAIL	LocalSpace/app/services [build failed]
   FAIL
3. Implemented the minimal helpers and decoupling changes needed to satisfy the tests.
4. Re-ran the targeted tests to green.

Tests run
1. Command:
   go test ./app/services -run "TestShouldExposeWebSearch|TestResolveMetadataTools|TestNewAgentService"
   Output:
   ok  	LocalSpace/app/services	0.226s

2. Command:
   go test ./app/tools -run "TestToolRegistry|TestWebSearchToolExecuteBoundsResults"
   Output:
   ok  	LocalSpace/app/tools	0.462s

Self-review notes
- Confirmed the app/tools package is generic again and no longer imports metadata-agent types, which removes the import-cycle source.
- Confirmed metadata web-search gating now lives only at the service layer.
- Confirmed no AnalyzeMetadata or metadata-agent runtime wiring was added in this task.
- Confirmed bounded web-search behavior remains generic and deterministic.
- Confirmed NewAgentService still registers the web_search tool in its registry while legacy agents no longer depend on passing []tools.Tool through AgentConfig in this task.

Concerns
- resolveMetadataTools is intentionally not yet consumed by runtime execution paths; that integration is deferred to later tasks per scope correction.
- Existing unrelated working tree changes were left untouched.

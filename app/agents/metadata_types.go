package agents

import "LocalSpace/app/models"

// MetadataGenerationInput represents input for metadata generation.
type MetadataGenerationInput struct {
	FileName        string
	FileType        string
	UserKeywords    string
	UserTags        []string
	UserDescription string
	Metadata        models.Metadata
}

// MetadataAnalysis represents structured metadata output.
type MetadataAnalysis struct {
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
}

// MetadataTrace captures metadata generation diagnostics.
type MetadataTrace struct {
	AgentEnabled   bool
	ToolsAvailable []string
	ToolsUsed      []string
	SearchQueries  []string
	FallbackReason string
	RawOutput      string
}

// MetadataAnalysisResult wraps analysis output with trace data.
type MetadataAnalysisResult struct {
	Analysis *MetadataAnalysis
	Trace    *MetadataTrace
}

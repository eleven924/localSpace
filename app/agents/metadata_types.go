package agents

import (
	"LocalSpace/app/models"
	"LocalSpace/app/tools"
)

// AnalysisRequest represents the unified backend input for one AI analysis run.
//
// FileID/FilePath/CollectionID are kept for backend orchestration and must not be
// copied into the model prompt unless a dedicated, sanitized projection allows it.
type AnalysisRequest struct {
	FileID          uint
	FilePath        string
	FileName        string
	FileType        string
	FileSubType     string
	UserKeywords    string
	UserTags        []string
	UserDescription string
	CollectionID    *uint
	NativeMetadata  models.Metadata
	Evidence        []tools.EvidenceItem
	// Metadata keeps source compatibility with the original agent input contract.
	// New callers should prefer NativeMetadata.
	Metadata models.Metadata
}

// MetadataGenerationInput is kept as an alias while callers migrate to AnalysisRequest.
type MetadataGenerationInput = AnalysisRequest

// MetadataAnalysis represents structured metadata output.
type MetadataAnalysis struct {
	Tags                  []string `json:"tags"`
	Description           string   `json:"description"`
	RelatedTags           []string `json:"relatedTags,omitempty"`
	SuggestedCollection   string   `json:"suggestedCollection,omitempty"`
	RecommendationReasons []string `json:"recommendationReasons,omitempty"`
}

// MetadataTrace captures metadata generation diagnostics.
type MetadataTrace struct {
	RunID            string
	Status           string
	Model            string
	AgentEnabled     bool
	Confidence       float64
	NeedsReview      bool
	QualityStatus    string
	CurrentStage     string
	StageHistory     []string
	Errors           []string
	ToolsAvailable   []string
	ToolsUsed        []string
	SearchQueries    []string
	EvidenceSources  []string
	FallbackReason   string
	RawOutput        string
	OutputDiagnostic string
	ToolCalls        []models.AIToolCall `json:"toolCalls,omitempty"`
}

// MetadataAnalysisResult wraps analysis output with trace data.
type MetadataAnalysisResult struct {
	Analysis *MetadataAnalysis
	Trace    *MetadataTrace
}

// NativeFileMetadata returns the preferred native metadata while supporting old callers.
func (r *AnalysisRequest) NativeFileMetadata() models.Metadata {
	if r == nil {
		return models.Metadata{}
	}
	if r.NativeMetadata != (models.Metadata{}) {
		return r.NativeMetadata
	}
	return r.Metadata
}

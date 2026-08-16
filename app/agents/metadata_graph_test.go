package agents

import (
	"context"
	"strings"
	"testing"

	"LocalSpace/app/models"
	"LocalSpace/app/tools"
)

type graphMetadataAgent struct {
	result *MetadataAnalysisResult
	calls  int
}

func (a *graphMetadataAgent) Analyze(_ context.Context, _ *MetadataGenerationInput, _ *models.AIConfig, _ []tools.Tool) (*MetadataAnalysisResult, error) {
	a.calls++
	return a.result, nil
}

func TestMetadataAnalysisGraphRoutesAndRecordsStages(t *testing.T) {
	tests := []struct {
		name     string
		fileType string
		fileName string
		route    string
	}{
		{name: "image", fileType: "image", fileName: "cover.png", route: metadataRouteImage},
		{name: "document", fileType: "document", fileName: "notes.md", route: metadataRouteDocument},
		{name: "media", fileType: "video", fileName: "movie.mp4", route: metadataRouteMedia},
		{name: "archive", fileType: "archive", fileName: "backup.zip", route: metadataRouteArchive},
		{name: "generic", fileType: "binary", fileName: "payload.bin", route: metadataRouteGeneric},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			agent := &graphMetadataAgent{result: &MetadataAnalysisResult{
				Analysis: &MetadataAnalysis{Tags: []string{"Video", " video ", "Document"}, Description: "  generated metadata  "},
			}}
			graph := NewMetadataAnalysisGraph(agent)
			result, err := graph.Run(context.Background(), &AnalysisRequest{
				FileName: test.fileName,
				FileType: test.fileType,
			}, &models.AIConfig{Enabled: true, EnableAgent: true}, nil)
			if err != nil {
				t.Fatalf("run graph: %v", err)
			}
			if agent.calls != 1 {
				t.Fatalf("expected agent to run once, got %d", agent.calls)
			}
			if result == nil || result.Trace == nil {
				t.Fatal("expected graph trace")
			}
			if result.Trace.CurrentStage != metadataRouteReview {
				t.Fatalf("expected final quality stage %q, got %q", metadataRouteReview, result.Trace.CurrentStage)
			}
			if !containsStage(result.Trace.StageHistory, test.route) {
				t.Fatalf("expected route %q in stage history %v", test.route, result.Trace.StageHistory)
			}
			if !containsStage(result.Trace.StageHistory, metadataGraphMerge) || !containsStage(result.Trace.StageHistory, metadataGraphAnalyze) {
				t.Fatalf("expected merge and agent stages in history %v", result.Trace.StageHistory)
			}
			if !result.Trace.NeedsReview || result.Trace.QualityStatus != "needs_review" {
				t.Fatalf("expected review quality status, got needsReview=%v status=%q", result.Trace.NeedsReview, result.Trace.QualityStatus)
			}
			if len(result.Analysis.Tags) != 2 || result.Analysis.Tags[0] != "video" || result.Analysis.Tags[1] != "document" {
				t.Fatalf("expected normalized unique tags, got %v", result.Analysis.Tags)
			}
			if result.Analysis.Description != "generated metadata" {
				t.Fatalf("expected trimmed description, got %q", result.Analysis.Description)
			}
		})
	}
}

func TestMetadataAnalysisGraphAutoApplyWhenEvidenceIsSufficient(t *testing.T) {
	agent := &graphMetadataAgent{result: &MetadataAnalysisResult{
		Analysis: &MetadataAnalysis{Tags: []string{"invoice"}, Description: "an invoice"},
	}}
	graph := NewMetadataAnalysisGraph(agent)
	result, err := graph.Run(context.Background(), &AnalysisRequest{
		FileName: "invoice.txt",
		FileType: "document",
		Evidence: []tools.EvidenceItem{
			{Kind: "file_metadata", Source: "local:file_metadata", Content: "{}"},
			{Kind: "text", Source: "local:document_text", Content: "invoice"},
		},
	}, &models.AIConfig{}, nil)
	if err != nil {
		t.Fatalf("run graph: %v", err)
	}
	if result.Trace.NeedsReview || result.Trace.QualityStatus != "auto_apply" {
		t.Fatalf("expected auto-apply quality status, got needsReview=%v status=%q", result.Trace.NeedsReview, result.Trace.QualityStatus)
	}
	if result.Trace.Confidence < metadataAutoApplyThreshold {
		t.Fatalf("expected confidence >= %.2f, got %.2f", metadataAutoApplyThreshold, result.Trace.Confidence)
	}
}

func TestMetadataAnalysisGraphBuildsReusableTagHints(t *testing.T) {
	agent := &graphMetadataAgent{result: &MetadataAnalysisResult{
		Analysis: &MetadataAnalysis{Tags: []string{"go"}, Description: "development notes"},
	}}
	graph := NewMetadataAnalysisGraph(agent)
	result, err := graph.Run(context.Background(), &AnalysisRequest{
		FileName:     "notes.md",
		FileType:     "document",
		UserTags:     []string{"Project", "go"},
		UserKeywords: "backend, archive",
	}, &models.AIConfig{}, nil)
	if err != nil {
		t.Fatalf("run graph: %v", err)
	}
	if len(result.Analysis.RelatedTags) != 3 {
		t.Fatalf("expected reusable tag hints, got %v", result.Analysis.RelatedTags)
	}
	if result.Analysis.RelatedTags[0] != "project" || result.Analysis.RelatedTags[1] != "backend" || result.Analysis.RelatedTags[2] != "archive" {
		t.Fatalf("unexpected reusable tag hints %v", result.Analysis.RelatedTags)
	}
	if !containsStage(result.Trace.StageHistory, metadataGraphRecommend) {
		t.Fatalf("expected recommendation stage, got %v", result.Trace.StageHistory)
	}
}

func TestMetadataAnalysisGraphInfersRouteFromFilename(t *testing.T) {
	agent := &graphMetadataAgent{result: &MetadataAnalysisResult{
		Analysis: &MetadataAnalysis{Tags: []string{"image"}, Description: "image"},
	}}
	graph := NewMetadataAnalysisGraph(agent)
	result, err := graph.Run(context.Background(), &AnalysisRequest{FileName: "wallpaper.jpg"}, &models.AIConfig{}, nil)
	if err != nil {
		t.Fatalf("run graph: %v", err)
	}
	if !containsStage(result.Trace.StageHistory, metadataRouteImage) {
		t.Fatalf("expected image route, got %v", result.Trace.StageHistory)
	}
}

func TestMetadataAnalysisGraphRejectsEmptyResult(t *testing.T) {
	graph := NewMetadataAnalysisGraph(&graphMetadataAgent{result: &MetadataAnalysisResult{
		Analysis: &MetadataAnalysis{},
	}})
	_, err := graph.Run(context.Background(), &AnalysisRequest{FileName: "unknown.bin"}, &models.AIConfig{}, nil)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !containsText(err.Error(), metadataGraphValidate) {
		t.Fatalf("expected validation stage in error, got %v", err)
	}
}

func containsStage(stages []string, expected string) bool {
	for _, stage := range stages {
		if stage == expected {
			return true
		}
	}
	return false
}

func containsText(value, expected string) bool {
	return strings.Contains(value, expected)
}

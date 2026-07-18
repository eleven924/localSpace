package agents

import "testing"

func TestParseMetadataOutput_JSON(t *testing.T) {
	raw := `{"tags":["video","action","action","thriller","movie","extra"],"description":"A very long description that should still be normalized by the parser layer into a bounded description value."}`

	analysis, err := ParseMetadataOutput(raw)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(analysis.Tags) != 5 {
		t.Fatalf("expected 5 tags after normalization, got %d", len(analysis.Tags))
	}

	if analysis.Tags[0] != "video" {
		t.Fatalf("expected first tag to remain 'video', got %q", analysis.Tags[0])
	}

	if analysis.Description == "" {
		t.Fatal("expected normalized description to be non-empty")
	}
}

func TestParseMetadataOutput_FencedJSON(t *testing.T) {
	raw := "```json\n{\"tags\":[\"video\"],\"description\":\"影片\"}\n```"

	analysis, err := ParseMetadataOutput(raw)
	if err != nil {
		t.Fatalf("expected fenced JSON to parse, got %v", err)
	}
	if analysis.Description != "影片" {
		t.Fatalf("expected parsed description, got %q", analysis.Description)
	}
}

func TestParseMetadataOutput_PrefixedAndSuffixedJSON(t *testing.T) {
	raw := "Here is the metadata:\n{\"tags\":[\"document\"],\"description\":\"报告\"}\nHope this helps."

	analysis, err := ParseMetadataOutput(raw)
	if err != nil {
		t.Fatalf("expected embedded JSON to parse, got %v", err)
	}
	if len(analysis.Tags) != 1 || analysis.Tags[0] != "document" {
		t.Fatalf("expected parsed tags, got %v", analysis.Tags)
	}
}

func TestParseMetadataOutput_InvalidJSON(t *testing.T) {
	if _, err := ParseMetadataOutput("not-json"); err == nil {
		t.Fatal("expected invalid JSON to return error")
	}
}

func TestNormalizeMetadataAnalysis_NilInput(t *testing.T) {
	normalized := NormalizeMetadataAnalysis(nil)
	if normalized == nil {
		t.Fatal("expected non-nil normalized analysis")
	}
	if len(normalized.Tags) != 0 {
		t.Fatalf("expected no tags, got %d", len(normalized.Tags))
	}
}

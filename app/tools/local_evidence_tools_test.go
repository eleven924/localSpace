package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectLocalEvidenceExtractsBoundedText(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "notes.md")
	if err := os.WriteFile(filePath, []byte("# LocalSpace\nThis is local evidence."), 0600); err != nil {
		t.Fatalf("failed to write test document: %v", err)
	}

	evidence, err := CollectLocalEvidence(context.Background(), filePath, "document", "md")
	if err != nil {
		t.Fatalf("CollectLocalEvidence() returned error: %v", err)
	}
	if len(evidence) != 2 {
		t.Fatalf("expected metadata and text evidence, got %d items", len(evidence))
	}
	if !strings.Contains(evidence[1].Content, "LocalSpace") {
		t.Fatalf("expected document text evidence, got %q", evidence[1].Content)
	}
}

func TestLocalEvidenceToolDoesNotAcceptAFilePathFromModel(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(filePath, []byte("safe local content"), 0600); err != nil {
		t.Fatalf("failed to write test document: %v", err)
	}

	toolList := NewLocalEvidenceTools(filePath, "document", "txt")
	if len(toolList) != 3 {
		t.Fatalf("expected three local evidence tools, got %d", len(toolList))
	}
	output, err := toolList[1].Execute(context.Background(), `{"max_chars":100}`)
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}
	if !strings.Contains(output, "safe local content") {
		t.Fatalf("expected bound file content, got %q", output)
	}
}

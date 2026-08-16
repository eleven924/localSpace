package tools

import (
	"archive/zip"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	defaultEvidenceMaxChars = 12000
	maxEvidenceChars        = 20000
	maxArchiveEntries       = 100
)

// EvidenceItem is bounded, model-safe evidence extracted from a local file.
type EvidenceItem struct {
	Kind      string `json:"kind"`
	Source    string `json:"source"`
	Content   string `json:"content"`
	Truncated bool   `json:"truncated,omitempty"`
}

// CollectLocalEvidence performs deterministic local extraction before the Agent decides on extra tools.
func CollectLocalEvidence(ctx context.Context, filePath, fileType, fileSubType string) ([]EvidenceItem, error) {
	if strings.TrimSpace(filePath) == "" {
		return nil, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat local evidence file: %w", err)
	}

	evidence := []EvidenceItem{{
		Kind:   "file_metadata",
		Source: "local:file_metadata",
		Content: mustMarshalJSON(map[string]any{
			"file_type":    fileType,
			"file_subtype": fileSubType,
			"size":         fileInfo.Size(),
			"extension":    strings.ToLower(filepath.Ext(filePath)),
		}),
	}}

	ext := strings.ToLower(filepath.Ext(filePath))
	switch {
	case isTextEvidenceExtension(ext):
		item, extractErr := extractTextEvidence(filePath, defaultEvidenceMaxChars)
		if extractErr != nil {
			return evidence, extractErr
		}
		evidence = append(evidence, item)
	case ext == ".docx":
		item, extractErr := extractDocxTextEvidence(filePath, defaultEvidenceMaxChars)
		if extractErr != nil {
			return evidence, extractErr
		}
		evidence = append(evidence, item)
	case isZipEvidenceExtension(ext):
		item, extractErr := extractZipEntriesEvidence(filePath)
		if extractErr != nil {
			return evidence, extractErr
		}
		evidence = append(evidence, item)
	}

	return evidence, nil
}

// NewLocalEvidenceTools creates tools bound to one backend-validated file path.
// The model only supplies extraction options; it never supplies a filesystem path.
func NewLocalEvidenceTools(filePath, fileType, fileSubType string) []Tool {
	if strings.TrimSpace(filePath) == "" {
		return nil
	}
	return []Tool{
		&localEvidenceTool{name: "get_local_file_metadata", description: "Read bounded metadata for the current local file", filePath: filePath, fileType: fileType, fileSubType: fileSubType},
		&localEvidenceTool{name: "extract_document_text", description: "Extract a bounded text excerpt from the current local document", filePath: filePath, fileType: fileType, fileSubType: fileSubType},
		&localEvidenceTool{name: "list_archive_entries", description: "List bounded entries in the current local ZIP archive", filePath: filePath, fileType: fileType, fileSubType: fileSubType},
	}
}

type localEvidenceTool struct {
	name        string
	description string
	filePath    string
	fileType    string
	fileSubType string
}

func (t *localEvidenceTool) Name() string        { return t.name }
func (t *localEvidenceTool) Description() string { return t.description }

func (t *localEvidenceTool) Execute(ctx context.Context, input string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	switch t.name {
	case "get_local_file_metadata":
		return t.executeMetadata()
	case "extract_document_text":
		limit := parseEvidenceLimit(input)
		item, err := extractDocumentEvidence(t.filePath, limit)
		if err != nil {
			return "", err
		}
		return mustMarshalJSON(item), nil
	case "list_archive_entries":
		item, err := extractZipEntriesEvidence(t.filePath)
		if err != nil {
			return "", err
		}
		return mustMarshalJSON(item), nil
	default:
		return "", fmt.Errorf("unsupported local evidence tool: %s", t.name)
	}
}

func (t *localEvidenceTool) executeMetadata() (string, error) {
	info, err := os.Stat(t.filePath)
	if err != nil {
		return "", fmt.Errorf("failed to stat local file: %w", err)
	}
	return mustMarshalJSON(map[string]any{
		"file_type":    t.fileType,
		"file_subtype": t.fileSubType,
		"size":         info.Size(),
		"extension":    strings.ToLower(filepath.Ext(t.filePath)),
	}), nil
}

func extractDocumentEvidence(filePath string, maxChars int) (EvidenceItem, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".docx" {
		return extractDocxTextEvidence(filePath, maxChars)
	}
	if !isTextEvidenceExtension(ext) {
		return EvidenceItem{}, fmt.Errorf("document text extraction is unsupported for %s", ext)
	}
	return extractTextEvidence(filePath, maxChars)
}

func extractTextEvidence(filePath string, maxChars int) (EvidenceItem, error) {
	maxChars = normalizeEvidenceLimit(maxChars)
	file, err := os.Open(filePath)
	if err != nil {
		return EvidenceItem{}, fmt.Errorf("failed to open document: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, int64(maxChars*4+1)))
	if err != nil {
		return EvidenceItem{}, fmt.Errorf("failed to read document: %w", err)
	}
	content := strings.ToValidUTF8(string(data), "�")
	content, truncated := truncateEvidence(content, maxChars)
	return EvidenceItem{Kind: "document_text", Source: "local:document_text", Content: content, Truncated: truncated}, nil
}

func extractDocxTextEvidence(filePath string, maxChars int) (EvidenceItem, error) {
	maxChars = normalizeEvidenceLimit(maxChars)
	archive, err := zip.OpenReader(filePath)
	if err != nil {
		return EvidenceItem{}, fmt.Errorf("failed to open docx archive: %w", err)
	}
	defer archive.Close()

	for _, entry := range archive.File {
		if entry.Name != "word/document.xml" {
			continue
		}
		reader, err := entry.Open()
		if err != nil {
			return EvidenceItem{}, fmt.Errorf("failed to open docx content: %w", err)
		}
		content, readErr := readXMLText(reader, maxChars)
		_ = reader.Close()
		if readErr != nil {
			return EvidenceItem{}, readErr
		}
		content, truncated := truncateEvidence(content, maxChars)
		return EvidenceItem{Kind: "document_text", Source: "local:docx_text", Content: content, Truncated: truncated}, nil
	}
	return EvidenceItem{}, fmt.Errorf("docx document.xml is missing")
}

func readXMLText(reader io.Reader, maxChars int) (string, error) {
	decoder := xml.NewDecoder(io.LimitReader(reader, 2*1024*1024))
	var builder strings.Builder
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to parse document XML: %w", err)
		}
		if charData, ok := token.(xml.CharData); ok {
			text := strings.TrimSpace(string(charData))
			if text != "" {
				if builder.Len() > 0 {
					builder.WriteByte(' ')
				}
				builder.WriteString(text)
				if utf8.RuneCountInString(builder.String()) >= maxChars {
					break
				}
			}
		}
	}
	return builder.String(), nil
}

func extractZipEntriesEvidence(filePath string) (EvidenceItem, error) {
	archive, err := zip.OpenReader(filePath)
	if err != nil {
		return EvidenceItem{}, fmt.Errorf("failed to open ZIP archive: %w", err)
	}
	defer archive.Close()

	entries := make([]string, 0, minInt(len(archive.File), maxArchiveEntries))
	for index, entry := range archive.File {
		if index >= maxArchiveEntries {
			break
		}
		name := entry.Name
		if len(name) > 240 {
			name = name[:240] + "..."
		}
		entries = append(entries, name)
	}
	return EvidenceItem{Kind: "archive_entries", Source: "local:archive_entries", Content: mustMarshalJSON(map[string]any{
		"count":   len(archive.File),
		"entries": entries,
	})}, nil
}

func isTextEvidenceExtension(ext string) bool {
	switch ext {
	case ".txt", ".md", ".markdown", ".csv", ".json", ".yaml", ".yml", ".xml", ".html", ".htm", ".log", ".ini", ".conf", ".rtf":
		return true
	default:
		return false
	}
}

func isZipEvidenceExtension(ext string) bool {
	switch ext {
	case ".zip", ".docx", ".xlsx", ".pptx":
		return true
	default:
		return false
	}
}

func parseEvidenceLimit(input string) int {
	var payload struct {
		MaxChars int `json:"max_chars"`
		Limit    int `json:"limit"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(input)), &payload) == nil {
		if payload.MaxChars > 0 {
			return normalizeEvidenceLimit(payload.MaxChars)
		}
		if payload.Limit > 0 {
			return normalizeEvidenceLimit(payload.Limit)
		}
	}
	if value, err := strconv.Atoi(strings.TrimSpace(input)); err == nil && value > 0 {
		return normalizeEvidenceLimit(value)
	}
	return defaultEvidenceMaxChars
}

func normalizeEvidenceLimit(limit int) int {
	if limit <= 0 {
		return defaultEvidenceMaxChars
	}
	if limit > maxEvidenceChars {
		return maxEvidenceChars
	}
	return limit
}

func truncateEvidence(content string, maxChars int) (string, bool) {
	maxChars = normalizeEvidenceLimit(maxChars)
	if utf8.RuneCountInString(content) <= maxChars {
		return content, false
	}
	runes := []rune(content)
	return string(runes[:maxChars]), true
}

func mustMarshalJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

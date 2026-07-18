package agents

import (
	"encoding/json"
	"errors"
	"strings"
)

// ParseMetadataOutput parses structured metadata output from the model.
func ParseMetadataOutput(raw string) (*MetadataAnalysis, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("empty metadata output")
	}

	var analysis MetadataAnalysis
	if err := json.Unmarshal([]byte(raw), &analysis); err != nil {
		return nil, err
	}

	return NormalizeMetadataAnalysis(&analysis), nil
}

// NormalizeMetadataAnalysis enforces the metadata output shape.
func NormalizeMetadataAnalysis(analysis *MetadataAnalysis) *MetadataAnalysis {
	if analysis == nil {
		return &MetadataAnalysis{Tags: []string{}, Description: ""}
	}

	normalized := &MetadataAnalysis{
		Tags:        make([]string, 0, len(analysis.Tags)),
		Description: strings.TrimSpace(analysis.Description),
	}

	seen := make(map[string]struct{}, len(analysis.Tags))
	for _, tag := range analysis.Tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		normalized.Tags = append(normalized.Tags, tag)
		if len(normalized.Tags) == 5 {
			break
		}
	}

	if len([]rune(normalized.Description)) > 200 {
		normalized.Description = string([]rune(normalized.Description)[:200]) + "..."
	}

	return normalized
}

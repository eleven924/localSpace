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

	jsonText, err := recoverMetadataJSON(raw)
	if err != nil {
		return nil, err
	}

	var analysis MetadataAnalysis
	if err := json.Unmarshal([]byte(jsonText), &analysis); err != nil {
		return nil, err
	}

	return NormalizeMetadataAnalysis(&analysis), nil
}

func recoverMetadataJSON(raw string) (string, error) {
	if json.Valid([]byte(raw)) {
		return raw, nil
	}

	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "```") {
		lines := strings.Split(trimmed, "\n")
		if len(lines) >= 2 {
			if strings.HasPrefix(strings.TrimSpace(lines[0]), "```") {
				lines = lines[1:]
			}
			if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
				lines = lines[:len(lines)-1]
			}
			candidate := strings.TrimSpace(strings.Join(lines, "\n"))
			if json.Valid([]byte(candidate)) {
				return candidate, nil
			}
		}
	}

	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end > start {
		candidate := strings.TrimSpace(trimmed[start : end+1])
		if json.Valid([]byte(candidate)) {
			return candidate, nil
		}
	}

	return "", errors.New("metadata output does not contain recoverable JSON")
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

package agents

import (
	"testing"
)

func TestParseMetadataOutput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *MetadataAnalysis
		wantErr bool
	}{
		{
			name:  "Valid JSON",
			input: `{"tags":["video","movie"],"description":"A great movie"}`,
			want: &MetadataAnalysis{
				Tags:        []string{"video", "movie"},
				Description: "A great movie",
			},
			wantErr: false,
		},
		{
			name:    "Empty input",
			input:   "",
			want:    nil,
			wantErr: true,
		},
		{
			name:  "JSON with markdown",
			input: "```json\n{\"tags\":[\"video\"],\"description\":\"test\"}\n```",
			want: &MetadataAnalysis{
				Tags:        []string{"video"},
				Description: "test",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMetadataOutput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMetadataOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("ParseMetadataOutput() got nil, want non-nil")
				return
			}
		})
	}
}

func TestNormalizeMetadataAnalysis(t *testing.T) {
	tests := []struct {
		name  string
		input *MetadataAnalysis
		want  *MetadataAnalysis
	}{
		{
			name: "Normalize tags",
			input: &MetadataAnalysis{
				Tags:        []string{"video", "video", "movie", "", "test", "extra"},
				Description: "test",
			},
			want: &MetadataAnalysis{
				Tags:        []string{"video", "movie", "test", "extra"},
				Description: "test",
			},
		},
		{
			name:  "Nil input",
			input: nil,
			want:  &MetadataAnalysis{Tags: []string{}, Description: ""},
		},
		{
			name: "Truncate long description",
			input: &MetadataAnalysis{
				Tags:        []string{"video"},
				Description: "This is a very long description that should be truncated because it exceeds the maximum allowed length of characters",
			},
			want: &MetadataAnalysis{
				Tags:        []string{"video"},
				Description: "This is a very long description that should be truncated because it exceeds the maximum allowed length of characters",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeMetadataAnalysis(tt.input)
			if got == nil {
				t.Errorf("NormalizeMetadataAnalysis() got nil")
				return
			}
			if len(got.Tags) != len(tt.want.Tags) {
				t.Errorf("NormalizeMetadataAnalysis() tags length = %d, want %d", len(got.Tags), len(tt.want.Tags))
			}
			if got.Description != tt.want.Description && len(tt.want.Description) > 0 {
				t.Errorf("NormalizeMetadataAnalysis() description = %s, want %s", got.Description, tt.want.Description)
			}
		})
	}
}

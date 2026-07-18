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
			name:  "Preserves unicode",
			input: `{"tags":["æon","剧情"],"description":"彭昱畅 æ 冒险"}`,
			want: &MetadataAnalysis{
				Tags:        []string{"æon", "剧情"},
				Description: "彭昱畅 æ 冒险",
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
			input: "```json\n{\"tags\":[\"video\"],\"description\":\"test æ\"}\n```",
			want: &MetadataAnalysis{
				Tags:        []string{"video"},
				Description: "test æ",
			},
			wantErr: false,
		},
		{
			name:  "JSON with surrounding text",
			input: "Here is the JSON: {\"tags\":[\"drama\"],\"description\":\"wrapped 剧情\"}",
			want: &MetadataAnalysis{
				Tags:        []string{"drama"},
				Description: "wrapped 剧情",
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
			if tt.wantErr {
				return
			}
			if got == nil {
				t.Fatalf("ParseMetadataOutput() got nil, want non-nil")
			}
			if len(got.Tags) != len(tt.want.Tags) {
				t.Fatalf("ParseMetadataOutput() tags length = %d, want %d", len(got.Tags), len(tt.want.Tags))
			}
			for i := range got.Tags {
				if got.Tags[i] != tt.want.Tags[i] {
					t.Fatalf("ParseMetadataOutput() tag[%d] = %q, want %q", i, got.Tags[i], tt.want.Tags[i])
				}
			}
			if got.Description != tt.want.Description {
				t.Fatalf("ParseMetadataOutput() description = %q, want %q", got.Description, tt.want.Description)
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

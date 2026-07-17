package agents

import (
	"strings"
	"testing"
)

func TestPromptBuilderBuildDescriptionPrompt(t *testing.T) {
	builder := NewPromptBuilder()

	testCases := []struct {
		name            string
		fileName        string
		fileType        string
		userKeywords    string
		userTags        []string
		userDescription string
		expectedParts   []string
	}{
		{
			name:            "document with description",
			fileName:        "report.pdf",
			fileType:        "document",
			userKeywords:    "business",
			userTags:        []string{"work"},
			userDescription: "Annual financial report",
			expectedParts: []string{
				"report.pdf",
				"document",
				"business",
				"work",
				"Annual financial report",
				"简洁准确的描述",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prompt := builder.BuildDescriptionPrompt(
				tc.fileName,
				tc.fileType,
				tc.userKeywords,
				tc.userTags,
				tc.userDescription,
			)

			for _, part := range tc.expectedParts {
				if !strings.Contains(prompt, part) {
					t.Errorf("Expected prompt to contain '%s'", part)
				}
			}
		})
	}
}
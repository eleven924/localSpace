package agents

import (
	"context"
	"strings"
	"testing"
	"time"

	"LocalSpace/app/models"
)

func TestDescriptionAgentGenerate(t *testing.T) {
	config := &AgentConfig{
		Name:    "description-agent",
		Timeout: 30 * time.Second,
	}

	agent := NewDescriptionAgent(config)
	if agent == nil {
		t.Fatal("Expected description agent to be created")
	}

	input := &DescriptionGenerationInput{
		FileName:        "test-document.pdf",
		FileType:        "document",
		UserKeywords:    "business report",
		UserTags:        []string{"work"},
		UserDescription: "Annual financial report",
	}

	ctx := context.Background()
	description, err := agent.Generate(ctx, input, &models.AIConfig{
		ID:          1,
		APIKey:      "test-key",
		Model:       "gpt-4",
		BaseURL:     "https://api.openai.com/v1",
		Enabled:     true,
		EnableAgent: true,
	})

	// Should return user description
	if description != "Annual financial report" {
		t.Errorf("Expected user description 'Annual financial report', got '%s'", description)
	}

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func TestDescriptionAgentGenerateBasicDescription(t *testing.T) {
	config := &AgentConfig{
		Name:    "description-agent",
		Timeout: 30 * time.Second,
	}

	agent := NewDescriptionAgent(config)

	testCases := []struct {
		name            string
		input           *DescriptionGenerationInput
		contains        []string
		maxLength       int
	}{
		{
			name: "document with keywords",
			input: &DescriptionGenerationInput{
				FileName:     "report.pdf",
				FileType:     "document",
				UserKeywords: "business",
				UserTags:     []string{"work"},
			},
			contains:  []string{"文档", "business", "work"},
			maxLength: 50,
		},
		{
			name: "minimal input",
			input: &DescriptionGenerationInput{
				FileName: "movie.mp4",
				FileType: "video",
			},
			contains:  []string{"视频"},
			maxLength: 50,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			description := agent.generateBasicDescription(tc.input)

			// Check that expected parts are present
			for _, expected := range tc.contains {
				if !strings.Contains(description, expected) {
					t.Errorf("Expected description to contain '%s', got '%s'", expected, description)
				}
			}

			// Check length limit
			if len(description) > tc.maxLength {
				t.Errorf("Description too long: %d chars, max %d", len(description), tc.maxLength)
			}
		})
	}
}

func TestDescriptionAgentUserDescription(t *testing.T) {
	config := &AgentConfig{
		Name:    "description-agent",
		Timeout: 30 * time.Second,
	}

	agent := NewDescriptionAgent(config)

	input := &DescriptionGenerationInput{
		FileName:        "test.pdf",
		FileType:        "document",
		UserDescription: "This is a test document for testing purposes",
	}

	ctx := context.Background()
	aiConfig := &models.AIConfig{
		ID:          1,
		APIKey:      "test-key",
		Model:       "gpt-4",
		BaseURL:     "https://api.openai.com/v1",
		Enabled:     true,
		EnableAgent: true,
	}

	description, err := agent.Generate(ctx, input, aiConfig)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should truncate long descriptions
	if len(description) > 53 { // 50 + "..."
		t.Errorf("Description too long: %d chars", len(description))
	}

	if !strings.Contains(description, "This is a test document") {
		t.Errorf("Expected description to contain user input, got '%s'", description)
	}
}

func TestDescriptionAgentDecideNeedSearch(t *testing.T) {
	config := &AgentConfig{
		Name:    "description-agent",
		Timeout: 30 * time.Second,
	}

	agent := NewDescriptionAgent(config)

	aiConfig := &models.AIConfig{
		Enabled:         true,
		EnableAgent:     true,
		EnableWebSearch: true,
	}

	testCases := []struct {
		name        string
		input       *DescriptionGenerationInput
		needSearch  bool
	}{
		{
			name: "no user input, descriptive filename",
			input: &DescriptionGenerationInput{
				FileName: "MyAwesomeDocument.pdf",
				FileType: "document",
			},
			needSearch: true,
		},
		{
			name: "has user description",
			input: &DescriptionGenerationInput{
				FileName:        "doc.pdf",
				FileType:        "document",
				UserDescription: "A document",
			},
			needSearch: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			needSearch, err := agent.decideNeedSearch(ctx, tc.input, aiConfig)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if needSearch != tc.needSearch {
				t.Errorf("Expected needSearch=%v, got %v", tc.needSearch, needSearch)
			}
		})
	}
}
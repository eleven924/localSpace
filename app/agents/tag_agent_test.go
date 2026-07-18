package agents

import (
	"context"
	"testing"
	"time"

	"LocalSpace/app/models"
)

func TestTagAgentGenerate(t *testing.T) {
	// Create agent config
	config := &AgentConfig{
		Name:    "tag-agent",
		Timeout: 30 * time.Second,
	}

	agent := NewTagAgent(config)
	if agent == nil {
		t.Fatal("Expected tag agent to be created")
	}

	// Create input
	input := &TagGenerationInput{
		FileName:        "test-video.mp4",
		FileType:        "video",
		UserKeywords:    "action movie",
		UserTags:        []string{"entertainment"},
		UserDescription: "An action movie",
	}

	// Disabled/unconfigured AI should use safe basic fallback tags without returning an error.
	ctx := context.Background()
	tags, err := agent.Generate(ctx, input, &models.AIConfig{})

	// Should return basic fallback tags and nil error for unconfigured agent.
	if len(tags) == 0 {
		t.Fatal("Expected fallback tags for unconfigured agent")
	}

	expected := map[string]bool{"video": false, "entertainment": false}
	for _, tag := range tags {
		if _, ok := expected[tag]; ok {
			expected[tag] = true
		}
	}
	for tag, found := range expected {
		if !found {
			t.Errorf("Expected fallback tag %q in %v", tag, tags)
		}
	}

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func TestTagAgentGenerateBasicTags(t *testing.T) {
	config := &AgentConfig{
		Name:    "tag-agent",
		Timeout: 30 * time.Second,
	}

	agent := NewTagAgent(config)

	testCases := []struct {
		name     string
		input    *TagGenerationInput
		expected []string
	}{
		{
			name: "video with user tags",
			input: &TagGenerationInput{
				FileName:     "movie.mp4",
				FileType:     "video",
				UserTags:     []string{"action", "drama"},
				UserKeywords: "film",
			},
			expected: []string{"video", "action", "drama", "film"},
		},
		{
			name: "minimal input",
			input: &TagGenerationInput{
				FileName: "document.pdf",
				FileType: "document",
			},
			expected: []string{"document"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tags := agent.generateBasicTags(tc.input)

			// Check that all expected tags are present
			for _, expectedTag := range tc.expected {
				found := false
				for _, tag := range tags {
					if tag == expectedTag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected tag '%s' not found in %v", expectedTag, tags)
				}
			}

			// Check tag limit
			if len(tags) > 5 {
				t.Errorf("Expected max 5 tags, got %d", len(tags))
			}
		})
	}
}

func TestTagAgentDecideNeedSearch(t *testing.T) {
	config := &AgentConfig{
		Name:    "tag-agent",
		Timeout: 30 * time.Second,
	}

	agent := NewTagAgent(config)

	aiConfig := &models.AIConfig{
		Enabled:         true,
		EnableAgent:     true,
		EnableWebSearch: true,
	}

	testCases := []struct {
		name       string
		input      *TagGenerationInput
		needSearch bool
	}{
		{
			name: "no user input, descriptive filename",
			input: &TagGenerationInput{
				FileName: "MyAwesomeMovie.mp4",
				FileType: "video",
			},
			needSearch: true,
		},
		{
			name: "has user keywords",
			input: &TagGenerationInput{
				FileName:     "movie.mp4",
				FileType:     "video",
				UserKeywords: "action",
			},
			needSearch: false,
		},
		{
			name: "has user tags",
			input: &TagGenerationInput{
				FileName: "movie.mp4",
				FileType: "video",
				UserTags: []string{"action"},
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

package models

import "testing"

func TestAIConfigAgentSettings(t *testing.T) {
	config := AIConfig{
		ID:              1,
		APIKey:          "test-key",
		Model:           "gpt-4",
		BaseURL:         "https://api.openai.com/v1",
		Enabled:         true,
		EnableAgent:     true,
		EnableWebSearch: true,
		MaxTokens:       500,
		Timeout:         30,
	}

	if !config.EnableAgent {
		t.Errorf("Expected EnableAgent to be true")
	}

	if !config.EnableWebSearch {
		t.Errorf("Expected EnableWebSearch to be true")
	}

	if config.MaxTokens != 500 {
		t.Errorf("Expected MaxTokens to be 500, got %d", config.MaxTokens)
	}

	if config.Timeout != 30 {
		t.Errorf("Expected Timeout to be 30, got %d", config.Timeout)
	}

	// Test default values
	defaultConfig := AIConfig{
		ID:      2,
		APIKey:  "test-key",
		Model:   "gpt-4",
		BaseURL: "https://api.openai.com/v1",
		Enabled: true,
	}

	if defaultConfig.EnableAgent {
		t.Errorf("Expected default EnableAgent to be false")
	}

	if defaultConfig.MaxTokens != 0 {
		t.Errorf("Expected default MaxTokens to be 0, got %d", defaultConfig.MaxTokens)
	}
}
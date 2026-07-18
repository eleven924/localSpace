package models

import "testing"

func TestAIConfigWebSearchSettings(t *testing.T) {
	config := AIConfig{
		WebSearchProvider:   "mock-http",
		WebSearchBaseURL:    "https://search.example.com",
		WebSearchAPIKey:     "search-key",
		WebSearchTimeout:    10,
		WebSearchMaxResults: 3,
	}

	if config.WebSearchProvider != "mock-http" {
		t.Fatalf("expected provider to round-trip, got %q", config.WebSearchProvider)
	}
	if config.WebSearchBaseURL != "https://search.example.com" {
		t.Fatalf("expected search base URL to round-trip, got %q", config.WebSearchBaseURL)
	}
	if config.WebSearchAPIKey != "search-key" {
		t.Fatalf("expected search API key to round-trip, got %q", config.WebSearchAPIKey)
	}
	if config.WebSearchTimeout != 10 {
		t.Fatalf("expected search timeout to round-trip, got %d", config.WebSearchTimeout)
	}
	if config.WebSearchMaxResults != 3 {
		t.Fatalf("expected max results to round-trip, got %d", config.WebSearchMaxResults)
	}
}

func TestAIConfigWebSearchDefaults(t *testing.T) {
	defaultConfig := AIConfig{
		ID:      2,
		APIKey:  "test-key",
		Model:   "gpt-4",
		BaseURL: "https://api.openai.com/v1",
		Enabled: true,
	}

	if defaultConfig.WebSearchProvider != "" {
		t.Fatalf("expected default search provider to be empty, got %q", defaultConfig.WebSearchProvider)
	}
	if defaultConfig.WebSearchBaseURL != "" {
		t.Fatalf("expected default search base URL to be empty, got %q", defaultConfig.WebSearchBaseURL)
	}
	if defaultConfig.WebSearchAPIKey != "" {
		t.Fatalf("expected default search API key to be empty, got %q", defaultConfig.WebSearchAPIKey)
	}
	if defaultConfig.WebSearchTimeout != 0 {
		t.Fatalf("expected zero-value timeout before repository defaults, got %d", defaultConfig.WebSearchTimeout)
	}
	if defaultConfig.WebSearchMaxResults != 0 {
		t.Fatalf("expected zero-value max results before repository defaults, got %d", defaultConfig.WebSearchMaxResults)
	}
}

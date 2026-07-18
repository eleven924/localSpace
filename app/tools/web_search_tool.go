package tools

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"LocalSpace/app/models"
)

// Tool is the minimal tool contract shared with the metadata runtime.
type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, input string) (string, error)
}

// SearchItem represents one bounded search result.
type SearchItem struct {
	Title   string
	URL     string
	Snippet string
}

// SearchResult represents a bounded tool search response.
type SearchResult struct {
	Query   string
	Results []SearchItem
}

// WebSearchTool executes bounded searches through a configured client.
type WebSearchTool struct {
	client     WebSearchClient
	timeout    time.Duration
	maxResults int
}

// NewWebSearchTool creates a new web search tool with sane defaults.
func NewWebSearchTool(client WebSearchClient, timeout time.Duration, maxResults int) *WebSearchTool {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	if maxResults <= 0 {
		maxResults = 3
	}
	if maxResults > 5 {
		maxResults = 5
	}

	return &WebSearchTool{
		client:     client,
		timeout:    timeout,
		maxResults: maxResults,
	}
}

// NewConfiguredWebSearchTool builds a tool from persisted AI config.
func NewConfiguredWebSearchTool(config *models.AIConfig) *WebSearchTool {
	if config == nil {
		return NewWebSearchTool(nil, 10*time.Second, 3)
	}

	timeout := time.Duration(config.WebSearchTimeout) * time.Second
	if config.WebSearchTimeout <= 0 {
		timeout = 10 * time.Second
	}

	client := NewHTTPWebSearchClient(
		config.WebSearchBaseURL,
		config.WebSearchAPIKey,
		config.WebSearchProvider,
		&http.Client{Timeout: timeout},
	)

	return NewWebSearchTool(client, timeout, config.WebSearchMaxResults)
}

// Name returns the runtime-visible tool name.
func (t *WebSearchTool) Name() string {
	return "web_search"
}

// Description returns the user-facing tool purpose.
func (t *WebSearchTool) Description() string {
	return "Search the web for short factual result snippets"
}

// Search validates, executes, and bounds one search request.
func (t *WebSearchTool) Search(ctx context.Context, query string) (*SearchResult, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return nil, errors.New("search query cannot be empty")
	}
	if len(trimmed) > 500 {
		return nil, errors.New("search query too long")
	}
	if t.client == nil {
		return nil, errors.New("web search client is not configured")
	}

	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	items, err := t.client.Search(ctx, trimmed, t.maxResults)
	if err != nil {
		return nil, err
	}

	if len(items) > t.maxResults {
		items = items[:t.maxResults]
	}
	for i := range items {
		if len(items[i].Snippet) > 240 {
			items[i].Snippet = items[i].Snippet[:240] + "..."
		}
	}

	return &SearchResult{Query: trimmed, Results: items}, nil
}

// Execute formats the bounded result for agent/runtime consumption.
func (t *WebSearchTool) Execute(ctx context.Context, input string) (string, error) {
	result, err := t.Search(ctx, input)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString("query: ")
	builder.WriteString(result.Query)
	builder.WriteString("\nresults:\n")
	for i, item := range result.Results {
		builder.WriteString(fmt.Sprintf("%d. Title: %s\n   URL: %s\n   Snippet: %s\n", i+1, item.Title, item.URL, item.Snippet))
	}
	if len(result.Results) == 0 {
		builder.WriteString("0. No results found\n")
	}

	return strings.TrimSpace(builder.String()), nil
}

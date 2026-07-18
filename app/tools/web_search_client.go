package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// WebSearchClient defines the search provider contract used by WebSearchTool.
type WebSearchClient interface {
	Search(ctx context.Context, query string, limit int) ([]SearchItem, error)
}

// HTTPWebSearchClient executes bounded HTTP GET searches against a configured endpoint.
type HTTPWebSearchClient struct {
	baseURL    string
	apiKey     string
	provider   string
	httpClient *http.Client
}

type httpWebSearchResponse struct {
	Results []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Snippet string `json:"snippet"`
	} `json:"results"`
}

// NewHTTPWebSearchClient creates a search client backed by a simple HTTP JSON API.
func NewHTTPWebSearchClient(baseURL, apiKey, provider string, httpClient *http.Client) *HTTPWebSearchClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &HTTPWebSearchClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		provider:   provider,
		httpClient: httpClient,
	}
}

// Search performs a single provider request and maps the bounded response format.
func (c *HTTPWebSearchClient) Search(ctx context.Context, query string, limit int) ([]SearchItem, error) {
	if strings.TrimSpace(c.baseURL) == "" {
		return nil, fmt.Errorf("web search base URL is not configured")
	}
	if strings.TrimSpace(c.apiKey) == "" {
		return nil, fmt.Errorf("web search API key is not configured")
	}

	endpoint, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid web search base URL: %w", err)
	}

	params := endpoint.Query()
	params.Set("q", query)
	params.Set("limit", fmt.Sprintf("%d", limit))
	if strings.TrimSpace(c.provider) != "" {
		params.Set("provider", c.provider)
	}
	endpoint.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build web search request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute web search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("web search returned status %d", resp.StatusCode)
	}

	var payload httpWebSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode web search response: %w", err)
	}

	items := make([]SearchItem, 0, len(payload.Results))
	for _, result := range payload.Results {
		items = append(items, SearchItem{
			Title:   strings.TrimSpace(result.Title),
			URL:     strings.TrimSpace(result.URL),
			Snippet: strings.TrimSpace(result.Snippet),
		})
	}

	return items, nil
}

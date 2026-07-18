package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		Content string `json:"content"` // Tavily使用content而不是snippet
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

	// Tavily API 使用 POST 请求，端点需要添加 /search
	searchEndpoint := strings.TrimRight(c.baseURL, "/") + "/search"

	// Tavily API 请求体格式
	requestBody := map[string]any{
		"query":        query,
		"search_depth": "basic",
		"max_results":  limit,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	fmt.Printf("[DEBUG] Tavily API Request - URL: %s, Query: %s, Limit: %d\n", searchEndpoint, query, limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, searchEndpoint, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("build web search request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute web search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 读取错误响应体以获取更多信息
		errorBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("[DEBUG] Tavily API Error - Status: %d, Body: %s\n", resp.StatusCode, string(errorBody))
		return nil, fmt.Errorf("web search returned status %d, response: %s", resp.StatusCode, string(errorBody))
	}

	var payload httpWebSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode web search response: %w", err)
	}

	fmt.Printf("[DEBUG] Tavily API Response - Results count: %d\n", len(payload.Results))

	items := make([]SearchItem, 0, len(payload.Results))
	for _, result := range payload.Results {
		items = append(items, SearchItem{
			Title:   strings.TrimSpace(result.Title),
			URL:     strings.TrimSpace(result.URL),
			Snippet: strings.TrimSpace(result.Content), // Tavily使用content字段
		})
	}

	return items, nil
}

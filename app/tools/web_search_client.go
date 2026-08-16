package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
		Content string `json:"content"` // Tavily使用content而不是snippet
	} `json:"results"`
}

// NewHTTPWebSearchClient creates a search client backed by a simple HTTP JSON API.
func NewHTTPWebSearchClient(baseURL, apiKey, provider string, httpClient *http.Client) *HTTPWebSearchClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &HTTPWebSearchClient{
		baseURL:    normalizeWebSearchBaseURL(baseURL),
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

	// Tavily API 使用 POST 请求；同时兼容用户填写 host 或完整 /search 地址。
	searchEndpoint, err := buildWebSearchEndpoint(c.baseURL)
	if err != nil {
		return nil, err
	}

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

// normalizeWebSearchBaseURL cleans values copied from JSON, environment files,
// or the settings form before they reach net/http.
func normalizeWebSearchBaseURL(raw string) string {
	value := strings.TrimSpace(raw)
	for len(value) >= 2 {
		wrappedInDoubleQuotes := value[0] == '"' && value[len(value)-1] == '"'
		wrappedInSingleQuotes := value[0] == '\'' && value[len(value)-1] == '\''
		if !wrappedInDoubleQuotes && !wrappedInSingleQuotes {
			break
		}
		value = strings.TrimSpace(value[1 : len(value)-1])
	}
	return strings.TrimRight(value, "/")
}

// buildWebSearchEndpoint validates the URL and appends /search only when needed.
func buildWebSearchEndpoint(baseURL string) (string, error) {
	value := normalizeWebSearchBaseURL(baseURL)
	if value == "" {
		return "", fmt.Errorf("web search base URL is not configured")
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("web search base URL is invalid: %q", value)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("web search base URL must use http or https")
	}
	if !strings.HasSuffix(strings.TrimRight(parsed.Path, "/"), "/search") {
		parsed.Path = strings.TrimRight(parsed.Path, "/") + "/search"
	}
	return parsed.String(), nil
}

package tools

import (
	"context"
	"encoding/json"
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
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"content"`
}

// SearchResult represents a bounded tool search response.
type SearchResult struct {
	Query   string       `json:"query"`
	Results []SearchItem `json:"results"`
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

	// 优化搜索查询：去除文件扩展名，提取关键词
	optimizedQuery := optimizeSearchQuery(trimmed)

	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	items, err := t.client.Search(ctx, optimizedQuery, t.maxResults)
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

// Execute formats the bounded result for agent/runtime consumption in JSON format.
func (t *WebSearchTool) Execute(ctx context.Context, input string) (string, error) {
	result, err := t.Search(ctx, input)
	if err != nil {
		return "", err
	}

	// Return structured JSON format for AI to process
	toolResult := map[string]any{
		"type":    "search_results",
		"query":   result.Query,
		"results": result.Results,
	}

	jsonOutput, err := json.Marshal(toolResult)
	if err != nil {
		return "", fmt.Errorf("failed to marshal search results to JSON: %w", err)
	}

	return string(jsonOutput), nil
}

// optimizeSearchQuery 优化搜索查询，去除文件扩展名和不必要的字符
func optimizeSearchQuery(query string) string {
	// 去除常见的文件扩展名
	extensions := []string{
		".mp4", ".mkv", ".avi", ".mov", ".flv", ".wmv", ".webm",
		".mp3", ".flac", ".wav", ".aac", ".m4a",
		".pdf", ".doc", ".docx", ".txt", ".rtf",
		".jpg", ".jpeg", ".png", ".gif", ".bmp",
		".zip", ".rar", ".7z", ".tar", ".gz",
		".exe", ".msi", ".dmg", ".app",
	}

	optimized := query
	for _, ext := range extensions {
		optimized = strings.ReplaceAll(optimized, ext, " ")
	}

	// 去除特殊字符和多余空格
	optimized = strings.ReplaceAll(optimized, "_", " ")
	optimized = strings.ReplaceAll(optimized, "-", " ")
	optimized = strings.ReplaceAll(optimized, ".", " ")

	// 去除连续空格
	spaces := strings.Fields(optimized)
	if len(spaces) == 0 {
		return query
	}

	// 重新组合，保留有意义的词汇
	var meaningful []string
	for _, word := range spaces {
		// 去除纯数字或短数字后缀
		if isPureNumeric(word) || (len(word) <= 2 && isNumeric(word)) {
			continue
		}
		meaningful = append(meaningful, word)
	}

	if len(meaningful) == 0 {
		return query // 如果过滤后没有内容，返回原查询
	}

	result := strings.Join(meaningful, " ")

	// 如果结果太短，保留部分原始查询
	if len(result) < 3 {
		return query
	}

	return result
}

// isPureNumeric 检查字符串是否纯数字
func isPureNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// isNumeric 检查字符串是否包含数字
func isNumeric(s string) bool {
	for _, c := range s {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	return false
}

# Web Search Tool Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a real HTTP-backed `web_search` tool and wire the metadata agent runtime so agents can execute the tool during metadata generation and record accurate trace data.

**Architecture:** Keep `AgentService` responsible for tool registration, exposure policy, and fallback while making `WebSearchTool` a thin wrapper around a configurable `WebSearchClient`. Extend `DefaultAgentRuntime` from a single direct model call to a bounded tool-calling loop that can expose `web_search`, execute it, feed results back to the model, and preserve `ToolsUsed` / `SearchQueries` trace data.

**Tech Stack:** Go, SQLite migrations, Wails/Vue, CloudWeGo Eino OpenAI-compatible chat model integration, `net/http`, Go testing package.

## Global Constraints

- Preserve the existing service-layer gating rule for whether `web_search` is available.
- Do not build a general multi-tool orchestration platform.
- Do not introduce webpage fetching, crawling, or full-document reading.
- Keep the first runtime loop intentionally bounded to one or two tool rounds.
- Keep search provider configuration independent from LLM model configuration.
- Search failures and runtime failures must not block metadata generation or file import.
- `ToolsAvailable` must remain service-derived; `ToolsUsed` and `SearchQueries` must reflect actual runtime execution only.
- Keep tool results bounded: reject empty queries, cap query length, cap returned results to 3–5, and truncate verbose snippets/output.

---

## File Structure

- `app/tools/web_search_tool.go` — keep the `Tool` implementation, but refactor it into a thin orchestrator that validates input, calls a client, and formats bounded output.
- `app/tools/web_search_client.go` — add search-provider abstractions and the default HTTP client implementation.
- `app/tools/web_search_tool_test.go` — cover tool validation, output formatting, result bounding, and provider error handling.
- `app/agents/agent_runtime.go` — extend the metadata runtime to support a bounded model → tool → model loop and accurate tool trace capture.
- `app/agents/agent_runtime_test.go` — cover no-tool behavior, tool execution behavior, trace capture, and failure handling.
- `app/models/config.go` — extend `AIConfig` with independent search-provider settings.
- `app/models/config_test.go` — cover default values and struct field expectations for the new search config.
- `app/repositories/config_repository.go` — persist and load the new search-provider fields with safe defaults.
- `app/database/migrations.go` — add a migration to extend `ai_configs` with search-provider fields.
- `frontend/src/components/AIConfigForm.vue` — expose the new search-provider configuration inputs in the UI.
- `docs/agent-integration.md` — update documentation so it describes a real tool-calling runtime instead of tool readiness only.

### Task 1: Add Search Provider Configuration Plumbing

**Files:**
- Modify: `app/models/config.go`
- Modify: `app/models/config_test.go`
- Modify: `app/repositories/config_repository.go`
- Modify: `app/database/migrations.go`
- Modify: `frontend/src/components/AIConfigForm.vue`

**Interfaces:**
- Consumes: existing `models.AIConfig`, `ConfigRepository.GetAIConfig() (*models.AIConfig, error)`, `ConfigRepository.SetAIConfig(config *models.AIConfig) error`
- Produces:
  - `type AIConfig struct { WebSearchProvider string; WebSearchBaseURL string; WebSearchAPIKey string; WebSearchTimeout int; WebSearchMaxResults int }`
  - repository read/write support for those fields
  - UI fields bound to `config.webSearchProvider`, `config.webSearchBaseURL`, `config.webSearchAPIKey`, `config.webSearchTimeout`, `config.webSearchMaxResults`

- [ ] **Step 1: Write the failing model test**

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./app/models -run TestAIConfigWebSearchSettings -count=1`
Expected: FAIL with unknown `WebSearchProvider` / `WebSearchBaseURL` / `WebSearchAPIKey` / `WebSearchTimeout` / `WebSearchMaxResults` fields on `AIConfig`.

- [ ] **Step 3: Add the new config fields to `AIConfig`**

```go
type AIConfig struct {
	ID      uint   `json:"id"`
	APIKey  string `json:"apiKey"`
	Model   string `json:"model"`
	BaseURL string `json:"baseURL"`
	Enabled bool   `json:"enabled"`

	EnableAgent     bool `json:"enableAgent"`
	EnableWebSearch bool `json:"enableWebSearch"`
	MaxTokens       int  `json:"maxTokens"`
	Timeout         int  `json:"timeout"`

	WebSearchProvider   string `json:"webSearchProvider"`
	WebSearchBaseURL    string `json:"webSearchBaseURL"`
	WebSearchAPIKey     string `json:"webSearchAPIKey"`
	WebSearchTimeout    int    `json:"webSearchTimeout"`
	WebSearchMaxResults int    `json:"webSearchMaxResults"`
}
```

- [ ] **Step 4: Run the model test to verify it passes**

Run: `go test ./app/models -run TestAIConfigWebSearchSettings -count=1`
Expected: PASS

- [ ] **Step 5: Write the failing repository default test**

Add this test to `app/models/config_test.go` or a repository-focused config test file if one already exists nearby:

```go
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
```

- [ ] **Step 6: Run the config tests**

Run: `go test ./app/models -run 'TestAIConfig(WebSearchSettings|WebSearchDefaults)' -count=1`
Expected: PASS for the struct tests.

- [ ] **Step 7: Write the failing repository/defaults implementation test**

Add a repository test like this in `app/repositories/config_repository_test.go`:

```go
func TestGetAIConfigReturnsWebSearchDefaultsWhenRowMissing(t *testing.T) {
	repo := newTestConfigRepository(t)

	config, err := repo.GetAIConfig()
	if err != nil {
		t.Fatalf("GetAIConfig returned error: %v", err)
	}

	if config.WebSearchTimeout != 10 {
		t.Fatalf("expected default web search timeout 10, got %d", config.WebSearchTimeout)
	}
	if config.WebSearchMaxResults != 3 {
		t.Fatalf("expected default web search max results 3, got %d", config.WebSearchMaxResults)
	}
	if config.WebSearchProvider != "" {
		t.Fatalf("expected default provider empty, got %q", config.WebSearchProvider)
	}
}
```

- [ ] **Step 8: Run the repository test to verify it fails**

Run: `go test ./app/repositories -run TestGetAIConfigReturnsWebSearchDefaultsWhenRowMissing -count=1`
Expected: FAIL because the repository does not yet populate the new fields/defaults.

- [ ] **Step 9: Add a migration for the new `ai_configs` columns**

Append a new migration entry and migration functions in `app/database/migrations.go`.

```go
{
	Version: 6,
	Name:    "add_ai_config_web_search_provider_fields",
	Up:      migration006_Up,
	Down:    migration006_Down,
},
```

```go
func migration006_Up(db *sql.DB) error {
	columns := []struct {
		name       string
		definition string
	}{
		{name: "web_search_provider", definition: "TEXT DEFAULT ''"},
		{name: "web_search_base_url", definition: "TEXT DEFAULT ''"},
		{name: "web_search_api_key", definition: "TEXT DEFAULT ''"},
		{name: "web_search_timeout", definition: "INTEGER DEFAULT 10"},
		{name: "web_search_max_results", definition: "INTEGER DEFAULT 3"},
	}

	for _, column := range columns {
		var exists bool
		err := db.QueryRow(`
			SELECT COUNT(*) > 0
			FROM pragma_table_info('ai_configs')
			WHERE name = ?
		`, column.name).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check %s column existence: %w", column.name, err)
		}
		if exists {
			continue
		}
		if _, err := db.Exec(fmt.Sprintf("ALTER TABLE ai_configs ADD COLUMN %s %s", column.name, column.definition)); err != nil {
			return fmt.Errorf("failed to add %s column: %w", column.name, err)
		}
	}

	return nil
}

func migration006_Down(db *sql.DB) error {
	return fmt.Errorf("SQLite rollback not supported for column additions")
}
```

- [ ] **Step 10: Update repository load/save SQL and defaults**

Update `GetAIConfig()` and `SetAIConfig()` in `app/repositories/config_repository.go`.

```go
query := `SELECT id, api_key, model, base_url, enabled, enable_agent, enable_web_search, max_tokens, timeout, web_search_provider, web_search_base_url, web_search_api_key, web_search_timeout, web_search_max_results FROM ai_configs WHERE id = 1`
```

```go
err := r.db.QueryRow(query).Scan(
	&config.ID,
	&config.APIKey,
	&config.Model,
	&config.BaseURL,
	&config.Enabled,
	&config.EnableAgent,
	&config.EnableWebSearch,
	&config.MaxTokens,
	&config.Timeout,
	&config.WebSearchProvider,
	&config.WebSearchBaseURL,
	&config.WebSearchAPIKey,
	&config.WebSearchTimeout,
	&config.WebSearchMaxResults,
)
```

Default return when no row exists:

```go
return &models.AIConfig{
	ID:                  0,
	APIKey:              "",
	Model:               "gpt-3.5-turbo",
	BaseURL:             "https://api.openai.com/v1",
	Enabled:             false,
	EnableAgent:         false,
	EnableWebSearch:     false,
	MaxTokens:           500,
	Timeout:             30,
	WebSearchProvider:   "",
	WebSearchBaseURL:    "",
	WebSearchAPIKey:     "",
	WebSearchTimeout:    10,
	WebSearchMaxResults: 3,
}, nil
```

Update `INSERT ... ON CONFLICT`:

```go
INSERT INTO ai_configs (
	id, api_key, model, base_url, enabled,
	enable_agent, enable_web_search, max_tokens, timeout,
	web_search_provider, web_search_base_url, web_search_api_key, web_search_timeout, web_search_max_results,
	updated_at
)
VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(id) DO UPDATE SET
	api_key = excluded.api_key,
	model = excluded.model,
	base_url = excluded.base_url,
	enabled = excluded.enabled,
	enable_agent = excluded.enable_agent,
	enable_web_search = excluded.enable_web_search,
	max_tokens = excluded.max_tokens,
	timeout = excluded.timeout,
	web_search_provider = excluded.web_search_provider,
	web_search_base_url = excluded.web_search_base_url,
	web_search_api_key = excluded.web_search_api_key,
	web_search_timeout = excluded.web_search_timeout,
	web_search_max_results = excluded.web_search_max_results,
	updated_at = CURRENT_TIMESTAMP
```

Normalize save defaults before the `Exec` call:

```go
webSearchTimeout := config.WebSearchTimeout
if webSearchTimeout == 0 {
	webSearchTimeout = 10
}
webSearchMaxResults := config.WebSearchMaxResults
if webSearchMaxResults == 0 {
	webSearchMaxResults = 3
}
```

`Exec` argument tail:

```go
config.APIKey,
config.Model,
config.BaseURL,
config.Enabled,
enableAgent,
enableWebSearch,
maxTokens,
timeout,
config.WebSearchProvider,
config.WebSearchBaseURL,
config.WebSearchAPIKey,
webSearchTimeout,
webSearchMaxResults,
```

- [ ] **Step 11: Run the repository test to verify it passes**

Run: `go test ./app/repositories -run TestGetAIConfigReturnsWebSearchDefaultsWhenRowMissing -count=1`
Expected: PASS

- [ ] **Step 12: Add the UI fields for search-provider config**

Extend the `AIConfig` interface and initial state in `frontend/src/components/AIConfigForm.vue`.

```ts
interface AIConfig {
  id?: number
  enabled: boolean
  apiKey: string
  model: string
  baseURL: string
  timeout?: number
  maxTokens?: number
  enableAgent?: boolean
  enableWebSearch?: boolean
  webSearchProvider?: string
  webSearchBaseURL?: string
  webSearchAPIKey?: string
  webSearchTimeout?: number
  webSearchMaxResults?: number
}
```

```ts
const config = ref<AIConfig>({
  enabled: false,
  apiKey: '',
  model: 'gpt-3.5-turbo',
  baseURL: 'https://api.openai.com/v1',
  timeout: 30,
  maxTokens: 1000,
  enableAgent: true,
  enableWebSearch: false,
  webSearchProvider: '',
  webSearchBaseURL: '',
  webSearchAPIKey: '',
  webSearchTimeout: 10,
  webSearchMaxResults: 3,
})
```

Hydration block in `loadConfig()`:

```ts
webSearchProvider: loadedConfig.webSearchProvider || '',
webSearchBaseURL: loadedConfig.webSearchBaseURL || '',
webSearchAPIKey: loadedConfig.webSearchAPIKey || '',
webSearchTimeout: loadedConfig.webSearchTimeout || 10,
webSearchMaxResults: loadedConfig.webSearchMaxResults || 3,
```

Add these inputs inside the `v-if="config.enableAgent"` block, nested under `v-if="config.enableWebSearch"`:

```vue
<div v-if="config.enableWebSearch" class="search-config-fields">
  <div class="form-group">
    <label for="web-search-provider">
      <span class="label-text">搜索 Provider</span>
    </label>
    <input
      id="web-search-provider"
      v-model="config.webSearchProvider"
      type="text"
      placeholder="例如: mock-http"
      @input="handleConfigChange"
    />
    <p class="form-hint">标识当前网络搜索服务类型</p>
  </div>

  <div class="form-group">
    <label for="web-search-base-url">
      <span class="label-text">搜索 Base URL</span>
      <span class="label-required">*</span>
    </label>
    <input
      id="web-search-base-url"
      v-model="config.webSearchBaseURL"
      type="text"
      placeholder="https://search.example.com"
      @input="handleConfigChange"
    />
    <p class="form-hint">网络搜索服务端点地址</p>
  </div>

  <div class="form-group">
    <label for="web-search-api-key">
      <span class="label-text">搜索 API Key</span>
      <span class="label-required">*</span>
    </label>
    <input
      id="web-search-api-key"
      v-model="config.webSearchAPIKey"
      type="password"
      placeholder="输入网络搜索服务 API Key"
      @input="handleConfigChange"
    />
    <p class="form-hint">仅用于 web_search tool，不复用模型 API Key</p>
  </div>

  <div class="form-group">
    <label for="web-search-timeout">
      <span class="label-text">搜索超时时间</span>
      <span class="label-hint">秒</span>
    </label>
    <input
      id="web-search-timeout"
      v-model.number="config.webSearchTimeout"
      type="number"
      min="1"
      max="60"
      @input="handleConfigChange"
    />
    <p class="form-hint">单次搜索请求超时时间，默认 10 秒</p>
  </div>

  <div class="form-group">
    <label for="web-search-max-results">
      <span class="label-text">搜索结果数量上限</span>
    </label>
    <input
      id="web-search-max-results"
      v-model.number="config.webSearchMaxResults"
      type="number"
      min="1"
      max="5"
      @input="handleConfigChange"
    />
    <p class="form-hint">限制返回给 Agent 的结果数量，建议 3</p>
  </div>
</div>
```

- [ ] **Step 13: Run the targeted Go tests**

Run: `go test ./app/models ./app/repositories -count=1`
Expected: PASS

- [ ] **Step 14: Commit**

```bash
git add app/models/config.go app/models/config_test.go app/repositories/config_repository.go app/database/migrations.go frontend/src/components/AIConfigForm.vue
git commit -m "feat: add web search provider config"
```

### Task 2: Implement the HTTP-Backed Web Search Tool

**Files:**
- Create: `app/tools/web_search_client.go`
- Modify: `app/tools/web_search_tool.go`
- Modify: `app/tools/web_search_tool_test.go`

**Interfaces:**
- Consumes: `models.AIConfig.WebSearchProvider`, `models.AIConfig.WebSearchBaseURL`, `models.AIConfig.WebSearchAPIKey`, `models.AIConfig.WebSearchTimeout`, `models.AIConfig.WebSearchMaxResults`
- Produces:
  - `type WebSearchClient interface { Search(ctx context.Context, query string, limit int) ([]SearchItem, error) }`
  - `func NewHTTPWebSearchClient(baseURL, apiKey, provider string, httpClient *http.Client) *HTTPWebSearchClient`
  - `func NewConfiguredWebSearchTool(config *models.AIConfig) *WebSearchTool`
  - `func (t *WebSearchTool) Execute(ctx context.Context, input string) (string, error)` returning bounded formatted results

- [ ] **Step 1: Write the failing tool-formatting test**

Replace the placeholder execute assertion in `app/tools/web_search_tool_test.go` with a fake-client-based test:

```go
type fakeWebSearchClient struct {
	results []SearchItem
	err     error
	query   string
	limit   int
}

func (f *fakeWebSearchClient) Search(ctx context.Context, query string, limit int) ([]SearchItem, error) {
	f.query = query
	f.limit = limit
	return f.results, f.err
}

func TestWebSearchToolExecuteFormatsBoundedResults(t *testing.T) {
	client := &fakeWebSearchClient{results: []SearchItem{
		{Title: "Result 1", URL: "https://example.com/1", Snippet: "First snippet"},
		{Title: "Result 2", URL: "https://example.com/2", Snippet: "Second snippet"},
		{Title: "Result 3", URL: "https://example.com/3", Snippet: "Third snippet"},
		{Title: "Result 4", URL: "https://example.com/4", Snippet: "Fourth snippet"},
	}}
	tool := NewWebSearchTool(client, 10*time.Second, 3)

	result, err := tool.Execute(context.Background(), "movie")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if client.query != "movie" {
		t.Fatalf("expected query to be forwarded, got %q", client.query)
	}
	if client.limit != 3 {
		t.Fatalf("expected limit 3, got %d", client.limit)
	}
	if !strings.Contains(result, "query: movie") {
		t.Fatalf("expected formatted query in output, got %q", result)
	}
	if strings.Contains(result, "Result 4") {
		t.Fatalf("expected output to exclude truncated result, got %q", result)
	}
}
```

- [ ] **Step 2: Run the tool test to verify it fails**

Run: `go test ./app/tools -run TestWebSearchToolExecuteFormatsBoundedResults -count=1`
Expected: FAIL because `WebSearchTool` does not yet accept a client or return formatted bounded results.

- [ ] **Step 3: Add the search client abstraction and HTTP client**

Create `app/tools/web_search_client.go` with this implementation skeleton:

```go
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type WebSearchClient interface {
	Search(ctx context.Context, query string, limit int) ([]SearchItem, error)
}

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
```

- [ ] **Step 4: Refactor `WebSearchTool` into a thin orchestrator**

Update `app/tools/web_search_tool.go` like this:

```go
type WebSearchTool struct {
	client     WebSearchClient
	timeout    time.Duration
	maxResults int
}

func NewWebSearchTool(client WebSearchClient, timeout time.Duration, maxResults int) *WebSearchTool {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	if maxResults <= 0 {
		maxResults = 3
	}
	return &WebSearchTool{
		client:     client,
		timeout:    timeout,
		maxResults: maxResults,
	}
}
```

```go
func NewConfiguredWebSearchTool(config *models.AIConfig) *WebSearchTool {
	if config == nil {
		return NewWebSearchTool(nil, 10*time.Second, 3)
	}
	timeout := time.Duration(config.WebSearchTimeout) * time.Second
	if config.WebSearchTimeout <= 0 {
		timeout = 10 * time.Second
	}
	client := NewHTTPWebSearchClient(config.WebSearchBaseURL, config.WebSearchAPIKey, config.WebSearchProvider, &http.Client{Timeout: timeout})
	return NewWebSearchTool(client, timeout, config.WebSearchMaxResults)
}
```

```go
func (t *WebSearchTool) Search(ctx context.Context, query string) (*SearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("search query cannot be empty")
	}
	if len(query) > 500 {
		return nil, errors.New("search query too long")
	}
	if t.client == nil {
		return nil, errors.New("web search client is not configured")
	}

	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	items, err := t.client.Search(ctx, strings.TrimSpace(query), t.maxResults)
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

	return &SearchResult{Query: strings.TrimSpace(query), Results: items}, nil
}
```

```go
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
```

- [ ] **Step 5: Add missing imports to `web_search_tool.go`**

Ensure the file imports include:

```go
import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"LocalSpace/app/models"
)
```

- [ ] **Step 6: Add the provider-config-missing test**

```go
func TestWebSearchToolSearchRequiresConfiguredClient(t *testing.T) {
	tool := NewWebSearchTool(nil, 10*time.Second, 3)

	_, err := tool.Search(context.Background(), "movie")
	if err == nil {
		t.Fatal("expected missing client to return error")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("expected not configured error, got %v", err)
	}
}
```

- [ ] **Step 7: Add the HTTP client success test**

```go
func TestHTTPWebSearchClientSearchMapsResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q"); got != "movie" {
			t.Fatalf("expected query movie, got %q", got)
		}
		if got := r.URL.Query().Get("limit"); got != "3" {
			t.Fatalf("expected limit 3, got %q", got)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer search-key" {
			t.Fatalf("expected bearer auth, got %q", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"title":"Movie Title","url":"https://example.com/movie","snippet":"Movie snippet"}]}`))
	}))
	defer server.Close()

	client := NewHTTPWebSearchClient(server.URL, "search-key", "mock-http", server.Client())
	items, err := client.Search(context.Background(), "movie", 3)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Title != "Movie Title" {
		t.Fatalf("expected mapped title, got %q", items[0].Title)
	}
}
```

- [ ] **Step 8: Run the targeted tool tests**

Run: `go test ./app/tools -run 'Test(WebSearchTool|HTTPWebSearchClient)' -count=1`
Expected: PASS

- [ ] **Step 9: Commit**

```bash
git add app/tools/web_search_client.go app/tools/web_search_tool.go app/tools/web_search_tool_test.go
git commit -m "feat: implement http-backed web search tool"
```

### Task 3: Teach the Metadata Runtime to Execute Tools

**Files:**
- Modify: `app/agents/agent_runtime.go`
- Modify: `app/agents/agent_runtime_test.go`

**Interfaces:**
- Consumes: `AgentRunRequest`, `AgentRunResponse`, `tools.Tool`, `Tool.Name() string`, `Tool.Description() string`, `Tool.Execute(ctx context.Context, input string) (string, error)`
- Produces:
  - bounded runtime loop in `func (r *DefaultAgentRuntime) Run(ctx context.Context, req *AgentRunRequest) (*AgentRunResponse, error)`
  - helper methods:
    - `func buildToolMap(tools []tools.Tool) map[string]tools.Tool`
    - `func appendToolDefinitions(messages []*schema.Message, availableTools []tools.Tool) []*schema.Message` (or an equivalent helper if the Eino API exposes tool binding another way)
    - `func recordSearchQuery(toolName, input string, searchQueries []string) []string`

- [ ] **Step 1: Write the failing runtime tool-usage test**

Add these test doubles to `app/agents/agent_runtime_test.go`:

```go
type runtimeTestTool struct {
	name  string
	calls []string
	result string
}

func (t *runtimeTestTool) Name() string { return t.name }
func (t *runtimeTestTool) Description() string { return "test tool" }
func (t *runtimeTestTool) Execute(ctx context.Context, input string) (string, error) {
	t.calls = append(t.calls, input)
	return t.result, nil
}
```

Then add a focused test around a helper you introduce for tool execution:

```go
func TestExecuteToolCallRecordsUsageAndSearchQuery(t *testing.T) {
	tool := &runtimeTestTool{name: "web_search", result: "query: movie\nresults:\n1. Title: Movie"}
	runtime := NewDefaultAgentRuntime()

	output, toolsUsed, searchQueries, err := runtime.executeToolCall(context.Background(), map[string]tools.Tool{
		"web_search": tool,
	}, "web_search", "movie", nil, nil)
	if err != nil {
		t.Fatalf("executeToolCall returned error: %v", err)
	}
	if output != "query: movie\nresults:\n1. Title: Movie" {
		t.Fatalf("unexpected tool output %q", output)
	}
	if len(toolsUsed) != 1 || toolsUsed[0] != "web_search" {
		t.Fatalf("expected web_search recorded in ToolsUsed, got %v", toolsUsed)
	}
	if len(searchQueries) != 1 || searchQueries[0] != "movie" {
		t.Fatalf("expected movie recorded in SearchQueries, got %v", searchQueries)
	}
}
```

- [ ] **Step 2: Run the runtime test to verify it fails**

Run: `go test ./app/agents -run TestExecuteToolCallRecordsUsageAndSearchQuery -count=1`
Expected: FAIL because `executeToolCall` does not exist yet.

- [ ] **Step 3: Add the helper for executing named tools**

Add this helper to `app/agents/agent_runtime.go`:

```go
func (r *DefaultAgentRuntime) executeToolCall(
	ctx context.Context,
	toolMap map[string]tools.Tool,
	toolName string,
	input string,
	toolsUsed []string,
	searchQueries []string,
) (string, []string, []string, error) {
	tool, ok := toolMap[toolName]
	if !ok {
		return "", toolsUsed, searchQueries, fmt.Errorf("tool not found: %s", toolName)
	}

	output, err := tool.Execute(ctx, input)
	if err != nil {
		return "", toolsUsed, searchQueries, fmt.Errorf("execute tool %s: %w", toolName, err)
	}

	toolsUsed = append(toolsUsed, toolName)
	if toolName == "web_search" && strings.TrimSpace(input) != "" {
		searchQueries = append(searchQueries, strings.TrimSpace(input))
	}

	return output, toolsUsed, searchQueries, nil
}
```

- [ ] **Step 4: Run the helper test to verify it passes**

Run: `go test ./app/agents -run TestExecuteToolCallRecordsUsageAndSearchQuery -count=1`
Expected: PASS

- [ ] **Step 5: Add a failing no-tool regression test**

```go
func TestDefaultAgentRuntimeRun_WithoutToolsReturnsDirectOutput(t *testing.T) {
	runtime := &DefaultAgentRuntime{
		generate: func(ctx context.Context, messages []*schema.Message, req *AgentRunRequest) (*schema.Message, error) {
			return schema.AssistantMessage(`{"tags":["video"],"description":"影片"}`), nil
		},
	}

	resp, err := runtime.Run(context.Background(), &AgentRunRequest{
		SystemPrompt: "sys",
		UserPrompt:   "user",
		AIConfig:     &models.AIConfig{APIKey: "key", Model: "model", BaseURL: "https://api.example.com/v1"},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if resp.Output != `{"tags":["video"],"description":"影片"}` {
		t.Fatalf("unexpected output %q", resp.Output)
	}
	if len(resp.ToolsUsed) != 0 {
		t.Fatalf("expected no tools used, got %v", resp.ToolsUsed)
	}
}
```

- [ ] **Step 6: Introduce an injectable `generate` seam**

Refactor `DefaultAgentRuntime` from:

```go
type DefaultAgentRuntime struct{}
```

to:

```go
type DefaultAgentRuntime struct {
	generate func(ctx context.Context, messages []*schema.Message, req *AgentRunRequest) (*schema.Message, error)
}
```

Update the constructor:

```go
func NewDefaultAgentRuntime() *DefaultAgentRuntime {
	runtime := &DefaultAgentRuntime{}
	runtime.generate = runtime.generateWithChatModel
	return runtime
}
```

Add a helper that preserves the existing model creation path:

```go
func (r *DefaultAgentRuntime) generateWithChatModel(ctx context.Context, messages []*schema.Message, req *AgentRunRequest) (*schema.Message, error) {
	chatConfig := &openai.ChatModelConfig{
		APIKey:  req.AIConfig.APIKey,
		Model:   req.AIConfig.Model,
		BaseURL: req.AIConfig.BaseURL,
	}
	if req.AIConfig.Timeout > 0 {
		chatConfig.Timeout = time.Duration(req.AIConfig.Timeout) * time.Second
	}
	if req.AIConfig.MaxTokens > 0 {
		maxTokens := req.AIConfig.MaxTokens
		chatConfig.MaxTokens = &maxTokens
	}
	temperature := float32(0.2)
	chatConfig.Temperature = &temperature

	chatModel, err := openai.NewChatModel(ctx, chatConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}

	resp, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}
	if resp.Content == "" {
		return nil, fmt.Errorf("empty response from AI")
	}
	return schema.AssistantMessage(resp.Content), nil
}
```

- [ ] **Step 7: Implement the bounded runtime loop**

Replace the body of `Run` with this shape:

```go
func (r *DefaultAgentRuntime) Run(ctx context.Context, req *AgentRunRequest) (*AgentRunResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("agent run request is nil")
	}
	if req.AIConfig == nil {
		return nil, fmt.Errorf("ai config is nil")
	}
	if req.AIConfig.APIKey == "" || req.AIConfig.Model == "" || req.AIConfig.BaseURL == "" {
		return nil, fmt.Errorf("invalid ai config")
	}
	if r.generate == nil {
		r.generate = r.generateWithChatModel
	}

	messages := []*schema.Message{
		schema.SystemMessage(req.SystemPrompt),
		schema.UserMessage(req.UserPrompt),
	}
	toolMap := buildToolMap(req.Tools)
	toolsUsed := []string{}
	searchQueries := []string{}

	assistantMsg, err := r.generate(ctx, messages, req)
	if err != nil {
		return nil, err
	}
	messages = append(messages, assistantMsg)

	toolName, toolInput, wantsTool := parseToolCall(assistantMsg.Content)
	maxRounds := 2
	for wantsTool && maxRounds > 0 {
		toolOutput, updatedToolsUsed, updatedSearchQueries, err := r.executeToolCall(ctx, toolMap, toolName, toolInput, toolsUsed, searchQueries)
		if err != nil {
			return nil, err
		}
		toolsUsed = updatedToolsUsed
		searchQueries = updatedSearchQueries

		messages = append(messages, schema.ToolMessage(toolOutput, toolName))
		assistantMsg, err = r.generate(ctx, messages, req)
		if err != nil {
			return nil, err
		}
		messages = append(messages, assistantMsg)
		toolName, toolInput, wantsTool = parseToolCall(assistantMsg.Content)
		maxRounds--
	}

	if wantsTool {
		return nil, fmt.Errorf("tool loop exceeded max rounds")
	}
	if assistantMsg.Content == "" {
		return nil, fmt.Errorf("empty response from AI")
	}

	return &AgentRunResponse{
		Output:        assistantMsg.Content,
		ToolsUsed:     toolsUsed,
		SearchQueries: searchQueries,
	}, nil
}
```

- [ ] **Step 8: Add the parsing helpers used by the loop**

Because the first version uses a simple single-string tool contract, add a deterministic JSON-like parser the tests can control:

```go
func buildToolMap(availableTools []tools.Tool) map[string]tools.Tool {
	toolMap := make(map[string]tools.Tool, len(availableTools))
	for _, tool := range availableTools {
		if tool == nil {
			continue
		}
		toolMap[tool.Name()] = tool
	}
	return toolMap
}
```

```go
type toolCallEnvelope struct {
	Tool  string `json:"tool"`
	Input string `json:"input"`
}

func parseToolCall(content string) (string, string, bool) {
	var envelope toolCallEnvelope
	if err := json.Unmarshal([]byte(content), &envelope); err != nil {
		return "", "", false
	}
	if strings.TrimSpace(envelope.Tool) == "" {
		return "", "", false
	}
	return strings.TrimSpace(envelope.Tool), strings.TrimSpace(envelope.Input), true
}
```

This keeps the initial runtime deterministic and testable. The system prompt can later instruct the model to emit this envelope when it wants to call a tool.

- [ ] **Step 9: Add the runtime loop test**

```go
func TestDefaultAgentRuntimeRun_ExecutesToolLoop(t *testing.T) {
	tool := &runtimeTestTool{name: "web_search", result: "query: movie\nresults:\n1. Title: Movie"}
	calls := 0
	runtime := &DefaultAgentRuntime{
		generate: func(ctx context.Context, messages []*schema.Message, req *AgentRunRequest) (*schema.Message, error) {
			calls++
			if calls == 1 {
				return schema.AssistantMessage(`{"tool":"web_search","input":"movie"}`), nil
			}
			return schema.AssistantMessage(`{"tags":["video","movie"],"description":"Movie metadata"}`), nil
		},
	}

	resp, err := runtime.Run(context.Background(), &AgentRunRequest{
		SystemPrompt: "sys",
		UserPrompt:   "user",
		Tools:        []tools.Tool{tool},
		AIConfig:     &models.AIConfig{APIKey: "key", Model: "model", BaseURL: "https://api.example.com/v1"},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if resp.Output != `{"tags":["video","movie"],"description":"Movie metadata"}` {
		t.Fatalf("unexpected output %q", resp.Output)
	}
	if len(resp.ToolsUsed) != 1 || resp.ToolsUsed[0] != "web_search" {
		t.Fatalf("expected web_search in ToolsUsed, got %v", resp.ToolsUsed)
	}
	if len(resp.SearchQueries) != 1 || resp.SearchQueries[0] != "movie" {
		t.Fatalf("expected movie in SearchQueries, got %v", resp.SearchQueries)
	}
}
```

- [ ] **Step 10: Run the targeted runtime tests**

Run: `go test ./app/agents -run 'Test(DefaultAgentRuntimeRun_WithoutToolsReturnsDirectOutput|DefaultAgentRuntimeRun_ExecutesToolLoop|ExecuteToolCallRecordsUsageAndSearchQuery)' -count=1`
Expected: PASS

- [ ] **Step 11: Commit**

```bash
git add app/agents/agent_runtime.go app/agents/agent_runtime_test.go
git commit -m "feat: add bounded metadata tool loop"
```

### Task 4: Wire the Real Tool into AgentService and Update Docs

**Files:**
- Modify: `app/services/agent_service.go`
- Modify: `app/services/agent_service_test.go`
- Modify: `docs/agent-integration.md`

**Interfaces:**
- Consumes: `func NewConfiguredWebSearchTool(config *models.AIConfig) *tools.WebSearchTool`, `func (s *AgentService) resolveMetadataTools(config *models.AIConfig, input *agents.MetadataGenerationInput) []tools.Tool`
- Produces:
  - `func (s *AgentService) ensureConfiguredTools(config *models.AIConfig)` (or equivalent lazy registration helper)
  - updated docs describing real runtime tool execution

- [ ] **Step 1: Write the failing service test for config-aware tool registration**

Add this test to `app/services/agent_service_test.go`:

```go
func TestResolveMetadataToolsUsesConfiguredWebSearchTool(t *testing.T) {
	service := &AgentService{toolRegistry: tools.NewToolRegistry()}
	config := &models.AIConfig{
		Enabled:             true,
		EnableAgent:         true,
		EnableWebSearch:     true,
		WebSearchProvider:   "mock-http",
		WebSearchBaseURL:    "https://search.example.com",
		WebSearchAPIKey:     "search-key",
		WebSearchTimeout:    10,
		WebSearchMaxResults: 3,
	}

	service.ensureConfiguredTools(config)
	resolved := service.resolveMetadataTools(config, &agents.MetadataGenerationInput{FileName: "MyAwesomeMovie.mp4", FileType: "video"})
	if len(resolved) != 1 {
		t.Fatalf("expected one resolved tool, got %d", len(resolved))
	}
	if resolved[0].Name() != "web_search" {
		t.Fatalf("expected web_search tool, got %q", resolved[0].Name())
	}
}
```

- [ ] **Step 2: Run the service test to verify it fails**

Run: `go test ./app/services -run TestResolveMetadataToolsUsesConfiguredWebSearchTool -count=1`
Expected: FAIL because `ensureConfiguredTools` does not exist yet.

- [ ] **Step 3: Stop registering a static tool in `NewAgentService`**

In `app/services/agent_service.go`, replace:

```go
webSearchTool := tools.NewWebSearchTool(10 * time.Second)
toolRegistry.Register(webSearchTool.Name(), webSearchTool)
```

with:

```go
service := &AgentService{
	metadataAgent: metadataAgent,
	configRepo:    configRepo,
	toolRegistry:  toolRegistry,
}

return service
```

- [ ] **Step 4: Add lazy config-aware registration**

In `app/services/agent_service.go`, add:

```go
func (s *AgentService) ensureConfiguredTools(config *models.AIConfig) {
	if s == nil || s.toolRegistry == nil {
		return
	}
	webSearchTool := tools.NewConfiguredWebSearchTool(config)
	_ = s.toolRegistry.Register(webSearchTool.Name(), webSearchTool)
}
```

Call it before resolving tools in `analyzeMetadataWithConfig`:

```go
s.ensureConfiguredTools(config)
availableTools := s.resolveMetadataTools(config, input)
```

- [ ] **Step 5: Run the service test to verify it passes**

Run: `go test ./app/services -run TestResolveMetadataToolsUsesConfiguredWebSearchTool -count=1`
Expected: PASS

- [ ] **Step 6: Update the constructor test**

Adjust `TestNewAgentService` so it no longer expects `web_search` to be pre-registered at construction time. Replace the old registry assertion with this:

```go
if len(service.toolRegistry.List()) != 0 {
	t.Fatalf("expected no eagerly registered tools, got %v", service.toolRegistry.List())
}
```

- [ ] **Step 7: Update `docs/agent-integration.md`**

Revise the sections that currently describe phase-one readiness only.

Replace text like this:

```md
- **WebSearchTool**: Exists as a trace-ready, service-gated capability for future expansion.
```

with:

```md
- **WebSearchTool**: A real service-gated capability that can execute bounded HTTP-backed search requests during metadata analysis.
```

Replace this runtime statement:

```md
- `DefaultAgentRuntime` does not yet execute a tool-calling loop.
```

with:

```md
- `DefaultAgentRuntime` executes a bounded metadata-focused tool loop when tools are available.
```

Update the execution flow bullets so they say the runtime may execute `web_search`, append tool output, and record `ToolsUsed` / `SearchQueries`.

- [ ] **Step 8: Run the targeted service tests**

Run: `go test ./app/services -run 'Test(NewAgentService|ResolveMetadataTools|ResolveMetadataToolsUsesConfiguredWebSearchTool)' -count=1`
Expected: PASS

- [ ] **Step 9: Commit**

```bash
git add app/services/agent_service.go app/services/agent_service_test.go docs/agent-integration.md
git commit -m "feat: wire configured web search into agent service"
```

### Task 5: Verify the End-to-End Agent-Facing Behavior

**Files:**
- Modify: `app/agents/metadata_agent_test.go`
- Modify: `app/services/agent_service_test.go`
- Optional Test Helper: add local fakes in the touched test files only if needed

**Interfaces:**
- Consumes: `EinoMetadataAgent.Analyze(ctx context.Context, input *MetadataGenerationInput, aiConfig *models.AIConfig, availableTools []tools.Tool) (*MetadataAnalysisResult, error)`, `AgentService.analyzeMetadataWithConfig(...)`
- Produces:
  - regression coverage that the agent-facing path distinguishes `ToolsAvailable` from `ToolsUsed`
  - regression coverage that fallback still preserves import-safe behavior when tool execution fails

- [ ] **Step 1: Write the failing metadata-agent trace test**

Add this test to `app/agents/metadata_agent_test.go`:

```go
func TestEinoMetadataAgentAnalyze_PropagatesRuntimeToolTrace(t *testing.T) {
	agent := NewEinoMetadataAgent(&fakeRuntime{
		response: &AgentRunResponse{
			Output:        `{"tags":["video","movie"],"description":"Movie metadata"}`,
			ToolsUsed:     []string{"web_search"},
			SearchQueries: []string{"movie"},
		},
	})

	result, err := agent.Analyze(
		context.Background(),
		&MetadataGenerationInput{FileName: "movie.mp4", FileType: "video"},
		&models.AIConfig{Enabled: true, EnableAgent: true},
		[]tools.Tool{&fakeTool{name: "web_search"}},
	)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}
	if len(result.Trace.ToolsAvailable) != 1 || result.Trace.ToolsAvailable[0] != "web_search" {
		t.Fatalf("expected ToolsAvailable to include web_search, got %v", result.Trace.ToolsAvailable)
	}
	if len(result.Trace.ToolsUsed) != 1 || result.Trace.ToolsUsed[0] != "web_search" {
		t.Fatalf("expected ToolsUsed to include web_search, got %v", result.Trace.ToolsUsed)
	}
	if len(result.Trace.SearchQueries) != 1 || result.Trace.SearchQueries[0] != "movie" {
		t.Fatalf("expected SearchQueries to include movie, got %v", result.Trace.SearchQueries)
	}
}
```

- [ ] **Step 2: Run the metadata-agent test**

Run: `go test ./app/agents -run TestEinoMetadataAgentAnalyze_PropagatesRuntimeToolTrace -count=1`
Expected: PASS or FAIL only if trace propagation regressed during runtime changes. If it already passes, keep it as regression coverage and continue.

- [ ] **Step 3: Write the failing fallback-on-tool-error service test**

Add a tool error fake to `app/services/agent_service_test.go`:

```go
type errorTool struct {
	name string
}

func (e *errorTool) Name() string { return e.name }
func (e *errorTool) Description() string { return "error tool" }
func (e *errorTool) Execute(ctx context.Context, input string) (string, error) {
	return "", errors.New("search backend unavailable")
}
```

Then add:

```go
func TestAgentServiceAnalyzeMetadata_FallbackIncludesResolvedToolsWhenRuntimeUsesFailingTool(t *testing.T) {
	service := &AgentService{
		metadataAgent: &fakeMetadataAgent{err: errors.New("execute tool web_search: search backend unavailable")},
		toolRegistry:  tools.NewToolRegistry(),
	}
	if err := service.toolRegistry.Register("web_search", &errorTool{name: "web_search"}); err != nil {
		t.Fatalf("register tool: %v", err)
	}

	result, err := service.analyzeMetadataWithConfig(
		context.Background(),
		&agents.MetadataGenerationInput{FileName: "MyAwesomeMovie.mp4", FileType: "video"},
		&models.AIConfig{Enabled: true, EnableAgent: true, EnableWebSearch: true},
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result.Trace.FallbackReason != "execute tool web_search: search backend unavailable" {
		t.Fatalf("unexpected fallback reason %q", result.Trace.FallbackReason)
	}
	if len(result.Trace.ToolsAvailable) != 1 || result.Trace.ToolsAvailable[0] != "web_search" {
		t.Fatalf("expected resolved tool captured in fallback trace, got %v", result.Trace.ToolsAvailable)
	}
}
```

- [ ] **Step 4: Run the targeted verification tests**

Run: `go test ./app/agents ./app/services -run 'Test(EinoMetadataAgentAnalyze_PropagatesRuntimeToolTrace|AgentServiceAnalyzeMetadata_FallbackIncludesResolvedToolsWhenRuntimeUsesFailingTool)' -count=1`
Expected: PASS

- [ ] **Step 5: Run the broader regression suites**

Run: `go test ./app/tools ./app/agents ./app/services ./app/models ./app/repositories -count=1`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add app/agents/metadata_agent_test.go app/services/agent_service_test.go
git commit -m "test: verify agent-facing web search behavior"
```

### Task 6: Final Documentation and Sanity Review

**Files:**
- Modify: `docs/superpowers/specs/2026-07-18-web-search-tool-design.md`
- Modify: `docs/superpowers/plans/2026-07-18-web-search-tool-implementation.md`

**Interfaces:**
- Consumes: all tasks above
- Produces: updated implementation notes if execution diverged, and a final checked plan record

- [ ] **Step 1: Re-read the spec and compare the implementation diff**

Run: `git diff --stat HEAD~5..HEAD`
Expected: shows focused changes in tools, runtime, config, service, tests, docs.

- [ ] **Step 2: Update the spec only if implementation intentionally diverged**

If the implementation differs from the spec in a deliberate way, add a short note under the relevant section such as:

```md
Implementation note: The first shipped runtime loop uses a small JSON tool-call envelope (`{"tool":"...","input":"..."}`) for deterministic testing before migrating to richer provider-native tool schemas.
```

If there is no intentional divergence, do not edit the spec.

- [ ] **Step 3: Check the plan file boxes or add execution notes**

At minimum, add an execution note near the top of this plan once implementation is complete:

```md
> Execution note: Implemented on 2026-07-18. All targeted Go test suites passed after Task 5.
```

- [ ] **Step 4: Run the final verification command**

Run: `go test ./app/tools ./app/agents ./app/services ./app/models ./app/repositories -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add docs/superpowers/specs/2026-07-18-web-search-tool-design.md docs/superpowers/plans/2026-07-18-web-search-tool-implementation.md
git commit -m "docs: finalize web search implementation notes"
```

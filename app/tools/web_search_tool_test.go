package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

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
	if !strings.Contains(result, `"query":"movie"`) {
		t.Fatalf("expected formatted query in output, got %q", result)
	}
	if strings.Contains(result, "Result 4") {
		t.Fatalf("expected output to exclude truncated result, got %q", result)
	}
}

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

func TestHTTPWebSearchClientSearchMapsResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/search" {
			t.Fatalf("expected path /search, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if got := body["query"]; got != "movie" {
			t.Fatalf("expected query movie, got %v", got)
		}
		if got := body["max_results"]; got != float64(3) {
			t.Fatalf("expected limit 3, got %v", got)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer search-key" {
			t.Fatalf("expected bearer auth, got %q", auth)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Fatalf("expected content type application/json, got %q", ct)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"title":"Movie Title","url":"https://example.com/movie","content":"Movie snippet"}]}`))
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
	if items[0].Snippet != "Movie snippet" {
		t.Fatalf("expected mapped snippet from content, got %q", items[0].Snippet)
	}
	if items[0].URL != "https://example.com/movie" {
		t.Fatalf("expected mapped URL, got %q", items[0].URL)
	}
}

package tools

import (
	"context"
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
	if !strings.Contains(result, "query: movie") {
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
		if got := r.URL.Query().Get("q"); got != "movie" {
			t.Fatalf("expected query movie, got %q", got)
		}
		if got := r.URL.Query().Get("limit"); got != "3" {
			t.Fatalf("expected limit 3, got %q", got)
		}
		if got := r.URL.Query().Get("provider"); got != "mock-http" {
			t.Fatalf("expected provider mock-http, got %q", got)
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

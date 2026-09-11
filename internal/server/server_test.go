package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/config"
	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/orchestrator"
	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/searxng"
	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/vault"
	"github.com/mark3labs/mcp-go/mcp"
)

func setupTestServer(t *testing.T) (*Server, *httptest.Server) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/search" {
			_, _ = w.Write([]byte(`{
				"results": [
					{"title": "Go Lang", "url": "https://go.dev", "content": "The Go language", "engine": "google"}
				]
			}`))
		} else {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<html><body><h1>Example Page</h1><p>Scraped content</p></body></html>"))
		}
	}))

	cfg := config.Config{
		SearXNGURL:        ts.URL,
		DefaultMaxResults: 10,
		MaxAllowedResults: 50,
		DefaultLanguage:   "auto",
		SearchTimeout:     2 * time.Second,
		CascadeTimeout:    2 * time.Second,
		RRFK:              60,
		AllowPrivateScrape: true,
	}

	searxClient := searxng.NewClient(ts.URL, 2*time.Second)
	v := vault.NewVault(t.TempDir())
	orc := orchestrator.NewOrchestrator(searxClient, v, nil, 60, true)

	srv := NewServer(searxClient, orc, cfg)
	return srv, ts
}

func TestHandleSearchWeb(t *testing.T) {
	srv, ts := setupTestServer(t)
	defer ts.Close()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "search_web",
			Arguments: map[string]interface{}{
				"query": "go test",
			},
		},
	}

	res, err := srv.handleSearchWeb(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.Content) == 0 {
		t.Fatalf("expected content in result")
	}

	textContent, ok := res.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", res.Content[0])
	}

	var searchRes searxng.SearchResult
	if err := json.Unmarshal([]byte(textContent.Text), &searchRes); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if searchRes.Count != 1 {
		t.Errorf("expected 1 result, got %d", searchRes.Count)
	}
	if searchRes.Results[0].Title != "Go Lang" {
		t.Errorf("expected title 'Go Lang', got '%s'", searchRes.Results[0].Title)
	}
}

func TestHandleFetchPage(t *testing.T) {
	srv, ts := setupTestServer(t)
	defer ts.Close()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "fetch_page",
			Arguments: map[string]interface{}{
				"url": ts.URL + "/page",
			},
		},
	}

	res, err := srv.handleFetchPage(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	textContent, ok := res.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", res.Content[0])
	}

	var scrapeRes struct {
		URL      string `json:"url"`
		Provider string `json:"provider"`
		Markdown string `json:"markdown"`
	}
	if err := json.Unmarshal([]byte(textContent.Text), &scrapeRes); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if scrapeRes.Markdown != "Example Page  Scraped content" {
		t.Errorf("unexpected markdown: '%s'", scrapeRes.Markdown)
	}
}

func TestHandleDeepResearch(t *testing.T) {
	srv, ts := setupTestServer(t)
	defer ts.Close()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "deep_research",
			Arguments: map[string]interface{}{
				"query": "quantum computing",
			},
		},
	}

	res, err := srv.handleDeepResearch(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	textContent, ok := res.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", res.Content[0])
	}

	var deepRes orchestrator.DeepResearchResult
	if err := json.Unmarshal([]byte(textContent.Text), &deepRes); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if deepRes.Count != 1 {
		t.Errorf("expected 1 result, got %d", deepRes.Count)
	}
}

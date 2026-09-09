package orchestrator

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/searxng"
	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/vault"
)

func TestOrchestratorDeepResearch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"results": [
				{"title": "SearXNG Result 1", "url": "https://example.com/1", "content": "Content 1", "engine": "google"}
			]
		}`))
	}))
	defer ts.Close()

	searxClient := searxng.NewClient(ts.URL, 2*time.Second)
	v := vault.NewVault(t.TempDir()) // Empty vault for unit test

	orc := NewOrchestrator(searxClient, v, nil, 60)
	res := orc.DeepResearch(context.Background(), "test query", 5)

	if res.Count != 1 {
		t.Fatalf("expected 1 result from SearXNG, got %d", res.Count)
	}
	if res.Results[0].Title != "SearXNG Result 1" {
		t.Errorf("unexpected title: %s", res.Results[0].Title)
	}
	if res.ProviderDeep != "none" {
		t.Errorf("expected provider_deep 'none', got '%s'", res.ProviderDeep)
	}
}

func TestOrchestratorScrapeNativeFallback(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body><h1>Article Header</h1><p>Full article body</p></body></html>"))
	}))
	defer ts.Close()

	searxClient := searxng.NewClient(ts.URL, 2*time.Second)
	v := vault.NewVault(t.TempDir()) // Empty vault forces fallback to native scraper

	orc := NewOrchestrator(searxClient, v, nil, 60)
	res := orc.ScrapePage(context.Background(), ts.URL)

	if res.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", res.StatusCode)
	}
	if res.Provider != "native (HTTP fallback)" {
		t.Errorf("expected provider 'native (HTTP fallback)', got '%s'", res.Provider)
	}
	if res.Markdown != "Article Header  Full article body" {
		t.Errorf("unexpected content: '%s'", res.Markdown)
	}
}

func TestOrchestratorScrapeTruncation(t *testing.T) {
	hugeBody := ""
	for i := 0; i < 4000; i++ {
		hugeBody += "0123456789 "
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(hugeBody))
	}))
	defer ts.Close()

	searxClient := searxng.NewClient(ts.URL, 2*time.Second)
	v := vault.NewVault(t.TempDir())

	orc := NewOrchestrator(searxClient, v, nil, 60)
	res := orc.ScrapePage(context.Background(), ts.URL)

	if len(res.Markdown) > MaxMarkdownLength+150 {
		t.Errorf("markdown exceeded max length: %d", len(res.Markdown))
	}
	if len(res.Markdown) <= MaxMarkdownLength {
		t.Errorf("expected markdown to have truncation footer, got length: %d", len(res.Markdown))
	}
}

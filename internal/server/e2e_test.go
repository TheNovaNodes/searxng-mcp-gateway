//go:build integration

package server

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/config"
	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/orchestrator"
	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/searxng"
	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/vault"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestLiveE2E(t *testing.T) {
	cfg := config.Load()

	searxClient := searxng.NewClient(cfg.SearXNGURL, 5*time.Second)
	v := vault.NewVault(cfg.VaultDir)
	orc := orchestrator.NewOrchestrator(searxClient, v, nil, 60)
	srv := NewServer(searxClient, orc, cfg)

	// 1. Test live search_web
	reqSearch := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "search_web",
			Arguments: map[string]interface{}{
				"query":       "open source",
				"max_results": 3,
			},
		},
	}
	resSearch, err := srv.handleSearchWeb(context.Background(), reqSearch)
	if err != nil {
		t.Fatalf("search_web failed: %v", err)
	}
	textSearch := resSearch.Content[0].(mcp.TextContent).Text
	var sResult searxng.SearchResult
	_ = json.Unmarshal([]byte(textSearch), &sResult)
	t.Logf("search_web: count=%d, latency=%.1fms", sResult.Count, sResult.LatencyMs)
	if sResult.Count == 0 {
		t.Errorf("expected live search_web to find results")
	}

	// 2. Test live deep_research (SearXNG + Exa AI RRF)
	reqDeep := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "deep_research",
			Arguments: map[string]interface{}{
				"query":       "golang 1.22",
				"max_results": 3,
			},
		},
	}
	resDeep, err := srv.handleDeepResearch(context.Background(), reqDeep)
	if err != nil {
		t.Fatalf("deep_research failed: %v", err)
	}
	textDeep := resDeep.Content[0].(mcp.TextContent).Text
	var dResult orchestrator.DeepResearchResult
	_ = json.Unmarshal([]byte(textDeep), &dResult)
	t.Logf("deep_research: provider_deep=%s, count=%d", dResult.ProviderDeep, dResult.Count)
	if dResult.Count == 0 {
		t.Errorf("expected deep_research to find results")
	}

	// 3. Test live fetch_page
	reqFetch := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "fetch_page",
			Arguments: map[string]interface{}{
				"url": "http://127.0.0.1:8889",
			},
		},
	}
	resFetch, err := srv.handleFetchPage(context.Background(), reqFetch)
	if err != nil {
		t.Fatalf("fetch_page failed: %v", err)
	}
	textFetch := resFetch.Content[0].(mcp.TextContent).Text
	var fResult struct {
		URL      string `json:"url"`
		Provider string `json:"provider"`
		Markdown string `json:"markdown"`
	}
	_ = json.Unmarshal([]byte(textFetch), &fResult)
	t.Logf("fetch_page: provider=%s, len=%d", fResult.Provider, len(fResult.Markdown))
	if len(fResult.Markdown) == 0 {
		t.Errorf("expected fetch_page to return markdown")
	}
}

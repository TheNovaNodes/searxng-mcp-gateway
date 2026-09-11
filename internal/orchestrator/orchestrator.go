package orchestrator

import (
	"context"
	"fmt"
	"time"

	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/echelon"
	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/fusion"
	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/searxng"
	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/vault"
	"golang.org/x/sync/errgroup"
)

// CascadeResult contains results from external search API cascade.
type CascadeResult struct {
	Provider string
	Results  []fusion.Item
	Error    string
}

// Orchestrator coordinates multi-echelon search and scraping.
type Orchestrator struct {
	searxClient   *searxng.Client
	vault         *vault.Vault
	cb            *echelon.CircuitBreaker
	balancers     map[string]*echelon.RoundRobinBalancer
	exaClient     *echelon.ExaClient
	tavilyClient  *echelon.TavilyClient
	firecrawl     *echelon.FirecrawlClient
	olostep       *echelon.OlostepClient
	nativeScraper      *echelon.NativeScraper
	rrfK               int
	allowPrivateScrape bool
}

// NewOrchestrator creates a new Orchestrator instance.
func NewOrchestrator(
	searxClient *searxng.Client,
	v *vault.Vault,
	cb *echelon.CircuitBreaker,
	rrfK int,
	allowPrivateScrape bool,
) *Orchestrator {
	if cb == nil {
		cb = echelon.NewCircuitBreaker(60 * time.Second)
	}
	if rrfK <= 0 {
		rrfK = 60
	}

	balancers := map[string]*echelon.RoundRobinBalancer{
		"exa":       echelon.NewRoundRobinBalancer(),
		"tavily":    echelon.NewRoundRobinBalancer(),
		"firecrawl": echelon.NewRoundRobinBalancer(),
		"olostep":   echelon.NewRoundRobinBalancer(),
	}

	return &Orchestrator{
		searxClient:        searxClient,
		vault:              v,
		cb:                 cb,
		balancers:          balancers,
		exaClient:          echelon.NewExaClient(6 * time.Second),
		tavilyClient:       echelon.NewTavilyClient(6 * time.Second),
		firecrawl:          echelon.NewFirecrawlClient(12 * time.Second),
		olostep:            echelon.NewOlostepClient(15 * time.Second),
		nativeScraper:      echelon.NewNativeScraper(8*time.Second, allowPrivateScrape),
		rrfK:               rrfK,
		allowPrivateScrape: allowPrivateScrape,
	}
}

// SearchCascade executes external API search with fallback (Exa -> Tavily).
func (o *Orchestrator) SearchCascade(ctx context.Context, query string, maxResults int) *CascadeResult {
	// 1. Try Exa AI (primary proven echelon)
	exaKeys := o.vault.GetKeys("exa")
	if selectedKey, ok := o.balancers["exa"].GetNext(exaKeys, o.cb); ok {
		items, err := o.exaClient.Search(ctx, selectedKey, query, maxResults, o.cb)
		if err == nil && len(items) > 0 {
			var fItems []fusion.Item
			for _, it := range items {
				fItems = append(fItems, fusion.Item{
					Title:   it.Title,
					URL:     it.URL,
					Content: it.Content,
					Engine:  "exa",
				})
			}
			return &CascadeResult{Provider: "exa", Results: fItems}
		}
	}

	// 2. Fallback to Tavily
	tavilyKeys := o.vault.GetKeys("tavily")
	if selectedKey, ok := o.balancers["tavily"].GetNext(tavilyKeys, o.cb); ok {
		items, err := o.tavilyClient.Search(ctx, selectedKey, query, maxResults, o.cb)
		if err == nil && len(items) > 0 {
			var fItems []fusion.Item
			for _, it := range items {
				fItems = append(fItems, fusion.Item{
					Title:   it.Title,
					URL:     it.URL,
					Content: it.Content,
					Engine:  "tavily",
				})
			}
			return &CascadeResult{Provider: "tavily", Results: fItems}
		}
	}

	return &CascadeResult{
		Provider: "none",
		Results:  nil,
		Error:    "All external deep search APIs are on cooldown or unavailable",
	}
}

// MaxMarkdownLength defines the maximum allowed characters in scraped markdown (approx. 8,500 tokens).
const MaxMarkdownLength = 35000

func truncateMarkdown(md string) string {
	if len(md) <= MaxMarkdownLength {
		return md
	}
	return md[:MaxMarkdownLength] + "\n\n... [Content truncated to 35,000 characters by gateway guardrail] ..."
}

// ScrapePage performs intelligent page scraping (Firecrawl -> Olostep WAF bypass -> Native).
func (o *Orchestrator) ScrapePage(ctx context.Context, targetURL string) *echelon.ScrapeResult {
	if _, err := echelon.ValidateTargetURL(targetURL, o.allowPrivateScrape); err != nil {
		return &echelon.ScrapeResult{
			URL:      targetURL,
			Provider: "none",
			Degraded: true,
			Error:    err.Error(),
		}
	}

	// 1. Try Firecrawl
	fcKeys := o.vault.GetKeys("firecrawl")
	if selectedKey, ok := o.balancers["firecrawl"].GetNext(fcKeys, o.cb); ok {
		res, err := o.firecrawl.Scrape(ctx, selectedKey, targetURL, o.cb)
		if err == nil && res != nil && res.Markdown != "" {
			res.Markdown = truncateMarkdown(res.Markdown)
			return res
		}
		// If WAF or error detected, proceed to Olostep
		if res != nil && echelon.DetectWAF(res.StatusCode, res.Error) {
			// WAF triggered, fall through to Olostep
		}
	}

	// 2. Fallback to Olostep (WAF Bypass)
	oloKeys := o.vault.GetKeys("olostep")
	if selectedKey, ok := o.balancers["olostep"].GetNext(oloKeys, o.cb); ok {
		res, err := o.olostep.Scrape(ctx, selectedKey, targetURL, o.cb)
		if err == nil && res != nil && res.Markdown != "" {
			res.Provider = "firecrawl -> olostep (WAF Bypassed)"
			res.Markdown = truncateMarkdown(res.Markdown)
			return res
		}
	}

	// 3. Fallback to Native Scraper
	res, err := o.nativeScraper.Scrape(ctx, targetURL)
	if err == nil && res != nil && res.Markdown != "" {
		res.Provider = "native (HTTP fallback)"
		res.Markdown = truncateMarkdown(res.Markdown)
		return res
	}

	if res != nil {
		return res
	}
	return &echelon.ScrapeResult{
		URL:      targetURL,
		Provider: "none",
		Degraded: true,
		Error:    fmt.Sprintf("failed to scrape URL: %v", err),
	}
}

// DeepResearchResult represents the combined RRF output of local SearXNG and external APIs.
type DeepResearchResult struct {
	Query        string        `json:"query"`
	ProviderDeep string        `json:"provider_deep"`
	Count        int           `json:"count"`
	Results      []fusion.Item `json:"results"`
	Degraded     bool          `json:"degraded,omitempty"`
}

// DeepResearch executes parallel search over SearXNG and external APIs and fuses with RRF.
func (o *Orchestrator) DeepResearch(ctx context.Context, query string, maxResults int) *DeepResearchResult {
	if maxResults <= 0 {
		maxResults = 10
	} else if maxResults > 50 {
		maxResults = 50
	}

	g, gCtx := errgroup.WithContext(ctx)

	var (
		searxItems []fusion.Item
		deepRes    *CascadeResult
	)

	// Thread 1: SearXNG search
	g.Go(func() error {
		sCtx, cancel := context.WithTimeout(gCtx, 5*time.Second)
		defer cancel()
		res, err := o.searxClient.Search(sCtx, searxng.SearchParams{
			Query:      query,
			MaxResults: maxResults,
		})
		if err == nil && res != nil {
			for _, r := range res.Results {
				eng := r.Engine
				searxItems = append(searxItems, fusion.Item{
					Title:   r.Title,
					URL:     r.URL,
					Content: r.Content,
					Engine:  eng,
				})
			}
		}
		return nil
	})

	// Thread 2: External cascade (Exa / Tavily)
	g.Go(func() error {
		dCtx, cancel := context.WithTimeout(gCtx, 6*time.Second)
		defer cancel()
		deepRes = o.SearchCascade(dCtx, query, maxResults)
		return nil
	})

	_ = g.Wait()

	var externalItems []fusion.Item
	providerDeep := "none"
	if deepRes != nil {
		externalItems = deepRes.Results
		providerDeep = deepRes.Provider
	}

	fused := fusion.ReciprocalRankFusion([][]fusion.Item{searxItems, externalItems}, o.rrfK)

	limit := maxResults
	if len(fused) < limit {
		limit = len(fused)
	}

	return &DeepResearchResult{
		Query:        query,
		ProviderDeep: providerDeep,
		Count:        limit,
		Results:      fused[:limit],
		Degraded:     len(searxItems) == 0 && len(externalItems) == 0,
	}
}

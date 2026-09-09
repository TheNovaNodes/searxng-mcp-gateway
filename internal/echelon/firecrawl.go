package echelon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// FirecrawlClient handles web scraping via Firecrawl.
type FirecrawlClient struct {
	httpClient *http.Client
}

// NewFirecrawlClient creates a Firecrawl client.
func NewFirecrawlClient(timeout time.Duration) *FirecrawlClient {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &FirecrawlClient{
		httpClient: &http.Client{Timeout: timeout},
	}
}

type firecrawlRequest struct {
	URL string `json:"url"`
}

type firecrawlResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Markdown string `json:"markdown"`
		Content  string `json:"content"`
		Metadata struct {
			Title      string `json:"title"`
			StatusCode int    `json:"statusCode"`
		} `json:"metadata"`
	} `json:"data"`
	Error string `json:"error"`
}

// Scrape calls Firecrawl scrape API.
func (c *FirecrawlClient) Scrape(ctx context.Context, apiKey, targetURL string, cb *CircuitBreaker) (*ScrapeResult, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("missing Firecrawl API key")
	}

	payload := firecrawlRequest{URL: targetURL}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.firecrawl.dev/v1/scrape", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("User-Agent", "TheNovaNodes-SearXNG-Gateway/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if cb != nil && cb.ShouldTrip(0, err) {
			cb.RecordFailure(apiKey)
		}
		return &ScrapeResult{
			URL:        targetURL,
			Provider:   "firecrawl",
			StatusCode: 0,
			Degraded:   true,
			Error:      err.Error(),
		}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if cb != nil && cb.ShouldTrip(resp.StatusCode, nil) {
			cb.RecordFailure(apiKey)
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return &ScrapeResult{
			URL:        targetURL,
			Provider:   "firecrawl",
			StatusCode: resp.StatusCode,
			Degraded:   true,
			Error:      fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
		}, fmt.Errorf("status %d", resp.StatusCode)
	}

	var res firecrawlResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return &ScrapeResult{
			URL:        targetURL,
			Provider:   "firecrawl",
			StatusCode: resp.StatusCode,
			Degraded:   true,
			Error:      err.Error(),
		}, err
	}

	md := res.Data.Markdown
	if md == "" {
		md = res.Data.Content
	}

	return &ScrapeResult{
		URL:        targetURL,
		Provider:   "firecrawl",
		Markdown:   strings.TrimSpace(md),
		Title:      res.Data.Metadata.Title,
		StatusCode: resp.StatusCode,
		Degraded:   false,
	}, nil
}

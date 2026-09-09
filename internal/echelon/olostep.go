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

// OlostepClient handles web scraping with automatic WAF bypass.
type OlostepClient struct {
	httpClient *http.Client
}

// NewOlostepClient creates an Olostep client.
func NewOlostepClient(timeout time.Duration) *OlostepClient {
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	return &OlostepClient{
		httpClient: &http.Client{Timeout: timeout},
	}
}

type olostepRequest struct {
	URL       string `json:"url"`
	BypassWAF bool   `json:"bypass_waf"`
}

type olostepResponse struct {
	Markdown string `json:"markdown"`
	Text     string `json:"text"`
	Title    string `json:"title"`
	Error    string `json:"error"`
}

// Scrape calls Olostep scrape endpoint with bypass_waf enabled.
func (c *OlostepClient) Scrape(ctx context.Context, apiKey, targetURL string, cb *CircuitBreaker) (*ScrapeResult, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("missing Olostep API key")
	}

	payload := olostepRequest{URL: targetURL, BypassWAF: true}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.olostep.com/v1/scrape", bytes.NewReader(jsonBody))
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
			Provider:   "olostep",
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
			Provider:   "olostep",
			StatusCode: resp.StatusCode,
			Degraded:   true,
			Error:      fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
		}, fmt.Errorf("status %d", resp.StatusCode)
	}

	var res olostepResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return &ScrapeResult{
			URL:        targetURL,
			Provider:   "olostep",
			StatusCode: resp.StatusCode,
			Degraded:   true,
			Error:      err.Error(),
		}, err
	}

	md := res.Markdown
	if md == "" {
		md = res.Text
	}

	return &ScrapeResult{
		URL:        targetURL,
		Provider:   "olostep",
		Markdown:   strings.TrimSpace(md),
		Title:      res.Title,
		StatusCode: resp.StatusCode,
		Degraded:   false,
	}, nil
}

// DetectWAF returns true if status code or error suggests WAF/Cloudflare blockage.
func DetectWAF(statusCode int, errStr string) bool {
	if statusCode == http.StatusForbidden || statusCode == http.StatusUnauthorized {
		return true
	}
	lower := strings.ToLower(errStr)
	return strings.Contains(lower, "cloudflare") ||
		strings.Contains(lower, "waf") ||
		strings.Contains(lower, "access denied") ||
		strings.Contains(lower, "captcha")
}

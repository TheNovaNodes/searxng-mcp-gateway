package searxng

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client interfaces with a SearXNG instance.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient initializes a SearXNG HTTP client.
func NewClient(baseURL string, timeout time.Duration) *Client {
	cleanURL := strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL: cleanURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// BaseURL returns the configured base URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// rawSearxResponse is the internal representation of SearXNG JSON output.
type rawSearxResponse struct {
	Results []struct {
		Title         string   `json:"title"`
		URL           string   `json:"url"`
		Content       string   `json:"content"`
		Engine        string   `json:"engine"`
		Engines       []string `json:"engines"`
		Score         float64  `json:"score"`
		Category      string   `json:"category"`
		PublishedDate string   `json:"publishedDate"`
	} `json:"results"`
	Answers             []string        `json:"answers"`
	Suggestions         []string        `json:"suggestions"`
	UnresponsiveEngines [][]interface{} `json:"unresponsive_engines"`
}

// Search queries SearXNG and normalizes the results.
func (c *Client) Search(ctx context.Context, params SearchParams) (*SearchResult, error) {
	if c.httpClient.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.httpClient.Timeout)
		defer cancel()
	}

	t0 := time.Now()

	limit := params.MaxResults
	if limit <= 0 {
		limit = 10
	} else if limit > 50 {
		limit = 50
	}

	searchEndpoint := fmt.Sprintf("%s/search", c.baseURL)
	reqURL, err := url.Parse(searchEndpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid search URL: %w", err)
	}

	q := reqURL.Query()
	q.Set("q", params.Query)
	q.Set("format", "json")
	if params.Language != "" {
		q.Set("language", params.Language)
	}
	q.Set("safesearch", strconv.Itoa(params.SafeSearch))
	if params.Categories != "" {
		q.Set("categories", params.Categories)
	}
	if params.Engines != "" {
		q.Set("engines", params.Engines)
	}
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "TheNovaNodes-SearXNG-Gateway/1.0")

	resp, err := c.httpClient.Do(req)
	latencyMs := math.Round(float64(time.Since(t0).Microseconds())/100.0) / 10.0

	if err != nil {
		sanitizedErr := fmt.Errorf("searxng request failed: connection error or timeout")
		return &SearchResult{
			Query:     params.Query,
			Count:     0,
			Results:   []ResultItem{},
			LatencyMs: latencyMs,
			Degraded:  true,
			Error:     sanitizedErr.Error(),
		}, sanitizedErr
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		err := fmt.Errorf("searxng responded with status %d: %s", resp.StatusCode, string(body))
		return &SearchResult{
			Query:     params.Query,
			Count:     0,
			Results:   []ResultItem{},
			LatencyMs: latencyMs,
			Degraded:  true,
			Error:     err.Error(),
		}, err
	}

	var raw rawSearxResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return &SearchResult{
			Query:     params.Query,
			Count:     0,
			Results:   []ResultItem{},
			LatencyMs: latencyMs,
			Degraded:  true,
			Error:     fmt.Sprintf("JSON decode error: %v", err),
		}, err
	}

	var results []ResultItem
	for i, r := range raw.Results {
		if i >= limit {
			break
		}
		item := ResultItem{
			Title:         r.Title,
			URL:           r.URL,
			Content:       r.Content,
			Engine:        r.Engine,
			Engines:       r.Engines,
			Score:         r.Score,
			Category:      r.Category,
			PublishedDate: r.PublishedDate,
		}
		results = append(results, item)
	}

	return &SearchResult{
		Query:               params.Query,
		Count:               len(results),
		Results:             results,
		Answers:             raw.Answers,
		Suggestions:         raw.Suggestions,
		UnresponsiveEngines: raw.UnresponsiveEngines,
		LatencyMs:           latencyMs,
	}, nil
}

// Health checks SearXNG connectivity and reports status.
func (c *Client) Health(ctx context.Context) *HealthInfo {
	if c.httpClient.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.httpClient.Timeout)
		defer cancel()
	}

	info := &HealthInfo{
		Status:     "unknown",
		SearXNGURL: c.baseURL,
		Reachable:  false,
	}

	t0 := time.Now()
	searchURL := fmt.Sprintf("%s/search?q=healthcheck&format=json", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		info.Status = "down"
		info.Error = "invalid request"
		return info
	}

	resp, err := c.httpClient.Do(req)
	info.LatencyMs = math.Round(float64(time.Since(t0).Microseconds())/100.0) / 10.0

	if err != nil {
		info.Status = "down"
		info.Error = "searxng health check failed: connection error or timeout"
		return info
	}
	defer resp.Body.Close()

	info.StatusCode = resp.StatusCode
	info.Reachable = resp.StatusCode == http.StatusOK

	if resp.StatusCode == http.StatusOK {
		var raw rawSearxResponse
		if err := json.NewDecoder(resp.Body).Decode(&raw); err == nil {
			info.ResultCount = len(raw.Results)
			info.UnresponsiveEngines = raw.UnresponsiveEngines
			info.Status = "ok"
		} else {
			info.Status = "degraded"
		}
	} else {
		info.Status = "degraded"
	}

	// Fetch version from /config if possible
	configURL := fmt.Sprintf("%s/config", c.baseURL)
	if cfgReq, err := http.NewRequestWithContext(ctx, http.MethodGet, configURL, nil); err == nil {
		if cfgResp, err := c.httpClient.Do(cfgReq); err == nil {
			defer cfgResp.Body.Close()
			if cfgResp.StatusCode == http.StatusOK {
				var cfgData struct {
					Version string `json:"version"`
				}
				if err := json.NewDecoder(cfgResp.Body).Decode(&cfgData); err == nil && cfgData.Version != "" {
					info.Version = cfgData.Version
				}
			}
		}
	}

	return info
}

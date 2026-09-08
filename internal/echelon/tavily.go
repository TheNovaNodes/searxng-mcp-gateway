package echelon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TavilyClient handles search requests to Tavily.
type TavilyClient struct {
	httpClient *http.Client
}

// NewTavilyClient creates a Tavily API client.
func NewTavilyClient(timeout time.Duration) *TavilyClient {
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	return &TavilyClient{
		httpClient: &http.Client{Timeout: timeout},
	}
}

type tavilyRequest struct {
	APIKey            string `json:"api_key"`
	Query             string `json:"query"`
	SearchDepth       string `json:"search_depth"`
	IncludeAnswer     bool   `json:"include_answer"`
	IncludeImages     bool   `json:"include_images"`
	IncludeRawContent bool   `json:"include_raw_content"`
	MaxResults        int    `json:"max_results,omitempty"`
}

type tavilyResponse struct {
	Results []struct {
		Title   string  `json:"title"`
		URL     string  `json:"url"`
		Content string  `json:"content"`
		Score   float64 `json:"score"`
	} `json:"results"`
	Answer string `json:"answer"`
	Error  string `json:"error"`
}

// Search calls Tavily search endpoint.
func (c *TavilyClient) Search(ctx context.Context, apiKey, query string, maxResults int, cb *CircuitBreaker) ([]EchelonItem, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("missing Tavily API key")
	}

	payload := tavilyRequest{
		APIKey:            apiKey,
		Query:             query,
		SearchDepth:       "basic",
		IncludeAnswer:     false,
		IncludeImages:     false,
		IncludeRawContent: false,
		MaxResults:        maxResults,
	}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.tavily.com/search", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "TheNovaNodes-SearXNG-Gateway/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if cb != nil && cb.ShouldTrip(0, err) {
			cb.RecordFailure(apiKey)
		}
		return nil, fmt.Errorf("tavily request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if cb != nil && cb.ShouldTrip(resp.StatusCode, nil) {
			cb.RecordFailure(apiKey)
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("tavily status %d: %s", resp.StatusCode, string(body))
	}

	var res tavilyResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("tavily decode error: %w", err)
	}

	var items []EchelonItem
	for _, r := range res.Results {
		items = append(items, EchelonItem{
			Title:   r.Title,
			URL:     r.URL,
			Content: r.Content,
			Engine:  "tavily",
			Score:   r.Score,
		})
	}
	return items, nil
}

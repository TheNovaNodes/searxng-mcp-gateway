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

// ExaClient handles search requests to Exa AI.
type ExaClient struct {
	httpClient *http.Client
}

// NewExaClient creates an Exa API client.
func NewExaClient(timeout time.Duration) *ExaClient {
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	return &ExaClient{
		httpClient: &http.Client{Timeout: timeout},
	}
}

type exaRequest struct {
	Query         string `json:"query"`
	UseAutoprompt bool   `json:"useAutoprompt"`
	NumResults    int    `json:"numResults,omitempty"`
}

type exaResponse struct {
	Results []struct {
		Title   string  `json:"title"`
		URL     string  `json:"url"`
		Text    string  `json:"text"`
		Snippet string  `json:"snippet"`
		Score   float64 `json:"score"`
	} `json:"results"`
	Error string `json:"error"`
}

// Search calls Exa AI search endpoint.
func (c *ExaClient) Search(ctx context.Context, apiKey, query string, maxResults int, cb *CircuitBreaker) ([]EchelonItem, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("missing Exa API key")
	}

	payload := exaRequest{
		Query:         query,
		UseAutoprompt: true,
		NumResults:    maxResults,
	}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.exa.ai/search", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("User-Agent", "TheNovaNodes-SearXNG-Gateway/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if cb != nil && cb.ShouldTrip(0, err) {
			cb.RecordFailure(apiKey)
		}
		return nil, fmt.Errorf("exa request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if cb != nil && cb.ShouldTrip(resp.StatusCode, nil) {
			cb.RecordFailure(apiKey)
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("exa status %d: %s", resp.StatusCode, string(body))
	}

	var res exaResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("exa decode error: %w", err)
	}

	var items []EchelonItem
	for _, r := range res.Results {
		content := r.Text
		if content == "" {
			content = r.Snippet
		}
		items = append(items, EchelonItem{
			Title:   r.Title,
			URL:     r.URL,
			Content: content,
			Engine:  "exa",
			Score:   r.Score,
		})
	}
	return items, nil
}

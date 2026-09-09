package echelon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewFirecrawlClient(t *testing.T) {
	c1 := NewFirecrawlClient(0)
	if c1.httpClient.Timeout != 15*time.Second {
		t.Errorf("expected 15s timeout, got %v", c1.httpClient.Timeout)
	}

	c2 := NewFirecrawlClient(5 * time.Second)
	if c2.httpClient.Timeout != 5*time.Second {
		t.Errorf("expected 5s timeout, got %v", c2.httpClient.Timeout)
	}
}

func TestFirecrawlScrape(t *testing.T) {
	tests := []struct {
		name         string
		apiKey       string
		handler      http.HandlerFunc
		timeout      bool
		wantErr      bool
		wantProvider string
		wantDegraded bool
	}{
		{
			name:    "missing api key",
			apiKey:  "",
			wantErr: true,
		},
		{
			name:   "successful response",
			apiKey: "test-key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success":true,"data":{"markdown":"Test markdown","content":"","metadata":{"title":"Test","statusCode":200}}}`))
			},
			wantErr:      false,
			wantProvider: "firecrawl",
			wantDegraded: false,
		},
		{
			name:   "api error 400",
			apiKey: "test-key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error":"bad request"}`))
			},
			wantErr:      true,
			wantProvider: "firecrawl",
			wantDegraded: true,
		},
		{
			name:   "api error 429",
			apiKey: "test-key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"rate limit"}`))
			},
			wantErr:      true,
			wantProvider: "firecrawl",
			wantDegraded: true,
		},
		{
			name:   "api error 500",
			apiKey: "test-key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"internal error"}`))
			},
			wantErr:      true,
			wantProvider: "firecrawl",
			wantDegraded: true,
		},
		{
			name:   "decode error",
			apiKey: "test-key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`invalid json`))
			},
			wantErr:      true,
			wantProvider: "firecrawl",
			wantDegraded: true,
		},
		{
			name:         "timeout error",
			apiKey:       "test-key",
			timeout:      true,
			wantErr:      true,
			wantProvider: "firecrawl",
			wantDegraded: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewFirecrawlClient(time.Second)

			client.httpClient.Transport = &roundTripperFunc{
				roundTrip: func(req *http.Request) (*http.Response, error) {
					if tt.timeout {
						return nil, context.DeadlineExceeded
					}
					req.URL.Scheme = "http"
					req.URL.Host = server.Listener.Addr().String()
					return http.DefaultTransport.RoundTrip(req)
				},
			}

			cb := NewCircuitBreaker(time.Second)
			res, err := client.Scrape(context.Background(), tt.apiKey, "http://test.com", cb)

			if (err != nil) != tt.wantErr {
				t.Errorf("Scrape() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if res.Provider != tt.wantProvider {
					t.Errorf("Scrape() provider = %v, want %v", res.Provider, tt.wantProvider)
				}
				if res.Degraded != tt.wantDegraded {
					t.Errorf("Scrape() degraded = %v, want %v", res.Degraded, tt.wantDegraded)
				}
			} else if res != nil && res.Provider != tt.wantProvider { // For error cases that return a Degraded ScrapeResult
				t.Errorf("Scrape() error provider = %v, want %v", res.Provider, tt.wantProvider)
			}
		})
	}
}

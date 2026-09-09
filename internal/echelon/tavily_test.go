package echelon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewTavilyClient(t *testing.T) {
	c1 := NewTavilyClient(0)
	if c1.httpClient.Timeout != 8*time.Second {
		t.Errorf("expected 8s timeout, got %v", c1.httpClient.Timeout)
	}

	c2 := NewTavilyClient(5 * time.Second)
	if c2.httpClient.Timeout != 5*time.Second {
		t.Errorf("expected 5s timeout, got %v", c2.httpClient.Timeout)
	}
}

func TestTavilySearch(t *testing.T) {
	tests := []struct {
		name       string
		apiKey     string
		handler    http.HandlerFunc
		timeout    bool
		wantErr    bool
		wantEngine string
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
				w.Write([]byte(`{"results":[{"title":"Test","url":"http://test.com","content":"Test content","score":0.9}]}`))
			},
			wantErr:    false,
			wantEngine: "tavily",
		},
		{
			name:   "api error 400",
			apiKey: "test-key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error":"bad request"}`))
			},
			wantErr: true,
		},
		{
			name:   "api error 429",
			apiKey: "test-key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"rate limit"}`))
			},
			wantErr: true,
		},
		{
			name:   "api error 500",
			apiKey: "test-key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"internal error"}`))
			},
			wantErr: true,
		},
		{
			name:   "decode error",
			apiKey: "test-key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`invalid json`))
			},
			wantErr: true,
		},
		{
			name:    "timeout error",
			apiKey:  "test-key",
			timeout: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewTavilyClient(time.Second)

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
			items, err := client.Search(context.Background(), tt.apiKey, "query", 1, cb)

			if (err != nil) != tt.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(items) > 0 && items[0].Engine != tt.wantEngine {
				t.Errorf("Search() engine = %v, want %v", items[0].Engine, tt.wantEngine)
			}
		})
	}
}

package echelon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewOlostepClient(t *testing.T) {
	c1 := NewOlostepClient(0)
	if c1.httpClient.Timeout != 20*time.Second {
		t.Errorf("expected 20s timeout, got %v", c1.httpClient.Timeout)
	}

	c2 := NewOlostepClient(5 * time.Second)
	if c2.httpClient.Timeout != 5*time.Second {
		t.Errorf("expected 5s timeout, got %v", c2.httpClient.Timeout)
	}
}

func TestOlostepScrape(t *testing.T) {
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
				w.Write([]byte(`{"markdown":"Test markdown","text":"","title":"Test"}`))
			},
			wantErr:      false,
			wantProvider: "olostep",
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
			wantProvider: "olostep",
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
			wantProvider: "olostep",
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
			wantProvider: "olostep",
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
			wantProvider: "olostep",
			wantDegraded: true,
		},
		{
			name:         "timeout error",
			apiKey:       "test-key",
			timeout:      true,
			wantErr:      true,
			wantProvider: "olostep",
			wantDegraded: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewOlostepClient(time.Second)

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

func TestDetectWAF(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		errStr     string
		want       bool
	}{
		{"403 Forbidden", http.StatusForbidden, "", true},
		{"401 Unauthorized", http.StatusUnauthorized, "", true},
		{"Cloudflare error string", 500, "Blocked by Cloudflare", true},
		{"WAF error string", 500, "WAF blockage", true},
		{"Access Denied string", 500, "access denied", true},
		{"Captcha string", 500, "please complete captcha", true},
		{"Normal 500", 500, "internal server error", false},
		{"Normal 200", 200, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectWAF(tt.statusCode, tt.errStr)
			if got != tt.want {
				t.Errorf("DetectWAF(%d, %q) = %v, want %v", tt.statusCode, tt.errStr, got, tt.want)
			}
		})
	}
}

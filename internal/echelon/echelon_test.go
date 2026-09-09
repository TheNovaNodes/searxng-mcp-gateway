package echelon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCircuitBreakerCooldown(t *testing.T) {
	cb := NewCircuitBreaker(100 * time.Millisecond)
	key := "test-key-1"

	if !cb.IsAvailable(key) {
		t.Errorf("expected key to be available initially")
	}

	cb.RecordFailure(key)
	if cb.IsAvailable(key) {
		t.Errorf("expected key to be on cooldown after failure")
	}

	time.Sleep(150 * time.Millisecond)
	if !cb.IsAvailable(key) {
		t.Errorf("expected key to recover after cooldown expired")
	}
}

func TestCircuitBreakerShouldTrip(t *testing.T) {
	cb := NewCircuitBreaker(time.Minute)

	if !cb.ShouldTrip(429, nil) {
		t.Errorf("expected 429 to trip")
	}
	if !cb.ShouldTrip(401, nil) {
		t.Errorf("expected 401 to trip")
	}
	if !cb.ShouldTrip(502, nil) {
		t.Errorf("expected 502 to trip")
	}
	if cb.ShouldTrip(200, nil) {
		t.Errorf("expected 200 NOT to trip")
	}
	if !cb.ShouldTrip(0, context.DeadlineExceeded) {
		t.Errorf("expected deadline exceeded to trip")
	}
}

func TestRoundRobinBalancer(t *testing.T) {
	b := NewRoundRobinBalancer()
	cb := NewCircuitBreaker(time.Minute)
	keys := []string{"k1", "k2", "k3"}

	k, ok := b.GetNext(keys, cb)
	if !ok || k != "k1" {
		t.Errorf("expected k1, got %s", k)
	}

	k, ok = b.GetNext(keys, cb)
	if !ok || k != "k2" {
		t.Errorf("expected k2, got %s", k)
	}

	// k3 fails
	cb.RecordFailure("k3")

	k, ok = b.GetNext(keys, cb)
	// k3 is on cooldown, should skip to k1
	if !ok || k != "k1" {
		t.Errorf("expected k1 (skipping k3), got %s", k)
	}
}

func TestNativeScraper(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><head><script>bad();</script></head><body><h1>Hello World</h1><p>Sample text</p></body></html>"))
	}))
	defer ts.Close()

	scraper := NewNativeScraper(2 * time.Second)
	res, err := scraper.Scrape(context.Background(), ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", res.StatusCode)
	}
	if res.Markdown != "Hello World  Sample text" {
		t.Errorf("unexpected content: '%s'", res.Markdown)
	}
}

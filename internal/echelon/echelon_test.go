package echelon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"net"
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

	scraper := NewNativeScraper(2 * time.Second, true)
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

func TestRoundRobinBalancerEmpty(t *testing.T) {
	b := NewRoundRobinBalancer()
	cb := NewCircuitBreaker(time.Minute)
	keys := []string{}

	k, ok := b.GetNext(keys, cb)
	if ok || k != "" {
		t.Errorf("expected empty string and false, got %s, %v", k, ok)
	}
}

func TestCircuitBreakerShouldTripNetError(t *testing.T) {
	cb := NewCircuitBreaker(time.Minute)

	// Context DeadlineExceeded is tested above.
	// Net timeout
	if !cb.ShouldTrip(0, &netErrorMock{timeout: true}) {
		t.Errorf("expected net.Error with Timeout()=true to trip")
	}

	if cb.ShouldTrip(0, &netErrorMock{timeout: false}) {
		t.Errorf("expected net.Error with Timeout()=false NOT to trip, unless it's another handled type")
	}

	// OpError
	if !cb.ShouldTrip(0, &net.OpError{}) {
		t.Errorf("expected *net.OpError to trip")
	}
}

type netErrorMock struct {
	timeout bool
}

func (e *netErrorMock) Error() string   { return "netErrorMock" }
func (e *netErrorMock) Timeout() bool   { return e.timeout }
func (e *netErrorMock) Temporary() bool { return false }

func TestRoundRobinBalancerSelection(t *testing.T) {
	b := NewRoundRobinBalancer()
	cb := NewCircuitBreaker(time.Minute)
	keys := []string{"k1", "k2", "k3"}

	// Sequence: k1, k2, k3, k1...
	for i, want := range []string{"k1", "k2", "k3", "k1"} {
		k, ok := b.GetNext(keys, cb)
		if !ok || k != want {
			t.Errorf("step %d: expected %s, got %s", i, want, k)
		}
	}

	// k2 fails
	cb.RecordFailure("k2")

	// Next should be k2, but it's failed, so skip to k3
	k, ok := b.GetNext(keys, cb)
	if !ok || k != "k3" {
		t.Errorf("expected k3 (skipping k2), got %s", k)
	}

	// All fail
	cb.RecordFailure("k1")
	cb.RecordFailure("k3")
	k, ok = b.GetNext(keys, cb)
	if ok || k != "" {
		t.Errorf("expected empty string and false, got %s, %v", k, ok)
	}
}

func TestCircuitBreakerStateTransitions(t *testing.T) {
	cb := NewCircuitBreaker(100 * time.Millisecond)
	key := "test-key-cb"

	// Closed -> Open (on failure)
	if !cb.IsAvailable(key) {
		t.Errorf("expected key to be available initially (Closed state)")
	}
	cb.RecordFailure(key)
	if cb.IsAvailable(key) {
		t.Errorf("expected key to be unavailable after failure (Open state)")
	}

	// Open -> Half-Open/Closed (after timeout)
	time.Sleep(150 * time.Millisecond)
	if !cb.IsAvailable(key) {
		t.Errorf("expected key to be available after timeout (Closed/Half-Open state)")
	}
}

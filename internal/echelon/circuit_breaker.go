package echelon

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"
)

// CircuitBreaker tracks API key failures and enforces cooldown periods.
type CircuitBreaker struct {
	mu              sync.RWMutex
	cooldowns       map[string]time.Time
	cooldownSeconds time.Duration
}

// NewCircuitBreaker creates a circuit breaker with the given cooldown duration.
func NewCircuitBreaker(cooldown time.Duration) *CircuitBreaker {
	if cooldown <= 0 {
		cooldown = 60 * time.Second
	}
	return &CircuitBreaker{
		cooldowns:       make(map[string]time.Time),
		cooldownSeconds: cooldown,
	}
}

// RecordFailure marks a key as failing at current time.
func (cb *CircuitBreaker) RecordFailure(key string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.cooldowns[key] = time.Now()
}

// IsAvailable checks if the key is ready for use (not in cooldown).
func (cb *CircuitBreaker) IsAvailable(key string) bool {
	cb.mu.RLock()
	lastFailure, exists := cb.cooldowns[key]
	cb.mu.RUnlock()

	if !exists {
		return true
	}

	if time.Since(lastFailure) > cb.cooldownSeconds {
		cb.mu.Lock()
		delete(cb.cooldowns, key)
		cb.mu.Unlock()
		return true
	}
	return false
}

// ShouldTrip determines if an HTTP status code or error warrants tripping the circuit breaker.
func (cb *CircuitBreaker) ShouldTrip(statusCode int, err error) bool {
	// Quota exhaustion or authorization / rate limit
	if statusCode == http.StatusUnauthorized ||
		statusCode == http.StatusPaymentRequired ||
		statusCode == http.StatusForbidden ||
		statusCode == http.StatusTooManyRequests ||
		statusCode == http.StatusBadGateway ||
		statusCode == http.StatusServiceUnavailable ||
		statusCode == http.StatusGatewayTimeout {
		return true
	}

	if err != nil {
		// Context deadline exceeded
		if errors.Is(err, context.DeadlineExceeded) {
			return true
		}
		// Network timeout
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return true
		}
		// Connection refused or reset
		var opErr *net.OpError
		if errors.As(err, &opErr) {
			return true
		}
	}
	return false
}

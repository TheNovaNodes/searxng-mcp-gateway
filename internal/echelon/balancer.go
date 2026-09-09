package echelon

import (
	"sync"
)

// RoundRobinBalancer selects keys in a round-robin order respecting circuit breaker state.
type RoundRobinBalancer struct {
	mu  sync.Mutex
	idx int
}

// NewRoundRobinBalancer creates a new balancer.
func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{idx: 0}
}

// GetNext picks the next healthy key from the list.
func (b *RoundRobinBalancer) GetNext(keys []string, cb *CircuitBreaker) (string, bool) {
	if len(keys) == 0 {
		return "", false
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	n := len(keys)
	for i := 0; i < n; i++ {
		key := keys[(b.idx+i)%n]
		if cb == nil || cb.IsAvailable(key) {
			b.idx = (b.idx + i + 1) % n
			return key, true
		}
	}
	return "", false
}

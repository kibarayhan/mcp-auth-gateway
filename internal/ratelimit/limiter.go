package ratelimit

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

// Limiter manages per-user per-server rate limits.
type Limiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
}

// New creates a new rate limiter.
func New() *Limiter {
	return &Limiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

// Configure sets up a rate limit for a user+server combination.
func (l *Limiter) Configure(user, server string, ratePerSecond float64) {
	key := user + ":" + server
	burst := int(ratePerSecond) + 1
	if burst < 1 {
		burst = 1
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.limiters[key] = rate.NewLimiter(rate.Limit(ratePerSecond), burst)
}

// Allow checks if a request is within the rate limit.
// Returns true if allowed, false if rate limited.
// If no limiter is configured for this user+server, always allows.
func (l *Limiter) Allow(user, server string) bool {
	key := user + ":" + server
	l.mu.Lock()
	lim, ok := l.limiters[key]
	l.mu.Unlock()

	if !ok {
		return true
	}
	return lim.Allow()
}

// ParseRate parses a rate string like "100/hour", "10/minute", "5/second"
// into requests per second.
func ParseRate(s string) (float64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty rate string")
	}

	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid rate format %q (expected N/unit)", s)
	}

	count, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid rate count %q: %w", parts[0], err)
	}

	switch strings.ToLower(parts[1]) {
	case "second", "sec", "s":
		return count, nil
	case "minute", "min", "m":
		return count / 60, nil
	case "hour", "hr", "h":
		return count / 3600, nil
	default:
		return 0, fmt.Errorf("unknown rate unit %q (use second/minute/hour)", parts[1])
	}
}

package middleware

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// RateLimiter implements a sliding window rate limiter
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	// Start cleanup goroutine
	go rl.cleanup()
	return rl
}

// cleanup removes old entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, times := range rl.requests {
			// Remove entries older than the window
			var valid []time.Time
			for _, t := range times {
				if now.Sub(t) < rl.window {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = valid
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request from the given key should be allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Get existing requests for this key
	times := rl.requests[key]

	// Filter to only requests within the window
	var valid []time.Time
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	// Check if under limit
	if len(valid) >= rl.limit {
		rl.requests[key] = valid
		return false
	}

	// Add this request
	valid = append(valid, now)
	rl.requests[key] = valid
	return true
}

// Remaining returns the number of requests remaining for a key
func (rl *RateLimiter) Remaining(key string) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	times := rl.requests[key]
	count := 0
	for _, t := range times {
		if t.After(cutoff) {
			count++
		}
	}

	remaining := rl.limit - count
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

// RateLimitMiddleware creates rate limiting middleware
type RateLimitMiddleware struct {
	general *RateLimiter
	auth    *RateLimiter
	public  *RateLimiter
}

// NewRateLimitMiddleware creates new rate limiting middleware with configurable limits
func NewRateLimitMiddleware() *RateLimitMiddleware {
	// Get limits from environment or use defaults
	generalLimit := 60 // requests per minute
	if limit := os.Getenv("RATE_LIMIT_REQUESTS_PER_MINUTE"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			generalLimit = l
		}
	}

	authLimit := 10 // auth requests per minute
	if limit := os.Getenv("RATE_LIMIT_AUTH_REQUESTS_PER_MINUTE"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			authLimit = l
		}
	}

	return &RateLimitMiddleware{
		general: NewRateLimiter(generalLimit, time.Minute),
		auth:    NewRateLimiter(authLimit, time.Minute),
		public:  NewRateLimiter(30, time.Minute), // 30 public form submissions per minute
	}
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		if idx := len(xff); idx > 0 {
			for i, c := range xff {
				if c == ',' {
					return xff[:i]
				}
			}
			return xff
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// General returns middleware for general rate limiting
func (m *RateLimitMiddleware) General(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := getClientIP(r)

		if !m.general.Allow(key) {
			w.Header().Set("Retry-After", "60")
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(m.general.limit))
			w.Header().Set("X-RateLimit-Remaining", "0")
			http.Error(w, `{"error": "rate_limit_exceeded", "message": "Too many requests, please try again later"}`, http.StatusTooManyRequests)
			return
		}

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(m.general.limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(m.general.Remaining(key)))
		next.ServeHTTP(w, r)
	})
}

// Auth returns middleware for authentication rate limiting
func (m *RateLimitMiddleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := getClientIP(r)

		if !m.auth.Allow(key) {
			w.Header().Set("Retry-After", "60")
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(m.auth.limit))
			w.Header().Set("X-RateLimit-Remaining", "0")
			http.Error(w, `{"error": "rate_limit_exceeded", "message": "Too many authentication attempts, please try again later"}`, http.StatusTooManyRequests)
			return
		}

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(m.auth.limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(m.auth.Remaining(key)))
		next.ServeHTTP(w, r)
	})
}

// Public returns middleware for public endpoint rate limiting
func (m *RateLimitMiddleware) Public(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := getClientIP(r)

		if !m.public.Allow(key) {
			w.Header().Set("Retry-After", "60")
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(m.public.limit))
			w.Header().Set("X-RateLimit-Remaining", "0")
			http.Error(w, `{"error": "rate_limit_exceeded", "message": "Too many requests to public endpoint, please try again later"}`, http.StatusTooManyRequests)
			return
		}

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(m.public.limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(m.public.Remaining(key)))
		next.ServeHTTP(w, r)
	})
}

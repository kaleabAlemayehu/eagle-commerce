package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiterConfig struct {
	RequestsPerMinute int
	Burst             int
	CleanupInterval   time.Duration
}

func NewRateLimit(config RateLimiterConfig) func(http.Handler) http.Handler {
	limiter := &rateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rate:     rate.Every(time.Minute / time.Duration(config.RequestsPerMinute)),
		burst:    config.Burst,
	}

	// Start cleanup goroutine
	go limiter.cleanup(config.CleanupInterval)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getIP(r)

			if !limiter.allow(ip) {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type rateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	if _, exists := rl.visitors[ip]; !exists {
		rl.visitors[ip] = rate.NewLimiter(rl.rate, rl.burst)
	}
	visitor := rl.visitors[ip]
	rl.mu.Unlock()

	return visitor.Allow()
}

func (rl *rateLimiter) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		if len(rl.visitors) > 1000 {
			rl.visitors = make(map[string]*rate.Limiter)
		}
		rl.mu.Unlock()
	}
}

func getIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	return r.RemoteAddr
}

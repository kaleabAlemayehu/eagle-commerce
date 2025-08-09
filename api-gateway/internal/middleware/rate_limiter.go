package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiterConfig struct {
	RequestsPerMinute int
	Burst             int
	CleanupInterval   time.Duration
	Timeout           time.Duration
}

func NewRateLimit(config RateLimiterConfig) func(http.Handler) http.Handler {
	limiter := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate.Every(time.Minute / time.Duration(config.RequestsPerMinute)),
		burst:    config.Burst,
	}

	// Start cleanup goroutine
	go limiter.cleanup(config.CleanupInterval, config.Timeout)

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
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter.Allow()
	}
	v.lastSeen = time.Now()
	return v.limiter.Allow()
}

func (rl *rateLimiter) cleanup(interval, timeout time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > timeout {
				delete(rl.visitors, ip)
			}
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

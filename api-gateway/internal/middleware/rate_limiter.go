// api-gateway/internal/middleware/rate_limit.go
package middleware

import (
	"net/http"
	"sync"
	"time"
)

type rateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
}

type visitor struct {
	limiter  *time.Ticker
	lastSeen time.Time
}

var limiter = &rateLimiter{
	visitors: make(map[string]*visitor),
}

func RateLimit() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getIP(r)

			limiter.mu.Lock()
			v, exists := limiter.visitors[ip]
			if !exists {
				v = &visitor{
					limiter:  time.NewTicker(time.Minute / 100), // 100 requests per minute
					lastSeen: time.Now(),
				}
				limiter.visitors[ip] = v
			}
			v.lastSeen = time.Now()
			limiter.mu.Unlock()

			select {
			case <-v.limiter.C:
				next.ServeHTTP(w, r)
			default:
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			}
		})
	}
}

func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

// Cleanup old visitors
func init() {
	go func() {
		for {
			time.Sleep(time.Minute)
			limiter.mu.Lock()
			for ip, v := range limiter.visitors {
				if time.Since(v.lastSeen) > 3*time.Minute {
					v.limiter.Stop()
					delete(limiter.visitors, ip)
				}
			}
			limiter.mu.Unlock()
		}
	}()
}

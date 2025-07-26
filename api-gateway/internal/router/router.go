package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/kaleabAlemayehu/eagle-commerce/api-gateway/internal/handler"
	gatewayMiddleware "github.com/kaleabAlemayehu/eagle-commerce/api-gateway/internal/middleware"
)

func NewRouter(proxyHandler *handler.ProxyHandler) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(gatewayMiddleware.CORS())

	r.Use(gatewayMiddleware.NewRateLimit(gatewayMiddleware.RateLimiterConfig{
		RequestsPerMinute: 100,
		Burst:             10,
		CleanupInterval:   5 * time.Minute,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("API Gateway is healthy"))
	})

	// Service routing
	r.Route("/api/v1", func(r chi.Router) {
		// User service routes
		r.Route("/users", func(r chi.Router) {
			r.HandleFunc("/*", proxyHandler.ProxyRequest("user"))
		})

		// Product service routes
		r.Route("/products", func(r chi.Router) {
			r.Use(gatewayMiddleware.Auth())
			r.HandleFunc("/*", proxyHandler.ProxyRequest("product"))
		})

		// Order service routes (protected)
		r.Route("/orders", func(r chi.Router) {
			r.Use(gatewayMiddleware.Auth())
			r.HandleFunc("/*", proxyHandler.ProxyRequest("order"))
		})

		// Payment service routes (protected)
		r.Route("/payments", func(r chi.Router) {
			r.Use(gatewayMiddleware.Auth())
			r.HandleFunc("/*", proxyHandler.ProxyRequest("payment"))
		})
	})

	return r
}

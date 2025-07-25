package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/interfaces/http/handler"
	sharedMiddleware "github.com/kaleabAlemayehu/eagle-commerce/shared/middleware"
)

// TODO: there is a lot of missing handlers

func NewRouter(orderHandler *handler.OrderHandler) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(sharedMiddleware.LoggingMiddleware())
	r.Use(middleware.Heartbeat("/health"))

	// Swagger
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/orders", func(r chi.Router) {
			// Public routes (with auth)
			r.Group(func(r chi.Router) {
				r.Use(sharedMiddleware.AuthMiddleware())

				r.Post("/", orderHandler.CreateOrder)
				r.Get("/", orderHandler.ListOrders)
				r.Get("/{id}", orderHandler.GetOrder)
				r.Put("/{id}/status", orderHandler.UpdateOrderStatus)
				r.Put("/{id}/cancel", orderHandler.CancelOrder)

				// User-specific routes
				r.Get("/user/{user_id}", orderHandler.GetUserOrders)
				// r.Get("/{id}/tracking", orderHandler.GetOrderTracking)
			})

			// Admin routes (additional auth needed)
			r.Group(func(r chi.Router) {
				r.Use(sharedMiddleware.AuthMiddleware())
				// Add admin-only middleware here

				// r.Get("/admin/summary", orderHandler.GetOrderSummary)
				// r.Put("/admin/{id}/ship", orderHandler.ShipOrder)
			})
		})
	})

	return r
}

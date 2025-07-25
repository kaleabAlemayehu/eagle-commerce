package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/interfaces/http/handler"
	sharedMiddleware "github.com/kaleabAlemayehu/eagle-commerce/shared/middleware"
)

func NewRouter(paymentHandler *handler.PaymentHandler) *chi.Mux {
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
		r.Route("/payments", func(r chi.Router) {
			// Protected routes
			r.Group(func(r chi.Router) {
				r.Use(sharedMiddleware.AuthMiddleware())

				r.Post("/", paymentHandler.ProcessPayment)
				r.Get("/", paymentHandler.ListPayments)
				r.Get("/{id}", paymentHandler.GetPayment)
				r.Get("/order/{order_id}", paymentHandler.GetPaymentByOrder)
				r.Post("/{id}/refund", paymentHandler.RefundPayment)
			})

			// Webhook routes (no auth required)
			r.Group(func(r chi.Router) {
				r.Post("/webhook", paymentHandler.HandleWebhook)
			})
		})
	})

	return r
}

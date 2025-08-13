package router

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/interfaces/http/handler"
	sharedMiddleware "github.com/kaleabAlemayehu/eagle-commerce/shared/middleware"
)

func NewRouter(productHandler *handler.ProductHandler, logger *slog.Logger, mode string) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	if mode == "production" {
		r.Use(sharedMiddleware.SlogMiddleware(logger))
	} else {
		r.Use(sharedMiddleware.LoggingMiddleware())
	}

	r.Use(middleware.Heartbeat("/health"))

	// Swagger
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/products", func(r chi.Router) {
			r.Post("/", productHandler.CreateProduct)
			r.Get("/", productHandler.ListProducts)
			r.Get("/search", productHandler.SearchProducts)
			r.Post("/check-stock", productHandler.CheckStock)
			r.Post("/reserve-stock", productHandler.ReserveStock)
			r.Get("/{id}", productHandler.GetProduct)
			r.Put("/{id}", productHandler.UpdateProduct)
			r.Delete("/{id}", productHandler.DeleteProduct)
		})
	})

	return r
}

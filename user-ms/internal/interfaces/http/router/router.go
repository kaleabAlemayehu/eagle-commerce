package router

import (
	"github.com/go-chi/chi/v5"
	mw "github.com/go-chi/chi/v5/middleware"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/kaleabAlemayehu/eagle-commerce/services/user-ms/internal/interfaces/http/handler"
	sharedMiddleware "github.com/kaleabAlemayehu/eagle-commerce/shared/middleware"
)

func NewRouter(userHandler *handler.UserHandler, auth *sharedMiddleware.Auth) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(mw.Logger)
	r.Use(mw.Recoverer)
	r.Use(mw.RequestID)
	r.Use(mw.Heartbeat("/health"))

	// Swagger
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/users", func(r chi.Router) {
			r.Post("/login", userHandler.LoginUser)
			r.Post("/signup", userHandler.RegisterUser)
			// Protected routes (with auth)
			r.Group(func(r chi.Router) {
				r.Use(auth.AuthMiddleware())
				r.Get("/", userHandler.ListUsers)
				r.Get("/{id}", userHandler.GetUser)
				r.Put("/{id}", userHandler.UpdateUser)
				r.Delete("/{id}", userHandler.DeleteUser)
			})
		})
	})

	return r
}

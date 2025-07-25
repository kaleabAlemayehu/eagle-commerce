// services/user-ms/internal/interfaces/http/middleware/auth_middleware.go
package middleware

import (
	"net/http"

	sharedMiddleware "github.com/kaleabAlemayehu/eagle-commerce/shared/middleware"
)

func AuthMiddleware() func(http.Handler) http.Handler {
	return sharedMiddleware.AuthMiddleware()
}

func OptionalAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				// If auth header is present, validate it
				authMiddleware := sharedMiddleware.AuthMiddleware()
				authMiddleware(next).ServeHTTP(w, r)
			} else {
				// If no auth header, proceed without authentication
				next.ServeHTTP(w, r)
			}
		})
	}
}

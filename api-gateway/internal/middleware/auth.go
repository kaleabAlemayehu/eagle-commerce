package middleware

import (
	"net/http"

	sharedMiddleware "github.com/kaleabAlemayehu/eagle-commerce/shared/middleware"
)

func Auth() func(http.Handler) http.Handler {
	return sharedMiddleware.AuthMiddleware()
}

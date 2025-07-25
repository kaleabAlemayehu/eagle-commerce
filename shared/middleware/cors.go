package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

func CORS() func(http.Handler) http.Handler {
	return middleware.SetHeader("Access-Control-Allow-Origin", "*")
}

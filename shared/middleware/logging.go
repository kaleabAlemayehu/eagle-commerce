package middleware

import (
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	bytes      int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	bytes, err := rw.ResponseWriter.Write(b)
	rw.bytes += bytes
	return bytes, err
}

func LoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Get request ID from context
			requestID := middleware.GetReqID(r.Context())

			defer func() {
				duration := time.Since(start)
				log.Printf(
					"[%s] %s %s %d %d bytes %v",
					requestID,
					r.Method,
					r.URL.Path,
					rw.statusCode,
					rw.bytes,
					duration,
				)
			}()

			next.ServeHTTP(rw, r)
		})
	}
}

func SlogMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := middleware.GetReqID(r.Context())

			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			// Log the request details using the structured logger
			logger.Info("incoming request",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.statusCode,
				"duration_ms", duration.Milliseconds(),
				"bytes_written", rw.bytes,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

package middleware

import (
	"github.com/go-chi/chi/v5/middleware"
	sharedLogger "github.com/kaleabAlemayehu/eagle-commerce/shared/logger"
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

func SlogMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := middleware.GetReqID(r.Context())
			reqLogger := logger.With(
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("user_agent", r.UserAgent()),
			)

			ctx := sharedLogger.WithLogger(r.Context(), reqLogger)

			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rw, r.WithContext(ctx))

			duration := time.Since(start)

			reqLogger.Info("incoming request",
				slog.Int("status", rw.statusCode),
				slog.Int64("duration_ms", duration.Milliseconds()),
				slog.Int("bytes_written", rw.bytes))
		})
	}
}

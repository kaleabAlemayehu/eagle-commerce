package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
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

func StructuredLogMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rw, r)

			log.Printf(`{
                "timestamp": "%s",
                "method": "%s",
                "path": "%s",
                "status": %d,
                "bytes": %d,
                "duration": "%v",
                "user_agent": "%s",
                "remote_addr": "%s"
            }`,
				start.Format(time.RFC3339),
				r.Method,
				r.URL.Path,
				rw.statusCode,
				rw.bytes,
				time.Since(start),
				r.UserAgent(),
				r.RemoteAddr,
			)
		})
	}
}

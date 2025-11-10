package middleware

import (
	"backend-form/m/internal/metrics"
	"net/http"
	"time"
)

// MetricsMiddleware records HTTP request metrics
func MetricsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		endpoint := r.URL.Path

		// Wrap response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call next handler
		next(rw, r)

		// Record metrics
		duration := time.Since(start)
		m := metrics.GetMetrics()
		m.IncrementHTTPRequest(endpoint)
		m.RecordHTTPDuration(endpoint, duration)

		// Record errors (4xx and 5xx)
		if rw.statusCode >= 400 {
			m.IncrementHTTPError(endpoint)
		}
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}


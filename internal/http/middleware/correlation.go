package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"backend-form/m/internal/logger"

	"go.uber.org/zap"
)

type correlationKey string

const CorrelationIDKey correlationKey = "correlation_id"
const CorrelationIDHeader = "X-Request-ID"

// CorrelationMiddleware adds a unique request ID to each request for tracing
func CorrelationMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if client provided a correlation ID
		correlationID := r.Header.Get(CorrelationIDHeader)

		// Generate new correlation ID if not provided
		if correlationID == "" {
			correlationID = generateCorrelationID()
		}

		// Add to response header for client tracing
		w.Header().Set(CorrelationIDHeader, correlationID)

		// Add to request context
		ctx := context.WithValue(r.Context(), CorrelationIDKey, correlationID)
		r = r.WithContext(ctx)

		// Add to logger context for this request
		logger.Info("Request started",
			zap.String("correlation_id", correlationID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
		)

		next(w, r)
	}
}

// GetCorrelationID retrieves the correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return id
	}
	return ""
}

// generateCorrelationID generates a unique correlation ID
func generateCorrelationID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return "fallback-id"
	}
	return hex.EncodeToString(bytes)
}


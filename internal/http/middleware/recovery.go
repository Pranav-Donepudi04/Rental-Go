package middleware

import (
	"backend-form/m/internal/logger"
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"
)

// RecoveryMiddleware recovers from panics and logs them with stack traces
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log panic with stack trace
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("remote_addr", r.RemoteAddr),
					zap.String("stack", string(debug.Stack())),
				)

				// Return 500 error
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}


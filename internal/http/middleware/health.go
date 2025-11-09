package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

// DatabaseHealthCheck checks database connectivity
type DatabaseHealthCheck struct {
	db *sql.DB
}

// NewDatabaseHealthCheck creates a new database health check
func NewDatabaseHealthCheck(db *sql.DB) *DatabaseHealthCheck {
	return &DatabaseHealthCheck{db: db}
}

// Check performs a health check on the database
func (h *DatabaseHealthCheck) Check() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.db.PingContext(ctx)
	return err == nil, err
}

// HealthCheckHandler returns a handler for the /health endpoint with database check
func (h *DatabaseHealthCheck) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	healthy, err := h.Check()

	status := "ok"
	statusCode := http.StatusOK
	if !healthy {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	response := map[string]interface{}{
		"status": status,
		"database": map[string]interface{}{
			"connected": healthy,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if err != nil {
		response["database"].(map[string]interface{})["error"] = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

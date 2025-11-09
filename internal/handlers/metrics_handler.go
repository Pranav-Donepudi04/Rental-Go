package handlers

import (
	"backend-form/m/internal/metrics"
	"encoding/json"
	"net/http"
)

// MetricsHandler handles metrics endpoint
type MetricsHandler struct{}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
}

// GetMetrics returns current metrics as JSON
func (h *MetricsHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	snapshot := metrics.GetMetrics().GetSnapshot()

	// Set headers before encoding
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// Encode JSON
	if err := json.NewEncoder(w).Encode(snapshot); err != nil {
		http.Error(w, "Failed to encode metrics", http.StatusInternalServerError)
		return
	}
}

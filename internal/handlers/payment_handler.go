package handlers

import (
	"backend-form/m/internal/domain"
	"backend-form/m/internal/metrics"
	"backend-form/m/internal/service"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PaymentHandler handles all payment-related HTTP requests (owner only)
type PaymentHandler struct {
	paymentService            *service.PaymentService
	paymentQueryService       *service.PaymentQueryService
	paymentTransactionService *service.PaymentTransactionService
	paymentHistoryService     *service.PaymentHistoryService
	dashboardService          *service.DashboardService
	userRepo                  interface{} // For future use if needed
}

// NewPaymentHandler creates a new PaymentHandler
func NewPaymentHandler(
	paymentService *service.PaymentService,
	paymentQueryService *service.PaymentQueryService,
	paymentTransactionService *service.PaymentTransactionService,
	paymentHistoryService *service.PaymentHistoryService,
	dashboardService *service.DashboardService,
) *PaymentHandler {
	return &PaymentHandler{
		paymentService:            paymentService,
		paymentQueryService:       paymentQueryService,
		paymentTransactionService: paymentTransactionService,
		paymentHistoryService:     paymentHistoryService,
		dashboardService:          dashboardService,
	}
}

// GetPayments returns all payments as JSON
func (h *PaymentHandler) GetPayments(w http.ResponseWriter, r *http.Request) {
	payments, err := h.paymentQueryService.GetAllPayments()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
}

// MarkPaymentAsPaid marks a payment as paid (legacy method - supports both old and new flow)
func (h *PaymentHandler) MarkPaymentAsPaid(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PaymentID     int    `json:"payment_id"`     // For legacy flow
		TransactionID string `json:"transaction_id"` // For new transaction verification flow
		Amount        int    `json:"amount"`         // Amount being verified
		PaymentDate   string `json:"payment_date"`   // Legacy: payment date
		Notes         string `json:"notes"`          // Legacy: notes
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get user from context (already validated by middleware)
	user, ok := r.Context().Value("user").(*domain.User)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// NEW: Transaction verification flow
	if req.TransactionID != "" {
		if req.Amount <= 0 {
			http.Error(w, "Amount must be greater than 0", http.StatusBadRequest)
			return
		}

		if err := h.paymentTransactionService.VerifyTransaction(req.TransactionID, req.Amount, user.ID); err != nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		// Invalidate dashboard cache since payment data changed
		h.dashboardService.InvalidateDashboardCache()

		// Track business metric
		metrics.GetMetrics().IncrementPaymentVerified()

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Transaction verified and payment updated",
		})
		return
	}

	// LEGACY: Old flow (mark full payment as paid)
	if req.PaymentID == 0 {
		http.Error(w, "Either payment_id or transaction_id required", http.StatusBadRequest)
		return
	}

	// Parse payment date
	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	if err := h.paymentService.MarkPaymentAsPaid(req.PaymentID, paymentDate, req.Notes); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Invalidate dashboard cache since payment data changed
	h.dashboardService.InvalidateDashboardCache()

	// Track business metric
	metrics.GetMetrics().IncrementPaymentProcessed()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Payment marked as paid",
	})
}

// GetPendingVerifications returns all pending transaction verifications (owner only)
func (h *PaymentHandler) GetPendingVerifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get tenant ID from query (optional - if provided, filter by tenant)
	tenantID := 0
	if tenantIDStr := r.URL.Query().Get("tenant_id"); tenantIDStr != "" {
		fmt.Sscanf(tenantIDStr, "%d", &tenantID)
	}

	// Get all pending verifications (0 = all tenants)
	pending, err := h.paymentTransactionService.GetPendingVerifications(tenantID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    pending,
	})
}

// SyncPaymentHistory syncs historical payment records for an existing tenant
func (h *PaymentHandler) SyncPaymentHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TenantID int                             `json:"tenant_id"`
		Payments []service.HistoricalPaymentData `json:"payments"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.TenantID <= 0 {
		http.Error(w, "tenant_id is required", http.StatusBadRequest)
		return
	}

	if len(req.Payments) == 0 {
		http.Error(w, "payments array is required", http.StatusBadRequest)
		return
	}

	// Sync payment history
	createdPayments, err := h.paymentHistoryService.SyncPaymentHistory(req.TenantID, req.Payments)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Invalidate dashboard cache since payment data changed
	h.dashboardService.InvalidateDashboardCache()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  fmt.Sprintf("Synced %d payment(s)", len(createdPayments)),
		"payments": createdPayments,
	})
}

// AdjustPaymentDueDate adjusts the due date of the first unpaid payment for a tenant
func (h *PaymentHandler) AdjustPaymentDueDate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TenantID int    `json:"tenant_id"`
		DueDate  string `json:"due_date"` // Format: "2006-01-02"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.TenantID <= 0 {
		http.Error(w, "tenant_id is required", http.StatusBadRequest)
		return
	}

	// Parse due date
	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// Adjust due date
	if err := h.paymentHistoryService.AdjustFirstPaymentDueDate(req.TenantID, dueDate); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Payment due date adjusted successfully",
	})
}

// RejectTransaction rejects a pending transaction
func (h *PaymentHandler) RejectTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TransactionID string `json:"transaction_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.TransactionID == "" {
		http.Error(w, "transaction_id is required", http.StatusBadRequest)
		return
	}

	// Reject transaction
	if err := h.paymentTransactionService.RejectTransaction(req.TransactionID); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Invalidate dashboard cache since transaction status changed
	h.dashboardService.InvalidateDashboardCache()

	// Track business metric
	metrics.GetMetrics().IncrementTransactionRejected()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Transaction rejected successfully",
	})
}

// CreatePayment creates a new payment (owner only)
func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Method not allowed",
		})
		return
	}

	var req struct {
		TenantID int    `json:"tenant_id"`
		UnitID   int    `json:"unit_id"`
		Label    string `json:"label"` // rent, water_bill, current_bill, maintenance
		Amount   int    `json:"amount"`
		DueDate  string `json:"due_date"` // Format: "2006-01-02"
		Notes    string `json:"notes"`    // Optional
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid JSON",
		})
		return
	}

	// Validate required fields
	if req.TenantID <= 0 {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "tenant_id is required",
		})
		return
	}
	if req.UnitID <= 0 {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "unit_id is required",
		})
		return
	}
	if req.Amount <= 0 {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "amount must be greater than 0",
		})
		return
	}
	if req.DueDate == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "due_date is required",
		})
		return
	}
	if req.Label == "" {
		req.Label = domain.PaymentLabelRent // Default to rent
	}

	// Parse due date
	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid date format. Use YYYY-MM-DD",
		})
		return
	}

	// Create payment using service
	createdPayment, err := h.paymentService.CreateCustomPayment(
		req.TenantID,
		req.UnitID,
		dueDate,
		req.Amount,
		req.Label,
		req.Notes,
	)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Invalidate dashboard cache since payment data changed
	h.dashboardService.InvalidateDashboardCache()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Payment created successfully",
		"payment": createdPayment,
	})
}

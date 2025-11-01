package handlers

import (
	"backend-form/m/internal/domain"
	"backend-form/m/internal/repository/interfaces"
	"backend-form/m/internal/service"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

// RentalHandler handles all rental management HTTP requests
type RentalHandler struct {
	unitService    *service.UnitService
	tenantService  *service.TenantService
	paymentService *service.PaymentService
	templates      *template.Template
	authService    *service.AuthService
	userRepo       interfaces.UserRepository
	cookieName     string
}

// NewRentalHandler creates a new RentalHandler
func NewRentalHandler(unitService *service.UnitService, tenantService *service.TenantService, paymentService *service.PaymentService, templates *template.Template, auth *service.AuthService) *RentalHandler {
	return &RentalHandler{
		unitService:    unitService,
		tenantService:  tenantService,
		paymentService: paymentService,
		templates:      templates,
		authService:    auth,
		cookieName:     "sid",
	}
}

// SetUserRepository sets the user repository (called from main.go after creation)
func (h *RentalHandler) SetUserRepository(userRepo interfaces.UserRepository) {
	h.userRepo = userRepo
}

// Dashboard renders the main dashboard
func (h *RentalHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get dashboard data
	units, err := h.unitService.GetAllUnits()
	if err != nil {
		http.Error(w, "Failed to load units", http.StatusInternalServerError)
		return
	}

	tenants, err := h.tenantService.GetAllTenants()
	if err != nil {
		http.Error(w, "Failed to load tenants", http.StatusInternalServerError)
		return
	}

	payments, err := h.paymentService.GetAllPayments()
	if err != nil {
		http.Error(w, "Failed to load payments", http.StatusInternalServerError)
		return
	}

	// Get summaries
	unitSummary, err := h.unitService.GetRentalSummary()
	if err != nil {
		http.Error(w, "Failed to load unit summary", http.StatusInternalServerError)
		return
	}

	paymentSummary, err := h.paymentService.GetPaymentSummary()
	if err != nil {
		http.Error(w, "Failed to load payment summary", http.StatusInternalServerError)
		return
	}

	// Prepare dashboard data
	dashboardData := map[string]interface{}{
		"Units":          units,
		"Tenants":        tenants,
		"Payments":       payments,
		"UnitSummary":    unitSummary,
		"PaymentSummary": paymentSummary,
	}

	if err := h.templates.ExecuteTemplate(w, "dashboard.html", dashboardData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// API Handlers for JSON responses

// GetUnits returns all units as JSON
func (h *RentalHandler) GetUnits(w http.ResponseWriter, r *http.Request) {
	units, err := h.unitService.GetAllUnits()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(units)
}

// GetTenants returns all tenants as JSON
func (h *RentalHandler) GetTenants(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.tenantService.GetAllTenants()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tenants)
}

// GetPayments returns all payments as JSON
func (h *RentalHandler) GetPayments(w http.ResponseWriter, r *http.Request) {
	payments, err := h.paymentService.GetAllPayments()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
}

// CreateTenant creates a new tenant
func (h *RentalHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tenant struct {
		Name             string `json:"name"`
		Phone            string `json:"phone"`
		AadharNumber     string `json:"aadhar_number"`
		MoveInDate       string `json:"move_in_date"`
		NumberOfPeople   int    `json:"number_of_people"`
		UnitID           int    `json:"unit_id"`
		IsExistingTenant bool   `json:"is_existing_tenant"` // If true, skip first payment creation
	}

	if err := json.NewDecoder(r.Body).Decode(&tenant); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Parse move-in date
	moveInDate, err := time.Parse("2006-01-02", tenant.MoveInDate)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	newTenant := &domain.Tenant{
		Name:           tenant.Name,
		Phone:          tenant.Phone,
		AadharNumber:   tenant.AadharNumber,
		MoveInDate:     moveInDate,
		NumberOfPeople: tenant.NumberOfPeople,
		UnitID:         tenant.UnitID,
	}

	if err := h.tenantService.CreateTenant(newTenant, tenant.IsExistingTenant); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Create login credentials for tenant and return temp password
	temp, err := h.authService.CreateTenantCredentials(newTenant.Phone, newTenant.ID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Tenant created, but failed to create credentials",
			"tenant":  newTenant,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"message":       "Tenant created successfully",
		"tenant":        newTenant,
		"temp_password": temp,
	})
}

// MarkPaymentAsPaid marks a payment as paid (legacy method - supports both old and new flow)
func (h *RentalHandler) MarkPaymentAsPaid(w http.ResponseWriter, r *http.Request) {
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

	// Get user from session (owner)
	cookie, err := r.Cookie(h.cookieName)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	sess, err := h.authService.ValidateSession(cookie.Value)
	if err != nil || sess == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if h.userRepo == nil {
		http.Error(w, "User repository not configured", http.StatusInternalServerError)
		return
	}
	user, err := h.userRepo.GetByID(sess.UserID)
	if err != nil || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// NEW: Transaction verification flow
	if req.TransactionID != "" {
		if req.Amount <= 0 {
			http.Error(w, "Amount must be greater than 0", http.StatusBadRequest)
			return
		}

		if err := h.paymentService.VerifyTransaction(req.TransactionID, req.Amount, user.ID); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Payment marked as paid",
	})
}

// GetSummary returns dashboard summary
func (h *RentalHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	unitSummary, err := h.unitService.GetRentalSummary()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	paymentSummary, err := h.paymentService.GetPaymentSummary()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	summary := map[string]interface{}{
		"units":    unitSummary,
		"payments": paymentSummary,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// VacateTenant handles tenant move-out
func (h *RentalHandler) VacateTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TenantID int `json:"tenant_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.tenantService.MoveOutTenant(req.TenantID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Tenant moved out successfully",
	})
}

// UnitDetails renders the unit detail page
func (h *RentalHandler) UnitDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract unit ID from URL path
	unitIDStr := r.URL.Path[len("/unit/"):]
	unitID := 0
	fmt.Sscanf(unitIDStr, "%d", &unitID)

	if unitID == 0 {
		http.Error(w, "Invalid unit ID", http.StatusBadRequest)
		return
	}

	// Get unit details
	unit, err := h.unitService.GetUnitByID(unitID)
	if err != nil {
		http.Error(w, "Unit not found", http.StatusNotFound)
		return
	}

	// Get tenant details if occupied
	var tenant *domain.Tenant
	var payments []*domain.Payment
	var pendingVerifications []*domain.PaymentTransaction
	if unit.IsOccupied {
		tenants, err := h.tenantService.GetTenantsByUnitID(unitID)
		if err == nil && len(tenants) > 0 {
			tenant = tenants[0] // Get the primary tenant

			// Get payment history for this tenant (transactions are loaded automatically)
			payments, err = h.paymentService.GetPaymentsByTenantID(tenant.ID)
			if err != nil {
				payments = []*domain.Payment{} // Empty slice if error
			}

			// Get pending verifications for this tenant
			pendingVerifications, err = h.paymentService.GetPendingVerifications(tenant.ID)
			if err != nil {
				pendingVerifications = []*domain.PaymentTransaction{} // Empty slice if error
			}
		}
	}

	// Prepare unit detail data
	unitData := map[string]interface{}{
		"Unit":                 unit,
		"Tenant":               tenant,
		"Payments":             payments,
		"PendingVerifications": pendingVerifications,
	}

	if err := h.templates.ExecuteTemplate(w, "unit-detail.html", unitData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Helper function to get payment status
func getPaymentStatus(payment *domain.Payment) string {
	if payment.IsPaid {
		return "Paid"
	}
	if time.Now().After(payment.DueDate) {
		return "Overdue"
	}
	return "Pending"
}

// GetPendingVerifications returns all pending transaction verifications (owner only)
func (h *RentalHandler) GetPendingVerifications(w http.ResponseWriter, r *http.Request) {
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
	pending, err := h.paymentService.GetPendingVerifications(tenantID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    pending,
	})
}

// SyncPaymentHistory syncs historical payment records for an existing tenant
// This allows marking past payments as paid when transitioning from manual to system management
func (h *RentalHandler) SyncPaymentHistory(w http.ResponseWriter, r *http.Request) {
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
	createdPayments, err := h.paymentService.SyncPaymentHistory(req.TenantID, req.Payments)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  fmt.Sprintf("Synced %d payment(s)", len(createdPayments)),
		"payments": createdPayments,
	})
}

// AdjustPaymentDueDate adjusts the due date of the first unpaid payment for a tenant
func (h *RentalHandler) AdjustPaymentDueDate(w http.ResponseWriter, r *http.Request) {
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
	if err := h.paymentService.AdjustFirstPaymentDueDate(req.TenantID, dueDate); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Payment due date adjusted successfully",
	})
}

// RejectTransaction rejects a pending transaction
func (h *RentalHandler) RejectTransaction(w http.ResponseWriter, r *http.Request) {
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
	if err := h.paymentService.RejectTransaction(req.TransactionID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Transaction rejected successfully",
	})
}

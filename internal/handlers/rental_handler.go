package handlers

import (
	"backend-form/m/internal/domain"
	"backend-form/m/internal/repository/interfaces"
	"backend-form/m/internal/service"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

// DashboardHandler handles dashboard and unit-related HTTP requests (owner only)
type DashboardHandler struct {
	unitService               *service.UnitService
	tenantService             *service.TenantService
	paymentService            *service.PaymentService
	paymentTransactionService *service.PaymentTransactionService
	dashboardService          *service.DashboardService
	templates                 *template.Template
}

// NewDashboardHandler creates a new DashboardHandler
func NewDashboardHandler(
	unitService *service.UnitService,
	tenantService *service.TenantService,
	paymentService *service.PaymentService,
	paymentTransactionService *service.PaymentTransactionService,
	dashboardService *service.DashboardService,
	templates *template.Template,
) *DashboardHandler {
	return &DashboardHandler{
		unitService:               unitService,
		tenantService:             tenantService,
		paymentService:            paymentService,
		paymentTransactionService: paymentTransactionService,
		dashboardService:          dashboardService,
		templates:                 templates,
	}
}

// RentalHandler is kept for backward compatibility - delegates to DashboardHandler
// Deprecated: Use DashboardHandler, PaymentHandler, and TenantManagementHandler instead
type RentalHandler struct {
	*DashboardHandler
	paymentHandler          *PaymentHandler
	tenantManagementHandler *TenantManagementHandler
}

// NewRentalHandler creates a new RentalHandler (backward compatibility wrapper)
func NewRentalHandler(
	unitService *service.UnitService,
	tenantService *service.TenantService,
	paymentService *service.PaymentService,
	paymentQueryService *service.PaymentQueryService,
	paymentTransactionService *service.PaymentTransactionService,
	paymentHistoryService *service.PaymentHistoryService,
	dashboardService *service.DashboardService,
	notificationService *service.NotificationService,
	templates *template.Template,
	auth *service.AuthService,
) *RentalHandler {
	dashboardHandler := NewDashboardHandler(
		unitService,
		tenantService,
		paymentService,
		paymentTransactionService,
		dashboardService,
		templates,
	)

	paymentHandler := NewPaymentHandler(
		paymentService,
		paymentQueryService,
		paymentTransactionService,
		paymentHistoryService,
		dashboardService,
	)

	tenantManagementHandler := NewTenantManagementHandler(
		tenantService,
		auth,
		dashboardService,
	)

	return &RentalHandler{
		DashboardHandler:        dashboardHandler,
		paymentHandler:          paymentHandler,
		tenantManagementHandler: tenantManagementHandler,
	}
}

// SetUserRepository sets the user repository (called from main.go after creation)
// Deprecated: No longer needed with new handler structure
func (h *RentalHandler) SetUserRepository(userRepo interfaces.UserRepository) {
	// No-op for backward compatibility
}

// Delegation method for backward compatibility
func (h *RentalHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	h.DashboardHandler.Dashboard(w, r)
}

// Dashboard renders the main dashboard
func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get dashboard data using DashboardService
	data, err := h.dashboardService.GetDashboardData()
	if err != nil {
		http.Error(w, "Failed to load dashboard data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare dashboard data for template (convert to map)
	dashboardData := map[string]interface{}{
		"Units":          data.Units,
		"Tenants":        data.Tenants,
		"Payments":       data.Payments,
		"UnitSummary":    data.UnitSummary,
		"PaymentSummary": data.PaymentSummary,
	}

	if err := h.templates.ExecuteTemplate(w, "dashboard.html", dashboardData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// API Handlers for JSON responses

// Delegation method for backward compatibility
func (h *RentalHandler) GetUnits(w http.ResponseWriter, r *http.Request) {
	h.DashboardHandler.GetUnits(w, r)
}

// GetUnits returns all units as JSON
func (h *DashboardHandler) GetUnits(w http.ResponseWriter, r *http.Request) {
	units, err := h.unitService.GetAllUnits()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(units)
}

// Delegation methods for backward compatibility
func (h *RentalHandler) GetTenants(w http.ResponseWriter, r *http.Request) {
	h.tenantManagementHandler.GetTenants(w, r)
}

func (h *RentalHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	h.tenantManagementHandler.CreateTenant(w, r)
}

func (h *RentalHandler) VacateTenant(w http.ResponseWriter, r *http.Request) {
	h.tenantManagementHandler.VacateTenant(w, r)
}

func (h *RentalHandler) RegenerateTenantPassword(w http.ResponseWriter, r *http.Request) {
	h.tenantManagementHandler.RegenerateTenantPassword(w, r)
}

func (h *RentalHandler) GetPayments(w http.ResponseWriter, r *http.Request) {
	h.paymentHandler.GetPayments(w, r)
}

func (h *RentalHandler) MarkPaymentAsPaid(w http.ResponseWriter, r *http.Request) {
	h.paymentHandler.MarkPaymentAsPaid(w, r)
}

func (h *RentalHandler) GetPendingVerifications(w http.ResponseWriter, r *http.Request) {
	h.paymentHandler.GetPendingVerifications(w, r)
}

func (h *RentalHandler) SyncPaymentHistory(w http.ResponseWriter, r *http.Request) {
	h.paymentHandler.SyncPaymentHistory(w, r)
}

func (h *RentalHandler) AdjustPaymentDueDate(w http.ResponseWriter, r *http.Request) {
	h.paymentHandler.AdjustPaymentDueDate(w, r)
}

func (h *RentalHandler) RejectTransaction(w http.ResponseWriter, r *http.Request) {
	h.paymentHandler.RejectTransaction(w, r)
}

func (h *RentalHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	h.paymentHandler.CreatePayment(w, r)
}

// Delegation method for backward compatibility
func (h *RentalHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	h.DashboardHandler.GetSummary(w, r)
}

// GetSummary returns dashboard summary
func (h *DashboardHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.dashboardService.GetDashboardSummary()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// Delegation method for backward compatibility
func (h *RentalHandler) UnitDetails(w http.ResponseWriter, r *http.Request) {
	h.DashboardHandler.UnitDetails(w, r)
}

// UnitDetails renders the unit detail page
func (h *DashboardHandler) UnitDetails(w http.ResponseWriter, r *http.Request) {
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
			pendingVerifications, err = h.paymentTransactionService.GetPendingVerifications(tenant.ID)
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

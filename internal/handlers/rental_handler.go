package handlers

import (
	"backend-form/m/internal/models"
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
}

// NewRentalHandler creates a new RentalHandler
func NewRentalHandler(unitService *service.UnitService, tenantService *service.TenantService, paymentService *service.PaymentService, templates *template.Template) *RentalHandler {
	return &RentalHandler{
		unitService:    unitService,
		tenantService:  tenantService,
		paymentService: paymentService,
		templates:      templates,
	}
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
		Name           string `json:"name"`
		Phone          string `json:"phone"`
		AadharNumber   string `json:"aadhar_number"`
		MoveInDate     string `json:"move_in_date"`
		NumberOfPeople int    `json:"number_of_people"`
		UnitID         int    `json:"unit_id"`
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

	newTenant := &models.Tenant{
		Name:           tenant.Name,
		Phone:          tenant.Phone,
		AadharNumber:   tenant.AadharNumber,
		MoveInDate:     moveInDate,
		NumberOfPeople: tenant.NumberOfPeople,
		UnitID:         tenant.UnitID,
	}

	if err := h.tenantService.CreateTenant(newTenant); err != nil {
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
		"message": "Tenant created successfully",
		"tenant":  newTenant,
	})
}

// MarkPaymentAsPaid marks a payment as paid
func (h *RentalHandler) MarkPaymentAsPaid(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PaymentID   int    `json:"payment_id"`
		PaymentDate string `json:"payment_date"`
		Notes       string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
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
	var tenant *models.Tenant
	var payments []*models.Payment
	if unit.IsOccupied {
		tenants, err := h.tenantService.GetTenantsByUnitID(unitID)
		if err == nil && len(tenants) > 0 {
			tenant = tenants[0] // Get the primary tenant

			// Get payment history for this tenant
			payments, err = h.paymentService.GetPaymentsByTenantID(tenant.ID)
			if err != nil {
				payments = []*models.Payment{} // Empty slice if error
			}
		}
	}

	// Prepare unit detail data
	unitData := map[string]interface{}{
		"Unit":     unit,
		"Tenant":   tenant,
		"Payments": payments,
	}

	if err := h.templates.ExecuteTemplate(w, "unit-detail.html", unitData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Helper function to get payment status
func getPaymentStatus(payment *models.Payment) string {
	if payment.IsPaid {
		return "Paid"
	}
	if time.Now().After(payment.DueDate) {
		return "Overdue"
	}
	return "Pending"
}

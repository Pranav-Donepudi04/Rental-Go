package handlers

import (
	"backend-form/m/internal/domain"
	"backend-form/m/internal/metrics"
	"backend-form/m/internal/service"
	"encoding/json"
	"net/http"
	"time"
)

// TenantManagementHandler handles owner-facing tenant management operations
type TenantManagementHandler struct {
	tenantService    *service.TenantService
	authService      *service.AuthService
	dashboardService *service.DashboardService
}

// NewTenantManagementHandler creates a new TenantManagementHandler
func NewTenantManagementHandler(
	tenantService *service.TenantService,
	authService *service.AuthService,
	dashboardService *service.DashboardService,
) *TenantManagementHandler {
	return &TenantManagementHandler{
		tenantService:    tenantService,
		authService:      authService,
		dashboardService: dashboardService,
	}
}

// GetTenants returns all tenants as JSON
func (h *TenantManagementHandler) GetTenants(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.tenantService.GetAllTenants()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(tenants)
}

// CreateTenant creates a new tenant
func (h *TenantManagementHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
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
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}); err != nil {
			return
		}
		return
	}

	// Create login credentials for tenant and return temp password
	temp, err := h.authService.CreateTenantCredentials(newTenant.Phone, newTenant.ID)
	if err != nil {
		// Invalidate dashboard cache even if credentials creation failed
		h.dashboardService.InvalidateDashboardCache()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Tenant created, but failed to create credentials",
			"tenant":  newTenant,
		}); err != nil {
			return
		}
		return
	}

	// Invalidate dashboard cache since tenant data changed
	h.dashboardService.InvalidateDashboardCache()

	// Track business metric
	metrics.GetMetrics().IncrementTenantCreated()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"message":       "Tenant created successfully",
		"tenant":        newTenant,
		"temp_password": temp,
	}); err != nil {
		return
	}
}

// VacateTenant handles tenant move-out
func (h *TenantManagementHandler) VacateTenant(w http.ResponseWriter, r *http.Request) {
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
		TenantID int `json:"tenant_id"`
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

	if err := h.tenantService.MoveOutTenant(req.TenantID); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Invalidate dashboard cache since tenant data changed
	h.dashboardService.InvalidateDashboardCache()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Tenant moved out successfully",
	})
}

// RegenerateTenantPassword regenerates the temporary password for an existing tenant
func (h *TenantManagementHandler) RegenerateTenantPassword(w http.ResponseWriter, r *http.Request) {
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
		TenantID int `json:"tenant_id"`
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

	// Get tenant to get phone number
	tenant, err := h.tenantService.GetTenantByID(req.TenantID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Tenant not found",
		})
		return
	}

	// Regenerate password using CreateTenantCredentials (it handles both new and existing users)
	tempPassword, err := h.authService.CreateTenantCredentials(tenant.Phone, tenant.ID)
	if err != nil {
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
		"success":       true,
		"message":       "Password regenerated successfully",
		"temp_password": tempPassword,
	})
}

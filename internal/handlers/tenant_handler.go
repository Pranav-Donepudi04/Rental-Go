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

type TenantHandler struct {
	tenantService             *service.TenantService
	paymentService            *service.PaymentService
	paymentTransactionService *service.PaymentTransactionService
	users                     interfaces.UserRepository
	templates                 *template.Template
	cookieName                string
	auth                      *service.AuthService
}

func NewTenantHandler(tenant *service.TenantService, payment *service.PaymentService, paymentTransaction *service.PaymentTransactionService, users interfaces.UserRepository, templates *template.Template, cookieName string, auth *service.AuthService) *TenantHandler {
	return &TenantHandler{
		tenantService:             tenant,
		paymentService:            payment,
		paymentTransactionService: paymentTransaction,
		users:                     users,
		templates:                 templates,
		cookieName:                cookieName,
		auth:                      auth,
	}
}

func (h *TenantHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Get user from context (already validated by middleware)
	// Uses the same string key that router middleware uses
	user, ok := r.Context().Value("user").(*domain.User)
	if !ok || user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if user.TenantID == nil {
		http.Error(w, "User account is not linked to a tenant. Please contact the owner to set up your tenant account.", http.StatusNotFound)
		return
	}
	tenant, err := h.tenantService.GetTenantByID(*user.TenantID)
	if err != nil {
		http.Error(w, "Failed to get tenant: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if tenant == nil {
		http.Error(w, "Tenant record not found. The tenant associated with your account may have been deleted.", http.StatusNotFound)
		return
	}
	payments, _ := h.paymentService.GetPaymentsByTenantID(tenant.ID)

	// Calculate family member limits for template
	maxFamilyMembers := tenant.NumberOfPeople - 1
	currentFamilyCount := len(tenant.FamilyMembers)
	isAtLimit := currentFamilyCount >= maxFamilyMembers

	// Get UPI ID and payment method for display
	upiID := h.paymentService.GetDefaultUPIID()
	paymentMethod := h.paymentService.GetDefaultPaymentMethod()

	data := map[string]interface{}{
		"Tenant":               tenant,
		"Payments":             payments,
		"MaxFamilyMembers":     maxFamilyMembers,
		"CurrentFamilyCount":   currentFamilyCount,
		"IsFamilyLimitReached": isAtLimit,
		"UPIID":                upiID,
		"PaymentMethod":        paymentMethod,
	}
	_ = h.templates.ExecuteTemplate(w, "tenant-dashboard.html", data)
}

func (h *TenantHandler) SubmitPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Method not allowed",
		})
		return
	}
	// Get user from context (already validated by middleware)
	// Uses the same string key that router middleware uses
	user, ok := r.Context().Value("user").(*domain.User)
	if !ok || user == nil || user.TenantID == nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}
	// Parse form values
	if err := r.ParseForm(); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Bad request",
		})
		return
	}
	txn := r.FormValue("txn_id")
	amountStr := r.FormValue("amount")

	if txn == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Transaction ID is required",
		})
		return
	}

	// Amount is required - validate it
	if amountStr == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Amount is required",
		})
		return
	}

	var amount int
	if _, err := fmt.Sscanf(amountStr, "%d", &amount); err != nil || amount <= 0 {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Amount must be a positive number",
		})
		return
	}

	suggestedAmount := &amount

	if err := h.paymentTransactionService.SubmitPaymentIntent(*user.TenantID, txn, suggestedAmount); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TenantHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Method not allowed",
		})
		return
	}
	// Get user from context (already validated by middleware)
	// Uses the same string key that router middleware uses
	user, ok := r.Context().Value("user").(*domain.User)
	if !ok || user == nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}
	var body struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid JSON body",
		})
		return
	}
	if err := h.auth.ValidatePasswordStrength(body.NewPassword); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	if !h.auth.ComparePassword(user.PasswordHash, body.OldPassword) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Old password is incorrect",
		})
		return
	}
	hashedPassword, err := h.auth.HashPassword(body.NewPassword)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to hash password",
		})
		return
	}
	if err := h.users.UpdatePassword(user.ID, hashedPassword); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to update password",
		})
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Password updated successfully",
	})
}

// AddFamilyMember handles adding a family member (tenant only)
func (h *TenantHandler) AddFamilyMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Get user from context (already validated by middleware)
	// Uses the same string key that router middleware uses
	user, ok := r.Context().Value("user").(*domain.User)
	if !ok || user == nil || user.TenantID == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	tenant, err := h.tenantService.GetTenantByID(*user.TenantID)
	if err != nil || tenant == nil {
		http.Error(w, "tenant not found", http.StatusNotFound)
		return
	}
	// Validate number of people limit
	// NumberOfPeople includes the tenant themselves, so max family members = NumberOfPeople - 1
	currentFamilyCount := len(tenant.FamilyMembers)
	maxAllowedFamilyMembers := tenant.NumberOfPeople - 1
	if currentFamilyCount >= maxAllowedFamilyMembers {
		http.Error(w, fmt.Sprintf("cannot add more family members: limit is %d total people (tenant + %d family member(s))", tenant.NumberOfPeople, maxAllowedFamilyMembers), http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	familyMember := &domain.FamilyMember{
		TenantID:     *user.TenantID,
		Name:         r.FormValue("name"),
		Age:          parseAge(r.FormValue("age")),
		Relationship: r.FormValue("relationship"),
		AadharNumber: r.FormValue("aadhar_number"),
	}
	if err := h.tenantService.AddFamilyMember(familyMember); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Family member added successfully",
	})
}

// Helper function to parse age
func parseAge(ageStr string) int {
	var age int
	fmt.Sscanf(ageStr, "%d", &age)
	return age
}

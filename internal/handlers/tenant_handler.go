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
	c, err := r.Cookie(h.cookieName)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	sess, err := h.auth.ValidateSession(c.Value)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	// Resolve user and tenant id
	user, err := h.users.GetByID(sess.UserID)
	if err != nil {
		http.Error(w, "Failed to get user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil {
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

	data := map[string]interface{}{
		"Tenant":               tenant,
		"Payments":             payments,
		"MaxFamilyMembers":     maxFamilyMembers,
		"CurrentFamilyCount":   currentFamilyCount,
		"IsFamilyLimitReached": isAtLimit,
	}
	_ = h.templates.ExecuteTemplate(w, "tenant-dashboard.html", data)
}

func (h *TenantHandler) SubmitPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	c, err := r.Cookie(h.cookieName)
	if err != nil {
		http.Error(w, "no cookie", http.StatusUnauthorized)
		return
	}
	sess, err := h.auth.ValidateSession(c.Value)
	if err != nil {
		http.Error(w, "invalid session", http.StatusUnauthorized)
		return
	}
	if sess == nil {
		http.Error(w, "session expired", http.StatusUnauthorized)
		return
	}
	// For simplicity use form value "txn_id"
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	txn := r.FormValue("txn_id")
	if txn == "" {
		http.Error(w, "txn_id required", http.StatusBadRequest)
		return
	}
	user, err := h.users.GetByID(sess.UserID)
	if err != nil || user == nil || user.TenantID == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err := h.paymentTransactionService.SubmitPaymentIntent(*user.TenantID, txn); err != nil {
		http.Error(w, "failed", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TenantHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	c, err := r.Cookie(h.cookieName)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	sess, err := h.auth.ValidateSession(c.Value)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if len(body.NewPassword) < 6 {
		http.Error(w, "password too short", http.StatusBadRequest)
		return
	}
	user, err := h.users.GetByID(sess.UserID)
	if err != nil || user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if !h.auth.ComparePassword(user.PasswordHash, body.OldPassword) {
		http.Error(w, "old password incorrect", http.StatusBadRequest)
		return
	}
	if err := h.users.UpdatePassword(user.ID, h.auth.HashPassword(body.NewPassword)); err != nil {
		http.Error(w, "failed to update", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
}

// AddFamilyMember handles adding a family member (tenant only)
func (h *TenantHandler) AddFamilyMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	c, err := r.Cookie(h.cookieName)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	sess, err := h.auth.ValidateSession(c.Value)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	user, err := h.users.GetByID(sess.UserID)
	if err != nil || user == nil || user.TenantID == nil {
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

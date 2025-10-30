package handlers

import (
	"backend-form/m/internal/repository/interfaces"
	"backend-form/m/internal/service"
	"encoding/json"
	"html/template"
	"net/http"
)

type TenantHandler struct {
	tenantService  *service.TenantService
	paymentService *service.PaymentService
	users          interfaces.UserRepository
	templates      *template.Template
	cookieName     string
	auth           *service.AuthService
}

func NewTenantHandler(tenant *service.TenantService, payment *service.PaymentService, users interfaces.UserRepository, templates *template.Template, cookieName string, auth *service.AuthService) *TenantHandler {
	return &TenantHandler{tenantService: tenant, paymentService: payment, users: users, templates: templates, cookieName: cookieName, auth: auth}
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
	if err != nil || user == nil || user.TenantID == nil {
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}
	tenant, err := h.tenantService.GetTenantByID(*user.TenantID)
	if err != nil || tenant == nil {
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}
	payments, _ := h.paymentService.GetPaymentsByTenantID(tenant.ID)
	data := map[string]interface{}{
		"Tenant":   tenant,
		"Payments": payments,
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
	if err := h.paymentService.SubmitPaymentIntent(*user.TenantID, txn); err != nil {
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

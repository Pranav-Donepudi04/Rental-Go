package handlers

import (
	"backend-form/m/internal/metrics"
	"backend-form/m/internal/service"
	"encoding/json"
	"html/template"
	"net/http"
	"time"
)

type AuthHandler struct {
	auth           *service.AuthService
	templates      *template.Template
	cookieName     string
	cookieSecure   bool
	cookieSameSite http.SameSite
}

func NewAuthHandler(auth *service.AuthService, templates *template.Template, cookieName string, cookieSecure bool, cookieSameSite string) *AuthHandler {
	sameSite := http.SameSiteStrictMode
	switch cookieSameSite {
	case "Lax":
		sameSite = http.SameSiteLaxMode
	case "None":
		sameSite = http.SameSiteNoneMode
	default:
		sameSite = http.SameSiteStrictMode
	}
	return &AuthHandler{
		auth:           auth,
		templates:      templates,
		cookieName:     cookieName,
		cookieSecure:   cookieSecure,
		cookieSameSite: sameSite,
	}
}

// GetAuthService returns the auth service for external access
func (h *AuthHandler) GetAuthService() *service.AuthService {
	return h.auth
}

// GetCookieName returns the cookie name for session management
func (h *AuthHandler) GetCookieName() string {
	return h.cookieName
}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = h.templates.ExecuteTemplate(w, "login.html", nil)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct{ Phone, Password, Role string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	sess, user, err := h.auth.Login(body.Phone, body.Password)
	if err != nil {
		metrics.GetMetrics().IncrementLoginFailure()
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	metrics.GetMetrics().IncrementLogin()
	if body.Role != "" && body.Role != string(user.UserType) {
		http.Error(w, "invalid role", http.StatusUnauthorized)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName,
		Value:    sess.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: h.cookieSameSite,
		Expires:  sess.ExpiresAt,
	})
	if user.UserType == "owner" {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/me", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	c, err := r.Cookie(h.cookieName)
	if err == nil {
		_ = h.auth.Logout(c.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: h.cookieSameSite,
		Expires:  time.Unix(0, 0),
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

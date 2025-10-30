package middleware

import (
	"backend-form/m/internal/domain"
	"backend-form/m/internal/repository/interfaces"
	"context"
	"net/http"
)

// AuthService interface for session validation
type AuthService interface {
	ValidateSession(sessionID string) (*domain.Session, error)
}

// LoadSession middleware attaches user to context if cookie is present
func LoadSession(authService AuthService, userRepo interfaces.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("sid")
			if err != nil {
				// No session cookie, continue without user in context
				next.ServeHTTP(w, r)
				return
			}

			session, err := authService.ValidateSession(cookie.Value)
			if err != nil || session == nil {
				// Invalid session, continue without user in context
				next.ServeHTTP(w, r)
				return
			}

			user, err := userRepo.GetByID(session.UserID)
			if err != nil || user == nil {
				// User not found, continue without user in context
				next.ServeHTTP(w, r)
				return
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireOwner middleware ensures the user is an owner
func RequireOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value("user").(*domain.User)
		if !ok || user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if user.UserType != domain.UserTypeOwner {
			http.Error(w, "Unauthorized: Owner access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireTenant middleware ensures the user is a tenant
func RequireTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value("user").(*domain.User)
		if !ok || user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if user.UserType != domain.UserTypeTenant {
			http.Error(w, "Unauthorized: Tenant access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext retrieves user from context
func GetUserFromContext(ctx context.Context) (*domain.User, bool) {
	user, ok := ctx.Value("user").(*domain.User)
	return user, ok
}

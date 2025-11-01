package http

import (
	"backend-form/m/internal/domain"
	"backend-form/m/internal/handlers"
	"backend-form/m/internal/repository/interfaces"
	"context"
	"fmt"
	"net/http"
)

// Router handles all HTTP routing
type Router struct {
	authHandler   *handlers.AuthHandler
	rentalHandler *handlers.RentalHandler
	tenantHandler *handlers.TenantHandler
	userRepo      interfaces.UserRepository
}

// UserContextKey is the key for storing user in context
type UserContextKey string

const UserKey UserContextKey = "user"

// NewRouter creates a new router with all handlers
func NewRouter(
	authHandler *handlers.AuthHandler,
	rentalHandler *handlers.RentalHandler,
	tenantHandler *handlers.TenantHandler,
	userRepo interfaces.UserRepository,
) *Router {
	return &Router{
		authHandler:   authHandler,
		rentalHandler: rentalHandler,
		tenantHandler: tenantHandler,
		userRepo:      userRepo,
	}
}

// SetupRoutes registers all HTTP routes
func (r *Router) SetupRoutes() {
	// Public routes
	http.HandleFunc("/login", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			r.authHandler.LoginPage(w, req)
			return
		}
		if req.Method == http.MethodPost {
			r.authHandler.Login(w, req)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
	http.HandleFunc("/logout", r.authHandler.Logout)

	// Redirect root to login
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
	})

	// Owner-only routes (with middleware)
	http.HandleFunc("/dashboard", r.requireOwner(r.rentalHandler.Dashboard))
	http.HandleFunc("/unit/", r.requireOwner(r.rentalHandler.UnitDetails))

	// Tenant-only routes
	http.HandleFunc("/me", r.requireTenant(r.tenantHandler.Me))

	// API routes
	http.HandleFunc("/api/units", r.requireOwner(r.rentalHandler.GetUnits))
	http.HandleFunc("/api/payments/submit", r.requireTenant(r.tenantHandler.SubmitPayment))
	http.HandleFunc("/api/me/change-password", r.requireTenant(r.tenantHandler.ChangePassword))
	http.HandleFunc("/api/me/family-members", r.requireTenant(r.tenantHandler.AddFamilyMember))
	http.HandleFunc("/api/tenants", r.requireOwner(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" {
			r.rentalHandler.GetTenants(w, req)
		} else if req.Method == "POST" {
			r.rentalHandler.CreateTenant(w, req)
		}
	}))
	http.HandleFunc("/api/payments", r.requireOwner(r.rentalHandler.GetPayments))
	http.HandleFunc("/api/payments/mark-paid", r.requireOwner(r.rentalHandler.MarkPaymentAsPaid))
	http.HandleFunc("/api/payments/pending-verifications", r.requireOwner(r.rentalHandler.GetPendingVerifications))
	http.HandleFunc("/api/tenants/vacate", r.requireOwner(r.rentalHandler.VacateTenant))
	http.HandleFunc("/api/summary", r.requireOwner(r.rentalHandler.GetSummary))
	http.HandleFunc("/api/payments/sync-history", r.requireOwner(r.rentalHandler.SyncPaymentHistory))
	http.HandleFunc("/api/payments/adjust-due-date", r.requireOwner(r.rentalHandler.AdjustPaymentDueDate))
	http.HandleFunc("/api/payments/reject-transaction", r.requireOwner(r.rentalHandler.RejectTransaction))
}

// SetUserRepository sets the user repository on the rental handler
func (r *Router) SetUserRepository(userRepo interfaces.UserRepository) {
	r.rentalHandler.SetUserRepository(userRepo)
}

// requireOwner middleware ensures the user is an owner
func (r *Router) requireOwner(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Load session and validate owner role
		user, err := r.loadSessionAndValidateRole(req, "owner")
		if err != nil {
			http.Redirect(w, req, "/login", http.StatusSeeOther)
			return
		}

		// Add user to context for handlers to use
		ctx := contextWithUser(req.Context(), user)
		next(w, req.WithContext(ctx))
	}
}

// requireTenant middleware ensures the user is a tenant
func (r *Router) requireTenant(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Load session and validate tenant role
		user, err := r.loadSessionAndValidateRole(req, "tenant")
		if err != nil {
			http.Redirect(w, req, "/login", http.StatusSeeOther)
			return
		}

		// Add user to context for handlers to use
		ctx := contextWithUser(req.Context(), user)
		next(w, req.WithContext(ctx))
	}
}

// loadSessionAndValidateRole loads session and validates user role
func (r *Router) loadSessionAndValidateRole(req *http.Request, requiredRole string) (*domain.User, error) {
	cookie, err := req.Cookie("sid")
	if err != nil {
		return nil, err
	}

	session, err := r.authHandler.GetAuthService().ValidateSession(cookie.Value)
	if err != nil || session == nil {
		return nil, err
	}

	user, err := r.userRepo.GetByID(session.UserID)
	if err != nil || user == nil {
		return nil, err
	}

	// Check role
	if requiredRole == "owner" && user.UserType != domain.UserTypeOwner {
		return nil, fmt.Errorf("unauthorized: owner required")
	}
	if requiredRole == "tenant" && user.UserType != domain.UserTypeTenant {
		return nil, fmt.Errorf("unauthorized: tenant required")
	}

	return user, nil
}

// contextWithUser adds user to context
func contextWithUser(ctx context.Context, user *domain.User) context.Context {
	return context.WithValue(ctx, UserKey, user)
}

// GetUserFromContext retrieves user from context
func GetUserFromContext(ctx context.Context) (*domain.User, bool) {
	user, ok := ctx.Value(UserKey).(*domain.User)
	return user, ok
}

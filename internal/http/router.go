package http

import (
	"backend-form/m/internal/domain"
	"backend-form/m/internal/handlers"
	"backend-form/m/internal/http/middleware"
	"backend-form/m/internal/repository/interfaces"
	"context"
	"fmt"
	"net/http"
)

// Router handles all HTTP routing
type Router struct {
	authHandler    *handlers.AuthHandler
	rentalHandler  *handlers.RentalHandler
	tenantHandler  *handlers.TenantHandler
	metricsHandler *handlers.MetricsHandler
	userRepo       interfaces.UserRepository
	loginLimiter   *middleware.RateLimiter
	dbHealthCheck  *middleware.DatabaseHealthCheck
}

// UserContextKey is the key for storing user in context
// Using string directly allows handlers to access it without import cycles
type UserContextKey string

const UserKey UserContextKey = "user"

// userContextKeyString is the string value for the context key
// Handlers can use this string directly to avoid import cycles
const userContextKeyString = "user"

// NewRouter creates a new router with all handlers
func NewRouter(
	authHandler *handlers.AuthHandler,
	rentalHandler *handlers.RentalHandler,
	tenantHandler *handlers.TenantHandler,
	userRepo interfaces.UserRepository,
	loginLimiter *middleware.RateLimiter,
	dbHealthCheck *middleware.DatabaseHealthCheck,
) *Router {
	return &Router{
		authHandler:    authHandler,
		rentalHandler:  rentalHandler,
		tenantHandler:  tenantHandler,
		metricsHandler: handlers.NewMetricsHandler(),
		userRepo:       userRepo,
		loginLimiter:   loginLimiter,
		dbHealthCheck:  dbHealthCheck,
	}
}

// SetupRoutes registers all HTTP routes
func (r *Router) SetupRoutes() {
	// Wrap all handlers with correlation, recovery, compression, and metrics middleware
	correlationWrapper := middleware.CorrelationMiddleware
	recoveryWrapper := middleware.RecoveryMiddleware
	compressionWrapper := middleware.CompressionMiddleware
	metricsWrapper := middleware.MetricsMiddleware

	// Public routes
	loginHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			r.authHandler.LoginPage(w, req)
			return
		}
		if req.Method == http.MethodPost {
			// Apply rate limiting to login POST requests
			if r.loginLimiter != nil {
				r.loginLimiter.Limit(r.authHandler.Login)(w, req)
			} else {
				r.authHandler.Login(w, req)
			}
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
	http.HandleFunc("/login", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(loginHandler)).ServeHTTP))))
	http.HandleFunc("/logout", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(http.HandlerFunc(r.authHandler.Logout))).ServeHTTP))))

	// Health check endpoint with database check
	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if r.dbHealthCheck != nil {
			r.dbHealthCheck.HealthCheckHandler(w, req)
		} else {
			// Fallback if health check not configured
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		}
	})
	http.HandleFunc("/health", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(healthHandler)).ServeHTTP))))

	// Metrics endpoint (owner only)
	http.HandleFunc("/metrics", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(r.requireOwner(r.metricsHandler.GetMetrics))).ServeHTTP))))

	// Static file: QR Code image
	http.HandleFunc("/static/qrcode.png", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "QRCode.png")
	}))).ServeHTTP))))

	// Redirect root to login
	rootHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
	})
	http.HandleFunc("/", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(rootHandler)).ServeHTTP))))

	// Owner-only routes (with middleware)
	http.HandleFunc("/dashboard", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(r.requireOwner(r.rentalHandler.Dashboard))).ServeHTTP))))
	http.HandleFunc("/unit/", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(r.requireOwner(r.rentalHandler.UnitDetails))).ServeHTTP))))

	// Tenant-only routes
	http.HandleFunc("/me", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(r.requireTenant(r.tenantHandler.Me))).ServeHTTP))))

	// API routes
	http.HandleFunc("/api/units", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(r.requireOwner(r.rentalHandler.GetUnits))).ServeHTTP))))
	http.HandleFunc("/api/payments/submit", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(r.requireTenant(r.tenantHandler.SubmitPayment))).ServeHTTP))))
	http.HandleFunc("/api/me/change-password", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(r.requireTenant(r.tenantHandler.ChangePassword))).ServeHTTP))))
	http.HandleFunc("/api/me/family-members", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(r.requireTenant(r.tenantHandler.AddFamilyMember))).ServeHTTP))))
	tenantsHandler := r.requireOwner(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" {
			r.rentalHandler.GetTenants(w, req)
		} else if req.Method == "POST" {
			r.rentalHandler.CreateTenant(w, req)
		}
	})
	http.HandleFunc("/api/tenants", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(tenantsHandler)).ServeHTTP))))
	http.HandleFunc("/api/payments", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(r.requireOwner(r.rentalHandler.GetPayments))).ServeHTTP))))
	http.HandleFunc("/api/payments/mark-paid", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(r.requireOwner(r.rentalHandler.MarkPaymentAsPaid))).ServeHTTP))))
	http.HandleFunc("/api/payments/pending-verifications", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(r.requireOwner(r.rentalHandler.GetPendingVerifications))).ServeHTTP))))
	http.HandleFunc("/api/tenants/vacate", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(r.requireOwner(r.rentalHandler.VacateTenant))).ServeHTTP))))
	http.HandleFunc("/api/tenants/regenerate-password", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(r.requireOwner(r.rentalHandler.RegenerateTenantPassword))).ServeHTTP))))
	http.HandleFunc("/api/summary", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(r.requireOwner(r.rentalHandler.GetSummary))).ServeHTTP))))
	http.HandleFunc("/api/payments/sync-history", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(r.requireOwner(r.rentalHandler.SyncPaymentHistory))).ServeHTTP))))
	http.HandleFunc("/api/payments/adjust-due-date", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(r.requireOwner(r.rentalHandler.AdjustPaymentDueDate))).ServeHTTP))))
	http.HandleFunc("/api/payments/reject-transaction", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(r.requireOwner(r.rentalHandler.RejectTransaction))).ServeHTTP))))
	// Create payment endpoint (owner only) - POST /api/payments/create
	http.HandleFunc("/api/payments/create", compressionWrapper(metricsWrapper(http.HandlerFunc(recoveryWrapper(correlationWrapper(r.requireOwner(r.rentalHandler.CreatePayment))).ServeHTTP))))
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
	cookieName := r.authHandler.GetCookieName()
	cookie, err := req.Cookie(cookieName)
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
	// Use string key so handlers can access without import cycles
	return context.WithValue(ctx, userContextKeyString, user)
}

// GetUserFromContext retrieves user from context
// Uses the same string key that contextWithUser uses
func GetUserFromContext(ctx context.Context) (*domain.User, bool) {
	user, ok := ctx.Value(userContextKeyString).(*domain.User)
	return user, ok
}

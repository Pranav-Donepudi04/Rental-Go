package main

import (
	"backend-form/m/internal/config"
	"backend-form/m/internal/handlers"
	httplib "backend-form/m/internal/http"
	"backend-form/m/internal/http/middleware"
	"backend-form/m/internal/logger"
	"backend-form/m/internal/repository/interfaces"
	repository "backend-form/m/internal/repository/postgres"
	"backend-form/m/internal/service"
	"context"
	"database/sql"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var templates = template.Must(template.ParseFiles(
	"templates/dashboard.html",
	"templates/unit-detail.html",
	"templates/login.html",
	"templates/tenant-dashboard.html",
))

// App holds all application dependencies
type App struct {
	Config                *config.Config
	DB                    *sql.DB
	Repositories          *Repositories
	Services              *Services
	Handlers              *Handlers
	Router                *httplib.Router
	Server                *http.Server
	NotificationScheduler *service.NotificationScheduler
}

// Repositories holds all repository instances
type Repositories struct {
	Unit         interfaces.UnitRepository
	Tenant       interfaces.TenantRepository
	Payment      interfaces.PaymentRepository
	User         interfaces.UserRepository
	Session      interfaces.SessionRepository
	Notification interfaces.NotificationRepository
}

// Services holds all service instances
type Services struct {
	Unit                  *service.UnitService
	Payment               *service.PaymentService
	PaymentQuery          *service.PaymentQueryService
	PaymentTransaction    *service.PaymentTransactionService
	PaymentHistory        *service.PaymentHistoryService
	Tenant                *service.TenantService
	Auth                  *service.AuthService
	Dashboard             *service.DashboardService
	Notification          *service.NotificationService
	NotificationScheduler *service.NotificationScheduler
}

// Handlers holds all HTTP handler instances
type Handlers struct {
	Auth    *handlers.AuthHandler
	Rental  *handlers.RentalHandler
	Tenant  *handlers.TenantHandler
	Metrics *handlers.MetricsHandler
}

func main() {
	// Initialize configuration
	cfg := initializeConfig()

	// Initialize logger
	initializeLogger(cfg)
	// Don't defer Sync() - call it explicitly before exit to avoid blocking

	// Setup application
	app := setupApplication(cfg)

	// Start server
	startServer(app)

	// Wait for shutdown signal
	waitForShutdown(app)

	// Sync logger before exit (with timeout to avoid blocking)
	syncLoggerWithTimeout()
}

// initializeConfig loads and validates configuration
func initializeConfig() *config.Config {
	config.LoadEnvFile()
	cfg := config.Load()

	if err := cfg.Validate(); err != nil {
		os.Stderr.WriteString("Configuration validation failed: " + err.Error() + "\n")
		os.Exit(1)
	}

	return cfg
}

// initializeLogger sets up structured logging
func initializeLogger(cfg *config.Config) {
	if err := logger.InitLogger(cfg.LogLevel, cfg.Environment); err != nil {
		os.Stderr.WriteString("Failed to initialize logger: " + err.Error() + "\n")
		os.Exit(1)
	}

	logger.Info("Starting Go Backend Server",
		zap.String("port", cfg.Port),
		zap.String("environment", cfg.Environment),
		zap.String("log_level", cfg.LogLevel),
	)

	logger.Info("Server configuration",
		zap.String("db_host", cfg.DBHost),
		zap.String("db_port", cfg.DBPort),
		zap.String("db_name", cfg.DBName),
		zap.String("db_user", cfg.DBUser),
		zap.String("db_ssl_mode", cfg.DBSSLMode),
		zap.Int("max_connections", cfg.MaxConnections),
		zap.Int("connection_timeout", cfg.ConnectionTimeout),
		zap.Bool("database_configured", cfg.DatabaseURL != ""),
	)
}

// setupApplication initializes all application components
func setupApplication(cfg *config.Config) *App {
	db := setupDatabase(cfg)
	repos := setupRepositories(db)
	services := setupServices(cfg, repos)
	handlers := setupHandlers(cfg, services, repos)
	router := setupRouter(cfg, handlers, repos, db)
	server := setupHTTPServer(cfg)
	notificationScheduler := setupNotificationScheduler(cfg, services.Notification)

	return &App{
		Config:                cfg,
		DB:                    db,
		Repositories:          repos,
		Services:              services,
		Handlers:              handlers,
		Router:                router,
		Server:                server,
		NotificationScheduler: notificationScheduler,
	}
}

// setupDatabase initializes and configures the database connection
func setupDatabase(cfg *config.Config) *sql.DB {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database",
			zap.Error(err),
		)
	}

	// Configure connection pool settings
	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.MaxConnections / 2) // Half of max for idle connections
	db.SetConnMaxLifetime(0)                   // Connections don't expire (let Neon handle it)

	// Test the connection
	if err := db.Ping(); err != nil {
		logger.Fatal("Failed to ping database",
			zap.Error(err),
		)
	}

	logger.Info("Database connection established")
	return db
}

// setupRepositories creates all repository instances
func setupRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		Unit:         repository.NewPostgresUnitRepository(db),
		Tenant:       repository.NewPostgresTenantRepository(db),
		Payment:      repository.NewPostgresPaymentRepository(db),
		User:         repository.NewPostgresUserRepository(db),
		Session:      repository.NewPostgresSessionRepository(db),
		Notification: repository.NewPostgresNotificationRepository(db),
	}
}

// setupServices creates all service instances
func setupServices(cfg *config.Config, repos *Repositories) *Services {
	// Note: PaymentService must be created before TenantService since TenantService depends on it
	unitService := service.NewUnitService(repos.Unit)
	paymentService := service.NewPaymentService(repos.Payment, repos.Tenant, repos.Unit, cfg.DefaultPaymentMethod, cfg.DefaultUPIID)
	paymentQueryService := service.NewPaymentQueryService(repos.Payment)
	paymentTransactionService := service.NewPaymentTransactionService(repos.Payment, paymentService)
	paymentHistoryService := service.NewPaymentHistoryService(repos.Payment, repos.Tenant, repos.Unit, paymentService)
	tenantService := service.NewTenantService(repos.Tenant, repos.Unit, paymentService)
	authService := service.NewAuthService(repos.User, repos.Session, 7*24*60*60*1e9)
	dashboardService := service.NewDashboardService(unitService, tenantService, paymentQueryService)

	notificationService := service.NewNotificationService(
		repos.Notification,
		repos.Payment,
		repos.Tenant,
		repos.Unit,
		cfg.TelegramBotToken,
		cfg.OwnerChatID,
	)
	notificationScheduler := service.NewNotificationScheduler(notificationService)

	return &Services{
		Unit:                  unitService,
		Payment:               paymentService,
		PaymentQuery:          paymentQueryService,
		PaymentTransaction:    paymentTransactionService,
		PaymentHistory:        paymentHistoryService,
		Tenant:                tenantService,
		Auth:                  authService,
		Dashboard:             dashboardService,
		Notification:          notificationService,
		NotificationScheduler: notificationScheduler,
	}
}

// setupHandlers creates all HTTP handler instances
func setupHandlers(cfg *config.Config, services *Services, repos *Repositories) *Handlers {
	rentalHandler := handlers.NewRentalHandler(
		services.Unit,
		services.Tenant,
		services.Payment,
		services.PaymentQuery,
		services.PaymentTransaction,
		services.PaymentHistory,
		services.Dashboard,
		services.Notification,
		templates,
		services.Auth,
	)

	authHandler := handlers.NewAuthHandler(
		services.Auth,
		templates,
		cfg.CookieName,
		cfg.CookieSecure,
		cfg.CookieSameSite,
	)

	tenantHandler := handlers.NewTenantHandler(
		services.Tenant,
		services.Payment,
		services.PaymentTransaction,
		repos.User,
		templates,
		cfg.CookieName,
		services.Auth,
	)

	return &Handlers{
		Auth:    authHandler,
		Rental:  rentalHandler,
		Tenant:  tenantHandler,
		Metrics: handlers.NewMetricsHandler(),
	}
}

// setupRouter creates and configures the HTTP router
func setupRouter(cfg *config.Config, handlers *Handlers, repos *Repositories, db *sql.DB) *httplib.Router {
	loginLimiter := middleware.NewRateLimiter(5.0/900.0, 5) // 5 requests per 900 seconds (15 minutes)
	dbHealthCheck := middleware.NewDatabaseHealthCheck(db)

	router := httplib.NewRouter(
		handlers.Auth,
		handlers.Rental,
		handlers.Tenant,
		repos.User,
		loginLimiter,
		dbHealthCheck,
	)
	router.SetUserRepository(repos.User) // Set user repo on rental handler for transaction verification
	router.SetupRoutes()

	return router
}

// setupHTTPServer creates and configures the HTTP server
// Note: Router uses http.HandleFunc which registers on DefaultServeMux,
// so we use nil as Handler to use the default ServeMux
func setupHTTPServer(cfg *config.Config) *http.Server {
	return &http.Server{
		Addr:         ":" + cfg.Port,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
		Handler:      nil, // Uses DefaultServeMux (router registers routes via http.HandleFunc)
	}
}

// setupNotificationScheduler starts the notification scheduler if configured
func setupNotificationScheduler(cfg *config.Config, notificationService *service.NotificationService) *service.NotificationScheduler {
	scheduler := service.NewNotificationScheduler(notificationService)

	if cfg.TelegramBotToken != "" && cfg.OwnerChatID != "" {
		scheduler.Start()
		logger.Info("Notification scheduler started")
	} else {
		logger.Warn("Telegram bot token or owner chat ID not configured. Notifications disabled.")
	}

	return scheduler
}

// startServer starts the HTTP server in a goroutine
func startServer(app *App) {
	go func() {
		logger.Info("Server starting",
			zap.String("port", app.Config.Port),
			zap.String("address", app.Server.Addr),
		)
		if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start",
				zap.Error(err),
			)
		}
	}()
}

// waitForShutdown waits for shutdown signal and performs graceful shutdown
func waitForShutdown(app *App) {
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Stop notification scheduler first (non-blocking)
	if app.Config.TelegramBotToken != "" && app.Config.OwnerChatID != "" {
		app.NotificationScheduler.Stop()
		// Give it a moment to stop, but don't wait too long
		time.Sleep(100 * time.Millisecond)
		logger.Info("Notification scheduler stop signal sent")
	}

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown",
			zap.Error(err),
		)
		// Force close if shutdown times out
		app.Server.Close()
	}

	// Close database connection
	if err := app.DB.Close(); err != nil {
		logger.Error("Error closing database connection",
			zap.Error(err),
		)
	}

	logger.Info("Server stopped gracefully")
}

// syncLoggerWithTimeout syncs the logger with a timeout to avoid blocking
func syncLoggerWithTimeout() {
	done := make(chan bool, 1)
	go func() {
		logger.Sync()
		done <- true
	}()

	select {
	case <-done:
		// Logger synced successfully
	case <-time.After(1 * time.Second):
		// Timeout - logger sync taking too long, exit anyway
	}
}

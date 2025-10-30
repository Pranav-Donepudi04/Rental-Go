package main

import (
	"backend-form/m/internal/config"
	"backend-form/m/internal/handlers"
	httplib "backend-form/m/internal/http"
	repository "backend-form/m/internal/repository/postgres"
	"backend-form/m/internal/service"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var templates = template.Must(template.ParseFiles(
	"templates/dashboard.html",
	"templates/unit-detail.html",
	"templates/login.html",
	"templates/tenant-dashboard.html",
))

func main() {
	fmt.Println("Starting Go Backend Server...")

	// Load configuration
	config.LoadEnvFile()
	cfg := config.Load()

	// Print configuration
	fmt.Printf("Server will start on port: %s\n", cfg.Port)
	fmt.Printf("Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("Database Host: %s\n", cfg.DBHost)
	fmt.Printf("Database Port: %s\n", cfg.DBPort)
	fmt.Printf("Database Name: %s\n", cfg.DBName)
	fmt.Printf("Database User: %s\n", cfg.DBUser)
	fmt.Printf("Database SSL Mode: %s\n", cfg.DBSSLMode)
	fmt.Printf("Max Connections: %d\n", cfg.MaxConnections)
	fmt.Printf("Connection Timeout: %d seconds\n", cfg.ConnectionTimeout)

	// Only print database URL if it's set (for security, don't print full URL with password)
	if cfg.DatabaseURL != "" {
		fmt.Printf("Database URL: [CONFIGURED]\n")
	} else {
		fmt.Printf("Database URL: [NOT SET]\n")
	}
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// DB schema should be provisioned externally (no auto-migrations here)
	// Create rental management repositories
	unitRepo := repository.NewPostgresUnitRepository(db)
	tenantRepo := repository.NewPostgresTenantRepository(db)
	paymentRepo := repository.NewPostgresPaymentRepository(db)
	userRepo := repository.NewPostgresUserRepository(db)
	sessionRepo := repository.NewPostgresSessionRepository(db)

	// Create rental management services
	unitService := service.NewUnitService(unitRepo)
	tenantService := service.NewTenantService(tenantRepo, unitRepo, paymentRepo)
	paymentService := service.NewPaymentService(paymentRepo, tenantRepo, unitRepo)
	authService := service.NewAuthService(userRepo, sessionRepo, 7*24*60*60*1e9)

	// Create rental management handler
	rentalHandler := handlers.NewRentalHandler(unitService, tenantService, paymentService, templates, authService)
	authHandler := handlers.NewAuthHandler(authService, templates, "sid")
	tenantHandler := handlers.NewTenantHandler(tenantService, paymentService, userRepo, templates, "sid", authService)

	// Owner/Tenant users must exist in DB prior to login

	// Create router and set up routes
	router := httplib.NewRouter(authHandler, rentalHandler, tenantHandler, userRepo)
	router.SetupRoutes()

	// Start the server using config
	fmt.Printf("Server starting on port %s...\n", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}

// (schema bootstrap removed intentionally)

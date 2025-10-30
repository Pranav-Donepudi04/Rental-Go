package main

import (
	"backend-form/m/internal/config"
	"backend-form/m/internal/handlers"
	"backend-form/m/internal/repository/postgres"
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
	// Create rental management repositories
	unitRepo := repository.NewPostgresUnitRepository(db)
	tenantRepo := repository.NewPostgresTenantRepository(db)
	paymentRepo := repository.NewPostgresPaymentRepository(db)

	// Create rental management services
	unitService := service.NewUnitService(unitRepo)
	tenantService := service.NewTenantService(tenantRepo, unitRepo, paymentRepo)
	paymentService := service.NewPaymentService(paymentRepo, tenantRepo, unitRepo)

	// Create rental management handler
	rentalHandler := handlers.NewRentalHandler(unitService, tenantService, paymentService, templates)

	// Set up HTTP routes
	http.HandleFunc("/", rentalHandler.Dashboard)
	http.HandleFunc("/dashboard", rentalHandler.Dashboard)
	http.HandleFunc("/unit/", rentalHandler.UnitDetails)

	// API routes
	http.HandleFunc("/api/units", rentalHandler.GetUnits)
	http.HandleFunc("/api/tenants", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			rentalHandler.GetTenants(w, r)
		} else if r.Method == "POST" {
			rentalHandler.CreateTenant(w, r)
		}
	})
	http.HandleFunc("/api/payments", rentalHandler.GetPayments)
	http.HandleFunc("/api/payments/mark-paid", rentalHandler.MarkPaymentAsPaid)
	http.HandleFunc("/api/tenants/vacate", rentalHandler.VacateTenant)
	http.HandleFunc("/api/summary", rentalHandler.GetSummary)

	// Start the server using config
	fmt.Printf("Server starting on port %s...\n", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}

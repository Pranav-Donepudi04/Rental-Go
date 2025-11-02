package main

import (
	"backend-form/m/internal/config"
	repository "backend-form/m/internal/repository/postgres"
	"backend-form/m/internal/service"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("🏠 Testing Rental Management System...")

	// Load configuration
	config.LoadEnvFile()
	cfg := config.Load()

	// Connect to database
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	fmt.Println("✅ Database connection successful!")

	// Create repositories
	unitRepo := repository.NewPostgresUnitRepository(db)
	tenantRepo := repository.NewPostgresTenantRepository(db)
	paymentRepo := repository.NewPostgresPaymentRepository(db)

	// Create services
	unitService := service.NewUnitService(unitRepo)
	paymentService := service.NewPaymentService(paymentRepo, tenantRepo, unitRepo)
	tenantService := service.NewTenantService(tenantRepo, unitRepo, paymentService)

	// Test 1: Get all units
	fmt.Println("\n📋 Test 1: Getting all units...")
	units, err := unitService.GetAllUnits()
	if err != nil {
		log.Printf("❌ Failed to get units: %v", err)
	} else {
		fmt.Printf("✅ Found %d units:\n", len(units))
		for _, unit := range units {
			fmt.Printf("   - %s: %s (₹%d/month, Due: %dth) - %s\n",
				unit.UnitCode, unit.UnitType, unit.MonthlyRent,
				unit.PaymentDueDay, getOccupancyStatus(unit.IsOccupied))
		}
	}

	// Test 2: Get rental summary
	fmt.Println("\n💰 Test 2: Getting rental summary...")
	summary, err := unitService.GetRentalSummary()
	if err != nil {
		log.Printf("❌ Failed to get rental summary: %v", err)
	} else {
		fmt.Printf("✅ Rental Summary:\n")
		fmt.Printf("   - Total Units: %d\n", summary.TotalUnits)
		fmt.Printf("   - Occupied Units: %d\n", summary.OccupiedUnits)
		fmt.Printf("   - Available Units: %d\n", summary.AvailableUnits)
		fmt.Printf("   - Occupied Rent: %s\n", summary.GetFormattedOccupiedRent())
		fmt.Printf("   - Total Rent: %s\n", summary.GetFormattedTotalRent())
	}

	// Test 3: Get all tenants
	fmt.Println("\n👥 Test 3: Getting all tenants...")
	tenants, err := tenantService.GetAllTenants()
	if err != nil {
		log.Printf("❌ Failed to get tenants: %v", err)
	} else {
		fmt.Printf("✅ Found %d tenants:\n", len(tenants))
		for _, tenant := range tenants {
			fmt.Printf("   - %s (%s) - Unit ID: %d\n", tenant.Name, tenant.Phone, tenant.UnitID)
		}
	}

	// Test 4: Get all payments
	fmt.Println("\n💳 Test 4: Getting all payments...")
	payments, err := paymentService.GetAllPayments()
	if err != nil {
		log.Printf("❌ Failed to get payments: %v", err)
	} else {
		fmt.Printf("✅ Found %d payments:\n", len(payments))
		for _, payment := range payments {
			fmt.Printf("   - Payment #%d: %s (Tenant ID: %d) - %s\n",
				payment.ID, payment.GetFormattedAmount(), payment.TenantID, payment.GetStatus())
		}
	}

	// Test 5: Get payment summary
	fmt.Println("\n📊 Test 5: Getting payment summary...")
	paymentSummary, err := paymentService.GetPaymentSummary()
	if err != nil {
		log.Printf("❌ Failed to get payment summary: %v", err)
	} else {
		fmt.Printf("✅ Payment Summary:\n")
		fmt.Printf("   - Total Payments: %d\n", paymentSummary.TotalPayments)
		fmt.Printf("   - Paid Payments: %d\n", paymentSummary.PaidPayments)
		fmt.Printf("   - Pending Payments: %d\n", paymentSummary.PendingPayments)
		fmt.Printf("   - Overdue Payments: %d\n", paymentSummary.OverduePayments)
		fmt.Printf("   - Total Amount: %s\n", paymentSummary.GetFormattedTotalAmount())
		fmt.Printf("   - Paid Amount: %s\n", paymentSummary.GetFormattedPaidAmount())
		fmt.Printf("   - Pending Amount: %s\n", paymentSummary.GetFormattedPendingAmount())
	}

	fmt.Println("\n🎉 All tests completed!")
}

func getOccupancyStatus(isOccupied bool) string {
	if isOccupied {
		return "Occupied"
	}
	return "Available"
}

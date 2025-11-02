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
	fmt.Println("======================================")

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

	// Create repositories (matching main.go structure)
	fmt.Println("\n📦 Initializing repositories...")
	unitRepo := repository.NewPostgresUnitRepository(db)
	tenantRepo := repository.NewPostgresTenantRepository(db)
	paymentRepo := repository.NewPostgresPaymentRepository(db)
	userRepo := repository.NewPostgresUserRepository(db)
	sessionRepo := repository.NewPostgresSessionRepository(db)
	fmt.Println("✅ All repositories initialized")

	// Create services (matching main.go structure and order)
	fmt.Println("\n⚙️  Initializing services...")
	unitService := service.NewUnitService(unitRepo)
	paymentService := service.NewPaymentService(paymentRepo, tenantRepo, unitRepo)
	paymentQueryService := service.NewPaymentQueryService(paymentRepo)
	paymentTransactionService := service.NewPaymentTransactionService(paymentRepo, paymentService)
	paymentHistoryService := service.NewPaymentHistoryService(paymentRepo, tenantRepo, unitRepo, paymentService)
	_ = paymentHistoryService // Keep for completeness (matches main.go structure)
	tenantService := service.NewTenantService(tenantRepo, unitRepo, paymentService)
	authService := service.NewAuthService(userRepo, sessionRepo, 7*24*60*60*1e9)
	dashboardService := service.NewDashboardService(unitService, tenantService, paymentQueryService)
	fmt.Println("✅ All services initialized")

	fmt.Println("\n🧪 Running Tests...")
	fmt.Println("===================")

	// ==================== UNIT SERVICE TESTS ====================
	testUnitService(unitService)

	// ==================== TENANT SERVICE TESTS ====================
	testTenantService(tenantService)

	// ==================== PAYMENT QUERY SERVICE TESTS ====================
	testPaymentQueryService(paymentQueryService)

	// ==================== PAYMENT SERVICE TESTS ====================
	testPaymentService(paymentService, paymentQueryService)

	// ==================== PAYMENT TRANSACTION SERVICE TESTS ====================
	testPaymentTransactionService(paymentTransactionService)

	// ==================== DASHBOARD SERVICE TESTS ====================
	testDashboardService(dashboardService)

	// ==================== AUTH SERVICE TESTS ====================
	testAuthService(authService)

	fmt.Println("\n🎉 All tests completed!")
}

// ==================== UNIT SERVICE TESTS ====================
func testUnitService(unitService *service.UnitService) {
	fmt.Println("\n📋 UNIT SERVICE TESTS")
	fmt.Println("--------------------")

	// Test 1: Get all units
	fmt.Println("\n  Test 1.1: Getting all units...")
	units, err := unitService.GetAllUnits()
	if err != nil {
		log.Printf("    ❌ Failed: %v", err)
	} else {
		fmt.Printf("    ✅ Found %d units\n", len(units))
		if len(units) > 0 {
			for i, unit := range units {
				if i < 3 { // Show first 3
					fmt.Printf("      - %s: %s (₹%d/month, Due: %dth) - %s\n",
						unit.UnitCode, unit.UnitType, unit.MonthlyRent,
						unit.PaymentDueDay, getOccupancyStatus(unit.IsOccupied))
				}
			}
			if len(units) > 3 {
				fmt.Printf("      ... and %d more units\n", len(units)-3)
			}
		}
	}

	// Test 2: Get rental summary
	fmt.Println("\n  Test 1.2: Getting rental summary...")
	summary, err := unitService.GetRentalSummary()
	if err != nil {
		log.Printf("    ❌ Failed: %v", err)
	} else {
		fmt.Printf("    ✅ Rental Summary:\n")
		fmt.Printf("      - Total Units: %d\n", summary.TotalUnits)
		fmt.Printf("      - Occupied Units: %d\n", summary.OccupiedUnits)
		fmt.Printf("      - Available Units: %d\n", summary.AvailableUnits)
		fmt.Printf("      - Occupied Rent: %s\n", summary.GetFormattedOccupiedRent())
		fmt.Printf("      - Total Rent: %s\n", summary.GetFormattedTotalRent())
	}

	// Test 3: Get available units
	fmt.Println("\n  Test 1.3: Getting available units...")
	available, err := unitService.GetAvailableUnits()
	if err != nil {
		log.Printf("    ❌ Failed: %v", err)
	} else {
		fmt.Printf("    ✅ Found %d available units\n", len(available))
	}

	// Test 4: Get occupied units
	fmt.Println("\n  Test 1.4: Getting occupied units...")
	occupied, err := unitService.GetOccupiedUnits()
	if err != nil {
		log.Printf("    ❌ Failed: %v", err)
	} else {
		fmt.Printf("    ✅ Found %d occupied units\n", len(occupied))
	}

	// Test 5: Get units by floor
	fmt.Println("\n  Test 1.5: Getting units by floor...")
	floorMap, err := unitService.GetUnitsByFloor()
	if err != nil {
		log.Printf("    ❌ Failed: %v", err)
	} else {
		fmt.Printf("    ✅ Found %d floors\n", len(floorMap))
		for floor, floorUnits := range floorMap {
			fmt.Printf("      - %s: %d units\n", floor, len(floorUnits))
		}
	}

	// Test 6: Get unit by ID (if units exist)
	if len(units) > 0 {
		fmt.Println("\n  Test 1.6: Getting unit by ID...")
		unit, err := unitService.GetUnitByID(units[0].ID)
		if err != nil {
			log.Printf("    ❌ Failed: %v", err)
		} else {
			fmt.Printf("    ✅ Found unit: %s\n", unit.UnitCode)
		}

		// Test 7: Get unit by code
		fmt.Println("\n  Test 1.7: Getting unit by code...")
		unitByCode, err := unitService.GetUnitByCode(units[0].UnitCode)
		if err != nil {
			log.Printf("    ❌ Failed: %v", err)
		} else {
			fmt.Printf("    ✅ Found unit: %s\n", unitByCode.UnitCode)
		}
	}
}

// ==================== TENANT SERVICE TESTS ====================
func testTenantService(tenantService *service.TenantService) {
	fmt.Println("\n👥 TENANT SERVICE TESTS")
	fmt.Println("----------------------")

	// Test 1: Get all tenants
	fmt.Println("\n  Test 2.1: Getting all tenants...")
	tenants, err := tenantService.GetAllTenants()
	if err != nil {
		log.Printf("    ❌ Failed: %v", err)
	} else {
		fmt.Printf("    ✅ Found %d tenants\n", len(tenants))
		if len(tenants) > 0 {
			for i, tenant := range tenants {
				if i < 3 { // Show first 3
					fmt.Printf("      - %s (%s) - Unit ID: %d\n", tenant.Name, tenant.Phone, tenant.UnitID)
				}
			}
			if len(tenants) > 3 {
				fmt.Printf("      ... and %d more tenants\n", len(tenants)-3)
			}
		}
	}

	// Test 2: Get tenant summary
	fmt.Println("\n  Test 2.2: Getting tenant summary...")
	tenantSummary, err := tenantService.GetTenantSummary()
	if err != nil {
		log.Printf("    ❌ Failed: %v", err)
	} else {
		fmt.Printf("    ✅ Tenant Summary:\n")
		fmt.Printf("      - Total Tenants: %d\n", tenantSummary.TotalTenants)
		fmt.Printf("      - New This Month: %d\n", tenantSummary.NewThisMonth)
		fmt.Printf("      - Total People: %d\n", tenantSummary.TotalPeople)
	}

	// Test 3: Get tenant by ID (if tenants exist)
	if len(tenants) > 0 {
		fmt.Println("\n  Test 2.3: Getting tenant by ID...")
		tenant, err := tenantService.GetTenantByID(tenants[0].ID)
		if err != nil {
			log.Printf("    ❌ Failed: %v", err)
		} else {
			fmt.Printf("    ✅ Found tenant: %s\n", tenant.Name)
		}

		// Test 4: Get family members
		fmt.Println("\n  Test 2.4: Getting family members...")
		familyMembers, err := tenantService.GetFamilyMembersByTenantID(tenants[0].ID)
		if err != nil {
			log.Printf("    ❌ Failed: %v", err)
		} else {
			fmt.Printf("    ✅ Found %d family members\n", len(familyMembers))
		}

		// Test 5: Get tenants by unit ID
		fmt.Println("\n  Test 2.5: Getting tenants by unit ID...")
		unitTenants, err := tenantService.GetTenantsByUnitID(tenants[0].UnitID)
		if err != nil {
			log.Printf("    ❌ Failed: %v", err)
		} else {
			fmt.Printf("    ✅ Found %d tenants for unit ID %d\n", len(unitTenants), tenants[0].UnitID)
		}
	}
}

// ==================== PAYMENT QUERY SERVICE TESTS ====================
func testPaymentQueryService(paymentQueryService *service.PaymentQueryService) {
	fmt.Println("\n💳 PAYMENT QUERY SERVICE TESTS")
	fmt.Println("------------------------------")

	// Test 1: Get all payments
	fmt.Println("\n  Test 3.1: Getting all payments...")
	payments, err := paymentQueryService.GetAllPayments()
	if err != nil {
		log.Printf("    ❌ Failed: %v", err)
	} else {
		fmt.Printf("    ✅ Found %d payments\n", len(payments))
		if len(payments) > 0 {
			for i, payment := range payments {
				if i < 3 { // Show first 3
					fmt.Printf("      - Payment #%d: %s (Tenant ID: %d) - %s\n",
						payment.ID, payment.GetFormattedAmount(), payment.TenantID, payment.GetStatus())
				}
			}
			if len(payments) > 3 {
				fmt.Printf("      ... and %d more payments\n", len(payments)-3)
			}
		}
	}

	// Test 2: Get payment summary
	fmt.Println("\n  Test 3.2: Getting payment summary...")
	paymentSummary, err := paymentQueryService.GetPaymentSummary()
	if err != nil {
		log.Printf("    ❌ Failed: %v", err)
	} else {
		fmt.Printf("    ✅ Payment Summary:\n")
		fmt.Printf("      - Total Payments: %d\n", paymentSummary.TotalPayments)
		fmt.Printf("      - Paid Payments: %d\n", paymentSummary.PaidPayments)
		fmt.Printf("      - Pending Payments: %d\n", paymentSummary.PendingPayments)
		fmt.Printf("      - Overdue Payments: %d\n", paymentSummary.OverduePayments)
		fmt.Printf("      - Total Amount: %s\n", paymentSummary.GetFormattedTotalAmount())
		fmt.Printf("      - Paid Amount: %s\n", paymentSummary.GetFormattedPaidAmount())
		fmt.Printf("      - Pending Amount: %s\n", paymentSummary.GetFormattedPendingAmount())
		if paymentSummary.OverdueAmount > 0 {
			fmt.Printf("      - Overdue Amount: %s\n", paymentSummary.GetFormattedOverdueAmount())
		}
	}

	// Test 3: Get overdue payments
	fmt.Println("\n  Test 3.3: Getting overdue payments...")
	overdue, err := paymentQueryService.GetOverduePayments()
	if err != nil {
		log.Printf("    ❌ Failed: %v", err)
	} else {
		fmt.Printf("    ✅ Found %d overdue payments\n", len(overdue))
	}

	// Test 4: Get pending payments
	fmt.Println("\n  Test 3.4: Getting pending payments...")
	pending, err := paymentQueryService.GetPendingPayments()
	if err != nil {
		log.Printf("    ❌ Failed: %v", err)
	} else {
		fmt.Printf("    ✅ Found %d pending payments\n", len(pending))
	}

	// Test 5: Get unpaid payments by tenant ID (if tenants exist)
	if len(payments) > 0 {
		fmt.Println("\n  Test 3.5: Getting unpaid payments by tenant ID...")
		unpaid, err := paymentQueryService.GetUnpaidPaymentsByTenantID(payments[0].TenantID)
		if err != nil {
			log.Printf("    ❌ Failed: %v", err)
		} else {
			fmt.Printf("    ✅ Found %d unpaid payments for tenant ID %d\n", len(unpaid), payments[0].TenantID)
		}
	}
}

// ==================== PAYMENT SERVICE TESTS ====================
func testPaymentService(paymentService *service.PaymentService, paymentQueryService *service.PaymentQueryService) {
	fmt.Println("\n💰 PAYMENT SERVICE TESTS")
	fmt.Println("------------------------")

	// Test 1: Get payment by ID (if payments exist)
	fmt.Println("\n  Test 4.1: Getting payment by ID...")
	// Get payments via PaymentQueryService
	allPayments, err := paymentQueryService.GetAllPayments()
	if err == nil && len(allPayments) > 0 {
		payment, err := paymentService.GetPaymentByID(allPayments[0].ID)
		if err != nil {
			log.Printf("    ❌ Failed: %v", err)
		} else {
			fmt.Printf("    ✅ Found payment #%d: %s\n", payment.ID, payment.GetFormattedAmount())
		}

		// Test 2: Get payments by tenant ID
		fmt.Println("\n  Test 4.2: Getting payments by tenant ID...")
		tenantPayments, err := paymentService.GetPaymentsByTenantID(allPayments[0].TenantID)
		if err != nil {
			log.Printf("    ❌ Failed: %v", err)
		} else {
			fmt.Printf("    ✅ Found %d payments for tenant ID %d\n", len(tenantPayments), allPayments[0].TenantID)
		}

		// Test 3: Get unpaid payments by tenant ID
		fmt.Println("\n  Test 4.3: Getting unpaid payments by tenant ID...")
		unpaidPayments, err := paymentService.GetUnpaidPaymentsByTenantID(allPayments[0].TenantID)
		if err != nil {
			log.Printf("    ❌ Failed: %v", err)
		} else {
			fmt.Printf("    ✅ Found %d unpaid payments for tenant ID %d\n", len(unpaidPayments), allPayments[0].TenantID)
		}
	} else {
		fmt.Println("    ⚠️  Skipped: No payments found in database")
	}
}

// ==================== PAYMENT TRANSACTION SERVICE TESTS ====================
func testPaymentTransactionService(paymentTransactionService *service.PaymentTransactionService) {
	fmt.Println("\n🔔 PAYMENT TRANSACTION SERVICE TESTS")
	fmt.Println("-----------------------------------")

	// Test 1: Get pending verifications (for all tenants)
	fmt.Println("\n  Test 5.1: Getting pending verifications...")
	pendingVerifications, err := paymentTransactionService.GetPendingVerifications(0) // 0 means all tenants
	if err != nil {
		log.Printf("    ❌ Failed: %v", err)
	} else {
		fmt.Printf("    ✅ Found %d pending verifications\n", len(pendingVerifications))
		if len(pendingVerifications) > 0 {
			for i, txn := range pendingVerifications {
				if i < 3 { // Show first 3
					amountStr := "Not verified"
					if txn.Amount != nil {
						amountStr = fmt.Sprintf("₹%d", *txn.Amount)
					}
					tenantIDStr := "N/A"
					if txn.Payment != nil {
						tenantIDStr = fmt.Sprintf("%d", txn.Payment.TenantID)
					}
					fmt.Printf("      - Transaction %s: %s (Payment ID: %d, Tenant ID: %s)\n",
						txn.TransactionID, amountStr, txn.PaymentID, tenantIDStr)
				}
			}
			if len(pendingVerifications) > 3 {
				fmt.Printf("      ... and %d more transactions\n", len(pendingVerifications)-3)
			}
		}
	}
}

// ==================== DASHBOARD SERVICE TESTS ====================
func testDashboardService(dashboardService *service.DashboardService) {
	fmt.Println("\n📊 DASHBOARD SERVICE TESTS")
	fmt.Println("--------------------------")

	// Test 1: Get dashboard data
	fmt.Println("\n  Test 6.1: Getting dashboard data...")
	dashboardData, err := dashboardService.GetDashboardData()
	if err != nil {
		log.Printf("    ❌ Failed: %v", err)
	} else {
		fmt.Printf("    ✅ Dashboard Data:\n")
		fmt.Printf("      - Units: %d\n", len(dashboardData.Units))
		fmt.Printf("      - Tenants: %d\n", len(dashboardData.Tenants))
		fmt.Printf("      - Payments: %d\n", len(dashboardData.Payments))
		if dashboardData.UnitSummary != nil {
			fmt.Printf("      - Unit Summary: %d total, %d occupied\n",
				dashboardData.UnitSummary.TotalUnits, dashboardData.UnitSummary.OccupiedUnits)
		}
		if dashboardData.PaymentSummary != nil {
			fmt.Printf("      - Payment Summary: %d total, %d paid, %d pending\n",
				dashboardData.PaymentSummary.TotalPayments,
				dashboardData.PaymentSummary.PaidPayments,
				dashboardData.PaymentSummary.PendingPayments)
		}
	}

	// Test 2: Get dashboard summary
	fmt.Println("\n  Test 6.2: Getting dashboard summary...")
	dashboardSummary, err := dashboardService.GetDashboardSummary()
	if err != nil {
		log.Printf("    ❌ Failed: %v", err)
	} else {
		fmt.Printf("    ✅ Dashboard Summary retrieved successfully\n")
		if dashboardSummary.Units != nil {
			fmt.Printf("      - Total Units: %d, Occupied: %d\n",
				dashboardSummary.Units.TotalUnits, dashboardSummary.Units.OccupiedUnits)
		}
		if dashboardSummary.Payments != nil {
			fmt.Printf("      - Total Payments: %d, Paid: %d, Pending: %d\n",
				dashboardSummary.Payments.TotalPayments,
				dashboardSummary.Payments.PaidPayments,
				dashboardSummary.Payments.PendingPayments)
		}
	}
}

// ==================== AUTH SERVICE TESTS ====================
func testAuthService(authService *service.AuthService) {
	fmt.Println("\n🔐 AUTH SERVICE TESTS")
	fmt.Println("--------------------")

	// Test 1: Hash password
	fmt.Println("\n  Test 7.1: Hashing password...")
	hashed := authService.HashPassword("testpassword123")
	if hashed != "" {
		fmt.Println("    ✅ Password hashed successfully")
	} else {
		fmt.Println("    ❌ Password hashing failed")
	}

	// Test 2: Compare password
	fmt.Println("\n  Test 7.2: Comparing password...")
	isValid := authService.ComparePassword(hashed, "testpassword123")
	if isValid {
		fmt.Println("    ✅ Password comparison successful")
	} else {
		fmt.Println("    ❌ Password comparison failed")
	}

	// Test 3: Generate temp password
	fmt.Println("\n  Test 7.3: Generating temporary password...")
	tempPassword, err := authService.GenerateTempPassword()
	if err != nil {
		log.Printf("    ❌ Failed: %v", err)
	} else {
		fmt.Printf("    ✅ Generated temp password: %s\n", tempPassword)
	}
}

// Helper function
func getOccupancyStatus(isOccupied bool) string {
	if isOccupied {
		return "Occupied"
	}
	return "Available"
}

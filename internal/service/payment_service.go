package service

import (
	"backend-form/m/internal/domain"
	interfaces "backend-form/m/internal/repository/interfaces"
	"fmt"
	"time"
)

// PaymentService handles payment-related business logic
type PaymentService struct {
	paymentRepo interfaces.PaymentRepository
	tenantRepo  interfaces.TenantRepository
	unitRepo    interfaces.UnitRepository
}

// NewPaymentService creates a new PaymentService
func NewPaymentService(paymentRepo interfaces.PaymentRepository, tenantRepo interfaces.TenantRepository, unitRepo interfaces.UnitRepository) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		tenantRepo:  tenantRepo,
		unitRepo:    unitRepo,
	}
}

// CreateMonthlyPayment creates a monthly payment record for a tenant
func (s *PaymentService) CreateMonthlyPayment(tenantID int, month time.Month, year int) (*domain.Payment, error) {
	// Get tenant and unit information
	tenant, err := s.tenantRepo.GetTenantByID(tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	unit, err := s.unitRepo.GetUnitByID(tenant.UnitID)
	if err != nil {
		return nil, fmt.Errorf("unit not found: %w", err)
	}

	// Calculate due date
	dueDate := time.Date(year, month, unit.PaymentDueDay, 0, 0, 0, 0, time.UTC)

	// Check if payment already exists for this month
	existingPayment, err := s.paymentRepo.GetPaymentByTenantAndMonth(tenantID, month, year)
	if err == nil && existingPayment != nil {
		return nil, fmt.Errorf("payment already exists for %s %d", month.String(), year)
	}

	// Create payment record
	payment := &domain.Payment{
		TenantID:      tenantID,
		UnitID:        tenant.UnitID,
		Amount:        unit.MonthlyRent,
		DueDate:       dueDate,
		IsPaid:        false,
		PaymentMethod: "UPI",
		UPIID:         "9848790200@ybl",
	}

	if err := s.paymentRepo.CreatePayment(payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return payment, nil
}

// MarkPaymentAsPaid marks a payment as paid (legacy method - kept for backward compatibility)
// For new partial payment flow, use VerifyTransaction instead
func (s *PaymentService) MarkPaymentAsPaid(paymentID int, paymentDate time.Time, notes string) error {
	payment, err := s.paymentRepo.GetPaymentByID(paymentID)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	if payment.IsFullyPaid {
		return fmt.Errorf("payment is already marked as paid")
	}

	// Update to fully paid
	payment.IsPaid = true
	payment.IsFullyPaid = true
	payment.AmountPaid = payment.Amount
	payment.RemainingBalance = 0
	payment.PaymentDate = &paymentDate
	payment.FullyPaidDate = &paymentDate
	payment.Notes = notes

	if err := s.paymentRepo.UpdatePayment(payment); err != nil {
		return err
	}

	// Auto-create next payment when fully paid
	return s.AutoCreateNextPayment(payment)
}

// GetPaymentByID returns a payment by ID with related data and transactions
func (s *PaymentService) GetPaymentByID(id int) (*domain.Payment, error) {
	payment, err := s.paymentRepo.GetPaymentByID(id)
	if err != nil {
		return nil, err
	}

	// Load related data
	if payment.TenantID > 0 {
		tenant, err := s.tenantRepo.GetTenantByID(payment.TenantID)
		if err == nil {
			payment.Tenant = tenant
		}
	}

	if payment.UnitID > 0 {
		unit, err := s.unitRepo.GetUnitByID(payment.UnitID)
		if err == nil {
			payment.Unit = unit
		}
	}

	// Load transactions for status calculation
	s.loadPaymentTransactions(payment)

	return payment, nil
}

// loadPaymentTransactions loads transactions for a payment
func (s *PaymentService) loadPaymentTransactions(payment *domain.Payment) {
	transactions, _ := s.paymentRepo.GetPaymentTransactionsByPaymentID(payment.ID)
	payment.Transactions = transactions
}

// GetPaymentsByTenantID returns all payments for a tenant with transactions loaded
func (s *PaymentService) GetPaymentsByTenantID(tenantID int) ([]*domain.Payment, error) {
	payments, err := s.paymentRepo.GetPaymentsByTenantID(tenantID)
	if err != nil {
		return nil, err
	}
	// Load transactions for each payment so domain can calculate status properly
	for _, payment := range payments {
		s.loadPaymentTransactions(payment)
	}
	return payments, nil
}

// GetUnpaidPaymentsByTenantID returns unpaid payments for a tenant
func (s *PaymentService) GetUnpaidPaymentsByTenantID(tenantID int) ([]*domain.Payment, error) {
	return s.paymentRepo.GetUnpaidPaymentsByTenantID(tenantID)
}

// DeletePayment deletes a payment by ID
func (s *PaymentService) DeletePayment(id int) error {
	return s.paymentRepo.DeletePayment(id)
}

// DeletePaymentsByTenantID deletes all payments for a tenant
func (s *PaymentService) DeletePaymentsByTenantID(tenantID int) error {
	return s.paymentRepo.DeletePaymentsByTenantID(tenantID)
}

// ============================================
// Payment Status & Queries
// ============================================
// NOTE: Query methods have been moved to PaymentQueryService.
// These methods are kept for backward compatibility and forward to PaymentQueryService.
// They will be removed in a future version.

// PaymentQueryService is injected for backward compatibility
// TODO: Remove this after all handlers are migrated
// var globalPaymentQueryService *PaymentQueryService

// SetPaymentQueryService sets the global query service for backward compatibility
// This is a temporary solution during migration
// func SetGlobalPaymentQueryService(queryService *PaymentQueryService) {
// 	globalPaymentQueryService = queryService
// }

// GetOverduePayments returns all overdue payments (DEPRECATED - use PaymentQueryService)
// func (s *PaymentService) GetOverduePayments() ([]*domain.Payment, error) {
// 	if globalPaymentQueryService != nil {
// 		return globalPaymentQueryService.GetOverduePayments()
// 	}
// 	// Fallback implementation if query service not set
// 	payments, err := s.paymentRepo.GetAllPayments()
// 	if err != nil {
// 		return nil, err
// 	}
// 	var overdue []*domain.Payment
// 	now := time.Now()
// 	for _, payment := range payments {
// 		if !payment.IsFullyPaid && now.After(payment.DueDate) {
// 			overdue = append(overdue, payment)
// 		}
// 	}
// 	return overdue, nil
// }

// GetPendingPayments returns all pending payments (DEPRECATED - use PaymentQueryService)
// func (s *PaymentService) GetPendingPayments() ([]*domain.Payment, error) {
// 	if globalPaymentQueryService != nil {
// 		return globalPaymentQueryService.GetPendingPayments()
// 	}
// 	// Fallback implementation
// 	payments, err := s.paymentRepo.GetAllPayments()
// 	if err != nil {
// 		return nil, err
// 	}
// 	var pending []*domain.Payment
// 	now := time.Now()
// 	for _, payment := range payments {
// 		if !payment.IsFullyPaid && now.Before(payment.DueDate) {
// 			pending = append(pending, payment)
// 		}
// 	}
// 	return pending, nil
// }

// GetPaymentSummary returns a summary of payments (DEPRECATED - use PaymentQueryService)
// func (s *PaymentService) GetPaymentSummary() (*PaymentSummary, error) {
// 	if globalPaymentQueryService != nil {
// 		return globalPaymentQueryService.GetPaymentSummary()
// 	}
// 	// Fallback implementation
// 	payments, err := s.paymentRepo.GetAllPayments()
// 	if err != nil {
// 		return nil, err
// 	}
// 	summary := &PaymentSummary{
// 		TotalPayments:   len(payments),
// 		PaidPayments:    0,
// 		PendingPayments: 0,
// 		OverduePayments: 0,
// 		TotalAmount:     0,
// 		PaidAmount:      0,
// 		PendingAmount:   0,
// 		OverdueAmount:   0,
// 	}
// 	now := time.Now()
// 	for _, payment := range payments {
// 		summary.TotalAmount += payment.Amount
// 		if payment.IsFullyPaid {
// 			summary.PaidPayments++
// 			summary.PaidAmount += payment.Amount
// 		} else if now.After(payment.DueDate) {
// 			summary.OverduePayments++
// 			summary.OverdueAmount += payment.RemainingBalance
// 		} else {
// 			summary.PendingPayments++
// 			summary.PendingAmount += payment.RemainingBalance
// 		}
// 	}
// 	return summary, nil
// }

// PaymentSummary type is defined in payment_query_service.go
// Since both services are in the same package, the type is accessible here

// ============================================
// Payment CRUD Operations
// ============================================

// GetAllPayments returns all payments (DEPRECATED - use PaymentQueryService)
// func (s *PaymentService) GetAllPayments() ([]*domain.Payment, error) {
// 	if globalPaymentQueryService != nil {
// 		return globalPaymentQueryService.GetAllPayments()
// 	}
// 	return s.paymentRepo.GetAllPayments()
// }

// ============================================
// Transaction Management
// ============================================
// NOTE: Transaction methods have been moved to PaymentTransactionService.
// These methods are kept for backward compatibility and forward to PaymentTransactionService.
// They will be removed in a future version.

// globalPaymentTransactionService is injected for backward compatibility
// TODO: Remove this after all handlers are migrated
// var globalPaymentTransactionService *PaymentTransactionService

// SetGlobalPaymentTransactionService sets the global transaction service for backward compatibility
// func SetGlobalPaymentTransactionService(transactionService *PaymentTransactionService) {
// 	globalPaymentTransactionService = transactionService
// }

// GetPendingVerifications returns all pending verifications for a tenant (DEPRECATED - use PaymentTransactionService)
// func (s *PaymentService) GetPendingVerifications(tenantID int) ([]*domain.PaymentTransaction, error) {
// 	if globalPaymentTransactionService != nil {
// 		return globalPaymentTransactionService.GetPendingVerifications(tenantID)
// 	}
// 	// Fallback implementation
// 	return s.paymentRepo.GetPendingVerifications(tenantID)
// }

// SubmitPaymentIntent creates a payment transaction record for a tenant (DEPRECATED - use PaymentTransactionService)
// func (s *PaymentService) SubmitPaymentIntent(tenantID int, txnID string) error {
// 	if globalPaymentTransactionService != nil {
// 		return globalPaymentTransactionService.SubmitPaymentIntent(tenantID, txnID)
// 	}
// 	// Fallback implementation
// 	payment, err := s.getOrCreateCurrentPayment(tenantID)
// 	if err != nil {
// 		return fmt.Errorf("get or create payment: %w", err)
// 	}
// 	existing, err := s.paymentRepo.GetTransactionByPaymentAndID(payment.ID, txnID)
// 	if err != nil {
// 		return fmt.Errorf("check existing transaction: %w", err)
// 	}
// 	if existing != nil {
// 		return nil
// 	}
// 	tx := &domain.PaymentTransaction{
// 		PaymentID:     payment.ID,
// 		TransactionID: txnID,
// 		Amount:        nil,
// 		SubmittedAt:   time.Now(),
// 	}
// 	if err := s.paymentRepo.CreatePaymentTransaction(tx); err != nil {
// 		return fmt.Errorf("create payment transaction: %w", err)
// 	}
// 	return nil
// }

// VerifyTransaction verifies a transaction (DEPRECATED - use PaymentTransactionService)
// func (s *PaymentService) VerifyTransaction(transactionID string, amount int, verifiedByUserID int) error {
// 	if globalPaymentTransactionService != nil {
// 		return globalPaymentTransactionService.VerifyTransaction(transactionID, amount, verifiedByUserID)
// 	}
// 	// Fallback implementation
// 	tx, err := s.paymentRepo.GetTransactionByID(transactionID)
// 	if err != nil {
// 		return fmt.Errorf("get transaction: %w", err)
// 	}
// 	if tx == nil {
// 		return fmt.Errorf("transaction not found")
// 	}
// 	if err := s.paymentRepo.VerifyTransaction(transactionID, amount, verifiedByUserID); err != nil {
// 		return fmt.Errorf("verify transaction: %w", err)
// 	}
// 	payment, err := s.paymentRepo.GetPaymentByID(tx.PaymentID)
// 	if err != nil {
// 		return fmt.Errorf("get updated payment: %w", err)
// 	}
// 	if payment.IsFullyPaid {
// 		return s.AutoCreateNextPayment(payment)
// 	}
// 	return nil
// }

// RejectTransaction rejects a pending transaction (DEPRECATED - use PaymentTransactionService)
// func (s *PaymentService) RejectTransaction(transactionID string) error {
// 	if globalPaymentTransactionService != nil {
// 		return globalPaymentTransactionService.RejectTransaction(transactionID)
// 	}
// 	// Fallback implementation
// 	tx, err := s.paymentRepo.GetTransactionByID(transactionID)
// 	if err != nil {
// 		return fmt.Errorf("get transaction: %w", err)
// 	}
// 	if tx == nil {
// 		return fmt.Errorf("transaction not found")
// 	}
// 	return s.paymentRepo.RejectTransaction(transactionID)
// }

// getOrCreateCurrentPayment gets the current unpaid payment or creates a new one
func (s *PaymentService) getOrCreateCurrentPayment(tenantID int) (*domain.Payment, error) {
	// Get unpaid payments
	unpaid, err := s.paymentRepo.GetUnpaidPaymentsByTenantID(tenantID)
	if err != nil {
		return nil, fmt.Errorf("get unpaid payments: %w", err)
	}

	// If exists and not fully paid, return it
	if len(unpaid) > 0 {
		return unpaid[0], nil
	}

	// Otherwise, create new payment
	tenant, err := s.tenantRepo.GetTenantByID(tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	unit, err := s.unitRepo.GetUnitByID(tenant.UnitID)
	if err != nil {
		return nil, fmt.Errorf("unit not found: %w", err)
	}

	// Calculate due date: Next 10th >= today
	now := time.Now()
	dueDate := time.Date(now.Year(), now.Month(), unit.PaymentDueDay, 0, 0, 0, 0, time.UTC)

	// If today is after due date in current month, use next month
	if now.Day() > unit.PaymentDueDay {
		dueDate = dueDate.AddDate(0, 1, 0) // Next month
	}

	// Use shared helper method
	return s.CreatePaymentForTenant(tenantID, tenant.UnitID, dueDate, unit.MonthlyRent)
}

// ============================================
// Payment Lifecycle & Helpers
// ============================================

// CreatePaymentForTenant creates a payment with explicit parameters
// Used by both TenantService (first payment) and PaymentService (auto-create)
func (s *PaymentService) CreatePaymentForTenant(
	tenantID int,
	unitID int,
	dueDate time.Time,
	amount int,
) (*domain.Payment, error) {
	payment := &domain.Payment{
		TenantID:         tenantID,
		UnitID:           unitID,
		Amount:           amount,
		AmountPaid:       0,
		RemainingBalance: amount,
		DueDate:          dueDate,
		IsPaid:           false,
		IsFullyPaid:      false,
		PaymentMethod:    "UPI",
		UPIID:            "9848790200@ybl",
	}

	if err := s.paymentRepo.CreatePayment(payment); err != nil {
		return nil, fmt.Errorf("create payment for tenant: %w", err)
	}

	return payment, nil
}

// CreateNextPayment creates the next payment for a tenant after current payment is fully paid
func (s *PaymentService) CreateNextPayment(currentPayment *domain.Payment) (*domain.Payment, error) {
	// Calculate next due date: currentPayment.DueDate + 1 month
	nextDueDate := currentPayment.DueDate.AddDate(0, 1, 0)

	// Use shared helper method
	return s.CreatePaymentForTenant(
		currentPayment.TenantID,
		currentPayment.UnitID,
		nextDueDate,
		currentPayment.Amount,
	)
}

// AutoCreateNextPayment automatically creates next payment when current is fully paid
// Exported for use by PaymentTransactionService
func (s *PaymentService) AutoCreateNextPayment(payment *domain.Payment) error {
	if !payment.IsFullyPaid {
		return nil // Not fully paid, no need to create next
	}

	// Check if next payment already exists
	nextDueDate := payment.DueDate.AddDate(0, 1, 0)
	existing, err := s.paymentRepo.GetPaymentByTenantAndMonth(payment.TenantID, nextDueDate.Month(), nextDueDate.Year())
	if err == nil && existing != nil {
		return nil // Next payment already exists
	}

	// Create next payment
	_, err = s.CreateNextPayment(payment)
	return err
}

// ============================================
// Historical Payment Management (for Existing Tenants)
// ============================================
// NOTE: Historical payment methods have been moved to PaymentHistoryService.
// These methods are kept for backward compatibility and forward to PaymentHistoryService.
// They will be removed in a future version.

// globalPaymentHistoryService is injected for backward compatibility
// TODO: Remove this after all handlers are migrated
// var globalPaymentHistoryService *PaymentHistoryService

// SetGlobalPaymentHistoryService sets the global history service for backward compatibility
// func SetGlobalPaymentHistoryService(historyService *PaymentHistoryService) {
// 	globalPaymentHistoryService = historyService
// }

// HistoricalPaymentData type is defined in payment_history_service.go
// Since both services are in the same package, the type is accessible here

// CreateHistoricalPaidPayment creates a payment record for a past month (DEPRECATED - use PaymentHistoryService)
// func (s *PaymentService) CreateHistoricalPaidPayment(
// 	tenantID int,
// 	month time.Month,
// 	year int,
// 	paymentDate time.Time,
// 	notes string,
// ) (*domain.Payment, error) {
// 	if globalPaymentHistoryService != nil {
// 		return globalPaymentHistoryService.CreateHistoricalPaidPayment(tenantID, month, year, paymentDate, notes)
// 	}
// 	return nil, fmt.Errorf("PaymentHistoryService not configured")
// }

// SyncPaymentHistory creates payment records for multiple months (DEPRECATED - use PaymentHistoryService)
// func (s *PaymentService) SyncPaymentHistory(
// 	tenantID int,
// 	payments []HistoricalPaymentData,
// ) ([]*domain.Payment, error) {
// 	if globalPaymentHistoryService != nil {
// 		return globalPaymentHistoryService.SyncPaymentHistory(tenantID, payments)
// 	}
// 	return nil, fmt.Errorf("PaymentHistoryService not configured")
// }

// AdjustFirstPaymentDueDate updates the first unpaid payment's due date (DEPRECATED - use PaymentHistoryService)
// func (s *PaymentService) AdjustFirstPaymentDueDate(tenantID int, newDueDate time.Time) error {
// 	if globalPaymentHistoryService != nil {
// 		return globalPaymentHistoryService.AdjustFirstPaymentDueDate(tenantID, newDueDate)
// 	}
// 	return fmt.Errorf("PaymentHistoryService not configured")
// }

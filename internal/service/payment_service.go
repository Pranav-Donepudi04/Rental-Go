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
	return s.autoCreateNextPayment(payment)
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

// GetOverduePayments returns all overdue payments
func (s *PaymentService) GetOverduePayments() ([]*domain.Payment, error) {
	payments, err := s.paymentRepo.GetAllPayments()
	if err != nil {
		return nil, err
	}

	var overdue []*domain.Payment
	now := time.Now()

	for _, payment := range payments {
		if !payment.IsFullyPaid && now.After(payment.DueDate) {
			overdue = append(overdue, payment)
		}
	}

	return overdue, nil
}

// GetPendingPayments returns all pending payments
func (s *PaymentService) GetPendingPayments() ([]*domain.Payment, error) {
	payments, err := s.paymentRepo.GetAllPayments()
	if err != nil {
		return nil, err
	}

	var pending []*domain.Payment
	now := time.Now()

	for _, payment := range payments {
		if !payment.IsFullyPaid && now.Before(payment.DueDate) {
			pending = append(pending, payment)
		}
	}

	return pending, nil
}

// GetPaymentSummary returns a summary of payments
func (s *PaymentService) GetPaymentSummary() (*PaymentSummary, error) {
	payments, err := s.paymentRepo.GetAllPayments()
	if err != nil {
		return nil, err
	}

	summary := &PaymentSummary{
		TotalPayments:   len(payments),
		PaidPayments:    0,
		PendingPayments: 0,
		OverduePayments: 0,
		TotalAmount:     0,
		PaidAmount:      0,
		PendingAmount:   0,
		OverdueAmount:   0,
	}

	now := time.Now()

	for _, payment := range payments {
		summary.TotalAmount += payment.Amount

		if payment.IsFullyPaid {
			summary.PaidPayments++
			summary.PaidAmount += payment.Amount
		} else if now.After(payment.DueDate) {
			summary.OverduePayments++
			summary.OverdueAmount += payment.RemainingBalance // Use remaining balance for overdue
		} else {
			summary.PendingPayments++
			summary.PendingAmount += payment.RemainingBalance // Use remaining balance for pending
		}
	}

	return summary, nil
}

// PaymentSummary represents payment summary
type PaymentSummary struct {
	TotalPayments   int `json:"total_payments"`
	PaidPayments    int `json:"paid_payments"`
	PendingPayments int `json:"pending_payments"`
	OverduePayments int `json:"overdue_payments"`
	TotalAmount     int `json:"total_amount"`
	PaidAmount      int `json:"paid_amount"`
	PendingAmount   int `json:"pending_amount"`
	OverdueAmount   int `json:"overdue_amount"`
}

// GetFormattedTotalAmount returns formatted total amount
func (ps *PaymentSummary) GetFormattedTotalAmount() string {
	return fmt.Sprintf("₹%d", ps.TotalAmount)
}

// GetFormattedPaidAmount returns formatted paid amount
func (ps *PaymentSummary) GetFormattedPaidAmount() string {
	return fmt.Sprintf("₹%d", ps.PaidAmount)
}

// GetFormattedPendingAmount returns formatted pending amount
func (ps *PaymentSummary) GetFormattedPendingAmount() string {
	return fmt.Sprintf("₹%d", ps.PendingAmount)
}

// GetFormattedOverdueAmount returns formatted overdue amount
func (ps *PaymentSummary) GetFormattedOverdueAmount() string {
	return fmt.Sprintf("₹%d", ps.OverdueAmount)
}

// ============================================
// Payment CRUD Operations
// ============================================

// GetAllPayments returns all payments
func (s *PaymentService) GetAllPayments() ([]*domain.Payment, error) {
	return s.paymentRepo.GetAllPayments()
}

// ============================================
// Transaction Management
// ============================================

// GetPendingVerifications returns all pending verifications for a tenant (0 = all tenants)
func (s *PaymentService) GetPendingVerifications(tenantID int) ([]*domain.PaymentTransaction, error) {
	return s.paymentRepo.GetPendingVerifications(tenantID)
}

// SubmitPaymentIntent creates a payment transaction record for a tenant
func (s *PaymentService) SubmitPaymentIntent(tenantID int, txnID string) error {
	// Get or create current unpaid payment
	payment, err := s.getOrCreateCurrentPayment(tenantID)
	if err != nil {
		return fmt.Errorf("get or create payment: %w", err)
	}

	// Check if transaction already exists
	existing, err := s.paymentRepo.GetTransactionByPaymentAndID(payment.ID, txnID)
	if err != nil {
		return fmt.Errorf("check existing transaction: %w", err)
	}
	if existing != nil {
		return nil // Already exists, no error
	}

	// Create payment transaction (amount NULL until owner verifies)
	tx := &domain.PaymentTransaction{
		PaymentID:     payment.ID,
		TransactionID: txnID,
		Amount:        nil, // NULL until owner verifies
		SubmittedAt:   time.Now(),
	}

	if err := s.paymentRepo.CreatePaymentTransaction(tx); err != nil {
		return fmt.Errorf("create payment transaction: %w", err)
	}

	return nil
}

// VerifyTransaction verifies a transaction by setting its amount and updating the payment
func (s *PaymentService) VerifyTransaction(transactionID string, amount int, verifiedByUserID int) error {
	// Get transaction directly by ID (efficient - O(1) instead of O(n))
	tx, err := s.paymentRepo.GetTransactionByID(transactionID)
	if err != nil {
		return fmt.Errorf("get transaction: %w", err)
	}
	if tx == nil {
		return fmt.Errorf("transaction not found")
	}

	paymentID := tx.PaymentID

	// Verify transaction (this also updates payment's amount_paid and remaining_balance)
	if err := s.paymentRepo.VerifyTransaction(transactionID, amount, verifiedByUserID); err != nil {
		return fmt.Errorf("verify transaction: %w", err)
	}

	// Get the updated payment to check if fully paid
	payment, err := s.paymentRepo.GetPaymentByID(paymentID)
	if err != nil {
		return fmt.Errorf("get updated payment: %w", err)
	}

	// Auto-create next payment if fully paid
	if payment.IsFullyPaid {
		return s.autoCreateNextPayment(payment)
	}

	return nil
}

// RejectTransaction rejects a pending transaction (deletes it)
func (s *PaymentService) RejectTransaction(transactionID string) error {
	// Check if transaction exists
	tx, err := s.paymentRepo.GetTransactionByID(transactionID)
	if err != nil {
		return fmt.Errorf("get transaction: %w", err)
	}
	if tx == nil {
		return fmt.Errorf("transaction not found")
	}

	// Reject (delete) the transaction
	return s.paymentRepo.RejectTransaction(transactionID)
}

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

// autoCreateNextPayment automatically creates next payment when current is fully paid
func (s *PaymentService) autoCreateNextPayment(payment *domain.Payment) error {
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

// CreateHistoricalPaidPayment creates a payment record for a past month and marks it as fully paid
// Used when migrating existing tenants who have already paid manually
func (s *PaymentService) CreateHistoricalPaidPayment(
	tenantID int,
	month time.Month,
	year int,
	paymentDate time.Time,
	notes string,
) (*domain.Payment, error) {
	// Get tenant and unit information
	tenant, err := s.tenantRepo.GetTenantByID(tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	unit, err := s.unitRepo.GetUnitByID(tenant.UnitID)
	if err != nil {
		return nil, fmt.Errorf("unit not found: %w", err)
	}

	// Calculate due date for the selected month/year
	dueDate := time.Date(year, month, unit.PaymentDueDay, 0, 0, 0, 0, tenant.MoveInDate.Location())

	// Calculate the first valid payment due date based on move-in date
	// Logic: First payment due date is the payment due day in the move-in month, or next month if move-in is after due day
	moveInDate := tenant.MoveInDate
	firstValidDueDate := time.Date(moveInDate.Year(), moveInDate.Month(), unit.PaymentDueDay, 0, 0, 0, 0, moveInDate.Location())

	// If move-in date is after due date in same month, first payment is next month
	if moveInDate.Day() > unit.PaymentDueDay {
		firstValidDueDate = firstValidDueDate.AddDate(0, 1, 0) // Next month
	}

	// Validate: Payment due date (month/year) must not be before the first valid payment due date
	if dueDate.Before(firstValidDueDate) {
		return nil, fmt.Errorf("cannot create payment for %s %d (due date: %s) - tenant moved in on %s, first payment due date is %s",
			month.String(), year,
			dueDate.Format("January 2, 2006"),
			moveInDate.Format("January 2, 2006"),
			firstValidDueDate.Format("January 2, 2006"))
	}

	// Check if payment already exists for this month
	existingPayment, err := s.paymentRepo.GetPaymentByTenantAndMonth(tenantID, month, year)
	if err == nil && existingPayment != nil {
		// If exists but not paid, mark it as paid
		if !existingPayment.IsFullyPaid {
			existingPayment.IsPaid = true
			existingPayment.IsFullyPaid = true
			existingPayment.AmountPaid = existingPayment.Amount
			existingPayment.RemainingBalance = 0
			existingPayment.PaymentDate = &paymentDate
			existingPayment.FullyPaidDate = &paymentDate
			existingPayment.Notes = notes
			if err := s.paymentRepo.UpdatePayment(existingPayment); err != nil {
				return nil, fmt.Errorf("failed to update existing payment: %w", err)
			}
			return existingPayment, nil
		}
		// Already fully paid, return it
		return existingPayment, nil
	}

	// Create payment record marked as fully paid
	payment := &domain.Payment{
		TenantID:         tenantID,
		UnitID:           tenant.UnitID,
		Amount:           unit.MonthlyRent,
		AmountPaid:       unit.MonthlyRent,
		RemainingBalance: 0,
		DueDate:          dueDate,
		PaymentDate:      &paymentDate,
		IsPaid:           true,
		IsFullyPaid:      true,
		FullyPaidDate:    &paymentDate,
		PaymentMethod:    "UPI",
		UPIID:            "9848790200@ybl",
		Notes:            notes,
	}

	if err := s.paymentRepo.CreatePayment(payment); err != nil {
		return nil, fmt.Errorf("failed to create historical payment: %w", err)
	}

	return payment, nil
}

// SyncPaymentHistory creates payment records for multiple months and marks them as paid
// This is useful when transitioning an existing tenant to the system
// After syncing, automatically creates the next pending payment
func (s *PaymentService) SyncPaymentHistory(
	tenantID int,
	payments []HistoricalPaymentData,
) ([]*domain.Payment, error) {
	var createdPayments []*domain.Payment
	var latestPaidDate time.Time

	// Validate and create payments
	// Note: CreateHistoricalPaidPayment already validates move-in date
	for _, paymentData := range payments {
		payment, err := s.CreateHistoricalPaidPayment(
			tenantID,
			paymentData.Month,
			paymentData.Year,
			paymentData.PaymentDate,
			paymentData.Notes,
		)
		if err != nil {
			// Return error immediately if validation fails (e.g., before move-in date)
			return nil, fmt.Errorf("failed to create payment for %s %d: %w", paymentData.Month.String(), paymentData.Year, err)
		}
		createdPayments = append(createdPayments, payment)

		// Track the latest payment due date
		if payment.DueDate.After(latestPaidDate) {
			latestPaidDate = payment.DueDate
		}
	}

	// Auto-create next payment after the latest synced payment
	if len(createdPayments) > 0 {
		if err := s.autoCreateNextPaymentAfterSync(tenantID, latestPaidDate); err != nil {
			// Log warning but don't fail the sync
			fmt.Printf("Warning: Failed to auto-create next payment after sync: %v\n", err)
		}
	}

	return createdPayments, nil
}

// autoCreateNextPaymentAfterSync creates the next payment after syncing historical payments
func (s *PaymentService) autoCreateNextPaymentAfterSync(tenantID int, latestPaidDate time.Time) error {
	// Get tenant and unit
	tenant, err := s.tenantRepo.GetTenantByID(tenantID)
	if err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	unit, err := s.unitRepo.GetUnitByID(tenant.UnitID)
	if err != nil {
		return fmt.Errorf("unit not found: %w", err)
	}

	// Calculate next due date: latest paid date + 1 month
	nextDueDate := latestPaidDate.AddDate(0, 1, 0)

	// Check if payment already exists for next month
	existing, err := s.paymentRepo.GetPaymentByTenantAndMonth(tenantID, nextDueDate.Month(), nextDueDate.Year())
	if err == nil && existing != nil {
		// Payment already exists, don't create duplicate
		return nil
	}

	// Create next payment as pending
	_, err = s.CreatePaymentForTenant(
		tenantID,
		tenant.UnitID,
		nextDueDate,
		unit.MonthlyRent,
	)
	return err
}

// HistoricalPaymentData represents data for creating a historical payment
type HistoricalPaymentData struct {
	Month       time.Month `json:"month"`        // 1-12
	Year        int        `json:"year"`         // e.g., 2024
	PaymentDate time.Time  `json:"payment_date"` // When the payment was actually made
	Notes       string     `json:"notes"`        // Optional notes
}

// AdjustFirstPaymentDueDate updates the first unpaid payment's due date for an existing tenant
// This is useful when the auto-created payment's due date doesn't match the actual payment schedule
func (s *PaymentService) AdjustFirstPaymentDueDate(tenantID int, newDueDate time.Time) error {
	// Get unpaid payments
	unpaid, err := s.paymentRepo.GetUnpaidPaymentsByTenantID(tenantID)
	if err != nil {
		return fmt.Errorf("get unpaid payments: %w", err)
	}

	if len(unpaid) == 0 {
		return fmt.Errorf("no unpaid payments found for tenant %d", tenantID)
	}

	// Update the first unpaid payment's due date
	firstPayment := unpaid[0]
	firstPayment.DueDate = newDueDate

	if err := s.paymentRepo.UpdatePayment(firstPayment); err != nil {
		return fmt.Errorf("failed to update payment due date: %w", err)
	}

	return nil
}

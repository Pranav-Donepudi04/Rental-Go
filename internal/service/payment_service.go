package service

import (
	"backend-form/m/internal/domain"
	interfaces "backend-form/m/internal/repository/interfaces"
	"fmt"
	"time"
)

// PaymentService handles payment-related business logic
type PaymentService struct {
	paymentRepo          interfaces.PaymentRepository
	tenantRepo           interfaces.TenantRepository
	unitRepo             interfaces.UnitRepository
	defaultPaymentMethod string
	defaultUPIID         string
}

// NewPaymentService creates a new PaymentService
func NewPaymentService(paymentRepo interfaces.PaymentRepository, tenantRepo interfaces.TenantRepository, unitRepo interfaces.UnitRepository, defaultPaymentMethod, defaultUPIID string) *PaymentService {
	return &PaymentService{
		paymentRepo:          paymentRepo,
		tenantRepo:           tenantRepo,
		unitRepo:             unitRepo,
		defaultPaymentMethod: defaultPaymentMethod,
		defaultUPIID:         defaultUPIID,
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
		PaymentMethod: s.defaultPaymentMethod,
		UPIID:         s.defaultUPIID,
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
// Use PaymentQueryService for all payment queries.

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
		PaymentMethod:    s.defaultPaymentMethod,
		UPIID:            s.defaultUPIID,
		Label:            domain.PaymentLabelRent, // Auto-created payments are always rent
	}

	if err := s.paymentRepo.CreatePayment(payment); err != nil {
		return nil, fmt.Errorf("create payment for tenant: %w", err)
	}

	return payment, nil
}

// CreateCustomPayment creates a payment with custom label (for water bills, current bills, maintenance, etc.)
func (s *PaymentService) CreateCustomPayment(
	tenantID int,
	unitID int,
	dueDate time.Time,
	amount int,
	label string,
	notes string,
) (*domain.Payment, error) {
	// Validate tenant exists
	tenant, err := s.tenantRepo.GetTenantByID(tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Validate unit exists
	_, err = s.unitRepo.GetUnitByID(unitID)
	if err != nil {
		return nil, fmt.Errorf("unit not found: %w", err)
	}

	// Ensure tenant belongs to this unit
	if tenant.UnitID != unitID {
		return nil, fmt.Errorf("tenant does not belong to this unit")
	}

	// Set default label if not provided
	if label == "" {
		label = domain.PaymentLabelRent
	}

	// Check if payment of this type already exists for this month
	// Only check for non-rent payments (rent is handled by auto-creation logic)
	if label != domain.PaymentLabelRent {
		existingPayments, err := s.paymentRepo.GetPaymentsByTenantID(tenantID)
		if err == nil {
			paymentMonth := dueDate.Month()
			paymentYear := dueDate.Year()
			for _, existing := range existingPayments {
				if existing.Label == label &&
					existing.DueDate.Month() == paymentMonth &&
					existing.DueDate.Year() == paymentYear {
					return nil, fmt.Errorf("payment of type '%s' already exists for %s %d", label, paymentMonth.String(), paymentYear)
				}
			}
		}
	}

	payment := &domain.Payment{
		TenantID:         tenantID,
		UnitID:           unitID,
		Amount:           amount,
		AmountPaid:       0,
		RemainingBalance: amount,
		DueDate:          dueDate,
		IsPaid:           false,
		IsFullyPaid:      false,
		PaymentMethod:    s.defaultPaymentMethod,
		UPIID:            s.defaultUPIID,
		Label:            label,
		Notes:            notes,
	}

	// Validate payment
	if err := payment.Validate(); err != nil {
		return nil, fmt.Errorf("invalid payment: %w", err)
	}

	if err := s.paymentRepo.CreatePayment(payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return payment, nil
}

// GetDefaultPaymentMethod returns the default payment method
func (s *PaymentService) GetDefaultPaymentMethod() string {
	return s.defaultPaymentMethod
}

// GetDefaultUPIID returns the default UPI ID
func (s *PaymentService) GetDefaultUPIID() string {
	return s.defaultUPIID
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
// Use PaymentHistoryService for all historical payment operations.

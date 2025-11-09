package service

import (
	"backend-form/m/internal/domain"
	interfaces "backend-form/m/internal/repository/interfaces"
	"fmt"
	"time"
)

// PaymentHistoryService handles historical payment management (migration/backfill)
// This service is used when transitioning existing tenants to the system
type PaymentHistoryService struct {
	paymentRepo    interfaces.PaymentRepository
	tenantRepo     interfaces.TenantRepository
	unitRepo       interfaces.UnitRepository
	paymentService *PaymentService // For CreatePaymentForTenant
}

// NewPaymentHistoryService creates a new PaymentHistoryService
func NewPaymentHistoryService(
	paymentRepo interfaces.PaymentRepository,
	tenantRepo interfaces.TenantRepository,
	unitRepo interfaces.UnitRepository,
	paymentService *PaymentService,
) *PaymentHistoryService {
	return &PaymentHistoryService{
		paymentRepo:    paymentRepo,
		tenantRepo:     tenantRepo,
		unitRepo:       unitRepo,
		paymentService: paymentService,
	}
}

// HistoricalPaymentData represents data for creating a historical payment
type HistoricalPaymentData struct {
	Month       time.Month `json:"month"`        // 1-12
	Year        int        `json:"year"`         // e.g., 2024
	PaymentDate time.Time  `json:"payment_date"` // When the payment was actually made
	Notes       string     `json:"notes"`        // Optional notes
}

// CreateHistoricalPaidPayment creates a payment record for a past month and marks it as fully paid
// Used when migrating existing tenants who have already paid manually
func (s *PaymentHistoryService) CreateHistoricalPaidPayment(
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
		PaymentMethod:    s.paymentService.GetDefaultPaymentMethod(),
		UPIID:            s.paymentService.GetDefaultUPIID(),
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
func (s *PaymentHistoryService) SyncPaymentHistory(
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
		if err := s.AutoCreateNextPaymentAfterSync(tenantID, latestPaidDate); err != nil {
			// Log warning but don't fail the sync
			fmt.Printf("Warning: Failed to auto-create next payment after sync: %v\n", err)
		}
	}

	return createdPayments, nil
}

// AutoCreateNextPaymentAfterSync creates the next payment after syncing historical payments
// Exported for use by PaymentHistoryService
func (s *PaymentHistoryService) AutoCreateNextPaymentAfterSync(tenantID int, latestPaidDate time.Time) error {
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

	// Create next payment as pending using PaymentService
	_, err = s.paymentService.CreatePaymentForTenant(
		tenantID,
		tenant.UnitID,
		nextDueDate,
		unit.MonthlyRent,
	)
	return err
}

// AdjustFirstPaymentDueDate updates the first unpaid payment's due date for an existing tenant
// This is useful when the auto-created payment's due date doesn't match the actual payment schedule
func (s *PaymentHistoryService) AdjustFirstPaymentDueDate(tenantID int, newDueDate time.Time) error {
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

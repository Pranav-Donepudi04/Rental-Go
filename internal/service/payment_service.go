package service

import (
	"backend-form/m/internal/domain"
	interfaces "backend-form/m/internal/repository/interfaces"
	"fmt"
	"strings"
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

// MarkPaymentAsPaid marks a payment as paid
func (s *PaymentService) MarkPaymentAsPaid(paymentID int, paymentDate time.Time, notes string) error {
	payment, err := s.paymentRepo.GetPaymentByID(paymentID)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	if payment.IsPaid {
		return fmt.Errorf("payment is already marked as paid")
	}

	payment.IsPaid = true
	payment.PaymentDate = &paymentDate
	payment.Notes = notes

	return s.paymentRepo.UpdatePayment(payment)
}

// GetPaymentByID returns a payment by ID with related data
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

	return payment, nil
}

// GetPaymentsByTenantID returns all payments for a tenant
func (s *PaymentService) GetPaymentsByTenantID(tenantID int) ([]*domain.Payment, error) {
	return s.paymentRepo.GetPaymentsByTenantID(tenantID)
}

// GetOverduePayments returns all overdue payments
func (s *PaymentService) GetOverduePayments() ([]*domain.Payment, error) {
	payments, err := s.paymentRepo.GetAllPayments()
	if err != nil {
		return nil, err
	}

	var overdue []*domain.Payment
	now := time.Now()

	for _, payment := range payments {
		if !payment.IsPaid && now.After(payment.DueDate) {
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
		if !payment.IsPaid && now.Before(payment.DueDate) {
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

		if payment.IsPaid {
			summary.PaidPayments++
			summary.PaidAmount += payment.Amount
		} else if now.After(payment.DueDate) {
			summary.OverduePayments++
			summary.OverdueAmount += payment.Amount
		} else {
			summary.PendingPayments++
			summary.PendingAmount += payment.Amount
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

// GetAllPayments returns all payments
func (s *PaymentService) GetAllPayments() ([]*domain.Payment, error) {
	return s.paymentRepo.GetAllPayments()
}

// SubmitPaymentIntent appends the provided txn id to the current month's payment notes for a tenant.
func (s *PaymentService) SubmitPaymentIntent(tenantID int, txnID string) error {
	now := time.Now()
	p, err := s.paymentRepo.GetPaymentByTenantAndMonth(tenantID, now.Month(), now.Year())
	if err != nil {
		return fmt.Errorf("fetch payment: %w", err)
	}
	if p == nil {
		// Create a payment record for the current month if it doesn't exist
		tenant, err := s.tenantRepo.GetTenantByID(tenantID)
		if err != nil {
			return fmt.Errorf("fetch tenant: %w", err)
		}
		if tenant == nil {
			return fmt.Errorf("tenant not found")
		}

		// Get unit information for monthly rent
		unit, err := s.unitRepo.GetUnitByID(tenant.UnitID)
		if err != nil {
			return fmt.Errorf("fetch unit: %w", err)
		}
		if unit == nil {
			return fmt.Errorf("unit not found")
		}

		// Create payment for current month
		dueDate := time.Date(now.Year(), now.Month(), unit.PaymentDueDay, 0, 0, 0, 0, time.UTC)
		payment := &domain.Payment{
			TenantID:      tenantID,
			UnitID:        tenant.UnitID,
			Amount:        unit.MonthlyRent,
			DueDate:       dueDate,
			IsPaid:        false,
			PaymentMethod: "UPI",
			Notes:         "",
		}

		if err := s.paymentRepo.CreatePayment(payment); err != nil {
			return fmt.Errorf("create payment: %w", err)
		}
		p = payment
	}
	// avoid duplicate appends of the same txn id
	marker := "TXN:" + txnID
	if p.Notes != "" && (p.Notes == marker || containsTxn(p.Notes, marker)) {
		return nil
	}
	if p.Notes == "" {
		p.Notes = marker
	} else {
		p.Notes = p.Notes + "; " + marker
	}
	return s.paymentRepo.UpdatePayment(p)
}

// containsTxn checks if notes already contains the exact marker token boundary-separated by ';'
func containsTxn(notes, marker string) bool {
	// simple contains is sufficient given our consistent formatting with '; ' delimiter
	return strings.Contains(notes, marker)
}

package service

import (
	"backend-form/m/internal/domain"
	interfaces "backend-form/m/internal/repository/interfaces"
	"fmt"
	"time"
)

// PaymentQueryService handles read-only payment queries and aggregations
// This service has minimal dependencies (only paymentRepo) for efficient querying
type PaymentQueryService struct {
	paymentRepo interfaces.PaymentRepository
}

// NewPaymentQueryService creates a new PaymentQueryService
func NewPaymentQueryService(paymentRepo interfaces.PaymentRepository) *PaymentQueryService {
	return &PaymentQueryService{
		paymentRepo: paymentRepo,
	}
}

// GetAllPayments returns all payments
func (s *PaymentQueryService) GetAllPayments() ([]*domain.Payment, error) {
	return s.paymentRepo.GetAllPayments()
}

// GetUnpaidPaymentsByTenantID returns unpaid payments for a tenant
func (s *PaymentQueryService) GetUnpaidPaymentsByTenantID(tenantID int) ([]*domain.Payment, error) {
	return s.paymentRepo.GetUnpaidPaymentsByTenantID(tenantID)
}

// GetOverduePayments returns all overdue payments
func (s *PaymentQueryService) GetOverduePayments() ([]*domain.Payment, error) {
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
func (s *PaymentQueryService) GetPendingPayments() ([]*domain.Payment, error) {
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
func (s *PaymentQueryService) GetPaymentSummary() (*PaymentSummary, error) {
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
// Moved from payment_service.go for use by PaymentQueryService
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

package service

import (
	"backend-form/m/internal/domain"
	interfaces "backend-form/m/internal/repository/interfaces"
	"fmt"
	"time"
)

// PaymentTransactionService handles payment transaction operations (submit, verify, reject)
// This service manages the transaction workflow between tenants and owners
type PaymentTransactionService struct {
	paymentRepo    interfaces.PaymentRepository
	paymentService *PaymentService // For getOrCreateCurrentPayment and autoCreateNextPayment
}

// NewPaymentTransactionService creates a new PaymentTransactionService
func NewPaymentTransactionService(paymentRepo interfaces.PaymentRepository, paymentService *PaymentService) *PaymentTransactionService {
	return &PaymentTransactionService{
		paymentRepo:    paymentRepo,
		paymentService: paymentService,
	}
}

// GetPendingVerifications returns all pending verifications for a tenant (0 = all tenants)
func (s *PaymentTransactionService) GetPendingVerifications(tenantID int) ([]*domain.PaymentTransaction, error) {
	return s.paymentRepo.GetPendingVerifications(tenantID)
}

// SubmitPaymentIntent creates a payment transaction record for a tenant
// This is called when a tenant submits a UPI transaction ID
func (s *PaymentTransactionService) SubmitPaymentIntent(tenantID int, txnID string) error {
	// Get or create current unpaid payment using PaymentService
	payment, err := s.paymentService.getOrCreateCurrentPayment(tenantID)
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
// This is called when an owner verifies a transaction submitted by a tenant
func (s *PaymentTransactionService) VerifyTransaction(transactionID string, amount int, verifiedByUserID int) error {
	// Get transaction directly by ID (efficient - O(1))
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

	// Auto-create next payment if fully paid (using PaymentService)
	if payment.IsFullyPaid {
		return s.paymentService.AutoCreateNextPayment(payment)
	}

	return nil
}

// RejectTransaction rejects a pending transaction (deletes it)
// This is called when an owner rejects an invalid transaction
func (s *PaymentTransactionService) RejectTransaction(transactionID string) error {
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

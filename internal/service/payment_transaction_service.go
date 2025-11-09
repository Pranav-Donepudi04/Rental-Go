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
// suggestedAmount is optional - if provided, it's stored in notes for owner reference
func (s *PaymentTransactionService) SubmitPaymentIntent(tenantID int, txnID string, suggestedAmount *int) error {
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
	// Store suggested amount in notes if provided
	notes := ""
	if suggestedAmount != nil {
		notes = fmt.Sprintf("Suggested amount: ₹%d", *suggestedAmount)
	}

	tx := &domain.PaymentTransaction{
		PaymentID:     payment.ID,
		TransactionID: txnID,
		Amount:        nil, // NULL until owner verifies
		SubmittedAt:   time.Now(),
		Notes:         notes,
	}

	if err := s.paymentRepo.CreatePaymentTransaction(tx); err != nil {
		return fmt.Errorf("create payment transaction: %w", err)
	}

	return nil
}

// VerifyTransaction verifies a transaction by setting its amount and updating the payment
// This implements smart allocation: if amount exceeds the linked payment, excess is allocated to next payments
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

	// Check if transaction is already verified (early check to avoid unnecessary work)
	if tx.IsVerified() {
		verifiedAmount := "unknown"
		if tx.Amount != nil {
			verifiedAmount = fmt.Sprintf("₹%d", *tx.Amount)
		}
		return fmt.Errorf("transaction already verified with amount %s on %s", verifiedAmount, tx.GetFormattedVerifiedAt())
	}

	// Get the linked payment to get tenant ID
	linkedPayment, err := s.paymentRepo.GetPaymentByID(tx.PaymentID)
	if err != nil {
		return fmt.Errorf("get linked payment: %w", err)
	}

	tenantID := linkedPayment.TenantID

	// Get all unpaid payments for this tenant (ordered by due date - most overdue first)
	unpaidPayments, err := s.paymentRepo.GetUnpaidPaymentsByTenantID(tenantID)
	if err != nil {
		return fmt.Errorf("get unpaid payments: %w", err)
	}

	if len(unpaidPayments) == 0 {
		return fmt.Errorf("no unpaid payments found for tenant")
	}

	// Find the linked payment in the unpaid list (should be first or near first)
	// Allocate starting from linked payment, then continue to next payments
	remainingAmount := amount
	allocations := make(map[int]int) // paymentID -> amount to allocate

	// Find linked payment index
	linkedIndex := -1
	for i, p := range unpaidPayments {
		if p.ID == tx.PaymentID {
			linkedIndex = i
			break
		}
	}

	if linkedIndex == -1 {
		// Linked payment is already paid or doesn't exist, start from first unpaid
		linkedIndex = 0
	}

	// Allocate amount across payments starting from linked payment
	for i := linkedIndex; i < len(unpaidPayments) && remainingAmount > 0; i++ {
		payment := unpaidPayments[i]
		needed := payment.RemainingBalance

		if remainingAmount >= needed {
			// Fully pay this payment
			allocations[payment.ID] = needed
			remainingAmount -= needed
		} else {
			// Partially pay this payment
			allocations[payment.ID] = remainingAmount
			remainingAmount = 0
		}
	}

	// If there's still remaining amount after all unpaid payments, that's an overpayment
	// We'll allocate it to the last payment (or could reject - but allocating is more user-friendly)
	if remainingAmount > 0 && len(unpaidPayments) > 0 {
		lastPayment := unpaidPayments[len(unpaidPayments)-1]
		allocations[lastPayment.ID] += remainingAmount
	}

	// Apply all allocations in a single database transaction for atomicity
	now := time.Now()
	if err := s.paymentRepo.ApplySmartAllocation(transactionID, amount, verifiedByUserID, allocations, now); err != nil {
		return fmt.Errorf("apply smart allocation: %w", err)
	}

	// After successful allocation, check for fully paid payments and auto-create next (only for rent)
	for paymentID := range allocations {
		payment, err := s.paymentRepo.GetPaymentByID(paymentID)
		if err == nil && payment != nil && payment.IsFullyPaid && payment.Label == domain.PaymentLabelRent {
			if err := s.paymentService.AutoCreateNextPayment(payment); err != nil {
				// Log error but don't fail - next payment creation is best effort
				fmt.Printf("Warning: Failed to auto-create next payment for payment %d: %v\n", paymentID, err)
			}
		}
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

package interfaces

import (
	"backend-form/m/internal/domain"
	"time"
)

// PaymentRepository defines the interface for payment data operations
type PaymentRepository interface {
	// Existing payment methods
	CreatePayment(payment *domain.Payment) error
	GetPaymentByID(id int) (*domain.Payment, error)
	GetAllPayments() ([]*domain.Payment, error)
	UpdatePayment(payment *domain.Payment) error
	DeletePayment(id int) error
	GetPaymentsByTenantID(tenantID int) ([]*domain.Payment, error)
	GetPaymentByTenantAndMonth(tenantID int, month time.Month, year int) (*domain.Payment, error)
	DeletePaymentsByTenantID(tenantID int) error

	// NEW: Transaction methods
	CreatePaymentTransaction(tx *domain.PaymentTransaction) error
	GetPaymentTransactionsByPaymentID(paymentID int) ([]*domain.PaymentTransaction, error)
	GetTransactionByPaymentAndID(paymentID int, transactionID string) (*domain.PaymentTransaction, error)
	GetTransactionByID(transactionID string) (*domain.PaymentTransaction, error)
	GetPendingVerifications(tenantID int) ([]*domain.PaymentTransaction, error)
	VerifyTransaction(transactionID string, amount int, verifiedByUserID int) error
	VerifyTransactionRecord(transactionID string, amount int, verifiedByUserID int, verifiedAt time.Time) error
	ApplyPaymentAllocation(paymentID int, amount int, allocationTime time.Time) error
	ApplySmartAllocation(transactionID string, amount int, verifiedByUserID int, allocations map[int]int, allocationTime time.Time) error
	RejectTransaction(transactionID string) error

	// NEW: Auto-create helpers
	GetLatestPaymentByTenantID(tenantID int) (*domain.Payment, error)
	GetUnpaidPaymentsByTenantID(tenantID int) ([]*domain.Payment, error)

	// NEW: Notification helpers - get payments by due date
	GetUnpaidPaymentsByDueDate(dueDate time.Time) ([]*domain.Payment, error)
	GetUnpaidPaymentsDueInDays(days int) ([]*domain.Payment, error)
}

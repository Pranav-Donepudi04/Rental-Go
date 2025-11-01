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

	// NEW: Auto-create helpers
	GetLatestPaymentByTenantID(tenantID int) (*domain.Payment, error)
	GetUnpaidPaymentsByTenantID(tenantID int) ([]*domain.Payment, error)
}

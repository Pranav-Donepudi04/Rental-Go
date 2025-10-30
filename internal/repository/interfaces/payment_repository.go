package interfaces

import (
	"backend-form/m/internal/domain"
	"time"
)

// PaymentRepository defines the interface for payment data operations
type PaymentRepository interface {
	CreatePayment(payment *domain.Payment) error
	GetPaymentByID(id int) (*domain.Payment, error)
	GetAllPayments() ([]*domain.Payment, error)
	UpdatePayment(payment *domain.Payment) error
	DeletePayment(id int) error
	GetPaymentsByTenantID(tenantID int) ([]*domain.Payment, error)
	GetPaymentByTenantAndMonth(tenantID int, month time.Month, year int) (*domain.Payment, error)
	DeletePaymentsByTenantID(tenantID int) error
}

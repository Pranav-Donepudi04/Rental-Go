package repository

import (
	"backend-form/m/internal/models"
	"time"
)

// PaymentRepository defines the interface for payment data operations
type PaymentRepository interface {
	CreatePayment(payment *models.Payment) error
	GetPaymentByID(id int) (*models.Payment, error)
	GetAllPayments() ([]*models.Payment, error)
	UpdatePayment(payment *models.Payment) error
	DeletePayment(id int) error
	GetPaymentsByTenantID(tenantID int) ([]*models.Payment, error)
	GetPaymentByTenantAndMonth(tenantID int, month time.Month, year int) (*models.Payment, error)
	DeletePaymentsByTenantID(tenantID int) error
}

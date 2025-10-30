package domain

import (
	"fmt"
	"time"
)

// Payment represents a rent payment
type Payment struct {
	ID            int        `json:"id" db:"id"`
	TenantID      int        `json:"tenant_id" db:"tenant_id"`
	UnitID        int        `json:"unit_id" db:"unit_id"`
	Amount        int        `json:"amount" db:"amount"`
	PaymentDate   *time.Time `json:"payment_date" db:"payment_date"`
	DueDate       time.Time  `json:"due_date" db:"due_date"`
	IsPaid        bool       `json:"is_paid" db:"is_paid"`
	PaymentMethod string     `json:"payment_method" db:"payment_method"`
	UPIID         string     `json:"upi_id" db:"upi_id"`
	Notes         string     `json:"notes" db:"notes"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`

	// Related data (populated by joins)
	Tenant *Tenant `json:"tenant,omitempty"`
	Unit   *Unit   `json:"unit,omitempty"`
}

// Validate validates the payment data
func (p *Payment) Validate() error {
	if p.TenantID <= 0 {
		return fmt.Errorf("tenant ID is required")
	}
	if p.UnitID <= 0 {
		return fmt.Errorf("unit ID is required")
	}
	if p.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if p.DueDate.IsZero() {
		return fmt.Errorf("due date is required")
	}
	return nil
}

// GetStatus returns the payment status as a string
func (p *Payment) GetStatus() string {
	if p.IsPaid {
		return "Paid"
	}
	if time.Now().After(p.DueDate) {
		return "Overdue"
	}
	return "Pending"
}

// GetDaysOverdue returns the number of days overdue (0 if not overdue)
func (p *Payment) GetDaysOverdue() int {
	if p.IsPaid || time.Now().Before(p.DueDate) {
		return 0
	}
	return int(time.Since(p.DueDate).Hours() / 24)
}

// GetFormattedAmount returns the amount formatted as currency
func (p *Payment) GetFormattedAmount() string {
	return fmt.Sprintf("₹%d", p.Amount)
}

// GetFormattedDueDate returns the due date formatted as a string
func (p *Payment) GetFormattedDueDate() string {
	return p.DueDate.Format("Jan 2, 2006")
}

// GetFormattedPaymentDate returns the payment date formatted as a string
func (p *Payment) GetFormattedPaymentDate() string {
	if p.PaymentDate == nil {
		return "Not paid"
	}
	return p.PaymentDate.Format("Jan 2, 2006")
}

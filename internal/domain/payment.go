package domain

import (
	"fmt"
	"time"
)

// Payment represents a rent payment
type Payment struct {
	ID               int        `json:"id" db:"id"`
	TenantID         int        `json:"tenant_id" db:"tenant_id"`
	UnitID           int        `json:"unit_id" db:"unit_id"`
	Amount           int        `json:"amount" db:"amount"`
	AmountPaid       int        `json:"amount_paid" db:"amount_paid"`
	RemainingBalance int        `json:"remaining_balance" db:"remaining_balance"`
	PaymentDate      *time.Time `json:"payment_date" db:"payment_date"`
	DueDate          time.Time  `json:"due_date" db:"due_date"`
	IsPaid           bool       `json:"is_paid" db:"is_paid"` // Legacy field, kept for backward compatibility
	IsFullyPaid      bool       `json:"is_fully_paid" db:"is_fully_paid"`
	FullyPaidDate    *time.Time `json:"fully_paid_date" db:"fully_paid_date"`
	PaymentMethod    string     `json:"payment_method" db:"payment_method"`
	UPIID            string     `json:"upi_id" db:"upi_id"`
	Notes            string     `json:"notes" db:"notes"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`

	// Related data (populated by joins)
	Tenant       *Tenant               `json:"tenant,omitempty"`
	Unit         *Unit                 `json:"unit,omitempty"`
	Transactions []*PaymentTransaction `json:"transactions,omitempty"`
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

// RecalculateBalance recalculates remaining balance and is_fully_paid status
func (p *Payment) RecalculateBalance() {
	p.RemainingBalance = p.Amount - p.AmountPaid
	p.IsFullyPaid = (p.RemainingBalance <= 0)
	if p.IsFullyPaid && p.FullyPaidDate == nil {
		now := time.Now()
		p.FullyPaidDate = &now
	}
}

// HasPendingVerification checks if payment has any unverified transactions
func (p *Payment) HasPendingVerification() bool {
	for _, tx := range p.Transactions {
		if tx != nil && tx.VerifiedAt == nil {
			return true
		}
	}
	return false
}

// GetUserFacingStatus refines status: shows "Pending verification" when a txn is submitted
// Now considers partial payments and transaction verification status
func (p *Payment) GetUserFacingStatus() string {
	if p.IsFullyPaid {
		return "Fully Paid"
	}
	hasPendingVerification := p.HasPendingVerification()
	if p.AmountPaid > 0 {
		// Partially paid
		if hasPendingVerification {
			return "Partially Paid (Pending Verification)"
		}
		return "Partially Paid"
	}
	// Not paid yet
	if hasPendingVerification {
		return "Pending Verification"
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

// GetFormattedAmountPaid returns the amount paid formatted as currency
func (p *Payment) GetFormattedAmountPaid() string {
	return fmt.Sprintf("₹%d", p.AmountPaid)
}

// GetFormattedRemainingBalance returns the remaining balance formatted as currency
func (p *Payment) GetFormattedRemainingBalance() string {
	return fmt.Sprintf("₹%d", p.RemainingBalance)
}

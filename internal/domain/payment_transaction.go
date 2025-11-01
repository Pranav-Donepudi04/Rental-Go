package domain

import (
	"fmt"
	"time"
)

// PaymentTransaction represents a transaction submitted by a tenant
// for a payment. The amount is NULL until the owner verifies it.
type PaymentTransaction struct {
	ID               int        `json:"id" db:"id"`
	PaymentID        int        `json:"payment_id" db:"payment_id"`
	TransactionID    string     `json:"transaction_id" db:"transaction_id"`
	Amount           *int       `json:"amount" db:"amount"` // NULL until owner verifies
	SubmittedAt      time.Time  `json:"submitted_at" db:"submitted_at"`
	VerifiedAt       *time.Time `json:"verified_at" db:"verified_at"`
	VerifiedByUserID *int       `json:"verified_by_user_id" db:"verified_by_user_id"`
	Notes            string     `json:"notes" db:"notes"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`

	// Related data (populated by joins)
	Payment *Payment `json:"payment,omitempty"`
}

// IsVerified returns true if the transaction has been verified by owner
func (pt *PaymentTransaction) IsVerified() bool {
	return pt.VerifiedAt != nil && pt.Amount != nil
}

// IsPending returns true if transaction is pending verification
func (pt *PaymentTransaction) IsPending() bool {
	return !pt.IsVerified()
}

// GetFormattedAmount returns the amount formatted as currency, or "Not verified" if NULL
func (pt *PaymentTransaction) GetFormattedAmount() string {
	if pt.Amount == nil {
		return "Not verified"
	}
	return fmt.Sprintf("₹%d", *pt.Amount)
}

// GetFormattedSubmittedAt returns the submitted date formatted
func (pt *PaymentTransaction) GetFormattedSubmittedAt() string {
	return pt.SubmittedAt.Format("Jan 2, 2006 3:04 PM")
}

// GetFormattedVerifiedAt returns the verified date formatted, or "Not verified"
func (pt *PaymentTransaction) GetFormattedVerifiedAt() string {
	if pt.VerifiedAt == nil {
		return "Not verified"
	}
	return pt.VerifiedAt.Format("Jan 2, 2006 3:04 PM")
}

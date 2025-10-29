package models

import (
	"fmt"
	"time"
)

// Unit represents a rental unit in the property
type Unit struct {
	ID              int       `json:"id" db:"id"`
	UnitCode        string    `json:"unit_code" db:"unit_code"`
	Floor           string    `json:"floor" db:"floor"`
	UnitType        string    `json:"unit_type" db:"unit_type"`
	MonthlyRent     int       `json:"monthly_rent" db:"monthly_rent"`
	SecurityDeposit int       `json:"security_deposit" db:"security_deposit"`
	PaymentDueDay   int       `json:"payment_due_day" db:"payment_due_day"`
	IsOccupied      bool      `json:"is_occupied" db:"is_occupied"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// Validate validates the unit data
func (u *Unit) Validate() error {
	if u.UnitCode == "" {
		return fmt.Errorf("unit code is required")
	}
	if u.Floor == "" {
		return fmt.Errorf("floor is required")
	}
	if u.UnitType == "" {
		return fmt.Errorf("unit type is required")
	}
	if u.MonthlyRent <= 0 {
		return fmt.Errorf("monthly rent must be greater than 0")
	}
	if u.SecurityDeposit <= 0 {
		return fmt.Errorf("security deposit must be greater than 0")
	}
	if u.PaymentDueDay < 1 || u.PaymentDueDay > 31 {
		return fmt.Errorf("payment due day must be between 1 and 31")
	}
	return nil
}

// GetDisplayName returns a formatted display name for the unit
func (u *Unit) GetDisplayName() string {
	return fmt.Sprintf("%s - %s (%s)", u.UnitCode, u.UnitType, u.Floor)
}

// GetRentInfo returns formatted rent information
func (u *Unit) GetRentInfo() string {
	return fmt.Sprintf("₹%d/month, Deposit: ₹%d, Due: %dth",
		u.MonthlyRent, u.SecurityDeposit, u.PaymentDueDay)
}

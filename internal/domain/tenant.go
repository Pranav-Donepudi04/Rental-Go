package domain

import (
	"fmt"
	"strings"
	"time"
)

// Tenant represents the primary rent payer
type Tenant struct {
	ID             int       `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	Phone          string    `json:"phone" db:"phone"`
	AadharNumber   string    `json:"aadhar_number" db:"aadhar_number"`
	MoveInDate     time.Time `json:"move_in_date" db:"move_in_date"`
	NumberOfPeople int       `json:"number_of_people" db:"number_of_people"`
	UnitID         int       `json:"unit_id" db:"unit_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`

	// Related data (populated by joins)
	Unit          *Unit           `json:"unit,omitempty"`
	FamilyMembers []*FamilyMember `json:"family_members,omitempty"`
}

// Validate validates the tenant data
func (t *Tenant) Validate() error {
	if strings.TrimSpace(t.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(t.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	if strings.TrimSpace(t.AadharNumber) == "" {
		return fmt.Errorf("aadhar number is required")
	}
	if len(t.AadharNumber) != 12 {
		return fmt.Errorf("aadhar number must be 12 digits")
	}
	if t.MoveInDate.IsZero() {
		return fmt.Errorf("move-in date is required")
	}
	if t.NumberOfPeople <= 0 {
		return fmt.Errorf("number of people must be greater than 0")
	}
	if t.UnitID <= 0 {
		return fmt.Errorf("unit ID is required")
	}
	return nil
}

// GetDisplayName returns a formatted display name for the tenant
func (t *Tenant) GetDisplayName() string {
	return fmt.Sprintf("%s (%s)", t.Name, t.Phone)
}

// GetMoveInDuration returns how long the tenant has been staying
func (t *Tenant) GetMoveInDuration() string {
	duration := time.Since(t.MoveInDate)
	months := int(duration.Hours() / 24 / 30)
	if months < 1 {
		return "Less than 1 month"
	}
	return fmt.Sprintf("%d months", months)
}

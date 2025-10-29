package models

import (
	"fmt"
	"strings"
	"time"
)

// FamilyMember represents a family member of a tenant
type FamilyMember struct {
	ID           int       `json:"id" db:"id"`
	TenantID     int       `json:"tenant_id" db:"tenant_id"`
	Name         string    `json:"name" db:"name"`
	Age          int       `json:"age" db:"age"`
	Relationship string    `json:"relationship" db:"relationship"`
	AadharNumber string    `json:"aadhar_number" db:"aadhar_number"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Validate validates the family member data
func (fm *FamilyMember) Validate() error {
	if strings.TrimSpace(fm.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if fm.Age <= 0 {
		return fmt.Errorf("age must be greater than 0")
	}
	if strings.TrimSpace(fm.Relationship) == "" {
		return fmt.Errorf("relationship is required")
	}
	if fm.TenantID <= 0 {
		return fmt.Errorf("tenant ID is required")
	}
	// Aadhar number is optional, but if provided, should be 12 digits
	if fm.AadharNumber != "" && len(fm.AadharNumber) != 12 {
		return fmt.Errorf("aadhar number must be 12 digits if provided")
	}
	return nil
}

// GetDisplayName returns a formatted display name for the family member
func (fm *FamilyMember) GetDisplayName() string {
	return fmt.Sprintf("%s (%s, %d years)", fm.Name, fm.Relationship, fm.Age)
}

// HasAadhar returns true if the family member has an aadhar number
func (fm *FamilyMember) HasAadhar() bool {
	return fm.AadharNumber != ""
}

package service

import (
	"backend-form/m/internal/domain"
	interfaces "backend-form/m/internal/repository/interfaces"
	"fmt"
	"time"
)

// TenantService handles tenant-related business logic
type TenantService struct {
	tenantRepo     interfaces.TenantRepository
	unitRepo       interfaces.UnitRepository
	paymentService *PaymentService
}

// NewTenantService creates a new TenantService
func NewTenantService(tenantRepo interfaces.TenantRepository, unitRepo interfaces.UnitRepository, paymentService *PaymentService) *TenantService {
	return &TenantService{
		tenantRepo:     tenantRepo,
		unitRepo:       unitRepo,
		paymentService: paymentService,
	}
}

// CreateTenant creates a new tenant and updates unit occupancy
// skipFirstPayment: if true, skips automatic first payment creation (useful for existing tenants)
func (s *TenantService) CreateTenant(tenant *domain.Tenant, skipFirstPayment bool) error {
	// Validate tenant data
	if err := tenant.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if unit exists and is available
	unit, err := s.unitRepo.GetUnitByID(tenant.UnitID)
	if err != nil {
		return fmt.Errorf("unit not found: %w", err)
	}

	if unit.IsOccupied {
		return fmt.Errorf("unit %s is already occupied", unit.UnitCode)
	}

	// Create tenant
	if err := s.tenantRepo.CreateTenant(tenant); err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	// Update unit occupancy
	if err := s.unitRepo.UpdateUnitOccupancy(tenant.UnitID, true); err != nil {
		// Rollback tenant creation if unit update fails
		s.tenantRepo.DeleteTenant(tenant.ID)
		return fmt.Errorf("failed to update unit occupancy: %w", err)
	}

	// Create first payment immediately (unless skipped for existing tenants)
	if !skipFirstPayment {
		if err := s.createFirstPayment(tenant); err != nil {
			// Log error but don't fail tenant creation
			// Payment can be created manually if needed
			fmt.Printf("Warning: Failed to create first payment for tenant %d: %v\n", tenant.ID, err)
		}
	}

	return nil
}

// createFirstPayment creates the first payment for a tenant based on move-in date
func (s *TenantService) createFirstPayment(tenant *domain.Tenant) error {
	// Get unit
	unit, err := s.unitRepo.GetUnitByID(tenant.UnitID)
	if err != nil {
		return fmt.Errorf("unit not found: %w", err)
	}

	// Calculate first due date: Next 10th >= move_in_date
	moveInDate := tenant.MoveInDate
	firstDueDate := time.Date(moveInDate.Year(), moveInDate.Month(), unit.PaymentDueDay, 0, 0, 0, 0, moveInDate.Location())

	// If move-in date is after due date in same month, use next month
	if moveInDate.Day() > unit.PaymentDueDay {
		firstDueDate = firstDueDate.AddDate(0, 1, 0) // Next month
	}

	// Use PaymentService to create payment (ensures consistent business logic)
	_, err = s.paymentService.CreatePaymentForTenant(
		tenant.ID,
		tenant.UnitID,
		firstDueDate,
		unit.MonthlyRent,
	)
	return err
}

// GetTenantByID returns a tenant by ID with related data
func (s *TenantService) GetTenantByID(id int) (*domain.Tenant, error) {
	tenant, err := s.tenantRepo.GetTenantByID(id)
	if err != nil {
		return nil, err
	}

	// Load related data
	if tenant.UnitID > 0 {
		unit, err := s.unitRepo.GetUnitByID(tenant.UnitID)
		if err == nil {
			tenant.Unit = unit
		}
	}

	// Load family members
	familyMembers, err := s.tenantRepo.GetFamilyMembersByTenantID(id)
	if err == nil {
		tenant.FamilyMembers = familyMembers
	}

	return tenant, nil
}

// GetAllTenants returns all tenants with related data
func (s *TenantService) GetAllTenants() ([]*domain.Tenant, error) {
	tenants, err := s.tenantRepo.GetAllTenants()
	if err != nil {
		return nil, err
	}

	// Load related data for each tenant
	for _, tenant := range tenants {
		if tenant.UnitID > 0 {
			unit, err := s.unitRepo.GetUnitByID(tenant.UnitID)
			if err == nil {
				tenant.Unit = unit
			}
		}
	}

	return tenants, nil
}

// UpdateTenant updates tenant information
func (s *TenantService) UpdateTenant(tenant *domain.Tenant) error {
	if err := tenant.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return s.tenantRepo.UpdateTenant(tenant)
}

// MoveOutTenant handles tenant move-out process
func (s *TenantService) MoveOutTenant(tenantID int) error {
	tenant, err := s.tenantRepo.GetTenantByID(tenantID)
	if err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	// Delete all payments for this tenant before deleting tenant
	// This is required because of foreign key constraint: payments.tenant_id → tenants.id
	// Note: If you want to keep payment history, you'd need to either:
	// 1. Soft delete tenants (add is_deleted flag) instead of hard delete
	// 2. Allow NULL tenant_id in payments (schema change)
	// 3. Move payments to a separate history table before deleting tenant
	if err := s.paymentService.DeletePaymentsByTenantID(tenantID); err != nil {
		// Log warning but continue - if deletion fails, tenant deletion will also fail with FK error
		fmt.Printf("Warning: Failed to delete payments for tenant %d: %v\n", tenantID, err)
	}

	// Delete tenant (this will cascade delete family members and payment_transactions)
	if err := s.tenantRepo.DeleteTenant(tenantID); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	// Update unit occupancy
	if err := s.unitRepo.UpdateUnitOccupancy(tenant.UnitID, false); err != nil {
		return fmt.Errorf("failed to update unit occupancy: %w", err)
	}

	return nil
}

// AddFamilyMember adds a family member to a tenant
func (s *TenantService) AddFamilyMember(familyMember *domain.FamilyMember) error {
	if err := familyMember.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return s.tenantRepo.CreateFamilyMember(familyMember)
}

// GetFamilyMembersByTenantID returns family members for a tenant
func (s *TenantService) GetFamilyMembersByTenantID(tenantID int) ([]*domain.FamilyMember, error) {
	return s.tenantRepo.GetFamilyMembersByTenantID(tenantID)
}

// GetTenantsByUnitID returns tenants for a specific unit
func (s *TenantService) GetTenantsByUnitID(unitID int) ([]*domain.Tenant, error) {
	return s.tenantRepo.GetTenantsByUnitID(unitID)
}

// GetTenantSummary returns a summary of tenants
func (s *TenantService) GetTenantSummary() (*TenantSummary, error) {
	tenants, err := s.tenantRepo.GetAllTenants()
	if err != nil {
		return nil, err
	}

	summary := &TenantSummary{
		TotalTenants: len(tenants),
		NewThisMonth: 0,
		TotalPeople:  0,
	}

	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	for _, tenant := range tenants {
		summary.TotalPeople += tenant.NumberOfPeople
		if tenant.MoveInDate.After(firstOfMonth) {
			summary.NewThisMonth++
		}
	}

	return summary, nil
}

// TenantSummary represents tenant summary
type TenantSummary struct {
	TotalTenants int `json:"total_tenants"`
	NewThisMonth int `json:"new_this_month"`
	TotalPeople  int `json:"total_people"`
}

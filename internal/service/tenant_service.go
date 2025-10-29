package service

import (
	"backend-form/m/internal/models"
	"backend-form/m/internal/repository"
	"fmt"
	"time"
)

// TenantService handles tenant-related business logic
type TenantService struct {
	tenantRepo  repository.TenantRepository
	unitRepo    repository.UnitRepository
	paymentRepo repository.PaymentRepository
}

// NewTenantService creates a new TenantService
func NewTenantService(tenantRepo repository.TenantRepository, unitRepo repository.UnitRepository, paymentRepo repository.PaymentRepository) *TenantService {
	return &TenantService{
		tenantRepo:  tenantRepo,
		unitRepo:    unitRepo,
		paymentRepo: paymentRepo,
	}
}

// CreateTenant creates a new tenant and updates unit occupancy
func (s *TenantService) CreateTenant(tenant *models.Tenant) error {
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

	return nil
}

// GetTenantByID returns a tenant by ID with related data
func (s *TenantService) GetTenantByID(id int) (*models.Tenant, error) {
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
func (s *TenantService) GetAllTenants() ([]*models.Tenant, error) {
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
func (s *TenantService) UpdateTenant(tenant *models.Tenant) error {
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

	// Delete all payments for this tenant first (to avoid foreign key constraint)
	if err := s.paymentRepo.DeletePaymentsByTenantID(tenantID); err != nil {
		return fmt.Errorf("failed to delete tenant payments: %w", err)
	}

	// Delete tenant (this will cascade delete family members)
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
func (s *TenantService) AddFamilyMember(familyMember *models.FamilyMember) error {
	if err := familyMember.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return s.tenantRepo.CreateFamilyMember(familyMember)
}

// GetFamilyMembersByTenantID returns family members for a tenant
func (s *TenantService) GetFamilyMembersByTenantID(tenantID int) ([]*models.FamilyMember, error) {
	return s.tenantRepo.GetFamilyMembersByTenantID(tenantID)
}

// GetTenantsByUnitID returns tenants for a specific unit
func (s *TenantService) GetTenantsByUnitID(unitID int) ([]*models.Tenant, error) {
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

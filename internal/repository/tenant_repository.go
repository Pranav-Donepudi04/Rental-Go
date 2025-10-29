package repository

import "backend-form/m/internal/models"

// TenantRepository defines the interface for tenant data operations
type TenantRepository interface {
	CreateTenant(tenant *models.Tenant) error
	GetTenantByID(id int) (*models.Tenant, error)
	GetAllTenants() ([]*models.Tenant, error)
	UpdateTenant(tenant *models.Tenant) error
	DeleteTenant(id int) error
	GetTenantsByUnitID(unitID int) ([]*models.Tenant, error)

	// Family member operations
	CreateFamilyMember(familyMember *models.FamilyMember) error
	GetFamilyMembersByTenantID(tenantID int) ([]*models.FamilyMember, error)
	UpdateFamilyMember(familyMember *models.FamilyMember) error
	DeleteFamilyMember(id int) error
}

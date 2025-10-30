package interfaces

import "backend-form/m/internal/domain"

// TenantRepository defines the interface for tenant data operations
type TenantRepository interface {
	CreateTenant(tenant *domain.Tenant) error
	GetTenantByID(id int) (*domain.Tenant, error)
	GetAllTenants() ([]*domain.Tenant, error)
	UpdateTenant(tenant *domain.Tenant) error
	DeleteTenant(id int) error
	GetTenantsByUnitID(unitID int) ([]*domain.Tenant, error)

	// Family member operations
	CreateFamilyMember(familyMember *domain.FamilyMember) error
	GetFamilyMembersByTenantID(tenantID int) ([]*domain.FamilyMember, error)
	UpdateFamilyMember(familyMember *domain.FamilyMember) error
	DeleteFamilyMember(id int) error
}

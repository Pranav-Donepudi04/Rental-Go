package interfaces

import "backend-form/m/internal/domain"

type UserRepository interface {
	GetByID(id int) (*domain.User, error)
	GetByPhone(phone string) (*domain.User, error)
	CreateTenantUser(user *domain.User) error
	UpdatePassword(userID int, newHash string) error
	LinkTenant(userID int, tenantID int) error
}

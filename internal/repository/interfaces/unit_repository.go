package interfaces

import "backend-form/m/internal/domain"

// UnitRepository defines the interface for unit data operations
type UnitRepository interface {
	GetAllUnits() ([]*domain.Unit, error)
	GetUnitByID(id int) (*domain.Unit, error)
	GetUnitByCode(code string) (*domain.Unit, error)
	UpdateUnitOccupancy(unitID int, isOccupied bool) error
}

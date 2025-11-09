package interfaces

import "backend-form/m/internal/domain"

// UnitRepository defines the interface for unit data operations
type UnitRepository interface {
	GetAllUnits() ([]*domain.Unit, error)
	GetUnitByID(id int) (*domain.Unit, error)
	GetUnitByCode(code string) (*domain.Unit, error)
	GetUnitsByIDs(ids []int) (map[int]*domain.Unit, error) // Bulk load units by IDs (fixes N+1)
	UpdateUnitOccupancy(unitID int, isOccupied bool) error
}

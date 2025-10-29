package repository

import "backend-form/m/internal/models"

// UnitRepository defines the interface for unit data operations
type UnitRepository interface {
	GetAllUnits() ([]*models.Unit, error)
	GetUnitByID(id int) (*models.Unit, error)
	GetUnitByCode(code string) (*models.Unit, error)
	UpdateUnitOccupancy(unitID int, isOccupied bool) error
}

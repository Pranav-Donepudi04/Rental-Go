package service

import (
	"backend-form/m/internal/domain"
	interfaces "backend-form/m/internal/repository/interfaces"
	"fmt"
)

// UnitService handles unit-related business logic
type UnitService struct {
	unitRepo interfaces.UnitRepository
}

// NewUnitService creates a new UnitService
func NewUnitService(unitRepo interfaces.UnitRepository) *UnitService {
	return &UnitService{
		unitRepo: unitRepo,
	}
}

// GetAllUnits returns all units
func (s *UnitService) GetAllUnits() ([]*domain.Unit, error) {
	return s.unitRepo.GetAllUnits()
}

// GetUnitByID returns a unit by ID
func (s *UnitService) GetUnitByID(id int) (*domain.Unit, error) {
	return s.unitRepo.GetUnitByID(id)
}

// GetUnitByCode returns a unit by unit code
func (s *UnitService) GetUnitByCode(code string) (*domain.Unit, error) {
	return s.unitRepo.GetUnitByCode(code)
}

// GetAvailableUnits returns units that are not occupied
func (s *UnitService) GetAvailableUnits() ([]*domain.Unit, error) {
	units, err := s.unitRepo.GetAllUnits()
	if err != nil {
		return nil, err
	}

	var available []*domain.Unit
	for _, unit := range units {
		if !unit.IsOccupied {
			available = append(available, unit)
		}
	}

	return available, nil
}

// GetOccupiedUnits returns units that are occupied
func (s *UnitService) GetOccupiedUnits() ([]*domain.Unit, error) {
	units, err := s.unitRepo.GetAllUnits()
	if err != nil {
		return nil, err
	}

	var occupied []*domain.Unit
	for _, unit := range units {
		if unit.IsOccupied {
			occupied = append(occupied, unit)
		}
	}

	return occupied, nil
}

// UpdateUnitOccupancy updates the occupancy status of a unit
func (s *UnitService) UpdateUnitOccupancy(unitID int, isOccupied bool) error {
	return s.unitRepo.UpdateUnitOccupancy(unitID, isOccupied)
}

// GetUnitsByFloor returns units grouped by floor
func (s *UnitService) GetUnitsByFloor() (map[string][]*domain.Unit, error) {
	units, err := s.unitRepo.GetAllUnits()
	if err != nil {
		return nil, err
	}

	floorMap := make(map[string][]*domain.Unit)
	for _, unit := range units {
		floorMap[unit.Floor] = append(floorMap[unit.Floor], unit)
	}

	return floorMap, nil
}

// GetRentalSummary returns a summary of rental income
func (s *UnitService) GetRentalSummary() (*RentalSummary, error) {
	units, err := s.unitRepo.GetAllUnits()
	if err != nil {
		return nil, err
	}

	summary := &RentalSummary{
		TotalUnits:          len(units),
		OccupiedUnits:       0,
		AvailableUnits:      0,
		TotalMonthlyRent:    0,
		OccupiedMonthlyRent: 0,
	}

	for _, unit := range units {
		summary.TotalMonthlyRent += unit.MonthlyRent
		if unit.IsOccupied {
			summary.OccupiedUnits++
			summary.OccupiedMonthlyRent += unit.MonthlyRent
		} else {
			summary.AvailableUnits++
		}
	}

	return summary, nil
}

// RentalSummary represents rental income summary
type RentalSummary struct {
	TotalUnits          int `json:"total_units"`
	OccupiedUnits       int `json:"occupied_units"`
	AvailableUnits      int `json:"available_units"`
	TotalMonthlyRent    int `json:"total_monthly_rent"`
	OccupiedMonthlyRent int `json:"occupied_monthly_rent"`
}

// GetFormattedTotalRent returns formatted total rent
func (rs *RentalSummary) GetFormattedTotalRent() string {
	return fmt.Sprintf("₹%d", rs.TotalMonthlyRent)
}

// GetFormattedOccupiedRent returns formatted occupied rent
func (rs *RentalSummary) GetFormattedOccupiedRent() string {
	return fmt.Sprintf("₹%d", rs.OccupiedMonthlyRent)
}

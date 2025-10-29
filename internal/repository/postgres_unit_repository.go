package repository

import (
	"backend-form/m/internal/models"
	"database/sql"
	"fmt"
)

// PostgresUnitRepository implements UnitRepository interface
type PostgresUnitRepository struct {
	db *sql.DB
}

// NewPostgresUnitRepository creates a new PostgresUnitRepository
func NewPostgresUnitRepository(db *sql.DB) UnitRepository {
	return &PostgresUnitRepository{db: db}
}

// GetAllUnits returns all units
func (r *PostgresUnitRepository) GetAllUnits() ([]*models.Unit, error) {
	query := `
		SELECT id, unit_code, floor, unit_type, monthly_rent, security_deposit, 
		       payment_due_day, is_occupied, created_at
		FROM units
		ORDER BY floor, unit_code`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query units: %w", err)
	}
	defer rows.Close()

	var units []*models.Unit
	for rows.Next() {
		unit := &models.Unit{}
		err := rows.Scan(
			&unit.ID,
			&unit.UnitCode,
			&unit.Floor,
			&unit.UnitType,
			&unit.MonthlyRent,
			&unit.SecurityDeposit,
			&unit.PaymentDueDay,
			&unit.IsOccupied,
			&unit.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan unit: %w", err)
		}
		units = append(units, unit)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating units: %w", err)
	}

	return units, nil
}

// GetUnitByID returns a unit by ID
func (r *PostgresUnitRepository) GetUnitByID(id int) (*models.Unit, error) {
	query := `
		SELECT id, unit_code, floor, unit_type, monthly_rent, security_deposit, 
		       payment_due_day, is_occupied, created_at
		FROM units
		WHERE id = $1`

	unit := &models.Unit{}
	err := r.db.QueryRow(query, id).Scan(
		&unit.ID,
		&unit.UnitCode,
		&unit.Floor,
		&unit.UnitType,
		&unit.MonthlyRent,
		&unit.SecurityDeposit,
		&unit.PaymentDueDay,
		&unit.IsOccupied,
		&unit.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("unit with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get unit: %w", err)
	}

	return unit, nil
}

// GetUnitByCode returns a unit by unit code
func (r *PostgresUnitRepository) GetUnitByCode(code string) (*models.Unit, error) {
	query := `
		SELECT id, unit_code, floor, unit_type, monthly_rent, security_deposit, 
		       payment_due_day, is_occupied, created_at
		FROM units
		WHERE unit_code = $1`

	unit := &models.Unit{}
	err := r.db.QueryRow(query, code).Scan(
		&unit.ID,
		&unit.UnitCode,
		&unit.Floor,
		&unit.UnitType,
		&unit.MonthlyRent,
		&unit.SecurityDeposit,
		&unit.PaymentDueDay,
		&unit.IsOccupied,
		&unit.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("unit with code %s not found", code)
		}
		return nil, fmt.Errorf("failed to get unit: %w", err)
	}

	return unit, nil
}

// UpdateUnitOccupancy updates the occupancy status of a unit
func (r *PostgresUnitRepository) UpdateUnitOccupancy(unitID int, isOccupied bool) error {
	query := `UPDATE units SET is_occupied = $1 WHERE id = $2`

	result, err := r.db.Exec(query, isOccupied, unitID)
	if err != nil {
		return fmt.Errorf("failed to update unit occupancy: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("unit with ID %d not found", unitID)
	}

	return nil
}

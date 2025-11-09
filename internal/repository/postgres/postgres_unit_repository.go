package repository

import (
	domain "backend-form/m/internal/domain"
	"backend-form/m/internal/repository/interfaces"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

// PostgresUnitRepository implements UnitRepository interface
type PostgresUnitRepository struct {
	db *sql.DB
}

// NewPostgresUnitRepository creates a new PostgresUnitRepository
func NewPostgresUnitRepository(db *sql.DB) interfaces.UnitRepository {
	return &PostgresUnitRepository{db: db}
}

// GetAllUnits returns all units
func (r *PostgresUnitRepository) GetAllUnits() ([]*domain.Unit, error) {
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

	var units []*domain.Unit
	for rows.Next() {
		unit := &domain.Unit{}
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
func (r *PostgresUnitRepository) GetUnitByID(id int) (*domain.Unit, error) {
	query := `
		SELECT id, unit_code, floor, unit_type, monthly_rent, security_deposit, 
		       payment_due_day, is_occupied, created_at
		FROM units
		WHERE id = $1`

	unit := &domain.Unit{}
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
func (r *PostgresUnitRepository) GetUnitByCode(code string) (*domain.Unit, error) {
	query := `
		SELECT id, unit_code, floor, unit_type, monthly_rent, security_deposit, 
		       payment_due_day, is_occupied, created_at
		FROM units
		WHERE unit_code = $1`

	unit := &domain.Unit{}
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

// GetUnitsByIDs returns units by their IDs in a map for efficient lookup (fixes N+1 queries)
func (r *PostgresUnitRepository) GetUnitsByIDs(ids []int) (map[int]*domain.Unit, error) {
	if len(ids) == 0 {
		return make(map[int]*domain.Unit), nil
	}

	// Build query with IN clause using PostgreSQL array
	query := `
		SELECT id, unit_code, floor, unit_type, monthly_rent, security_deposit, 
		       payment_due_day, is_occupied, created_at
		FROM units
		WHERE id = ANY($1)
		ORDER BY id`

	// Convert []int to PostgreSQL array
	rows, err := r.db.Query(query, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("failed to query units by IDs: %w", err)
	}
	defer rows.Close()

	units := make(map[int]*domain.Unit)
	for rows.Next() {
		unit := &domain.Unit{}
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
		units[unit.ID] = unit
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating units: %w", err)
	}

	return units, nil
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

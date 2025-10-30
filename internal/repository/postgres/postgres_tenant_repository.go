package repository

import (
	domain "backend-form/m/internal/domain"
	"backend-form/m/internal/repository/interfaces"
	"database/sql"
	"fmt"
)

// PostgresTenantRepository implements TenantRepository interface
type PostgresTenantRepository struct {
	db *sql.DB
}

// NewPostgresTenantRepository creates a new PostgresTenantRepository
func NewPostgresTenantRepository(db *sql.DB) interfaces.TenantRepository {
	return &PostgresTenantRepository{db: db}
}

// CreateTenant creates a new tenant
func (r *PostgresTenantRepository) CreateTenant(tenant *domain.Tenant) error {
	query := `
		INSERT INTO tenants (name, phone, aadhar_number, move_in_date, number_of_people, unit_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	err := r.db.QueryRow(query,
		tenant.Name,
		tenant.Phone,
		tenant.AadharNumber,
		tenant.MoveInDate,
		tenant.NumberOfPeople,
		tenant.UnitID,
	).Scan(&tenant.ID, &tenant.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	return nil
}

// GetTenantByID returns a tenant by ID
func (r *PostgresTenantRepository) GetTenantByID(id int) (*domain.Tenant, error) {
	query := `
		SELECT id, name, phone, aadhar_number, move_in_date, number_of_people, unit_id, created_at
		FROM tenants
		WHERE id = $1`

	tenant := &domain.Tenant{}
	err := r.db.QueryRow(query, id).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Phone,
		&tenant.AadharNumber,
		&tenant.MoveInDate,
		&tenant.NumberOfPeople,
		&tenant.UnitID,
		&tenant.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}

// GetAllTenants returns all tenants
func (r *PostgresTenantRepository) GetAllTenants() ([]*domain.Tenant, error) {
	query := `
		SELECT id, name, phone, aadhar_number, move_in_date, number_of_people, unit_id, created_at
		FROM tenants
		ORDER BY name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*domain.Tenant
	for rows.Next() {
		tenant := &domain.Tenant{}
		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Phone,
			&tenant.AadharNumber,
			&tenant.MoveInDate,
			&tenant.NumberOfPeople,
			&tenant.UnitID,
			&tenant.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenants = append(tenants, tenant)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tenants: %w", err)
	}

	return tenants, nil
}

// UpdateTenant updates tenant information
func (r *PostgresTenantRepository) UpdateTenant(tenant *domain.Tenant) error {
	query := `
		UPDATE tenants 
		SET name = $1, phone = $2, aadhar_number = $3, move_in_date = $4, 
		    number_of_people = $5, unit_id = $6
		WHERE id = $7`

	result, err := r.db.Exec(query,
		tenant.Name,
		tenant.Phone,
		tenant.AadharNumber,
		tenant.MoveInDate,
		tenant.NumberOfPeople,
		tenant.UnitID,
		tenant.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tenant with ID %d not found", tenant.ID)
	}

	return nil
}

// DeleteTenant deletes a tenant
func (r *PostgresTenantRepository) DeleteTenant(id int) error {
	query := `DELETE FROM tenants WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tenant with ID %d not found", id)
	}

	return nil
}

// GetTenantsByUnitID returns tenants for a specific unit
func (r *PostgresTenantRepository) GetTenantsByUnitID(unitID int) ([]*domain.Tenant, error) {
	query := `
		SELECT id, name, phone, aadhar_number, move_in_date, number_of_people, unit_id, created_at
		FROM tenants
		WHERE unit_id = $1
		ORDER BY name`

	rows, err := r.db.Query(query, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tenants by unit: %w", err)
	}
	defer rows.Close()

	var tenants []*domain.Tenant
	for rows.Next() {
		tenant := &domain.Tenant{}
		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Phone,
			&tenant.AadharNumber,
			&tenant.MoveInDate,
			&tenant.NumberOfPeople,
			&tenant.UnitID,
			&tenant.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenants = append(tenants, tenant)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tenants: %w", err)
	}

	return tenants, nil
}

// CreateFamilyMember creates a new family member
func (r *PostgresTenantRepository) CreateFamilyMember(familyMember *domain.FamilyMember) error {
	query := `
		INSERT INTO family_members (tenant_id, name, age, relationship, aadhar_number)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	err := r.db.QueryRow(query,
		familyMember.TenantID,
		familyMember.Name,
		familyMember.Age,
		familyMember.Relationship,
		familyMember.AadharNumber,
	).Scan(&familyMember.ID, &familyMember.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create family member: %w", err)
	}

	return nil
}

// GetFamilyMembersByTenantID returns family members for a tenant
func (r *PostgresTenantRepository) GetFamilyMembersByTenantID(tenantID int) ([]*domain.FamilyMember, error) {
	query := `
		SELECT id, tenant_id, name, age, relationship, aadhar_number, created_at
		FROM family_members
		WHERE tenant_id = $1
		ORDER BY name`

	rows, err := r.db.Query(query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query family members: %w", err)
	}
	defer rows.Close()

	var familyMembers []*domain.FamilyMember
	for rows.Next() {
		familyMember := &domain.FamilyMember{}
		err := rows.Scan(
			&familyMember.ID,
			&familyMember.TenantID,
			&familyMember.Name,
			&familyMember.Age,
			&familyMember.Relationship,
			&familyMember.AadharNumber,
			&familyMember.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan family member: %w", err)
		}
		familyMembers = append(familyMembers, familyMember)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating family members: %w", err)
	}

	return familyMembers, nil
}

// UpdateFamilyMember updates family member information
func (r *PostgresTenantRepository) UpdateFamilyMember(familyMember *domain.FamilyMember) error {
	query := `
		UPDATE family_members 
		SET name = $1, age = $2, relationship = $3, aadhar_number = $4
		WHERE id = $5`

	result, err := r.db.Exec(query,
		familyMember.Name,
		familyMember.Age,
		familyMember.Relationship,
		familyMember.AadharNumber,
		familyMember.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update family member: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("family member with ID %d not found", familyMember.ID)
	}

	return nil
}

// DeleteFamilyMember deletes a family member
func (r *PostgresTenantRepository) DeleteFamilyMember(id int) error {
	query := `DELETE FROM family_members WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete family member: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("family member with ID %d not found", id)
	}

	return nil
}

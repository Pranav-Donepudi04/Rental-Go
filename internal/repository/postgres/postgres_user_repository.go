package repository

import (
	"backend-form/m/internal/domain"
	"backend-form/m/internal/repository/interfaces"
	"database/sql"
	"fmt"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) interfaces.UserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) GetByPhone(phone string) (*domain.User, error) {
	const q = `SELECT id, phone, password_hash, user_type, tenant_id, is_active, created_at FROM users WHERE phone = $1`
	u := &domain.User{}
	var tenantID sql.NullInt64
	if err := r.db.QueryRow(q, phone).Scan(&u.ID, &u.Phone, &u.PasswordHash, &u.UserType, &tenantID, &u.IsActive, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by phone: %w", err)
	}
	if tenantID.Valid {
		v := int(tenantID.Int64)
		u.TenantID = &v
	}
	return u, nil
}

func (r *PostgresUserRepository) GetByID(id int) (*domain.User, error) {
	const q = `SELECT id, phone, password_hash, user_type, tenant_id, is_active, created_at FROM users WHERE id = $1`
	u := &domain.User{}
	var tenantID sql.NullInt64
	if err := r.db.QueryRow(q, id).Scan(&u.ID, &u.Phone, &u.PasswordHash, &u.UserType, &tenantID, &u.IsActive, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	if tenantID.Valid {
		v := int(tenantID.Int64)
		u.TenantID = &v
	}
	return u, nil
}

func (r *PostgresUserRepository) CreateTenantUser(user *domain.User) error {
	const q = `INSERT INTO users (phone, password_hash, user_type, tenant_id, is_active) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	var tenantID interface{}
	if user.TenantID != nil {
		tenantID = *user.TenantID
	}
	if err := r.db.QueryRow(q, user.Phone, user.PasswordHash, user.UserType, tenantID, user.IsActive).Scan(&user.ID, &user.CreatedAt); err != nil {
		return fmt.Errorf("create tenant user: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) UpdatePassword(userID int, newHash string) error {
	const q = `UPDATE users SET password_hash = $1 WHERE id = $2`
	if _, err := r.db.Exec(q, newHash, userID); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) LinkTenant(userID int, tenantID int) error {
	const q = `UPDATE users SET tenant_id = $1 WHERE id = $2`
	if _, err := r.db.Exec(q, tenantID, userID); err != nil {
		return fmt.Errorf("link tenant: %w", err)
	}
	return nil
}

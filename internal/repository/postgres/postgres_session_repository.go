package repository

import (
	"backend-form/m/internal/domain"
	"backend-form/m/internal/repository/interfaces"
	"database/sql"
	"fmt"
)

type PostgresSessionRepository struct {
	db *sql.DB
}

func NewPostgresSessionRepository(db *sql.DB) interfaces.SessionRepository {
	return &PostgresSessionRepository{db: db}
}

func (r *PostgresSessionRepository) Create(s *domain.Session) error {
	const q = `INSERT INTO sessions (user_id, token, created_at, expires_at) VALUES ($1, $2, $3, $4) RETURNING id`
	if err := r.db.QueryRow(q, s.UserID, s.Token, s.CreatedAt, s.ExpiresAt).Scan(&s.ID); err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (r *PostgresSessionRepository) GetByToken(token string) (*domain.Session, error) {
	const q = `SELECT id, user_id, token, created_at, expires_at FROM sessions WHERE token = $1`
	s := &domain.Session{}
	if err := r.db.QueryRow(q, token).Scan(&s.ID, &s.UserID, &s.Token, &s.CreatedAt, &s.ExpiresAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	return s, nil
}

func (r *PostgresSessionRepository) Delete(token string) error {
	const q = `DELETE FROM sessions WHERE token = $1`
	if _, err := r.db.Exec(q, token); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

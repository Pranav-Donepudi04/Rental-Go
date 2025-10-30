package interfaces

import "backend-form/m/internal/domain"

type SessionRepository interface {
	Create(session *domain.Session) error
	GetByToken(token string) (*domain.Session, error)
	Delete(token string) error
}

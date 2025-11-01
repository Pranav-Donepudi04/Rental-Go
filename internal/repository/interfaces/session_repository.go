package interfaces

import "backend-form/m/internal/domain"

type SessionRepository interface {
	Create(session *domain.Session) error
	GetByToken(token string) (*domain.Session, error)
	Delete(token string) error
	DeleteExpiredByUserID(userID int) error // Delete expired sessions for a user
}

package service

import (
	"backend-form/m/internal/domain"
	"backend-form/m/internal/repository/interfaces"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

type AuthService struct {
	users      interfaces.UserRepository
	sessions   interfaces.SessionRepository
	sessionTTL time.Duration
}

func NewAuthService(users interfaces.UserRepository, sessions interfaces.SessionRepository, sessionTTL time.Duration) *AuthService {
	return &AuthService{users: users, sessions: sessions, sessionTTL: sessionTTL}
}

func (s *AuthService) HashPassword(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

func (s *AuthService) ComparePassword(hash, plain string) bool {
	return hash == s.HashPassword(plain)
}

func (s *AuthService) GenerateTempPassword() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate temp password: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (s *AuthService) Login(phone, password string) (*domain.Session, *domain.User, error) {
	user, err := s.users.GetByPhone(phone)
	if err != nil {
		return nil, nil, err
	}
	if user == nil || !user.IsActive {
		return nil, nil, errors.New("invalid credentials")
	}
	if !s.ComparePassword(user.PasswordHash, password) {
		return nil, nil, errors.New("invalid credentials")
	}

	// Clean up expired sessions for this user before creating a new one
	// This prevents session accumulation
	_ = s.sessions.DeleteExpiredByUserID(user.ID)

	token, err := s.generateToken()
	if err != nil {
		return nil, nil, err
	}
	now := time.Now()
	sess := &domain.Session{UserID: user.ID, Token: token, CreatedAt: now, ExpiresAt: now.Add(s.sessionTTL)}
	if err := s.sessions.Create(sess); err != nil {
		return nil, nil, err
	}
	return sess, user, nil
}

func (s *AuthService) Logout(token string) error {
	return s.sessions.Delete(token)
}

func (s *AuthService) generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ValidateSession returns session if valid and not expired.
func (s *AuthService) ValidateSession(token string) (*domain.Session, error) {
	sess, err := s.sessions.GetByToken(token)
	if err != nil || sess == nil {
		return nil, errors.New("invalid session")
	}
	if time.Now().After(sess.ExpiresAt) {
		_ = s.sessions.Delete(token)
		return nil, errors.New("session expired")
	}
	return sess, nil
}

// CreateTenantCredentials creates a login for a tenant and returns a temporary password.
func (s *AuthService) CreateTenantCredentials(phone string, tenantID int) (string, error) {
	// if user exists, just generate and set a new temp password
	user, err := s.users.GetByPhone(phone)
	if err != nil {
		return "", err
	}
	temp, err := s.GenerateTempPassword()
	if err != nil {
		return "", err
	}
	hash := s.HashPassword(temp)
	if user == nil {
		// create new tenant user
		u := &domain.User{Phone: phone, PasswordHash: hash, UserType: domain.UserTypeTenant, TenantID: &tenantID, IsActive: true}
		if err := s.users.CreateTenantUser(u); err != nil {
			return "", err
		}
		return temp, nil
	}
	// existing user: update password and ensure linkage
	if err := s.users.UpdatePassword(user.ID, hash); err != nil {
		return "", err
	}
	if user.TenantID == nil {
		if err := s.users.LinkTenant(user.ID, tenantID); err != nil {
			return "", err
		}
	}
	return temp, nil
}

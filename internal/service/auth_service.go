package service

import (
	"backend-form/m/internal/domain"
	"backend-form/m/internal/repository/interfaces"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users      interfaces.UserRepository
	sessions   interfaces.SessionRepository
	sessionTTL time.Duration
}

func NewAuthService(users interfaces.UserRepository, sessions interfaces.SessionRepository, sessionTTL time.Duration) *AuthService {
	return &AuthService{users: users, sessions: sessions, sessionTTL: sessionTTL}
}

// HashPassword hashes a password using bcrypt (production-ready)
// Cost factor of 10 provides good security/performance balance
func (s *AuthService) HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// hashSHA256 is a legacy method kept only for backward compatibility during migration
// DO NOT USE for new passwords - use HashPassword instead
func (s *AuthService) hashSHA256(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

// isBcryptHash checks if a hash is a bcrypt hash (starts with $2a$, $2b$, or $2y$)
func (s *AuthService) isBcryptHash(hash string) bool {
	return strings.HasPrefix(hash, "$2a$") || strings.HasPrefix(hash, "$2b$") || strings.HasPrefix(hash, "$2y$")
}

// ValidatePasswordStrength validates password complexity requirements
// Returns error if password doesn't meet requirements, nil otherwise
func (s *AuthService) ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	var missing []string
	if !hasUpper {
		missing = append(missing, "uppercase letter")
	}
	if !hasLower {
		missing = append(missing, "lowercase letter")
	}
	if !hasNumber {
		missing = append(missing, "number")
	}
	if !hasSpecial {
		missing = append(missing, "special character")
	}

	if len(missing) > 0 {
		return fmt.Errorf("password must contain at least one: %s", strings.Join(missing, ", "))
	}

	return nil
}

// ComparePassword compares a password with a hash, supporting both bcrypt (new) and SHA256 (legacy)
// Returns true if password matches, false otherwise
func (s *AuthService) ComparePassword(hash, plain string) bool {
	// Check if it's a bcrypt hash (new format)
	if s.isBcryptHash(hash) {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
		return err == nil
	}
	// Legacy SHA256 support for backward compatibility
	sha256Hash := s.hashSHA256(plain)
	return hash == sha256Hash
}

func (s *AuthService) GenerateTempPassword() (string, error) {
	// Generate a simple 6-character alphanumeric password
	// Using uppercase letters and numbers only for easy sharing
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Excludes easily confused chars (I, O, 0, 1)
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate temp password: %w", err)
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
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

	// Auto-upgrade legacy SHA256 passwords to bcrypt on successful login
	if !s.isBcryptHash(user.PasswordHash) {
		newHash, err := s.HashPassword(password)
		if err == nil {
			// Silently upgrade password hash (non-blocking - if it fails, user can still login)
			_ = s.users.UpdatePassword(user.ID, newHash)
		}
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
	hash, err := s.HashPassword(temp)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
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
	// Always ensure tenant is linked (update if different, link if nil)
	// This handles cases where tenant was recreated or user was unlinked
	if user.TenantID == nil || *user.TenantID != tenantID {
		if err := s.users.LinkTenant(user.ID, tenantID); err != nil {
			return "", fmt.Errorf("failed to link tenant: %w", err)
		}
	}
	return temp, nil
}

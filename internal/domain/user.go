package domain

import "time"

type UserType string

const (
	UserTypeOwner  UserType = "owner"
	UserTypeTenant UserType = "tenant"
)

type User struct {
	ID           int
	Phone        string
	PasswordHash string
	UserType     UserType
	TenantID     *int
	IsActive     bool
	CreatedAt    time.Time
}

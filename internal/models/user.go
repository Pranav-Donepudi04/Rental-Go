package models

import (
	"fmt"
	"strings"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       string    `json:"age"`
	Phone     string    `json:"phone"`
	Website   string    `json:"website"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func (u *User) Validate() error {
	var errors []string

	if strings.TrimSpace(u.Name) == "" {
		errors = append(errors, "name is required")
	}

	if strings.TrimSpace(u.Email) == "" {
		errors = append(errors, "email is required")
	} else if !strings.Contains(u.Email, "@") {
		errors = append(errors, "email format is invalid")
	}

	if strings.TrimSpace(u.Phone) == "" {
		errors = append(errors, "phone number is required")
	}

	if len(u.Message) > 500 {
		errors = append(errors, "message must be less than 500 characters")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, ", "))
	}

	return nil
}

package domain

import (
	"strings"
	"testing"
	"time"
)

func TestTenant_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tenant  *Tenant
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid tenant",
			tenant: &Tenant{
				Name:           "John Doe",
				Phone:          "9876543210",
				AadharNumber:   "123456789012",
				MoveInDate:     time.Now(),
				NumberOfPeople: 2,
				UnitID:         1,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			tenant: &Tenant{
				Name:           "",
				Phone:          "9876543210",
				AadharNumber:   "123456789012",
				MoveInDate:     time.Now(),
				NumberOfPeople: 2,
				UnitID:         1,
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "whitespace only name",
			tenant: &Tenant{
				Name:           "   ",
				Phone:          "9876543210",
				AadharNumber:   "123456789012",
				MoveInDate:     time.Now(),
				NumberOfPeople: 2,
				UnitID:         1,
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "empty phone",
			tenant: &Tenant{
				Name:           "John Doe",
				Phone:          "",
				AadharNumber:   "123456789012",
				MoveInDate:     time.Now(),
				NumberOfPeople: 2,
				UnitID:         1,
			},
			wantErr: true,
			errMsg:  "phone is required",
		},
		{
			name: "empty aadhar",
			tenant: &Tenant{
				Name:           "John Doe",
				Phone:          "9876543210",
				AadharNumber:   "",
				MoveInDate:     time.Now(),
				NumberOfPeople: 2,
				UnitID:         1,
			},
			wantErr: true,
			errMsg:  "aadhar number is required",
		},
		{
			name: "aadhar too short",
			tenant: &Tenant{
				Name:           "John Doe",
				Phone:          "9876543210",
				AadharNumber:   "12345678901",
				MoveInDate:     time.Now(),
				NumberOfPeople: 2,
				UnitID:         1,
			},
			wantErr: true,
			errMsg:  "aadhar number must be 12 digits",
		},
		{
			name: "aadhar too long",
			tenant: &Tenant{
				Name:           "John Doe",
				Phone:          "9876543210",
				AadharNumber:   "1234567890123",
				MoveInDate:     time.Now(),
				NumberOfPeople: 2,
				UnitID:         1,
			},
			wantErr: true,
			errMsg:  "aadhar number must be 12 digits",
		},
		{
			name: "aadhar exactly 12 digits",
			tenant: &Tenant{
				Name:           "John Doe",
				Phone:          "9876543210",
				AadharNumber:   "123456789012",
				MoveInDate:     time.Now(),
				NumberOfPeople: 2,
				UnitID:         1,
			},
			wantErr: false,
		},
		{
			name: "zero move-in date",
			tenant: &Tenant{
				Name:           "John Doe",
				Phone:          "9876543210",
				AadharNumber:   "123456789012",
				MoveInDate:     time.Time{},
				NumberOfPeople: 2,
				UnitID:         1,
			},
			wantErr: true,
			errMsg:  "move-in date is required",
		},
		{
			name: "negative number of people",
			tenant: &Tenant{
				Name:           "John Doe",
				Phone:          "9876543210",
				AadharNumber:   "123456789012",
				MoveInDate:     time.Now(),
				NumberOfPeople: -1,
				UnitID:         1,
			},
			wantErr: true,
			errMsg:  "number of people must be greater than 0",
		},
		{
			name: "zero number of people",
			tenant: &Tenant{
				Name:           "John Doe",
				Phone:          "9876543210",
				AadharNumber:   "123456789012",
				MoveInDate:     time.Now(),
				NumberOfPeople: 0,
				UnitID:         1,
			},
			wantErr: true,
			errMsg:  "number of people must be greater than 0",
		},
		{
			name: "negative unit ID",
			tenant: &Tenant{
				Name:           "John Doe",
				Phone:          "9876543210",
				AadharNumber:   "123456789012",
				MoveInDate:     time.Now(),
				NumberOfPeople: 2,
				UnitID:         -1,
			},
			wantErr: true,
			errMsg:  "unit ID is required",
		},
		{
			name: "zero unit ID",
			tenant: &Tenant{
				Name:           "John Doe",
				Phone:          "9876543210",
				AadharNumber:   "123456789012",
				MoveInDate:     time.Now(),
				NumberOfPeople: 2,
				UnitID:         0,
			},
			wantErr: true,
			errMsg:  "unit ID is required",
		},
		{
			name: "all fields invalid",
			tenant: &Tenant{
				Name:           "",
				Phone:          "",
				AadharNumber:   "",
				MoveInDate:     time.Time{},
				NumberOfPeople: 0,
				UnitID:         0,
			},
			wantErr: true,
			// Should fail on first validation (name)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tenant.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Tenant.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Tenant.Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr error
	}{
		{
			name: "valid user with all fields",
			user: User{
				ID:        ID("user-123"),
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "valid user with minimal fields",
			user: User{
				FirstName: "Jane",
				LastName:  "Doe",
				Email:     "jane@example.com",
			},
			wantErr: nil,
		},
		{
			name: "missing email",
			user: User{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "",
			},
			wantErr: ErrInvalidEmail,
		},
		{
			name: "whitespace-only email",
			user: User{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "   ",
			},
			wantErr: ErrInvalidEmail,
		},
		{
			name: "missing first name",
			user: User{
				FirstName: "",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
			wantErr: ErrInvalidFirstName,
		},
		{
			name: "whitespace-only first name",
			user: User{
				FirstName: "   ",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
			wantErr: ErrInvalidFirstName,
		},
		{
			name: "missing last name",
			user: User{
				FirstName: "John",
				LastName:  "",
				Email:     "john@example.com",
			},
			wantErr: ErrInvalidLastName,
		},
		{
			name: "whitespace-only last name",
			user: User{
				FirstName: "John",
				LastName:  "   ",
				Email:     "john@example.com",
			},
			wantErr: ErrInvalidLastName,
		},
		{
			name: "missing both email and names returns email error first",
			user: User{
				FirstName: "",
				LastName:  "",
				Email:     "",
			},
			wantErr: ErrInvalidEmail,
		},
		{
			name: "valid email but missing names returns first name error",
			user: User{
				FirstName: "",
				LastName:  "",
				Email:     "john@example.com",
			},
			wantErr: ErrInvalidFirstName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

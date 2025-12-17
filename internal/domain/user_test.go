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
				Name:      "John Doe",
				Email:     "john@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "valid user with minimal fields",
			user: User{
				Name:  "Jane Doe",
				Email: "jane@example.com",
			},
			wantErr: nil,
		},
		{
			name: "missing email",
			user: User{
				Name:  "John Doe",
				Email: "",
			},
			wantErr: ErrInvalidEmail,
		},
		{
			name: "whitespace-only email",
			user: User{
				Name:  "John Doe",
				Email: "   ",
			},
			wantErr: ErrInvalidEmail,
		},
		{
			name: "missing name",
			user: User{
				Name:  "",
				Email: "john@example.com",
			},
			wantErr: ErrInvalidUserName,
		},
		{
			name: "whitespace-only name",
			user: User{
				Name:  "   ",
				Email: "john@example.com",
			},
			wantErr: ErrInvalidUserName,
		},
		{
			name: "missing both email and name returns email error first",
			user: User{
				Name:  "",
				Email: "",
			},
			wantErr: ErrInvalidEmail,
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

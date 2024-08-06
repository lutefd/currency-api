package model_test

import (
	"testing"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestToUser(t *testing.T) {
	tests := []struct {
		name     string
		userDB   model.UserDB
		expected model.User
	}{
		{
			name: "RoleUser",
			userDB: model.UserDB{
				ID:        uuid.New(),
				Username:  "testuser",
				Password:  "testpass",
				Role:      model.RoleUser,
				APIKey:    "testapikey",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expected: model.User{
				ID:       uuid.New(),
				Username: "testuser",
				Role:     model.RoleUser,
				APIKey:   "testapikey",
			},
		},
		{
			name: "RoleAdmin",
			userDB: model.UserDB{
				ID:        uuid.New(),
				Username:  "adminuser",
				Password:  "adminpass",
				Role:      model.RoleAdmin,
				APIKey:    "adminkey",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expected: model.User{
				ID:       uuid.New(),
				Username: "adminuser",
				Role:     model.RoleAdmin,
				APIKey:   "adminkey",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.userDB.ToUser()

			tt.expected.ID = tt.userDB.ID

			assert.Equal(t, tt.expected, actual)
		})
	}
}

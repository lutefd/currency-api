package model

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type UserDB struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Role      Role      `json:"role"`
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Role     Role      `json:"role"`
	APIKey   string    `json:"api_key"`
}

func (u *UserDB) ToUser() User {
	return User{
		ID:       u.ID,
		Username: u.Username,
		Role:     u.Role,
		APIKey:   u.APIKey,
	}
}

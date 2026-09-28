package models

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleUser      Role = "USER"
	RoleModerator Role = "MODERATOR"
	RoleAdmin     Role = "ADMIN"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Avatar       *string   `json:"avatar,omitempty"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

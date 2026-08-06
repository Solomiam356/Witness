package domain

import (
	"time"
)

type UserRole string

const (
	RoleUser	    UserRole = "user"
	RoleModerator	UserRole = "moderator"
	RoleAdmin       UserRole = "admin"
)

type User struct {
	ID            string     `json:"id"`
	Email         string     `json:"email"`
	PasswordHash  string     `json:"-"` 
	DisplayName   string     `json:"display_name"`
	Role          UserRole   `json:"role"`
	EmailVerified bool       `json:"email_verified"`
	CurrentStreak int        `json:"current_streak"`
	LastActiveAt  time.Time  `json:"last_active_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}
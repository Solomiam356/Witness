package domain

import (
	"time"
)

type Session struct {
	ID               string     `json:"id"`
	UserID           string     `json:"user_id"`
	RefreshTokenHash string     `json:"-"`
	DeviceInfo       string     `json:"device_info"`
	ExpiresAt        time.Time  `json:"expires_at"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

func (s *Session) IsValid() bool {
	return time.Now().Before(s.ExpiresAt) && s.RevokedAt == nil
}
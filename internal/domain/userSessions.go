package domain

import (
	"time"
)

type UserSession struct {
	UserID       int64     `db:"user_id"`
	RefreshToken string    `db:"refresh_token"`
	ExpiresAt    time.Time `db:"expires_at"`
}

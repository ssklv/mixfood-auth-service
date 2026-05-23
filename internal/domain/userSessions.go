package domain

import "time"

type UserSession struct {
	ID           int64     `json:"id" db:"id"`
	UserID       int64     `json:"userId" db:"user_id"`
	RefreshToken string    `json:"refreshToken" db:"refresh_token"`
	ExpiresAt    time.Time `json:"expiresAt" db:"expires_at"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}

package domain

import (
	"time"
)

type UserRole string

const (
	RoleUser  UserRole = "user" //client
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID           int64     `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Phone        string    `json:"phone" db:"phone"`
	Email        string    `json:"email" db:"email"`
	Address      string    `json:"address" db:"address"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         UserRole  `json:"role" db:"role"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

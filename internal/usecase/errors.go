package usecase

import (
	"errors"
)

var (
	ErrInvalidName            = errors.New("name is too short or exceeds maximum length")
	ErrInvalidPhone           = errors.New("invalid phone number format")
	ErrInvalidEmail           = errors.New("invalid email format")
	ErrInvalidAddress         = errors.New("address is too long")
	ErrInvalidPasswordTooWeak = errors.New("password must be at least 6 characters long")

	//ErrInvalidPasswod

	ErrUserAlreadyExists  = errors.New("user with this phone number already exists")
	ErrInvalidCredentials = errors.New("invalid phone number or password")
	ErrSessionNotFound    = errors.New("active session not found")
	ErrSessionExpired     = errors.New("session expired")
)

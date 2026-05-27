package infrastructure

import "errors"

var (
	//для бд и профилей юзеров
	ErrUserNotFound     = errors.New("user not found")
	ErrDuplicatePhone   = errors.New("user with this phone number already exists")
	ErrDuplicateEmail   = errors.New("this email is already taken")
	ErrDatabaseInternal = errors.New("internal database error")
	//
	ErrInvalidToken     = errors.New("invalid or expired token")
	ErrPasswordMismatch = errors.New("invalid phone number or password")
	ErrSessionNotFound  = errors.New("session not found")

	ErrAddressNotFound = errors.New("address not found")
)

//адрес спросить

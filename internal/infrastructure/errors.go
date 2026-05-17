package infrastructure

import (
	"errors"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrNoChanges      = errors.New("no changes provided")
	ErrDuplicatePhone = errors.New("duplicate phone")
	ErrDuplicateEmail = errors.New("duplicate email")
)

//адрес в сервис с доставкой

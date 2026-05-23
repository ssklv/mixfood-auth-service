package usecase

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	minPasswordLen = 6
	maxNameLen     = 30
	maxAddressLen  = 100
	maxEmailLen    = 255
)

var phoneRegex = regexp.MustCompile(`^(?:\+7|7|8)?\d{10}$`)

// ErrInvalindPhone
// ErrInvalidName
// ErrInvalidPasswodTooWeek
// ErrInvalidEmail
// ErrInvalidAddress -- спросить про адрес

func validatePassword(password string) error {
	if utf8.RuneCountInString(password) < minPasswordLen {
		return ErrInvalidPasswordTooWeak
	}
	return nil
}

func validatePhone(phone string) error {
	if !phoneRegex.MatchString(phone) {
		return ErrInvalidPhone
	}
	return nil
}

func validateName(name string) error {
	count := utf8.RuneCountInString(strings.TrimSpace(name))
	if count == 0 || count > maxNameLen {
		return ErrInvalidName
	}
	return nil
}

// /
func validateEmail(email string) error {
	if email == "" || len(email) > maxEmailLen {
		return ErrInvalidEmail
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmail
	}
	return nil
}

func validateAddress(address string) error {
	count := utf8.RuneCountInString(strings.TrimSpace(address))
	if count > maxAddressLen {
		return ErrInvalidAddress
	}
	return nil
}

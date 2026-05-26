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

	maxApartmentLen = 20
	maxEntranceLen  = 10
	maxFloorLen     = 10
	maxDoorCodeLen  = 20
)

var phoneRegex = regexp.MustCompile(`^(?:\+7|7|8)?\d{10}$`)

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

func validateEmail(email string) error {
	if email == "" || len(email) > maxEmailLen {
		return ErrInvalidEmail
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmail
	}
	return nil
}

func validateStreetHouse(streetHouse string) error {
	count := utf8.RuneCountInString(strings.TrimSpace(streetHouse))
	if count == 0 || count > maxAddressLen {
		return ErrInvalidAddress
	}
	return nil
}

func validateApartment(val string) error {
	if utf8.RuneCountInString(val) > maxApartmentLen {
		return ErrInvalidAddress
	}
	return nil
}

func validateEntrance(val string) error {
	if utf8.RuneCountInString(val) > maxEntranceLen {
		return ErrInvalidAddress
	}
	return nil
}

func validateFloor(val string) error {
	if utf8.RuneCountInString(val) > maxFloorLen {
		return ErrInvalidAddress
	}
	return nil
}

func validateDoorCode(val string) error {
	if utf8.RuneCountInString(val) > maxDoorCodeLen {
		return ErrInvalidAddress
	}
	return nil
}

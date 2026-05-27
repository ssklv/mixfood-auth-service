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
	maxEmailLen    = 255

	maxStreetHouse  = 100
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

// //
func validateStreetHouse(streetHouse string) error {
	count := utf8.RuneCountInString(strings.TrimSpace(streetHouse))
	if count == 0 || count > maxStreetHouse {
		return ErrInvalidAddress
	}
	return nil
}

func validateApartment(apartment string) error {
	if utf8.RuneCountInString(apartment) > maxApartmentLen {
		return ErrInvalidAddress
	}
	return nil
}

func validateEntrance(entrance string) error {
	if utf8.RuneCountInString(entrance) > maxEntranceLen {
		return ErrInvalidAddress
	}
	return nil
}

func validateFloor(floor string) error {
	if utf8.RuneCountInString(floor) > maxFloorLen {
		return ErrInvalidAddress
	}
	return nil
}

func validateDoorCode(doorCode string) error {
	if utf8.RuneCountInString(doorCode) > maxDoorCodeLen {
		return ErrInvalidAddress
	}
	return nil
}

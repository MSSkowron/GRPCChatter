package validation

import (
	"errors"
	"strings"
)

var (
	// ErrInvalidUsername is returned when an invalid username is provided.
	ErrInvalidUsername = errors.New("user name must must not be empty and have at least 6 characters, including digits")
	// ErrInvalidPassword is returned when an invalid password is provided.
	ErrInvalidPassword = errors.New("password must not be empty and must have at least 6 characters, including 1 uppercase letter, 1 lowercase letter, 1 digit and 1 special character")
)

// ValidateUsername validates the provided username.
// It checks if the username is at least 6 characters long, contains at least one digit,
// and does not contain any special characters.
func ValidateUsername(username string) error {
	valid := len(username) >= 6 &&
		strings.ContainsAny(username, "0123456789") &&
		!strings.ContainsAny(username, "!@#$%^&*()_+[]{};':,.<>?/")

	if !valid {
		return ErrInvalidUsername
	}

	return nil
}

// ValidatePassword validates the provided password.
// It checks if the password is at least 6 characters long, contains at least one uppercase letter,
// one lowercase letter, one digit, and one special character.
func ValidatePassword(password string) error {
	valid := len(password) >= 6 &&
		strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") &&
		strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") &&
		strings.ContainsAny(password, "0123456789") &&
		strings.ContainsAny(password, "!@#$%^&*()_+[]{};':,.<>?/")

	if !valid {
		return ErrInvalidPassword
	}

	return nil
}

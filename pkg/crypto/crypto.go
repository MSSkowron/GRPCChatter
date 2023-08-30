package crypto

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidCredentials is returned when the credentials provided are invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")
)

const (
	// Cost is the cost used to hash passwords.
	Cost = 10
)

// HashPassword hashes a password with bcrypt and the default cost of 10.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), Cost)
	return string(bytes), err
}

// CheckPassword checks if a password matches a hash with bcrypt.
func CheckPassword(password, hash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		}

		return err
	}

	return nil
}

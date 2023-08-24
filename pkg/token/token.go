package token

import (
	"errors"

	"github.com/golang-jwt/jwt"
)

const (
	claimUserNameKey  = "userName"
	claimShortCodeKey = "shortCode"
)

var (
	// ErrInvalidToken is returned when the token is invalid.
	ErrInvalidToken = errors.New("invalid token")
	// ErrInvalidSignature is returned when the token signature is invalid.
	ErrInvalidSignature = errors.New("invalid signature")
)

// Generate generates a new JWT token.
// The token is signed with the given secret.
// The token contains the user name and short code.
func Generate(userName, shortCode, secret string) (tokenString string, err error) {
	claims := &jwt.MapClaims{
		"userName":  userName,
		"shortCode": shortCode,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

// Validate validates the given JWT token.
func Validate(tokenString, secret string) error {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		_, ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidSignature
		}

		return []byte(secret), nil
	})
	if err != nil {
		return ErrInvalidToken
	}

	if !token.Valid {
		return ErrInvalidToken
	}

	userName, ok := token.Claims.(jwt.MapClaims)[claimUserNameKey].(string)
	if !ok || userName == "" {
		return ErrInvalidToken
	}

	shortCode, ok := token.Claims.(jwt.MapClaims)[claimShortCodeKey].(string)
	if !ok || shortCode == "" {
		return ErrInvalidToken
	}

	return nil
}

// GetClaimValue retrieves a claim value with the given key from the given JWT token.
func GetClaimValue(tokenString, secret, key string) (string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		_, ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidSignature
		}

		return []byte(secret), nil
	})
	if err != nil {
		return "", ErrInvalidToken
	}

	userName, ok := token.Claims.(jwt.MapClaims)[key].(string)
	if !ok {
		return "", ErrInvalidToken
	}

	return userName, nil
}

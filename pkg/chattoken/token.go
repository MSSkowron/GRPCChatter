package chattoken

import (
	"errors"

	"github.com/golang-jwt/jwt"
)

const (
	ClaimUserNameKey  = "userName"
	ClaimShortCodeKey = "shortCode"
)

var (
	// ErrInvalidToken is returned when the token is invalid.
	ErrInvalidToken = errors.New("invalid token")
)

// Generate generates a new JWT token.
// The token is signed with the given secret.
// The token contains the user name and short code.
func Generate(userName, shortCode, secret string) (tokenString string, err error) {
	claims := &jwt.MapClaims{
		ClaimUserNameKey:  userName,
		ClaimShortCodeKey: shortCode,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

// Validate validates the given JWT token.
func Validate(tokenString, secret string) error {
	token, err := parse(tokenString, secret)
	if err != nil || !token.Valid {
		return ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ErrInvalidToken
	}

	userName, ok := claims[ClaimUserNameKey].(string)
	if !ok || userName == "" {
		return ErrInvalidToken
	}

	shortCode, ok := claims[ClaimShortCodeKey].(string)
	if !ok || shortCode == "" {
		return ErrInvalidToken
	}

	return nil
}

// GetClaim retrieves a claim value with the given key from the given JWT token.
func GetClaim(tokenString, secret, key string) (string, error) {
	token, err := parse(tokenString, secret)
	if err != nil || !token.Valid {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidToken
	}

	value, ok := claims[key].(string)
	if !ok {
		return "", ErrInvalidToken
	}

	return value, nil
}

func parse(tokenString, secret string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		_, ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
}

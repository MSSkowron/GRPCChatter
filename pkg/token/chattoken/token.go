package chattoken

import (
	"errors"

	"github.com/MSSkowron/GRPCChatter/pkg/token"
	"github.com/golang-jwt/jwt"
)

const (
	ClaimUserNameKey  = "userName"
	ClaimShortCodeKey = "shortCode"
)

// ErrInvalidToken is returned when the token is invalid.
var ErrInvalidToken = errors.New("invalid token")

// Generate generates a new JWT token with user name and short code.
func Generate(userName, shortCode, secret string) (string, error) {
	claims := &jwt.MapClaims{
		ClaimUserNameKey:  userName,
		ClaimShortCodeKey: shortCode,
	}

	return token.NewWithClaims(claims, secret)
}

// Validate validates the given JWT token.
func Validate(tokenString, secret string) error {
	token, err := token.Parse(tokenString, secret)
	if err != nil || !token.Valid {
		return ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ErrInvalidToken
	}

	_, ok = claims[ClaimUserNameKey].(string)
	if !ok {
		return ErrInvalidToken
	}

	_, ok = claims[ClaimShortCodeKey].(string)
	if !ok {
		return ErrInvalidToken
	}

	return nil
}

// GetClaim retrieves a claim value with the given key from the given JWT token.
func GetClaim(tokenString, secret, key string) (string, error) {
	token, err := token.Parse(tokenString, secret)
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

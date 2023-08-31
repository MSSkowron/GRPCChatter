package usertoken

import (
	"errors"
	"time"

	"github.com/MSSkowron/GRPCChatter/pkg/token"
	"github.com/golang-jwt/jwt"
)

const (
	ClaimUserIDKey    = "id"
	ClaimUserNameKey  = "userName"
	ClaimExpiresAtKey = "expiresAt"
)

var (
	// ErrInvalidToken is returned when the token is invalid.
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when the token is expired.
	ErrExpiredToken = errors.New("expired token")
)

// Generate generates a new JWT token with user ID, user name, and expiration time.
func Generate(userID int, userName string, expirationTime time.Duration, secret string) (string, error) {
	expiration := time.Now().Add(expirationTime).Unix()
	claims := &jwt.MapClaims{
		ClaimUserIDKey:    userID,
		ClaimUserNameKey:  userName,
		ClaimExpiresAtKey: expiration,
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

	expiresAt, ok := token.Claims.(jwt.MapClaims)[ClaimExpiresAtKey].(float64)
	if !ok {
		return ErrInvalidToken
	}

	if int64(expiresAt) < time.Now().Local().Unix() {
		return ErrExpiredToken
	}

	_, ok = claims[ClaimUserIDKey].(float64)
	if !ok {
		return ErrInvalidToken
	}

	_, ok = claims[ClaimUserNameKey].(string)
	if !ok {
		return ErrInvalidToken
	}

	return nil
}

type Claim interface {
	string | float64
}

// GetClaim retrieves a claim value with the given key from the given JWT token.
func GetClaim[T Claim](tokenString, secret, key string) (T, error) {
	var value T

	token, err := token.Parse(tokenString, secret)
	if err != nil || !token.Valid {
		return value, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return value, ErrInvalidToken
	}

	value, ok = claims[key].(T)
	if !ok {
		return value, ErrInvalidToken
	}

	return value, nil
}

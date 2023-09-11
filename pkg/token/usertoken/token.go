package usertoken

import (
	"errors"
	"time"

	"github.com/MSSkowron/GRPCChatter/pkg/token"
	"github.com/golang-jwt/jwt"
)

const (
	// ClaimUserIDKey is the key for user ID claim.
	ClaimUserIDKey = "id"
	// ClaimUserNameKey is the key for user name claim.
	ClaimUserNameKey = "userName"
	// ClaimUserRolle is the key for user role claim.
	ClaimUserRoleKey = "role"
	// ClaimExpiresAtKey is the key for expiration time claim.
	ClaimExpiresAtKey = "expiresAt"
)

var (
	// ErrInvalidToken is returned when the token is invalid.
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when the token is expired.
	ErrExpiredToken = errors.New("expired token")
)

// Generate generates a new JWT token with user ID, user name, user role and expiration time.
func Generate(userID int, userName string, role string, expirationTime time.Duration, secret string) (string, error) {
	expiration := time.Now().Add(expirationTime).Unix()
	claims := &jwt.MapClaims{
		ClaimUserIDKey:    userID,
		ClaimUserNameKey:  userName,
		ClaimUserRoleKey:  role,
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

	if _, ok := claims[ClaimUserIDKey].(float64); !ok {
		return ErrInvalidToken
	}

	if _, ok := claims[ClaimUserNameKey].(string); !ok {
		return ErrInvalidToken
	}

	if _, ok := claims[ClaimUserRoleKey].(string); !ok {
		return ErrInvalidToken
	}

	return nil
}

// Claim is a generic type for JWT claims.
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

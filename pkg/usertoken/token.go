package usertoken

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

type Claim interface {
	string | float64
}

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

// Generate generates a new JWT token.
// The token is signed with the given secret.
// The token contains the user id, user name, and expiration time.
func Generate(userID int, userName string, expirationTime time.Duration, secret string) (string, error) {
	claims := &jwt.MapClaims{
		ClaimUserIDKey:    userID,
		ClaimUserNameKey:  userName,
		ClaimExpiresAtKey: time.Now().Add(expirationTime).Unix(),
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

// GetClaim retrieves a claim value with the given key from the given JWT token.
func GetClaim[T Claim](tokenString, secret, key string) (T, error) {
	var value T

	token, err := parse(tokenString, secret)
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

func parse(tokenString, secret string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		_, ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
}

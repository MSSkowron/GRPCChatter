package token

import (
	"fmt"

	"github.com/golang-jwt/jwt"
)

// DefaultSigningMethod is the default signing method for JWT tokens.
var DefaultSigningMethod = jwt.SigningMethodHS256

// NewWithClaims creates a JWT token with the provided claims and secret.
func NewWithClaims(claims *jwt.MapClaims, secret string) (string, error) {
	token := jwt.NewWithClaims(DefaultSigningMethod, claims)
	return token.SignedString([]byte(secret))
}

// Parse parses a JWT token string using the provided secret and returns the token.
func Parse(tokenString, secret string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if token.Method != DefaultSigningMethod {
			return nil, fmt.Errorf("invalid signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}

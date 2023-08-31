package service

import (
	"errors"
	"fmt"
	"time"

	token "github.com/MSSkowron/GRPCChatter/pkg/token/usertoken"
)

// ErrInvalidUserToken is returned when the token is invalid.
var ErrInvalidUserToken = errors.New("invalid token")

// UserTokenService is an interface that defines the methods required for user token management.
type UserTokenService interface {
	// GenerateToken generates a user token.
	GenerateToken(int, string) (string, error)

	// ValidateToken validates a user token.
	ValidateToken(string) error

	// GetUserIDFromToken retrieves the user ID from a user token.
	GetUserIDFromToken(string) (int, error)

	// GetUserNameFromToken retrieves the user name from a user token.
	GetUserNameFromToken(string) (string, error)
}

// UserTokenServiceImpl implements the UserTokenService interface.
type UserTokenServiceImpl struct {
	secret   string
	duration time.Duration
}

// NewUserTokenService creates a new UserTokenServiceImpl instance with the provided secret and duration.
func NewUserTokenService(secret string, duration time.Duration) *UserTokenServiceImpl {
	return &UserTokenServiceImpl{
		secret:   secret,
		duration: duration,
	}
}

func (s *UserTokenServiceImpl) GenerateToken(id int, userName string) (string, error) {
	token, err := token.Generate(id, userName, s.duration, s.secret)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return token, nil
}

func (s *UserTokenServiceImpl) ValidateToken(t string) error {
	if err := token.Validate(t, s.secret); err != nil {
		return ErrInvalidChatToken
	}
	return nil
}

func (s *UserTokenServiceImpl) GetUserIDFromToken(t string) (int, error) {
	userID, err := token.GetClaim[float64](t, s.secret, token.ClaimUserIDKey)
	if err != nil {
		return 0, ErrInvalidChatToken
	}
	return int(userID), nil
}

func (s *UserTokenServiceImpl) GetUserNameFromToken(t string) (string, error) {
	userName, err := token.GetClaim[string](t, s.secret, token.ClaimUserNameKey)
	if err != nil {
		return "", ErrInvalidChatToken
	}
	return userName, nil
}

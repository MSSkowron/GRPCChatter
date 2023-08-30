package service

import (
	"errors"
	"fmt"

	"github.com/MSSkowron/GRPCChatter/pkg/token"
)

// ErrInvalidChatToken is returned when the token is invalid.
var ErrInvalidChatToken = errors.New("invalid token")

// ChatTokenService is an interface that defines the methods required for token management.
type ChatTokenService interface {
	// GenerateToken generates a token for a given username and short code.
	// It returns the generated token and an error if the generation fails.
	GenerateToken(username, shortCode string) (string, error)

	// ValidateToken validates a token and returns an error if it's invalid.
	ValidateToken(token string) error

	// GetUserNameFromToken retrieves the username from a token.
	// It returns the username and an error if the retrieval fails.
	GetUserNameFromToken(token string) (string, error)

	// GetShortCodeFromToken retrieves the short code from a token.
	// It returns the short code and an error if the retrieval fails.
	GetShortCodeFromToken(token string) (string, error)
}

// ChatTokenServiceImpl implements the ChatTokenService interface.
type ChatTokenServiceImpl struct {
	secret string
}

// NewChatTokenService creates a new ChatTokenServiceImpl instance with the provided secret.
func NewChatTokenService(secret string) ChatTokenService {
	return &ChatTokenServiceImpl{
		secret: secret,
	}
}

// GenerateToken generates a token for a given username and short code.
// It returns the generated token and an error if the generation fails.
func (s *ChatTokenServiceImpl) GenerateToken(username, shortCode string) (string, error) {
	token, err := token.Generate(username, shortCode, s.secret)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return token, nil
}

// ValidateToken validates a token and returns an error if it's invalid.
func (s *ChatTokenServiceImpl) ValidateToken(t string) error {
	if err := token.Validate(t, s.secret); err != nil {
		return ErrInvalidChatToken
	}
	return nil
}

// GetUserNameFromToken retrieves the username from a token.
// It returns the username and an error if the retrieval fails.
func (s *ChatTokenServiceImpl) GetUserNameFromToken(t string) (string, error) {
	userName, err := token.GetClaim(t, s.secret, token.ClaimUserNameKey)
	if err != nil {
		return "", ErrInvalidChatToken
	}
	return userName, nil
}

// GetShortCodeFromToken retrieves the short code from a token.
// It returns the short code and an error if the retrieval fails.
func (s *ChatTokenServiceImpl) GetShortCodeFromToken(t string) (string, error) {
	shortCode, err := token.GetClaim(t, s.secret, token.ClaimShortCodeKey)
	if err != nil {
		return "", ErrInvalidChatToken
	}
	return shortCode, nil
}

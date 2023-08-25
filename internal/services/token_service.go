package services

import "github.com/MSSkowron/GRPCChatter/pkg/token"

// TokenService is an interface that defines the methods required for token management.
type TokenService interface {
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

// TokenServiceImpl implements the TokenService interface.
type TokenServiceImpl struct {
	secret string
}

// NewTokenService creates a new TokenServiceImpl instance with the provided secret.
func NewTokenService(secret string) *TokenServiceImpl {
	return &TokenServiceImpl{
		secret: secret,
	}
}

func (s *TokenServiceImpl) GenerateToken(username, shortCode string) (string, error) {
	return token.Generate(username, shortCode, s.secret)
}

func (s *TokenServiceImpl) ValidateToken(t string) error {
	return token.Validate(t, s.secret)
}

func (s *TokenServiceImpl) GetUserNameFromToken(t string) (string, error) {
	return token.GetClaim(t, s.secret, token.ClaimUserNameKey)
}

func (s *TokenServiceImpl) GetShortCodeFromToken(t string) (string, error) {
	return token.GetClaim(t, s.secret, token.ClaimShortCodeKey)
}

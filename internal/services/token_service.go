package services

import "github.com/MSSkowron/GRPCChatter/pkg/token"

// TokenService is an interface that defines the methods that the TokenService must implement.
type TokenService interface {
	GenerateToken(string, string) (string, error)
	ValidateToken(string) error
	GetUserNameFromToken(string) (string, error)
	GetShortCodeFromToken(string) (string, error)
}

// TokenServiceImpl implements the TokenService interface.
type TokenServiceImpl struct {
	secret string
}

// NewTokenService creates a new TokenServiceImpl.
func NewTokenService(sercret string) *TokenServiceImpl {
	return &TokenServiceImpl{
		secret: sercret,
	}
}

// GenerateToken generates a token.
func (s *TokenServiceImpl) GenerateToken(userName, shortCode string) (string, error) {
	return token.Generate(userName, shortCode, s.secret)
}

// ValidateToken validates a token.
func (s *TokenServiceImpl) ValidateToken(t string) error {
	return token.Validate(t, s.secret)
}

// GetUserNameFromToken retrieves the user name from a token.
func (s *TokenServiceImpl) GetUserNameFromToken(t string) (string, error) {
	return token.GetClaim(t, s.secret, token.ClaimUserNameKey)
}

// GetShortCodeFromToken retrieves the short code from a token.
func (s *TokenServiceImpl) GetShortCodeFromToken(t string) (string, error) {
	return token.GetClaim(t, s.secret, token.ClaimShortCodeKey)
}
